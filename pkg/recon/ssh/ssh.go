package ssh

import (
	"fmt"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
)

var DefaultPath = filepath.Join(os.Getenv("HOME"), ".ssh", "config")

type SSHProvider struct {
	ConfigPath     string
	StartDirectory string
}

func (p SSHProvider) Name() string {
	return "ssh"
}

func (p SSHProvider) Options() ([]recon.Option, error) {
	var options []recon.Option

	// parse ssh config
	sshConfig, err := ParseFile(p.ConfigPath)
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

			// option
			opt := recon.Option{
				ProviderName:   p.Name(),
				Id:             name,
				DisplayName:    fmt.Sprintf("%s [%s%s]", name, user, hostname),
				Name:           name,
				StartDirectory: p.StartDirectory,
				Tags:           tags,
				Context: map[string]string{
					"host": hostname,
					"user": strings.TrimRight(user, "@"),
				},
			}

			// add to list
			options = append(options, opt)
		}
	}

	return options, nil
}

func (p SSHProvider) OptionsOrCache(maxAge float64) ([]recon.Option, error) {
	options, err := recon.LoadOptions(p.Name(), maxAge)
	if err == nil {
		return options, nil
	}

	options, err = p.Options()
	if err != nil {
		return nil, fmt.Errorf("failed to get options: %w", err)
	}

	err = recon.SaveOptions(p.Name(), options)
	if err != nil {
		log.Warn().Err(err).Msg("failed to save options to cache")
	}

	return options, nil
}

func (p SSHProvider) SelectOption(option *recon.Option) error {
	err := option.CreateStartDirectoryIfMissing()
	if err != nil {
		return err
	}

	return nil
}

func (p SSHProvider) Columns() []recon.Column {
	return append(recon.DefaultColumns(),
		recon.Column{Key: "host", Name: "Host"},
		recon.Column{Key: "user", Name: "User"},
	)
}

func NewSSHProvider(configPath string, startDirectory string) SSHProvider {
	if configPath == "" {
		configPath = DefaultPath
	}

	return SSHProvider{
		ConfigPath:     configPath,
		StartDirectory: startDirectory,
	}
}
