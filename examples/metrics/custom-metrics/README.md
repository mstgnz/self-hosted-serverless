# Custom Metrics Example

This example demonstrates how to record and use custom metrics in your serverless functions.

## Overview

The Self-Hosted Serverless framework automatically collects standard metrics for all function executions, but you can also record custom metrics for specific operations or business logic within your functions.

This example shows how to:

1. Create a metrics collector
2. Record custom metrics for different operations
3. Track success and error metrics separately
4. Use metrics to monitor function performance

## Function Structure

The function in this example:

- Accepts an `operation` parameter that can be one of: `default`, `fast`, `slow`, or `error`
- Simulates different processing times for each operation type
- Records custom metrics for each operation
- Tracks success and error counts separately

## How to Use

### Deploy the Function

```bash
serverless deploy -f custom-metrics
```

### Invoke the Function

You can invoke the function with different operations:

```bash
# Default operation
serverless invoke -f custom-metrics

# Fast operation
serverless invoke -f custom-metrics -d '{"operation": "fast"}'

# Slow operation
serverless invoke -f custom-metrics -d '{"operation": "slow"}'

# Error operation
serverless invoke -f custom-metrics -d '{"operation": "error"}'
```

### View the Metrics

After invoking the function several times with different operations, you can view the metrics:

```bash
serverless metrics -f custom-metrics
```

You'll see both the standard metrics and your custom metrics:

```
Function: custom-metrics
Executions: 10
Average Duration: 85.2ms
Error Count: 2
Cold Starts: 1
Average Cold Start Latency: 120.5ms

Custom Metrics:
- operation.default: 4
- operation.fast: 3
- operation.slow: 1
- operation.error: 2
- errors.total: 2
- success.total: 8
```

## Metrics Dashboard

For a visual representation of your metrics, you can use the metrics dashboard example:

```bash
serverless invoke -f metrics-dashboard
```

This will start a web dashboard that displays all metrics, including your custom metrics.

## Stress Testing

To generate a large number of metrics for testing, you can use the stress test example:

```bash
serverless invoke -f metrics-stress-test -d '{"target_function": "custom-metrics", "concurrency": 5, "duration": 60}'
```

This will invoke your function multiple times with different operations, generating a variety of metrics.

## Code Explanation

The key parts of the code that handle custom metrics:

```go
// getMetricsCollector returns the metrics collector
func (h *FunctionHandler) getMetricsCollector() *function.MetricsCollector {
    if h.metricsCollector == nil {
        h.metricsCollector = function.NewMetricsCollector()
    }
    return h.metricsCollector
}

// recordMetrics records custom metrics for an operation
func (h *FunctionHandler) recordMetrics(operation string, duration time.Duration, err error) {
    // Get metrics collector
    metrics := h.getMetricsCollector()

    // Record operation count
    metrics.RecordExecution(fmt.Sprintf("operation.%s", operation), duration, err)

    // Record custom metrics
    if err != nil {
        metrics.RecordExecution("errors.total", duration, err)
    } else {
        metrics.RecordExecution("success.total", duration, nil)
    }
}
```

This pattern allows you to track any custom metrics you need for your specific application.
