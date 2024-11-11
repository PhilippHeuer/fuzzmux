package ssh

import (
	"fmt"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon"
	"os"
	"path/filepath"
	"strings"
)

const moduleType = "ssh"

var DefaultPath = filepath.Join(os.Getenv("HOME"), ".ssh", "config")

type Module struct {
	Config ModuleConfig
}

type ModuleConfig struct {
	// Name is used to override the default module name
	Name string `yaml:"name,omitempty"`

	// DisplayName is a template string to render a custom display name
	DisplayName string `yaml:"display-name"`

	// StartDirectory is a template string that defines the start directory
	StartDirectory string `yaml:"start-directory"`

	// ConfigFile is used in case your ssh config is not in the default location
	ConfigFile string `yaml:"file"`

	// Mode controls how sessions or windows are created for SSH connections
	Mode SSHMode `yaml:"mode"`
}

type SSHMode string

const (
	SSHSessionMode SSHMode = "session"
	SSHWindowMode  SSHMode = "window"
)

func (p Module) Name() string {
	if p.Config.Name != "" {
		return p.Config.Name
	}
	return moduleType
}

func (p Module) Type() string {
	return moduleType
}

func (p Module) Options() ([]recon.Option, error) {
	var result []recon.Option

	// parse ssh config
	sshConfig, err := ParseFile(p.Config.ConfigFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ssh config: %w", err)
	}
	for _, host := range sshConfig.Hosts {
		for _, pattern := range host.Patterns {
			// skip wildcards
			if strings.Contains(pattern.String(), "*") || strings.Contains(pattern.String(), "?") {
				continue
			}

			// parse
			name := pattern.String()
			hostname := ""
			user := "root@"
			var tags []string
			for _, node := range host.Nodes {
				line := strings.TrimSpace(node.String())
				if strings.HasPrefix(line, "Hostname ") {
					hostname = strings.TrimSpace(strings.TrimPrefix(line, "Hostname "))
				}
				if strings.HasPrefix(line, "User ") {
					user = strings.TrimSpace(strings.TrimPrefix(line, "User ")) + "@"
				}

				if strings.HasPrefix(line, "# tag: ") {
					tags = append(tags, strings.TrimSpace(strings.TrimPrefix(line, "# tag: ")))
				}
			}

			// result
			opt := recon.Option{
				ProviderName:   p.Name(),
				ProviderType:   p.Type(),
				Id:             name,
				DisplayName:    fmt.Sprintf("%s [%s%s]", name, user, hostname),
				Name:           name,
				StartDirectory: p.Config.StartDirectory,
				Tags:           tags,
				Context: map[string]string{
					"host": hostname,
					"user": strings.TrimRight(user, "@"),
				},
			}
			opt.ProcessUserTemplateStrings(p.Config.DisplayName, p.Config.StartDirectory)
			result = append(result, opt)
		}
	}

	return result, nil
}

func (p Module) OptionsOrCache(maxAge float64) ([]recon.Option, error) {
	return recon.OptionsOrCache(p, maxAge)
}

func (p Module) SelectOption(option *recon.Option) error {
	err := option.CreateStartDirectoryIfMissing()
	if err != nil {
		return err
	}

	return nil
}

func (p Module) Columns() []recon.Column {
	return append(recon.DefaultColumns(),
		recon.Column{Key: "host", Name: "Host"},
		recon.Column{Key: "user", Name: "User"},
	)
}

func NewModule(config ModuleConfig) Module {
	if config.ConfigFile == "" {
		config.ConfigFile = DefaultPath
	}

	return Module{
		Config: config,
	}
}
