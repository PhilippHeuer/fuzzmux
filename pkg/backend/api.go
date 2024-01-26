package backend

import (
	"strings"

	"github.com/PhilippHeuer/fuzzmux/pkg/config"
	"github.com/PhilippHeuer/fuzzmux/pkg/provider"
)

type Opts struct {
	SessionName string
	Layout      config.Layout
	AppendMode  AppendMode
	BaseIndex   int
}

type AppendMode string

const (
	CreateOrAttachSession AppendMode = "session"
)

type Provider interface {
	Name() string
	Run(option *provider.Option, opts Opts) error
}

func expandCommand(option *provider.Option, command string) string {
	command = expandPlaceholders(command, "name", option.Name)
	command = expandPlaceholders(command, "displayName", option.DisplayName)
	command = expandPlaceholders(command, "startDirectory", option.StartDirectory)
	for k, v := range option.Context {
		command = expandPlaceholders(command, k, v)
	}

	return command
}

func expandPlaceholders(command string, key string, value string) string {
	command = strings.ReplaceAll(command, "$"+key, value)
	command = strings.ReplaceAll(command, "${"+key+"}", value)
	command = strings.ReplaceAll(command, "$"+strings.ToUpper(key), value)
	command = strings.ReplaceAll(command, "${"+strings.ToUpper(key)+"}", value)
	command = strings.ReplaceAll(command, "$"+strings.ToLower(key), value)
	command = strings.ReplaceAll(command, "${"+strings.ToLower(key)+"}", value)

	return command
}
