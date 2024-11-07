package util

import "testing"

func TestConvertADTimeToRFC3339(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Valid date string",
			input:    "20050917134246.0Z",
			expected: "2005-09-17T13:42:46Z",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Invalid format string",
			input:    "2005-09-17T13:42:46Z",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertLDAPTimeToRFC3339(tt.input)
			if result != tt.expected {
				t.Errorf("ConvertLDAPTimeToRFC3339(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}
