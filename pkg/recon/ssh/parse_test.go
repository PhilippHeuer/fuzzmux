package ssh

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseIncludeLine(t *testing.T) {
	tests := []struct {
		line      string
		paths     []string
		isInclude bool
	}{
		{"Include config.d/*", []string{"config.d/*"}, true},
		{"  include a b # comment", []string{"a", "b"}, true},
		{"Include=a", []string{"a"}, true},
		{"IncludeFoo bar", nil, false},
		{"Host include", nil, false},
	}

	for _, tt := range tests {
		paths, isInclude := parseIncludeLine(tt.line)
		assert.Equal(t, tt.isInclude, isInclude, tt.line)
		assert.Equal(t, tt.paths, paths, tt.line)
	}
}

func TestParseFileWithIncludes(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "config.d"), 0700))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config"), []byte("include config.d/*\nInclude config.d/missing-*\n"), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.d", "a"), []byte("Host a\n  HostName a.example.com\n  user alice\n  # tag: prod\n"), 0600))

	module := NewModule(ModuleConfig{ConfigFile: filepath.Join(dir, "config")})
	options, err := module.Options()
	require.NoError(t, err)
	require.Len(t, options, 1)
	assert.Equal(t, "a", options[0].Id)
	assert.Equal(t, "a.example.com", options[0].Context["host"])
	assert.Equal(t, "alice", options[0].Context["user"])
	assert.Equal(t, []string{"prod"}, options[0].Tags)
}
