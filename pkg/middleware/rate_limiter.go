package middleware

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

// RateLimiterConfig holds rate limiter configuration
type RateLimiterConfig struct {
	RedisClient *redis.Client
	MaxRequests int64         // Maximum number of requests
	Window      time.Duration // Time window for rate limiting
}

// RateLimiter implements rate limiting using Redis
type RateLimiter struct {
	config RateLimiterConfig
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config RateLimiterConfig) *RateLimiter {
	return &RateLimiter{
		config: config,
	}
}

// Middleware returns a rate limiting middleware
func (r *RateLimiter) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get client IP
		ip := c.IP()
		if ip == "" {
			ip = "unknown"
		}

		// Create Redis key for this IP
		key := fmt.Sprintf("rate_limit:%s", ip)

		// Get current count
		ctx := c.Context()
		count, err := r.config.RedisClient.Get(ctx, key).Int64()
		if err != nil && err != redis.Nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Internal server error",
			})
		}

		// If count exceeds limit, return error
		if count >= r.config.MaxRequests {
			return c.Status(429).JSON(fiber.Map{
				"error":       "Too many requests",
				"retry_after": r.config.Window.Seconds(),
			})
		}

		// Increment counter
		pipe := r.config.RedisClient.Pipeline()
		incr := pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, r.config.Window)
		_, err = pipe.Exec(ctx)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Internal server error",
			})
		}

		// Add rate limit headers
		c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", r.config.MaxRequests))
		c.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", r.config.MaxRequests-incr.Val()))
		c.Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(r.config.Window).Unix()))

		return c.Next()
	}
}
