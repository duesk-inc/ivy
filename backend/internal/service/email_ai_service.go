package service

import (
	"context"
	"encoding/json"
)

// EmailClassification メール分類結果
type EmailClassification struct {
	Classification string  `json:"classification"` // "job", "engineer", "other"
	Confidence     float64 `json:"confidence"`
}

// EmailAIService メール分類・パース用AIサービスインターフェース（既存AIServiceとは分離）
type EmailAIService interface {
	ClassifyEmail(ctx context.Context, subject, body, sender string) (*EmailClassification, error)
	ParseJobFromEmail(ctx context.Context, emailText string) (json.RawMessage, error)
	ParseEngineerFromEmail(ctx context.Context, emailText, attachmentText string) (json.RawMessage, error)
}
