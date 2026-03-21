package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/duesk/ivy/internal/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// SettingsRepository 設定リポジトリインターフェース
type SettingsRepository interface {
	GetAll(ctx context.Context) ([]model.Setting, error)
	GetByKey(ctx context.Context, key string) (*model.Setting, error)
	Update(ctx context.Context, key string, value json.RawMessage, updatedBy string) error
}

type settingsRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewSettingsRepository 設定リポジトリを作成
func NewSettingsRepository(db *gorm.DB, logger *zap.Logger) SettingsRepository {
	return &settingsRepository{db: db, logger: logger}
}

func (r *settingsRepository) GetAll(ctx context.Context) ([]model.Setting, error) {
	var settings []model.Setting
	if err := r.db.WithContext(ctx).Find(&settings).Error; err != nil {
		return nil, fmt.Errorf("設定一覧取得失敗: %w", err)
	}
	return settings, nil
}

func (r *settingsRepository) GetByKey(ctx context.Context, key string) (*model.Setting, error) {
	var setting model.Setting
	if err := r.db.WithContext(ctx).Where("key = ?", key).First(&setting).Error; err != nil {
		return nil, fmt.Errorf("設定取得失敗 (key=%s): %w", key, err)
	}
	return &setting, nil
}

func (r *settingsRepository) Update(ctx context.Context, key string, value json.RawMessage, updatedBy string) error {
	result := r.db.WithContext(ctx).Model(&model.Setting{}).
		Where("key = ?", key).
		Updates(map[string]interface{}{
			"value":      value,
			"updated_by": updatedBy,
		})
	if result.Error != nil {
		return fmt.Errorf("設定更新失敗 (key=%s): %w", key, result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
