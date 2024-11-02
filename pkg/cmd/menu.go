package cmd

import (
	"errors"
	"github.com/PhilippHeuer/fuzzmux/pkg/app"
	"github.com/PhilippHeuer/fuzzmux/pkg/launcher"
	"github.com/PhilippHeuer/fuzzmux/pkg/layout"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon/static"
	"slices"

	"github.com/PhilippHeuer/fuzzmux/pkg/config"
	"github.com/PhilippHeuer/fuzzmux/pkg/finder"
	"github.com/PhilippHeuer/fuzzmux/pkg/types"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func menuCmd() *cobra.Command {
	flags := RootFlags{}

	cmd := &cobra.Command{
		Use:   "menu",
		Short: "menu with all available option providers to select from",
		Run: func(cmd *cobra.Command, args []string) {
			// load config
			conf, err := config.ResolvedConfig()
			if err != nil {
				log.Fatal().Err(err).Msg("failed to load configuration")
			}

			// select recon
			selected, err := providerMenuFuzzyFinder(conf, args)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to get selected recon module")
			}
			log.Debug().Str("display-name", selected.DisplayName).Str("name", selected.Name).Str("directory", selected.StartDirectory).Interface("context", selected.Context).Msg("selected recon from menu")

			// select recon options (static options by tag, providers by name)
			if len(selected.Tags) > 0 {
				selected, err = optionFuzzyFinder(conf, []string{static.StaticProviderName}, RootFlags{
					backend:  flags.backend,
					template: flags.template,
					showTags: []string{selected.Id},
				})
			} else {
				selected, err = optionFuzzyFinder(conf, []string{selected.Id}, RootFlags{
					backend:  flags.backend,
					template: flags.template,
				})
			}
			if err != nil {
				log.Fatal().Err(err).Msg("failed to get selected option")
			}
			log.Debug().Str("display-name", selected.DisplayName).Str("name", selected.Name).Str("directory", selected.StartDirectory).Interface("context", selected.Context).Msg("selected option")

			// layout
			defaultLayout := selected.ProviderName
			if selected.Context["layout"] != "" {
				defaultLayout = selected.Context["layout"]
			}
			if defaultLayout == "static" && len(selected.Tags) == 1 {
				defaultLayout = selected.Tags[0]
			}
			log.Debug().Str("default-layout", defaultLayout).Msg("default layout, if template is not specified")

			// template
			templateName, _ := cmd.Flags().GetString("template")
			template, err := layout.GetLayout(conf, &selected, templateName, defaultLayout)
			if err != nil {
				log.Fatal().Err(err).Str("name", templateName).Msg("failed to read template")
			}

			// create session or window and attach
			be, err := app.FindLauncher(flags.backend)
			if err != nil {
				log.Fatal().Err(err).Msg("no suitable launcher found")
			}
			err = be.Run(&selected, launcher.Opts{
				SessionName: selected.Name,
				Layout:      template,
				AppendMode:  launcher.CreateOrAttachSession,
			})
			if err != nil {
				log.Fatal().Err(err).Msg("failed to modify tmux state")
			}
		},
	}

	cmd.PersistentFlags().StringVar(&flags.backend, "launcher", "", "specify the launcher to use, auto-detected if not set (valid: tmux, sway, i3)")
	cmd.PersistentFlags().StringVarP(&flags.template, "template", "t", "", "template to create the tmux session")

	return cmd
}

func providerMenuFuzzyFinder(conf config.Config, filter []string) (recon.Option, error) {
	// collect options from providers
	providers := app.ConfigToReconModules(conf)
	var options []recon.Option
	for _, p := range providers {
		options = append(options, recon.Option{
			ProviderName: p.Name(),
			Id:           p.Name(),
			DisplayName:  p.Name(),
			Name:         p.Name(),
		})

		if p.Name() == static.StaticProviderName {
			opts, err := p.Options()
			if err != nil {
				return recon.Option{}, errors.Join(types.ErrFailedToGetOptionsFromProvider, err)
			}

			var addedProviderNames []string
			for _, o := range opts {
				if len(o.Tags) != 1 {
					continue
				}

				providerName := o.Tags[0]
				if !slices.Contains(addedProviderNames, providerName) {
					options = append(options, recon.Option{
						ProviderName: p.Name(),
						Id:           providerName,
						DisplayName:  providerName,
						Name:         providerName,
						Tags:         []string{providerName},
					})

					addedProviderNames = append(addedProviderNames, providerName)
				}
			}
		}
	}
	if len(filter) > 0 {
		var filteredOptions []recon.Option
		for _, o := range options {
			if slices.Contains(filter, o.ProviderName) || (len(o.Tags) > 0 && slices.Contains(filter, o.Tags[0])) {
				filteredOptions = append(filteredOptions, o)
			}
		}
		options = filteredOptions
	}
	if len(options) == 0 {
		return recon.Option{}, types.ErrNoProvidersAvailable
	}

	// fuzzy finder
	selected, err := finder.FuzzyFinder(options, config.FinderConfig{
		Executable:   conf.Finder.Executable,
		Preview:      false,
		FZFDelimiter: conf.Finder.FZFDelimiter,
	})
	if err != nil {
		return recon.Option{}, errors.Join(types.ErrNoOptionSelected, err)
	}

	return selected, nil
}
