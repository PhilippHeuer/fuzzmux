package cmd

import (
	"errors"
	"github.com/PhilippHeuer/fuzzmux/pkg/app"
	"github.com/PhilippHeuer/fuzzmux/pkg/launcher"
	"github.com/PhilippHeuer/fuzzmux/pkg/layout"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon"
	"os"
	"strings"

	"github.com/PhilippHeuer/fuzzmux/pkg/config"
	"github.com/PhilippHeuer/fuzzmux/pkg/extensions"
	"github.com/PhilippHeuer/fuzzmux/pkg/finder"
	"github.com/PhilippHeuer/fuzzmux/pkg/types"
	"github.com/cidverse/cidverseutils/zerologconfig"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var cfg zerologconfig.LogConfig

type RootFlags struct {
	backend     string
	template    string
	mode        string
	selected    string
	maxCacheAge int
	showTags    []string
	hideTags    []string
}

func rootCmd() *cobra.Command {
	flags := RootFlags{}

	cmd := &cobra.Command{
		Use:   `tmx`,
		Short: `scans source directories for projects to create tmux sessions`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			zerologconfig.Configure(cfg)
		},
		Args: cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			// load config
			conf, err := config.ResolvedConfig()
			if err != nil {
				log.Fatal().Err(err).Msg("failed to load configuration")
			}

			// fuzzy finder
			selected, err := optionFuzzyFinder(conf, args, flags)

			// layout
			defaultLayout := selected.ProviderName
			if selected.Context["layout"] != "" {
				defaultLayout = selected.Context["layout"]
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

	cmd.PersistentFlags().StringVar(&cfg.LogLevel, "log-level", "info", "log level - allowed: "+strings.Join(zerologconfig.ValidLogLevels, ","))
	cmd.PersistentFlags().StringVar(&cfg.LogFormat, "log-format", "color", "log format - allowed: "+strings.Join(zerologconfig.ValidLogFormats, ","))
	cmd.PersistentFlags().BoolVar(&cfg.LogCaller, "log-caller", false, "include caller in log functions")

	cmd.PersistentFlags().StringVar(&flags.backend, "launcher", "", "specify the launcher to use, auto-detected if not set (valid: tmux, hyprland, sway, i3)")
	cmd.PersistentFlags().StringVarP(&flags.template, "template", "t", "", "template to create the tmux session")
	cmd.PersistentFlags().StringVar(&flags.mode, "mode", "", "return data in custom format to use an external fuzzy finder (valid: telescope)")
	cmd.PersistentFlags().StringVar(&flags.selected, "select", "", "skips the finder and directly selects the given id")
	cmd.PersistentFlags().IntVar(&flags.maxCacheAge, "cache-age", 300, "maximum age of the cache in seconds")
	cmd.PersistentFlags().StringSliceVar(&flags.showTags, "show-tags", []string{}, "only show elements with the given tags, all others will be hidden")
	cmd.PersistentFlags().StringSliceVar(&flags.hideTags, "hide-tags", []string{}, "tags to hide from the fuzzy finder")

	cmd.AddCommand(menuCmd())
	cmd.AddCommand(previewCmd())
	cmd.AddCommand(exportCmd())
	cmd.AddCommand(versionCmd())
	cmd.AddCommand(killCmd())
	cmd.AddCommand(killAllCmd())

	return cmd
}

// Execute executes the root command.
func Execute() error {
	return rootCmd().Execute()
}

func optionFuzzyFinder(conf config.Config, args []string, flags RootFlags) (recon.Option, error) {
	// collect options
	modules, options := app.GatherReconOptions(conf, args, flags.showTags, flags.hideTags, flags.maxCacheAge)
	if len(options) == 0 {
		log.Fatal().Msg("no options found")
	}
	if len(options) == 0 {
		return recon.Option{}, types.ErrNoOptionsAvailable
	}

	// custom output mode for external finder
	if flags.mode != "" {
		err := extensions.OptionsForFinder(flags.mode, options)
		if err != nil {
			return recon.Option{}, errors.Join(types.ErrFailedToRenderOptions, err)
		}
		os.Exit(0) // exit after rendering options for external tools, TODO: move this somewhere else
	}

	// fuzzy finder or direct selection
	var selected recon.Option
	if flags.selected == "" {
		s, err := finder.FuzzyFinder(options, *conf.Finder)
		if err != nil {
			return recon.Option{}, errors.Join(types.ErrNoOptionSelected, err)
		}
		selected = s
	} else {
		for _, o := range options {
			if o.Id == flags.selected {
				selected = o
				break
			}
		}
	}
	log.Debug().Str("display-name", selected.DisplayName).Str("name", selected.Name).Str("directory", selected.StartDirectory).Interface("context", selected.Context).Msg("selected item")

	// call select
	selectedProvider, err := app.FindReconModuleByName(modules, selected.ProviderName)
	if err != nil {
		log.Fatal().Err(err).Str("recon", selected.ProviderName).Msg("failed to get recon of selected item")
	}
	err = selectedProvider.SelectOption(&selected)
	if err != nil {
		log.Fatal().Err(err).Str("recon", selected.ProviderName).Msg("failed to run select")
	}

	return selected, nil
}
