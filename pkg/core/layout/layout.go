package layout

import (
	"fmt"

	"github.com/PhilippHeuer/fuzzmux/pkg/config"
	"github.com/PhilippHeuer/fuzzmux/pkg/provider"
	"github.com/cidverse/go-rules/pkg/expr"
	"github.com/rs/zerolog/log"
)

func GetLayout(conf config.Config, selected *provider.Option, templateName string, defaultName string) (config.Layout, error) {
	// context
	ruleContext := map[string]interface{}{
		"PROVIDER_NAME":   selected.ProviderName,
		"ID":              selected.Id,
		"NAME":            selected.Name,
		"DISPLAY_NAME":    selected.DisplayName,
		"START_DIRECTORY": selected.StartDirectory,
		"CONTEXT":         selected.Context,
		"TAGS":            selected.Tags,
	}

	// auto-detect template if not specified
	if templateName == "" {
		log.Debug().Str("template-name", templateName).Interface("layouts", conf.Layouts).Interface("context", ruleContext).Msg("no template specified, auto-detecting")
		for key, l := range conf.Layouts {
			if len(l.Rules) > 0 && evalRules(l.Rules, ruleContext) > 0 {
				templateName = key
				break
			}
		}
	}

	// fallback to default
	if templateName == "" {
		templateName = defaultName
	}

	// get template
	template, exists := conf.Layouts[templateName]
	if !exists {
		return config.Layout{}, fmt.Errorf("template '%s' not found", templateName)
	}

	// filter windows and commands
	template.Windows = FilterWindows(template.Windows, ruleContext)

	return template, nil
}

func FilterWindows(windows []config.Window, ruleContext map[string]interface{}) []config.Window {
	var result []config.Window

	for _, window := range windows {
		if len(window.Rules) == 0 || evalRules(window.Rules, ruleContext) > 0 {
			// filter commands
			window.Commands = FilterCommands(window.Commands, ruleContext)

			result = append(result, window)
		}
	}

	return result
}

func FilterCommands(commands []config.Command, ruleContext map[string]interface{}) []config.Command {
	var result []config.Command

	for _, command := range commands {
		if len(command.Rules) == 0 || evalRules(command.Rules, ruleContext) > 0 {
			result = append(result, command)
		}
	}

	return result
}

func evalRules(rules []string, ctx map[string]interface{}) int {
	count, err := expr.EvaluateRules(rules, ctx)
	if err != nil {
		log.Warn().Interface("rules", rules).Interface("context", ctx).Msg("failed to evaluate rules")
		return 0
	}
	return count
}
