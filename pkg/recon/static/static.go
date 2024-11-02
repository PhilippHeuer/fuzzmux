package static

import (
	"fmt"
	"github.com/PhilippHeuer/fuzzmux/pkg/config"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon"
)

const StaticProviderName = "static"

type StaticProvider struct {
	StaticOptions []config.StaticOption
}

func (p StaticProvider) Name() string {
	return StaticProviderName
}

func (p StaticProvider) Options() ([]recon.Option, error) {
	var options []recon.Option

	for _, staticOption := range p.StaticOptions {
		op := recon.Option{
			ProviderName:   p.Name(),
			Id:             staticOption.Id,
			DisplayName:    staticOption.DisplayName,
			Name:           staticOption.Name,
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

func (p StaticProvider) OptionsOrCache(maxAge float64) ([]recon.Option, error) {
	options, err := p.Options()
	if err != nil {
		return nil, fmt.Errorf("failed to get options: %w", err)
	}

	return options, nil
}

func (p StaticProvider) SelectOption(option *recon.Option) error {
	if option.Context["preview"] != "" {
		fmt.Print(option.Context["preview"])
		return nil
	}

	return nil
}
