package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GmailSyncState Gmail同期位置管理（Phase 2）
type GmailSyncState struct {
	ID            string    `gorm:"type:uuid;primaryKey" json:"id"`
	LastHistoryID int64     `gorm:"not null;default:0" json:"last_history_id"`
	LastSyncedAt  time.Time `gorm:"not null;default:NOW()" json:"last_synced_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (GmailSyncState) TableName() string {
	return "gmail_sync_state"
}

func (g *GmailSyncState) BeforeCreate(tx *gorm.DB) error {
	if g.ID == "" {
		g.ID = uuid.New().String()
	}
	return nil
}
