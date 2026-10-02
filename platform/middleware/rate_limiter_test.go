package middleware_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"foodeli/platform/middleware"
	platformRedis "foodeli/platform/redis"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func getTestRedis(t *testing.T) *redis.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	client, err := platformRedis.NewRedisClient(ctx, "localhost", "6379", "", 0)
	if err != nil {
		t.Skipf("Skipping Redis integration test: Redis unavailable (%v)", err)
		return nil
	}
	return client
}

func TestRateLimiter_NilRedis_FailOpen(t *testing.T) {
	r := gin.New()
	r.Use(middleware.RateLimiter(nil, middleware.RateLimiterConfig{
		Prefix: "test:nil",
		Limit:  1,
		Window: 1 * time.Minute,
	}))

	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200 under fail-open, got %d", w.Code)
		}
	}
}

func TestRateLimiter_EnforcesLimitAndHeaders(t *testing.T) {
	rdb := getTestRedis(t)
	defer rdb.Close()

	prefix := fmt.Sprintf("test:limit:%s", uuid.New().String())
	const limit = int64(3)
	const window = 2 * time.Second

	r := gin.New()
	r.Use(middleware.RateLimiter(rdb, middleware.RateLimiterConfig{
		Prefix: prefix,
		Limit:  limit,
		Window: window,
	}))

	r.POST("/login", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	// Clean up key on test finish
	defer func() {
		ctx := context.Background()
		rdb.Del(ctx, fmt.Sprintf("ratelimit:%s:127.0.0.1", prefix))
	}()

	// 1st request: should succeed with remaining = 2
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/login", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("request 1: expected status 200, got %d", w.Code)
	}
	if w.Header().Get("X-RateLimit-Limit") != "3" {
		t.Errorf("request 1: expected X-RateLimit-Limit 3, got %s", w.Header().Get("X-RateLimit-Limit"))
	}
	if w.Header().Get("X-RateLimit-Remaining") != "2" {
		t.Errorf("request 1: expected X-RateLimit-Remaining 2, got %s", w.Header().Get("X-RateLimit-Remaining"))
	}

	// 2nd request: should succeed with remaining = 1
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/login", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("request 2: expected status 200, got %d", w.Code)
	}
	if w.Header().Get("X-RateLimit-Remaining") != "1" {
		t.Errorf("request 2: expected X-RateLimit-Remaining 1, got %s", w.Header().Get("X-RateLimit-Remaining"))
	}

	// 3rd request: should succeed with remaining = 0
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/login", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("request 3: expected status 200, got %d", w.Code)
	}
	if w.Header().Get("X-RateLimit-Remaining") != "0" {
		t.Errorf("request 3: expected X-RateLimit-Remaining 0, got %s", w.Header().Get("X-RateLimit-Remaining"))
	}

	// 4th request: exceeds limit, should return 429 Too Many Requests
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/login", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("request 4: expected status 429, got %d", w.Code)
	}
	if w.Header().Get("Retry-After") == "" {
		t.Errorf("request 4: expected Retry-After header to be present")
	}

	// Wait for window to expire
	time.Sleep(window + 200*time.Millisecond)

	// 5th request: window reset, should succeed again
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/login", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("request 5: expected status 200 after window expiry, got %d", w.Code)
	}
	if w.Header().Get("X-RateLimit-Remaining") != "2" {
		t.Errorf("request 5: expected X-RateLimit-Remaining 2 after reset, got %s", w.Header().Get("X-RateLimit-Remaining"))
	}
}
