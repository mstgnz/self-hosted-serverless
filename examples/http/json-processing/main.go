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
	Name:        "json-processing",
	Description: "A function that processes JSON data",
	Runtime:     "go",
}

// FunctionHandler implements the serverless function
type FunctionHandler struct{}

// Execute executes the function with the given input
func (h *FunctionHandler) Execute(input map[string]interface{}) (interface{}, error) {
	// Extract user data from input
	userData, ok := input["user"].(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid user data")
	}

	// Extract and validate name
	name, ok := userData["name"].(string)
	if !ok {
		return nil, errors.New("name is required and must be a string")
	}

	// Extract and validate age
	age, ok := userData["age"].(float64)
	if !ok {
		return nil, errors.New("age is required and must be a number")
	}

	// Process the data
	return map[string]interface{}{
		"greeting": fmt.Sprintf("Hello, %s!", name),
		"message":  fmt.Sprintf("You are %d years old.", int(age)),
		"adult":    age >= 18,
	}, nil
}
