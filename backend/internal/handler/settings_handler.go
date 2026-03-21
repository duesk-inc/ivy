package handler

import (
	"errors"
	"net/http"

	"github.com/duesk/ivy/internal/dto"
	"github.com/duesk/ivy/internal/repository"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// SettingsHandler 設定ハンドラー
type SettingsHandler struct {
	settingsRepo repository.SettingsRepository
	logger       *zap.Logger
}

// NewSettingsHandler 設定ハンドラーを作成
func NewSettingsHandler(settingsRepo repository.SettingsRepository, logger *zap.Logger) *SettingsHandler {
	return &SettingsHandler{settingsRepo: settingsRepo, logger: logger}
}

// GetAll 全設定取得
func (h *SettingsHandler) GetAll(c *gin.Context) {
	settings, err := h.settingsRepo.GetAll(c.Request.Context())
	if err != nil {
		h.logger.Error("設定取得失敗", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "設定の取得に失敗しました"})
		return
	}

	items := make([]dto.SettingItem, len(settings))
	for i, s := range settings {
		items[i] = dto.SettingItem{
			Key:   s.Key,
			Value: s.Value,
		}
	}

	c.JSON(http.StatusOK, dto.SettingsResponse{Settings: items})
}

// Update 設定更新（adminのみ）
func (h *SettingsHandler) Update(c *gin.Context) {
	key := c.Param("key")

	var req dto.UpdateSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "リクエストが不正です"})
		return
	}

	userID := c.GetString("user_id")

	if err := h.settingsRepo.Update(c.Request.Context(), key, req.Value, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "設定が見つかりません"})
			return
		}
		h.logger.Error("設定更新失敗", zap.Error(err), zap.String("key", key))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "設定の更新に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{Message: "設定を更新しました"})
}
