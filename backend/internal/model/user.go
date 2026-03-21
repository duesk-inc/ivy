package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Role ユーザーロール
type Role string

const (
	RoleAdmin Role = "admin"
	RoleSales Role = "sales"
)

// User ユーザーモデル（JITプロビジョニングで自動作成）
type User struct {
	ID         string    `gorm:"type:uuid;primaryKey" json:"id"`
	CognitoSub string   `gorm:"uniqueIndex;not null" json:"cognito_sub"`
	Email      string   `gorm:"size:255" json:"email"`
	Name       string   `gorm:"size:255;not null" json:"name"`
	Role       Role     `gorm:"size:50;not null;default:sales" json:"role"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// TableName テーブル名を指定
func (User) TableName() string {
	return "users"
}

// BeforeCreate UUID自動生成
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return nil
}

// IsAdmin 管理者かどうか
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}
