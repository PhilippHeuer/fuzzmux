package util

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShellQuote(t *testing.T) {
	assert.Equal(t, `'abc'`, ShellQuote("abc"))
	assert.Equal(t, `''`, ShellQuote(""))
	assert.Equal(t, `'it'\''s'`, ShellQuote("it's"))
}

func TestShellQuoteRoundTrip(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("sh not available")
	}

	for _, value := range []string{"plain", "with space", `$(id) ${HOME} "dq" 'sq' \ ; &&`, "`id`"} {
		out, err := exec.Command("sh", "-c", "printf %s "+ShellQuote(value)).Output()
		require.NoError(t, err)
		assert.Equal(t, value, string(out))
	}
}

func TestGetTerminalCommand(t *testing.T) {
	cmd, err := GetTerminalCommand("alacritty", "/tmp/a b", `echo "$(id)"`)
	require.NoError(t, err)
	assert.Equal(t, `alacritty --working-directory '/tmp/a b' -e bash -c 'echo "$(id)"'`, cmd)

	_, err = GetTerminalCommand("unknown", "/tmp", "true")
	assert.Error(t, err)
}
