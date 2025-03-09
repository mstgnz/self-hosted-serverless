.PHONY: build run docker-build docker-run clean test proto

# Build the application
build:
	go build -o bin/go-serverless ./cmd/main.go

# Run the application
run:
	go run ./cmd/main.go

# Generate protobuf files
proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/function.proto

# Build the Docker image
docker-build:
	docker build -t go-serverless .

# Run the Docker container
docker-run:
	docker run -p 8080:8080 go-serverless

# Run with Docker Compose
docker-compose-up:
	docker-compose up -d

# Stop Docker Compose
docker-compose-down:
	docker-compose down

# Create a new function
create-function:
	@read -p "Enter function name: " name; \
	go run ./cmd/main.go create function $$name

# Build a function
build-function:
	@read -p "Enter function name: " name; \
	cd functions/$$name && go build -buildmode=plugin -o $$name.so

# Run a function
run-function:
	@read -p "Enter function name: " name; \
	go run ./cmd/main.go run $$name

# List all functions
list-functions:
	go run ./cmd/main.go list

# Clean build artifacts
clean:
	rm -rf bin/
	find functions -name "*.so" -delete

# Run tests
test:
	go test -v ./... 