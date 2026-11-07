package util

import (
	"strings"
)

func ExpandPlaceholders(command string, key string, value string) string {
	if value == "" {
		return command
	}

	// prepare
	rawPlaceholder := "{{!" + key + "}}"
	quotedPlaceholder := "{{" + key + "}}"

	// replace
	command = strings.ReplaceAll(command, rawPlaceholder, value)
	command = strings.ReplaceAll(command, quotedPlaceholder, escapeDoubleQuoted(value))

	return command
}

// ExpandPlaceholdersRaw replaces both placeholder forms with the unmodified value, for templates that are not evaluated by a shell (display names, paths)
func ExpandPlaceholdersRaw(input string, key string, value string) string {
	if value == "" {
		return input
	}

	input = strings.ReplaceAll(input, "{{!"+key+"}}", value)
	input = strings.ReplaceAll(input, "{{"+key+"}}", value)

	return input
}

// doubleQuoteEscaper escapes all characters that keep a special meaning inside a double-quoted shell string
var doubleQuoteEscaper = strings.NewReplacer(
	`\`, `\\`,
	`"`, `\"`,
	`$`, `\$`,
	"`", "\\`",
)

// escapeDoubleQuoted makes the value safe to use inside a double-quoted shell string
func escapeDoubleQuoted(value string) string {
	return doubleQuoteEscaper.Replace(value)
}
