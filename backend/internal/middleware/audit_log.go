package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuditLog 監査ログミドルウェア（設計書セクション8）
func AuditLog(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// リクエストボディを読み取り（監査用）
		var requestBody []byte
		if c.Request.Body != nil && (c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "DELETE") {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		c.Next()

		duration := time.Since(start)
		userID, _ := c.Get("user_id")
		role, _ := c.Get("role")

		// 変更を伴うリクエストのみ監査ログを記録
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "DELETE" {
			logger.Info("AUDIT",
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.Int("status", c.Writer.Status()),
				zap.Duration("duration", duration),
				zap.Any("user_id", userID),
				zap.Any("role", role),
				zap.String("client_ip", c.ClientIP()),
				zap.Int("body_size", len(requestBody)),
			)
		}
	}
}
