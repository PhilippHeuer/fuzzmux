package finder

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/PhilippHeuer/fuzzmux/pkg/provider"
)

// FuzzyFinderFZF uses fzf to find the selected option
func FuzzyFinderFZF(options []provider.Option) (*provider.Option, error) {
	// write options to file
	var builder strings.Builder
	for _, option := range options {
		builder.WriteString(fmt.Sprintf("%s: %s\n", option.Id, option.DisplayName))
	}
	optionFile, err := os.CreateTemp("/tmp", "tms-fzf")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file for options: %w", err)
	}
	defer os.Remove(optionFile.Name())
	_, err = optionFile.WriteString(builder.String())
	if err != nil {
		return nil, fmt.Errorf("failed to write options to file: %w", err)
	}

	// run fzf (capture output as var)
	cmd := exec.Command("bash", "-c", "cat "+optionFile.Name()+" | fzf --with-nth=2")
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	// execute command
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run fzf: %w", err)
	}
	optionId := strings.Split(string(out), ":")
	if len(optionId) == 0 {
		return nil, fmt.Errorf("failed to parse fzf output")
	}

	// find option
	for _, option := range options {
		if option.Id == optionId[0] {
			return &option, nil
		}
	}

	return nil, fmt.Errorf("failed to find option")
}
