package gotmuxutil

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/PhilippHeuer/tmux-tms/pkg/config"
	"github.com/PhilippHeuer/tmux-tms/pkg/provider"
	gotmux "github.com/jubnzv/go-tmux"
	"github.com/rs/zerolog/log"
)

var server = new(gotmux.Server)

type Opts struct {
	SessionName string
	Windows     []config.Window
	AppendMode  AppendMode
	BaseIndex   int
}

type AppendMode string

const (
	CreateOrAttachSession AppendMode = "session"
	//CreateOrAttachWindow  AppendMode = "rel"
)

// Run will take the selected option from the provider and apply changes to tmux, taking into account the config for the template.
func Run(option *provider.Option, opts Opts) error {
	// references
	var session *gotmux.Session
	var windowCommands = make(map[string][]string)
	var defaultWindowId = opts.BaseIndex

	// resolve vars
	startDirectory := os.ExpandEnv(option.StartDirectory)

	// session lookup
	session, err := FindSession(opts.SessionName)
	if err != nil {
		return fmt.Errorf("failed to find session: %w", err)
	}
	log.Debug().Interface("session", session).Str("search-key", opts.SessionName).Msg("session search result")

	// for CreateOrAttachSession: attach to existing session
	if opts.AppendMode == CreateOrAttachSession && session != nil {
		err = session.AttachSession()
		if err != nil {
			return fmt.Errorf("failed to attach to existing session %s [%d]: %w", session.Name, session.Id, err)
		}
		return nil
	}

	// create session if it doesn't exist
	if session == nil {
		windows, windowIds := applyWindows([]gotmux.Window{}, opts.Windows, opts.BaseIndex, startDirectory)
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
		for i, w := range opts.Windows {
			if len(w.Commands) > 0 {
				windowCommands[strconv.Itoa(windowIds[i])] = w.Commands
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
				err = pane.RunCommand(expandCommand(option, command))
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
func applyWindows(windows []gotmux.Window, add []config.Window, baseIndex int, startDirectory string) ([]gotmux.Window, []int) {
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
			windows = append(windows, gotmux.Window{
				Name:           w.Name,
				Id:             len(windows) + baseIndex,
				StartDirectory: startDirectory,
			})
			windowIds = append(windowIds, len(windows)+baseIndex)
		}
	}

	return windows, windowIds
}

func expandCommand(option *provider.Option, command string) string {
	command = expandPlaceholders(command, "name", option.Name)
	command = expandPlaceholders(command, "displayName", option.DisplayName)
	command = expandPlaceholders(command, "startDirectory", option.StartDirectory)
	for k, v := range option.Context {
		command = expandPlaceholders(command, k, v)
	}

	return command
}

func expandPlaceholders(command string, key string, value string) string {
	command = strings.ReplaceAll(command, "$"+key, value)
	command = strings.ReplaceAll(command, "${"+key+"}", value)
	command = strings.ReplaceAll(command, "$"+strings.ToUpper(key), value)
	command = strings.ReplaceAll(command, "${"+strings.ToUpper(key)+"}", value)
	command = strings.ReplaceAll(command, "$"+strings.ToLower(key), value)
	command = strings.ReplaceAll(command, "${"+strings.ToLower(key)+"}", value)

	return command
}
