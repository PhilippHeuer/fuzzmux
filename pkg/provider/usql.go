package provider

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/PhilippHeuer/fuzzmux/pkg/core/parser/usql"
	"github.com/PhilippHeuer/fuzzmux/pkg/core/util"
	"github.com/rs/zerolog/log"
)

var USQLConfigDefaultPath = filepath.Join(os.Getenv("HOME"), ".config", "usql", "config.yaml")

type USQLProvider struct {
	ConfigPath string
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
			DisplayName:    fmt.Sprintf("%s @ %s:%d [%s]", conn.Username, conn.Hostname, conn.Port, conn.Database),
			Name:           key,
			StartDirectory: "",
		}
		if conn.Instance != "" {
			opt.DisplayName = fmt.Sprintf("%s @ %s:%d [%s:%s]", conn.Username, conn.Hostname, conn.Port, conn.Instance, conn.Database)
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
	if option.StartDirectory == "" {
		return nil
	}

	// create startDirectory
	if _, err := os.Stat(option.StartDirectory); os.IsNotExist(err) {
		err = os.MkdirAll(option.StartDirectory, 0755)
		if err != nil {
			return errors.Join(ErrFailedToCreateStartDirectory, err)
		}
	}

	return nil
}

func NewUSQLProvider(configPath string) USQLProvider {
	if configPath == "" {
		configPath = USQLConfigDefaultPath
	}

	return USQLProvider{
		ConfigPath: configPath,
	}
}
