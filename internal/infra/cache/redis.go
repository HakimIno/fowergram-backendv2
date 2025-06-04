package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache implements cache using Redis
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new Redis cache client
func NewRedisCache(redisURL string) (*RedisCache, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	client := redis.NewClient(opt)

	// Test connection
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	return &RedisCache{
		client: client,
	}, nil
}

// GetClient returns the underlying Redis client
func (r *RedisCache) GetClient() *redis.Client {
	return r.client
}

// Close closes the Redis connection
func (r *RedisCache) Close() error {
	return r.client.Close()
}

// Get retrieves a value by key
func (r *RedisCache) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// Set stores a value by key
func (r *RedisCache) Set(ctx context.Context, key, value string) error {
	return r.client.Set(ctx, key, value, 0).Err()
}

// SetWithExpiration stores a value by key with expiration
func (r *RedisCache) SetWithExpiration(ctx context.Context, key, value string, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}
