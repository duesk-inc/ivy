package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// HealthHandler ヘルスチェックハンドラー
type HealthHandler struct {
	db     *gorm.DB
	redis  *redis.Client
	logger *zap.Logger
}

// NewHealthHandler ヘルスチェックハンドラーを作成
func NewHealthHandler(db *gorm.DB, redis *redis.Client, logger *zap.Logger) *HealthHandler {
	return &HealthHandler{db: db, redis: redis, logger: logger}
}

// HealthCheck ヘルスチェック
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	status := "ok"
	checks := make(map[string]string)

	// DB接続チェック
	sqlDB, err := h.db.DB()
	if err != nil {
		status = "degraded"
		checks["database"] = "error: " + err.Error()
	} else if err := sqlDB.Ping(); err != nil {
		status = "degraded"
		checks["database"] = "error: " + err.Error()
	} else {
		checks["database"] = "ok"
	}

	// Redis接続チェック
	if h.redis != nil {
		if err := h.redis.Ping(c.Request.Context()).Err(); err != nil {
			checks["redis"] = "error: " + err.Error()
		} else {
			checks["redis"] = "ok"
		}
	}

	statusCode := http.StatusOK
	if status != "ok" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, gin.H{
		"status": status,
		"checks": checks,
	})
}
