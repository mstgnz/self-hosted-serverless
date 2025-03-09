package main

import (
	"context"
	"errors"

	"github.com/mstgnz/self-hosted-serverless/internal/common"
	"github.com/mstgnz/self-hosted-serverless/internal/event"
)

// Handler is the function handler
var Handler = &FunctionHandler{}

// Info contains metadata about the function
var Info = common.FunctionInfo{
	Name:        "event-publisher",
	Description: "A function that publishes events",
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

	// Get event payload from input
	payloadValue, ok := input["payload"]
	if !ok {
		return nil, errors.New("payload is required")
	}

	payload, ok := payloadValue.(map[string]interface{})
	if !ok {
		return nil, errors.New("payload must be an object")
	}

	// Create event
	evt := event.Event{
		Type:    eventType,
		Payload: payload,
	}

	// Publish event
	ctx := context.Background()
	errs := eventBus.Publish(ctx, evt)

	// Check for errors
	if len(errs) > 0 {
		errorMessages := make([]string, len(errs))
		for i, err := range errs {
			errorMessages[i] = err.Error()
		}

		return map[string]interface{}{
			"published":  false,
			"event_type": eventType,
			"errors":     errorMessages,
		}, nil
	}

	return map[string]interface{}{
		"published":  true,
		"event_type": eventType,
	}, nil
}
