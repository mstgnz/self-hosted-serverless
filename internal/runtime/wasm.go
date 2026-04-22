package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/tetratelabs/wazero/sys"
)

type cachedModule struct {
	module  wazero.CompiledModule
	modTime time.Time
}

// WasmRuntime represents a WebAssembly runtime for executing WASM functions
type WasmRuntime struct {
	runtime wazero.Runtime
	cache   map[string]cachedModule
	mu      sync.RWMutex
}

var instanceCounter atomic.Int64

// NewWasmRuntime creates a new WebAssembly runtime
func NewWasmRuntime() (*WasmRuntime, error) {
	ctx := context.Background()
	r := wazero.NewRuntime(ctx)

	_, err := wasi_snapshot_preview1.Instantiate(ctx, r)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate WASI: %w", err)
	}

	return &WasmRuntime{
		runtime: r,
		cache:   make(map[string]cachedModule),
	}, nil
}

// getCompiledModule returns a cached compiled module, or compiles and caches it.
func (r *WasmRuntime) getCompiledModule(ctx context.Context, wasmFile string) (wazero.CompiledModule, error) {
	info, err := os.Stat(wasmFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read WebAssembly file: %w", err)
	}
	modTime := info.ModTime()

	r.mu.RLock()
	cached, ok := r.cache[wasmFile]
	r.mu.RUnlock()

	if ok && cached.modTime.Equal(modTime) {
		return cached.module, nil
	}

	wasmBytes, err := os.ReadFile(wasmFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read WebAssembly file: %w", err)
	}

	module, err := r.runtime.CompileModule(ctx, wasmBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to compile WebAssembly module: %w", err)
	}

	r.mu.Lock()
	r.cache[wasmFile] = cachedModule{module: module, modTime: modTime}
	r.mu.Unlock()

	return module, nil
}

// ExecuteWASI runs a WASI command module using JSON-over-stdio for I/O.
// The module reads its input as a JSON object from stdin and must write
// its result as a JSON value to stdout before exiting with code 0.
func (r *WasmRuntime) ExecuteWASI(wasmFile string, input map[string]any) (any, error) {
	ctx := context.Background()

	module, err := r.getCompiledModule(ctx, wasmFile)
	if err != nil {
		return nil, err
	}

	inputJSON, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input: %w", err)
	}

	var stdout bytes.Buffer
	// Each instantiation needs a unique name so the runtime can host concurrent calls.
	instanceName := fmt.Sprintf("%s#%d", filepath.Base(wasmFile), instanceCounter.Add(1))
	config := wazero.NewModuleConfig().
		WithStdout(&stdout).
		WithStderr(os.Stderr).
		WithStdin(bytes.NewReader(inputJSON)).
		WithName(instanceName)

	// InstantiateModule runs _start automatically for WASI command modules.
	// When the module calls proc_exit(0), wazero returns a *sys.ExitError with code 0.
	_, err = r.runtime.InstantiateModule(ctx, module, config)
	if err != nil {
		var exitErr *sys.ExitError
		if !errors.As(err, &exitErr) || exitErr.ExitCode() != 0 {
			return nil, fmt.Errorf("failed to execute WebAssembly module: %w", err)
		}
		// exit code 0 = normal completion
	}

	if stdout.Len() == 0 {
		return nil, nil
	}

	var result any
	if jsonErr := json.Unmarshal(stdout.Bytes(), &result); jsonErr != nil {
		// Return raw string output if it isn't valid JSON.
		return stdout.String(), nil
	}
	return result, nil
}

// ExecuteFunction calls a named export directly with primitive numeric arguments.
// Suitable for low-level WASM modules; prefer ExecuteWASI for application-level functions.
func (r *WasmRuntime) ExecuteFunction(wasmFile string, functionName string, args ...any) (any, error) {
	ctx := context.Background()

	module, err := r.getCompiledModule(ctx, wasmFile)
	if err != nil {
		return nil, err
	}

	instanceName := fmt.Sprintf("%s#%d", filepath.Base(wasmFile), instanceCounter.Add(1))
	config := wazero.NewModuleConfig().
		WithStdout(os.Stdout).
		WithStderr(os.Stderr).
		WithStdin(os.Stdin).
		WithName(instanceName)

	instance, err := r.runtime.InstantiateModule(ctx, module, config)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate WebAssembly module: %w", err)
	}
	defer instance.Close(ctx)

	fn := instance.ExportedFunction(functionName)
	if fn == nil {
		return nil, fmt.Errorf("function %s not found", functionName)
	}

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

	results, err := fn.Call(ctx, wasmArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to call function: %w", err)
	}

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
