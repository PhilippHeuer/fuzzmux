package desktop

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/PhilippHeuer/fuzzmux/pkg/recon"
	"github.com/PhilippHeuer/fuzzmux/pkg/util"
	"github.com/adrg/xdg"
	"github.com/rs/zerolog/log"
)

const moduleType = "desktop"

// tags used by the default layout to decide between launching the app directly or within a terminal
const (
	tagGUI      = "gui"
	tagTerminal = "terminal"
)

type Module struct {
	Config ModuleConfig

	// searchDirs are the applications directories in order of precedence
	searchDirs []string
}

type ModuleConfig struct {
	// Name is used to override the default module name
	Name string `yaml:"name,omitempty"`

	// DisplayName is a template string to render a custom display name
	DisplayName string `yaml:"display-name"`

	// StartDirectory is a template string that defines the start directory
	StartDirectory string `yaml:"start-directory"`

	// Directories are additional directories to search for .desktop files (lowest precedence)
	Directories []string `yaml:"directories"`

	// ShowHidden includes entries flagged with NoDisplay=true
	ShowHidden bool `yaml:"show-hidden"`

	// Categories only includes entries with at least one of the given categories (e.g. Development, Network)
	Categories []string `yaml:"categories"`

	// ExcludeCategories excludes entries with any of the given categories
	ExcludeCategories []string `yaml:"exclude-categories"`

	// Exclude is a list of desktop file ids to exclude (e.g. org.gnome.Settings or org.gnome.Settings.desktop)
	Exclude []string `yaml:"exclude"`
}

func (p Module) Name() string {
	if p.Config.Name != "" {
		return p.Config.Name
	}
	return moduleType
}

func (p Module) Type() string {
	return moduleType
}

func (p Module) Options() ([]recon.Option, error) {
	locales := localeCandidates()
	currentDesktops := splitDesktops(os.Getenv("XDG_CURRENT_DESKTOP"))

	var result []recon.Option
	for _, file := range discoverDesktopFiles(p.searchDirs) {
		entry, err := ParseFile(file.path, locales)
		if err != nil {
			log.Debug().Err(err).Str("file", file.path).Msg("skipping desktop file")
			continue
		}
		entry.ID = file.id

		if !p.include(entry, currentDesktops) {
			continue
		}

		args, err := ParseExec(entry.Exec, entry.Name, entry.Icon, entry.Path)
		if err != nil {
			log.Debug().Err(err).Str("file", entry.Path).Msg("skipping desktop file with invalid exec value")
			continue
		}

		result = append(result, p.toOption(entry, ShellCommand(args)))
	}

	slices.SortFunc(result, func(a, b recon.Option) int {
		return strings.Compare(strings.ToLower(a.DisplayName), strings.ToLower(b.DisplayName))
	})

	return result, nil
}

func (p Module) include(entry *Entry, currentDesktops []string) bool {
	if entry.Type != "Application" || entry.Hidden || entry.Exec == "" || entry.Name == "" {
		return false
	}
	if entry.NoDisplay && !p.Config.ShowHidden {
		return false
	}
	if len(entry.OnlyShowIn) > 0 && !intersects(entry.OnlyShowIn, currentDesktops) {
		return false
	}
	if intersects(entry.NotShowIn, currentDesktops) {
		return false
	}
	if len(p.Config.Categories) > 0 && !intersects(entry.Categories, p.Config.Categories) {
		return false
	}
	if intersects(entry.Categories, p.Config.ExcludeCategories) {
		return false
	}
	if slices.Contains(p.Config.Exclude, entry.ID) || slices.Contains(p.Config.Exclude, strings.TrimSuffix(entry.ID, ".desktop")) {
		return false
	}
	if entry.TryExec != "" && !executableExists(entry.TryExec) {
		return false
	}
	return true
}

func (p Module) toOption(entry *Entry, command string) recon.Option {
	id := strings.TrimSuffix(entry.ID, ".desktop")

	displayName := entry.Name
	if entry.GenericName != "" && entry.GenericName != entry.Name {
		displayName += " (" + entry.GenericName + ")"
	}

	tags := []string{moduleType, tagGUI}
	if entry.Terminal {
		tags = []string{moduleType, tagTerminal}
	}

	startDirectory := entry.WorkDir
	if startDirectory == "" {
		startDirectory = "~"
	}

	op := recon.Option{
		ProviderName:   p.Name(),
		ProviderType:   p.Type(),
		Id:             p.Name() + "-" + id,
		Name:           entry.Name,
		DisplayName:    displayName,
		Description:    entry.Comment,
		StartDirectory: startDirectory,
		Tags:           tags,
		Context: map[string]string{
			"desktopId":   id,
			"desktopFile": entry.Path,
			"exec":        command,
			"genericName": entry.GenericName,
			"icon":        entry.Icon,
			"categories":  strings.Join(entry.Categories, ";"),
			"keywords":    strings.Join(entry.Keywords, ";"),
			"terminal":    boolString(entry.Terminal),
		},
	}
	op.ProcessUserTemplateStrings(p.Config.DisplayName, p.Config.StartDirectory)

	return op
}

func (p Module) OptionsOrCache(maxAge float64) ([]recon.Option, error) {
	return recon.OptionsOrCache(p, maxAge)
}

func (p Module) SelectOption(option *recon.Option) error {
	return option.CreateStartDirectoryIfMissing()
}

func (p Module) Columns() []recon.Column {
	return recon.DefaultColumns()
}

func NewModule(config ModuleConfig) Module {
	return Module{
		Config:     config,
		searchDirs: defaultSearchDirs(config.Directories),
	}
}

// defaultSearchDirs returns $XDG_DATA_HOME/applications, $XDG_DATA_DIRS/*/applications and the user-provided directories, in order of precedence
func defaultSearchDirs(extra []string) []string {
	var dirs []string
	for _, d := range append([]string{xdg.DataHome}, xdg.DataDirs...) {
		dirs = append(dirs, filepath.Join(d, "applications"))
	}
	for _, d := range extra {
		dirs = append(dirs, util.ExpandTilde(os.ExpandEnv(d)))
	}
	return dirs
}

type desktopFile struct {
	id   string
	path string
}

// discoverDesktopFiles finds all .desktop files, the first file for each desktop file id wins (user entries override system entries)
func discoverDesktopFiles(dirs []string) []desktopFile {
	var result []desktopFile
	seen := map[string]bool{}

	for _, dir := range dirs {
		_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil // missing dir or unreadable subdir
			}
			if d.IsDir() || !strings.HasSuffix(path, ".desktop") {
				return nil
			}

			rel, err := filepath.Rel(dir, path)
			if err != nil {
				return nil
			}
			id := strings.ReplaceAll(rel, string(filepath.Separator), "-")
			if seen[id] {
				return nil
			}
			seen[id] = true

			result = append(result, desktopFile{id: id, path: path})
			return nil
		})
	}

	return result
}

func executableExists(name string) bool {
	if filepath.IsAbs(name) {
		info, err := os.Stat(name)
		return err == nil && !info.IsDir() && info.Mode()&0111 != 0
	}
	_, err := exec.LookPath(name)
	return err == nil
}

func splitDesktops(value string) []string {
	if value == "" {
		return nil
	}
	return strings.Split(value, ":")
}

func intersects(a, b []string) bool {
	for _, v := range a {
		for _, w := range b {
			if strings.EqualFold(v, w) {
				return true
			}
		}
	}
	return false
}

func boolString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
