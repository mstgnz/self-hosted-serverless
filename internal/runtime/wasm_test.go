package runtime

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewWasmRuntime(t *testing.T) {
	runtime, err := NewWasmRuntime()
	assert.NoError(t, err)
	assert.NotNil(t, runtime)
	assert.NotNil(t, runtime.runtime)
}

// This test requires a real WebAssembly file to test with
// Since we don't have one in the test environment, we'll create a simple test
// that checks for the expected error when trying to load a non-existent file
func TestExecuteFunction_FileNotFound(t *testing.T) {
	runtime, err := NewWasmRuntime()
	assert.NoError(t, err)

	// Try to execute a non-existent file
	_, err = runtime.ExecuteFunction("non-existent.wasm", "main")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read WebAssembly file")
}

// TestExecuteFunction_InvalidWasm tests executing an invalid WebAssembly file
func TestExecuteFunction_InvalidWasm(t *testing.T) {
	runtime, err := NewWasmRuntime()
	assert.NoError(t, err)

	// Create a temporary file with invalid WebAssembly content
	tempDir, err := os.MkdirTemp("", "wasm-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	invalidWasmFile := filepath.Join(tempDir, "invalid.wasm")
	err = os.WriteFile(invalidWasmFile, []byte("invalid wasm content"), 0644)
	assert.NoError(t, err)

	// Try to execute the invalid file
	_, err = runtime.ExecuteFunction(invalidWasmFile, "main")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to compile WebAssembly module")
}
