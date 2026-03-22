package service

import (
	"context"
	"fmt"

	"github.com/duesk/ivy/internal/dto"
	"github.com/duesk/ivy/internal/model"
	"go.uber.org/zap"
)

// MockGmailService Gmail未設定時のモックサービス
type MockGmailService struct {
	logger *zap.Logger
}

// NewMockGmailService モックGmailServiceを作成
func NewMockGmailService(logger *zap.Logger) GmailService {
	return &MockGmailService{logger: logger}
}

// SyncEmails Gmail未設定のためエラーを返す
func (s *MockGmailService) SyncEmails(ctx context.Context) (*dto.EmailSyncResponse, error) {
	return nil, fmt.Errorf("Gmail連携が設定されていません。GMAIL_ENABLED=true および関連する環境変数を設定してください")
}

// GetSyncState Gmail未設定のためエラーを返す
func (s *MockGmailService) GetSyncState(ctx context.Context) (*model.GmailSyncState, error) {
	return nil, fmt.Errorf("Gmail連携が設定されていません")
}
