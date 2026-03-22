package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/duesk/ivy/internal/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ProcessedEmailRepository 処理済みメールリポジトリインターフェース
type ProcessedEmailRepository interface {
	Create(ctx context.Context, email *model.ProcessedEmail) error
	ExistsByContentHash(ctx context.Context, hash string) (bool, error)
	DeleteExpired(ctx context.Context, cutoffDate time.Time) (int64, error)
}

type processedEmailRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewProcessedEmailRepository 処理済みメールリポジトリを作成
func NewProcessedEmailRepository(db *gorm.DB, logger *zap.Logger) ProcessedEmailRepository {
	return &processedEmailRepository{db: db, logger: logger}
}

func (r *processedEmailRepository) Create(ctx context.Context, email *model.ProcessedEmail) error {
	if err := r.db.WithContext(ctx).Create(email).Error; err != nil {
		return fmt.Errorf("処理済みメール作成失敗: %w", err)
	}
	return nil
}

func (r *processedEmailRepository) ExistsByContentHash(ctx context.Context, hash string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.ProcessedEmail{}).Where("content_hash = ?", hash).Count(&count).Error; err != nil {
		return false, fmt.Errorf("メールハッシュ確認失敗: %w", err)
	}
	return count > 0, nil
}

func (r *processedEmailRepository) DeleteExpired(ctx context.Context, cutoffDate time.Time) (int64, error) {
	result := r.db.WithContext(ctx).Where("processed_at < ?", cutoffDate).Delete(&model.ProcessedEmail{})
	if result.Error != nil {
		return 0, fmt.Errorf("期限切れメール削除失敗: %w", result.Error)
	}
	return result.RowsAffected, nil
}
