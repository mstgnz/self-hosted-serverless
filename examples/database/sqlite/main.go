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
	Name:        "sqlite-example",
	Description: "A function that demonstrates SQLite integration",
	Runtime:     "go",
}

// FunctionHandler implements the serverless function
type FunctionHandler struct{}

// Execute executes the function with the given input
func (h *FunctionHandler) Execute(input map[string]interface{}) (interface{}, error) {
	// Get SQLite connection
	db, err := db.GetSQLiteDB()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SQLite: %w", err)
	}

	// Get limit from input, default to 10
	limit := 10
	if l, ok := input["limit"].(float64); ok {
		limit = int(l)
	}

	// Execute a query
	rows, err := db.Query("SELECT id, title, content FROM notes WHERE id > ? LIMIT ?", 0, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Process results
	var notes []map[string]interface{}
	for rows.Next() {
		var id int
		var title, content string
		if err := rows.Scan(&id, &title, &content); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		notes = append(notes, map[string]interface{}{
			"id":      id,
			"title":   title,
			"content": content,
		})
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return map[string]interface{}{
		"notes": notes,
	}, nil
}
