package util

import "fmt"

func GetTerminalCommand(term string, startDirectory string, script string) (string, error) {
	switch term {
	case "alacritty":
		return fmt.Sprintf("alacritty --working-directory %q -e bash -c %q", startDirectory, script), nil
	case "foot":
		return fmt.Sprintf("foot --working-directory %q -e bash -c %q", startDirectory, script), nil
	case "kitty", "xterm-kitty":
		return fmt.Sprintf("kitty -d %q bash -c %q", startDirectory, script), nil
	case "gnome-terminal":
		return fmt.Sprintf("gnome-terminal --working-directory=%q -- bash -c %q", startDirectory, script), nil
	case "xfce4-terminal":
		return fmt.Sprintf("xfce4-terminal --working-directory=%q --command=bash -c %q", startDirectory, script), nil
	case "konsole":
		return fmt.Sprintf("konsole --workdir %q -e bash -c %q", startDirectory, script), nil
	case "xterm":
		return fmt.Sprintf("xterm -e 'cd %q; bash -c %q'", startDirectory, script), nil
	default:
		return "", fmt.Errorf("unsupported terminal: %s", term)
	}
}
