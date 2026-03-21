package service

import (
	"context"
	"strings"
	"testing"

	"github.com/duesk/ivy/internal/config"
	"github.com/duesk/ivy/internal/dto"
	"github.com/duesk/ivy/internal/model"
	"go.uber.org/zap"
)

type mockUserRepo struct{}

func (m *mockUserRepo) GetByID(ctx context.Context, id string) (*model.User, error) {
	return nil, nil
}

func (m *mockUserRepo) GetByCognitoSub(ctx context.Context, sub string) (*model.User, error) {
	return nil, nil
}

func (m *mockUserRepo) Create(ctx context.Context, user *model.User) error {
	return nil
}

func newTestAuthService() AuthService {
	logger := zap.NewNop()
	cfg := &config.Config{
		Cognito: config.CognitoConfig{
			Enabled: false,
		},
	}
	cfg.Cognito.SetDefaults()
	return NewAuthService(cfg, &mockUserRepo{}, logger)
}

func TestAuthService_DevLogin(t *testing.T) {
	svc := newTestAuthService()

	req := dto.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	resp, err := svc.Login(context.Background(), req)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if resp.AccessToken == "" {
		t.Error("expected non-empty AccessToken")
	}
	if !strings.HasPrefix(resp.AccessToken, "dev.") {
		t.Errorf("expected AccessToken to start with 'dev.', got %q", resp.AccessToken)
	}
	if resp.RefreshToken != "dev-refresh-token" {
		t.Errorf("RefreshToken = %q, want %q", resp.RefreshToken, "dev-refresh-token")
	}
	if resp.ExpiresIn != 3600 {
		t.Errorf("ExpiresIn = %d, want 3600", resp.ExpiresIn)
	}
}

func TestAuthService_DevRefreshToken(t *testing.T) {
	svc := newTestAuthService()

	resp, err := svc.RefreshToken(context.Background(), "dev-refresh-token")
	if err != nil {
		t.Fatalf("RefreshToken failed: %v", err)
	}

	if resp.AccessToken == "" {
		t.Error("expected non-empty AccessToken")
	}
	if !strings.HasPrefix(resp.AccessToken, "dev.") {
		t.Errorf("expected AccessToken to start with 'dev.', got %q", resp.AccessToken)
	}
	if resp.RefreshToken != "dev-refresh-token" {
		t.Errorf("RefreshToken = %q, want %q", resp.RefreshToken, "dev-refresh-token")
	}
	if resp.ExpiresIn != 3600 {
		t.Errorf("ExpiresIn = %d, want 3600", resp.ExpiresIn)
	}
}

func TestAuthService_DevLogout(t *testing.T) {
	svc := newTestAuthService()

	err := svc.Logout(context.Background(), "dev-access-token")
	if err != nil {
		t.Fatalf("Logout failed: %v", err)
	}
}

func TestAuthService_DevLogin_ReturnsUser(t *testing.T) {
	svc := newTestAuthService()

	req := dto.LoginRequest{
		Email:    "admin@example.com",
		Password: "password123",
	}

	resp, err := svc.Login(context.Background(), req)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if resp.User.ID == "" {
		t.Error("expected non-empty User.ID")
	}
	if resp.User.ID != "00000000-0000-0000-0000-000000000001" {
		t.Errorf("User.ID = %q, want %q", resp.User.ID, "00000000-0000-0000-0000-000000000001")
	}
	if resp.User.Email != "admin@example.com" {
		t.Errorf("User.Email = %q, want %q", resp.User.Email, "admin@example.com")
	}
	if resp.User.Name == "" {
		t.Error("expected non-empty User.Name")
	}
	if resp.User.Role != "admin" {
		t.Errorf("User.Role = %q, want %q", resp.User.Role, "admin")
	}
}
