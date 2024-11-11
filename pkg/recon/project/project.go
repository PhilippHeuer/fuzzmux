package project

import (
	"fmt"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon"
	"github.com/PhilippHeuer/fuzzmux/pkg/util"
	"github.com/cidverse/repoanalyzer/analyzer"
)

const moduleType = "project"

var defaultChecks = []string{".git", ".gitignore", ".hg", ".hgignore", ".svn", ".vscode", ".idea"}

type Module struct {
	Config ModuleConfig
}

type ModuleConfig struct {
	// Name is used to override the default module name
	Name string `yaml:"name,omitempty"`

	// DisplayName is a template string to render a custom display name
	DisplayName string `yaml:"display-name"`

	// StartDirectory is a template string that defines the start directory
	StartDirectory string `yaml:"start-directory"`

	// Sources is a list of source directories that should be scanned
	SourceDirectories []SourceDirectory `yaml:"directories"`

	// Checks is a list of files or directories that should be checked, e.g. ".git", ".gitignore"
	Checks []string `yaml:"checks"`

	// DisplayFormat is the format that should be used to display the project name
	DisplayFormat ProjectDisplayFormat `yaml:"display-format"`
}

type SourceDirectory struct {
	// Directory is the absolute path to the source directory
	Directory string `yaml:"path"`

	// Depth is the maximum depth of subdirectories that should be scanned
	Depth int `yaml:"depth"`

	// Exclude is a list of directories that should be excluded from the scan
	Exclude []string `yaml:"exclude"`

	// Tags can be used to filter directories
	Tags []string `yaml:"tags"`
}

type ProjectDisplayFormat string

const (
	AbsolutePath ProjectDisplayFormat = "absolute"
	RelativePath ProjectDisplayFormat = "relative"
	BaseName     ProjectDisplayFormat = "base"
)

func (p Module) Name() string {
	if p.Config.Name != "" {
		return p.Config.Name
	}
	return moduleType
}

func (p Module) Type() string {
	return moduleType
}

func (p Module) Options() ([]recon.Option, error) {
	var result []recon.Option

	// search for projects
	projects, err := SearchProjectDirectories(p.Config.SourceDirectories, p.Config.Checks)
	if err != nil {
		return result, fmt.Errorf("failed to scan for projects: %w", err)
	}

	for _, project := range projects {
		opt := recon.Option{
			ProviderName:   p.Name(),
			ProviderType:   p.Type(),
			Id:             project.Path,
			DisplayName:    renderProjectDisplayName(project, p.Config.DisplayFormat),
			Name:           project.Name,
			StartDirectory: project.Path,
			Tags:           project.Tags,
		}
		opt.ProcessUserTemplateStrings(p.Config.DisplayName, p.Config.StartDirectory)
		result = append(result, opt)
	}

	return result, nil
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

func NewModule(config ModuleConfig) Module {
	if config.Checks == nil || len(config.Checks) == 0 {
		config.Checks = defaultChecks
	}

	return Module{
		Config: config,
	}
}

func renderProjectDisplayName(project Project, displayFormat ProjectDisplayFormat) string {
	output := project.Name
	if displayFormat == AbsolutePath {
		output = project.Path
	} else if displayFormat == RelativePath {
		output = project.RelativePath
	}

	return output
}
