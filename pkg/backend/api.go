package backend

import (
	"fmt"

	"github.com/PhilippHeuer/fuzzmux/pkg/config"
	"github.com/PhilippHeuer/fuzzmux/pkg/provider"
)

type Opts struct {
	SessionName string
	Layout      config.Layout
	AppendMode  AppendMode
}

type AppendMode string

const (
	CreateOrAttachSession AppendMode = "session"
)

type Provider interface {
	Name() string
	Check() bool
	Order() int
	Run(option *provider.Option, opts Opts) error
}

func getTerminalCommand(term string, startDirectory string, script string) (string, error) {
	switch term {
	case "alacritty":
		return fmt.Sprintf("alacritty --working-directory %q -e bash -c %q", startDirectory, script), nil
	case "kitty":
		return fmt.Sprintf("kitty -d %q bash -c %q", startDirectory, script), nil
	case "xterm-kitty":
		return fmt.Sprintf("kitty -d %q bash -c %q", startDirectory, script), nil
	default:
		return "", fmt.Errorf("unsupported terminal: %s", term)
	}
}
