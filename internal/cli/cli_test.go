package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateFunction(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "cli-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	assert.NoError(t, err)
	defer os.Chdir(originalDir)

	err = os.Chdir(tempDir)
	assert.NoError(t, err)

	// Create the functions directory
	err = os.Mkdir("functions", 0755)
	assert.NoError(t, err)

	// Test creating a function
	functionName := "test-function"
	CreateFunction(functionName)

	// Check if the function directory and file were created
	functionDir := filepath.Join("functions", functionName)
	assert.DirExists(t, functionDir)

	functionFile := filepath.Join(functionDir, "main.go")
	assert.FileExists(t, functionFile)

	// Check the content of the file
	content, err := os.ReadFile(functionFile)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "package main")
	assert.Contains(t, string(content), "Name:        \""+functionName+"\"")
}
