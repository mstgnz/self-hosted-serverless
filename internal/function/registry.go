package function

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"plugin"
	"strconv"
	"sync"
	"time"

	"github.com/mstgnz/self-hosted-serverless/internal/common"
	"github.com/mstgnz/self-hosted-serverless/internal/runtime"
)

// WasmFunctionHandler implements FunctionHandler for WebAssembly functions using WASI stdio.
type WasmFunctionHandler struct {
	runtime  *runtime.WasmRuntime
	wasmFile string
}

func (h *WasmFunctionHandler) Execute(input map[string]any) (any, error) {
	return h.runtime.ExecuteWASI(h.wasmFile, input)
}

type execResult struct {
	value any
	err   error
}

// Registry manages the serverless functions
type Registry struct {
	functions       map[string]common.FunctionHandler
	metadata        map[string]common.FunctionInfo
	wasmRuntime     *runtime.WasmRuntime
	metrics         *MetricsCollector
	mutex           sync.RWMutex
	functionTimeout time.Duration
}

// NewRegistry creates a new function registry
func NewRegistry() *Registry {
	timeout := 30 * time.Second
	if secs := os.Getenv("FUNCTION_TIMEOUT_SECS"); secs != "" {
		if v, err := strconv.Atoi(secs); err == nil && v > 0 {
			timeout = time.Duration(v) * time.Second
		}
	}

	wasmRuntime, err := runtime.NewWasmRuntime()
	if err != nil {
		fmt.Printf("Warning: Failed to initialize WebAssembly runtime: %v\n", err)
	}

	registry := &Registry{
		functions:       make(map[string]common.FunctionHandler),
		metadata:        make(map[string]common.FunctionInfo),
		wasmRuntime:     wasmRuntime,
		metrics:         NewMetricsCollector(),
		functionTimeout: timeout,
	}

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
func (r *Registry) RegisterWasmFunction(name string, wasmFile string, info common.FunctionInfo) error {
	if r.wasmRuntime == nil {
		return errors.New("WebAssembly runtime not initialized")
	}

	handler := &WasmFunctionHandler{
		runtime:  r.wasmRuntime,
		wasmFile: wasmFile,
	}

	r.Register(name, handler, info)
	return nil
}

// Execute executes a function by name with a configurable timeout and panic recovery.
func (r *Registry) Execute(name string, input map[string]any) (any, error) {
	r.mutex.RLock()
	handler, exists := r.functions[name]
	r.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("function %s not found", name)
	}

	ctx, cancel := context.WithTimeout(context.Background(), r.functionTimeout)
	defer cancel()

	ch := make(chan execResult, 1)
	go func() {
		var res execResult
		defer func() {
			if rec := recover(); rec != nil {
				res.err = fmt.Errorf("function panicked: %v", rec)
			}
			ch <- res
		}()
		res.value, res.err = handler.Execute(input)
	}()

	startTime := time.Now()
	select {
	case <-ctx.Done():
		r.metrics.RecordExecution(name, time.Since(startTime), ctx.Err())
		return nil, fmt.Errorf("function %s: execution timed out after %v", name, r.functionTimeout)
	case res := <-ch:
		r.metrics.RecordExecution(name, time.Since(startTime), res.err)
		return res.value, res.err
	}
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

// GetMetrics returns metrics for all functions
func (r *Registry) GetMetrics() map[string]FunctionMetrics {
	return r.metrics.GetMetrics()
}

// GetFunctionMetrics returns metrics for a specific function
func (r *Registry) GetFunctionMetrics(name string) (FunctionMetrics, bool) {
	return r.metrics.GetFunctionMetrics(name)
}

// loadFunctions loads all functions from the functions directory
func (r *Registry) loadFunctions() error {
	functionsDir := "functions"

	if _, err := os.Stat(functionsDir); os.IsNotExist(err) {
		if err := os.MkdirAll(functionsDir, 0755); err != nil {
			return fmt.Errorf("failed to create functions directory: %w", err)
		}
		return nil
	}

	return filepath.WalkDir(functionsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		switch ext {
		case ".so":
			return r.loadGoPlugin(path)
		case ".wasm":
			if r.wasmRuntime != nil {
				name := filepath.Base(path)
				name = name[:len(name)-len(ext)]
				info := common.FunctionInfo{
					Name:        name,
					Description: fmt.Sprintf("WebAssembly function: %s", name),
					Runtime:     "wasm",
				}
				return r.RegisterWasmFunction(name, path, info)
			}
		}

		return nil
	})
}

// loadGoPlugin loads a Go plugin
func (r *Registry) loadGoPlugin(path string) error {
	p, err := plugin.Open(path)
	if err != nil {
		return fmt.Errorf("failed to load plugin %s: %w", path, err)
	}

	handlerSymbol, err := p.Lookup("Handler")
	if err != nil {
		return fmt.Errorf("plugin %s does not export Handler symbol: %w", path, err)
	}

	handler, ok := handlerSymbol.(common.FunctionHandler)
	if !ok {
		return errors.New("plugin Handler is not a FunctionHandler")
	}

	infoSymbol, err := p.Lookup("Info")
	if err != nil {
		return fmt.Errorf("plugin %s does not export Info symbol: %w", path, err)
	}

	info, ok := infoSymbol.(common.FunctionInfo)
	if !ok {
		return errors.New("plugin Info is not a FunctionInfo")
	}

	r.Register(info.Name, handler, info)

	return nil
}
