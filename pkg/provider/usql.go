package provider

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/PhilippHeuer/fuzzmux/pkg/core/parser/usql"
	"github.com/PhilippHeuer/fuzzmux/pkg/core/util"
	"github.com/rs/zerolog/log"
)

var USQLConfigDefaultPath = filepath.Join(os.Getenv("HOME"), ".config", "usql", "config.yaml")

type USQLProvider struct {
	ConfigPath     string
	StartDirectory string
}

func (p USQLProvider) Name() string {
	return "usql"
}

func (p USQLProvider) Options() ([]Option, error) {
	var options []Option

	// parse config
	conf, err := usql.ParseFile(p.ConfigPath)
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
		opt := Option{
			ProviderName:   p.Name(),
			Id:             "usql-" + key,
			DisplayName:    fmt.Sprintf("%s @ %s:%d", conn.Username, conn.Hostname, conn.Port),
			Name:           key,
			StartDirectory: p.StartDirectory,
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

func (p USQLProvider) OptionsOrCache(maxAge float64) ([]Option, error) {
	options, err := LoadOptions(p.Name(), maxAge)
	if err == nil {
		return options, nil
	}

	options, err = p.Options()
	if err != nil {
		return nil, fmt.Errorf("failed to get options: %w", err)
	}

	err = SaveOptions(p.Name(), options)
	if err != nil {
		log.Warn().Err(err).Msg("failed to save options to cache")
	}

	return options, nil
}

func (p USQLProvider) SelectOption(option *Option) error {
	err := option.CreateStartDirectoryIfMissing()
	if err != nil {
		return err
	}

	return nil
}

func NewUSQLProvider(configPath string, startDirectory string) USQLProvider {
	if configPath == "" {
		configPath = USQLConfigDefaultPath
	}

	return USQLProvider{
		ConfigPath:     configPath,
		StartDirectory: startDirectory,
	}
}
