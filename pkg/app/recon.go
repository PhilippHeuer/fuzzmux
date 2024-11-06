package app

import (
	"errors"
	"fmt"
	"github.com/PhilippHeuer/fuzzmux/pkg/config"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon/kubernetes"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon/ldap"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon/project"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon/ssh"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon/static"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon/usql"
	"github.com/PhilippHeuer/fuzzmux/pkg/types"
	"github.com/rs/zerolog/log"
	"slices"
)

// ConfigToReconModules initializes the recon modules based on the provided config
func ConfigToReconModules(conf config.Config) []recon.Module {
	var modules []recon.Module

	for _, m := range conf.Modules {
		switch cfg := m.(type) {
		case *config.ProjectModuleConfig:
			modules = append(modules, project.NewModule(*cfg))
		case *config.SSHModuleConfig:
			modules = append(modules, ssh.NewModule(*cfg))
		case *config.KubernetesModuleConfig:
			modules = append(modules, kubernetes.NewModule(*cfg))
		case *config.USQLModuleConfig:
			modules = append(modules, usql.NewModule(*cfg))
		case *config.StaticModuleConfig:
			modules = append(modules, static.NewModule(*cfg))
		case *config.LDAPModuleConfig:
			modules = append(modules, ldap.NewModule(*cfg))
		default:
			log.Error().Interface("module", m).Msg("unrecognized module type")
		}
	}

	return modules
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
