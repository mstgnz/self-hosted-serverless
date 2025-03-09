package runtime

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// WasmRuntime represents a WebAssembly runtime for executing WASM functions
type WasmRuntime struct {
	runtime wazero.Runtime
}

// NewWasmRuntime creates a new WebAssembly runtime
func NewWasmRuntime() (*WasmRuntime, error) {
	// Create a new WebAssembly runtime with context
	ctx := context.Background()
	r := wazero.NewRuntime(ctx)

	// Instantiate WASI
	_, err := wasi_snapshot_preview1.Instantiate(ctx, r)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate WASI: %w", err)
	}

	return &WasmRuntime{
		runtime: r,
	}, nil
}

// ExecuteFunction executes a WebAssembly function
func (r *WasmRuntime) ExecuteFunction(wasmFile string, functionName string, args ...any) (any, error) {
	// Create context
	ctx := context.Background()

	// Read the WebAssembly module
	wasmBytes, err := os.ReadFile(wasmFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read WebAssembly file: %w", err)
	}

	// Compile the WebAssembly module
	module, err := r.runtime.CompileModule(ctx, wasmBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to compile WebAssembly module: %w", err)
	}

	// Configure the module
	config := wazero.NewModuleConfig().
		WithStdout(os.Stdout).
		WithStderr(os.Stderr).
		WithStdin(os.Stdin)

	// Instantiate the module
	instance, err := r.runtime.InstantiateModule(ctx, module, config)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate WebAssembly module: %w", err)
	}
	defer instance.Close(ctx)

	// Get the function
	fn := instance.ExportedFunction(functionName)
	if fn == nil {
		return nil, fmt.Errorf("function %s not found", functionName)
	}

	// Convert arguments to WebAssembly values
	wasmArgs := make([]uint64, len(args))
	for i, arg := range args {
		switch v := arg.(type) {
		case int:
			wasmArgs[i] = uint64(v)
		case int32:
			wasmArgs[i] = uint64(v)
		case int64:
			wasmArgs[i] = uint64(v)
		case uint:
			wasmArgs[i] = uint64(v)
		case uint32:
			wasmArgs[i] = uint64(v)
		case uint64:
			wasmArgs[i] = v
		case float32:
			wasmArgs[i] = api.EncodeF32(v)
		case float64:
			wasmArgs[i] = api.EncodeF64(v)
		default:
			return nil, fmt.Errorf("unsupported argument type: %T", arg)
		}
	}

	// Call the function
	results, err := fn.Call(ctx, wasmArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to call function: %w", err)
	}

	// Return the result
	if len(results) == 0 {
		return nil, nil
	}
	return results[0], nil
}

// Close closes the WebAssembly runtime
func (r *WasmRuntime) Close() error {
	if r.runtime != nil {
		ctx := context.Background()
		return r.runtime.Close(ctx)
	}
	return nil
}

// WasmFunctionHandler implements the FunctionHandler interface for WebAssembly functions
type WasmFunctionHandler struct {
	wasmFile     string
	functionName string
	runtime      *WasmRuntime
}

// NewWasmFunctionHandler creates a new WebAssembly function handler
func NewWasmFunctionHandler(wasmFile, functionName string) (*WasmFunctionHandler, error) {
	runtime, err := NewWasmRuntime()
	if err != nil {
		return nil, err
	}

	return &WasmFunctionHandler{
		wasmFile:     wasmFile,
		functionName: functionName,
		runtime:      runtime,
	}, nil
}

// Execute executes the WebAssembly function
func (h *WasmFunctionHandler) Execute(input map[string]any) (any, error) {
	// Convert input to arguments
	// This is a simplified implementation; in a real-world scenario,
	// you would need to serialize the input to a format that the WebAssembly function can understand
	args := make([]any, 0)
	for _, value := range input {
		switch v := value.(type) {
		case int, int32, int64, uint, uint32, uint64, float32, float64:
			args = append(args, v)
		default:
			return nil, errors.New("unsupported input type")
		}
	}

	// Execute the function
	return h.runtime.ExecuteFunction(h.wasmFile, h.functionName, args...)
}

// Close closes the WebAssembly function handler
func (h *WasmFunctionHandler) Close() error {
	if h.runtime != nil {
		return h.runtime.Close()
	}
	return nil
}
