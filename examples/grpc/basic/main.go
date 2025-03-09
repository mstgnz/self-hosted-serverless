package main

import (
	"fmt"

	"github.com/mstgnz/self-hosted-serverless/internal/common"
)

// Handler is the function handler
var Handler = &FunctionHandler{}

// Info contains metadata about the function
var Info = common.FunctionInfo{
	Name:        "grpc-basic",
	Description: "A basic gRPC function example",
	Runtime:     "go",
}

// FunctionHandler implements the serverless function
type FunctionHandler struct{}

// Execute executes the function with the given input
func (h *FunctionHandler) Execute(input map[string]interface{}) (interface{}, error) {
	// Extract name from input, default to "World"
	name := "World"
	if n, ok := input["name"].(string); ok {
		name = n
	}

	// Return a greeting message
	return map[string]interface{}{
		"message": fmt.Sprintf("Hello, %s from gRPC!", name),
	}, nil
}
