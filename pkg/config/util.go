package config

import (
	"fmt"
)

func GetTemplate(conf Config, templateName string, defaultName string) ([]Window, error) {
	if templateName == "" {
		templateName = defaultName
	}

	template, exists := conf.WindowTemplates[templateName]
	if !exists {
		return nil, fmt.Errorf("template '%s' not found", templateName)
	}

	return template, nil
}
