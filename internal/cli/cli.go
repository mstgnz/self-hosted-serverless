package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"text/template"
)

const (
	functionTemplate = `package main

import (
	"fmt"
	
	"github.com/mstgnz/self-hosted-serverless/internal/common"
)

// Handler is the function handler
var Handler = &FunctionHandler{}

// Info contains metadata about the function
var Info = common.FunctionInfo{
	Name:        "{{.Name}}",
	Description: "{{.Description}}",
	Runtime:     "go",
}

// FunctionHandler implements the serverless function
type FunctionHandler struct{}

// Execute executes the function with the given input
func (h *FunctionHandler) Execute(input map[string]any) (any, error) {
	// TODO: Implement your function logic here
	return map[string]any{
		"message": fmt.Sprintf("Hello from %s function!", Info.Name),
		"input":   input,
	}, nil
}
`
)

// CreateFunction creates a new serverless function
func CreateFunction(name string) {
	// Create the functions directory if it doesn't exist
	functionsDir := "functions"
	if _, err := os.Stat(functionsDir); os.IsNotExist(err) {
		if err := os.MkdirAll(functionsDir, 0755); err != nil {
			log.Fatalf("Failed to create functions directory: %v", err)
		}
	}

	// Create the function directory
	functionDir := filepath.Join(functionsDir, name)
	if _, err := os.Stat(functionDir); os.IsNotExist(err) {
		if err := os.MkdirAll(functionDir, 0755); err != nil {
			log.Fatalf("Failed to create function directory: %v", err)
		}
	}

	// Create the main.go file
	mainFile := filepath.Join(functionDir, "main.go")
	if _, err := os.Stat(mainFile); !os.IsNotExist(err) {
		log.Fatalf("Function %s already exists", name)
	}

	// Parse the template
	tmpl, err := template.New("function").Parse(functionTemplate)
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}

	// Create the file
	file, err := os.Create(mainFile)
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}
	defer file.Close()

	// Execute the template
	err = tmpl.Execute(file, struct {
		Name        string
		Description string
	}{
		Name:        name,
		Description: fmt.Sprintf("A serverless function named %s", name),
	})
	if err != nil {
		log.Fatalf("Failed to execute template: %v", err)
	}

	fmt.Printf("Created function %s in %s\n", name, mainFile)
	fmt.Println("To build the function, run:")
	fmt.Printf("  cd %s && go build -buildmode=plugin -o %s.so\n", functionDir, name)
}

// RunFunction runs a serverless function locally
func RunFunction(name string) {
	// Prepare the request body
	requestBody, err := json.Marshal(map[string]any{
		"key": "value",
	})
	if err != nil {
		log.Fatalf("Failed to marshal request body: %v", err)
	}

	// Send the request to the server
	resp, err := http.Post(fmt.Sprintf("http://localhost:8080/run/%s", name), "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response: %v", err)
	}

	// Print the response
	fmt.Println(string(body))
}

// ListFunctions lists all available serverless functions
func ListFunctions() {
	// Send the request to the server
	resp, err := http.Get("http://localhost:8080/functions")
	if err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response: %v", err)
	}

	// Parse the response
	var result struct {
		Functions []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Runtime     string `json:"runtime"`
		} `json:"functions"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Fatalf("Failed to parse response: %v", err)
	}

	// Print the functions
	fmt.Println("Available functions:")
	for _, function := range result.Functions {
		fmt.Printf("  %s - %s (%s)\n", function.Name, function.Description, function.Runtime)
	}
}
