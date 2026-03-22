package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/duesk/ivy/internal/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// EngineerProfileRepository 人材プロファイルリポジトリインターフェース
type EngineerProfileRepository interface {
	Create(ctx context.Context, profile *model.EngineerProfile) error
	GetByID(ctx context.Context, id string) (*model.EngineerProfile, error)
	List(ctx context.Context, params ListEngineerProfileParams) ([]model.EngineerProfile, int64, error)
	ExistsByContentHash(ctx context.Context, hash string) (bool, error)
	DeleteExpired(ctx context.Context) (int64, error)
	UpdateStatus(ctx context.Context, id string, status model.EngineerProfileStatus) error
	ListActive(ctx context.Context, startMonthFrom, startMonthTo string) ([]model.EngineerProfile, error)
}

// ListEngineerProfileParams 人材一覧検索パラメータ
type ListEngineerProfileParams struct {
	Page       int
	PageSize   int
	StartMonth string
	Status     string
}

type engineerProfileRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewEngineerProfileRepository 人材プロファイルリポジトリを作成
func NewEngineerProfileRepository(db *gorm.DB, logger *zap.Logger) EngineerProfileRepository {
	return &engineerProfileRepository{db: db, logger: logger}
}

func (r *engineerProfileRepository) Create(ctx context.Context, profile *model.EngineerProfile) error {
	if err := r.db.WithContext(ctx).Create(profile).Error; err != nil {
		return fmt.Errorf("人材プロファイル作成失敗: %w", err)
	}
	return nil
}

func (r *engineerProfileRepository) GetByID(ctx context.Context, id string) (*model.EngineerProfile, error) {
	var profile model.EngineerProfile
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&profile).Error; err != nil {
		return nil, fmt.Errorf("人材プロファイル取得失敗 (id=%s): %w", id, err)
	}
	return &profile, nil
}

func (r *engineerProfileRepository) List(ctx context.Context, params ListEngineerProfileParams) ([]model.EngineerProfile, int64, error) {
	var profiles []model.EngineerProfile
	var total int64

	query := r.db.WithContext(ctx).Model(&model.EngineerProfile{})

	if params.StartMonth != "" {
		query = query.Where("start_month = ?", params.StartMonth)
	}
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("人材件数取得失敗: %w", err)
	}

	offset := (params.Page - 1) * params.PageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(params.PageSize).Find(&profiles).Error; err != nil {
		return nil, 0, fmt.Errorf("人材一覧取得失敗: %w", err)
	}

	return profiles, total, nil
}

func (r *engineerProfileRepository) ExistsByContentHash(ctx context.Context, hash string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.EngineerProfile{}).Where("content_hash = ?", hash).Count(&count).Error; err != nil {
		return false, fmt.Errorf("人材ハッシュ確認失敗: %w", err)
	}
	return count > 0, nil
}

func (r *engineerProfileRepository) DeleteExpired(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).Where("expires_at IS NOT NULL AND expires_at < ?", time.Now()).Delete(&model.EngineerProfile{})
	if result.Error != nil {
		return 0, fmt.Errorf("期限切れ人材削除失敗: %w", result.Error)
	}
	return result.RowsAffected, nil
}

func (r *engineerProfileRepository) UpdateStatus(ctx context.Context, id string, status model.EngineerProfileStatus) error {
	result := r.db.WithContext(ctx).Model(&model.EngineerProfile{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		return fmt.Errorf("人材ステータス更新失敗: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *engineerProfileRepository) ListActive(ctx context.Context, startMonthFrom, startMonthTo string) ([]model.EngineerProfile, error) {
	var profiles []model.EngineerProfile
	query := r.db.WithContext(ctx).Where("status = ?", model.EngineerProfileStatusActive)

	if startMonthFrom != "" {
		query = query.Where("start_month >= ?", startMonthFrom)
	}
	if startMonthTo != "" {
		query = query.Where("start_month <= ?", startMonthTo)
	}

	if err := query.Find(&profiles).Error; err != nil {
		return nil, fmt.Errorf("アクティブ人材取得失敗: %w", err)
	}
	return profiles, nil
}
