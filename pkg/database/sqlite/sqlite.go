package sqlite

import (
	"database/sql"
	"fmt"
)

type DBConnection interface {
	Query(query string, args ...any) (*sql.Rows, error)
}

const sqliteDriverName = `sqlite3`

func NewDB(dbFilename string) (DBConnection, error) {
	connectionString := fmt.Sprintf("file:%s?mode=ro&cache=shared&immutable=1", dbFilename)
	conn, err := sql.Open(sqliteDriverName, connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite DB: %v", err)
	}
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping SQLite DB: %v", err)
	}

	return conn, nil
}
