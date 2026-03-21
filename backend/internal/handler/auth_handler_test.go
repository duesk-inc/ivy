package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/duesk/ivy/internal/dto"
	"github.com/duesk/ivy/internal/model"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type mockAuthService struct {
	loginResp   *dto.LoginResponse
	loginErr    error
	refreshResp *dto.LoginResponse
	refreshErr  error
	logoutErr   error
}

func (m *mockAuthService) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	return m.loginResp, m.loginErr
}

func (m *mockAuthService) RefreshToken(ctx context.Context, token string) (*dto.LoginResponse, error) {
	return m.refreshResp, m.refreshErr
}

func (m *mockAuthService) Logout(ctx context.Context, token string) error {
	return m.logoutErr
}

func setupAuthRouter(handler *AuthHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/auth/login", handler.Login)
	r.POST("/auth/refresh", handler.Refresh)
	r.POST("/auth/logout", handler.Logout)
	r.GET("/auth/me", handler.Me)
	return r
}

func TestAuthHandler_Login_Success(t *testing.T) {
	mock := &mockAuthService{
		loginResp: &dto.LoginResponse{
			AccessToken:  "access-token-123",
			RefreshToken: "refresh-token-456",
			ExpiresIn:    3600,
			User: dto.UserResponse{
				ID:    "user-1",
				Email: "test@example.com",
				Name:  "Test User",
				Role:  "admin",
			},
		},
	}
	handler := NewAuthHandler(mock, zap.NewNop())
	r := setupAuthRouter(handler)

	body := `{"email":"test@example.com","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp dto.LoginResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.AccessToken != "access-token-123" {
		t.Errorf("expected access token 'access-token-123', got '%s'", resp.AccessToken)
	}
	if resp.RefreshToken != "refresh-token-456" {
		t.Errorf("expected refresh token 'refresh-token-456', got '%s'", resp.RefreshToken)
	}
}

func TestAuthHandler_Login_InvalidJSON(t *testing.T) {
	mock := &mockAuthService{}
	handler := NewAuthHandler(mock, zap.NewNop())
	r := setupAuthRouter(handler)

	body := `{"invalid json`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestAuthHandler_Login_AuthError(t *testing.T) {
	mock := &mockAuthService{
		loginErr: errors.New("invalid credentials"),
	}
	handler := NewAuthHandler(mock, zap.NewNop())
	r := setupAuthRouter(handler)

	body := `{"email":"test@example.com","password":"wrong"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}

	var resp dto.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Error == "" {
		t.Error("expected non-empty error message")
	}
}

func TestAuthHandler_Refresh_Success(t *testing.T) {
	mock := &mockAuthService{
		refreshResp: &dto.LoginResponse{
			AccessToken:  "new-access-token",
			RefreshToken: "new-refresh-token",
			ExpiresIn:    3600,
		},
	}
	handler := NewAuthHandler(mock, zap.NewNop())
	r := setupAuthRouter(handler)

	body := `{"refresh_token":"old-refresh-token"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp dto.LoginResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.AccessToken != "new-access-token" {
		t.Errorf("expected access token 'new-access-token', got '%s'", resp.AccessToken)
	}
}

func TestAuthHandler_Refresh_InvalidJSON(t *testing.T) {
	mock := &mockAuthService{}
	handler := NewAuthHandler(mock, zap.NewNop())
	r := setupAuthRouter(handler)

	body := `not json`
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestAuthHandler_Refresh_Error(t *testing.T) {
	mock := &mockAuthService{
		refreshErr: errors.New("token expired"),
	}
	handler := NewAuthHandler(mock, zap.NewNop())
	r := setupAuthRouter(handler)

	body := `{"refresh_token":"expired-token"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestAuthHandler_Logout_Success(t *testing.T) {
	mock := &mockAuthService{}
	handler := NewAuthHandler(mock, zap.NewNop())
	r := setupAuthRouter(handler)

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer some-access-token")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp dto.SuccessResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Message == "" {
		t.Error("expected non-empty success message")
	}
}

func TestAuthHandler_Me_Success(t *testing.T) {
	mock := &mockAuthService{}
	handler := NewAuthHandler(mock, zap.NewNop())

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/auth/me", func(c *gin.Context) {
		c.Set("user", &model.User{
			ID:    "user-1",
			Email: "test@example.com",
			Name:  "Test User",
			Role:  model.RoleAdmin,
		})
		handler.Me(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp dto.UserResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.ID != "user-1" {
		t.Errorf("expected user ID 'user-1', got '%s'", resp.ID)
	}
	if resp.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got '%s'", resp.Email)
	}
	if resp.Role != "admin" {
		t.Errorf("expected role 'admin', got '%s'", resp.Role)
	}
}

func TestAuthHandler_Me_NoUser(t *testing.T) {
	mock := &mockAuthService{}
	handler := NewAuthHandler(mock, zap.NewNop())
	r := setupAuthRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestAuthHandler_Me_InvalidUserType(t *testing.T) {
	mock := &mockAuthService{}
	handler := NewAuthHandler(mock, zap.NewNop())

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/auth/me", func(c *gin.Context) {
		c.Set("user", "not-a-user-struct")
		handler.Me(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}
