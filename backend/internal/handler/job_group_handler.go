package handler

import (
	"errors"
	"net/http"

	"github.com/duesk/ivy/internal/dto"
	"github.com/duesk/ivy/internal/model"
	"github.com/duesk/ivy/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// JobGroupHandler 案件グループハンドラー
type JobGroupHandler struct {
	matchingService service.MatchingService
	logger          *zap.Logger
}

// NewJobGroupHandler 案件グループハンドラーを作成
func NewJobGroupHandler(matchingService service.MatchingService, logger *zap.Logger) *JobGroupHandler {
	return &JobGroupHandler{matchingService: matchingService, logger: logger}
}

// Create 案件グループ作成
func (h *JobGroupHandler) Create(c *gin.Context) {
	var req dto.CreateJobGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "リクエストが不正です"})
		return
	}

	userID := c.GetString("user_id")

	resp, err := h.matchingService.CreateJobGroup(c.Request.Context(), userID, req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "マッチング結果が見つかりません"})
			return
		}
		h.logger.Error("案件グループ作成失敗", zap.Error(err), zap.String("user_id", userID))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "案件グループの作成に失敗しました"})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// Get 案件グループ取得
func (h *JobGroupHandler) Get(c *gin.Context) {
	id := c.Param("id")

	resp, err := h.matchingService.GetJobGroup(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "案件グループが見つかりません"})
			return
		}
		h.logger.Error("案件グループ取得失敗", zap.Error(err), zap.String("id", id))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "データの取得に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// List 案件グループ一覧取得
func (h *JobGroupHandler) List(c *gin.Context) {
	userID := c.GetString("user_id")
	roleStr := c.GetString("role")
	role := model.Role(roleStr)

	resp, err := h.matchingService.ListJobGroups(c.Request.Context(), userID, role)
	if err != nil {
		h.logger.Error("案件グループ一覧取得失敗", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "データの取得に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Delete 案件グループ削除
func (h *JobGroupHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("user_id")

	if err := h.matchingService.DeleteJobGroup(c.Request.Context(), id, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "案件グループが見つかりません"})
			return
		}
		h.logger.Error("案件グループ削除失敗", zap.Error(err), zap.String("id", id))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "削除に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{Message: "削除しました"})
}
