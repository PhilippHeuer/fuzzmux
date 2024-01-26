package finder

import (
	"os/exec"

	"github.com/PhilippHeuer/fuzzmux/pkg/provider"
)

// FuzzyFinder uses the best available fuzzy finder
func FuzzyFinder(finder string, options []provider.Option) (*provider.Option, error) {
	// user specified?
	if finder == "fzf" {
		return FuzzyFinderFZF(options)
	} else if finder == "embedded" {
		return FuzzyFinderEmbedded(options)
	}

	// choose best available option
	_, err := exec.LookPath("fzf")
	if err == nil {
		return FuzzyFinderFZF(options)
	}

	return FuzzyFinderEmbedded(options)
}
