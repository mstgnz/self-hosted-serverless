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
	Name:        "error-handling",
	Description: "A function that demonstrates error handling",
	Runtime:     "go",
}

// FunctionHandler implements the serverless function
type FunctionHandler struct{}

// Execute executes the function with the given input
func (h *FunctionHandler) Execute(input map[string]interface{}) (interface{}, error) {
	// Check if required fields are present
	actionValue, ok := input["action"]
	if !ok {
		return nil, errors.New("action field is required")
	}

	// Type assertion
	action, ok := actionValue.(string)
	if !ok {
		return nil, errors.New("action must be a string")
	}

	// Handle different actions
	switch action {
	case "success":
		return map[string]interface{}{
			"status":  "success",
			"message": "Operation completed successfully",
		}, nil
	case "error":
		return nil, errors.New("operation failed")
	case "validation":
		// Simulate a validation error
		if _, ok := input["data"]; !ok {
			return nil, errors.New("data field is required for validation action")
		}
		return map[string]interface{}{
			"status":  "validated",
			"message": "Data validation passed",
		}, nil
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}
