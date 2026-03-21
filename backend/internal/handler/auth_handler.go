package handler

import (
	"net/http"
	"strings"

	"github.com/duesk/ivy/internal/dto"
	"github.com/duesk/ivy/internal/model"
	"github.com/duesk/ivy/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuthHandler 認証ハンドラー
type AuthHandler struct {
	authService service.AuthService
	logger      *zap.Logger
}

// NewAuthHandler 認証ハンドラーを作成
func NewAuthHandler(authService service.AuthService, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{authService: authService, logger: logger}
}

// Login ログイン
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "リクエストが不正です"})
		return
	}

	resp, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		h.logger.Error("ログイン失敗", zap.Error(err), zap.String("email", req.Email))
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "メールアドレスまたはパスワードが正しくありません"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Refresh トークンリフレッシュ
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "リクエストが不正です"})
		return
	}

	resp, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		h.logger.Error("トークンリフレッシュ失敗", zap.Error(err))
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "トークンリフレッシュに失敗しました"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Logout ログアウト
func (h *AuthHandler) Logout(c *gin.Context) {
	token := ""
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 {
			token = parts[1]
		}
	}

	if err := h.authService.Logout(c.Request.Context(), token); err != nil {
		h.logger.Warn("ログアウト処理エラー", zap.Error(err))
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{Message: "ログアウトしました"})
}

// Me 現在のユーザー情報を返す
func (h *AuthHandler) Me(c *gin.Context) {
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "ユーザー情報が取得できません"})
		return
	}

	user, ok := userInterface.(*model.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "ユーザー情報の形式が無効です"})
		return
	}

	c.JSON(http.StatusOK, dto.UserResponse{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
		Role:  string(user.Role),
	})
}
