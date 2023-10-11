package cmd

import (
	"github.com/PhilippHeuer/tmux-tms/pkg/config"
	"github.com/PhilippHeuer/tmux-tms/pkg/extensions"
	"github.com/PhilippHeuer/tmux-tms/pkg/gotmuxutil"
	"github.com/PhilippHeuer/tmux-tms/pkg/provider"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func projectCmd() *cobra.Command {
	const providerName = "project"
	flags := RootFlags{}

	cmd := &cobra.Command{
		Use:     "project",
		Aliases: []string{"p"},
		Short:   "Searches for projects with a fuzzy finder to start a new session.",
		Run: func(cmd *cobra.Command, args []string) {
			// load config
			conf, err := config.ResolvedConfig()
			if err != nil {
				log.Fatal().Err(err).Msg("failed to load configuration")
			}

			// template
			templateName, _ := cmd.Flags().GetString("template")
			template, err := config.GetTemplate(conf, templateName, providerName)
			if err != nil {
				log.Fatal().Err(err).Str("name", templateName).Msg("failed to read template")
			}

			// provider
			p, err := provider.GetProviderByName(conf, providerName)
			if err != nil {
				log.Fatal().Err(err).Str("provider", providerName).Msg("failed to get provider")
			}
			options, err := p.OptionsOrCache(float64(flags.maxCacheAge))
			if err != nil {
				log.Fatal().Err(err).Str("provider", p.Name()).Msg("failed to get options")
			}

			// custom output mode for external finder
			if flags.mode != "" {
				err = extensions.OptionsForFinder(flags.mode, options)
				if err != nil {
					log.Fatal().Err(err).Str("mode", flags.mode).Msg("failed to render options")
				}
				return
			}

			// fuzzy finder or direct selection
			var selected *provider.Option
			if flags.selected == "" {
				selected, err = provider.FuzzyFinder(options)
				if err != nil {
					log.Fatal().Err(err).Msg("failed to get selected option")
				}
			} else {
				for _, o := range options {
					if o.Id == flags.selected {
						selected = &o
						break
					}
				}
			}
			log.Debug().Str("display-name", selected.DisplayName).Str("name", selected.Name).Str("directory", selected.StartDirectory).Interface("context", selected.Context).Msg("selected item")

			// create session or window and attach
			err = gotmuxutil.Run(selected, gotmuxutil.Opts{
				SessionName: selected.Name,
				Windows:     template,
				AppendMode:  gotmuxutil.CreateOrAttachSession,
				BaseIndex:   conf.TMUXBaseIndex,
			})
			if err != nil {
				log.Fatal().Err(err).Msg("failed to modify tmux state")
			}
		},
	}

	cmd.PersistentFlags().StringVarP(&flags.template, "template", "t", "", "template to create the tmux session")
	cmd.PersistentFlags().StringVar(&flags.mode, "mode", "", "return data in custom format to use an external fuzzy finder (valid: telescope)")
	cmd.PersistentFlags().StringVar(&flags.selected, "select", "", "skips the finder and directly selects the given id")
	cmd.PersistentFlags().IntVar(&flags.maxCacheAge, "cache-age", 300, "maximum age of the cache in seconds")

	return cmd
}
