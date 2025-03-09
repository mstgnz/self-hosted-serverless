package function

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"plugin"
	"sync"

	"github.com/mstgnz/self-hosted-serverless/internal/common"
	"github.com/mstgnz/self-hosted-serverless/internal/runtime"
)

// FunctionHandler is the interface that all serverless functions must implement
// Deprecated: Use common.FunctionHandler instead
type FunctionHandler = common.FunctionHandler

// FunctionInfo represents metadata about a registered function
// Deprecated: Use common.FunctionInfo instead
type FunctionInfo = common.FunctionInfo

// WasmFunctionHandler implements FunctionHandler for WebAssembly functions
type WasmFunctionHandler struct {
	runtime    *runtime.WasmRuntime
	wasmFile   string
	exportName string
}

// Execute executes a WebAssembly function
func (h *WasmFunctionHandler) Execute(input map[string]any) (any, error) {
	return h.runtime.ExecuteFunction(h.wasmFile, h.exportName, input)
}

// Registry manages the serverless functions
type Registry struct {
	functions   map[string]common.FunctionHandler
	metadata    map[string]common.FunctionInfo
	wasmRuntime *runtime.WasmRuntime
	mutex       sync.RWMutex
}

// NewRegistry creates a new function registry
func NewRegistry() *Registry {
	// Initialize WebAssembly runtime
	wasmRuntime, err := runtime.NewWasmRuntime()
	if err != nil {
		// Log error but continue without WebAssembly support
		fmt.Printf("Warning: Failed to initialize WebAssembly runtime: %v\n", err)
	}

	registry := &Registry{
		functions:   make(map[string]common.FunctionHandler),
		metadata:    make(map[string]common.FunctionInfo),
		wasmRuntime: wasmRuntime,
	}

	// Load all functions from the functions directory
	registry.loadFunctions()

	return registry
}

// Register registers a new function
func (r *Registry) Register(name string, handler common.FunctionHandler, info common.FunctionInfo) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.functions[name] = handler
	r.metadata[name] = info
}

// RegisterWasmFunction registers a WebAssembly function
func (r *Registry) RegisterWasmFunction(name string, wasmFile string, exportName string, info common.FunctionInfo) error {
	if r.wasmRuntime == nil {
		return errors.New("WebAssembly runtime not initialized")
	}

	handler := &WasmFunctionHandler{
		runtime:    r.wasmRuntime,
		wasmFile:   wasmFile,
		exportName: exportName,
	}

	r.Register(name, handler, info)
	return nil
}

// Execute executes a function by name
func (r *Registry) Execute(name string, input map[string]any) (any, error) {
	r.mutex.RLock()
	handler, exists := r.functions[name]
	r.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("function %s not found", name)
	}

	return handler.Execute(input)
}

// ListFunctions returns a list of all registered functions
func (r *Registry) ListFunctions() []common.FunctionInfo {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	functions := make([]common.FunctionInfo, 0, len(r.metadata))
	for _, info := range r.metadata {
		functions = append(functions, info)
	}

	return functions
}

// loadFunctions loads all functions from the functions directory
func (r *Registry) loadFunctions() error {
	functionsDir := "functions"

	// Create the functions directory if it doesn't exist
	if _, err := os.Stat(functionsDir); os.IsNotExist(err) {
		if err := os.MkdirAll(functionsDir, 0755); err != nil {
			return fmt.Errorf("failed to create functions directory: %w", err)
		}
		return nil
	}

	// Walk through the functions directory
	return filepath.WalkDir(functionsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Handle different function types based on file extension
		ext := filepath.Ext(path)
		switch ext {
		case ".so":
			// Load Go plugin
			return r.loadGoPlugin(path)
		case ".wasm":
			// Load WebAssembly module
			if r.wasmRuntime != nil {
				name := filepath.Base(path)
				name = name[:len(name)-len(ext)]
				info := common.FunctionInfo{
					Name:        name,
					Description: fmt.Sprintf("WebAssembly function: %s", name),
					Runtime:     "wasm",
				}
				return r.RegisterWasmFunction(name, path, "execute", info)
			}
		}

		return nil
	})
}

// loadGoPlugin loads a Go plugin
func (r *Registry) loadGoPlugin(path string) error {
	// Load the plugin
	p, err := plugin.Open(path)
	if err != nil {
		return fmt.Errorf("failed to load plugin %s: %w", path, err)
	}

	// Look up the Handler symbol
	handlerSymbol, err := p.Lookup("Handler")
	if err != nil {
		return fmt.Errorf("plugin %s does not export Handler symbol: %w", path, err)
	}

	// Assert that the symbol is a FunctionHandler
	handler, ok := handlerSymbol.(common.FunctionHandler)
	if !ok {
		return errors.New("plugin Handler is not a FunctionHandler")
	}

	// Look up the Info symbol
	infoSymbol, err := p.Lookup("Info")
	if err != nil {
		return fmt.Errorf("plugin %s does not export Info symbol: %w", path, err)
	}

	// Assert that the symbol is a FunctionInfo
	info, ok := infoSymbol.(common.FunctionInfo)
	if !ok {
		return errors.New("plugin Info is not a FunctionInfo")
	}

	// Register the function
	r.Register(info.Name, handler, info)

	return nil
}
