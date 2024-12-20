package firefox

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSearchBookmarks(t *testing.T) {
	// query
	dbPath, _ := filepath.Abs("testdata/places.sqlite")
	firefoxModule := NewModule(ModuleConfig{
		DbFile: dbPath,
	})
	options, err := firefoxModule.Options()
	require.NoError(t, err)

	// verify
	require.NotEmpty(t, options)
	require.Len(t, options, 3)

	require.Equal(t, "1", options[0].Id)
	require.Equal(t, "Example Site 1", options[0].Name)
	require.Equal(t, "https://example.com", options[0].Context["url"])

	require.Equal(t, "2", options[1].Id)
	require.Equal(t, "Example Site 2", options[1].Name)
	require.Equal(t, "https://example.org", options[1].Context["url"])

	require.Equal(t, "3", options[2].Id)
	require.Equal(t, "Example Site 3", options[2].Name)
	require.Equal(t, "https://example.net", options[2].Context["url"])
}
