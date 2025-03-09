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
	Name:        "grpc-error",
	Description: "A gRPC function that demonstrates error handling",
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
			"message": "Operation completed successfully via gRPC",
		}, nil
	case "error":
		return nil, errors.New("operation failed via gRPC")
	case "panic":
		// This will be caught by the gRPC server and returned as an error
		panic("This is a simulated panic in the gRPC function")
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}
