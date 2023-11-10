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
	hideTags = append(hideTags, "hidden") // always hide hidden options, used for e.g. git ssh hosts

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

			var builder strings.Builder
			builder.WriteString(options[i].DisplayName + "\n\n")
			builder.WriteString("Provider: " + options[i].ProviderName + "\n")
			if options[i].StartDirectory != "" {
				builder.WriteString("Directory: " + options[i].StartDirectory + "\n")
			}
			if len(options[i].Tags) > 0 {
				builder.WriteString("Tags: " + strings.Join(options[i].Tags, ", ") + "\n")
			}

			// k8s, openshift
			if options[i].Context["clusterName"] != "" {
				builder.WriteString("K8S Cluster Name: " + options[i].Context["clusterName"] + "\n")
			}
			if options[i].Context["clusterHost"] != "" {
				builder.WriteString("K8S Cluster API: " + options[i].Context["clusterHost"] + "\n")
			}
			if options[i].Context["clusterUser"] != "" {
				builder.WriteString("K8S Cluster User: " + options[i].Context["clusterUser"] + "\n")
			}
			if options[i].Context["clusterType"] != "" {
				builder.WriteString("K8S Cluster Type: " + options[i].Context["clusterType"] + "\n")
			}

			// free-text description
			if options[i].Context["description"] != "" {
				builder.WriteString("\n\n" + options[i].Context["description"] + "\n")
			}

			return builder.String()
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find option: %w", err)
	}

	return &options[idx], nil
}

func addTagToOption(option *Option, tag string) {
	if !slices.Contains(option.Tags, tag) {
		option.Tags = append(option.Tags, tag)
	}
}
