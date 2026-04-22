package db

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// DatabaseType represents the type of database
type DatabaseType string

const (
	PostgreSQL DatabaseType = "postgres"
	SQLite     DatabaseType = "sqlite"
	Redis      DatabaseType = "redis"
)

// Service provides a unified interface for database operations
type Service struct {
	dbType      DatabaseType
	sqlDB       *sql.DB
	redisClient *redis.Client
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
		service.redisClient, err = GetRedisClient()
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return service, nil
}

// Close closes the database connection(s)
func (s *Service) Close() error {
	if s.sqlDB != nil {
		if err := s.sqlDB.Close(); err != nil {
			return err
		}
	}
	if s.redisClient != nil {
		if err := s.redisClient.Close(); err != nil {
			return err
		}
	}
	return nil
}

// Query executes a SQL query and returns the rows (SQL databases only)
func (s *Service) Query(query string, args ...any) (*sql.Rows, error) {
	if s.sqlDB == nil {
		return nil, errors.New("SQL database not initialized or using non-SQL database")
	}
	return s.sqlDB.Query(query, args...)
}

// Exec executes a SQL statement (SQL databases only)
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
func (s *Service) GetRedis() *redis.Client {
	return s.redisClient
}
