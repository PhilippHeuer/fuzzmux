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

	// TODO: custom option render logic should move into the option provider
	switch option.ProviderName {
	case "kubernetes":
		builder.WriteString("\n")
		if option.Context["clusterName"] != "" {
			builder.WriteString(fmt.Sprintf("K8S Cluster Name: %s\n", option.Context["clusterName"]))
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
	case "usql":
		builder.WriteString("\n")
		if option.Context["name"] != "" {
			builder.WriteString(fmt.Sprintf("Name: %s\n", option.Context["name"]))
		}
		if option.Context["hostname"] != "" {
			builder.WriteString(fmt.Sprintf("DB Host: %s\n", option.Context["hostname"]))
		}
		if option.Context["port"] != "" {
			builder.WriteString(fmt.Sprintf("DB Port: %s\n", option.Context["port"]))
		}
		if option.Context["username"] != "" {
			builder.WriteString(fmt.Sprintf("DB Username: %s\n", option.Context["username"]))
		}
		if option.Context["instance"] != "" {
			builder.WriteString(fmt.Sprintf("DB Instance/SID: %s\n", option.Context["instance"]))
		}
		if option.Context["database"] != "" {
			builder.WriteString(fmt.Sprintf("DB Database: %s\n", option.Context["database"]))
		}
	default:
		builder.WriteString("\n")
		if len(option.Context) > 0 {
			builder.WriteString("Context:\n")
			for k, v := range option.Context {
				builder.WriteString(fmt.Sprintf("  %s: %s\n", k, v))
			}
		}
	}

	// free-text description
	if option.Context["description"] != "" {
		builder.WriteString("\n\n" + option.Context["description"] + "\n")
	}

	return builder.String()
}
