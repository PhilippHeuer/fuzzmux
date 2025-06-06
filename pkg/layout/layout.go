package layout

import (
	"fmt"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon"
	"slices"

	"github.com/PhilippHeuer/fuzzmux/pkg/config"
	"github.com/cidverse/go-rules/pkg/expr"
	"github.com/rs/zerolog/log"
)

func GetLayout(conf config.Config, selected *recon.Option, templateName string, defaultName string) (config.Layout, error) {
	// context
	ruleContext := map[string]interface{}{
		"PROVIDER_NAME":   selected.ProviderName,
		"PROVIDER_TYPE":   selected.ProviderType,
		"ID":              selected.Id,
		"NAME":            selected.Name,
		"DISPLAY_NAME":    selected.DisplayName,
		"DESCRIPTION":     selected.Description,
		"WEB":             selected.Web,
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
	} else if _, ok := conf.Layouts[templateName]; !ok {
		log.Warn().Str("template-name", templateName).Msg("template name not found - check your configuration, falling back to default")
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
