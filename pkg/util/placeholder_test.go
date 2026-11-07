package util

import (
	"testing"
)

func TestExpandPlaceholders(t *testing.T) {
	tests := []struct {
		command string
		key     string
		value   string
		want    string
	}{
		{`echo "{{name}}"`, "name", "Fuzz", `echo "Fuzz"`},
		{`echo "{{name}}"`, "name", "Fuzz\"; rm -rf /tmp/test", `echo "Fuzz\"; rm -rf /tmp/test"`},
		{`echo "{{!name}}"`, "name", "Fuzz\"; rm -rf /tmp/test", `echo "Fuzz"; rm -rf /tmp/test"`},
		{`echo "{{name}}"`, "name", "$(rm -rf /tmp/test)", `echo "\$(rm -rf /tmp/test)"`},
		{`echo "{{name}}"`, "name", "`id` ${HOME} a\\b", "echo \"\\`id\\` \\${HOME} a\\\\b\""},
		{`echo "{{name}}"`, "name", "", `echo "{{name}}"`},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			got := ExpandPlaceholders(tt.command, tt.key, tt.value)
			if got != tt.want {
				t.Errorf("expandPlaceholders(%s, %s, %s) = %v, want %v", tt.command, tt.key, tt.value, got, tt.want)
			}
		})
	}
}

func TestExpandPlaceholdersRaw(t *testing.T) {
	tests := []struct {
		input string
		value string
		want  string
	}{
		{`{{name}}`, "Cost $5 \"x\"", "Cost $5 \"x\""},
		{`{{!name}}`, "a\\b", "a\\b"},
		{`{{name}}`, "", `{{name}}`},
	}

	for _, tt := range tests {
		if got := ExpandPlaceholdersRaw(tt.input, "name", tt.value); got != tt.want {
			t.Errorf("ExpandPlaceholdersRaw(%s, %s) = %v, want %v", tt.input, tt.value, got, tt.want)
		}
	}
}
