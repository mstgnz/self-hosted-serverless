package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/mstgnz/self-hosted-serverless/internal/common"
)

//go:embed templates
var templateFS embed.FS

// Handler is the function handler
var Handler = &FunctionHandler{}

// Info contains metadata about the function
var Info = common.FunctionInfo{
	Name:        "metrics-dashboard",
	Description: "A dashboard for visualizing function metrics",
	Runtime:     "go",
}

// FunctionHandler implements the serverless function
type FunctionHandler struct{}

// Execute executes the function with the given input
func (h *FunctionHandler) Execute(input map[string]interface{}) (interface{}, error) {
	// Get port from input, default to 8080
	port := 8080
	if p, ok := input["port"].(float64); ok {
		port = int(p)
	}

	// Start the dashboard server
	go startDashboard(port)

	return map[string]interface{}{
		"status":  "running",
		"message": fmt.Sprintf("Metrics dashboard is running on http://localhost:%d", port),
		"port":    port,
	}, nil
}

// MetricsData represents the metrics data structure
type MetricsData struct {
	FunctionMetrics map[string]FunctionMetric `json:"functionMetrics"`
	CustomMetrics   map[string]interface{}    `json:"customMetrics"`
	SystemMetrics   map[string]interface{}    `json:"systemMetrics"`
	LastUpdated     string                    `json:"lastUpdated"`
}

// FunctionMetric represents metrics for a single function
type FunctionMetric struct {
	ExecutionCount          int     `json:"executionCount"`
	AverageDuration         float64 `json:"averageDuration"`
	ErrorCount              int     `json:"errorCount"`
	ColdStartCount          int     `json:"coldStartCount"`
	AverageColdStartLatency float64 `json:"averageColdStartLatency"`
}

// metricsData is the global metrics data
var metricsData = MetricsData{
	FunctionMetrics: make(map[string]FunctionMetric),
	CustomMetrics:   make(map[string]interface{}),
	SystemMetrics:   make(map[string]interface{}),
	LastUpdated:     time.Now().Format(time.RFC3339),
}

// GetAllMetrics returns all metrics data
func GetAllMetrics() MetricsData {
	return metricsData
}

// GetMetricsJSON returns metrics data as JSON
func GetMetricsJSON() []byte {
	data, err := json.Marshal(metricsData)
	if err != nil {
		log.Printf("Error marshaling metrics data: %v", err)
		return []byte("{}")
	}
	return data
}

// RefreshMetrics refreshes the metrics data
func RefreshMetrics() {
	// In a real implementation, this would fetch metrics from the function runtime
	// For this example, we'll generate some sample data

	// Update function metrics
	functions := []string{"custom-metrics", "metrics-dashboard", "metrics-stress-test"}
	for _, fn := range functions {
		// Generate or update metrics for this function
		metric, exists := metricsData.FunctionMetrics[fn]
		if !exists {
			metric = FunctionMetric{}
		}

		// Update with some random data
		metric.ExecutionCount += rand.Intn(10)
		metric.AverageDuration = float64(50 + rand.Intn(200))
		metric.ErrorCount += rand.Intn(3)
		metric.ColdStartCount += rand.Intn(2)
		metric.AverageColdStartLatency = float64(100 + rand.Intn(300))

		metricsData.FunctionMetrics[fn] = metric
	}

	// Update custom metrics
	customMetricNames := []string{
		"operation.default", "operation.fast", "operation.slow", "operation.error",
		"errors.total", "success.total",
	}
	for _, name := range customMetricNames {
		// Generate or update custom metric
		value, exists := metricsData.CustomMetrics[name]
		if !exists {
			value = 0
		}

		// Update with some random data
		intValue, _ := value.(int)
		metricsData.CustomMetrics[name] = intValue + rand.Intn(5)
	}

	// Update system metrics
	metricsData.SystemMetrics["cpu_usage"] = fmt.Sprintf("%.1f%%", 10+rand.Float64()*30)
	metricsData.SystemMetrics["memory_usage"] = fmt.Sprintf("%.1f MB", 100+rand.Float64()*200)
	metricsData.SystemMetrics["disk_usage"] = fmt.Sprintf("%.1f GB", 1+rand.Float64()*5)

	// Update last updated timestamp
	metricsData.LastUpdated = time.Now().Format(time.RFC3339)
}

// startDashboard starts the dashboard server
func startDashboard(port int) {
	// Parse templates
	tmpl, err := template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}

	// Create HTTP server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Get metrics data
		metricsData := GetAllMetrics()

		// Render dashboard template
		err := tmpl.ExecuteTemplate(w, "dashboard.html", metricsData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	// Start API endpoints for metrics data
	http.HandleFunc("/api/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(GetMetricsJSON())
	})

	// Start HTTP server
	log.Printf("Starting metrics dashboard on port %d", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatalf("Failed to start dashboard server: %v", err)
	}
}

// init initializes the function
func init() {
	// Create templates directory if it doesn't exist
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(templateFS))))

	// Register metrics collector refresh
	go func() {
		for {
			RefreshMetrics()
			time.Sleep(5 * time.Second)
		}
	}()
}
