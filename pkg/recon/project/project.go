package project

import (
	"fmt"
	"github.com/PhilippHeuer/fuzzmux/pkg/config"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon"
	"github.com/PhilippHeuer/fuzzmux/pkg/util"
	"github.com/cidverse/repoanalyzer/analyzer"
)

const moduleName = "project"

var defaultChecks = []string{".git", ".gitignore", ".hg", ".hgignore", ".svn", ".vscode", ".idea"}

type Module struct {
	Config config.ProjectModuleConfig
}

func (p Module) Name() string {
	if p.Config.Name != "" {
		return p.Config.Name
	}
	return moduleName
}

func (p Module) Type() string {
	return moduleName
}

func (p Module) Options() ([]recon.Option, error) {
	var options []recon.Option

	// search for projects
	projects, err := SearchProjectDirectories(p.Config.SourceDirectories, p.Config.Checks)
	if err != nil {
		return options, fmt.Errorf("failed to scan for projects: %w", err)
	}

	for _, project := range projects {
		options = append(options, recon.Option{
			ProviderName:   p.Name(),
			ProviderType:   p.Type(),
			Id:             project.Path,
			DisplayName:    renderProjectDisplayName(project, p.Config.DisplayFormat),
			Name:           project.Name,
			StartDirectory: project.Path,
			Tags:           project.Tags,
		})
	}

	return options, nil
}

func (p Module) OptionsOrCache(maxAge float64) ([]recon.Option, error) {
	return recon.OptionsOrCache(p, maxAge)
}

func (p Module) SelectOption(option *recon.Option) error {
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

func (p Module) Columns() []recon.Column {
	return append(recon.DefaultColumns(),
		recon.Column{Key: "directory", Name: "Directory"},
	)
}

func NewModule(config config.ProjectModuleConfig) Module {
	if config.Checks == nil || len(config.Checks) == 0 {
		config.Checks = defaultChecks
	}

	return Module{
		Config: config,
	}
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
