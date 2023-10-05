package cmd

import (
	gotmux "github.com/jubnzv/go-tmux"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func killAllCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "kill-all",
		Aliases: []string{},
		Short:   "Kill all tmux sessions.",
		Run: func(cmd *cobra.Command, args []string) {
			server := new(gotmux.Server)

			// query sessions
			sessions, err := server.ListSessions()
			if err != nil {
				log.Fatal().Err(err).Msg("failed to list sessions")
			}

			// kill sessions
			for _, s := range sessions {
				err := server.KillSession(s.Name)
				if err != nil {
					log.Warn().Int("session_id", s.Id).Str("session_name", s.Name).Msg("failed to kill session")
				}
			}
		},
	}
}
