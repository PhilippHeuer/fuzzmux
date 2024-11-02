package util

import (
	"strconv"
	"strings"
)

func ExpandPlaceholders(command string, key string, value string) string {
	if value == "" {
		return command
	}

	// raw value
	command = strings.ReplaceAll(command, "{{!"+key+"}}", value)

	// quoted values
	quotedValue := strconv.Quote(value)
	trimmedQuotedValue := quotedValue[1 : len(quotedValue)-1]
	command = strings.ReplaceAll(command, "{{"+key+"}}", trimmedQuotedValue)

	return command
}
