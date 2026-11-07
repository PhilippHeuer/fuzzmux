package desktop

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/PhilippHeuer/fuzzmux/pkg/recon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testModule(t *testing.T, config ModuleConfig) Module {
	t.Helper()
	m := NewModule(config)
	m.searchDirs = []string{
		filepath.Join("testdata", "user", "applications"),
		filepath.Join("testdata", "system", "applications"),
	}
	return m
}

func optionIds(options []recon.Option) []string {
	var ids []string
	for _, o := range options {
		ids = append(ids, o.Id)
	}
	return ids
}

func TestDesktopOptions(t *testing.T) {
	t.Setenv("LC_ALL", "C")
	t.Setenv("XDG_CURRENT_DESKTOP", "Hyprland")

	options, err := testModule(t, ModuleConfig{}).Options()
	require.NoError(t, err)

	// masked (Hidden in user dir), nodisplay, kde-only, tryexec and link entries are filtered
	assert.Equal(t, []string{"desktop-kde-dolphin", "desktop-firefox", "desktop-htop"}, optionIds(options))

	firefox := options[1]
	assert.Equal(t, "Firefox", firefox.Name)
	assert.Equal(t, "Firefox (Web Browser)", firefox.DisplayName)
	assert.Equal(t, "Browse the Web", firefox.Description)
	assert.Equal(t, "firefox --name 'my firefox'", firefox.Context["exec"])
	assert.Equal(t, "false", firefox.Context["terminal"])
	assert.Equal(t, "Network;WebBrowser", firefox.Context["categories"])
	assert.Equal(t, []string{"desktop", "gui"}, firefox.Tags)

	// user entry overrides system entry with the same id
	htop := options[2]
	assert.Equal(t, "Htop (User)", htop.Name)
	assert.Equal(t, "htop --tree", htop.Context["exec"])
	assert.Equal(t, []string{"desktop", "terminal"}, htop.Tags)

	// nested directories are part of the desktop file id, file codes are dropped
	assert.Equal(t, "dolphin", options[0].Context["exec"])
}

func TestDesktopOptionsFilters(t *testing.T) {
	t.Setenv("LC_ALL", "C")
	t.Setenv("XDG_CURRENT_DESKTOP", "ubuntu:KDE")

	options, err := testModule(t, ModuleConfig{ShowHidden: true}).Options()
	require.NoError(t, err)
	assert.Contains(t, optionIds(options), "desktop-nodisplay")
	assert.Contains(t, optionIds(options), "desktop-kde-only")
	assert.NotContains(t, optionIds(options), "desktop-masked")

	options, err = testModule(t, ModuleConfig{Categories: []string{"network"}}).Options()
	require.NoError(t, err)
	assert.Equal(t, []string{"desktop-firefox"}, optionIds(options))

	options, err = testModule(t, ModuleConfig{ExcludeCategories: []string{"Monitor"}, Exclude: []string{"firefox.desktop", "kde-dolphin"}}).Options()
	require.NoError(t, err)
	assert.Equal(t, []string{"desktop-kde-only"}, optionIds(options))
}

func TestDesktopLocalization(t *testing.T) {
	t.Setenv("LC_ALL", "de_DE.UTF-8")
	t.Setenv("XDG_CURRENT_DESKTOP", "")

	options, err := testModule(t, ModuleConfig{Categories: []string{"Network"}, DisplayName: "{{name}} - {{genericName}}"}).Options()
	require.NoError(t, err)
	require.Len(t, options, 1)
	assert.Equal(t, "Firefox DE", options[0].Name)
	assert.Equal(t, "Firefox DE - Webbrowser", options[0].DisplayName)
}

func TestParseEntryIgnoresOtherGroups(t *testing.T) {
	entry, err := ParseEntry(strings.NewReader("[Desktop Entry]\nName=App\nExec=app\n[Desktop Action x]\nName=Other\nExec=other\n"), nil)
	require.NoError(t, err)
	assert.Equal(t, "App", entry.Name)
	assert.Equal(t, "app", entry.Exec)
}

func TestParseExec(t *testing.T) {
	tests := []struct {
		exec string
		want []string
	}{
		{"firefox %u", []string{"firefox"}},
		{"code --new-window %F", []string{"code", "--new-window"}},
		{`app "arg with space" --x=%f`, []string{"app", "arg with space", "--x="}},
		{`sh -c "echo \"hi\" \$HOME"`, []string{"sh", "-c", `echo "hi" $HOME`}},
		{"app %i %c %k 100%%", []string{"app", "--icon", "icon", "My App", "/path/app.desktop", "100%"}},
		{"app   --double-space", []string{"app", "--double-space"}},
		{`app ""`, []string{"app", ""}},
	}
	for _, tt := range tests {
		got, err := ParseExec(tt.exec, "My App", "icon", "/path/app.desktop")
		require.NoError(t, err, tt.exec)
		assert.Equal(t, tt.want, got, tt.exec)
	}

	_, err := ParseExec(`app "unterminated`, "", "", "")
	assert.Error(t, err)
	_, err = ParseExec("%f", "", "", "")
	assert.Error(t, err)
}

func TestShellCommand(t *testing.T) {
	assert.Equal(t, `sh -c 'echo "hi" $HOME' 'it'\''s' ''`, ShellCommand([]string{"sh", "-c", `echo "hi" $HOME`, "it's", ""}))
	assert.Equal(t, "env FOO=bar /usr/bin/app --flag=1", ShellCommand([]string{"env", "FOO=bar", "/usr/bin/app", "--flag=1"}))
}

func TestLocaleCandidates(t *testing.T) {
	t.Setenv("LC_ALL", "")
	t.Setenv("LC_MESSAGES", "")
	t.Setenv("LANG", "sr_RS.UTF-8@latin")
	assert.Equal(t, []string{"sr_RS@latin", "sr_RS", "sr@latin", "sr"}, localeCandidates())

	t.Setenv("LANG", "C.UTF-8")
	assert.Nil(t, localeCandidates())
}
