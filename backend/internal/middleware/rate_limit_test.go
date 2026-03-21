package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func TestInMemoryRateLimiter_FirstRequest(t *testing.T) {
	logger := zap.NewNop()
	limiter := NewInMemoryRateLimiter(logger)

	allowed := limiter.Allow("test-key", 10, time.Minute)
	if !allowed {
		t.Error("first request should be allowed")
	}
}

func TestInMemoryRateLimiter_WithinLimit(t *testing.T) {
	logger := zap.NewNop()
	limiter := NewInMemoryRateLimiter(logger)

	limit := 5
	for i := 0; i < limit; i++ {
		allowed := limiter.Allow("test-key", limit, time.Minute)
		if !allowed {
			t.Errorf("request %d should be allowed (limit=%d)", i+1, limit)
		}
	}
}

func TestInMemoryRateLimiter_ExceedsLimit(t *testing.T) {
	logger := zap.NewNop()
	limiter := NewInMemoryRateLimiter(logger)

	limit := 3
	for i := 0; i < limit; i++ {
		limiter.Allow("test-key", limit, time.Minute)
	}

	allowed := limiter.Allow("test-key", limit, time.Minute)
	if allowed {
		t.Error("request exceeding limit should be denied")
	}
}

func TestInMemoryRateLimiter_WindowReset(t *testing.T) {
	logger := zap.NewNop()
	limiter := NewInMemoryRateLimiter(logger)

	limit := 2
	window := 100 * time.Millisecond

	for i := 0; i < limit; i++ {
		limiter.Allow("test-key", limit, window)
	}

	allowed := limiter.Allow("test-key", limit, window)
	if allowed {
		t.Error("should be denied before window reset")
	}

	time.Sleep(150 * time.Millisecond)

	allowed = limiter.Allow("test-key", limit, window)
	if !allowed {
		t.Error("should be allowed after window reset")
	}
}

func TestInMemoryRateLimiter_DifferentKeys(t *testing.T) {
	logger := zap.NewNop()
	limiter := NewInMemoryRateLimiter(logger)

	limit := 1

	allowed := limiter.Allow("key-a", limit, time.Minute)
	if !allowed {
		t.Error("key-a first request should be allowed")
	}

	allowed = limiter.Allow("key-a", limit, time.Minute)
	if allowed {
		t.Error("key-a second request should be denied")
	}

	allowed = limiter.Allow("key-b", limit, time.Minute)
	if !allowed {
		t.Error("key-b first request should be allowed (separate from key-a)")
	}
}

func TestRateLimitMiddleware_Allowed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	logger := zap.NewNop()
	limiter := NewInMemoryRateLimiter(logger)

	r.Use(RateLimitMiddleware(limiter, 10, time.Minute))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestRateLimitMiddleware_Denied(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := zap.NewNop()
	limiter := NewInMemoryRateLimiter(logger)

	handler := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}

	limit := 2

	for i := 0; i < limit; i++ {
		w := httptest.NewRecorder()
		_, r := gin.CreateTestContext(w)
		r.Use(RateLimitMiddleware(limiter, limit, time.Minute))
		r.GET("/test", handler)
		req := httptest.NewRequest("GET", "/test", nil)
		r.ServeHTTP(w, req)
	}

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.Use(RateLimitMiddleware(limiter, limit, time.Minute))
	r.GET("/test", handler)
	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("status = %d, want %d", w.Code, http.StatusTooManyRequests)
	}
}

func TestLoginRateLimitMiddleware_Allowed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	logger := zap.NewNop()
	limiter := NewInMemoryRateLimiter(logger)

	r.Use(LoginRateLimitMiddleware(limiter))
	r.POST("/login", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("POST", "/login", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestLoginRateLimitMiddleware_Denied(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := zap.NewNop()
	limiter := NewInMemoryRateLimiter(logger)

	handler := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}

	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		_, r := gin.CreateTestContext(w)
		r.Use(LoginRateLimitMiddleware(limiter))
		r.POST("/login", handler)
		req := httptest.NewRequest("POST", "/login", nil)
		r.ServeHTTP(w, req)
	}

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.Use(LoginRateLimitMiddleware(limiter))
	r.POST("/login", handler)
	req := httptest.NewRequest("POST", "/login", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("status = %d, want %d", w.Code, http.StatusTooManyRequests)
	}
}
