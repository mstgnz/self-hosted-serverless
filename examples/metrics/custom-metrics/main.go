package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/mstgnz/self-hosted-serverless/internal/common"
	"github.com/mstgnz/self-hosted-serverless/internal/function"
)

// Handler is the function handler
var Handler = &FunctionHandler{}

// Info contains metadata about the function
var Info = common.FunctionInfo{
	Name:        "custom-metrics",
	Description: "A function that demonstrates custom metrics",
	Runtime:     "go",
}

// FunctionHandler implements the serverless function
type FunctionHandler struct {
	metricsCollector *function.MetricsCollector
}

// getMetricsCollector returns the metrics collector
func (h *FunctionHandler) getMetricsCollector() *function.MetricsCollector {
	if h.metricsCollector == nil {
		h.metricsCollector = function.NewMetricsCollector()
	}
	return h.metricsCollector
}

// Execute executes the function with the given input
func (h *FunctionHandler) Execute(input map[string]interface{}) (interface{}, error) {
	// Get operation from input, default to "default"
	operation := "default"
	if op, ok := input["operation"].(string); ok {
		operation = op
	}

	// Simulate different operations with different metrics
	switch operation {
	case "fast":
		return h.fastOperation(input)
	case "slow":
		return h.slowOperation(input)
	case "error":
		return h.errorOperation(input)
	default:
		return h.defaultOperation(input)
	}
}

// defaultOperation performs a default operation
func (h *FunctionHandler) defaultOperation(input map[string]interface{}) (interface{}, error) {
	// Record start time
	startTime := time.Now()

	// Simulate some work
	time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)

	// Record custom metrics
	duration := time.Since(startTime)
	h.recordMetrics("default", duration, nil)

	return map[string]interface{}{
		"operation": "default",
		"duration":  duration.String(),
		"message":   "Default operation completed successfully",
	}, nil
}

// fastOperation performs a fast operation
func (h *FunctionHandler) fastOperation(input map[string]interface{}) (interface{}, error) {
	// Record start time
	startTime := time.Now()

	// Simulate some work
	time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)

	// Record custom metrics
	duration := time.Since(startTime)
	h.recordMetrics("fast", duration, nil)

	return map[string]interface{}{
		"operation": "fast",
		"duration":  duration.String(),
		"message":   "Fast operation completed successfully",
	}, nil
}

// slowOperation performs a slow operation
func (h *FunctionHandler) slowOperation(input map[string]interface{}) (interface{}, error) {
	// Record start time
	startTime := time.Now()

	// Simulate some work
	time.Sleep(time.Duration(rand.Intn(200)+200) * time.Millisecond)

	// Record custom metrics
	duration := time.Since(startTime)
	h.recordMetrics("slow", duration, nil)

	return map[string]interface{}{
		"operation": "slow",
		"duration":  duration.String(),
		"message":   "Slow operation completed successfully",
	}, nil
}

// errorOperation performs an operation that results in an error
func (h *FunctionHandler) errorOperation(input map[string]interface{}) (interface{}, error) {
	// Record start time
	startTime := time.Now()

	// Simulate some work
	time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)

	// Create an error
	err := fmt.Errorf("simulated error in error operation")

	// Record custom metrics
	duration := time.Since(startTime)
	h.recordMetrics("error", duration, err)

	return nil, err
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
