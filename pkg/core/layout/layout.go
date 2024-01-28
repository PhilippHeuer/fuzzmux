package layout

import (
	"fmt"
	"slices"

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
	template.Apps = FilterApps(template.Apps, ruleContext)

	return template, nil
}

func FilterApps(apps []config.App, ruleContext map[string]interface{}) []config.App {
	var result []config.App

	var groupIDs []string
	for _, app := range apps {
		if len(app.Rules) == 0 || evalRules(app.Rules, ruleContext) > 0 {
			// groupID check
			if app.Group != "" {
				if slices.Contains(groupIDs, app.Group) {
					continue
				}

				groupIDs = append(groupIDs, app.Group)
			}

			// filter commands
			app.Commands = FilterCommands(app.Commands, ruleContext)

			result = append(result, app)
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
