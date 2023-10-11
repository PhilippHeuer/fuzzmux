package cmd

import (
	"os"
	"slices"
	"strings"

	"github.com/PhilippHeuer/tmux-tms/pkg/config"
	"github.com/PhilippHeuer/tmux-tms/pkg/extensions"
	"github.com/PhilippHeuer/tmux-tms/pkg/gotmuxutil"
	"github.com/PhilippHeuer/tmux-tms/pkg/provider"
	"github.com/mattn/go-colorable"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	cfg = struct {
		LogLevel  string
		LogFormat string
		LogCaller bool
	}{}
	validLogLevels  = []string{"trace", "debug", "info", "warn", "error"}
	validLogFormats = []string{"plain", "color", "json"}
)

type RootFlags struct {
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
		Use:   `tms`,
		Short: `scans source directories for projects to create tmux sessions`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// log format
			if !slices.Contains(validLogFormats, cfg.LogFormat) {
				log.Error().Str("current", cfg.LogFormat).Strs("valid", validLogFormats).Msg("invalid log format specified")
				os.Exit(1)
			}
			var logContext zerolog.Context
			if cfg.LogFormat == "plain" {
				logContext = zerolog.New(os.Stderr).Output(zerolog.ConsoleWriter{Out: os.Stderr, NoColor: true}).With().Timestamp()
			} else if cfg.LogFormat == "color" {
				colorableOutput := colorable.NewColorableStdout()
				logContext = zerolog.New(os.Stderr).Output(zerolog.ConsoleWriter{Out: colorableOutput, NoColor: false}).With().Timestamp()
			} else if cfg.LogFormat == "json" {
				logContext = zerolog.New(os.Stderr).Output(os.Stderr).With().Timestamp()
			}
			if cfg.LogCaller {
				logContext = logContext.Caller()
			}
			log.Logger = logContext.Logger()

			// log time format
			zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

			// log level
			if !slices.Contains(validLogLevels, cfg.LogLevel) {
				log.Error().Str("current", cfg.LogLevel).Strs("valid", validLogLevels).Msg("invalid log level specified")
				os.Exit(1)
			}
			if cfg.LogLevel == "trace" {
				zerolog.SetGlobalLevel(zerolog.TraceLevel)
			} else if cfg.LogLevel == "debug" {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			} else if cfg.LogLevel == "info" {
				zerolog.SetGlobalLevel(zerolog.InfoLevel)
			} else if cfg.LogLevel == "warn" {
				zerolog.SetGlobalLevel(zerolog.WarnLevel)
			} else if cfg.LogLevel == "error" {
				zerolog.SetGlobalLevel(zerolog.ErrorLevel)
			}

			// logging config
			log.Debug().Str("log-level", cfg.LogLevel).Str("log-format", cfg.LogFormat).Bool("log-caller", cfg.LogCaller).Msg("configured logging")
		},
		Run: func(cmd *cobra.Command, args []string) {
			// load config
			conf, err := config.ResolvedConfig()
			if err != nil {
				log.Fatal().Err(err).Msg("failed to load configuration")
			}

			// collect options from providers
			providers := provider.GetProviders(conf)
			var options []provider.Option
			for _, p := range providers {
				opts, err := p.OptionsOrCache(float64(flags.maxCacheAge))
				if err != nil {
					log.Fatal().Err(err).Str("provider", p.Name()).Msg("failed to get options")
				}

				options = append(options, opts...)
			}
			options = provider.FilterOptions(options, flags.showTags, flags.hideTags)

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

			// template
			templateName, _ := cmd.Flags().GetString("template")
			template, err := config.GetTemplate(conf, templateName, selected.ProviderName)
			if err != nil {
				log.Fatal().Err(err).Str("name", templateName).Msg("failed to read template")
			}

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

	cmd.PersistentFlags().StringVar(&cfg.LogLevel, "log-level", "info", "log level - allowed: "+strings.Join(validLogLevels, ","))
	cmd.PersistentFlags().StringVar(&cfg.LogFormat, "log-format", "color", "log format - allowed: "+strings.Join(validLogFormats, ","))
	cmd.PersistentFlags().BoolVar(&cfg.LogCaller, "log-caller", false, "include caller in log functions")

	cmd.PersistentFlags().StringVarP(&flags.template, "template", "t", "", "template to create the tmux session")
	cmd.PersistentFlags().StringVar(&flags.mode, "mode", "", "return data in custom format to use an external fuzzy finder (valid: telescope)")
	cmd.PersistentFlags().StringVar(&flags.selected, "select", "", "skips the finder and directly selects the given id")
	cmd.PersistentFlags().IntVar(&flags.maxCacheAge, "cache-age", 300, "maximum age of the cache in seconds")
	cmd.PersistentFlags().StringSliceVar(&flags.showTags, "show-tags", []string{}, "only show elements with the given tags, all others will be hidden")
	cmd.PersistentFlags().StringSliceVar(&flags.hideTags, "hide-tags", []string{}, "tags to hide from the fuzzy finder")

	cmd.AddCommand(versionCmd())
	cmd.AddCommand(projectCmd())
	cmd.AddCommand(sshCmd())
	cmd.AddCommand(killCmd())
	cmd.AddCommand(killAllCmd())

	return cmd
}

// Execute executes the root command.
func Execute() error {
	return rootCmd().Execute()
}
