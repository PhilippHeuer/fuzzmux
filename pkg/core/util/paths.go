package util

import (
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

func GetHomeDir() string {
	homeDir, homeErr := os.UserHomeDir()
	if homeErr != nil {
		log.Fatal().Err(homeErr).Msg("failed to get user home directory")
	}

	return homeDir
}

func ResolvePath(input string) string {
	// expand ~
	if input[0] == '~' {
		input = filepath.Join(GetHomeDir(), input[1:])
	}

	// env vars
	input = os.ExpandEnv(input)

	return input
}
