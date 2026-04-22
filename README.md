# Self-Hosted Serverless

A production-ready, self-hostable alternative to AWS Lambda and Google Cloud Functions. Built with Go, it provides a fast, secure, and independent serverless platform you can run on your own infrastructure.

## Overview

Self-Hosted Serverless gives you complete control over your serverless environment:

- **Cloud Independence**: No vendor lock-in, runs on any Linux host or container
- **Performance**: Go's concurrency model and compiled WASM modules give fast cold starts
- **Security**: API key auth, per-IP rate limiting, request size limits, path traversal prevention
- **Observability**: Built-in execution metrics, error rates, and cold start tracking

Functions run as Go plugins or WASM modules. A central dispatcher handles HTTP and gRPC requests with timeout enforcement and panic recovery on every execution.

## Features

- **HTTP and gRPC**: Dual protocol support
- **WebAssembly Runtime**: WASI-compatible modules via [wazero](https://github.com/tetratelabs/wazero), compiled module caching
- **Go Plugins**: Native Go functions loaded as shared libraries
- **Database Integration**: PostgreSQL, SQLite, and Redis
- **Event Bus**: Publish and subscribe to events across functions
- **API Key Authentication**: `X-API-Key` header or `Authorization: Bearer` token
- **Rate Limiting**: Per-IP sliding window, configurable via env var
- **Function Timeout**: Configurable per-execution deadline with goroutine-level enforcement
- **Panic Recovery**: Bad functions cannot crash the server
- **Metrics**: Execution count, average duration, error rate, cold start count

## Architecture

```
                        ┌─────────────────────────────┐
                        │         HTTP :8080           │
                        │  CORS, Rate Limit, Auth      │
                        └────────────┬────────────────┘
                                     │
                        ┌────────────▼────────────────┐
                        │        gRPC :9090            │
                        └────────────┬────────────────┘
                                     │
                        ┌────────────▼────────────────┐
                        │      Function Registry       │
                        │  timeout + panic recovery    │
                        └──────┬──────────────┬───────┘
                               │              │
                    ┌──────────▼──┐    ┌──────▼──────┐
                    │ Go Plugin   │    │ WASM (WASI)  │
                    │   (.so)     │    │   (.wasm)    │
                    └─────────────┘    └─────────────┘
```

## Requirements

- Go 1.24+
- CGO-capable toolchain (required for go-sqlite3 and the plugin system)
- Docker (optional)
- PostgreSQL, Redis (optional)

## Quick Start

```sh
git clone https://github.com/mstgnz/self-hosted-serverless.git
cd self-hosted-serverless

go mod tidy

# Start without auth (development)
go run cmd/main.go

# Start with auth
API_KEY=secret go run cmd/main.go
```

### Docker

```sh
# Single container
docker build -t self-hosted-serverless .
docker run -p 8080:8080 -p 9090:9090 -e API_KEY=secret self-hosted-serverless

# Full stack (Postgres + Redis)
API_KEY=secret POSTGRES_PASSWORD=strongpassword docker compose up -d
```

## Configuration

All configuration is via environment variables.

| Variable | Default | Description |
|---|---|---|
| `API_KEY` | _(empty)_ | When set, all endpoints (except `/health`) require this key. Leave empty for development only. |
| `FUNCTION_TIMEOUT_SECS` | `30` | Maximum seconds a single function execution may run |
| `RATE_LIMIT_PER_MIN` | `100` | Maximum requests per IP per minute |
| `POSTGRES_HOST` | `localhost` | |
| `POSTGRES_PORT` | `5432` | |
| `POSTGRES_USER` | `postgres` | |
| `POSTGRES_PASSWORD` | `postgres` | |
| `POSTGRES_DB` | `serverless` | |
| `REDIS_HOST` | `localhost` | |
| `REDIS_PORT` | `6379` | |
| `REDIS_PASSWORD` | _(empty)_ | |
| `SQLITE_DB_PATH` | `data/serverless.db` | |

## Authentication

When `API_KEY` is set, every protected endpoint requires the key. Send it via either header:

```sh
curl -H "X-API-Key: secret" http://localhost:8080/functions
# or
curl -H "Authorization: Bearer secret" http://localhost:8080/functions
```

The `/health` endpoint is always public.

## CLI

```sh
# Create a new function scaffold
go run cmd/main.go create function myFunction

# Invoke a function running on the local server
go run cmd/main.go run myFunction

# List registered functions
go run cmd/main.go list

# Show metrics for all functions
go run cmd/main.go metrics

# Show metrics for one function
go run cmd/main.go metrics myFunction
```

## HTTP API

| Method | Path | Description |
|---|---|---|
| `GET` | `/health` | Health check (public) |
| `POST` | `/run/{name}` | Execute a function |
| `GET` | `/functions` | List all registered functions |
| `POST` | `/events` | Publish an event |
| `POST` | `/db` | Execute a SELECT query |
| `GET` | `/metrics` | Metrics for all functions |
| `GET` | `/metrics/{name}` | Metrics for one function |

### Execute a function

```sh
curl -X POST http://localhost:8080/run/myFunction \
  -H "X-API-Key: secret" \
  -H "Content-Type: application/json" \
  -d '{"name": "world"}'
```

### Publish an event

```sh
curl -X POST http://localhost:8080/events \
  -H "X-API-Key: secret" \
  -H "Content-Type: application/json" \
  -d '{"type": "user.created", "payload": {"id": 42}}'
```

### Query the database

Only `SELECT` and `WITH` statements are accepted. Write operations must go through your functions.

```sh
curl -X POST http://localhost:8080/db \
  -H "X-API-Key: secret" \
  -H "Content-Type: application/json" \
  -d '{"query": "SELECT id, name FROM users WHERE active = $1", "args": [true]}'
```

## Writing Functions

### Go Plugin

Create a function with the CLI, then build it as a shared library:

```sh
go run cmd/main.go create function myFunction
cd functions/myFunction
go build -buildmode=plugin -o myFunction.so .
```

The plugin must export two symbols:

```go
package main

import "github.com/mstgnz/self-hosted-serverless/internal/common"

var Handler = &MyHandler{}

var Info = common.FunctionInfo{
    Name:        "myFunction",
    Description: "Does something useful",
    Runtime:     "go",
}

type MyHandler struct{}

func (h *MyHandler) Execute(input map[string]any) (any, error) {
    return map[string]any{"message": "hello"}, nil
}
```

> **Note:** Go plugins require CGO and must be compiled with the same Go version and build flags as the server. Linux is the most reliable target.

### WebAssembly (WASI)

The runtime uses WASI stdio for I/O. The module receives input as a JSON object on stdin and must write its result as JSON to stdout before exiting with code 0.

Example in Go (compile with `GOOS=wasip1 GOARCH=wasm`):

```go
package main

import (
    "encoding/json"
    "fmt"
    "os"
)

func main() {
    var input map[string]any
    json.NewDecoder(os.Stdin).Decode(&input)

    result := map[string]any{
        "message": fmt.Sprintf("Hello, %v!", input["name"]),
    }
    json.NewEncoder(os.Stdout).Encode(result)
}
```

```sh
GOOS=wasip1 GOARCH=wasm go build -o functions/hello.wasm .
```

The server picks up `.wasm` files automatically from the `functions/` directory on startup.

## gRPC

Connect to port 9090. The service definition is in [`proto/function.proto`](./proto/function.proto).

```sh
# List functions
grpcurl -plaintext localhost:9090 function.FunctionService/ListFunctions

# Execute a function
grpcurl -plaintext -d '{"name": "myFunction", "input": {"key": "value"}}' \
  localhost:9090 function.FunctionService/ExecuteFunction
```

## Examples

See the [`examples/`](./examples) directory for runnable examples:

- [HTTP Functions](./examples/http)
- [gRPC Functions](./examples/grpc)
- [WebAssembly Functions](./examples/wasm)
- [Database Integration](./examples/database)
- [Event-Driven Architecture](./examples/events)
- [Metrics and Observability](./examples/metrics)

## Use Cases

- **API Backends**: Lightweight HTTP services without a framework
- **Webhook Processing**: Isolated handlers for each integration
- **Data Processing**: Batch or stream processing functions
- **ML Inference**: Serve models with predictable latency
- **Event-Driven Pipelines**: Chain functions via the event bus

## Contributing

```sh
git clone https://github.com/mstgnz/self-hosted-serverless.git
cd self-hosted-serverless
git checkout -b feature-branch
# make your changes
go test ./...
```

## Roadmap

- [x] WebAssembly support (WASI stdio I/O)
- [x] Function metrics and observability
- [x] API key authentication
- [x] Per-IP rate limiting
- [x] Function execution timeout and panic recovery
- [x] CORS support
- [x] Docker Compose with health checks
- [ ] Kubernetes manifests
- [ ] Horizontal scaling with a shared function registry
- [ ] Structured JSON logging
- [ ] CLI authentication support

## License

MIT
