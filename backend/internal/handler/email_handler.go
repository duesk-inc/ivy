package handler

import (
	"net/http"

	"github.com/duesk/ivy/internal/dto"
	"github.com/duesk/ivy/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// EmailHandler メール同期ハンドラー
type EmailHandler struct {
	gmailService service.GmailService
	logger       *zap.Logger
}

// NewEmailHandler メール同期ハンドラーを作成
func NewEmailHandler(gmailService service.GmailService, logger *zap.Logger) *EmailHandler {
	return &EmailHandler{gmailService: gmailService, logger: logger}
}

// Sync Gmail同期実行
func (h *EmailHandler) Sync(c *gin.Context) {
	userID := c.GetString("user_id")
	h.logger.Info("メール同期開始", zap.String("user_id", userID))

	resp, err := h.gmailService.SyncEmails(c.Request.Context())
	if err != nil {
		h.logger.Error("メール同期失敗", zap.Error(err), zap.String("user_id", userID))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "メール同期に失敗しました: " + err.Error()})
		return
	}

	h.logger.Info("メール同期完了",
		zap.Int("total_processed", resp.TotalProcessed),
		zap.Int("new_jobs", resp.NewJobs),
		zap.Int("new_engineers", resp.NewEngineers),
	)

	c.JSON(http.StatusOK, resp)
}

// GetSyncState 同期状態取得
func (h *EmailHandler) GetSyncState(c *gin.Context) {
	state, err := h.gmailService.GetSyncState(c.Request.Context())
	if err != nil {
		h.logger.Error("同期状態取得失敗", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "同期状態の取得に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, state)
}
