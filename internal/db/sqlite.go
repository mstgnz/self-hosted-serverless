package db

import (
	"database/sql"
	"os"
	"path/filepath"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

var (
	sqliteOnce sync.Once
	sqliteDB   *sql.DB
	sqliteErr  error
)

// GetSQLiteDB returns a singleton SQLite database connection
func GetSQLiteDB() (*sql.DB, error) {
	sqliteOnce.Do(func() {
		// Get database file path from environment variable or use default
		dbPath := getEnv("SQLITE_DB_PATH", "data/serverless.db")

		// Create the directory if it doesn't exist
		dir := filepath.Dir(dbPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			sqliteErr = err
			return
		}

		// Open the database connection
		sqliteDB, sqliteErr = sql.Open("sqlite3", dbPath)
		if sqliteErr != nil {
			return
		}

		// Test the connection
		sqliteErr = sqliteDB.Ping()
	})

	return sqliteDB, sqliteErr
}

// CloseSQLiteDB closes the SQLite database connection
func CloseSQLiteDB() error {
	if sqliteDB != nil {
		return sqliteDB.Close()
	}
	return nil
}
