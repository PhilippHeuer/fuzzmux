package backend

import (
	"fmt"
)

func ChooseBackend(backend string) (Provider, error) {
	// tmux
	tmux := TMUX{}
	if tmux.Check() {
		return tmux, nil
	}

	// sway
	sway := SWAY{}
	if sway.Check() {
		return sway, nil
	}

	return nil, fmt.Errorf("no available backend found")
}
