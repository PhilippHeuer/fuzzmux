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
	Name() string                                    // Name returns the name of the provider
	Options() ([]Option, error)                      // Options returns the options
	OptionsOrCache(maxAge float64) ([]Option, error) // OptionsOrCache returns the options from cache or calls Options
	SelectOption(options *Option) error              // Select can be used to run actions / enrich the context before opening the session
}

func GetProviders(config config.Config) []Provider {
	var providers []Provider

	if config.ProjectProvider != nil {
		providers = append(providers, ProjectProvider{
			SourceDirectories: config.ProjectProvider.SourceDirectories,
			Checks:            config.ProjectProvider.Checks,
			DisplayFormat:     config.ProjectProvider.DisplayFormat,
		})
	}
	if config.SSHProvider != nil {
		providers = append(providers, SSHProvider{
			// Mode: config.SSHProviderConfig.Mode,
		})
	}
	if config.KubernetesProvider != nil {
		providers = append(providers, KubernetesProvider{
			Clusters: config.KubernetesProvider.Clusters,
		})
	}
	if config.OpenShiftProvider != nil {
		providers = append(providers, OpenShiftProvider{
			Clusters: config.OpenShiftProvider.Clusters,
		})
	}

	return providers
}

func GetProviderByName(providers []Provider, name string) (Provider, error) {
	for _, p := range providers {
		if p.Name() == name {
			return p, nil
		}
	}

	return nil, fmt.Errorf("provider '%s' not found", name)
}

func GetProvidersByName(providers []Provider, names []string) []Provider {
	var result []Provider

	for _, p := range providers {
		if slices.Contains(names, p.Name()) {
			result = append(result, p)
		}
	}

	return result
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
