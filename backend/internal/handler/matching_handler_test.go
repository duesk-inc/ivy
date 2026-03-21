package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/duesk/ivy/internal/dto"
	"github.com/duesk/ivy/internal/model"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type mockMatchingService struct {
	executeResp *dto.MatchingResponse
	executeErr  error
	getByIDResp *dto.MatchingDetailResponse
	getByIDErr  error
	listResp    *dto.MatchingListResponse
	listErr     error
	deleteErr   error
}

func (m *mockMatchingService) Execute(ctx context.Context, userID string, req dto.MatchingRequest) (*dto.MatchingResponse, error) {
	return m.executeResp, m.executeErr
}

func (m *mockMatchingService) GetByID(ctx context.Context, id string) (*dto.MatchingDetailResponse, error) {
	return m.getByIDResp, m.getByIDErr
}

func (m *mockMatchingService) List(ctx context.Context, userID string, role model.Role, req dto.MatchingListRequest) (*dto.MatchingListResponse, error) {
	return m.listResp, m.listErr
}

func (m *mockMatchingService) Delete(ctx context.Context, id string, userID string) error {
	return m.deleteErr
}
func (m *mockMatchingService) CreateJobGroup(ctx context.Context, userID string, req dto.CreateJobGroupRequest) (*dto.JobGroupResponse, error) {
	return nil, nil
}
func (m *mockMatchingService) GetJobGroup(ctx context.Context, id string) (*dto.JobGroupResponse, error) {
	return nil, nil
}
func (m *mockMatchingService) ListJobGroups(ctx context.Context, userID string, role model.Role) ([]dto.JobGroupResponse, error) {
	return nil, nil
}
func (m *mockMatchingService) DeleteJobGroup(ctx context.Context, id string, userID string) error {
	return nil
}
func (m *mockMatchingService) LinkToJobGroup(ctx context.Context, matchingID string, jobGroupID string, userID string) error {
	return nil
}
func (m *mockMatchingService) UnlinkFromJobGroup(ctx context.Context, matchingID string, userID string) error {
	return nil
}

func setupMatchingRouter(handler *MatchingHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	authorized := r.Group("/")
	authorized.Use(func(c *gin.Context) {
		c.Set("user_id", "test-user-id")
		c.Set("role", "admin")
		c.Next()
	})

	authorized.POST("/matchings", handler.Execute)
	authorized.GET("/matchings/:id", handler.GetByID)
	authorized.GET("/matchings", handler.List)
	authorized.DELETE("/matchings/:id", handler.Delete)

	return r
}

func TestMatchingHandler_Execute_Success(t *testing.T) {
	mock := &mockMatchingService{
		executeResp: &dto.MatchingResponse{
			ID:         "match-1",
			TotalScore: 85,
			Grade:      "A",
			GradeLabel: "提案推奨",
			Result:     json.RawMessage(`{"skills":90}`),
			ModelUsed:  "claude-3-opus",
			TokensUsed: 1500,
			CreatedAt:  time.Now(),
		},
	}
	handler := NewMatchingHandler(mock, zap.NewNop())
	r := setupMatchingRouter(handler)

	body := `{"job_text":"Go developer needed","engineer_text":"5 years Go experience"}`
	req := httptest.NewRequest(http.MethodPost, "/matchings", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp dto.MatchingResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.ID != "match-1" {
		t.Errorf("expected ID 'match-1', got '%s'", resp.ID)
	}
	if resp.TotalScore != 85 {
		t.Errorf("expected total score 85, got %d", resp.TotalScore)
	}
	if resp.Grade != "A" {
		t.Errorf("expected grade 'A', got '%s'", resp.Grade)
	}
}

func TestMatchingHandler_Execute_InvalidJSON(t *testing.T) {
	mock := &mockMatchingService{}
	handler := NewMatchingHandler(mock, zap.NewNop())
	r := setupMatchingRouter(handler)

	body := `{bad json`
	req := httptest.NewRequest(http.MethodPost, "/matchings", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestMatchingHandler_Execute_NoEngineerInfo(t *testing.T) {
	mock := &mockMatchingService{}
	handler := NewMatchingHandler(mock, zap.NewNop())
	r := setupMatchingRouter(handler)

	body := `{"job_text":"Go developer needed","engineer_text":"","engineer_file_key":""}`
	req := httptest.NewRequest(http.MethodPost, "/matchings", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var resp dto.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if !strings.Contains(resp.Error, "エンジニア情報") {
		t.Errorf("expected error about engineer info, got '%s'", resp.Error)
	}
}

func TestMatchingHandler_Execute_ServiceError(t *testing.T) {
	mock := &mockMatchingService{
		executeErr: errors.New("AI service unavailable"),
	}
	handler := NewMatchingHandler(mock, zap.NewNop())
	r := setupMatchingRouter(handler)

	body := `{"job_text":"Go developer needed","engineer_text":"5 years Go experience"}`
	req := httptest.NewRequest(http.MethodPost, "/matchings", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

func TestMatchingHandler_GetByID_Success(t *testing.T) {
	mock := &mockMatchingService{
		getByIDResp: &dto.MatchingDetailResponse{
			ID:           "match-1",
			JobText:      "Go developer needed",
			EngineerText: "5 years Go experience",
			TotalScore:   85,
			Grade:        "A",
			GradeLabel:   "提案推奨",
			Result:       json.RawMessage(`{"skills":90}`),
			ModelUsed:    "claude-3-opus",
			TokensUsed:   1500,
			CreatedAt:    time.Now(),
		},
	}
	handler := NewMatchingHandler(mock, zap.NewNop())
	r := setupMatchingRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/matchings/match-1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp dto.MatchingDetailResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.ID != "match-1" {
		t.Errorf("expected ID 'match-1', got '%s'", resp.ID)
	}
}

func TestMatchingHandler_GetByID_NotFound(t *testing.T) {
	mock := &mockMatchingService{
		getByIDErr: gorm.ErrRecordNotFound,
	}
	handler := NewMatchingHandler(mock, zap.NewNop())
	r := setupMatchingRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/matchings/nonexistent", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestMatchingHandler_GetByID_Error(t *testing.T) {
	mock := &mockMatchingService{
		getByIDErr: errors.New("database error"),
	}
	handler := NewMatchingHandler(mock, zap.NewNop())
	r := setupMatchingRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/matchings/match-1", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

func TestMatchingHandler_List_Success(t *testing.T) {
	mock := &mockMatchingService{
		listResp: &dto.MatchingListResponse{
			Items: []dto.MatchingListItem{
				{
					ID:         "match-1",
					TotalScore: 85,
					Grade:      "A",
					GradeLabel: "提案推奨",
					ModelUsed:  "claude-3-opus",
					CreatedAt:  time.Now(),
				},
			},
			Total:      1,
			Page:       1,
			PageSize:   20,
			TotalPages: 1,
		},
	}
	handler := NewMatchingHandler(mock, zap.NewNop())
	r := setupMatchingRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/matchings?page=1&page_size=20", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp dto.MatchingListResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if len(resp.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(resp.Items))
	}
	if resp.Total != 1 {
		t.Errorf("expected total 1, got %d", resp.Total)
	}
}

func TestMatchingHandler_List_Error(t *testing.T) {
	mock := &mockMatchingService{
		listErr: errors.New("database error"),
	}
	handler := NewMatchingHandler(mock, zap.NewNop())
	r := setupMatchingRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/matchings", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

func TestMatchingHandler_Delete_Success(t *testing.T) {
	mock := &mockMatchingService{}
	handler := NewMatchingHandler(mock, zap.NewNop())
	r := setupMatchingRouter(handler)

	req := httptest.NewRequest(http.MethodDelete, "/matchings/match-1", nil)
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

func TestMatchingHandler_Delete_NotFound(t *testing.T) {
	mock := &mockMatchingService{
		deleteErr: gorm.ErrRecordNotFound,
	}
	handler := NewMatchingHandler(mock, zap.NewNop())
	r := setupMatchingRouter(handler)

	req := httptest.NewRequest(http.MethodDelete, "/matchings/nonexistent", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}
