package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/mstgnz/self-hosted-serverless/internal/common"
	"github.com/mstgnz/self-hosted-serverless/internal/db"
	"github.com/redis/go-redis/v9"
)

// Handler is the function handler
var Handler = &FunctionHandler{}

// Info contains metadata about the function
var Info = common.FunctionInfo{
	Name:        "redis-example",
	Description: "A function that demonstrates Redis integration",
	Runtime:     "go",
}

// FunctionHandler implements the serverless function
type FunctionHandler struct{}

// Execute executes the function with the given input
func (h *FunctionHandler) Execute(input map[string]interface{}) (interface{}, error) {
	// Get Redis client
	client, err := db.GetRedisClient()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	// Get operation from input
	operation, ok := input["operation"].(string)
	if !ok {
		return nil, errors.New("operation is required and must be a string")
	}

	ctx := context.Background()

	switch operation {
	case "get":
		// Get key from input
		key, ok := input["key"].(string)
		if !ok {
			return nil, errors.New("key is required and must be a string")
		}

		// Get value from Redis
		value, err := client.Get(ctx, key).Result()
		if err != nil {
			if err == redis.Nil {
				return map[string]interface{}{
					"exists": false,
					"key":    key,
				}, nil
			}
			return nil, fmt.Errorf("failed to get value from Redis: %w", err)
		}

		return map[string]interface{}{
			"exists": true,
			"key":    key,
			"value":  value,
		}, nil

	case "set":
		// Get key and value from input
		key, ok := input["key"].(string)
		if !ok {
			return nil, errors.New("key is required and must be a string")
		}

		value, ok := input["value"].(string)
		if !ok {
			return nil, errors.New("value is required and must be a string")
		}

		// Set value in Redis
		err := client.Set(ctx, key, value, 0).Err()
		if err != nil {
			return nil, fmt.Errorf("failed to set value in Redis: %w", err)
		}

		return map[string]interface{}{
			"success": true,
			"key":     key,
			"value":   value,
		}, nil

	case "keys":
		// Get pattern from input, default to "*"
		pattern := "*"
		if p, ok := input["pattern"].(string); ok {
			pattern = p
		}

		// Get keys from Redis
		keys, err := client.Keys(ctx, pattern).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to get keys from Redis: %w", err)
		}

		return map[string]interface{}{
			"keys": keys,
		}, nil

	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}
}
