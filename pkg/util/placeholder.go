package util

import (
	"strconv"
	"strings"
)

func ExpandPlaceholders(command string, key string, value string) string {
	if value == "" {
		return command
	}

	// prepare
	rawPlaceholder := "{{!" + key + "}}"
	quotedPlaceholder := "{{" + key + "}}"
	trimmedQuotedValue := strconv.Quote(value)[1 : len(strconv.Quote(value))-1]

	// replace
	command = strings.ReplaceAll(command, rawPlaceholder, value)
	command = strings.ReplaceAll(command, quotedPlaceholder, trimmedQuotedValue)

	return command
}
