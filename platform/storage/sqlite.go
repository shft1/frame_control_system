package storage

import (
	"database/sql"
	"fmt"
)

// sqlite3 driver
import _ "github.com/mattn/go-sqlite3"

func OpenSQLite(path string) (*sql.DB, error) {
	dsn := fmt.Sprintf("file:%s?_busy_timeout=5000&_journal_mode=WAL&_fk=1", path)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}


