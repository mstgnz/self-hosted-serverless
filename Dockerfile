FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o go-serverless ./cmd/main.go

# Create a minimal image
FROM alpine:latest

WORKDIR /app

# Install required packages
RUN apk --no-cache add ca-certificates libc6-compat

# Copy the binary from the builder stage
COPY --from=builder /app/go-serverless /app/go-serverless

# Create functions directory
RUN mkdir -p /app/functions

# Expose the port
EXPOSE 8080

# Set the entrypoint
ENTRYPOINT ["/app/go-serverless"] 