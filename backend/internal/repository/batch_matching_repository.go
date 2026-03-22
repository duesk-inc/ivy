package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/duesk/ivy/internal/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// BatchMatchingRepository バッチマッチングリポジトリインターフェース
type BatchMatchingRepository interface {
	Create(ctx context.Context, batch *model.BatchMatching) error
	GetByID(ctx context.Context, id string) (*model.BatchMatching, error)
	UpdateProgress(ctx context.Context, id string, successCount, failureCount int) error
	Complete(ctx context.Context, id string, results json.RawMessage) error
	Fail(ctx context.Context, id string) error
	RecoverStaleRunning(ctx context.Context) (int64, error)
	ListByUserID(ctx context.Context, userID string, page, pageSize int) ([]model.BatchMatching, int64, error)
}

type batchMatchingRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewBatchMatchingRepository バッチマッチングリポジトリを作成
func NewBatchMatchingRepository(db *gorm.DB, logger *zap.Logger) BatchMatchingRepository {
	return &batchMatchingRepository{db: db, logger: logger}
}

func (r *batchMatchingRepository) Create(ctx context.Context, batch *model.BatchMatching) error {
	if err := r.db.WithContext(ctx).Create(batch).Error; err != nil {
		return fmt.Errorf("バッチマッチング作成失敗: %w", err)
	}
	return nil
}

func (r *batchMatchingRepository) GetByID(ctx context.Context, id string) (*model.BatchMatching, error) {
	var batch model.BatchMatching
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&batch).Error; err != nil {
		return nil, fmt.Errorf("バッチマッチング取得失敗 (id=%s): %w", id, err)
	}
	return &batch, nil
}

func (r *batchMatchingRepository) UpdateProgress(ctx context.Context, id string, successCount, failureCount int) error {
	result := r.db.WithContext(ctx).
		Model(&model.BatchMatching{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"success_count": successCount,
			"failure_count": failureCount,
		})
	if result.Error != nil {
		return fmt.Errorf("バッチ進捗更新失敗: %w", result.Error)
	}
	return nil
}

func (r *batchMatchingRepository) Complete(ctx context.Context, id string, results json.RawMessage) error {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&model.BatchMatching{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":       model.BatchMatchingStatusCompleted,
			"results":      results,
			"completed_at": &now,
		})
	if result.Error != nil {
		return fmt.Errorf("バッチ完了更新失敗: %w", result.Error)
	}
	return nil
}

func (r *batchMatchingRepository) Fail(ctx context.Context, id string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&model.BatchMatching{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":       model.BatchMatchingStatusFailed,
			"completed_at": &now,
		})
	if result.Error != nil {
		return fmt.Errorf("バッチ失敗更新失敗: %w", result.Error)
	}
	return nil
}

// RecoverStaleRunning 起動時リカバリ: running状態のバッチをfailedに更新
func (r *batchMatchingRepository) RecoverStaleRunning(ctx context.Context) (int64, error) {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&model.BatchMatching{}).
		Where("status = ?", model.BatchMatchingStatusRunning).
		Updates(map[string]any{
			"status":       model.BatchMatchingStatusFailed,
			"completed_at": &now,
		})
	if result.Error != nil {
		return 0, fmt.Errorf("staleバッチリカバリ失敗: %w", result.Error)
	}
	return result.RowsAffected, nil
}

func (r *batchMatchingRepository) ListByUserID(ctx context.Context, userID string, page, pageSize int) ([]model.BatchMatching, int64, error) {
	var batches []model.BatchMatching
	var total int64

	query := r.db.WithContext(ctx).Model(&model.BatchMatching{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("バッチ件数取得失敗: %w", err)
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&batches).Error; err != nil {
		return nil, 0, fmt.Errorf("バッチ一覧取得失敗: %w", err)
	}

	return batches, total, nil
}
