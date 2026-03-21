package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// JobGroup 案件グループ（同一案件の複数経路をグルーピング）
type JobGroup struct {
	ID        string    `gorm:"type:uuid;primaryKey" json:"id"`
	Name      string    `gorm:"size:500;not null" json:"name"`
	UserID    string    `gorm:"type:uuid;not null;index" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (JobGroup) TableName() string {
	return "job_groups"
}

func (g *JobGroup) BeforeCreate(tx *gorm.DB) error {
	if g.ID == "" {
		g.ID = uuid.New().String()
	}
	return nil
}
