package store

import (
	"database/sql"
	"strings"

	_ "modernc.org/sqlite"
)

// Open opens path read-only so we never block WAL writers and never create
// -wal/-shm files next to the database.
func Open(path string) (*sql.DB, error) {
	dsn := "file:" + escapeURI(path) + "?mode=ro"
	return sql.Open("sqlite", dsn)
}

func escapeURI(path string) string {
	r := strings.NewReplacer("%", "%25", "?", "%3F", "#", "%23")
	return r.Replace(path)
}
