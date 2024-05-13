package provider

import (
	"fmt"

	"github.com/PhilippHeuer/fuzzmux/pkg/config"
)

type StaticProvider struct {
	StaticOptions []config.StaticOption
}

func (p StaticProvider) Name() string {
	return "static"
}

func (p StaticProvider) Options() ([]Option, error) {
	var options []Option

	for _, staticOption := range p.StaticOptions {
		options = append(options, Option{
			ProviderName:   p.Name(),
			Id:             staticOption.Id,
			DisplayName:    staticOption.DisplayName,
			Name:           staticOption.Name,
			StartDirectory: staticOption.StartDirectory,
			Tags:           staticOption.Tags,
			Context: map[string]string{
				"preview": staticOption.Preview,
				"layout":  staticOption.Layout,
			},
		})
	}

	return options, nil
}

func (p StaticProvider) OptionsOrCache(maxAge float64) ([]Option, error) {
	options, err := p.Options()
	if err != nil {
		return nil, fmt.Errorf("failed to get options: %w", err)
	}

	return options, nil
}

func (p StaticProvider) SelectOption(option *Option) error {
	if option.Context["preview"] != "" {
		fmt.Print(option.Context["preview"])
		return nil
	}

	return nil
}
