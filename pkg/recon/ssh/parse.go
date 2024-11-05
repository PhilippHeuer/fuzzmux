package ssh

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kevinburke/ssh_config"
)

// ParseFile parses an ssh config file
// NOTE: This manually resolves Include statements due to limitations in the ssh_config library
func ParseFile(path string) (*ssh_config.Config, error) {
	// read file
	configBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read SSH config file: %w", err)
	}

	// resolve includes
	configBytes, err = resolveIncludes(path, configBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve includes: %w", err)
	}

	// parse
	sshConfig, err := ssh_config.DecodeBytes(configBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ssh config: %w", err)
	}

	return sshConfig, nil
}

// resolveIncludes resolves Include statements in the SSH configuration file
func resolveIncludes(rootFile string, configBytes []byte) ([]byte, error) {
	// Split the configuration content into lines
	lines := strings.Split(string(configBytes), "\n")
	var resolvedConfig []string

	for _, line := range lines {
		if strings.HasPrefix(line, "Include") {
			// Extract the file path from the Include statement
			includePath := strings.TrimPrefix(line, "Include")
			includePath = strings.TrimSpace(includePath)
			if includePath == "" {
				return nil, fmt.Errorf("include statement is missing a file path")
			}

			// resolve relative paths
			includePath = strings.Replace(includePath, "~", os.Getenv("HOME"), 1)
			includePath = os.ExpandEnv(includePath)
			if !filepath.IsAbs(includePath) {
				includePath = filepath.Join(filepath.Dir(rootFile), includePath)
			}

			// Read the included file content
			includeContent, err := os.ReadFile(includePath)
			if err != nil {
				return nil, fmt.Errorf("failed to read included file '%s': %w", includePath, err)
			}

			// Insert the included content into the resolved configuration
			resolvedConfig = append(resolvedConfig, string(includeContent))
		} else {
			// Preserve other lines as they are
			resolvedConfig = append(resolvedConfig, line)
		}
	}

	// Join the resolved lines and convert them back to bytes
	resolvedBytes := []byte(strings.Join(resolvedConfig, "\n"))

	return resolvedBytes, nil
}
