package db

import (
	"database/sql"
	"fmt"
	"os"
	"sync"

	_ "github.com/lib/pq"
)

var (
	postgresOnce sync.Once
	postgresDB   *sql.DB
	postgresErr  error
)

// GetPostgresDB returns a singleton PostgreSQL database connection
func GetPostgresDB() (*sql.DB, error) {
	postgresOnce.Do(func() {
		// Get connection parameters from environment variables
		host := getEnv("POSTGRES_HOST", "localhost")
		port := getEnv("POSTGRES_PORT", "5432")
		user := getEnv("POSTGRES_USER", "postgres")
		password := getEnv("POSTGRES_PASSWORD", "postgres")
		dbname := getEnv("POSTGRES_DB", "serverless")

		// Create the connection string
		connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			host, port, user, password, dbname)

		// Open the database connection
		postgresDB, postgresErr = sql.Open("postgres", connStr)
		if postgresErr != nil {
			return
		}

		// Test the connection
		postgresErr = postgresDB.Ping()
	})

	return postgresDB, postgresErr
}

// ClosePostgresDB closes the PostgreSQL database connection
func ClosePostgresDB() error {
	if postgresDB != nil {
		return postgresDB.Close()
	}
	return nil
}

// Helper function to get environment variables with default values
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
