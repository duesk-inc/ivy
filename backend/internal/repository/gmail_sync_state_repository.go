package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/duesk/ivy/internal/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// GmailSyncStateRepository Gmail同期状態リポジトリインターフェース
type GmailSyncStateRepository interface {
	GetOrCreate(ctx context.Context) (*model.GmailSyncState, error)
	UpdateLastSync(ctx context.Context, historyID int64, syncedAt time.Time) error
}

type gmailSyncStateRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewGmailSyncStateRepository Gmail同期状態リポジトリを作成
func NewGmailSyncStateRepository(db *gorm.DB, logger *zap.Logger) GmailSyncStateRepository {
	return &gmailSyncStateRepository{db: db, logger: logger}
}

func (r *gmailSyncStateRepository) GetOrCreate(ctx context.Context) (*model.GmailSyncState, error) {
	var state model.GmailSyncState
	err := r.db.WithContext(ctx).First(&state).Error
	if err == gorm.ErrRecordNotFound {
		state = model.GmailSyncState{
			LastHistoryID: 0,
			LastSyncedAt:  time.Time{},
		}
		if err := r.db.WithContext(ctx).Create(&state).Error; err != nil {
			return nil, fmt.Errorf("Gmail同期状態初期化失敗: %w", err)
		}
		return &state, nil
	}
	if err != nil {
		return nil, fmt.Errorf("Gmail同期状態取得失敗: %w", err)
	}
	return &state, nil
}

func (r *gmailSyncStateRepository) UpdateLastSync(ctx context.Context, historyID int64, syncedAt time.Time) error {
	result := r.db.WithContext(ctx).
		Model(&model.GmailSyncState{}).
		Where("1 = 1").
		Updates(map[string]any{
			"last_history_id": historyID,
			"last_synced_at":  syncedAt,
			"updated_at":      time.Now(),
		})
	if result.Error != nil {
		return fmt.Errorf("Gmail同期状態更新失敗: %w", result.Error)
	}
	return nil
}
