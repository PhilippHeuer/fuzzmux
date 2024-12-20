package firefox

import (
	"context"
	"fmt"
	"github.com/PhilippHeuer/fuzzmux/pkg/database/sqlite"
	"github.com/PhilippHeuer/fuzzmux/pkg/recon"
	"github.com/cidverse/go-ptr"

	_ "github.com/mattn/go-sqlite3"
)

const moduleType = "firefox"

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

	// DbFile points to the local Firefox SQLite database
	DbFile string `yaml:"db-file"`
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
	var result []recon.Option

	// db connect
	dbConn, err := sqlite.NewDB(p.Config.DbFile)
	if err != nil {
		return nil, err
	}

	// db operator
	dbClient := NewDatabaseOperator(dbConn)

	// query bookmarks
	bookmarks, err := dbClient.GetBookmarks(context.Background())
	if err != nil {
		return nil, err
	}
	for _, bookmark := range bookmarks {
		result = append(result, recon.Option{
			ProviderName:   moduleType,
			ProviderType:   moduleType,
			Id:             fmt.Sprintf("%d", bookmark.ID),
			DisplayName:    bookmark.Title,
			Name:           bookmark.Title,
			Description:    ptr.Value(bookmark.URL),
			StartDirectory: p.Config.StartDirectory,
			Tags:           []string{"firefox", "bookmark"},
			Context: map[string]string{
				"url":    ptr.Value(bookmark.URL),
				"title":  bookmark.Title,
				"folder": bookmark.Folder,
				"id":     fmt.Sprintf("%d", bookmark.ID),
				"parent": fmt.Sprintf("%d", bookmark.Parent),
			},
		})
	}

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
	return Module{
		Config: config,
	}
}
