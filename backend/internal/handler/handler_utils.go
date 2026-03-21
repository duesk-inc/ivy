package handler

import (
	"net/http"

	"github.com/duesk/ivy/internal/dto"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// HandleError エラーレスポンスを返す（ログ出力付き）
func HandleError(c *gin.Context, status int, message string, logger *zap.Logger, err error, keyvals ...interface{}) {
	fields := []zap.Field{zap.Error(err)}
	for i := 0; i+1 < len(keyvals); i += 2 {
		key, ok := keyvals[i].(string)
		if !ok {
			continue
		}
		fields = append(fields, zap.Any(key, keyvals[i+1]))
	}
	logger.Error(message, fields...)
	c.JSON(status, dto.ErrorResponse{Error: message})
}

// RespondValidationError バリデーションエラーレスポンスを返す
func RespondValidationError(c *gin.Context, details map[string]string) {
	c.JSON(http.StatusBadRequest, dto.ErrorResponse{
		Error:   "リクエストが不正です",
		Code:    "VALIDATION_ERROR",
		Details: details,
	})
}
