package project

import (
	"fmt"
	"github.com/PhilippHeuer/fuzzmux/pkg/config"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon"
	"github.com/PhilippHeuer/fuzzmux/pkg/util"
	"github.com/cidverse/repoanalyzer/analyzer"
	"github.com/rs/zerolog/log"
)

type ProjectProvider struct {
	Checks            []string
	SourceDirectories []config.SourceDirectory
	DisplayFormat     config.ProjectDisplayFormat
}

func (p ProjectProvider) Name() string {
	return "project"
}

func (p ProjectProvider) Options() ([]recon.Option, error) {
	var options []recon.Option

	// search for projects
	projects, err := SearchProjectDirectories(p.SourceDirectories, p.Checks)
	if err != nil {
		return options, fmt.Errorf("failed to scan for projects: %w", err)
	}

	for _, project := range projects {
		options = append(options, recon.Option{
			ProviderName:   p.Name(),
			Id:             project.Path,
			DisplayName:    renderProjectDisplayName(project, p.DisplayFormat),
			Name:           project.Name,
			StartDirectory: project.Path,
			Tags:           project.Tags,
		})
	}

	return options, nil
}

func (p ProjectProvider) OptionsOrCache(maxAge float64) ([]recon.Option, error) {
	options, err := recon.LoadOptions(p.Name(), maxAge)
	if err == nil {
		return options, nil
	}

	options, err = p.Options()
	if err != nil {
		return nil, fmt.Errorf("failed to get options: %w", err)
	}

	err = recon.SaveOptions(p.Name(), options)
	if err != nil {
		log.Warn().Err(err).Msg("failed to save options to cache")
	}

	return options, nil
}

func (p ProjectProvider) SelectOption(option *recon.Option) error {
	// run repo analyzer
	modules := analyzer.ScanDirectory(option.StartDirectory)
	for _, m := range modules {
		// languages
		for k := range m.Language {
			option.Tags = util.AddToSet(option.Tags, "language-"+string(k))
		}

		// build system
		option.Tags = util.AddToSet(option.Tags, "buildsystem-"+string(m.BuildSystem))
	}

	return nil
}

func renderProjectDisplayName(project Project, displayFormat config.ProjectDisplayFormat) string {
	output := project.Name
	if displayFormat == config.AbsolutePath {
		output = project.Path
	} else if displayFormat == config.RelativePath {
		output = project.RelativePath
	}

	return output
}
