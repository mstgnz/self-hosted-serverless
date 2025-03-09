package main

import (
	"fmt"

	"github.com/mstgnz/self-hosted-serverless/internal/common"
)

// Handler is the function handler
var Handler = &FunctionHandler{}

// Info contains metadata about the function
var Info = common.FunctionInfo{
	Name:        "hello",
	Description: "A simple hello world function",
	Runtime:     "go",
}

// FunctionHandler implements the serverless function
type FunctionHandler struct{}

// Execute executes the function with the given input
func (h *FunctionHandler) Execute(input map[string]any) (any, error) {
	name := "World"
	if n, ok := input["name"].(string); ok {
		name = n
	}

	return map[string]any{
		"message": fmt.Sprintf("Hello, %s!", name),
		"input":   input,
	}, nil
}
