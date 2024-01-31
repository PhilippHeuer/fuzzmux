package finder

import (
	"os/exec"

	"github.com/PhilippHeuer/fuzzmux/pkg/config"
	"github.com/PhilippHeuer/fuzzmux/pkg/provider"
)

// FuzzyFinder uses the best available fuzzy finder
func FuzzyFinder(options []provider.Option, cfg config.FinderConfig) (*provider.Option, error) {
	// user specified?
	if cfg.Executable == "fzf" {
		return FuzzyFinderFZF(options, cfg)
	} else if cfg.Executable == "embedded" {
		return FuzzyFinderEmbedded(options, cfg)
	}

	// choose best available option
	_, err := exec.LookPath("fzf")
	if err == nil {
		return FuzzyFinderFZF(options, cfg)
	}

	return FuzzyFinderEmbedded(options, cfg)
}
