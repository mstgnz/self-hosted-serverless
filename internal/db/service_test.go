package db

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// TestNewService tests the creation of a new database service
func TestNewService(t *testing.T) {
	// This is a basic test that doesn't actually connect to a database
	// In a real test, we would use a mock or an in-memory database

	// Test with an unsupported database type
	service, err := NewService("unsupported")
	assert.Error(t, err)
	assert.Nil(t, service)
}

// TestServiceWithMockDB tests the database service with a mock database
func TestServiceWithMockDB(t *testing.T) {
	// Create a mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	// Create a service with the mock database
	service := &Service{
		dbType: PostgreSQL,
		sqlDB:  db,
	}

	// Test Query method
	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(1, "test1").
		AddRow(2, "test2")

	mock.ExpectQuery("SELECT (.+) FROM users").WillReturnRows(rows)

	result, err := service.Query("SELECT * FROM users")
	assert.NoError(t, err)

	// Verify the results
	var id int
	var name string

	// First row
	assert.True(t, result.Next())
	err = result.Scan(&id, &name)
	assert.NoError(t, err)
	assert.Equal(t, 1, id)
	assert.Equal(t, "test1", name)

	// Second row
	assert.True(t, result.Next())
	err = result.Scan(&id, &name)
	assert.NoError(t, err)
	assert.Equal(t, 2, id)
	assert.Equal(t, "test2", name)

	// No more rows
	assert.False(t, result.Next())

	// Test Exec method
	mock.ExpectExec("INSERT INTO users").WillReturnResult(sqlmock.NewResult(1, 1))

	_, err = service.Exec("INSERT INTO users (name) VALUES (?)", "test3")
	assert.NoError(t, err)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestGetDB tests the GetDB method
func TestGetDB(t *testing.T) {
	// Create a mock database
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	// Create a service with the mock database
	service := &Service{
		dbType: PostgreSQL,
		sqlDB:  db,
	}

	// Test GetDB method
	assert.Equal(t, db, service.GetDB())
}

// TestGetRedis tests the GetRedis method
func TestGetRedis(t *testing.T) {
	// Create a service with a mock Redis client
	mockRedisClient := "mock-redis-client"
	service := &Service{
		dbType:      Redis,
		redisClient: mockRedisClient,
	}

	// Test GetRedis method
	assert.Equal(t, mockRedisClient, service.GetRedis())
}

// TestClose tests the Close method
func TestClose(t *testing.T) {
	// Create a mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}

	// Expect Close to be called
	mock.ExpectClose()

	// Create a service with the mock database
	service := &Service{
		dbType: PostgreSQL,
		sqlDB:  db,
	}

	// Test Close method
	err = service.Close()
	assert.NoError(t, err)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())

	// Test Close with nil database
	service = &Service{
		dbType: Redis,
	}

	err = service.Close()
	assert.NoError(t, err)
}
