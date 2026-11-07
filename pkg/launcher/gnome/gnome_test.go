package gnome

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseGVariantUint(t *testing.T) {
	v, err := parseGVariantUint("(uint32 3,)\n")
	require.NoError(t, err)
	assert.Equal(t, "3", v)

	_, err = parseGVariantUint("(uint32,)")
	assert.Error(t, err)
}

func TestParseWorkspaceWindows(t *testing.T) {
	windows, err := parseWorkspaceWindows(`('[{"id":12,"pid":100,"focus":true},{"id":"13"},{"title":"no id"}]',)` + "\n")
	require.NoError(t, err)
	require.Len(t, windows, 2)

	assert.Equal(t, "12", windows[0].ID)
	assert.Equal(t, 100, windows[0].PID)
	require.NotNil(t, windows[0].Focus)
	assert.True(t, *windows[0].Focus)

	assert.Equal(t, "13", windows[1].ID)
	assert.Equal(t, 0, windows[1].PID)
	assert.Nil(t, windows[1].Focus)
}
