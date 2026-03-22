package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Setting 設定モデル（key-valueストア）
type Setting struct {
	ID        string          `gorm:"type:uuid;primaryKey" json:"id"`
	Key       string          `gorm:"size:100;uniqueIndex;not null" json:"key"`
	Value     json.RawMessage `gorm:"type:jsonb;not null" json:"value"`
	UpdatedBy *string         `gorm:"type:uuid" json:"updated_by,omitempty"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// TableName テーブル名を指定
func (Setting) TableName() string {
	return "settings"
}

// BeforeCreate UUID自動生成
func (s *Setting) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return nil
}

// 設定キー定数
const (
	SettingKeyMargin        = "margin"
	SettingKeyAIModel       = "ai_model"
	SettingKeyDataRetention = "data_retention"
)

// MarginSetting マージン設定
type MarginSetting struct {
	Type   string `json:"type"`   // "fixed" or "percentage"
	Amount int    `json:"amount"` // 固定金額（円）またはパーセンテージ
}

// AIModelSetting AIモデル設定
type AIModelSetting struct {
	Model string `json:"model"`
}

// DataRetentionSetting データ保持期間設定
type DataRetentionSetting struct {
	JobsDays             int `json:"jobs_days"`
	EngineersDays        int `json:"engineers_days"`
	MatchingsDays        int `json:"matchings_days"`
	ProcessedEmailsDays  int `json:"processed_emails_days"`
}

// GetProcessedEmailsDaysOrDefault processed_emails_daysのデフォルト値フォールバック
func (d *DataRetentionSetting) GetProcessedEmailsDaysOrDefault() int {
	if d.ProcessedEmailsDays <= 0 {
		return 180
	}
	return d.ProcessedEmailsDays
}

// ParseMarginSetting JSON→MarginSetting
func ParseMarginSetting(data json.RawMessage) (*MarginSetting, error) {
	var s MarginSetting
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// ParseAIModelSetting JSON→AIModelSetting
func ParseAIModelSetting(data json.RawMessage) (*AIModelSetting, error) {
	var s AIModelSetting
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// ParseDataRetentionSetting JSON→DataRetentionSetting
func ParseDataRetentionSetting(data json.RawMessage) (*DataRetentionSetting, error) {
	var s DataRetentionSetting
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}
