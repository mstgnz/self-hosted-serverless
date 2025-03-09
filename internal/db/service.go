package db

import (
	"database/sql"
	"errors"
	"fmt"
)

// DatabaseType represents the type of database
type DatabaseType string

const (
	// PostgreSQL database type
	PostgreSQL DatabaseType = "postgres"
	// SQLite database type
	SQLite DatabaseType = "sqlite"
	// Redis database type (not SQL)
	Redis DatabaseType = "redis"
)

// Service provides a unified interface for database operations
type Service struct {
	dbType      DatabaseType
	sqlDB       *sql.DB
	redisClient any // We'll use any for now since we're not fully implementing Redis
}

// NewService creates a new database service
func NewService(dbType DatabaseType) (*Service, error) {
	service := &Service{
		dbType: dbType,
	}

	var err error
	switch dbType {
	case PostgreSQL:
		service.sqlDB, err = GetPostgresDB()
	case SQLite:
		service.sqlDB, err = GetSQLiteDB()
	case Redis:
		// Redis is not SQL-based, so we handle it differently
		service.redisClient, err = GetRedisClient()
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return service, nil
}

// Close closes the database connection
func (s *Service) Close() error {
	if s.sqlDB != nil {
		return s.sqlDB.Close()
	}
	// Handle Redis close if needed
	return nil
}

// Query executes a SQL query and returns the rows (for SQL databases only)
func (s *Service) Query(query string, args ...any) (*sql.Rows, error) {
	if s.sqlDB == nil {
		return nil, errors.New("SQL database not initialized or using non-SQL database")
	}
	return s.sqlDB.Query(query, args...)
}

// Exec executes a SQL statement (for SQL databases only)
func (s *Service) Exec(query string, args ...any) (sql.Result, error) {
	if s.sqlDB == nil {
		return nil, errors.New("SQL database not initialized or using non-SQL database")
	}
	return s.sqlDB.Exec(query, args...)
}

// GetDB returns the underlying SQL database connection
func (s *Service) GetDB() *sql.DB {
	return s.sqlDB
}

// GetRedis returns the underlying Redis client
func (s *Service) GetRedis() any {
	return s.redisClient
}
