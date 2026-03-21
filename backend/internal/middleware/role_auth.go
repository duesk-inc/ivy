package middleware

import (
	"net/http"

	"github.com/duesk/ivy/internal/model"
	"github.com/gin-gonic/gin"
)

// RoleRequired 指定されたロールのいずれかが必要なミドルウェア
func RoleRequired(allowedRoles ...model.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleStr, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "ロール情報が取得できません"})
			c.Abort()
			return
		}

		userRole := model.Role(roleStr.(string))
		for _, allowed := range allowedRoles {
			if userRole == allowed {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "権限が不足しています"})
		c.Abort()
	}
}

// AdminRequired 管理者権限が必要なミドルウェア
func AdminRequired() gin.HandlerFunc {
	return RoleRequired(model.RoleAdmin)
}

// SalesOrAdminRequired 営業または管理者権限が必要なミドルウェア
func SalesOrAdminRequired() gin.HandlerFunc {
	return RoleRequired(model.RoleAdmin, model.RoleSales)
}
