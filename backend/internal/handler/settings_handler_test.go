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
	"gorm.io/gorm"
)

type mockSettingsRepo struct {
	settings  []model.Setting
	getAllErr error
	setting   *model.Setting
	getErr    error
	updateErr error
}

func (m *mockSettingsRepo) GetAll(ctx context.Context) ([]model.Setting, error) {
	return m.settings, m.getAllErr
}

func (m *mockSettingsRepo) GetByKey(ctx context.Context, key string) (*model.Setting, error) {
	return m.setting, m.getErr
}

func (m *mockSettingsRepo) Update(ctx context.Context, key string, value json.RawMessage, updatedBy string) error {
	return m.updateErr
}

func setupSettingsRouter(handler *SettingsHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	authorized := r.Group("/")
	authorized.Use(func(c *gin.Context) {
		c.Set("user_id", "test-user-id")
		c.Set("role", "admin")
		c.Next()
	})

	authorized.GET("/settings", handler.GetAll)
	authorized.PUT("/settings/:key", handler.Update)

	return r
}

func TestSettingsHandler_GetAll_Success(t *testing.T) {
	mock := &mockSettingsRepo{
		settings: []model.Setting{
			{
				ID:    "setting-1",
				Key:   "ai_model",
				Value: json.RawMessage(`"claude-3-opus"`),
			},
			{
				ID:    "setting-2",
				Key:   "margin_rate",
				Value: json.RawMessage(`10`),
			},
		},
	}
	handler := NewSettingsHandler(mock, zap.NewNop())
	r := setupSettingsRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/settings", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp dto.SettingsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if len(resp.Settings) != 2 {
		t.Errorf("expected 2 settings, got %d", len(resp.Settings))
	}
	if resp.Settings[0].Key != "ai_model" {
		t.Errorf("expected key 'ai_model', got '%s'", resp.Settings[0].Key)
	}
}

func TestSettingsHandler_GetAll_Error(t *testing.T) {
	mock := &mockSettingsRepo{
		getAllErr: errors.New("database error"),
	}
	handler := NewSettingsHandler(mock, zap.NewNop())
	r := setupSettingsRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/settings", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

func TestSettingsHandler_Update_Success(t *testing.T) {
	mock := &mockSettingsRepo{}
	handler := NewSettingsHandler(mock, zap.NewNop())
	r := setupSettingsRouter(handler)

	body := `{"value":"claude-3-5-sonnet"}`
	req := httptest.NewRequest(http.MethodPut, "/settings/ai_model", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
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

func TestSettingsHandler_Update_InvalidJSON(t *testing.T) {
	mock := &mockSettingsRepo{}
	handler := NewSettingsHandler(mock, zap.NewNop())
	r := setupSettingsRouter(handler)

	body := `{invalid`
	req := httptest.NewRequest(http.MethodPut, "/settings/ai_model", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestSettingsHandler_Update_NotFound(t *testing.T) {
	mock := &mockSettingsRepo{
		updateErr: gorm.ErrRecordNotFound,
	}
	handler := NewSettingsHandler(mock, zap.NewNop())
	r := setupSettingsRouter(handler)

	body := `{"value":"some-value"}`
	req := httptest.NewRequest(http.MethodPut, "/settings/nonexistent_key", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestSettingsHandler_Update_Error(t *testing.T) {
	mock := &mockSettingsRepo{
		updateErr: errors.New("database error"),
	}
	handler := NewSettingsHandler(mock, zap.NewNop())
	r := setupSettingsRouter(handler)

	body := `{"value":"some-value"}`
	req := httptest.NewRequest(http.MethodPut, "/settings/ai_model", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}
