package repository

import (
	"context"
	"fmt"

	"github.com/duesk/ivy/internal/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// JobGroupRepository 案件グループリポジトリインターフェース
type JobGroupRepository interface {
	Create(ctx context.Context, jobGroup *model.JobGroup) error
	GetByID(ctx context.Context, id string) (*model.JobGroup, error)
	List(ctx context.Context, userID string) ([]model.JobGroup, error)
	Delete(ctx context.Context, id string) error
}

type jobGroupRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewJobGroupRepository 案件グループリポジトリを作成
func NewJobGroupRepository(db *gorm.DB, logger *zap.Logger) JobGroupRepository {
	return &jobGroupRepository{db: db, logger: logger}
}

func (r *jobGroupRepository) Create(ctx context.Context, jobGroup *model.JobGroup) error {
	if err := r.db.WithContext(ctx).Create(jobGroup).Error; err != nil {
		return fmt.Errorf("案件グループ作成失敗: %w", err)
	}
	return nil
}

func (r *jobGroupRepository) GetByID(ctx context.Context, id string) (*model.JobGroup, error) {
	var jobGroup model.JobGroup
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&jobGroup).Error; err != nil {
		return nil, fmt.Errorf("案件グループ取得失敗 (id=%s): %w", id, err)
	}
	return &jobGroup, nil
}

func (r *jobGroupRepository) List(ctx context.Context, userID string) ([]model.JobGroup, error) {
	var jobGroups []model.JobGroup
	query := r.db.WithContext(ctx).Order("created_at DESC")
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if err := query.Find(&jobGroups).Error; err != nil {
		return nil, fmt.Errorf("案件グループ一覧取得失敗: %w", err)
	}
	return jobGroups, nil
}

func (r *jobGroupRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.JobGroup{})
	if result.Error != nil {
		return fmt.Errorf("案件グループ削除失敗: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
