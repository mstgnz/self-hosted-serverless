# Event-Driven Architecture Examples

This directory contains examples of using the event-driven architecture in the Self-Hosted Serverless framework.

## Overview

The Self-Hosted Serverless framework includes an event bus that allows functions to communicate with each other through events. This enables building complex workflows and reactive systems.

## Event Publisher

The [publisher](./publisher) directory contains an example of a function that publishes events.

```go
// Function that publishes an event
func (h *FunctionHandler) Execute(input map[string]interface{}) (interface{}, error) {
    // Get event bus
    eventBus := event.GetGlobalBus()

    // Get event type from input
    eventType, ok := input["event_type"].(string)
    if !ok {
        return nil, errors.New("event_type is required and must be a string")
    }

    // Get event payload from input
    payload, ok := input["payload"].(map[string]interface{})
    if !ok {
        return nil, errors.New("payload is required and must be an object")
    }

    // Create event
    evt := event.Event{
        Type:    eventType,
        Payload: payload,
    }

    // Publish event
    ctx := context.Background()
    errors := eventBus.Publish(ctx, evt)

    return map[string]interface{}{
        "published": true,
        "event_type": eventType,
        "errors": errors,
    }, nil
}
```

### Invoking the Function

```sh
# Deploy the function
mkdir -p functions/event-publisher
cp examples/events/publisher/main.go functions/event-publisher/main.go

# Invoke via HTTP
curl -X POST http://localhost:8080/run/event-publisher -d '{
  "event_type": "user.created",
  "payload": {
    "user_id": 123,
    "name": "John Doe",
    "email": "john@example.com"
  }
}'
```

## Event Subscriber

The [subscriber](./subscriber) directory contains an example of a function that subscribes to events.

```go
// Function that subscribes to events
func (h *FunctionHandler) Execute(input map[string]interface{}) (interface{}, error) {
    // Get event bus
    eventBus := event.GetGlobalBus()

    // Get event type from input
    eventType, ok := input["event_type"].(string)
    if !ok {
        return nil, errors.New("event_type is required and must be a string")
    }

    // Subscribe to event
    eventBus.Subscribe(eventType, func(ctx context.Context, evt event.Event) error {
        // Process event
        log.Printf("Received event: %s with payload: %v", evt.Type, evt.Payload)
        return nil
    })

    return map[string]interface{}{
        "subscribed": true,
        "event_type": eventType,
    }, nil
}
```

### Invoking the Function

```sh
# Deploy the function
mkdir -p functions/event-subscriber
cp examples/events/subscriber/main.go functions/event-subscriber/main.go

# Invoke via HTTP
curl -X POST http://localhost:8080/run/event-subscriber -d '{
  "event_type": "user.created"
}'
```

## Event Chain

The [chain](./chain) directory contains an example of chaining multiple functions through events.

### Function 1: Initial Processor

```go
// Function that processes a request and publishes an event
func (h *FunctionHandler) Execute(input map[string]interface{}) (interface{}, error) {
    // Process input
    // ...

    // Publish event for next step
    eventBus := event.GetGlobalBus()
    eventBus.Publish(context.Background(), event.Event{
        Type: "process.step1.completed",
        Payload: map[string]interface{}{
            "result": "Step 1 result",
            "next_step": "step2",
        },
    })

    return map[string]interface{}{
        "status": "processing",
        "step": "step1",
    }, nil
}
```

### Function 2: Second Processor

```go
// Function that subscribes to events from the first processor
func init() {
    // Subscribe to events when the function is loaded
    eventBus := event.GetGlobalBus()
    eventBus.Subscribe("process.step1.completed", func(ctx context.Context, evt event.Event) error {
        // Process step 1 result
        // ...

        // Publish event for next step
        eventBus.Publish(ctx, event.Event{
            Type: "process.step2.completed",
            Payload: map[string]interface{}{
                "result": "Step 2 result",
                "next_step": "step3",
            },
        })

        return nil
    })
}
```

### Function 3: Final Processor

```go
// Function that subscribes to events from the second processor
func init() {
    // Subscribe to events when the function is loaded
    eventBus := event.GetGlobalBus()
    eventBus.Subscribe("process.step2.completed", func(ctx context.Context, evt event.Event) error {
        // Process step 2 result
        // ...

        // Publish final result
        eventBus.Publish(ctx, event.Event{
            Type: "process.completed",
            Payload: map[string]interface{}{
                "result": "Final result",
                "status": "completed",
            },
        })

        return nil
    })
}
```

## Event API

The Self-Hosted Serverless framework also provides an HTTP API for publishing events:

```sh
# Publish an event via the API
curl -X POST http://localhost:8080/events -d '{
  "type": "user.created",
  "payload": {
    "user_id": 123,
    "name": "John Doe",
    "email": "john@example.com"
  }
}'
```

This allows external systems to publish events to the serverless framework.
