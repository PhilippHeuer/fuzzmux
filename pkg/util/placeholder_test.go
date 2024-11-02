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
