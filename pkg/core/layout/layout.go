package layout

import (
	"fmt"

	"github.com/PhilippHeuer/tmux-tms/pkg/config"
	"github.com/PhilippHeuer/tmux-tms/pkg/core/rules"
	"github.com/PhilippHeuer/tmux-tms/pkg/provider"
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
			if len(l.Rules) > 0 && rules.EvaluateRules(l.Rules, ruleContext) > 0 {
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
	return template, nil
}
