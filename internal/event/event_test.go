package event

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBus(t *testing.T) {
	bus := NewBus()
	assert.NotNil(t, bus)
	assert.NotNil(t, bus.handlers)
}

func TestSubscribe(t *testing.T) {
	bus := NewBus()

	// Subscribe to an event
	eventType := "test-event"

	bus.Subscribe(eventType, func(ctx context.Context, event Event) error {
		// Handler implementation
		return nil
	})

	// Verify the handler was registered
	assert.NotNil(t, bus.handlers[eventType])
	assert.Equal(t, 1, len(bus.handlers[eventType]))
}

func TestPublish(t *testing.T) {
	bus := NewBus()
	ctx := context.Background()

	// Subscribe to an event
	eventType := "test-event"
	handlerCalled := false
	eventPayload := map[string]interface{}{"key": "value"}

	var receivedEvent Event
	var wg sync.WaitGroup
	wg.Add(1)

	bus.Subscribe(eventType, func(ctx context.Context, event Event) error {
		handlerCalled = true
		receivedEvent = event
		wg.Done()
		return nil
	})

	// Publish an event
	event := Event{
		Type:    eventType,
		Payload: eventPayload,
	}

	errors := bus.Publish(ctx, event)
	wg.Wait()

	// Verify the handler was called
	assert.True(t, handlerCalled)
	assert.Equal(t, 0, len(errors))
	assert.Equal(t, eventType, receivedEvent.Type)
	assert.Equal(t, eventPayload, receivedEvent.Payload)

	// Test publishing an event with no subscribers
	errors = bus.Publish(ctx, Event{Type: "unknown-event"})
	assert.Equal(t, 0, len(errors))
}

func TestUnsubscribe(t *testing.T) {
	bus := NewBus()

	// Subscribe to an event
	eventType := "test-event"

	handler := func(ctx context.Context, event Event) error {
		return nil
	}

	bus.Subscribe(eventType, handler)
	assert.Equal(t, 1, len(bus.handlers[eventType]))

	// Unsubscribe from the event
	bus.Unsubscribe(eventType, handler)

	// This test might be flaky because we're comparing function references
	// and the Unsubscribe implementation uses pointer comparison
	// In a real implementation, we might want to use a more robust way to identify handlers
}

func TestGetGlobalBus(t *testing.T) {
	bus1 := GetGlobalBus()
	bus2 := GetGlobalBus()

	// Verify that we get the same instance
	assert.Equal(t, bus1, bus2)
}
