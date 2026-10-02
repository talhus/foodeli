package middleware

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"foodeli/platform/response"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimiterConfig defines the limits and window duration for rate limiting.
type RateLimiterConfig struct {
	Prefix string        // Unique namespace prefix, e.g. "auth:login"
	Limit  int64         // Maximum allowed requests within the window
	Window time.Duration // Time window duration, e.g. 1*time.Minute
}

// RateLimiter creates a Gin middleware that enforces distributed rate limits per client IP using Redis.
// It leverages an atomic Redis Lua script to eliminate race conditions between INCR and EXPIRE commands.
func RateLimiter(rdb *redis.Client, cfg RateLimiterConfig) gin.HandlerFunc {
	// Lua script:
	// 1. Atomically increment the counter for the given key.
	// 2. If it's a new key (current == 1), set the expiration window.
	// 3. Ensure key has a TTL; if somehow unset (-1), enforce expiration.
	// 4. Return current request count and remaining TTL in seconds.
	const luaScript = `
local current = redis.call('INCR', KEYS[1])
if current == 1 then
    redis.call('EXPIRE', KEYS[1], ARGV[1])
end
local ttl = redis.call('TTL', KEYS[1])
if ttl == -1 then
    redis.call('EXPIRE', KEYS[1], ARGV[1])
    ttl = tonumber(ARGV[1])
end
return {current, ttl}
`

	return func(c *gin.Context) {
		if rdb == nil {
			// Fail-open: If Redis is unavailable or unconfigured, pass request through
			c.Next()
			return
		}

		clientIP := c.ClientIP()
		key := fmt.Sprintf("ratelimit:%s:%s", cfg.Prefix, clientIP)
		windowSeconds := int(cfg.Window.Seconds())
		if windowSeconds <= 0 {
			windowSeconds = 60
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 500*time.Millisecond)
		defer cancel()

		rawRes, err := rdb.Eval(ctx, luaScript, []string{key}, windowSeconds).Slice()
		if err != nil {
			// Fail-open: Do not block legitimate users if Redis has transient network errors
			log.Printf("[RateLimiter] Redis error for key %s: %v", key, err)
			c.Next()
			return
		}

		var currentCount int64
		var ttlSeconds int64

		if len(rawRes) == 2 {
			if v, ok := rawRes[0].(int64); ok {
				currentCount = v
			}
			if v, ok := rawRes[1].(int64); ok {
				ttlSeconds = v
			}
		}

		remaining := cfg.Limit - currentCount
		if remaining < 0 {
			remaining = 0
		}

		// Standard RFC RateLimit headers
		c.Header("X-RateLimit-Limit", strconv.FormatInt(cfg.Limit, 10))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))
		if ttlSeconds > 0 {
			c.Header("X-RateLimit-Reset", strconv.FormatInt(ttlSeconds, 10))
		}

		// Enforce limit
		if currentCount > cfg.Limit {
			if ttlSeconds > 0 {
				c.Header("Retry-After", strconv.FormatInt(ttlSeconds, 10))
			}
			response.TooManyRequests(c, fmt.Sprintf("Too many requests. Please try again after %d seconds.", ttlSeconds))
			c.Abort()
			return
		}

		c.Next()
	}
}
