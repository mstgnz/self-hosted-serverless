package function

import (
	"sync"
	"time"
)

// MetricsCollector collects metrics for function executions
type MetricsCollector struct {
	mutex            sync.RWMutex
	executionCounts  map[string]int64
	executionTimes   map[string]time.Duration
	executionErrors  map[string]int64
	lastExecutions   map[string]time.Time
	coldStartCounts  map[string]int64
	coldStartLatency map[string]time.Duration
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		executionCounts:  make(map[string]int64),
		executionTimes:   make(map[string]time.Duration),
		executionErrors:  make(map[string]int64),
		lastExecutions:   make(map[string]time.Time),
		coldStartCounts:  make(map[string]int64),
		coldStartLatency: make(map[string]time.Duration),
	}
}

// RecordExecution records a function execution
func (m *MetricsCollector) RecordExecution(functionName string, duration time.Duration, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Record execution count
	m.executionCounts[functionName]++

	// Record execution time
	m.executionTimes[functionName] += duration

	// Record error if any
	if err != nil {
		m.executionErrors[functionName]++
	}

	// Check if this is a cold start
	now := time.Now()
	lastExecution, exists := m.lastExecutions[functionName]
	if !exists || now.Sub(lastExecution) > 5*time.Minute {
		m.coldStartCounts[functionName]++
		m.coldStartLatency[functionName] += duration
	}

	// Update last execution time
	m.lastExecutions[functionName] = now
}

// GetMetrics returns metrics for all functions
func (m *MetricsCollector) GetMetrics() map[string]FunctionMetrics {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	metrics := make(map[string]FunctionMetrics)

	for name, count := range m.executionCounts {
		avgDuration := time.Duration(0)
		if count > 0 {
			avgDuration = time.Duration(int64(m.executionTimes[name]) / count)
		}

		avgColdStartLatency := time.Duration(0)
		coldStartCount := m.coldStartCounts[name]
		if coldStartCount > 0 {
			avgColdStartLatency = time.Duration(int64(m.coldStartLatency[name]) / coldStartCount)
		}

		metrics[name] = FunctionMetrics{
			Name:                name,
			ExecutionCount:      count,
			AverageDuration:     avgDuration,
			ErrorCount:          m.executionErrors[name],
			LastExecutionTime:   m.lastExecutions[name],
			ColdStartCount:      coldStartCount,
			AvgColdStartLatency: avgColdStartLatency,
		}
	}

	return metrics
}

// GetFunctionMetrics returns metrics for a specific function
func (m *MetricsCollector) GetFunctionMetrics(functionName string) (FunctionMetrics, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	count, exists := m.executionCounts[functionName]
	if !exists {
		return FunctionMetrics{}, false
	}

	avgDuration := time.Duration(0)
	if count > 0 {
		avgDuration = time.Duration(int64(m.executionTimes[functionName]) / count)
	}

	avgColdStartLatency := time.Duration(0)
	coldStartCount := m.coldStartCounts[functionName]
	if coldStartCount > 0 {
		avgColdStartLatency = time.Duration(int64(m.coldStartLatency[functionName]) / coldStartCount)
	}

	metrics := FunctionMetrics{
		Name:                functionName,
		ExecutionCount:      count,
		AverageDuration:     avgDuration,
		ErrorCount:          m.executionErrors[functionName],
		LastExecutionTime:   m.lastExecutions[functionName],
		ColdStartCount:      coldStartCount,
		AvgColdStartLatency: avgColdStartLatency,
	}

	return metrics, true
}

// FunctionMetrics represents metrics for a function
type FunctionMetrics struct {
	Name                string        `json:"name"`
	ExecutionCount      int64         `json:"execution_count"`
	AverageDuration     time.Duration `json:"average_duration"`
	ErrorCount          int64         `json:"error_count"`
	LastExecutionTime   time.Time     `json:"last_execution_time"`
	ColdStartCount      int64         `json:"cold_start_count"`
	AvgColdStartLatency time.Duration `json:"avg_cold_start_latency"`
}
