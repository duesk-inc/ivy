package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/duesk/ivy/internal/model"
	"github.com/duesk/ivy/internal/repository"
	"go.uber.org/zap"
)

// RetentionResult クリーンアップ結果
type RetentionResult struct {
	JobsDeleted           int64     `json:"jobs_deleted"`
	EngineersDeleted      int64     `json:"engineers_deleted"`
	ProcessedEmailsDeleted int64    `json:"processed_emails_deleted"`
	MatchingsDeleted      int64     `json:"matchings_deleted"`
	ExecutedAt            time.Time `json:"executed_at"`
}

// RetentionService データ保持期間管理サービスインターフェース
type RetentionService interface {
	RunCleanup(ctx context.Context) (*RetentionResult, error)
	ShouldRun(ctx context.Context) (bool, error)
	RecordLastCleanup(ctx context.Context) error
}

type retentionService struct {
	jobRepo            repository.JobRepository
	engineerProfileRepo repository.EngineerProfileRepository
	processedEmailRepo repository.ProcessedEmailRepository
	settingsRepo       repository.SettingsRepository
	logger             *zap.Logger
}

// NewRetentionService データ保持期間管理サービスを作成
func NewRetentionService(
	jobRepo repository.JobRepository,
	engineerProfileRepo repository.EngineerProfileRepository,
	processedEmailRepo repository.ProcessedEmailRepository,
	settingsRepo repository.SettingsRepository,
	logger *zap.Logger,
) RetentionService {
	return &retentionService{
		jobRepo:             jobRepo,
		engineerProfileRepo: engineerProfileRepo,
		processedEmailRepo:  processedEmailRepo,
		settingsRepo:        settingsRepo,
		logger:              logger,
	}
}

func (s *retentionService) RunCleanup(ctx context.Context) (*RetentionResult, error) {
	retention, err := s.getRetentionSettings(ctx)
	if err != nil {
		return nil, fmt.Errorf("保持期間設定取得失敗: %w", err)
	}

	result := &RetentionResult{
		ExecutedAt: time.Now(),
	}

	// 案件削除
	jobsDeleted, err := s.jobRepo.DeleteExpired(ctx)
	if err != nil {
		s.logger.Error("期限切れ案件削除失敗", zap.Error(err))
	} else {
		result.JobsDeleted = jobsDeleted
	}

	// 人材削除
	engineersDeleted, err := s.engineerProfileRepo.DeleteExpired(ctx)
	if err != nil {
		s.logger.Error("期限切れ人材削除失敗", zap.Error(err))
	} else {
		result.EngineersDeleted = engineersDeleted
	}

	// 処理済みメール削除
	emailCutoff := time.Now().AddDate(0, 0, -retention.GetProcessedEmailsDaysOrDefault())
	emailsDeleted, err := s.processedEmailRepo.DeleteExpired(ctx, emailCutoff)
	if err != nil {
		s.logger.Error("期限切れメール削除失敗", zap.Error(err))
	} else {
		result.ProcessedEmailsDeleted = emailsDeleted
	}

	s.logger.Info("データクリーンアップ完了",
		zap.Int64("jobs", result.JobsDeleted),
		zap.Int64("engineers", result.EngineersDeleted),
		zap.Int64("emails", result.ProcessedEmailsDeleted),
	)

	return result, nil
}

// ShouldRun 前回クリーンアップから7日以上経過しているか確認
func (s *retentionService) ShouldRun(ctx context.Context) (bool, error) {
	setting, err := s.settingsRepo.GetByKey(ctx, "last_cleanup_at")
	if err != nil {
		// 設定が存在しない場合は実行すべき
		return true, nil
	}

	var lastCleanup struct {
		At time.Time `json:"at"`
	}
	if err := json.Unmarshal(setting.Value, &lastCleanup); err != nil {
		return true, nil
	}

	return time.Since(lastCleanup.At) > 7*24*time.Hour, nil
}

// RecordLastCleanup クリーンアップ実行日時を記録
func (s *retentionService) RecordLastCleanup(ctx context.Context) error {
	value, _ := json.Marshal(map[string]any{
		"at": time.Now(),
	})
	return s.settingsRepo.Update(ctx, "last_cleanup_at", value, "")
}

func (s *retentionService) getRetentionSettings(ctx context.Context) (*model.DataRetentionSetting, error) {
	setting, err := s.settingsRepo.GetByKey(ctx, model.SettingKeyDataRetention)
	if err != nil {
		// デフォルト値を返す
		return &model.DataRetentionSetting{
			JobsDays:            90,
			EngineersDays:       180,
			MatchingsDays:       365,
			ProcessedEmailsDays: 180,
		}, nil
	}
	return model.ParseDataRetentionSetting(setting.Value)
}
