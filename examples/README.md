# Self-Hosted Serverless Examples

This directory contains examples demonstrating the various features of the Self-Hosted Serverless framework. Each subdirectory focuses on a specific feature with code samples and explanations.

## Table of Contents

- [HTTP Functions](#http-functions)
- [gRPC Functions](#grpc-functions)
- [WebAssembly Functions](#webassembly-functions)
- [Database Integration](#database-integration)
- [Event-Driven Architecture](#event-driven-architecture)
- [Metrics and Observability](#metrics-and-observability)

## HTTP Functions

The [http](./http) directory contains examples of creating and invoking serverless functions via HTTP.

- Basic HTTP function
- HTTP function with JSON input/output
- HTTP function with error handling
- HTTP function with custom headers

## gRPC Functions

The [grpc](./grpc) directory contains examples of creating and invoking serverless functions via gRPC.

- Basic gRPC function
- gRPC streaming function
- gRPC function with complex data types
- gRPC client examples

## WebAssembly Functions

The [wasm](./wasm) directory contains examples of using WebAssembly functions.

- Creating WebAssembly functions in different languages
- Deploying WebAssembly functions
- Invoking WebAssembly functions
- Passing data between host and WebAssembly

## Database Integration

The [database](./database) directory contains examples of integrating with databases.

- PostgreSQL integration
- SQLite integration
- Redis integration
- Database query examples

## Event-Driven Architecture

The [events](./events) directory contains examples of using the event-driven architecture.

- Publishing events
- Subscribing to events
- Event-driven function chains
- Event filtering

## Metrics and Observability

The [metrics](./metrics) directory contains examples of using the metrics and observability features.

- Collecting function metrics
- Viewing metrics via CLI
- Viewing metrics via HTTP API
- Custom metrics dashboards

## Running the Examples

Each example directory contains its own README with specific instructions, but generally you can run the examples as follows:

1. Start the serverless framework:

   ```sh
   go run cmd/main.go
   ```

2. Deploy the example function:

   ```sh
   # Copy the function to the functions directory
   cp examples/http/basic/main.go functions/basic/main.go
   ```

3. Invoke the function:

   ```sh
   # Via HTTP
   curl -X POST http://localhost:8080/run/basic -d '{"key": "value"}'

   # Via CLI
   go run cmd/main.go run basic
   ```

4. View metrics:
   ```sh
   go run cmd/main.go metrics basic
   ```
