package app

import (
	"errors"
	"fmt"
	"github.com/PhilippHeuer/fuzzmux/pkg/config"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon/kubernetes"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon/project"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon/ssh"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon/static"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon/usql"
	"github.com/PhilippHeuer/fuzzmux/pkg/types"
	"github.com/cidverse/cidverseutils/filesystem"
	"github.com/rs/zerolog/log"
	"slices"
)

// ConfigToReconModules initializes the recon modules based on the provided config
func ConfigToReconModules(config config.Config) []recon.Module {
	var providers []recon.Module

	// projects
	if config.ProjectProvider != nil && config.ProjectProvider.Enabled {
		providers = append(providers, project.ProjectProvider{
			SourceDirectories: config.ProjectProvider.SourceDirectories,
			Checks:            config.ProjectProvider.Checks,
			DisplayFormat:     config.ProjectProvider.DisplayFormat,
		})
	}

	// ssh
	if config.SSHProvider != nil && config.SSHProvider.Enabled {
		providers = append(providers, ssh.NewSSHProvider(config.SSHProvider.ConfigFile, config.SSHProvider.StartDirectory))
	} else if config.SSHProvider == nil && filesystem.FileExists(ssh.DefaultPath) {
		providers = append(providers, ssh.NewSSHProvider("", ""))
	}

	// k8s
	if config.KubernetesProvider != nil && config.KubernetesProvider.Enabled {
		providers = append(providers, kubernetes.NewKubernetesProvider(config.KubernetesProvider.Clusters, config.KubernetesProvider.StartDirectory))
	}

	// usql
	if config.USQLProvider != nil && config.USQLProvider.Enabled {
		providers = append(providers, usql.NewUSQLProvider(config.USQLProvider.ConfigFile, config.USQLProvider.StartDirectory))
	} else if config.USQLProvider == nil && filesystem.FileExists(usql.USQLConfigDefaultPath) {
		providers = append(providers, usql.NewUSQLProvider("", ""))
	}

	// static
	if config.StaticProvider != nil && config.StaticProvider.Enabled {
		providers = append(providers, static.StaticProvider{
			StaticOptions: config.StaticProvider.StaticOptions,
		})
	}

	return providers
}

// FindReconModuleByName finds a recon module in the list by name
func FindReconModuleByName(providers []recon.Module, name string) (recon.Module, error) {
	for _, p := range providers {
		if p.Name() == name {
			return p, nil
		}
	}

	return nil, errors.Join(types.ErrReconModuleNotFound, fmt.Errorf("%q not found", name))
}

// FindReconModulesByNames finds multiple recon modules in the list by name
func FindReconModulesByNames(providers []recon.Module, names []string) []recon.Module {
	var result []recon.Module

	for _, p := range providers {
		if slices.Contains(names, p.Name()) {
			result = append(result, p)
		}
	}

	return result
}

// GatherReconOptions collects options from specified recon modules or all available modules if none are specified.
func GatherReconOptions(conf config.Config, moduleNames []string, showTags []string, hideTags []string, maxCacheAge int) ([]recon.Module, []recon.Option) {
	modules := ConfigToReconModules(conf)
	if len(moduleNames) > 0 {
		modules = FindReconModulesByNames(modules, moduleNames)
	}

	var options []recon.Option
	options, errs := CollectOptions(modules, maxCacheAge)
	if len(options) == 0 && len(errs) > 0 {
		log.Fatal().Errs("errors", errs).Msg("failed to collect options")
	} else if len(errs) > 0 {
		log.Warn().Errs("errors", errs).Msg("at least one recon failed to collect options")
	}
	options = FilterOptions(options, showTags, hideTags)

	return modules, options
}
