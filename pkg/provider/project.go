package provider

import (
	"fmt"

	"github.com/PhilippHeuer/tmux-tms/pkg/config"
	"github.com/PhilippHeuer/tmux-tms/pkg/lookup"
)

type ProjectProvider struct {
	Checks            []string
	SourceDirectories []config.SourceDirectory
}

func (p ProjectProvider) Name() string {
	return "project"
}

func (p ProjectProvider) Options() ([]Option, error) {
	var options []Option

	// search for projects
	projects, err := lookup.ScanForProjects(p.SourceDirectories, p.Checks)
	if err != nil {
		return options, fmt.Errorf("failed to scan for projects: %w", err)
	}

	for _, project := range projects {
		options = append(options, Option{
			ProviderName:   p.Name(),
			DisplayName:    project.Name, // TODO: display name with additional information
			Name:           project.Name,
			StartDirectory: project.Path,
		})
	}

	return options, nil
}
