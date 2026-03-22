package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EngineerProfileStatus 人材プロファイルステータス
type EngineerProfileStatus string

const (
	EngineerProfileStatusActive   EngineerProfileStatus = "active"
	EngineerProfileStatusArchived EngineerProfileStatus = "archived"
)

// EngineerProfile メールから抽出した人材情報（Phase 2）
type EngineerProfile struct {
	ID            string                `gorm:"type:uuid;primaryKey" json:"id"`
	ContentHash   string                `gorm:"size:64;not null;uniqueIndex:uq_engineer_profiles_content_hash" json:"content_hash"`
	SourceEmailID string                `gorm:"size:255" json:"source_email_id,omitempty"`
	RawText       string                `gorm:"type:text;not null" json:"raw_text"`
	FileKey       string                `gorm:"size:500" json:"file_key,omitempty"`
	Parsed        json.RawMessage       `gorm:"type:jsonb;not null;default:'{}'" json:"parsed"`
	StartMonth    string                `gorm:"size:7;index" json:"start_month,omitempty"`
	Status        EngineerProfileStatus `gorm:"size:20;not null;default:'active';index" json:"status"`
	CreatedAt     time.Time             `json:"created_at"`
	UpdatedAt     time.Time             `json:"updated_at"`
	ExpiresAt     *time.Time            `gorm:"index" json:"expires_at,omitempty"`
}

func (EngineerProfile) TableName() string {
	return "engineer_profiles"
}

func (e *EngineerProfile) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = uuid.New().String()
	}
	return nil
}

// ParsedEngineerData 人材のパース結果構造体
type ParsedEngineerData struct {
	Initials       string   `json:"initials,omitempty"`
	Age            *int     `json:"age,omitempty"`
	Gender         string   `json:"gender,omitempty"`
	Skills         []string `json:"skills,omitempty"`
	Rate           *int     `json:"rate,omitempty"`
	StartMonth     string   `json:"start_month,omitempty"`
	Nationality    string   `json:"nationality,omitempty"`
	EmploymentType string   `json:"employment_type,omitempty"`
	Affiliation    string   `json:"affiliation,omitempty"`
	NearestStation string   `json:"nearest_station,omitempty"`
}

// GetParsedData Parsed JSONBをParsedEngineerDataに変換
func (e *EngineerProfile) GetParsedData() (*ParsedEngineerData, error) {
	var data ParsedEngineerData
	if err := json.Unmarshal(e.Parsed, &data); err != nil {
		return nil, err
	}
	return &data, nil
}
