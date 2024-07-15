package backend

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
		{"echo ${name}", "name", "John", `echo "John"`},
		{"echo !{name}", "name", "John", `echo John`},
		{"echo ${name} and !{name}", "name", "John", `echo "John" and John`},
		{"echo ${name}", "name", "John; rm -rf /", `echo "John; rm -rf /"`},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			got := expandPlaceholders(tt.command, tt.key, tt.value)
			if got != tt.want {
				t.Errorf("expandPlaceholders(%s, %s, %s) = %v, want %v", tt.command, tt.key, tt.value, got, tt.want)
			}
		})
	}
}
