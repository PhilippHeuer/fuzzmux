package util

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
)

func GetHomeDir() string {
	homeDir, homeErr := os.UserHomeDir()
	if homeErr != nil {
		log.Fatal().Err(homeErr).Msg("failed to get user home directory")
	}

	return homeDir
}

// ExpandTilde replaces a leading ~ (as in "~" or "~/...") with the home directory of the current user
func ExpandTilde(input string) string {
	if input == "~" || strings.HasPrefix(input, "~/") {
		home := os.Getenv("HOME")
		if home == "" {
			home = GetHomeDir()
		}
		return home + input[1:]
	}
	return input
}

func ResolvePath(input string) string {
	if input == "" {
		return ""
	}

	// expand ~
	if input == "~" || strings.HasPrefix(input, "~/") {
		input = filepath.Join(GetHomeDir(), input[1:])
	}

	// env vars
	input = os.ExpandEnv(input)

	return input
}
