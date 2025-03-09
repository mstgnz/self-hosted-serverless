# HTTP Function Examples

This directory contains examples of creating and invoking serverless functions via HTTP.

## Basic HTTP Function

The [basic](./basic) directory contains a simple HTTP function that returns a greeting message.

```go
// Function that returns a greeting message
func (h *FunctionHandler) Execute(input map[string]interface{}) (interface{}, error) {
    name := "World"
    if n, ok := input["name"].(string); ok {
        name = n
    }
    return map[string]interface{}{
        "message": fmt.Sprintf("Hello, %s!", name),
    }, nil
}
```

### Invoking the Function

```sh
# Deploy the function
mkdir -p functions/basic
cp examples/http/basic/main.go functions/basic/main.go

# Invoke via HTTP
curl -X POST http://localhost:8080/run/basic -d '{"name": "John"}'
# Output: {"message": "Hello, John!"}

# Invoke via CLI
go run cmd/main.go run basic '{"name": "John"}'
```

## JSON Processing Function

The [json-processing](./json-processing) directory contains a function that processes JSON data.

```go
// Function that processes JSON data
func (h *FunctionHandler) Execute(input map[string]interface{}) (interface{}, error) {
    // Extract user data from input
    userData, ok := input["user"].(map[string]interface{})
    if !ok {
        return nil, errors.New("invalid user data")
    }

    // Process the data
    name := userData["name"].(string)
    age, _ := userData["age"].(float64)

    return map[string]interface{}{
        "greeting": fmt.Sprintf("Hello, %s!", name),
        "message": fmt.Sprintf("You are %d years old.", int(age)),
        "adult": age >= 18,
    }, nil
}
```

### Invoking the Function

```sh
# Deploy the function
mkdir -p functions/json-processing
cp examples/http/json-processing/main.go functions/json-processing/main.go

# Invoke via HTTP
curl -X POST http://localhost:8080/run/json-processing -d '{"user": {"name": "John", "age": 25}}'
# Output: {"greeting": "Hello, John!", "message": "You are 25 years old.", "adult": true}
```

## Error Handling Function

The [error-handling](./error-handling) directory contains a function that demonstrates error handling.

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
mkdir -p functions/error-handling
cp examples/http/error-handling/main.go functions/error-handling/main.go

# Invoke with success action
curl -X POST http://localhost:8080/run/error-handling -d '{"action": "success"}'
# Output: {"status": "success", "message": "Operation completed successfully"}

# Invoke with error action
curl -X POST http://localhost:8080/run/error-handling -d '{"action": "error"}'
# Output: {"error": "operation failed"}
```

## Custom Headers Function

The [custom-headers](./custom-headers) directory contains a function that returns custom HTTP headers.

```go
// Function that returns custom HTTP headers
func (h *FunctionHandler) Execute(input map[string]interface{}) (interface{}, error) {
    // Return both data and headers
    return map[string]interface{}{
        "data": map[string]interface{}{
            "message": "Hello, World!",
        },
        "headers": map[string]string{
            "X-Custom-Header": "Custom Value",
            "X-Powered-By": "Self-Hosted Serverless",
        },
    }, nil
}
```

### Invoking the Function

```sh
# Deploy the function
mkdir -p functions/custom-headers
cp examples/http/custom-headers/main.go functions/custom-headers/main.go

# Invoke via HTTP
curl -v -X POST http://localhost:8080/run/custom-headers -d '{}'
# Output will include custom headers in the response
```
