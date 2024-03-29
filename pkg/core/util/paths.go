package util

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/rs/zerolog/log"
)

func GetAppDataDir() string {
	var appDataDir string

	switch runtime.GOOS {
	case "darwin":
		appDataDir, _ = os.UserHomeDir()
		appDataDir = filepath.Join(appDataDir, "Library", "Application Support")
	default:
		// On Linux and other platforms, follow XDG Base Directory Specification
		// Use XDG_CONFIG_HOME if set, otherwise fallback to the default
		if configHome := os.Getenv("XDG_CONFIG_HOME"); configHome != "" {
			appDataDir = configHome
		} else {
			// Default to home directory + .config
			appDataDir, _ = os.UserConfigDir()
		}
	}

	return appDataDir
}

func GetAppStateDir() string {
	var appDataDir string

	switch runtime.GOOS {
	case "darwin":
		appDataDir, _ = os.UserHomeDir()
		appDataDir = filepath.Join(appDataDir, "Library", "Application Support")
	default:
		// On Linux and other platforms, follow XDG Base Directory Specification
		// Use XDG_DATA_HOME if set, otherwise fallback to the default
		if dataHome := os.Getenv("XDG_DATA_HOME"); dataHome != "" {
			appDataDir = dataHome
		} else {
			// Default to home directory + .local/state
			appDataDir, _ = os.UserHomeDir()
			appDataDir = filepath.Join(appDataDir, ".local", "state")
		}
	}

	return appDataDir
}

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
