package app

import (
	"github.com/PhilippHeuer/fuzzmux/pkg/launcher"
	"github.com/PhilippHeuer/fuzzmux/pkg/launcher/hyprland"
	"github.com/PhilippHeuer/fuzzmux/pkg/launcher/i3"
	"github.com/PhilippHeuer/fuzzmux/pkg/launcher/shell"
	"github.com/PhilippHeuer/fuzzmux/pkg/launcher/sway"
	"github.com/PhilippHeuer/fuzzmux/pkg/launcher/tmux"
	"github.com/PhilippHeuer/fuzzmux/pkg/types"
	"sort"

	"github.com/rs/zerolog/log"
)

// FindLauncher finds the launcher implementation by name, or returns the first available one if name is empty
func FindLauncher(name string) (launcher.Provider, error) {
	// sort by order
	var appLaunchers = []launcher.Provider{
		tmux.TMUX{},
		hyprland.Hyprland{},
		sway.SWAY{},
		i3.I3{},
		shell.Shell{},
	}
	sort.Slice(appLaunchers, func(i, j int) bool {
		return appLaunchers[i].Order() > appLaunchers[j].Order()
	})

	// select launcher by calling check
	for _, p := range appLaunchers {
		if name == "" && p.Check() {
			log.Debug().Str("launcher", p.Name()).Msg("selected launcher implementation")
			return p, nil
		} else if name != "" && name == p.Name() {
			log.Debug().Str("launcher", p.Name()).Msg("selected launcher implementation")
			return p, nil
		}
	}

	return nil, types.ErrNoLauncherAvailable
}
