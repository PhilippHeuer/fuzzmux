package static

import (
	"fmt"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon"
)

const moduleType = "static"

type Module struct {
	Config ModuleConfig
}

type ModuleConfig struct {
	// Name is used to override the default module name
	Name string `yaml:"name,omitempty"`

	// Options is a list of static options
	StaticOptions []StaticOption `yaml:"options"`
}

type StaticOption struct {
	// Id is a unique identifier for the option
	Id string `yaml:"id"`

	// Name is the name of the option
	Name string `yaml:"name"`

	// DisplayName is the name that should be displayed in the fuzzy finder
	DisplayName string `yaml:"display-name"`

	// Description is a short description of the option
	Description string `yaml:"description"`

	// Web is the URL to open when the option is selected
	Web string `yaml:"web"`

	// StartDirectory is the initial working directory
	StartDirectory string `yaml:"start-directory"`

	// Tags can be used to filter options
	Tags []string `yaml:"tags"`

	// Context
	Context map[string]string `yaml:"context"`

	// Layout can be used to override the default layout used by the option (e.g. ssh/project)
	Layout string `yaml:"layout"`

	// Preview to render in the preview window
	Preview string `yaml:"preview"`
}

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
	var options []recon.Option

	for _, staticOption := range p.Config.StaticOptions {
		op := recon.Option{
			ProviderName:   p.Name(),
			ProviderType:   p.Type(),
			Id:             staticOption.Id,
			Name:           staticOption.Name,
			DisplayName:    staticOption.DisplayName,
			Description:    staticOption.Description,
			Web:            staticOption.Web,
			StartDirectory: staticOption.StartDirectory,
			Tags:           staticOption.Tags,
			Context:        staticOption.Context,
		}
		if op.Context == nil {
			op.Context = make(map[string]string)
		}
		if staticOption.Preview != "" {
			op.Context["preview"] = staticOption.Preview
		}
		if staticOption.Layout != "" {
			op.Context["layout"] = staticOption.Layout
		}

		options = append(options, op)
	}

	return options, nil
}

func (p Module) OptionsOrCache(maxAge float64) ([]recon.Option, error) {
	options, err := p.Options()
	if err != nil {
		return nil, fmt.Errorf("failed to get options: %w", err)
	}

	return options, nil
}

func (p Module) SelectOption(option *recon.Option) error {
	if option.Context["preview"] != "" {
		fmt.Print(option.Context["preview"])
		return nil
	}

	return nil
}

func (p Module) Columns() []recon.Column {
	return recon.DefaultColumns()
}

func NewModule(config ModuleConfig) Module {
	return Module{
		Config: config,
	}
}
