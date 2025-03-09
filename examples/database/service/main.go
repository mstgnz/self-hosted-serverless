package main

import (
	"fmt"

	"github.com/mstgnz/self-hosted-serverless/internal/common"
	"github.com/mstgnz/self-hosted-serverless/internal/db"
)

// Handler is the function handler
var Handler = &FunctionHandler{}

// Info contains metadata about the function
var Info = common.FunctionInfo{
	Name:        "db-service-example",
	Description: "A function that demonstrates the database service",
	Runtime:     "go",
}

// FunctionHandler implements the serverless function
type FunctionHandler struct{}

// Execute executes the function with the given input
func (h *FunctionHandler) Execute(input map[string]interface{}) (interface{}, error) {
	// Get database type from input, default to PostgreSQL
	dbType := db.PostgreSQL
	if t, ok := input["db_type"].(string); ok {
		switch t {
		case "postgres":
			dbType = db.PostgreSQL
		case "sqlite":
			dbType = db.SQLite
		case "redis":
			dbType = db.Redis
		default:
			return nil, fmt.Errorf("unsupported database type: %s", t)
		}
	}

	// Get database service
	service, err := db.NewService(dbType)
	if err != nil {
		return nil, fmt.Errorf("failed to create database service: %w", err)
	}
	defer service.Close()

	// Get query from input
	query, ok := input["query"].(string)
	if !ok {
		return nil, fmt.Errorf("query is required and must be a string")
	}

	// Get args from input
	var args []interface{}
	if argsValue, ok := input["args"].([]interface{}); ok {
		args = argsValue
	}

	// Execute the query
	rows, err := service.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get column names: %w", err)
	}

	// Process results
	var results []map[string]interface{}
	for rows.Next() {
		// Create a slice of interface{} to hold the values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		// Scan the result into the values
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Create a map for this row
		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			// Handle null values
			if val == nil {
				row[col] = nil
			} else {
				// Convert to string for simplicity
				row[col] = fmt.Sprintf("%v", val)
			}
		}
		results = append(results, row)
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return map[string]interface{}{
		"results": results,
	}, nil
}
