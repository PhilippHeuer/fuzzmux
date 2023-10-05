package provider

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kevinburke/ssh_config"
)

type SSHProvider struct {
}

func (p SSHProvider) Name() string {
	return "ssh"
}

func (p SSHProvider) Options() ([]Option, error) {
	var options []Option

	// parse ssh config
	f, _ := os.Open(filepath.Join(os.Getenv("HOME"), ".ssh", "config"))
	sshConfig, _ := ssh_config.Decode(f)
	for _, host := range sshConfig.Hosts {
		for _, pattern := range host.Patterns {
			// skip wildcards
			if strings.Contains(pattern.String(), "*") || strings.Contains(pattern.String(), "?") {
				continue
			}

			// parse
			name := host.Patterns[0].String()
			hostname := ""
			for _, node := range host.Nodes {
				line := strings.TrimSpace(node.String())
				if strings.HasPrefix(line, "Hostname ") {
					hostname = strings.TrimSpace(strings.TrimPrefix(line, "Hostname "))
				}
			}

			// add to list
			options = append(options, Option{
				DisplayName: fmt.Sprintf("%s [%s]", name, hostname),
				Name:        name,
				Context: map[string]string{
					"host": hostname,
				},
			})
		}
	}

	return options, nil
}
