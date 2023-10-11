package cmd

import (
	"github.com/PhilippHeuer/tmux-tms/pkg/config"
	"github.com/PhilippHeuer/tmux-tms/pkg/extensions"
	"github.com/PhilippHeuer/tmux-tms/pkg/gotmuxutil"
	"github.com/PhilippHeuer/tmux-tms/pkg/provider"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func sshCmd() *cobra.Command {
	const providerName = "ssh"
	flags := RootFlags{}

	cmd := &cobra.Command{
		Use:     "ssh",
		Aliases: []string{},
		Short:   "Fuzzy search for ssh hosts",
		Run: func(cmd *cobra.Command, args []string) {
			// load config
			conf, err := config.ResolvedConfig()
			if err != nil {
				log.Fatal().Err(err).Msg("failed to load configuration")
			}

			// template
			template, err := config.GetTemplate(conf, flags.template, providerName)
			if err != nil {
				log.Fatal().Err(err).Str("name", flags.template).Msg("failed to read template")
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
	cmd.PersistentFlags().StringSliceVar(&flags.showTags, "show-tags", []string{}, "tags to show in the fuzzy finder, all others will be hidden. Overrides --hide-tags.")
	cmd.PersistentFlags().StringSliceVar(&flags.hideTags, "hide-tags", []string{}, "tags to hide from the fuzzy finder")

	return cmd
}

/*
// sshConnectWindowMode uses a shared session and creates a new window for each connection
func sshConnectWindowMode(host Host) {
	server := new(gotmux.Server)

	// check if session exists
	session, sessionErr := gotmuxutil.FindSession(server, "ssh")
	if sessionErr != nil {
		log.Fatal().Err(sessionErr).Msg("failed to list sessions")
	}

	if session != nil {
		log.Debug().Msg("session exists, extending with new window")

		// create window
		window, err := session.NewWindow(host.Name)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create new window")
		}
		fmt.Printf("window: %+v\n", window)

		// connect
		panes, err := gotmuxutil.ListPanes(window)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to list panes")
		}
		err = panes[0].RunCommand("ssh " + host.Name)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to run ssh command")
		}

		// select
		err = panes[0].Select()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to select pane")
		}

		// attach
		err = session.AttachSession()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to attach to created session")
		}
	} else {
		log.Debug().Msg("session does not exist, creating new session with initial window")

		session = &gotmux.Session{Name: "ssh"}
		window := gotmux.Window{Name: host.Name, Id: 1}
		session.AddWindow(window)

		// apply to tmux server
		tmuxConfiguration := gotmux.Configuration{
			Server: server,
			Sessions: []*gotmux.Session{
				session,
			},
			ActiveSession: nil,
		}
		err := tmuxConfiguration.Apply()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to apply configuration")
		}

		// connect
		panes, err := gotmuxutil.ListPanes(window)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to list panes")
		}
		err = panes[0].RunCommand("ssh " + host.Name)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to run ssh command")
		}

		// attach
		err = session.AttachSession()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to attach to created session")
		}
	}
}
*/
