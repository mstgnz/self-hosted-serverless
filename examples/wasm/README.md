# WebAssembly Function Examples

This directory contains examples of creating and using WebAssembly functions with the Self-Hosted Serverless framework.

## Introduction to WebAssembly Functions

WebAssembly (Wasm) is a binary instruction format that allows you to run code written in multiple languages on the web at near-native speed. In the context of serverless functions, WebAssembly enables you to:

1. Write functions in languages other than Go (e.g., Rust, C/C++, AssemblyScript)
2. Achieve better isolation between functions
3. Potentially improve performance for certain workloads

## Basic WebAssembly Function

The [basic](./basic) directory contains a simple WebAssembly function written in Rust.

### Rust Source Code

```rust
// lib.rs
#[no_mangle]
pub extern "C" fn execute(input_ptr: i32, input_len: i32) -> i32 {
    // Parse input JSON
    let input_slice = unsafe {
        std::slice::from_raw_parts(input_ptr as *const u8, input_len as usize)
    };

    let input: serde_json::Value = match serde_json::from_slice(input_slice) {
        Ok(v) => v,
        Err(_) => return -1,
    };

    // Extract name from input
    let name = input["name"].as_str().unwrap_or("World");

    // Create response
    let response = format!("{{\"message\":\"Hello, {} from WebAssembly!\"}}", name);

    // Allocate memory for response
    let response_bytes = response.into_bytes();
    let response_ptr = Box::into_raw(response_bytes.into_boxed_slice()) as i32;

    // Return pointer to response
    response_ptr
}
```

### Building the WebAssembly Module

```sh
# Install Rust and wasm-pack
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
cargo install wasm-pack

# Build the WebAssembly module
cd examples/wasm/basic
wasm-pack build --target no-modules

# Copy the WebAssembly module to the functions directory
cp pkg/basic_bg.wasm ../../../functions/wasm-basic.wasm
```

### Invoking the Function

```sh
# Invoke via HTTP
curl -X POST http://localhost:8080/run/wasm-basic -d '{"name": "John"}'
# Output: {"message": "Hello, John from WebAssembly!"}
```

## Passing Complex Data

The [complex-data](./complex-data) directory contains a WebAssembly function that handles complex data types.

### AssemblyScript Source Code

```typescript
// index.ts
export function execute(input: string): string {
  const data = JSON.parse(input);

  // Extract user data
  const user = data.user || {};
  const name = user.name || "Unknown";
  const age = user.age || 0;

  // Extract items
  const items = data.items || [];

  // Process data
  const result = {
    greeting: `Hello, ${name}!`,
    age_in_months: age * 12,
    item_count: items.length,
    processed_items: items.map((item: string) => `Processed: ${item}`),
  };

  return JSON.stringify(result);
}
```

### Building the WebAssembly Module

```sh
# Install AssemblyScript
npm install -g assemblyscript

# Build the WebAssembly module
cd examples/wasm/complex-data
npm install
npm run asbuild

# Copy the WebAssembly module to the functions directory
cp build/optimized.wasm ../../../functions/wasm-complex.wasm
```

## WebAssembly with JavaScript

The [javascript](./javascript) directory contains an example of using JavaScript with WebAssembly.

### JavaScript Source Code

```javascript
// main.js
export function execute(input) {
  const data = JSON.parse(input);

  // Process data
  const result = {
    message: `Hello from JavaScript in WebAssembly!`,
    input: data,
    timestamp: new Date().toISOString(),
  };

  return JSON.stringify(result);
}
```

### Building with QuickJS

```sh
# Build using QuickJS compiler
cd examples/wasm/javascript
quickjs-wasm-build main.js -o main.wasm

# Copy the WebAssembly module to the functions directory
cp main.wasm ../../../functions/wasm-js.wasm
```

## Notes on WebAssembly Support

The Self-Hosted Serverless framework supports WebAssembly modules through the `internal/runtime/wasm.go` implementation. When a `.wasm` file is placed in the `functions` directory, it is automatically loaded and registered as a function.

For WebAssembly modules to work with the framework, they should:

1. Export a function named `execute` that takes input data and returns a result
2. Handle JSON serialization/deserialization for input and output
3. Manage memory allocation for the returned data

Different languages and toolchains have different approaches to building WebAssembly modules, so refer to the specific examples for each language.
