package event

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
)

// Event represents a serverless event
type Event struct {
	Type    string         `json:"type"`
	Payload map[string]any `json:"payload"`
}

// Handler is a function that handles an event
type Handler func(ctx context.Context, event Event) error

type handlerEntry struct {
	id      string
	handler Handler
}

// Bus is an event bus that dispatches events to registered handlers
type Bus struct {
	handlers map[string][]handlerEntry
	mutex    sync.RWMutex
}

var idCounter atomic.Int64

func nextID() string {
	return fmt.Sprintf("%d", idCounter.Add(1))
}

// NewBus creates a new event bus
func NewBus() *Bus {
	return &Bus{
		handlers: make(map[string][]handlerEntry),
	}
}

// Subscribe registers a handler for a specific event type and returns a cancel
// function that removes the handler when called.
func (b *Bus) Subscribe(eventType string, handler Handler) func() {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	id := nextID()
	if _, exists := b.handlers[eventType]; !exists {
		b.handlers[eventType] = make([]handlerEntry, 0)
	}
	b.handlers[eventType] = append(b.handlers[eventType], handlerEntry{id: id, handler: handler})

	return func() {
		b.mutex.Lock()
		defer b.mutex.Unlock()
		entries := b.handlers[eventType]
		for i, e := range entries {
			if e.id == id {
				b.handlers[eventType] = append(entries[:i], entries[i+1:]...)
				return
			}
		}
	}
}

// Publish publishes an event to all registered handlers
func (b *Bus) Publish(ctx context.Context, event Event) []error {
	b.mutex.RLock()
	entries, exists := b.handlers[event.Type]
	b.mutex.RUnlock()

	if !exists {
		return nil
	}

	var errors []error
	for _, entry := range entries {
		if err := entry.handler(ctx, event); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

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
