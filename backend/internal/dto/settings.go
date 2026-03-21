package dto

import "encoding/json"

// SettingsResponse 全設定レスポンス
type SettingsResponse struct {
	Settings []SettingItem `json:"settings"`
}

// SettingItem 設定アイテム
type SettingItem struct {
	Key   string          `json:"key"`
	Value json.RawMessage `json:"value"`
}

// UpdateSettingRequest 設定更新リクエスト
type UpdateSettingRequest struct {
	Value json.RawMessage `json:"value" binding:"required"`
}
