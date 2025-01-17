package config

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

var configDir = filepath.Join(xdg.ConfigHome, "fuzzmux")

//go:embed layouts.yaml
var layoutsConfig []byte

func ResolvedConfig() (Config, error) {
	// load config
	config, err := LoadConfig()
	if err != nil {
		return Config{}, err
	}

	// finder
	if config.Finder == nil {
		config.Finder = &FinderConfig{
			Executable: "",
			Preview:    true,
		}
	}
	if config.Finder.FZFDelimiter == "" {
		config.Finder.FZFDelimiter = "\x1F"
	}

	// load default templates
	if config.Layouts == nil {
		config.Layouts = make(map[string]Layout)
	}

	var layouts Config
	err = yaml.Unmarshal(layoutsConfig, &layouts)
	if err != nil {
		return Config{}, fmt.Errorf("failed to parse default layouts: %w", err)
	}
	for key, layout := range layouts.Layouts {
		if _, exists := config.Layouts[key]; !exists {
			config.Layouts[key] = layout
		}
	}

	// resolve config
	log.Debug().Interface("config", config).Msg("resolved config")

	return config, nil
}

func LoadConfig() (Config, error) {
	// main config
	config, err := loadConfig(filepath.Join(configDir, "fuzzmux.yaml"))
	if err != nil {
		return Config{}, err
	}

	// user config
	userConfig, err := loadConfig(filepath.Join(configDir, "fuzzmux.user.yaml"))
	if err == nil {
		config = MergeConfig(config, userConfig)
	}

	return config, nil
}

func loadConfig(path string) (Config, error) {
	var config Config

	file, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}

func SaveConfig(config Config) {
	// create directory if it doesn't exist
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		_ = os.MkdirAll(configDir, os.ModePerm)
	}

	// save file
	file, _ := json.MarshalIndent(config, "", " ")
	_ = os.WriteFile(filepath.Join(configDir, "fuzzmux.yaml"), file, 0644)
}

func CommandsAsStringSlice(commands []Command) []string {
	var result []string

	for _, cmd := range commands {
		result = append(result, cmd.Command)
	}

	return result
}

func MergeConfig(a Config, b Config) Config {
	// merge modules
	if b.Modules != nil {
		for _, module := range b.Modules {
			a.Modules = append(a.Modules, module)
		}
	}

	// merge layouts
	if b.Layouts != nil {
		for key, layout := range b.Layouts {
			a.Layouts[key] = layout
		}
	}

	return a
}
