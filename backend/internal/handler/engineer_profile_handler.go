package handler

import (
	"net/http"

	"github.com/duesk/ivy/internal/dto"
	"github.com/duesk/ivy/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// EngineerProfileHandler 人材プロファイルハンドラー
type EngineerProfileHandler struct {
	engineerProfileService service.EngineerProfileService
	logger                 *zap.Logger
}

// NewEngineerProfileHandler 人材プロファイルハンドラーを作成
func NewEngineerProfileHandler(engineerProfileService service.EngineerProfileService, logger *zap.Logger) *EngineerProfileHandler {
	return &EngineerProfileHandler{engineerProfileService: engineerProfileService, logger: logger}
}

// List 人材一覧取得
func (h *EngineerProfileHandler) List(c *gin.Context) {
	var req dto.EngineerProfileListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "リクエストパラメータが不正です"})
		return
	}

	resp, err := h.engineerProfileService.List(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("人材一覧取得失敗", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "人材一覧の取得に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, resp)
}
