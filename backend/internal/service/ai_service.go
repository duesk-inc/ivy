package service

import "context"

// MatchRequest AIマッチングリクエスト
type MatchRequest struct {
	JobText      string
	EngineerText string
	Supplement   map[string]interface{}
	MarginAmount int    // マージン金額（円）
	MarginType   string // "fixed" or "percentage"
}

// MatchResponse AIマッチングレスポンス
type MatchResponse struct {
	TotalScore int
	Grade      string
	GradeLabel string
	Result     []byte // JSON raw bytes
	TokensUsed int
	ModelUsed  string
}

// AIService AIマッチングサービスインターフェース
type AIService interface {
	Match(ctx context.Context, req MatchRequest) (*MatchResponse, error)
}
