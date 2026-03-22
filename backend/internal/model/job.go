package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// JobStatus 案件ステータス
type JobStatus string

const (
	JobStatusActive   JobStatus = "active"
	JobStatusArchived JobStatus = "archived"
)

// Job メールから抽出した案件情報（Phase 2）
type Job struct {
	ID            string          `gorm:"type:uuid;primaryKey" json:"id"`
	ContentHash   string          `gorm:"size:64;not null;uniqueIndex:uq_jobs_content_hash" json:"content_hash"`
	SourceEmailID string          `gorm:"size:255" json:"source_email_id,omitempty"`
	RawText       string          `gorm:"type:text;not null" json:"raw_text"`
	Parsed        json.RawMessage `gorm:"type:jsonb;not null;default:'{}'" json:"parsed"`
	StartMonth    string          `gorm:"size:7;index" json:"start_month,omitempty"`
	Status        JobStatus       `gorm:"size:20;not null;default:'active';index" json:"status"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	ExpiresAt     *time.Time      `gorm:"index" json:"expires_at,omitempty"`
}

func (Job) TableName() string {
	return "jobs"
}

func (j *Job) BeforeCreate(tx *gorm.DB) error {
	if j.ID == "" {
		j.ID = uuid.New().String()
	}
	return nil
}

// ParsedJobData 案件のパース結果構造体
type ParsedJobData struct {
	Name          string   `json:"name,omitempty"`
	Skills        []string `json:"skills,omitempty"`
	RateMin       *int     `json:"rate_min,omitempty"`
	RateMax       *int     `json:"rate_max,omitempty"`
	Location      string   `json:"location,omitempty"`
	Remote        string   `json:"remote,omitempty"`
	StartMonth    string   `json:"start_month,omitempty"`
	Settlement    string   `json:"settlement,omitempty"`
	NationalityOK *bool   `json:"nationality_ok,omitempty"`
	FreelanceOK   *bool   `json:"freelance_ok,omitempty"`
	AgeLimit      *int     `json:"age_limit,omitempty"`
	Conditions    string   `json:"conditions,omitempty"`
}

// GetParsedData Parsed JSONBをParsedJobDataに変換
func (j *Job) GetParsedData() (*ParsedJobData, error) {
	var data ParsedJobData
	if err := json.Unmarshal(j.Parsed, &data); err != nil {
		return nil, err
	}
	return &data, nil
}
