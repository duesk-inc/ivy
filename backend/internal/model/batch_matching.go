package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BatchMatchingStatus バッチマッチングステータス
type BatchMatchingStatus string

const (
	BatchMatchingStatusRunning   BatchMatchingStatus = "running"
	BatchMatchingStatusCompleted BatchMatchingStatus = "completed"
	BatchMatchingStatusFailed    BatchMatchingStatus = "failed"
)

// BatchType バッチ種別
type BatchType string

const (
	BatchTypeNtoN             BatchType = "n_to_n"
	BatchTypeJobToEngineers   BatchType = "job_to_engineers"
	BatchTypeEngineerToJobs   BatchType = "engineer_to_jobs"
)

// BatchMatching N:Nバッチ実行状態（Phase 2）
type BatchMatching struct {
	ID             string              `gorm:"type:uuid;primaryKey" json:"id"`
	UserID         string              `gorm:"type:uuid;not null;index" json:"user_id"`
	BatchType      BatchType           `gorm:"size:20;not null;default:'n_to_n'" json:"batch_type"`
	StartMonthFrom string              `gorm:"size:7;not null" json:"start_month_from"`
	StartMonthTo   string              `gorm:"size:7;not null" json:"start_month_to"`
	TotalPairs     int                 `gorm:"not null;default:0" json:"total_pairs"`
	SuccessCount   int                 `gorm:"not null;default:0" json:"success_count"`
	FailureCount   int                 `gorm:"not null;default:0" json:"failure_count"`
	Status         BatchMatchingStatus `gorm:"size:20;not null;default:'running';index" json:"status"`
	Results        json.RawMessage     `gorm:"type:jsonb;default:'[]'" json:"results"`
	CreatedAt      time.Time           `gorm:"index:idx_batch_matchings_created_at" json:"created_at"`
	CompletedAt    *time.Time          `json:"completed_at,omitempty"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (BatchMatching) TableName() string {
	return "batch_matchings"
}

func (b *BatchMatching) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	return nil
}

// BatchMatchingResultItem バッチ結果の個別アイテム
type BatchMatchingResultItem struct {
	JobID        string `json:"job_id"`
	EngineerID   string `json:"engineer_id"`
	JobName      string `json:"job_name"`
	EngineerName string `json:"engineer_name"`
	TotalScore   int    `json:"total_score"`
	Grade        string `json:"grade"`
	GradeLabel   string `json:"grade_label"`
	MatchingID   string `json:"matching_id,omitempty"`
}
