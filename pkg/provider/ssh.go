package provider

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/PhilippHeuer/fuzzmux/pkg/core/parser/sshconfig"
	"github.com/rs/zerolog/log"
)

var SSHConfigDefaultPath = filepath.Join(os.Getenv("HOME"), ".ssh", "config")

type SSHProvider struct {
	ConfigPath     string
	StartDirectory string
}

func (p SSHProvider) Name() string {
	return "ssh"
}

func (p SSHProvider) Options() ([]Option, error) {
	var options []Option

	// parse ssh config
	sshConfig, err := sshconfig.ParseFile(p.ConfigPath)
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
			opt := Option{
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

func (p SSHProvider) OptionsOrCache(maxAge float64) ([]Option, error) {
	options, err := LoadOptions(p.Name(), maxAge)
	if err == nil {
		return options, nil
	}

	options, err = p.Options()
	if err != nil {
		return nil, fmt.Errorf("failed to get options: %w", err)
	}

	err = SaveOptions(p.Name(), options)
	if err != nil {
		log.Warn().Err(err).Msg("failed to save options to cache")
	}

	return options, nil
}

func (p SSHProvider) SelectOption(option *Option) error {
	err := option.CreateStartDirectoryIfMissing()
	if err != nil {
		return err
	}

	return nil
}

func NewSSHProvider(configPath string, startDirectory string) SSHProvider {
	if configPath == "" {
		configPath = SSHConfigDefaultPath
	}

	return SSHProvider{
		ConfigPath:     configPath,
		StartDirectory: startDirectory,
	}
}
