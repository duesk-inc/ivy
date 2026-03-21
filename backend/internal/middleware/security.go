package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SecurityHeaders セキュリティヘッダーを設定するミドルウェア
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}

// HSTSHeader ALB経由のHTTPSリクエストにHSTSヘッダーを設定
func HSTSHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Header.Get("X-Forwarded-Proto") == "https" {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		c.Next()
	}
}

// JSONSizeLimit JSONペイロードのサイズ制限
func JSONSizeLimit(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxSize {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": "リクエストサイズが上限を超えています",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
