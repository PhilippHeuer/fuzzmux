package backend

import (
	"fmt"
)

func ChooseBackend(backend string) (Provider, error) {
	// tmux
	tmux := TMUX{}
	if tmux.Check() || backend == "tmux" {
		return tmux, nil
	}

	// sway
	sway := SWAY{}
	if sway.Check() || backend == "sway" {
		return sway, nil
	}

	// i3
	i3 := I3{}
	if i3.Check() || backend == "i3" {
		return i3, nil
	}

	// shell (fallback, exec in current shell)
	simple := Shell{}
	if simple.Check() || backend == "shell" {
		return simple, nil
	}

	return nil, fmt.Errorf("no available backend found")
}
