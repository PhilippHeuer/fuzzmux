package usql

import (
	"fmt"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon"
	"github.com/PhilippHeuer/fuzzmux/pkg/util"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
)

const moduleName = "usql"

var USQLConfigDefaultPath = filepath.Join(os.Getenv("HOME"), ".config", "usql", "config.yaml")

type Module struct {
	Config ModuleConfig
}

type ModuleConfig struct {
	// Name is used to override the default module name
	Name string `yaml:"name,omitempty"`

	// ConfigFile is used in case your usql config is not in the default location
	ConfigFile string `yaml:"file"`

	// StartDirectory is used to define the current working directory, supports template variables
	StartDirectory string `yaml:"start-directory"`
}

func (p Module) Name() string {
	if p.Config.Name != "" {
		return p.Config.Name
	}
	return moduleName
}

func (p Module) Type() string {
	return moduleName
}

func (p Module) Options() ([]recon.Option, error) {
	var options []recon.Option

	// parse config
	conf, err := ParseFile(p.Config.ConfigFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse usql config: %w", err)
	}

	for key, conn := range conf.Connections {
		// key needs to be alphanumeric
		if !util.IsAlphanumeric(key) {
			log.Warn().Str("key", key).Msg("skipping non-alphanumeric key")
			continue
		}

		// add to list
		opt := recon.Option{
			ProviderName:   p.Name(),
			ProviderType:   p.Type(),
			Id:             "usql-" + key,
			DisplayName:    fmt.Sprintf("%s @ %s:%d", conn.Username, conn.Hostname, conn.Port),
			Name:           key,
			StartDirectory: p.Config.StartDirectory,
			Context: map[string]string{
				"name":     conn.Name,
				"hostname": conn.Hostname,
				"port":     fmt.Sprintf("%d", conn.Port),
				"username": conn.Username,
				"instance": conn.Instance,
				"database": conn.Database,
			},
		}

		// additional info (SID, instance, database, ...)
		var additionalInfo []string
		if conn.Instance != "" {
			additionalInfo = append(additionalInfo, conn.Instance)
		}
		if conn.Database != "" {
			additionalInfo = append(additionalInfo, conn.Database)
		}
		if len(additionalInfo) > 0 {
			opt.DisplayName = fmt.Sprintf("%s @ %s:%d [%s]", conn.Username, conn.Hostname, conn.Port, strings.Join(additionalInfo, ":"))
		}

		// named connections
		if conn.Name != "" {
			opt.DisplayName = "[" + conn.Name + "] " + opt.DisplayName
		}

		options = append(options, opt)
	}

	return options, nil
}

func (p Module) OptionsOrCache(maxAge float64) ([]recon.Option, error) {
	return recon.OptionsOrCache(p, maxAge)
}

func (p Module) SelectOption(option *recon.Option) error {
	err := option.CreateStartDirectoryIfMissing()
	if err != nil {
		return err
	}

	return nil
}

func (p Module) Columns() []recon.Column {
	return append(recon.DefaultColumns(),
		recon.Column{Key: "host", Name: "Host"},
		recon.Column{Key: "user", Name: "User"},
	)
}

func NewModule(config ModuleConfig) Module {
	if config.ConfigFile == "" {
		config.ConfigFile = USQLConfigDefaultPath
	}

	return Module{
		Config: config,
	}
}
