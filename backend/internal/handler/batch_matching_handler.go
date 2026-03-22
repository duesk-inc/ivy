package handler

import (
	"errors"
	"net/http"

	"github.com/duesk/ivy/internal/dto"
	"github.com/duesk/ivy/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// BatchMatchingHandler バッチマッチングハンドラー
type BatchMatchingHandler struct {
	batchService service.BatchMatchingService
	logger       *zap.Logger
}

// NewBatchMatchingHandler バッチマッチングハンドラーを作成
func NewBatchMatchingHandler(batchService service.BatchMatchingService, logger *zap.Logger) *BatchMatchingHandler {
	return &BatchMatchingHandler{batchService: batchService, logger: logger}
}

// Preview バッチマッチングプレビュー
func (h *BatchMatchingHandler) Preview(c *gin.Context) {
	var req dto.BatchMatchingPreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "リクエストが不正です"})
		return
	}

	resp, err := h.batchService.Preview(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("バッチプレビュー失敗", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "プレビューの取得に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Execute バッチマッチング実行
func (h *BatchMatchingHandler) Execute(c *gin.Context) {
	var req dto.BatchMatchingExecuteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "リクエストが不正です"})
		return
	}

	userID := c.GetString("user_id")

	resp, err := h.batchService.Execute(c.Request.Context(), userID, req)
	if err != nil {
		h.logger.Error("バッチ実行失敗", zap.Error(err), zap.String("user_id", userID))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "バッチマッチングの開始に失敗しました"})
		return
	}

	c.JSON(http.StatusAccepted, resp)
}

// MatchJobToEngineers 案件→人材マッチング
func (h *BatchMatchingHandler) MatchJobToEngineers(c *gin.Context) {
	jobID := c.Param("id")
	userID := c.GetString("user_id")

	var req dto.OneToNMatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// ボディなしでもOK（フィルタなし）
		req = dto.OneToNMatchRequest{}
	}

	resp, err := h.batchService.MatchJobToEngineers(c.Request.Context(), userID, jobID, req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "案件が見つかりません"})
			return
		}
		h.logger.Error("案件→人材マッチング失敗", zap.Error(err), zap.String("job_id", jobID))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "マッチングの開始に失敗しました"})
		return
	}

	c.JSON(http.StatusAccepted, resp)
}

// MatchEngineerToJobs 人材→案件マッチング
func (h *BatchMatchingHandler) MatchEngineerToJobs(c *gin.Context) {
	engineerID := c.Param("id")
	userID := c.GetString("user_id")

	var req dto.OneToNMatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req = dto.OneToNMatchRequest{}
	}

	resp, err := h.batchService.MatchEngineerToJobs(c.Request.Context(), userID, engineerID, req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "人材が見つかりません"})
			return
		}
		h.logger.Error("人材→案件マッチング失敗", zap.Error(err), zap.String("engineer_id", engineerID))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "マッチングの開始に失敗しました"})
		return
	}

	c.JSON(http.StatusAccepted, resp)
}

// GetStatus バッチマッチングステータス取得
func (h *BatchMatchingHandler) GetStatus(c *gin.Context) {
	id := c.Param("id")

	resp, err := h.batchService.GetStatus(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "バッチマッチングが見つかりません"})
			return
		}
		h.logger.Error("バッチステータス取得失敗", zap.Error(err), zap.String("batch_id", id))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "ステータスの取得に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, resp)
}
