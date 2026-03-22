package dto

import (
	"encoding/json"
	"time"
)

// JobListRequest 案件一覧リクエスト
type JobListRequest struct {
	Page       int    `form:"page,default=1"`
	PageSize   int    `form:"page_size,default=20"`
	StartMonth string `form:"start_month"`
	Status     string `form:"status"`
}

// JobResponse 案件レスポンス
type JobResponse struct {
	ID         string          `json:"id"`
	RawText    string          `json:"raw_text"`
	Parsed     json.RawMessage `json:"parsed"`
	StartMonth string          `json:"start_month,omitempty"`
	Status     string          `json:"status"`
	CreatedAt  time.Time       `json:"created_at"`
	ExpiresAt  *time.Time      `json:"expires_at,omitempty"`
}

// JobListResponse 案件一覧レスポンス
type JobListResponse struct {
	Items      []JobResponse `json:"items"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	PageSize   int           `json:"page_size"`
	TotalPages int           `json:"total_pages"`
}
