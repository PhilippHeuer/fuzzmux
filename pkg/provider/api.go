package provider

import (
	"fmt"
	"slices"

	"github.com/PhilippHeuer/tmux-tms/pkg/config"
	"github.com/ktr0731/go-fuzzyfinder"
)

type Option struct {
	ProviderName   string            `json:"provider_name"`   // provider name
	Id             string            `json:"id"`              // unique id
	DisplayName    string            `json:"display_name"`    // display name for the fuzzy finder
	Name           string            `json:"name"`            // name
	StartDirectory string            `json:"start_directory"` // sets the initial working directory
	Context        map[string]string `json:"context"`         // additional context information
}

type Provider interface {
	Name() string
	Options() ([]Option, error)
	OptionsOrCache(maxAge float64) ([]Option, error)
}

func GetProviders(config config.Config) []Provider {
	var providers []Provider

	providers = append(providers, ProjectProvider{
		SourceDirectories: config.ProjectProvider.SourceDirectories,
		Checks:            config.ProjectProvider.Checks,
	})
	providers = append(providers, SSHProvider{
		// Mode: config.SSHProviderConfig.Mode,
	})

	return providers
}

func GetProviderByName(config config.Config, name string) (Provider, error) {
	for _, p := range GetProviders(config) {
		if p.Name() == name {
			return p, nil
		}
	}

	return nil, fmt.Errorf("provider '%s' not found", name)
}

func GetProvidersByName(config config.Config, names []string) []Provider {
	var providers []Provider

	for _, p := range GetProviders(config) {
		if slices.Contains(names, p.Name()) {
			providers = append(providers, p)
		}
	}

	return providers
}

func FuzzyFinder(options []Option) (*Option, error) {
	idx, err := fuzzyfinder.Find(
		options,
		func(i int) string {
			return options[i].DisplayName
		},
		/*
			fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
				if i == -1 {
					return ""
				}

				return fmt.Sprintf("%s\n\n%s", options[i].DisplayName, options[i].StartDirectory)
			}),
		*/
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find option: %w", err)
	}

	return &options[idx], nil
}
