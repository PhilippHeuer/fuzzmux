package provider

import (
	"fmt"

	"github.com/PhilippHeuer/tmux-tms/pkg/config"
	"github.com/PhilippHeuer/tmux-tms/pkg/lookup"
	"github.com/rs/zerolog/log"
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
			Id:             project.Path,
			DisplayName:    project.Name, // TODO: display name with additional information
			Name:           project.Name,
			StartDirectory: project.Path,
		})
	}

	return options, nil
}

func (p ProjectProvider) OptionsOrCache(maxAge float64) ([]Option, error) {
	options, err := LoadOptions(p.Name(), maxAge)
	if err == nil {
		return options, nil
	}

	options, err = p.Options()
	if err != nil {
		return nil, fmt.Errorf("failed to get options: %w", err)
	}

	err = SaveOptions(p.Name(), options)
	if err != nil {
		log.Warn().Err(err).Msg("failed to save options to cache")
	}

	return options, nil
}
