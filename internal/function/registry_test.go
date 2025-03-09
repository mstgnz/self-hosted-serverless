package function

import (
	"fmt"
	"testing"

	"github.com/mstgnz/self-hosted-serverless/internal/common"
	"github.com/stretchr/testify/assert"
)

// MockFunctionHandler is a mock implementation of FunctionHandler for testing
type MockFunctionHandler struct {
	ExecuteFunc func(input map[string]interface{}) (interface{}, error)
}

// Execute calls the mock ExecuteFunc
func (m *MockFunctionHandler) Execute(input map[string]interface{}) (interface{}, error) {
	return m.ExecuteFunc(input)
}

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()
	assert.NotNil(t, registry)
	assert.NotNil(t, registry.functions)
	assert.NotNil(t, registry.metadata)
}

func TestRegister(t *testing.T) {
	registry := NewRegistry()

	// Create a mock handler
	mockHandler := &MockFunctionHandler{
		ExecuteFunc: func(input map[string]interface{}) (interface{}, error) {
			return "test-result", nil
		},
	}

	// Create function info
	info := common.FunctionInfo{
		Name:        "test-function",
		Description: "Test function",
		Runtime:     "go",
	}

	// Register the function
	registry.Register(info.Name, mockHandler, info)

	// Verify the function was registered
	functions := registry.ListFunctions()
	assert.Equal(t, 1, len(functions))
	assert.Equal(t, "test-function", functions[0].Name)
	assert.Equal(t, "Test function", functions[0].Description)
	assert.Equal(t, "go", functions[0].Runtime)
}

func TestExecute(t *testing.T) {
	registry := NewRegistry()

	// Create a mock handler
	mockHandler := &MockFunctionHandler{
		ExecuteFunc: func(input map[string]interface{}) (interface{}, error) {
			// Return the input as the result for testing
			return input, nil
		},
	}

	// Create function info
	info := common.FunctionInfo{
		Name:        "test-function",
		Description: "Test function",
		Runtime:     "go",
	}

	// Register the function
	registry.Register(info.Name, mockHandler, info)

	// Execute the function
	input := map[string]interface{}{"key": "value"}
	result, err := registry.Execute("test-function", input)

	// Verify the result
	assert.NoError(t, err)
	assert.Equal(t, input, result)

	// Test executing a non-existent function
	_, err = registry.Execute("non-existent", input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestListFunctions(t *testing.T) {
	registry := NewRegistry()

	// Register multiple functions
	for i := 1; i <= 3; i++ {
		name := fmt.Sprintf("function-%d", i)
		mockHandler := &MockFunctionHandler{
			ExecuteFunc: func(input map[string]interface{}) (interface{}, error) {
				return "result", nil
			},
		}

		info := common.FunctionInfo{
			Name:        name,
			Description: fmt.Sprintf("Function %d", i),
			Runtime:     "go",
		}

		registry.Register(name, mockHandler, info)
	}

	// List functions
	functions := registry.ListFunctions()

	// Verify the functions
	assert.Equal(t, 3, len(functions))

	// Check that all functions are in the list
	functionNames := make(map[string]bool)
	for _, f := range functions {
		functionNames[f.Name] = true
	}

	assert.True(t, functionNames["function-1"])
	assert.True(t, functionNames["function-2"])
	assert.True(t, functionNames["function-3"])
}
