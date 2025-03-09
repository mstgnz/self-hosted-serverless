package main

import (
	"context"
	"fmt"
	"log"

	"github.com/mstgnz/self-hosted-serverless/internal/common"
	"github.com/mstgnz/self-hosted-serverless/internal/event"
)

// Handler is the function handler
var Handler = &FunctionHandler{}

// Info contains metadata about the function
var Info = common.FunctionInfo{
	Name:        "event-chain-step1",
	Description: "First step in an event chain",
	Runtime:     "go",
}

// FunctionHandler implements the serverless function
type FunctionHandler struct{}

// Execute executes the function with the given input
func (h *FunctionHandler) Execute(input map[string]interface{}) (interface{}, error) {
	// Process input
	processId := fmt.Sprintf("process-%d", input["id"])
	log.Printf("Step 1: Processing request %s", processId)

	// Extract data from input
	data := map[string]interface{}{
		"process_id": processId,
	}

	// Copy input data
	for k, v := range input {
		data[k] = v
	}

	// Add step 1 processing result
	data["step1_result"] = "Step 1 processing completed"
	data["next_step"] = "step2"

	// Publish event for next step
	eventBus := event.GetGlobalBus()
	eventBus.Publish(context.Background(), event.Event{
		Type:    "process.step1.completed",
		Payload: data,
	})

	return map[string]interface{}{
		"status":     "processing",
		"step":       "step1",
		"process_id": processId,
		"message":    "Step 1 processing initiated, event published for step 2",
	}, nil
}

// init function to subscribe to events
func init() {
	// Subscribe to process.completed events to log the final result
	eventBus := event.GetGlobalBus()
	eventBus.Subscribe("process.completed", func(ctx context.Context, evt event.Event) error {
		log.Printf("Process completed: %v", evt.Payload)
		return nil
	})
}
