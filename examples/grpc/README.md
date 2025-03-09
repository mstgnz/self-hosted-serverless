# gRPC Function Examples

This directory contains examples of creating and invoking serverless functions via gRPC.

## Basic gRPC Function

The [basic](./basic) directory contains a simple gRPC function that returns a greeting message.

```go
// Function that returns a greeting message
func (h *FunctionHandler) Execute(input map[string]interface{}) (interface{}, error) {
    name := "World"
    if n, ok := input["name"].(string); ok {
        name = n
    }
    return map[string]interface{}{
        "message": fmt.Sprintf("Hello, %s from gRPC!", name),
    }, nil
}
```

### Invoking the Function

```sh
# Deploy the function
mkdir -p functions/grpc-basic
cp examples/grpc/basic/main.go functions/grpc-basic/main.go

# Invoke via gRPC client
go run examples/grpc/client/main.go -function grpc-basic -input '{"name": "John"}'
# Output: {"message": "Hello, John from gRPC!"}
```

## Complex Data Types

The [complex-data](./complex-data) directory contains a function that handles complex data types.

```go
// Function that handles complex data types
func (h *FunctionHandler) Execute(input map[string]interface{}) (interface{}, error) {
    // Extract nested data
    user, ok := input["user"].(map[string]interface{})
    if !ok {
        return nil, errors.New("user data is required")
    }

    // Extract array data
    items, ok := input["items"].([]interface{})
    if !ok {
        return nil, errors.New("items array is required")
    }

    // Process the data
    return map[string]interface{}{
        "user_processed": map[string]interface{}{
            "name": user["name"],
            "id": user["id"],
        },
        "item_count": len(items),
        "items_processed": items,
    }, nil
}
```

### Invoking the Function

```sh
# Deploy the function
mkdir -p functions/grpc-complex
cp examples/grpc/complex-data/main.go functions/grpc-complex/main.go

# Invoke via gRPC client
go run examples/grpc/client/main.go -function grpc-complex -input '{"user": {"name": "John", "id": 123}, "items": ["item1", "item2", "item3"]}'
```

## Error Handling

The [error-handling](./error-handling) directory contains a function that demonstrates error handling in gRPC.

```go
// Function that demonstrates error handling
func (h *FunctionHandler) Execute(input map[string]interface{}) (interface{}, error) {
    // Check if required fields are present
    if _, ok := input["action"]; !ok {
        return nil, errors.New("action field is required")
    }

    action := input["action"].(string)

    switch action {
    case "success":
        return map[string]interface{}{
            "status": "success",
            "message": "Operation completed successfully",
        }, nil
    case "error":
        return nil, errors.New("operation failed")
    default:
        return nil, fmt.Errorf("unknown action: %s", action)
    }
}
```

### Invoking the Function

```sh
# Deploy the function
mkdir -p functions/grpc-error
cp examples/grpc/error-handling/main.go functions/grpc-error/main.go

# Invoke with success action
go run examples/grpc/client/main.go -function grpc-error -input '{"action": "success"}'

# Invoke with error action
go run examples/grpc/client/main.go -function grpc-error -input '{"action": "error"}'
```

## gRPC Client

The [client](./client) directory contains a simple gRPC client for invoking functions.

```go
package main

import (
    "context"
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "time"

    pb "github.com/mstgnz/self-hosted-serverless/internal/grpc/proto"
    "google.golang.org/grpc"
)

func main() {
    // Parse command line arguments
    functionName := flag.String("function", "", "Function name to execute")
    inputJSON := flag.String("input", "{}", "Input JSON")
    flag.Parse()

    // Connect to the gRPC server
    conn, err := grpc.Dial("localhost:9090", grpc.WithInsecure())
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer conn.Close()

    // Create a client
    client := pb.NewFunctionServiceClient(conn)

    // Parse input JSON
    var input map[string]string
    if err := json.Unmarshal([]byte(*inputJSON), &input); err != nil {
        log.Fatalf("Failed to parse input JSON: %v", err)
    }

    // Execute the function
    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()

    resp, err := client.ExecuteFunction(ctx, &pb.ExecuteFunctionRequest{
        Name:  *functionName,
        Input: input,
    })

    if err != nil {
        log.Fatalf("Error executing function: %v", err)
    }

    // Print the result
    fmt.Printf("Result: %v\n", resp.Result)
}
```

### Using the Client

```sh
# Build the client
go build -o grpc-client examples/grpc/client/main.go

# Invoke a function
./grpc-client -function grpc-basic -input '{"name": "John"}'
```
