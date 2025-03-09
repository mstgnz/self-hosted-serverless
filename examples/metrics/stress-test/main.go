package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/mstgnz/self-hosted-serverless/internal/common"
)

// Handler is the function handler
var Handler = &FunctionHandler{}

// Info contains metadata about the function
var Info = common.FunctionInfo{
	Name:        "metrics-stress-test",
	Description: "A stress test for function metrics",
	Runtime:     "go",
}

// FunctionHandler implements the serverless function
type FunctionHandler struct{}

// Execute executes the function with the given input
func (h *FunctionHandler) Execute(input map[string]interface{}) (interface{}, error) {
	// Get parameters from input
	concurrency := 5
	if c, ok := input["concurrency"].(float64); ok {
		concurrency = int(c)
	}

	duration := 60
	if d, ok := input["duration"].(float64); ok {
		duration = int(d)
	}

	targetFunction := "custom-metrics"
	if tf, ok := input["target_function"].(string); ok {
		targetFunction = tf
	}

	// Start the stress test
	go runStressTest(targetFunction, concurrency, duration)

	return map[string]interface{}{
		"status":          "running",
		"message":         fmt.Sprintf("Stress test started against function '%s'", targetFunction),
		"concurrency":     concurrency,
		"duration":        duration,
		"target_function": targetFunction,
	}, nil
}

// runStressTest runs a stress test against the specified function
func runStressTest(targetFunction string, concurrency, durationSeconds int) {
	fmt.Printf("Starting stress test against '%s' with %d concurrent workers for %d seconds\n",
		targetFunction, concurrency, durationSeconds)

	// Create a wait group to track workers
	var wg sync.WaitGroup

	// Create a channel to signal workers to stop
	stopCh := make(chan struct{})

	// Start workers
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			// Track metrics for this worker
			successCount := 0
			errorCount := 0

			// Run until stop signal
			for {
				select {
				case <-stopCh:
					fmt.Printf("Worker %d finished: %d successes, %d errors\n",
						workerID, successCount, errorCount)
					return
				default:
					// Generate random operation
					operations := []string{"default", "fast", "slow", "error"}
					operation := operations[rand.Intn(len(operations))]

					// Invoke function
					input := map[string]interface{}{
						"operation": operation,
						"worker_id": workerID,
					}

					// In a real implementation, we would use the function.Invoke method
					// For this example, we'll simulate the invocation
					err := simulateInvoke(targetFunction, input)
					if err != nil {
						errorCount++
						fmt.Printf("Worker %d error: %v\n", workerID, err)
					} else {
						successCount++
					}

					// Add some randomness to invocation rate
					time.Sleep(time.Duration(rand.Intn(200)) * time.Millisecond)
				}
			}
		}(i)
	}

	// Run for specified duration
	time.Sleep(time.Duration(durationSeconds) * time.Second)

	// Signal workers to stop
	close(stopCh)

	// Wait for all workers to finish
	wg.Wait()

	fmt.Println("Stress test completed")
}

// simulateInvoke simulates invoking a function
func simulateInvoke(functionName string, input map[string]interface{}) error {
	// Simulate some processing time
	time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)

	// Simulate errors occasionally
	if rand.Float32() < 0.1 {
		return fmt.Errorf("simulated error invoking function %s", functionName)
	}

	return nil
}
