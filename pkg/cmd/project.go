package cmd

import (
	"github.com/PhilippHeuer/tmux-tms/pkg/config"
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

			// fuzzy finder
			selected, err := provider.FuzzyFinder(options)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to get selected option")
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
	cmd.PersistentFlags().IntVar(&flags.maxCacheAge, "cache-age", 300, "maximum age of the cache in seconds")

	return cmd
}
