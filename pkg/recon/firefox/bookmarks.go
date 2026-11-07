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
	defer rows.Close()

	var bookmarks []Bookmark
	for rows.Next() {
		var bm Bookmark

		err = rows.Scan(&bm.ID, &bm.Parent, &bm.URL, &bm.Title)
		if err != nil {
			return nil, fmt.Errorf("failed to execute query: %v", err)
		}

		bookmarks = append(bookmarks, bm)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to read query results: %v", err)
	}

	// resolve folder names, entries without url are folders
	folders := make(map[int]string)
	for _, bm := range bookmarks {
		if bm.URL == nil {
			folders[bm.ID] = bm.Title
		}
	}
	for i, bm := range bookmarks {
		if bm.Parent != nil {
			bookmarks[i].Folder = folders[*bm.Parent]
		}
	}

	return bookmarks, nil
}

func NewDatabaseOperator(conn sqlite.DBConnection) BookmarkOperator {
	return &DatabaseOperator{
		db: conn,
	}
}
