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

// MatchingHandler マッチングハンドラー
type MatchingHandler struct {
	matchingService service.MatchingService
	logger          *zap.Logger
}

// NewMatchingHandler マッチングハンドラーを作成
func NewMatchingHandler(matchingService service.MatchingService, logger *zap.Logger) *MatchingHandler {
	return &MatchingHandler{matchingService: matchingService, logger: logger}
}

// Execute マッチング実行
func (h *MatchingHandler) Execute(c *gin.Context) {
	var req dto.MatchingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "リクエストが不正です"})
		return
	}

	if req.EngineerText == "" && req.EngineerFileKey == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "エンジニア情報（テキストまたはファイル）が必要です"})
		return
	}

	userID := c.GetString("user_id")

	resp, err := h.matchingService.Execute(c.Request.Context(), userID, req)
	if err != nil {
		h.logger.Error("マッチング実行失敗", zap.Error(err), zap.String("user_id", userID))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "AIサーバーが混雑しています。しばらく待ってから再度お試しください。"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetByID マッチング詳細取得
func (h *MatchingHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	resp, err := h.matchingService.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "マッチング結果が見つかりません"})
			return
		}
		h.logger.Error("マッチング詳細取得失敗", zap.Error(err), zap.String("id", id))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "データの取得に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// List マッチング一覧取得
func (h *MatchingHandler) List(c *gin.Context) {
	var req dto.MatchingListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "リクエストが不正です"})
		return
	}

	userID := c.GetString("user_id")
	roleStr := c.GetString("role")
	role := model.Role(roleStr)

	resp, err := h.matchingService.List(c.Request.Context(), userID, role, req)
	if err != nil {
		h.logger.Error("マッチング一覧取得失敗", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "データの取得に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Delete マッチング結果削除
func (h *MatchingHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("user_id")

	if err := h.matchingService.Delete(c.Request.Context(), id, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "マッチング結果が見つかりません"})
			return
		}
		h.logger.Error("マッチング削除失敗", zap.Error(err), zap.String("id", id))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "削除に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{Message: "削除しました"})
}

// LinkToJobGroup マッチングを案件グループに紐付け
func (h *MatchingHandler) LinkToJobGroup(c *gin.Context) {
	id := c.Param("id")

	var req dto.LinkToJobGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "リクエストが不正です"})
		return
	}

	userID := c.GetString("user_id")

	if err := h.matchingService.LinkToJobGroup(c.Request.Context(), id, req.JobGroupID, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "マッチング結果または案件グループが見つかりません"})
			return
		}
		h.logger.Error("案件グループ紐付け失敗", zap.Error(err), zap.String("matching_id", id))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "紐付けに失敗しました"})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{Message: "紐付けしました"})
}

// UnlinkFromJobGroup マッチングの案件グループ紐付けを解除
func (h *MatchingHandler) UnlinkFromJobGroup(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetString("user_id")

	if err := h.matchingService.UnlinkFromJobGroup(c.Request.Context(), id, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "マッチング結果が見つかりません"})
			return
		}
		h.logger.Error("案件グループ紐付け解除失敗", zap.Error(err), zap.String("matching_id", id))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "紐付け解除に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{Message: "紐付けを解除しました"})
}
