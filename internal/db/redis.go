package db

import (
	"context"
	"fmt"
	"sync"

	"github.com/redis/go-redis/v9"
)

var (
	redisOnce   sync.Once
	redisClient *redis.Client
	redisErr    error
)

// GetRedisClient returns a singleton Redis client
func GetRedisClient() (*redis.Client, error) {
	redisOnce.Do(func() {
		// Get connection parameters from environment variables
		host := getEnv("REDIS_HOST", "localhost")
		port := getEnv("REDIS_PORT", "6379")
		password := getEnv("REDIS_PASSWORD", "")
		db := 0 // Default database

		// Create the Redis client
		redisClient = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", host, port),
			Password: password,
			DB:       db,
		})

		// Test the connection
		ctx := context.Background()
		_, redisErr = redisClient.Ping(ctx).Result()
	})

	return redisClient, redisErr
}

// CloseRedisClient closes the Redis client
func CloseRedisClient() error {
	if redisClient != nil {
		return redisClient.Close()
	}
	return nil
}
