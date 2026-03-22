package service

import (
	"context"
	"encoding/json"
)

// mockEmailAIService テスト/開発用のモックメールAIサービス
type mockEmailAIService struct{}

// NewMockEmailAIService モックメールAIサービスを作成
func NewMockEmailAIService() EmailAIService {
	return &mockEmailAIService{}
}

// ClassifyEmail モック分類（常にjobを返す）
func (s *mockEmailAIService) ClassifyEmail(ctx context.Context, subject, body, sender string) (*EmailClassification, error) {
	return &EmailClassification{
		Classification: "job",
		Confidence:     0.9,
	}, nil
}

// ParseJobFromEmail モック案件パース（固定JSONを返す）
func (s *mockEmailAIService) ParseJobFromEmail(ctx context.Context, emailText string) (json.RawMessage, error) {
	result := map[string]any{
		"name":           "Webアプリケーション開発",
		"skills":         []string{"Java", "Spring Boot", "PostgreSQL"},
		"rate_min":       60,
		"rate_max":       75,
		"location":       "東京都渋谷区",
		"remote":         "一部リモート",
		"start_month":    "2026-04",
		"settlement":     "140-180h",
		"nationality_ok": true,
		"freelance_ok":   true,
		"age_limit":      nil,
		"conditions":     "Java経験3年以上、チーム開発経験必須",
	}
	resultJSON, _ := json.Marshal(result)
	return json.RawMessage(resultJSON), nil
}

// ParseEngineerFromEmail モック人材パース（固定JSONを返す）
func (s *mockEmailAIService) ParseEngineerFromEmail(ctx context.Context, emailText, attachmentText string) (json.RawMessage, error) {
	result := map[string]any{
		"initials":        "T.Y.",
		"age":             30,
		"gender":          "男性",
		"skills":          []string{"Java", "Spring Boot", "Docker", "PostgreSQL"},
		"rate":            60,
		"start_month":     "2026-04",
		"nationality":     "日本",
		"employment_type": "フリーランス",
		"affiliation":     "株式会社SasaTech",
		"nearest_station": "池袋駅",
	}
	resultJSON, _ := json.Marshal(result)
	return json.RawMessage(resultJSON), nil
}
