package handler

import (
	"net/http"

	"github.com/duesk/ivy/internal/dto"
	"github.com/duesk/ivy/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AdminHandler 管理者ハンドラー
type AdminHandler struct {
	retentionService service.RetentionService
	logger           *zap.Logger
}

// NewAdminHandler 管理者ハンドラーを作成
func NewAdminHandler(retentionService service.RetentionService, logger *zap.Logger) *AdminHandler {
	return &AdminHandler{retentionService: retentionService, logger: logger}
}

// RunRetention データクリーンアップ実行
func (h *AdminHandler) RunRetention(c *gin.Context) {
	userID := c.GetString("user_id")
	h.logger.Info("手動データクリーンアップ開始", zap.String("user_id", userID))

	result, err := h.retentionService.RunCleanup(c.Request.Context())
	if err != nil {
		h.logger.Error("データクリーンアップ失敗", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "データクリーンアップに失敗しました"})
		return
	}

	if err := h.retentionService.RecordLastCleanup(c.Request.Context()); err != nil {
		h.logger.Warn("クリーンアップ実行日時記録失敗", zap.Error(err))
	}

	c.JSON(http.StatusOK, result)
}
