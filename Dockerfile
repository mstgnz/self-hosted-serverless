FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy dependency files first to leverage layer caching
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build with trimpath and stripped debug info for a smaller, non-leaking binary.
# CGO_ENABLED=1 is required for go-sqlite3 and the plugin system.
RUN CGO_ENABLED=1 GOOS=linux \
    go build -trimpath -ldflags="-s -w" -o go-serverless ./cmd/main.go

FROM alpine:3.21

WORKDIR /app

RUN apk --no-cache add ca-certificates libgcc

# Run as a non-root user
RUN adduser -D -H -h /app appuser
USER appuser

COPY --from=builder /app/go-serverless /app/go-serverless

RUN mkdir -p /app/functions

EXPOSE 8080 9090

ENTRYPOINT ["/app/go-serverless"]
