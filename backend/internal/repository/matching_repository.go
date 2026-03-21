package repository

import (
	"context"
	"fmt"

	"github.com/duesk/ivy/internal/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// MatchingRepository マッチングリポジトリインターフェース
type MatchingRepository interface {
	Create(ctx context.Context, matching *model.Matching) error
	GetByID(ctx context.Context, id string) (*model.Matching, error)
	List(ctx context.Context, userID string, params ListMatchingParams) ([]model.Matching, int64, error)
	Delete(ctx context.Context, id string, userID string) error
	UpdateJobGroup(ctx context.Context, matchingID string, jobGroupID *string) error
	GetByJobGroupID(ctx context.Context, jobGroupID string) ([]model.Matching, error)
}

// ListMatchingParams 一覧検索パラメータ
type ListMatchingParams struct {
	Page     int
	PageSize int
	Grade    string
}

type matchingRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewMatchingRepository マッチングリポジトリを作成
func NewMatchingRepository(db *gorm.DB, logger *zap.Logger) MatchingRepository {
	return &matchingRepository{db: db, logger: logger}
}

func (r *matchingRepository) Create(ctx context.Context, matching *model.Matching) error {
	if err := r.db.WithContext(ctx).Create(matching).Error; err != nil {
		return fmt.Errorf("マッチング結果作成失敗: %w", err)
	}
	return nil
}

func (r *matchingRepository) GetByID(ctx context.Context, id string) (*model.Matching, error) {
	var matching model.Matching
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&matching).Error; err != nil {
		return nil, fmt.Errorf("マッチング結果取得失敗 (id=%s): %w", id, err)
	}
	return &matching, nil
}

func (r *matchingRepository) List(ctx context.Context, userID string, params ListMatchingParams) ([]model.Matching, int64, error) {
	var matchings []model.Matching
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Matching{})

	// salesロールの場合は自分のデータのみ（adminは全件閲覧可能）
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	if params.Grade != "" {
		query = query.Where("grade = ?", params.Grade)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("マッチング件数取得失敗: %w", err)
	}

	offset := (params.Page - 1) * params.PageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(params.PageSize).Find(&matchings).Error; err != nil {
		return nil, 0, fmt.Errorf("マッチング一覧取得失敗: %w", err)
	}

	return matchings, total, nil
}

func (r *matchingRepository) Delete(ctx context.Context, id string, userID string) error {
	result := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&model.Matching{})
	if result.Error != nil {
		return fmt.Errorf("マッチング結果削除失敗: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *matchingRepository) UpdateJobGroup(ctx context.Context, matchingID string, jobGroupID *string) error {
	result := r.db.WithContext(ctx).
		Model(&model.Matching{}).
		Where("id = ?", matchingID).
		Update("job_group_id", jobGroupID)
	if result.Error != nil {
		return fmt.Errorf("案件グループ紐付け更新失敗: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *matchingRepository) GetByJobGroupID(ctx context.Context, jobGroupID string) ([]model.Matching, error) {
	var matchings []model.Matching
	if err := r.db.WithContext(ctx).
		Where("job_group_id = ?", jobGroupID).
		Order("supply_chain_level ASC, created_at DESC").
		Find(&matchings).Error; err != nil {
		return nil, fmt.Errorf("案件グループのマッチング一覧取得失敗: %w", err)
	}
	return matchings, nil
}
