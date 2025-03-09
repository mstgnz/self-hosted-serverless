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
	Name:        "event-chain-step2",
	Description: "Second step in an event chain",
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
		"message": "Step 2 is ready and listening for events",
	}, nil
}

// init function to subscribe to events
func init() {
	// Subscribe to events from step 1
	eventBus := event.GetGlobalBus()
	eventBus.Subscribe("process.step1.completed", func(ctx context.Context, evt event.Event) error {
		// Process step 1 result
		processId, _ := evt.Payload["process_id"].(string)
		log.Printf("Step 2: Processing event for %s", processId)

		// Create data for step 3
		data := make(map[string]interface{})

		// Copy payload data
		for k, v := range evt.Payload {
			data[k] = v
		}

		// Add step 2 processing result
		data["step2_result"] = "Step 2 processing completed"
		data["next_step"] = "step3"

		// Publish event for next step
		eventBus.Publish(ctx, event.Event{
			Type:    "process.step2.completed",
			Payload: data,
		})

		log.Printf("Step 2: Published event for step 3")
		return nil
	})
}
