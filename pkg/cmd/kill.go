package cmd

import (
	"slices"

	gotmux "github.com/jubnzv/go-tmux"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func killCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "kill",
		Aliases: []string{"k"},
		Short:   "Kill the current tmux session and jump to another",
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			server := new(gotmux.Server)

			// query sessions
			sessions, err := server.ListSessions()
			if err != nil {
				log.Fatal().Err(err).Msg("failed to list sessions")
			}

			// kill sessions
			for _, s := range sessions {
				if !slices.Contains(args, s.Name) {
					continue
				}

				err := server.KillSession(s.Name)
				if err != nil {
					log.Warn().Int("session_id", s.Id).Str("session_name", s.Name).Msg("failed to kill session")
				}
			}
		},
	}
}
