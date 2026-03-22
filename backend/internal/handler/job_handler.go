package handler

import (
	"net/http"

	"github.com/duesk/ivy/internal/dto"
	"github.com/duesk/ivy/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// JobHandler 案件ハンドラー
type JobHandler struct {
	jobService service.JobService
	logger     *zap.Logger
}

// NewJobHandler 案件ハンドラーを作成
func NewJobHandler(jobService service.JobService, logger *zap.Logger) *JobHandler {
	return &JobHandler{jobService: jobService, logger: logger}
}

// List 案件一覧取得
func (h *JobHandler) List(c *gin.Context) {
	var req dto.JobListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "リクエストパラメータが不正です"})
		return
	}

	resp, err := h.jobService.List(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("案件一覧取得失敗", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "案件一覧の取得に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, resp)
}
