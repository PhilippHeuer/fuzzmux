package finder

import (
	"fmt"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon"
	"os"
	"os/exec"
	"strings"

	"github.com/PhilippHeuer/fuzzmux/pkg/config"
)

// FuzzyFinderFZF uses fzf to find the selected option
func FuzzyFinderFZF(options []recon.Option, cfg config.FinderConfig) (recon.Option, error) {
	// write options to file
	var builder strings.Builder
	for _, option := range options {
		builder.WriteString(fmt.Sprintf("%s%s%s\n", option.Id, cfg.FZFDelimiter, option.DisplayName))
	}
	optionFile, err := os.CreateTemp("/tmp", "tms-fzf")
	if err != nil {
		return recon.Option{}, fmt.Errorf("failed to create temp file for options: %w", err)
	}
	defer os.Remove(optionFile.Name())
	_, err = optionFile.WriteString(builder.String())
	if err != nil {
		return recon.Option{}, fmt.Errorf("failed to write options to file: %w", err)
	}

	// get executable path
	executablePath, err := os.Executable()
	if err != nil {
		return recon.Option{}, fmt.Errorf("failed to get executable path: %w", err)
	}

	// highlight
	previewCmd := ""
	if cfg.Preview {
		highlightCmd := ""
		if _, err := exec.LookPath("bat"); err == nil {
			highlightCmd = " | bat --color=always -l markdown --style=plain"
		}
		previewCmd = fmt.Sprintf("--preview=\"%s preview \"{}\"%s\"", executablePath, highlightCmd)
	}

	// run fzf (capture output as var)
	cmd := exec.Command("bash", "-c", fmt.Sprintf("cat %s | fzf -d %q --with-nth=2 %s", optionFile.Name(), cfg.FZFDelimiter, previewCmd))
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	// execute command
	out, err := cmd.Output()
	if err != nil {
		return recon.Option{}, fmt.Errorf("failed to run fzf: %w", err)
	}
	optionId := strings.Split(string(out), cfg.FZFDelimiter)
	if len(optionId) == 0 {
		return recon.Option{}, fmt.Errorf("failed to parse fzf output")
	}

	// find option
	for _, option := range options {
		if option.Id == optionId[0] {
			return option, nil
		}
	}

	return recon.Option{}, fmt.Errorf("failed to find option")
}
