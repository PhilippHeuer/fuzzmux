package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/PhilippHeuer/fuzzmux/pkg/config"
	"github.com/PhilippHeuer/fuzzmux/pkg/provider"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func previewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "preview",
		Short: "show option preview for a given id (preview for external fuzzy finders)",
		Run: func(cmd *cobra.Command, args []string) {
			// require option id in argument
			if len(args) != 1 {
				_ = cmd.Help()
				os.Exit(1)
			}
			optionId := args[0]

			// load config
			conf, confErr := config.ResolvedConfig()
			if confErr != nil {
				log.Fatal().Err(confErr).Msg("failed to load configuration")
			}
			providers := provider.GetProviders(conf)
			var options []provider.Option
			for _, p := range providers {
				opts, err := p.OptionsOrCache(3600)
				if err != nil {
					log.Debug().Err(err).Str("provider", p.Name()).Msg("failed to get options")
				}

				options = append(options, opts...)
			}

			// keep first part of option (fzf option id)
			if os.Getenv("FZF_PREVIEW_TOP") != "" {
				optionId = strings.Split(optionId, conf.Finder.FZFDelimiter)[0]
			}

			// query option from cache
			option, err := provider.OptionById(options, optionId)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to find option in cache")
			}

			// call select
			selectedProvider, err := provider.GetProviderByName(providers, option.ProviderName)
			if err != nil {
				log.Fatal().Err(err).Str("provider", option.ProviderName).Msg("failed to get provider of selected option")
			}
			err = selectedProvider.SelectOption(option)
			if err != nil {
				log.Fatal().Err(err).Str("provider", option.ProviderName).Msg("failed to run option select")
			}

			// print preview
			fmt.Printf("%s\n", renderPreview(option))
		},
	}

	return cmd
}

func renderPreview(option *provider.Option) string {
	var builder strings.Builder
	builder.WriteString("# " + option.DisplayName + "\n\n")
	builder.WriteString("Provider: " + option.ProviderName + "\n")
	builder.WriteString("Directory: " + option.ResolveStartDirectory(true) + " [" + option.StartDirectory + "]\n")
	if len(option.Tags) > 0 {
		builder.WriteString("\nTags:\n")
		for _, t := range option.Tags {
			builder.WriteString("- " + t + "\n")
		}
	}

	// k8s, openshift
	if option.Context["clusterName"] != "" {
		builder.WriteString(fmt.Sprintf("\nK8S Cluster Name: %s\n", option.Context["clusterName"]))
	}
	if option.Context["clusterHost"] != "" {
		builder.WriteString(fmt.Sprintf("K8S Cluster API: %s\n", option.Context["clusterHost"]))
	}
	if option.Context["clusterUser"] != "" {
		builder.WriteString(fmt.Sprintf("K8S Cluster User: %s\n", option.Context["clusterUser"]))
	}
	if option.Context["clusterType"] != "" {
		builder.WriteString(fmt.Sprintf("K8S Cluster Type: %s\n", option.Context["clusterType"]))
	}

	// free-text description
	if option.Context["description"] != "" {
		builder.WriteString("\n\n" + option.Context["description"] + "\n")
	}

	return builder.String()
}
