package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfigEmptyFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "empty.yaml")
	require.NoError(t, os.WriteFile(path, nil, 0600))

	conf, err := loadConfig(path)
	require.NoError(t, err)
	assert.Empty(t, conf.Modules)
}

func TestMergeConfigWithoutLayouts(t *testing.T) {
	merged := MergeConfig(Config{}, Config{Layouts: map[string]Layout{"a": {}}})
	assert.Contains(t, merged.Layouts, "a")
}
