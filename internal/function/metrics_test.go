package function

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewMetricsCollector(t *testing.T) {
	collector := NewMetricsCollector()
	assert.NotNil(t, collector)
	assert.NotNil(t, collector.executionCounts)
	assert.NotNil(t, collector.executionTimes)
	assert.NotNil(t, collector.executionErrors)
	assert.NotNil(t, collector.lastExecutions)
	assert.NotNil(t, collector.coldStartCounts)
	assert.NotNil(t, collector.coldStartLatency)
}

func TestRecordExecution(t *testing.T) {
	collector := NewMetricsCollector()

	// Record a successful execution
	functionName := "test-function"
	duration := 100 * time.Millisecond
	collector.RecordExecution(functionName, duration, nil)

	// Verify the metrics
	metrics, exists := collector.GetFunctionMetrics(functionName)
	assert.True(t, exists)
	assert.Equal(t, functionName, metrics.Name)
	assert.Equal(t, int64(1), metrics.ExecutionCount)
	assert.Equal(t, duration, metrics.AverageDuration)
	assert.Equal(t, int64(0), metrics.ErrorCount)
	assert.Equal(t, int64(1), metrics.ColdStartCount)
	assert.Equal(t, duration, metrics.AvgColdStartLatency)

	// Record another execution
	collector.RecordExecution(functionName, duration, nil)

	// Verify the metrics
	metrics, exists = collector.GetFunctionMetrics(functionName)
	assert.True(t, exists)
	assert.Equal(t, int64(2), metrics.ExecutionCount)
	assert.Equal(t, duration, metrics.AverageDuration)
	assert.Equal(t, int64(0), metrics.ErrorCount)
	assert.Equal(t, int64(1), metrics.ColdStartCount) // Still 1 because it's not a cold start

	// Record an execution with an error
	collector.RecordExecution(functionName, duration, errors.New("test error"))

	// Verify the metrics
	metrics, exists = collector.GetFunctionMetrics(functionName)
	assert.True(t, exists)
	assert.Equal(t, int64(3), metrics.ExecutionCount)
	assert.Equal(t, duration, metrics.AverageDuration)
	assert.Equal(t, int64(1), metrics.ErrorCount)
}

func TestGetMetrics(t *testing.T) {
	collector := NewMetricsCollector()

	// Record executions for multiple functions
	collector.RecordExecution("function1", 100*time.Millisecond, nil)
	collector.RecordExecution("function2", 200*time.Millisecond, nil)
	collector.RecordExecution("function3", 300*time.Millisecond, errors.New("test error"))

	// Get all metrics
	metrics := collector.GetMetrics()

	// Verify the metrics
	assert.Equal(t, 3, len(metrics))
	assert.Contains(t, metrics, "function1")
	assert.Contains(t, metrics, "function2")
	assert.Contains(t, metrics, "function3")

	assert.Equal(t, int64(1), metrics["function1"].ExecutionCount)
	assert.Equal(t, int64(1), metrics["function2"].ExecutionCount)
	assert.Equal(t, int64(1), metrics["function3"].ExecutionCount)

	assert.Equal(t, 100*time.Millisecond, metrics["function1"].AverageDuration)
	assert.Equal(t, 200*time.Millisecond, metrics["function2"].AverageDuration)
	assert.Equal(t, 300*time.Millisecond, metrics["function3"].AverageDuration)

	assert.Equal(t, int64(0), metrics["function1"].ErrorCount)
	assert.Equal(t, int64(0), metrics["function2"].ErrorCount)
	assert.Equal(t, int64(1), metrics["function3"].ErrorCount)
}

func TestGetFunctionMetrics(t *testing.T) {
	collector := NewMetricsCollector()

	// Record an execution
	functionName := "test-function"
	duration := 100 * time.Millisecond
	collector.RecordExecution(functionName, duration, nil)

	// Get metrics for the function
	metrics, exists := collector.GetFunctionMetrics(functionName)
	assert.True(t, exists)
	assert.Equal(t, functionName, metrics.Name)
	assert.Equal(t, int64(1), metrics.ExecutionCount)
	assert.Equal(t, duration, metrics.AverageDuration)

	// Get metrics for a non-existent function
	_, exists = collector.GetFunctionMetrics("non-existent")
	assert.False(t, exists)
}

func TestColdStart(t *testing.T) {
	collector := NewMetricsCollector()

	// Record an execution
	functionName := "test-function"
	duration := 100 * time.Millisecond
	collector.RecordExecution(functionName, duration, nil)

	// Verify it's a cold start
	metrics, _ := collector.GetFunctionMetrics(functionName)
	assert.Equal(t, int64(1), metrics.ColdStartCount)
	assert.Equal(t, duration, metrics.AvgColdStartLatency)

	// Record another execution immediately (not a cold start)
	collector.RecordExecution(functionName, duration, nil)

	// Verify cold start count hasn't changed
	metrics, _ = collector.GetFunctionMetrics(functionName)
	assert.Equal(t, int64(1), metrics.ColdStartCount)

	// Simulate time passing (> 5 minutes)
	collector.lastExecutions[functionName] = time.Now().Add(-6 * time.Minute)

	// Record another execution (should be a cold start)
	collector.RecordExecution(functionName, duration, nil)

	// Verify cold start count has increased
	metrics, _ = collector.GetFunctionMetrics(functionName)
	assert.Equal(t, int64(2), metrics.ColdStartCount)
}
