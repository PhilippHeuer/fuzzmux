package backend

import (
	"fmt"
	"sort"

	"github.com/rs/zerolog/log"
)

func ChooseBackend(backend string) (Provider, error) {
	// sort by order
	var providerBackends = []Provider{
		TMUX{},
		Hyprland{},
		SWAY{},
		I3{},
		Shell{},
	}
	sort.Slice(providerBackends, func(i, j int) bool {
		return providerBackends[i].Order() > providerBackends[j].Order()
	})

	// select backend by calling check
	for _, p := range providerBackends {
		if backend == "" && p.Check() {
			log.Debug().Str("backend", p.Name()).Msg("selected backend implementation")
			return p, nil
		} else if backend != "" && backend == p.Name() {
			log.Debug().Str("backend", p.Name()).Msg("selected backend implementation")
			return p, nil
		}
	}

	return nil, fmt.Errorf("no available backend found")
}
