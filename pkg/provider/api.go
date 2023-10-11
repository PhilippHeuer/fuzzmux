package provider

import (
	"fmt"
	"slices"
	"strings"

	"github.com/PhilippHeuer/tmux-tms/pkg/config"
	"github.com/ktr0731/go-fuzzyfinder"
)

type Option struct {
	ProviderName   string            `json:"provider_name"`   // provider name
	Id             string            `json:"id"`              // unique id
	DisplayName    string            `json:"display_name"`    // display name for the fuzzy finder
	Name           string            `json:"name"`            // name
	StartDirectory string            `json:"start_directory"` // sets the initial working directory
	Tags           []string          `json:"tags"`            // tags
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
		DisplayFormat:     config.ProjectProvider.DisplayFormat,
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

// FilterOptions filters the options, showTags are required, hideTags
func FilterOptions(options []Option, showTags []string, hideTags []string) []Option {
	var filtered []Option

	for _, o := range options {
		showTagFound := false
		for _, showTag := range showTags {
			if slices.Contains(o.Tags, showTag) {
				showTagFound = true
				break
			}
		}

		hideTagFound := false
		for _, hideTag := range hideTags {
			if slices.Contains(o.Tags, hideTag) {
				hideTagFound = true
				break
			}
		}

		if (showTagFound && !hideTagFound) || (len(showTags) == 0 && !hideTagFound) {
			filtered = append(filtered, o)
		}
	}

	return filtered
}

func FuzzyFinder(options []Option) (*Option, error) {
	idx, err := fuzzyfinder.Find(
		options,
		func(i int) string {
			return options[i].DisplayName
		},
		fuzzyfinder.WithCursorPosition(fuzzyfinder.CursorPositionBottom),
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}

			return fmt.Sprintf("%s\n\nProvider: %s\nDirectory: %s\nTags: %s\n", options[i].DisplayName, options[i].ProviderName, options[i].StartDirectory, strings.Join(options[i].Tags, ", "))
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find option: %w", err)
	}

	return &options[idx], nil
}
