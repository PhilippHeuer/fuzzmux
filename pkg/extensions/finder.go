package extensions

import (
	"encoding/json"
	"fmt"

	"github.com/PhilippHeuer/fuzzmux/pkg/provider"
)

func OptionsForFinder(mode string, options []provider.Option) error {
	if mode == "telescope" {
		return telescopeOptions(options)
	}

	return fmt.Errorf("mode [%s] is not implemented", mode)
}

type TelescopeOption struct {
	Ordinal string `json:"ordinal"`
	Display string `json:"display"`
	Value   string `json:"value"`
}

func telescopeOptions(options []provider.Option) error {
	var telescopeOptions []TelescopeOption

	for _, option := range options {
		telescopeOptions = append(telescopeOptions, TelescopeOption{
			Ordinal: option.Id,
			Display: option.DisplayName,
			Value:   option.Id,
		})
	}

	// output json
	bytes, err := json.Marshal(telescopeOptions)
	if err != nil {
		return fmt.Errorf("failed to marshal options: %w", err)
	}

	// print json
	fmt.Println(string(bytes))

	return nil
}
