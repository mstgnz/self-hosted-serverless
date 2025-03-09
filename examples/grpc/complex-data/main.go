package main

import (
	"errors"
	"fmt"

	"github.com/mstgnz/self-hosted-serverless/internal/common"
)

// Handler is the function handler
var Handler = &FunctionHandler{}

// Info contains metadata about the function
var Info = common.FunctionInfo{
	Name:        "grpc-complex",
	Description: "A gRPC function that handles complex data types",
	Runtime:     "go",
}

// FunctionHandler implements the serverless function
type FunctionHandler struct{}

// Execute executes the function with the given input
func (h *FunctionHandler) Execute(input map[string]interface{}) (interface{}, error) {
	// Extract nested data
	userValue, ok := input["user"]
	if !ok {
		return nil, errors.New("user data is required")
	}

	user, ok := userValue.(map[string]interface{})
	if !ok {
		return nil, errors.New("user must be an object")
	}

	// Extract array data
	itemsValue, ok := input["items"]
	if !ok {
		return nil, errors.New("items array is required")
	}

	items, ok := itemsValue.([]interface{})
	if !ok {
		return nil, errors.New("items must be an array")
	}

	// Process the data
	processedItems := make([]interface{}, len(items))
	for i, item := range items {
		if str, ok := item.(string); ok {
			processedItems[i] = fmt.Sprintf("Processed: %s", str)
		} else {
			processedItems[i] = item
		}
	}

	return map[string]interface{}{
		"user_processed": map[string]interface{}{
			"name": user["name"],
			"id":   user["id"],
		},
		"item_count":      len(items),
		"items_processed": processedItems,
	}, nil
}
