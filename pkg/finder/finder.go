package finder

import (
	"github.com/PhilippHeuer/fuzzmux/pkg/recon"
	"os/exec"

	"github.com/PhilippHeuer/fuzzmux/pkg/config"
)

// FuzzyFinder uses the best available fuzzy finder
func FuzzyFinder(options []recon.Option, cfg config.FinderConfig) (recon.Option, error) {
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
