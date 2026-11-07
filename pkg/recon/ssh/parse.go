package ssh

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/PhilippHeuer/fuzzmux/pkg/util"
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

// maxIncludeDepth limits nested Include statements, matches the limit used by OpenSSH
const maxIncludeDepth = 16

// resolveIncludes resolves Include statements in the SSH configuration file
func resolveIncludes(rootFile string, configBytes []byte) ([]byte, error) {
	return resolveIncludesDepth(filepath.Dir(rootFile), configBytes, 0)
}

func resolveIncludesDepth(baseDir string, configBytes []byte, depth int) ([]byte, error) {
	if depth > maxIncludeDepth {
		return nil, fmt.Errorf("too many nested include statements (max %d)", maxIncludeDepth)
	}

	lines := strings.Split(string(configBytes), "\n")
	var resolvedConfig []string

	for _, line := range lines {
		paths, isInclude := parseIncludeLine(line)
		if !isInclude {
			resolvedConfig = append(resolvedConfig, line)
			continue
		}
		if len(paths) == 0 {
			return nil, fmt.Errorf("include statement is missing a file path")
		}

		for _, includePath := range paths {
			// resolve home, env vars and relative paths
			includePath = util.ExpandTilde(includePath)
			includePath = os.ExpandEnv(includePath)
			if !filepath.IsAbs(includePath) {
				includePath = filepath.Join(baseDir, includePath)
			}

			// glob patterns, e.g. Include ~/.ssh/config.d/*
			matches, err := filepath.Glob(includePath)
			if err != nil {
				return nil, fmt.Errorf("invalid include pattern '%s': %w", includePath, err)
			}
			if len(matches) == 0 && !hasGlobMeta(includePath) {
				return nil, fmt.Errorf("failed to read included file '%s': %w", includePath, os.ErrNotExist)
			}

			for _, match := range matches {
				info, err := os.Stat(match)
				if err != nil {
					return nil, fmt.Errorf("failed to read included file '%s': %w", match, err)
				}
				if info.IsDir() {
					continue
				}

				includeContent, err := os.ReadFile(match)
				if err != nil {
					return nil, fmt.Errorf("failed to read included file '%s': %w", match, err)
				}

				// nested includes, relative paths stay relative to the root config directory
				includeContent, err = resolveIncludesDepth(baseDir, includeContent, depth+1)
				if err != nil {
					return nil, err
				}

				resolvedConfig = append(resolvedConfig, string(includeContent))
			}
		}
	}

	return []byte(strings.Join(resolvedConfig, "\n")), nil
}

// parseIncludeLine checks if the line is an Include statement (case-insensitive, optional "=") and returns the paths
func parseIncludeLine(line string) ([]string, bool) {
	trimmed := strings.TrimSpace(line)
	if len(trimmed) < len("include") || !strings.EqualFold(trimmed[:len("include")], "include") {
		return nil, false
	}

	rest := trimmed[len("include"):]
	if rest != "" && rest[0] != ' ' && rest[0] != '\t' && rest[0] != '=' {
		return nil, false // other keyword with the same prefix
	}
	rest = strings.TrimSpace(rest)
	rest = strings.TrimSpace(strings.TrimPrefix(rest, "="))

	// strip trailing comment
	if idx := strings.Index(rest, "#"); idx >= 0 {
		rest = rest[:idx]
	}

	return strings.Fields(rest), true
}

func hasGlobMeta(path string) bool {
	return strings.ContainsAny(path, "*?[")
}
