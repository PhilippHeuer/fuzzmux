package cmd

import (
	gotmux "github.com/jubnzv/go-tmux"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func switchCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "switch",
		Aliases: []string{"s"},
		Short:   "Display other sessions with a fuzzy finder and a preview window.",
		Run: func(cmd *cobra.Command, args []string) {
			server := new(gotmux.Server)

			// list sessions
			sessions, err := server.ListSessions()
			if err != nil {
				log.Fatal().Err(err).Msg("failed to list sessions")
			}

			// fuzzy finder
			idx, err := fuzzyfinder.Find(
				sessions,
				func(i int) string {
					return sessions[i].Name
				},
			)
			if err != nil {
				log.Err(err).Msg("failed to find session")
			}

			// attach
			err = sessions[idx].AttachSession()
			if err != nil {
				log.Fatal().Err(err).Int("session_id", sessions[idx].Id).Str("session_name", sessions[idx].Name).Msg("failed to attach to session")
			}
		},
	}
}
