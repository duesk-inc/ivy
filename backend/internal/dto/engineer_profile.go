package dto

import (
	"encoding/json"
	"time"
)

// EngineerProfileListRequest 人材一覧リクエスト
type EngineerProfileListRequest struct {
	Page       int    `form:"page,default=1"`
	PageSize   int    `form:"page_size,default=20"`
	StartMonth string `form:"start_month"`
	Status     string `form:"status"`
}

// EngineerProfileResponse 人材レスポンス
type EngineerProfileResponse struct {
	ID         string          `json:"id"`
	RawText    string          `json:"raw_text"`
	Parsed     json.RawMessage `json:"parsed"`
	FileKey    string          `json:"file_key,omitempty"`
	StartMonth string          `json:"start_month,omitempty"`
	Status     string          `json:"status"`
	CreatedAt  time.Time       `json:"created_at"`
	ExpiresAt  *time.Time      `json:"expires_at,omitempty"`
}

// EngineerProfileListResponse 人材一覧レスポンス
type EngineerProfileListResponse struct {
	Items      []EngineerProfileResponse `json:"items"`
	Total      int64                     `json:"total"`
	Page       int                       `json:"page"`
	PageSize   int                       `json:"page_size"`
	TotalPages int                       `json:"total_pages"`
}
