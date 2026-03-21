package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SupplyChainLevel 商流レベル
type SupplyChainLevel int

const (
	SupplyChainUnknown  SupplyChainLevel = 0 // 不明
	SupplyChainDirect   SupplyChainLevel = 1 // エンド直
	SupplyChainFirst    SupplyChainLevel = 2 // 1次請け（エンド→1社→デュスク）
	SupplyChainSecond   SupplyChainLevel = 3 // 2次請け
	SupplyChainThirdUp  SupplyChainLevel = 4 // 3次以上
)

// SupplyChainLabel 商流レベルのラベルを返す
func (l SupplyChainLevel) Label() string {
	switch l {
	case SupplyChainDirect:
		return "エンド直"
	case SupplyChainFirst:
		return "1次請け"
	case SupplyChainSecond:
		return "2次請け"
	case SupplyChainThirdUp:
		return "3次以上"
	default:
		return "不明"
	}
}

// Matching マッチング結果モデル
type Matching struct {
	ID                string           `gorm:"type:uuid;primaryKey" json:"id"`
	UserID            string           `gorm:"type:uuid;not null;index" json:"user_id"`
	JobGroupID        *string          `gorm:"type:uuid;index" json:"job_group_id,omitempty"`
	JobText           string           `gorm:"type:text;not null" json:"job_text"`
	EngineerText      string           `gorm:"type:text;not null" json:"engineer_text"`
	EngineerFileKey   string           `gorm:"size:500" json:"engineer_file_key,omitempty"`
	Supplement        json.RawMessage  `gorm:"type:jsonb;default:'{}'" json:"supplement"`
	SupplyChainLevel  SupplyChainLevel `gorm:"not null;default:0" json:"supply_chain_level"`
	SupplyChainSource string           `gorm:"size:255" json:"supply_chain_source,omitempty"`
	TotalScore        int              `gorm:"not null" json:"total_score"`
	Grade             string           `gorm:"size:1;not null" json:"grade"`
	Result            json.RawMessage  `gorm:"type:jsonb;not null" json:"result"`
	ModelUsed         string           `gorm:"size:100;not null" json:"model_used"`
	TokensUsed        int              `gorm:"not null;default:0" json:"tokens_used"`
	CreatedAt         time.Time        `gorm:"index:idx_matchings_created_at" json:"created_at"`

	User     User      `gorm:"foreignKey:UserID" json:"-"`
	JobGroup *JobGroup `gorm:"foreignKey:JobGroupID" json:"job_group,omitempty"`
}

// TableName テーブル名を指定
func (Matching) TableName() string {
	return "matchings"
}

// BeforeCreate UUID自動生成
func (m *Matching) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	return nil
}

// GradeLabel グレード判定ラベルを返す
func (m *Matching) GradeLabel() string {
	switch m.Grade {
	case "A":
		return "提案推奨"
	case "B":
		return "提案検討可"
	case "C":
		return "条件次第で検討"
	case "D":
		return "提案非推奨"
	default:
		return ""
	}
}

// CalculateGrade スコアからグレードを算出
func CalculateGrade(score int) string {
	switch {
	case score >= 80:
		return "A"
	case score >= 60:
		return "B"
	case score >= 40:
		return "C"
	default:
		return "D"
	}
}
