package firefox

import (
	"context"
	"fmt"
	"github.com/PhilippHeuer/fuzzmux/pkg/database/sqlite"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon"
	"github.com/adrg/xdg"
	"github.com/cidverse/go-ptr"
	"path"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

const moduleType = "firefox"

var dbFilePaths = []string{
	filepath.Join(".mozilla/firefox", "*.default*"),
	filepath.Join(".librewolf", "*.default*"),
	filepath.Join(".floorp", "*.default*"),
	filepath.Join(".zen", "*.default*"),
}

type Module struct {
	Config ModuleConfig
}

type ModuleConfig struct {
	// Name is used to override the default module name
	Name string `yaml:"name,omitempty"`

	// DisplayName is a template string to render a custom display name
	DisplayName string `yaml:"display-name"`

	// StartDirectory is a template string that defines the start directory
	StartDirectory string `yaml:"start-directory"`

	// ProfilePath is the path to the Firefox profile
	ProfilePath string `yaml:"profile-path"`
}

func (c *ModuleConfig) DecodeConfig() {
	c.ProfilePath = strings.Replace(c.ProfilePath, "~", xdg.Home, 1)
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
	p.Config.DecodeConfig()
	var result []recon.Option

	// db connect
	dbConn, err := sqlite.NewDB(path.Join(p.Config.ProfilePath, "places.sqlite"))
	if err != nil {
		return nil, fmt.Errorf("error opening db file %s: %v", p.Config.ProfilePath, err)
	}

	// db operator
	dbClient := NewDatabaseOperator(dbConn)

	// query bookmarks
	bookmarks, err := dbClient.GetBookmarks(context.Background())
	if err != nil {
		return nil, err
	}
	for _, bookmark := range bookmarks {
		if ptr.Value(bookmark.URL) == "" {
			continue // skip directories
		}

		result = append(result, recon.Option{
			ProviderName:   p.Name(),
			ProviderType:   p.Type(),
			Id:             fmt.Sprintf("%d", bookmark.ID),
			DisplayName:    bookmark.Title,
			Name:           bookmark.Title,
			Description:    "",
			Web:            ptr.Value(bookmark.URL),
			StartDirectory: p.Config.StartDirectory,
			Tags:           []string{"firefox", "bookmark"},
			Context: map[string]string{
				"folder": bookmark.Folder,
				"parent": fmt.Sprintf("%d", bookmark.Parent),
			},
		})
	}

	// TODO: query history

	return result, nil
}

func (p Module) OptionsOrCache(maxAge float64) ([]recon.Option, error) {
	return recon.OptionsOrCache(p, maxAge)
}

func (p Module) SelectOption(option *recon.Option) error {
	err := option.CreateStartDirectoryIfMissing()
	if err != nil {
		return err
	}

	return nil
}

func (p Module) Columns() []recon.Column {
	return recon.DefaultColumns()
}

func NewModule(config ModuleConfig) Module {
	// discover db file, if only one profile exists
	if config.ProfilePath == "" {
		for _, p := range dbFilePaths {
			matches, err := filepath.Glob(path.Join(xdg.Home, p))
			if err != nil {
				continue
			}

			if len(matches) == 1 {
				config.ProfilePath = matches[0]
				break
			}
		}
	}

	return Module{
		Config: config,
	}
}
