package main

import (
	"context"
	"log"

	"github.com/mstgnz/self-hosted-serverless/internal/common"
	"github.com/mstgnz/self-hosted-serverless/internal/event"
)

// Handler is the function handler
var Handler = &FunctionHandler{}

// Info contains metadata about the function
var Info = common.FunctionInfo{
	Name:        "event-chain-step3",
	Description: "Final step in an event chain",
	Runtime:     "go",
}

// FunctionHandler implements the serverless function
type FunctionHandler struct{}

// Execute executes the function with the given input
func (h *FunctionHandler) Execute(input map[string]interface{}) (interface{}, error) {
	// This function doesn't need to be called directly
	// It works by subscribing to events
	return map[string]interface{}{
		"status":  "ready",
		"message": "Step 3 is ready and listening for events",
	}, nil
}

// init function to subscribe to events
func init() {
	// Subscribe to events from step 2
	eventBus := event.GetGlobalBus()
	eventBus.Subscribe("process.step2.completed", func(ctx context.Context, evt event.Event) error {
		// Process step 2 result
		processId, _ := evt.Payload["process_id"].(string)
		log.Printf("Step 3: Processing event for %s", processId)

		// Create final result data
		data := make(map[string]interface{})

		// Copy payload data
		for k, v := range evt.Payload {
			data[k] = v
		}

		// Add step 3 processing result
		data["step3_result"] = "Step 3 processing completed"
		data["status"] = "completed"

		// Publish final result
		eventBus.Publish(ctx, event.Event{
			Type:    "process.completed",
			Payload: data,
		})

		log.Printf("Step 3: Process completed for %s", processId)
		return nil
	})
}
