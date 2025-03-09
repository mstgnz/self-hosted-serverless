package main

import (
	"context"
	"errors"
	"log"

	"github.com/mstgnz/self-hosted-serverless/internal/common"
	"github.com/mstgnz/self-hosted-serverless/internal/event"
)

// Handler is the function handler
var Handler = &FunctionHandler{}

// Info contains metadata about the function
var Info = common.FunctionInfo{
	Name:        "event-subscriber",
	Description: "A function that subscribes to events",
	Runtime:     "go",
}

// FunctionHandler implements the serverless function
type FunctionHandler struct{}

// Execute executes the function with the given input
func (h *FunctionHandler) Execute(input map[string]interface{}) (interface{}, error) {
	// Get event bus
	eventBus := event.GetGlobalBus()

	// Get event type from input
	eventTypeValue, ok := input["event_type"]
	if !ok {
		return nil, errors.New("event_type is required")
	}

	eventType, ok := eventTypeValue.(string)
	if !ok {
		return nil, errors.New("event_type must be a string")
	}

	// Get handler name from input, default to "default-handler"
	handlerName := "default-handler"
	if nameValue, ok := input["handler_name"].(string); ok {
		handlerName = nameValue
	}

	// Subscribe to event
	eventBus.Subscribe(eventType, func(ctx context.Context, evt event.Event) error {
		// Process event
		log.Printf("[%s] Received event: %s with payload: %v", handlerName, evt.Type, evt.Payload)
		return nil
	})

	return map[string]interface{}{
		"subscribed": true,
		"event_type": eventType,
		"handler":    handlerName,
		"message":    "Event handler registered successfully",
	}, nil
}
