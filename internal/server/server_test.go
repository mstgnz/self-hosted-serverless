package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mstgnz/self-hosted-serverless/internal/common"
	"github.com/mstgnz/self-hosted-serverless/internal/function"
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

func setupTestServer() *Server {
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

	return NewServer(8080, registry)
}

func TestHandleHealth(t *testing.T) {
	server := setupTestServer()

	// Create a request to the health endpoint
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Call the handler directly
	server.handleHealth(w, req)

	// Check the response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ok", response["status"])
}

func TestHandleRunFunction(t *testing.T) {
	server := setupTestServer()

	// Create a request to run a function
	input := map[string]interface{}{
		"key": "value",
	}
	body, _ := json.Marshal(input)
	req := httptest.NewRequest("POST", "/run/test-function", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	// Call the handler directly
	server.handleRunFunction(w, req)

	// Check the response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["result"])

	// Test with a non-existent function
	req = httptest.NewRequest("POST", "/run/non-existent", bytes.NewBuffer(body))
	w = httptest.NewRecorder()
	server.handleRunFunction(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleListFunctions(t *testing.T) {
	server := setupTestServer()

	// Create a request to list functions
	req := httptest.NewRequest("GET", "/functions", nil)
	w := httptest.NewRecorder()

	// Call the handler directly
	server.handleListFunctions(w, req)

	// Check the response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	functions, ok := response["functions"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 1, len(functions))

	// Test with invalid method
	req = httptest.NewRequest("POST", "/functions", nil)
	w = httptest.NewRecorder()
	server.handleListFunctions(w, req)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestHandlePublishEvent(t *testing.T) {
	server := setupTestServer()

	// Create a request to publish an event
	event := map[string]interface{}{
		"type": "test-event",
		"payload": map[string]interface{}{
			"key": "value",
		},
	}
	body, _ := json.Marshal(event)
	req := httptest.NewRequest("POST", "/events", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	// Call the handler directly
	server.handlePublishEvent(w, req)

	// Check the response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "published", response["status"])

	// Test with invalid method
	req = httptest.NewRequest("GET", "/events", nil)
	w = httptest.NewRecorder()
	server.handlePublishEvent(w, req)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}
