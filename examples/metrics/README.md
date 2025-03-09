# Metrics and Observability Examples

This directory contains examples of using the metrics and observability features of the Self-Hosted Serverless framework.

## Overview

The Self-Hosted Serverless framework includes built-in metrics collection for function executions. This allows you to monitor the performance and usage of your functions.

## Metrics Collection

The framework automatically collects the following metrics for each function:

- **Execution Count**: The number of times a function has been executed
- **Average Duration**: The average execution time of a function
- **Error Count**: The number of times a function has failed
- **Cold Start Count**: The number of cold starts for a function
- **Average Cold Start Latency**: The average latency of cold starts

## Viewing Metrics

### Via CLI

You can view metrics for all functions or a specific function using the CLI:

```sh
# View metrics for all functions
go run cmd/main.go metrics

# View metrics for a specific function
go run cmd/main.go metrics myFunction
```

### Via HTTP API

You can also view metrics via the HTTP API:

```sh
# Get metrics for all functions
curl http://localhost:8080/metrics

# Get metrics for a specific function
curl http://localhost:8080/metrics/myFunction
```

## Custom Metrics

The [custom-metrics](./custom-metrics) directory contains an example of a function that records custom metrics.

```go
// Function that records custom metrics
func (h *FunctionHandler) Execute(input map[string]interface{}) (interface{}, error) {
    // Get metrics collector
    metrics := h.getMetricsCollector()

    // Record a custom metric
    metrics.RecordCustomMetric("api_calls", 1)

    // Record a custom timing metric
    startTime := time.Now()
    // ... do some work ...
    duration := time.Since(startTime)
    metrics.RecordCustomTiming("api_latency", duration)

    return map[string]interface{}{
        "message": "Custom metrics recorded",
    }, nil
}
```

## Metrics Dashboard

The [dashboard](./dashboard) directory contains an example of a simple metrics dashboard.

### Running the Dashboard

```sh
# Start the dashboard
go run examples/metrics/dashboard/main.go

# Open the dashboard in your browser
open http://localhost:3000
```

## Stress Test

The [stress-test](./stress-test) directory contains a script for stress testing functions and observing metrics.

```sh
# Run the stress test
go run examples/metrics/stress-test/main.go -function myFunction -concurrency 10 -requests 100
```

This will execute the specified function 100 times with 10 concurrent requests and display the metrics afterwards.
