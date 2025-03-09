# Self-Hosted Serverless

A lightweight, self-hostable alternative to traditional serverless solutions like AWS Lambda and Google Cloud Functions. Built with Go, this framework provides a fast, scalable, and independent serverless platform that you can run on your own infrastructure.

## 🚀 Overview

Self-Hosted Serverless gives you complete control over your serverless environment with these key advantages:

- **Cloud Independence**: Run on your own infrastructure without vendor lock-in
- **Superior Performance**: Leverage Go's speed for significantly faster cold starts
- **Complete Control**: Manage your own resource limits, timeouts, and scaling
- **Cost Efficiency**: Reduce expenses compared to public cloud solutions

Functions can run as independent containers or processes, with a central dispatcher handling HTTP/gRPC requests. WebAssembly support enables running code written in multiple languages.

## ✨ Core Features

- **HTTP and gRPC Support**: Dual protocol support for maximum flexibility
- **WebAssembly Runtime**: Run functions compiled from various languages
- **Database Integration**: Built-in support for PostgreSQL, SQLite, and Redis
- **Event-Driven Architecture**: Publish and subscribe to events across functions
- **Lightweight Runtime**: Minimal resource footprint with maximum performance
- **Simple CLI**: Easy function management and deployment
- **Docker Ready**: Containerized deployment with included configurations

## 🛠️ Technical Architecture

- **Independent Operation**: No external cloud dependencies
- **High Performance**: Go's concurrency model enables efficient processing
- **Multi-Language Support**: Write functions in Go or any WebAssembly-compatible language
- **Scalable Design**: Process requests efficiently using goroutines and channels
- **Simple Integration**: Easy to connect with API gateways and database systems

## 🌐 Installation

### Requirements

- Go 1.24+
- Docker (optional for containerized deployment)
- PostgreSQL (optional for database features)

### Quick Start

```sh
# Clone the repository
git clone https://github.com/mstgnz/self-hosted-serverless.git
cd self-hosted-serverless

# Install dependencies
go mod tidy

# Start the server
go run cmd/main.go
```

### Docker Deployment

```sh
docker build -t self-hosted-serverless .
docker run -p 8080:8080 -p 9090:9090 self-hosted-serverless
```

## ⚙️ Usage Guide

### Creating a Function

```sh
go run cmd/main.go create function myFunction
```

This creates a `functions/myFunction/main.go` file with a template function.

### Executing Functions

**Via HTTP API:**

```sh
curl -X POST http://localhost:8080/run/myFunction -d '{"key": "value"}'
```

**Via gRPC:**
Use any gRPC client to connect to port 9090 and call the `ExecuteFunction` method.

### WebAssembly Functions

Place WebAssembly files in the functions directory for automatic loading:

```sh
cp my-function.wasm functions/
```

### Database Operations

Execute database queries through the API:

```sh
curl -X POST http://localhost:8080/db -d '{"query": "SELECT * FROM users", "args": []}'
```

## 🎯 Use Cases

- **API Backends**: Create lightweight API services with minimal overhead
- **Webhook Processing**: Handle incoming webhooks with custom logic
- **Data Processing**: Process data streams or batch operations efficiently
- **Machine Learning Inference**: Serve ML models with low latency
- **Real-Time Notifications**: Build event-driven notification systems
- **Database Operations**: Perform database operations without managing connections

## 🔧 Development

To contribute to the project:

```sh
git clone https://github.com/mstgnz/self-hosted-serverless.git
cd self-hosted-serverless
git checkout -b feature-branch
# Make your changes
go test ./...  # Run tests before submitting
```

## 🗺️ Roadmap

- [x] WebAssembly Support
- [ ] Advanced CLI with monitoring capabilities
- [ ] Kubernetes integration for orchestrated deployment
- [ ] Function metrics and observability
- [ ] Enhanced security features

## 📄 License

This project is provided under the MIT license.
