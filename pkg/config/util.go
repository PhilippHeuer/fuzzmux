package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func getAppDataDir() string {
	var appDataDir string

	switch runtime.GOOS {
	case "windows":
		appDataDir, _ = os.UserConfigDir()
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

func GetTemplate(conf Config, templateName string, defaultName string) ([]Window, error) {
	if templateName == "" {
		templateName = defaultName
	}

	template, exists := conf.WindowTemplates[templateName]
	if !exists {
		return nil, fmt.Errorf("template '%s' not found", templateName)
	}

	return template, nil
}
