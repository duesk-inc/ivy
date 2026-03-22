package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/duesk/ivy/internal/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// JobRepository 案件リポジトリインターフェース
type JobRepository interface {
	Create(ctx context.Context, job *model.Job) error
	GetByID(ctx context.Context, id string) (*model.Job, error)
	List(ctx context.Context, params ListJobParams) ([]model.Job, int64, error)
	ExistsByContentHash(ctx context.Context, hash string) (bool, error)
	DeleteExpired(ctx context.Context) (int64, error)
	UpdateStatus(ctx context.Context, id string, status model.JobStatus) error
	ListActive(ctx context.Context, startMonthFrom, startMonthTo string) ([]model.Job, error)
}

// ListJobParams 案件一覧検索パラメータ
type ListJobParams struct {
	Page       int
	PageSize   int
	StartMonth string
	Status     string
}

type jobRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewJobRepository 案件リポジトリを作成
func NewJobRepository(db *gorm.DB, logger *zap.Logger) JobRepository {
	return &jobRepository{db: db, logger: logger}
}

func (r *jobRepository) Create(ctx context.Context, job *model.Job) error {
	if err := r.db.WithContext(ctx).Create(job).Error; err != nil {
		return fmt.Errorf("案件作成失敗: %w", err)
	}
	return nil
}

func (r *jobRepository) GetByID(ctx context.Context, id string) (*model.Job, error) {
	var job model.Job
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&job).Error; err != nil {
		return nil, fmt.Errorf("案件取得失敗 (id=%s): %w", id, err)
	}
	return &job, nil
}

func (r *jobRepository) List(ctx context.Context, params ListJobParams) ([]model.Job, int64, error) {
	var jobs []model.Job
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Job{})

	if params.StartMonth != "" {
		query = query.Where("start_month = ?", params.StartMonth)
	}
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("案件件数取得失敗: %w", err)
	}

	offset := (params.Page - 1) * params.PageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(params.PageSize).Find(&jobs).Error; err != nil {
		return nil, 0, fmt.Errorf("案件一覧取得失敗: %w", err)
	}

	return jobs, total, nil
}

func (r *jobRepository) ExistsByContentHash(ctx context.Context, hash string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Job{}).Where("content_hash = ?", hash).Count(&count).Error; err != nil {
		return false, fmt.Errorf("案件ハッシュ確認失敗: %w", err)
	}
	return count > 0, nil
}

func (r *jobRepository) DeleteExpired(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).Where("expires_at IS NOT NULL AND expires_at < ?", time.Now()).Delete(&model.Job{})
	if result.Error != nil {
		return 0, fmt.Errorf("期限切れ案件削除失敗: %w", result.Error)
	}
	return result.RowsAffected, nil
}

func (r *jobRepository) UpdateStatus(ctx context.Context, id string, status model.JobStatus) error {
	result := r.db.WithContext(ctx).Model(&model.Job{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		return fmt.Errorf("案件ステータス更新失敗: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *jobRepository) ListActive(ctx context.Context, startMonthFrom, startMonthTo string) ([]model.Job, error) {
	var jobs []model.Job
	query := r.db.WithContext(ctx).Where("status = ?", model.JobStatusActive)

	if startMonthFrom != "" {
		query = query.Where("start_month >= ?", startMonthFrom)
	}
	if startMonthTo != "" {
		query = query.Where("start_month <= ?", startMonthTo)
	}

	if err := query.Find(&jobs).Error; err != nil {
		return nil, fmt.Errorf("アクティブ案件取得失敗: %w", err)
	}
	return jobs, nil
}
