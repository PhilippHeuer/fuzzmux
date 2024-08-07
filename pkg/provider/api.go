package provider

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/PhilippHeuer/fuzzmux/pkg/config"
	"github.com/PhilippHeuer/fuzzmux/pkg/core/util"
	"github.com/PhilippHeuer/fuzzmux/pkg/errtypes"
	"github.com/cidverse/cidverseutils/filesystem"
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

func (o Option) ResolveStartDirectory(full bool) string {
	startDirectory := o.StartDirectory
	if startDirectory == "" {
		startDirectory = "~"
	}
	startDirectory = util.ExpandPlaceholders(startDirectory, "name", o.Name)
	startDirectory = util.ExpandPlaceholders(startDirectory, "displayName", o.DisplayName)
	for k, v := range o.Context {
		startDirectory = util.ExpandPlaceholders(startDirectory, k, v)
	}

	if full {
		startDirectory = strings.Replace(startDirectory, "~", os.Getenv("HOME"), -1)
	}

	return startDirectory
}

func (o Option) CreateStartDirectoryIfMissing() error {
	if o.StartDirectory == "" || o.StartDirectory == "~" {
		return nil
	}

	dir := o.ResolveStartDirectory(true)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return errors.Join(ErrFailedToCreateStartDirectory, err)
		}
	}

	return nil
}

func (o Option) ResolvePlaceholders(input string) string {
	input = os.ExpandEnv(input)

	input = util.ExpandPlaceholders(input, "name", o.Name)
	input = util.ExpandPlaceholders(input, "displayName", o.DisplayName)
	input = util.ExpandPlaceholders(input, "startDirectory", o.ResolveStartDirectory(true))

	for k, v := range o.Context {
		input = util.ExpandPlaceholders(input, k, v)
	}

	return input
}

type Provider interface {
	Name() string                                    // Name returns the name of the provider
	Options() ([]Option, error)                      // Options returns the options
	OptionsOrCache(maxAge float64) ([]Option, error) // OptionsOrCache returns the options from cache or calls Options
	SelectOption(options *Option) error              // Select can be used to run actions / enrich the context before opening the session
}

func GetProviders(config config.Config) []Provider {
	var providers []Provider

	// projects
	if config.ProjectProvider != nil && config.ProjectProvider.Enabled {
		providers = append(providers, ProjectProvider{
			SourceDirectories: config.ProjectProvider.SourceDirectories,
			Checks:            config.ProjectProvider.Checks,
			DisplayFormat:     config.ProjectProvider.DisplayFormat,
		})
	}

	// ssh
	if config.SSHProvider != nil && config.SSHProvider.Enabled {
		providers = append(providers, NewSSHProvider(config.SSHProvider.ConfigFile, config.SSHProvider.StartDirectory))
	} else if config.SSHProvider == nil && filesystem.FileExists(SSHConfigDefaultPath) {
		providers = append(providers, NewSSHProvider("", ""))
	}

	// k8s
	if config.KubernetesProvider != nil && config.KubernetesProvider.Enabled {
		providers = append(providers, NewKubernetesProvider(config.KubernetesProvider.Clusters, config.KubernetesProvider.StartDirectory))
	}

	// usql
	if config.USQLProvider != nil && config.USQLProvider.Enabled {
		providers = append(providers, NewUSQLProvider(config.USQLProvider.ConfigFile, config.USQLProvider.StartDirectory))
	} else if config.USQLProvider == nil && filesystem.FileExists(USQLConfigDefaultPath) {
		providers = append(providers, NewUSQLProvider("", ""))
	}

	// static
	if config.StaticProvider != nil && config.StaticProvider.Enabled {
		providers = append(providers, StaticProvider{
			StaticOptions: config.StaticProvider.StaticOptions,
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

// CollectOptions collects the options from the providers, optionally filtered by name
func CollectOptions(providers []Provider, byName []string, maxCacheAge int) ([]Option, []error) {
	var options []Option
	var errs []error

	for _, p := range providers {
		if len(byName) > 0 && !slices.Contains(byName, p.Name()) {
			continue
		}

		opts, err := p.OptionsOrCache(float64(maxCacheAge))
		if err != nil {
			errs = append(errs, errors.Join(errtypes.ErrFailedToGetOptionsFromProvider, err))
		}

		options = append(options, opts...)
	}

	return options, errs
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

func addTagToOption(option *Option, tag string) {
	if !slices.Contains(option.Tags, tag) {
		option.Tags = append(option.Tags, tag)
	}
}
