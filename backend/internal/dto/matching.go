package dto

import (
	"encoding/json"
	"time"
)

// MatchingRequest マッチングリクエスト
type MatchingRequest struct {
	JobText         string          `json:"job_text" binding:"required"`
	EngineerText    string          `json:"engineer_text"`
	EngineerFileKey string          `json:"engineer_file_key"`
	Supplement      *SupplementInfo `json:"supplement"`
}

// SupplementInfo 補足情報
type SupplementInfo struct {
	AffiliationType   string `json:"affiliation_type,omitempty"` // "duesk" or "partner"
	AffiliationName   string `json:"affiliation_name,omitempty"`
	Rate              int    `json:"rate,omitempty"`             // 希望単価（万円）
	Nationality       string `json:"nationality,omitempty"`      // "japanese" or other
	EmploymentType    string `json:"employment_type,omitempty"`  // "employee", "freelance"
	AvailableFrom     string `json:"available_from,omitempty"`   // "2026-04"
	SupplyChainLevel  int    `json:"supply_chain_level,omitempty"`
	SupplyChainSource string `json:"supply_chain_source,omitempty"`
}

// MatchingResponse マッチングレスポンス
type MatchingResponse struct {
	ID                string          `json:"id"`
	TotalScore        int             `json:"total_score"`
	Grade             string          `json:"grade"`
	GradeLabel        string          `json:"grade_label"`
	Result            json.RawMessage `json:"result"`
	ModelUsed         string          `json:"model_used"`
	TokensUsed        int             `json:"tokens_used"`
	JobGroupID        *string         `json:"job_group_id,omitempty"`
	SupplyChainLevel  int             `json:"supply_chain_level"`
	SupplyChainSource string          `json:"supply_chain_source,omitempty"`
	CreatedAt         time.Time       `json:"created_at"`
}

// MatchingListRequest マッチング履歴一覧リクエスト
type MatchingListRequest struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=20"`
	Grade    string `form:"grade"`
}

// MatchingListResponse マッチング履歴一覧レスポンス
type MatchingListResponse struct {
	Items      []MatchingListItem `json:"items"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalPages int                `json:"total_pages"`
}

// MatchingListItem 履歴一覧の各アイテム
type MatchingListItem struct {
	ID                string          `json:"id"`
	TotalScore        int             `json:"total_score"`
	Grade             string          `json:"grade"`
	GradeLabel        string          `json:"grade_label"`
	JobSummary        json.RawMessage `json:"job_summary,omitempty"`
	ModelUsed         string          `json:"model_used"`
	JobGroupID        *string         `json:"job_group_id,omitempty"`
	SupplyChainLevel  int             `json:"supply_chain_level"`
	SupplyChainSource string          `json:"supply_chain_source,omitempty"`
	CreatedAt         time.Time       `json:"created_at"`
}

// MatchingDetailResponse マッチング詳細レスポンス
type MatchingDetailResponse struct {
	ID                string          `json:"id"`
	JobText           string          `json:"job_text"`
	EngineerText      string          `json:"engineer_text"`
	EngineerFileKey   string          `json:"engineer_file_key,omitempty"`
	Supplement        json.RawMessage `json:"supplement"`
	TotalScore        int             `json:"total_score"`
	Grade             string          `json:"grade"`
	GradeLabel        string          `json:"grade_label"`
	Result            json.RawMessage `json:"result"`
	ModelUsed         string          `json:"model_used"`
	TokensUsed        int             `json:"tokens_used"`
	JobGroupID        *string         `json:"job_group_id,omitempty"`
	SupplyChainLevel  int             `json:"supply_chain_level"`
	SupplyChainSource string          `json:"supply_chain_source,omitempty"`
	CreatedAt         time.Time       `json:"created_at"`
}

// CreateJobGroupRequest 案件グループ作成リクエスト
type CreateJobGroupRequest struct {
	Name       string `json:"name" binding:"required"`
	MatchingID string `json:"matching_id" binding:"required"`
}

// LinkToJobGroupRequest 案件グループ紐付けリクエスト
type LinkToJobGroupRequest struct {
	JobGroupID string `json:"job_group_id" binding:"required"`
}

// JobGroupResponse 案件グループレスポンス
type JobGroupResponse struct {
	ID        string             `json:"id"`
	Name      string             `json:"name"`
	Matchings []MatchingListItem `json:"matchings"`
	BestRoute *MatchingListItem  `json:"best_route,omitempty"`
	CreatedAt time.Time          `json:"created_at"`
}
