package gnome

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/PhilippHeuer/fuzzmux/pkg/launcher"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon"
	"github.com/rs/zerolog/log"
)

type GNOME struct{}

func (p GNOME) Name() string {
	return "gnome"
}

func (p GNOME) Check() bool {
	desktop := os.Getenv("XDG_CURRENT_DESKTOP")
	if strings.Contains(strings.ToLower(desktop), "gnome") {
		return true
	}

	_, err := exec.LookPath("gnome-shell")
	return err == nil
}

func (p GNOME) Order() int {
	return 99
}

func (p GNOME) Run(option *recon.Option, opts launcher.Opts) error {
	startDirectory := option.ResolveStartDirectory(true)

	if opts.Layout.ClearWorkspace {
		if err := clearWorkspace(); err != nil {
			log.Warn().Err(err).Msg("failed to clear workspace")
		}
	}

	terminal, err := findTerminal()
	if err != nil {
		log.Fatal().Err(err).Msg("no suitable terminal found")
	}
	log.Debug().Str("terminal", terminal).Msg("selected terminal emulator")

	for _, app := range opts.Layout.Apps {
		log.Debug().Str("name", app.Name).Msg("starting app")

		if len(app.Commands) == 0 {
			continue
		}

		script := strings.Builder{}
		for _, cmd := range app.Commands {
			script.WriteString(option.ResolvePlaceholders(cmd.Command))
			script.WriteString("; ")
		}
		scriptStr := strings.TrimSuffix(script.String(), "; ")

		if app.GUI && len(app.Commands) == 1 {
			cmd := option.ResolvePlaceholders(app.Commands[0].Command)
			log.Trace().Str("name", app.Name).Str("cmd", cmd).Msg("starting GUI app")
			if err := exec.Command("sh", "-c", fmt.Sprintf("cd %q && %s", startDirectory, cmd)).Start(); err != nil {
				log.Fatal().Err(err).Str("name", app.Name).Msg("failed to start GUI app")
			}
		} else {
			launchCmd, err := buildTerminalCommand(terminal, startDirectory, scriptStr)
			if err != nil {
				log.Fatal().Err(err).Str("name", app.Name).Msg("failed to prepare terminal command")
			}

			log.Trace().Str("name", app.Name).Str("cmd", launchCmd).Msg("starting terminal app")
			if err := exec.Command("sh", "-c", launchCmd).Start(); err != nil {
				log.Fatal().Err(err).Str("name", app.Name).Msg("failed to start terminal app")
			}
		}
	}

	return nil
}

// clearWorkspace closes all windows on the current GNOME virtual desktop via D-Bus.
// Requires: "Workspace D-Bus" GNOME extension (https://github.com/kemallette/ws-dbus)
func clearWorkspace() error {
	// Get active workspace index
	out, err := exec.Command("gdbus", "call", "--session",
		"--dest", "org.gnome.Shell.Extensions.WorkspaceDBus",
		"--object-path", "/org/gnome/Shell/Extensions/WorkspaceDBus",
		"--method", "org.gnome.Shell.Extensions.WorkspaceDBus.GetActive",
	).Output()
	if err != nil {
		return fmt.Errorf("clear-workspace requires the 'Workspace D-Bus' GNOME extension. Install from https://github.com/kemallette/ws-dbus")
	}

	// List windows on active workspace
	out, err = exec.Command("gdbus", "call", "--session",
		"--dest", "org.gnome.Shell.Extensions.WorkspaceDBus",
		"--object-path", "/org/gnome/Shell/Extensions/WorkspaceDBus",
		"--method", "org.gnome.Shell.Extensions.WorkspaceDBus.ListWindows",
		strings.TrimSpace(string(out)),
	).Output()
	if err != nil {
		return fmt.Errorf("failed to list windows: %w", err)
	}

	closed := 0
	for _, winId := range extractJsonIds(string(out)) {
		if closeErr := exec.Command("gdbus", "call", "--session",
			"--dest", "org.gnome.Shell.Extensions.WorkspaceDBus",
			"--object-path", "/org/gnome/Shell/Extensions/WorkspaceDBus",
			"--method", "org.gnome.Shell.Extensions.WorkspaceDBus.Close",
			winId,
		).Run(); closeErr != nil {
			log.Debug().Err(closeErr).Str("window-id", winId).Msg("failed to close window")
		} else {
			closed++
		}
	}

	log.Debug().Int("closed", closed).Msg("cleared workspace via D-Bus")
	return nil
}

// extractJsonIds extracts numeric "id" values from a simple JSON array.
func extractJsonIds(jsonStr string) []string {
	var ids []string
	for {
		idx := strings.Index(jsonStr, `"id":`)
		if idx < 0 {
			break
		}
		jsonStr = jsonStr[idx+4:]
		jsonStr = strings.TrimSpace(jsonStr)
		if strings.HasPrefix(jsonStr, ",") {
			jsonStr = jsonStr[1:]
			jsonStr = strings.TrimSpace(jsonStr)
		}
		numStr := ""
		for _, c := range jsonStr {
			if c >= '0' && c <= '9' {
				numStr += string(c)
			} else {
				break
			}
		}
		if numStr != "" {
			ids = append(ids, numStr)
		}
	}
	return ids
}

// findTerminal detects the user's default terminal emulator.
// Priority: gsettings -> .desktop file -> PATH scan.
func findTerminal() (string, error) {
	if term := gsettingsDefaultTerminal(); term != "" {
		return term, nil
	}

	for _, candidate := range knownTerminals {
		if _, err := exec.LookPath(candidate); err == nil {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("no supported terminal emulator found")
}

// gsettingsDefaultTerminal queries the GNOME default terminal setting.
func gsettingsDefaultTerminal() string {
	out, err := exec.Command("gsettings", "get", "org.gnome.desktop.default-applications.terminal", "exec").Output()
	if err != nil {
		return ""
	}

	// output looks like: 'gnome-terminal.desktop'
	raw := strings.TrimSpace(string(out))
	raw = strings.Trim(raw, "'\"")
	raw = strings.TrimSuffix(raw, ".desktop")

	if term, ok := desktopToCommand(raw); ok {
		if _, err := exec.LookPath(term); err == nil {
			return term
		}
	}

	return ""
}

// desktopToCommand maps a .desktop file base name to the executable.
func desktopToCommand(name string) (string, bool) {
	mapping := map[string]string{
		"gnome-terminal":   "gnome-terminal",
		"org.gnome.Terminal": "gnome-terminal",
		"konsole":          "konsole",
		"kitty":            "kitty",
		"alacritty":        "alacritty",
		"foot":             "foot",
		"xfce4-terminal":   "xfce4-terminal",
		"xterm":            "xterm",
		"tilix":            "tilix",
		"terminator":       "terminator",
		"wezterm":          "wezterm",
	}
	cmd, ok := mapping[name]
	return cmd, ok
}

var knownTerminals = []string{
	"gnome-terminal",
	"konsole",
	"kitty",
	"alacritty",
	"foot",
	"tilix",
	"terminator",
	"wezterm",
	"xterm",
}

func buildTerminalCommand(terminal, startDirectory, script string) (string, error) {
	switch terminal {
	case "gnome-terminal":
		return fmt.Sprintf("gnome-terminal --working-directory=%q -- bash -c %q", startDirectory, script), nil
	case "alacritty":
		return fmt.Sprintf("alacritty --working-directory %q -e bash -c %q", startDirectory, script), nil
	case "kitty":
		return fmt.Sprintf("kitty -d %q bash -c %q", startDirectory, script), nil
	case "foot":
		return fmt.Sprintf("foot --working-directory %q -e bash -c %q", startDirectory, script), nil
	case "konsole":
		return fmt.Sprintf("konsole --workdir %q -e bash -c %q", startDirectory, script), nil
	case "xfce4-terminal":
		return fmt.Sprintf("xfce4-terminal --working-directory=%q --command=bash -c %q", startDirectory, script), nil
	case "tilix":
		return fmt.Sprintf("tilix --working-directory=%q -e bash -c %q", startDirectory, script), nil
	case "terminator":
		return fmt.Sprintf("terminator -w %q -x bash -c %q", startDirectory, script), nil
	case "wezterm":
		return fmt.Sprintf("wezterm start --cwd %q bash -c %q", startDirectory, script), nil
	case "xterm":
		return fmt.Sprintf("xterm -e 'cd %q; bash -c %q'", startDirectory, script), nil
	default:
		return "", fmt.Errorf("unsupported terminal: %s", terminal)
	}
}
