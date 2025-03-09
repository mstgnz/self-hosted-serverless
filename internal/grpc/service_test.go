package grpc

import (
	"context"
	"testing"

	"github.com/mstgnz/self-hosted-serverless/internal/common"
	"github.com/mstgnz/self-hosted-serverless/internal/function"
	pb "github.com/mstgnz/self-hosted-serverless/internal/grpc/proto"
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

func setupTestService() *Service {
	registry := function.NewRegistry()

	// Register a test function
	mockHandler := &MockFunctionHandler{
		ExecuteFunc: func(input map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{
				"result": "success",
				"input":  input,
			}, nil
		},
	}

	info := common.FunctionInfo{
		Name:        "test-function",
		Description: "Test function",
		Runtime:     "go",
	}

	registry.Register(info.Name, mockHandler, info)

	return NewService(registry)
}

func TestExecuteFunction(t *testing.T) {
	service := setupTestService()
	ctx := context.Background()

	// Create a request to execute a function
	req := &pb.ExecuteFunctionRequest{
		Name: "test-function",
		Input: map[string]string{
			"key": "value",
		},
	}

	// Execute the function
	resp, err := service.ExecuteFunction(ctx, req)

	// Check the response
	assert.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "success", resp.Result["result"])

	// Test with a non-existent function
	req.Name = "non-existent"
	_, err = service.ExecuteFunction(ctx, req)
	assert.Error(t, err)
}

func TestListFunctions(t *testing.T) {
	service := setupTestService()
	ctx := context.Background()

	// Create a request to list functions
	req := &pb.ListFunctionsRequest{}

	// List functions
	resp, err := service.ListFunctions(ctx, req)

	// Check the response
	assert.NoError(t, err)
	assert.Equal(t, 1, len(resp.Functions))
	assert.Equal(t, "test-function", resp.Functions[0].Name)
	assert.Equal(t, "Test function", resp.Functions[0].Description)
	assert.Equal(t, "go", resp.Functions[0].Runtime)
}
