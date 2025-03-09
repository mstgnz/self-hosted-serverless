package main

import (
	"github.com/mstgnz/self-hosted-serverless/internal/common"
)

// Handler is the function handler
var Handler = &FunctionHandler{}

// Info contains metadata about the function
var Info = common.FunctionInfo{
	Name:        "custom-headers",
	Description: "A function that returns custom HTTP headers",
	Runtime:     "go",
}

// FunctionHandler implements the serverless function
type FunctionHandler struct{}

// Execute executes the function with the given input
func (h *FunctionHandler) Execute(input map[string]interface{}) (interface{}, error) {
	// Return both data and headers
	// The server will extract headers from the response and set them in the HTTP response
	return map[string]interface{}{
		"data": map[string]interface{}{
			"message":   "Hello, World!",
			"timestamp": input["timestamp"],
		},
		"headers": map[string]string{
			"X-Custom-Header": "Custom Value",
			"X-Powered-By":    "Self-Hosted Serverless",
			"X-Function-Name": "custom-headers",
			"Cache-Control":   "no-cache, no-store, must-revalidate",
		},
	}, nil
}
