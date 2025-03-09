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
	"time"
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

// GetMetrics gets metrics for all functions
func GetMetrics() {
	// Make a request to the metrics endpoint
	resp, err := http.Get("http://localhost:8080/metrics")
	if err != nil {
		log.Fatalf("Failed to get metrics: %v", err)
	}
	defer resp.Body.Close()

	// Parse the response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Fatalf("Failed to parse metrics response: %v", err)
	}

	// Print the metrics
	fmt.Println("Function Metrics:")
	fmt.Println("================")

	metrics, ok := result["metrics"].(map[string]interface{})
	if !ok {
		fmt.Println("No metrics available")
		return
	}

	if len(metrics) == 0 {
		fmt.Println("No functions have been executed yet")
		return
	}

	for name, metricData := range metrics {
		metric, ok := metricData.(map[string]interface{})
		if !ok {
			continue
		}

		fmt.Printf("Function: %s\n", name)
		fmt.Printf("  Executions: %v\n", metric["execution_count"])
		fmt.Printf("  Errors: %v\n", metric["error_count"])

		// Format average duration
		if avgDuration, ok := metric["average_duration"].(float64); ok {
			fmt.Printf("  Average Duration: %v\n", time.Duration(avgDuration))
		}

		// Format last execution time
		if lastExecStr, ok := metric["last_execution_time"].(string); ok {
			if lastExec, err := time.Parse(time.RFC3339, lastExecStr); err == nil {
				fmt.Printf("  Last Execution: %v\n", lastExec.Format(time.RFC3339))
			}
		}

		// Format cold start info
		if coldStarts, ok := metric["cold_start_count"].(float64); ok {
			fmt.Printf("  Cold Starts: %v\n", int64(coldStarts))
		}

		if avgColdStart, ok := metric["avg_cold_start_latency"].(float64); ok {
			fmt.Printf("  Average Cold Start Latency: %v\n", time.Duration(avgColdStart))
		}

		fmt.Println()
	}
}

// GetFunctionMetrics gets metrics for a specific function
func GetFunctionMetrics(functionName string) {
	// Make a request to the function metrics endpoint
	resp, err := http.Get(fmt.Sprintf("http://localhost:8080/metrics/%s", functionName))
	if err != nil {
		log.Fatalf("Failed to get metrics for function %s: %v", functionName, err)
	}
	defer resp.Body.Close()

	// Check if the function was found
	if resp.StatusCode == http.StatusNotFound {
		fmt.Printf("Function %s not found\n", functionName)
		return
	}

	// Parse the response
	var metric map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&metric); err != nil {
		log.Fatalf("Failed to parse metrics response: %v", err)
	}

	// Print the metrics
	fmt.Printf("Metrics for function: %s\n", functionName)
	fmt.Println("================")

	fmt.Printf("Executions: %v\n", metric["execution_count"])
	fmt.Printf("Errors: %v\n", metric["error_count"])

	// Format average duration
	if avgDuration, ok := metric["average_duration"].(float64); ok {
		fmt.Printf("Average Duration: %v\n", time.Duration(avgDuration))
	}

	// Format last execution time
	if lastExecStr, ok := metric["last_execution_time"].(string); ok {
		if lastExec, err := time.Parse(time.RFC3339, lastExecStr); err == nil {
			fmt.Printf("Last Execution: %v\n", lastExec.Format(time.RFC3339))
		}
	}

	// Format cold start info
	if coldStarts, ok := metric["cold_start_count"].(float64); ok {
		fmt.Printf("Cold Starts: %v\n", int64(coldStarts))
	}

	if avgColdStart, ok := metric["avg_cold_start_latency"].(float64); ok {
		fmt.Printf("Average Cold Start Latency: %v\n", time.Duration(avgColdStart))
	}
}
