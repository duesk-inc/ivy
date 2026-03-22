package dto

import (
	"encoding/json"
	"time"
)

// BatchMatchingPreviewRequest バッチマッチングプレビューリクエスト
type BatchMatchingPreviewRequest struct {
	StartMonthFrom string `json:"start_month_from" binding:"required"`
	StartMonthTo   string `json:"start_month_to" binding:"required"`
}

// BatchMatchingPreviewResponse バッチマッチングプレビューレスポンス
type BatchMatchingPreviewResponse struct {
	TotalJobs        int     `json:"total_jobs"`
	TotalEngineers   int     `json:"total_engineers"`
	PairsAfterFilter int     `json:"pairs_after_filter"`
	EstimatedCost    float64 `json:"estimated_cost"`
}

// BatchMatchingExecuteRequest バッチマッチング実行リクエスト
type BatchMatchingExecuteRequest struct {
	StartMonthFrom string `json:"start_month_from" binding:"required"`
	StartMonthTo   string `json:"start_month_to" binding:"required"`
}

// BatchMatchingResponse バッチマッチングレスポンス
type BatchMatchingResponse struct {
	ID             string          `json:"id"`
	BatchType      string          `json:"batch_type"`
	StartMonthFrom string          `json:"start_month_from"`
	StartMonthTo   string          `json:"start_month_to"`
	TotalPairs     int             `json:"total_pairs"`
	SuccessCount   int             `json:"success_count"`
	FailureCount   int             `json:"failure_count"`
	Status         string          `json:"status"`
	Results        json.RawMessage `json:"results"`
	CreatedAt      time.Time       `json:"created_at"`
	CompletedAt    *time.Time      `json:"completed_at,omitempty"`
}

// OneToNMatchRequest 1:Nマッチングリクエスト
type OneToNMatchRequest struct {
	StartMonthFrom string `json:"start_month_from"`
	StartMonthTo   string `json:"start_month_to"`
}
