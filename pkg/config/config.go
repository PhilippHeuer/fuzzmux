package config

import (
	"fmt"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon/backstage"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon/jira"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon/keycloak"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon/kubernetes"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon/ldap"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon/project"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon/rundeck"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon/ssh"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon/static"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon/usql"
	"gopkg.in/yaml.v3"
)

type Config struct {
	// Modules is a list of recon modules
	Modules []ModuleConfig `yaml:"-"`

	// Layouts is a map of tmux layouts
	Layouts map[string]Layout `yaml:"layouts"`

	// Finder
	Finder *FinderConfig `yaml:"finder"`
}

type ModuleConfig interface{}

func (c *Config) UnmarshalYAML(value *yaml.Node) error {
	// decode known fields, avoid infinite recursion
	aux := &struct {
		Modules []yaml.Node       `yaml:"modules"`
		Layouts map[string]Layout `yaml:"layouts"`
		Finder  *FinderConfig     `yaml:"finder"`
	}{}
	if err := value.Decode(aux); err != nil {
		return err
	}
	c.Modules = nil
	c.Finder = aux.Finder
	c.Layouts = aux.Layouts

	// parse the "recon" field into the appropriate ModuleConfig types
	for key, moduleNode := range aux.Modules {
		var typeInfo struct {
			Type string `yaml:"type"`
		}
		if err := moduleNode.Decode(&typeInfo); err != nil {
			return fmt.Errorf("failed to decode type for module at index %d: %w", key, err)
		}

		var module ModuleConfig
		switch typeInfo.Type {
		case "static":
			module = &static.ModuleConfig{}
		case "project":
			module = &project.ModuleConfig{}
		case "ssh":
			module = &ssh.ModuleConfig{}
		case "kubernetes":
			module = &kubernetes.ModuleConfig{}
		case "usql":
			module = &usql.ModuleConfig{}
		case "ldap":
			module = &ldap.ModuleConfig{}
		case "keycloak":
			module = &keycloak.ModuleConfig{}
		case "backstage":
			module = &backstage.ModuleConfig{}
		case "jira":
			module = &jira.ModuleConfig{}
		case "rundeck":
			module = &rundeck.ModuleConfig{}
		default:
			return fmt.Errorf("unknown module type '%s' for key %d", typeInfo.Type, key)
		}

		if err := moduleNode.Decode(module); err != nil {
			return fmt.Errorf("failed to decode module for key %d: %w", key, err)
		}

		c.Modules = append(c.Modules, module)
	}

	return nil
}
