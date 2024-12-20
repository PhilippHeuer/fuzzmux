package firefox

import (
	"context"
	"fmt"
	"github.com/PhilippHeuer/fuzzmux/pkg/database/sqlite"
)

type DatabaseOperator struct {
	db sqlite.DBConnection
}

type Bookmark struct {
	URL    *string `json:"url" yaml:"url"`
	Title  string  `json:"title" yaml:"title"`
	Folder string  `json:"folder" yaml:"folder"`
	ID     int     `json:"id" yaml:"id"`
	Parent *int    `json:"parent" yaml:"parent"`
}

type BookmarkOperator interface {
	GetBookmarks(context.Context) ([]Bookmark, error)
}

const (
	bookmarkQuery = `SELECT bookmarks.id, bookmarks.parent, places.URL, bookmarks.title
				FROM moz_places as places
				RIGHT JOIN moz_bookmarks as bookmarks
				ON places.id = bookmarks.fk`
)

func (d *DatabaseOperator) GetBookmarks(ctx context.Context) ([]Bookmark, error) {
	rows, err := d.db.Query(bookmarkQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query db: %v", err)
	}

	var bookmarks []Bookmark
	for rows.Next() {
		var bm Bookmark

		err = rows.Scan(&bm.ID, &bm.Parent, &bm.URL, &bm.Title)
		if err != nil {
			return nil, fmt.Errorf("failed to execute query: %v", err)
		}

		bookmarks = append(bookmarks, bm)
	}

	return bookmarks, nil
}

func NewDatabaseOperator(conn sqlite.DBConnection) BookmarkOperator {
	return &DatabaseOperator{
		db: conn,
	}
}
