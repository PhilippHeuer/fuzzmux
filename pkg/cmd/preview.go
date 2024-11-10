package cmd

import (
	"fmt"
	"github.com/PhilippHeuer/fuzzmux/pkg/app"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon"
	"os"
	"strings"

	"github.com/PhilippHeuer/fuzzmux/pkg/config"
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
			providers := app.ConfigToReconModules(conf)
			var options []recon.Option
			for _, p := range providers {
				opts, err := p.OptionsOrCache(3600)
				if err != nil {
					log.Debug().Err(err).Str("recon", p.Name()).Msg("failed to get options")
				}

				options = append(options, opts...)
			}

			// keep first part of option (fzf option id)
			if os.Getenv("FZF_PREVIEW_TOP") != "" {
				optionId = strings.Split(optionId, conf.Finder.FZFDelimiter)[0]
			}

			// query option from cache
			option, err := recon.OptionById(options, optionId)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to find option in cache")
			}

			// call select
			selectedProvider, err := app.FindReconModuleByName(providers, option.ProviderName)
			if err != nil {
				log.Fatal().Err(err).Str("recon", option.ProviderName).Msg("failed to get recon of selected option")
			}
			err = selectedProvider.SelectOption(option)
			if err != nil {
				log.Fatal().Err(err).Str("recon", option.ProviderName).Msg("failed to run option select")
			}

			// print preview
			fmt.Printf("%s\n", option.RenderPreview())
		},
	}

	return cmd
}
