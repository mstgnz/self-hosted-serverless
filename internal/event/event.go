package event

import (
	"context"
	"sync"
)

// Event represents a serverless event
type Event struct {
	Type    string         `json:"type"`
	Payload map[string]any `json:"payload"`
}

// Handler is a function that handles an event
type Handler func(ctx context.Context, event Event) error

// Bus is an event bus that dispatches events to registered handlers
type Bus struct {
	handlers map[string][]Handler
	mutex    sync.RWMutex
}

// NewBus creates a new event bus
func NewBus() *Bus {
	return &Bus{
		handlers: make(map[string][]Handler),
	}
}

// Subscribe registers a handler for a specific event type
func (b *Bus) Subscribe(eventType string, handler Handler) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if _, exists := b.handlers[eventType]; !exists {
		b.handlers[eventType] = make([]Handler, 0)
	}

	b.handlers[eventType] = append(b.handlers[eventType], handler)
}

// Publish publishes an event to all registered handlers
func (b *Bus) Publish(ctx context.Context, event Event) []error {
	b.mutex.RLock()
	handlers, exists := b.handlers[event.Type]
	b.mutex.RUnlock()

	if !exists {
		return nil
	}

	var errors []error
	for _, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

// Unsubscribe removes a handler for a specific event type
func (b *Bus) Unsubscribe(eventType string, handler Handler) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if handlers, exists := b.handlers[eventType]; exists {
		for i, h := range handlers {
			if &h == &handler {
				b.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
				break
			}
		}
	}
}

// Global event bus instance
var (
	globalBus  *Bus
	globalOnce sync.Once
)

// GetGlobalBus returns the global event bus instance
func GetGlobalBus() *Bus {
	globalOnce.Do(func() {
		globalBus = NewBus()
	})
	return globalBus
}
