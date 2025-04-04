package tmux

import (
	"fmt"
	"github.com/PhilippHeuer/fuzzmux/pkg/launcher"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon"
	"os"
	"os/user"
	"strconv"

	"github.com/PhilippHeuer/fuzzmux/pkg/config"
	gotmux "github.com/jubnzv/go-tmux"
	"github.com/rs/zerolog/log"
)

var server = new(gotmux.Server)

var tmuxBaseIndex = 1

type TMUX struct {
}

func (p TMUX) Name() string {
	return "tmux"
}

func (p TMUX) Check() bool {
	if _, ok := os.LookupEnv("TMUX"); ok {
		return true
	}

	// tmux server running?
	_, _, err := gotmux.RunCmd([]string{"list-sessions"})
	if err == nil {
		return true
	}

	return false
}

func (p TMUX) Order() int {
	if _, ok := os.LookupEnv("TMUX"); ok {
		return 1000
	}

	return 100
}

func (p TMUX) Run(option *recon.Option, opts launcher.Opts) error {
	// references
	var session *gotmux.Session
	var windowCommands = make(map[string][]string)
	var defaultWindowId = tmuxBaseIndex

	// resolve vars
	startDirectory := option.ResolveStartDirectory(true)

	// session lookup
	session, err := FindSession(opts.SessionName)
	if err != nil {
		return fmt.Errorf("failed to find session: %w", err)
	}
	log.Debug().Interface("session", session).Str("search-key", opts.SessionName).Msg("session search result")

	// for CreateOrAttachSession: attach to existing session
	if opts.AppendMode == launcher.CreateOrAttachSession && session != nil {
		err = session.AttachSession()
		if err != nil {
			return fmt.Errorf("failed to attach to existing session %s [%d]: %w", session.Name, session.Id, err)
		}
		return nil
	}

	// create session if it doesn't exist
	if session == nil {
		windows, windowIds := applyWindows([]gotmux.Window{}, opts.Layout.Apps, tmuxBaseIndex, startDirectory)
		session = &gotmux.Session{
			Name:           opts.SessionName,
			StartDirectory: startDirectory,
			Windows:        windows,
		}

		// apply to tmux server
		tmuxConfiguration := gotmux.Configuration{
			Server: server,
			Sessions: []*gotmux.Session{
				session,
			},
			ActiveSession: nil,
		}
		err = tmuxConfiguration.Apply()
		if err != nil {
			return fmt.Errorf("failed to apply configuration to tmux: %w", err)
		}

		windowCommands = make(map[string][]string)

		// exec commands
		for i, w := range opts.Layout.Apps {
			if len(w.Commands) > 0 {
				windowCommands[strconv.Itoa(windowIds[i])] = config.CommandsAsStringSlice(w.Commands)
			}
			if w.Default {
				defaultWindowId = windowIds[i]
			}
		}
	}

	// exec commands
	for id, commands := range windowCommands {
		log.Debug().Str("session-name", session.Name).Str("window-id", id).Interface("commands", commands).Msg("executing commands")
		panes, err := gotmux.ListPanes([]string{"-t", fmt.Sprintf("%s:%s", opts.SessionName, id)})
		if err != nil {
			return fmt.Errorf("failed to list panes: %w", err)
		}

		for _, pane := range panes {
			for _, command := range commands {
				err = pane.RunCommand(option.ResolvePlaceholders(command))
				if err != nil {
					return fmt.Errorf("failed to run command: %w", err)
				}
			}
		}
	}

	// select active window
	log.Debug().Str("session-name", session.Name).Int("window-id", defaultWindowId).Msg("selecting active window")
	_, _, err = gotmux.RunCmd([]string{"select-window", "-t", fmt.Sprintf("%s:%d", session.Name, defaultWindowId)})
	if err != nil {
		return fmt.Errorf("failed to set active window: %w", err)
	}

	// attach
	log.Debug().Str("session-name", session.Name).Int("session-id", session.Id).Msg("attaching to session")
	err = session.AttachSession()
	if err != nil {
		return fmt.Errorf("failed to attach to existing session %s [%d]: %w", session.Name, session.Id, err)
	}

	return nil
}

// applyWindows will add missing windows to the session
func applyWindows(windows []gotmux.Window, add []config.App, baseIndex int, startDirectory string) ([]gotmux.Window, []int) {
	var windowIds []int

	// create windows if none exist
	if len(windows) == 0 {
		for i, w := range add {
			windows = append(windows, gotmux.Window{
				Name:           w.Name,
				Id:             i + baseIndex,
				StartDirectory: startDirectory,
			})
			windowIds = append(windowIds, i+baseIndex)
		}
	}

	// add missing windows
	for _, w := range add {
		found := false
		for _, window := range windows {
			if window.Name == w.Name {
				found = true
				break
			}
		}

		if !found {
			startDirectoryOrHome := startDirectory
			if _, err := os.Stat(startDirectoryOrHome); os.IsNotExist(err) {
				usr, err := user.Current()
				if err != nil {
					log.Fatal().Err(err).Msg("failed to get current user")
				}
				startDirectoryOrHome = usr.HomeDir
			}

			windows = append(windows, gotmux.Window{
				Name:           w.Name,
				Id:             len(windows) + baseIndex,
				StartDirectory: startDirectoryOrHome,
			})
			windowIds = append(windowIds, len(windows)+baseIndex)
		}
	}

	return windows, windowIds
}

// ListPanes finds a window by id
func ListPanes(window gotmux.Window) ([]gotmux.Pane, error) {
	return gotmux.ListPanes([]string{"-t", strconv.Itoa(window.Id)})
}

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
