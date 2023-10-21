package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/PhilippHeuer/tmux-tms/pkg/core/util"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

var configDir = filepath.Join(util.GetAppDataDir(), "fuzzmux")

var defaultChecks = []string{".git", ".gitignore", ".hg", ".hgignore", ".svn", ".vscode", ".idea"}

func ResolvedConfig() (Config, error) {
	// load config
	config, err := LoadConfig()
	if err != nil {
		return Config{}, err
	}

	// ssh provider
	if config.SSHProvider == nil {
		config.SSHProvider = &SSHProviderConfig{Enabled: false}

	}
	if config.SSHProvider.Mode == "" {
		config.SSHProvider.Mode = SSHWindowMode
	}

	// project provider
	if config.ProjectProvider == nil || len(config.ProjectProvider.SourceDirectories) == 0 {
		config.ProjectProvider = &ProjectProviderConfig{Enabled: false}
	}
	if config.ProjectProvider.Checks == nil {
		config.ProjectProvider.Checks = defaultChecks
	}
	if config.ProjectProvider.DisplayFormat == "" {
		config.ProjectProvider.DisplayFormat = BaseName
	}

	// default templates
	if config.Layouts == nil {
		config.Layouts = make(map[string]Layout)
	}
	if _, exists := config.Layouts["default"]; !exists {
		config.Layouts["default"] = Layout{
			Windows: []Window{
				{
					Name: "bash",
				},
			},
		}
	}
	if _, exists := config.Layouts["ssh"]; !exists {
		config.Layouts["ssh"] = Layout{
			Windows: []Window{
				{
					Name:     "bash",
					Commands: []string{"exec ssh ${name}"},
				},
			},
		}
	}
	if _, exists := config.Layouts["project"]; !exists {
		config.Layouts["project"] = Layout{
			Windows: []Window{
				{
					Name:    "bash",
					Default: true,
				},
				{
					Name:     "nvim",
					Commands: []string{"nvim +'Telescope find_files hidden=false layout_config={height=0.9}'"},
				},
			},
		}
	}
	if _, exists := config.Layouts["kubernetes"]; !exists {
		config.Layouts["kubernetes"] = Layout{
			Windows: []Window{
				{
					Name:     "k9s",
					Commands: []string{"exec k9s --logoless --headless --readonly --kubeconfig \"${kubeConfig}\" --namespace \"${namespace}\""},
				},
				{
					Name: "kubectl",
					Commands: []string{
						"export KUBECONFIG=\"${kubeConfig}\"",
						"kubectl config set-context --current --namespace=\"${namespace}\"",
					},
				},
			},
		}
	}

	// resolve config
	log.Debug().Interface("config", config).Msg("resolved config")

	return config, nil
}

func LoadConfig() (Config, error) {
	var config Config

	file, err := os.Open(filepath.Join(configDir, "config.yaml"))
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
	_ = os.WriteFile(filepath.Join(configDir, "config.yaml"), file, 0644)
}
