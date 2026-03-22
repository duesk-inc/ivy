package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EmailClassification メール分類
type EmailClassification string

const (
	EmailClassificationJob      EmailClassification = "job"
	EmailClassificationEngineer EmailClassification = "engineer"
	EmailClassificationOther    EmailClassification = "other"
)

// ProcessedEmail 処理済みメール追跡（Phase 2）
type ProcessedEmail struct {
	ID             string              `gorm:"type:uuid;primaryKey" json:"id"`
	ContentHash    string              `gorm:"size:64;not null;uniqueIndex:uq_processed_emails_content_hash" json:"content_hash"`
	GmailMessageID string             `gorm:"size:255;not null;index" json:"gmail_message_id"`
	Classification EmailClassification `gorm:"size:20;not null" json:"classification"`
	ProcessedAt    time.Time           `gorm:"not null;default:NOW()" json:"processed_at"`
}

func (ProcessedEmail) TableName() string {
	return "processed_emails"
}

func (p *ProcessedEmail) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}
