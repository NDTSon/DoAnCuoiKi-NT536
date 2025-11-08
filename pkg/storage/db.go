package storage

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"
)

func NewDB(connString string) (*sql.DB, error) {
	// Prefer driver from URL scheme
	lower := strings.ToLower(connString)
	if strings.HasPrefix(lower, "postgres://") || strings.HasPrefix(lower, "postgresql://") || strings.HasPrefix(lower, "pgx://") {
		db, err := sql.Open("pgx", connString)
		if err == nil {
			if err = db.Ping(); err == nil {
				return db, nil
			}
			db.Close()
		}
		// Fallback to SQLite local if Postgres unreachable
		return openOrInitSQLite("file:data/dev.db?_pragma=busy_timeout(5000)")
	}

	if strings.HasPrefix(lower, "sqlite://") || strings.HasPrefix(lower, "file:") || strings.HasSuffix(lower, ".db") {
		return openOrInitSQLite(connString)
	}

	// Default: try pgx, else fallback to sqlite
	if db, err := sql.Open("pgx", connString); err == nil {
		if err = db.Ping(); err == nil {
			return db, nil
		}
		db.Close()
	}
	return openOrInitSQLite("file:data/dev.db?_pragma=busy_timeout(5000)")
}

func openOrInitSQLite(conn string) (*sql.DB, error) {
	// Extract file path from connection string
	// Handle formats like: "file:data/dev.db?_pragma=..." or "data/dev.db"
	filePath := conn
	if strings.HasPrefix(conn, "file:") {
		// Remove "file:" prefix and query parameters
		parts := strings.Split(conn, "?")
		filePath = strings.TrimPrefix(parts[0], "file:")
	}
	
	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
	}
	
	// Use absolute path to avoid issues
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, err
	}
	
	// Reconstruct connection string with absolute path
	if strings.HasPrefix(conn, "file:") {
		// Keep query parameters if any
		parts := strings.SplitN(conn, "?", 2)
		if len(parts) > 1 {
			conn = "file:" + absPath + "?" + parts[1]
		} else {
			conn = "file:" + absPath
		}
	} else {
		conn = absPath
	}
	
	db, err := sql.Open("sqlite", conn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	if err = ensureUserSchema(db); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func ensureUserSchema(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS users (
  id TEXT PRIMARY KEY,
  email TEXT NOT NULL UNIQUE,
  password_hash BLOB NOT NULL,
  display_name TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
)`)
	if err != nil {
		return err
	}
	// Ensure id is generated if not provided (for sqlite we can use randomblob)
	// Inserts from repo pass id via RETURNING in Postgres; for sqlite we let app layer treat id as TEXT and set via db
	return nil
}
