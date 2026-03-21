package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// InMemoryRateLimiter メモリベースのレート制限
type InMemoryRateLimiter struct {
	mu      sync.RWMutex
	buckets map[string]*bucket
	logger  *zap.Logger
}

type bucket struct {
	count      int
	resetAt    time.Time
	windowSize time.Duration
	mu         sync.Mutex
}

// NewInMemoryRateLimiter インスタンスを生成
func NewInMemoryRateLimiter(logger *zap.Logger) *InMemoryRateLimiter {
	limiter := &InMemoryRateLimiter{
		buckets: make(map[string]*bucket),
		logger:  logger,
	}
	go limiter.cleanup()
	return limiter
}

// Allow 指定されたキーに対してリクエストを許可するかチェック
func (r *InMemoryRateLimiter) Allow(key string, limit int, window time.Duration) bool {
	r.mu.Lock()
	b, exists := r.buckets[key]
	if !exists {
		b = &bucket{
			count:      0,
			resetAt:    time.Now().Add(window),
			windowSize: window,
		}
		r.buckets[key] = b
	}
	r.mu.Unlock()

	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	if now.After(b.resetAt) {
		b.count = 1
		b.resetAt = now.Add(b.windowSize)
		return true
	}

	if b.count >= limit {
		return false
	}

	b.count++
	return true
}

func (r *InMemoryRateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		r.mu.Lock()
		for key, b := range r.buckets {
			b.mu.Lock()
			if now.After(b.resetAt) {
				delete(r.buckets, key)
			}
			b.mu.Unlock()
		}
		r.mu.Unlock()
	}
}

// RateLimitMiddleware APIのレート制限ミドルウェア
func RateLimitMiddleware(limiter *InMemoryRateLimiter, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := fmt.Sprintf("api:%s", c.ClientIP())
		if !limiter.Allow(key, limit, window) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "リクエスト数が上限に達しました。しばらく待ってから再度お試しください",
				"code":  "RATE_LIMIT_EXCEEDED",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// LoginRateLimitMiddleware ログイン試行のレート制限ミドルウェア
func LoginRateLimitMiddleware(limiter *InMemoryRateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := fmt.Sprintf("login:%s", c.ClientIP())
		if !limiter.Allow(key, 5, 10*time.Minute) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "ログイン試行回数が上限に達しました。しばらく待ってから再度お試しください",
				"code":  "LOGIN_RATE_LIMIT_EXCEEDED",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
