package gotmuxutil

import (
	"fmt"

	gotmux "github.com/jubnzv/go-tmux"
)

// FindSession finds a session by name
func FindSession(sessionName string) (*gotmux.Session, error) {
	sessions, err := server.ListSessions()
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}
	for _, session := range sessions {
		if session.Name == sessionName {
			return &session, nil
		}
	}

	return nil, nil
}
