# Metrics Dashboard Example

This example demonstrates how to create a web-based dashboard for visualizing metrics from your serverless functions.

## Overview

The metrics dashboard provides a real-time view of:

- Function execution counts
- Average execution durations
- Error counts
- Cold start metrics
- Custom metrics
- System metrics

The dashboard automatically refreshes every 5 seconds to show the latest metrics data.

## How It Works

The dashboard is implemented as a serverless function that:

1. Starts an HTTP server on a specified port
2. Serves an HTML dashboard using Go templates
3. Provides a JSON API endpoint for metrics data
4. Periodically refreshes metrics data

## How to Use

### Deploy the Function

```bash
serverless deploy -f metrics-dashboard
```

### Start the Dashboard

```bash
# Start with default port (8080)
serverless invoke -f metrics-dashboard

# Start with custom port
serverless invoke -f metrics-dashboard -d '{"port": 3000}'
```

### Access the Dashboard

Open your browser and navigate to:

```
http://localhost:8080
```

(Or the custom port you specified)

### API Endpoint

The dashboard also provides a JSON API endpoint for metrics data:

```
http://localhost:8080/api/metrics
```

This endpoint returns all metrics data in JSON format, which you can use for integration with other monitoring tools.

## Dashboard Features

### Real-time Updates

The dashboard automatically refreshes every 5 seconds to show the latest metrics.

### Metrics Categories

The dashboard organizes metrics into several categories:

- **Function Execution Metrics**: Shows the number of executions for each function
- **Average Duration**: Shows the average execution time for each function
- **Error Metrics**: Shows the number of errors for each function
- **Cold Start Metrics**: Shows cold start information for each function
- **Custom Metrics**: Shows any custom metrics recorded by your functions
- **System Metrics**: Shows system-level metrics like CPU and memory usage

### Visual Design

The dashboard uses a clean, responsive design that works well on desktop and mobile devices. Each metrics category is displayed in a separate card for easy reading.

## Customization

You can customize the dashboard by modifying the HTML template in `templates/dashboard.html`. The template uses Go's template syntax and receives a `MetricsData` struct containing all metrics information.

## Integration with Other Examples

This dashboard works well with the other metrics examples:

- **Custom Metrics**: Deploy and invoke the custom-metrics example to see custom metrics in the dashboard
- **Stress Test**: Use the stress-test example to generate a large volume of metrics data

## Code Structure

- `main.go`: The main function code that starts the HTTP server and handles requests
- `templates/dashboard.html`: The HTML template for the dashboard UI
