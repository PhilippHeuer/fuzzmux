package util

import (
	"fmt"
	"strings"
)

// ShellQuote quotes the value for a POSIX shell, the result is always interpreted as a single literal word
func ShellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
}

// GetTerminalCommand returns a shell command that opens the terminal in startDirectory and runs the script with bash
func GetTerminalCommand(term string, startDirectory string, script string) (string, error) {
	dir := ShellQuote(startDirectory)
	bashCmd := "bash -c " + ShellQuote(script)

	switch term {
	case "alacritty":
		return fmt.Sprintf("alacritty --working-directory %s -e %s", dir, bashCmd), nil
	case "foot":
		return fmt.Sprintf("foot --working-directory %s -e %s", dir, bashCmd), nil
	case "kitty", "xterm-kitty":
		return fmt.Sprintf("kitty -d %s %s", dir, bashCmd), nil
	case "gnome-terminal":
		return fmt.Sprintf("gnome-terminal --working-directory=%s -- %s", dir, bashCmd), nil
	case "xfce4-terminal":
		// --command takes a single string that is parsed by the terminal itself
		return fmt.Sprintf("xfce4-terminal --working-directory=%s --command=%s", dir, ShellQuote(bashCmd)), nil
	case "konsole":
		return fmt.Sprintf("konsole --workdir %s -e %s", dir, bashCmd), nil
	case "tilix":
		// --command takes a single string that is parsed by the terminal itself
		return fmt.Sprintf("tilix --working-directory=%s --command=%s", dir, ShellQuote(bashCmd)), nil
	case "terminator":
		return fmt.Sprintf("terminator --working-directory=%s -x %s", dir, bashCmd), nil
	case "wezterm":
		return fmt.Sprintf("wezterm start --cwd %s -- %s", dir, bashCmd), nil
	case "xterm":
		return fmt.Sprintf("xterm -e bash -c %s", ShellQuote("cd "+dir+" && "+bashCmd)), nil
	default:
		return "", fmt.Errorf("unsupported terminal: %s", term)
	}
}
