package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/duesk/ivy/internal/dto"
	"github.com/duesk/ivy/internal/model"
	"github.com/duesk/ivy/internal/repository"
	"go.uber.org/zap"
)

// BatchMatchingService バッチマッチングサービスインターフェース
type BatchMatchingService interface {
	Preview(ctx context.Context, req dto.BatchMatchingPreviewRequest) (*dto.BatchMatchingPreviewResponse, error)
	Execute(ctx context.Context, userID string, req dto.BatchMatchingExecuteRequest) (*dto.BatchMatchingResponse, error)
	GetStatus(ctx context.Context, id string) (*dto.BatchMatchingResponse, error)
	MatchJobToEngineers(ctx context.Context, userID string, jobID string, req dto.OneToNMatchRequest) (*dto.BatchMatchingResponse, error)
	MatchEngineerToJobs(ctx context.Context, userID string, engineerID string, req dto.OneToNMatchRequest) (*dto.BatchMatchingResponse, error)
}

type batchMatchingService struct {
	jobRepo         repository.JobRepository
	engineerRepo    repository.EngineerProfileRepository
	batchRepo       repository.BatchMatchingRepository
	matchingService MatchingService
	prefilter       PrefilterService
	logger          *zap.Logger
}

// NewBatchMatchingService バッチマッチングサービスを作成
func NewBatchMatchingService(
	jobRepo repository.JobRepository,
	engineerRepo repository.EngineerProfileRepository,
	batchRepo repository.BatchMatchingRepository,
	matchingService MatchingService,
	prefilter PrefilterService,
	logger *zap.Logger,
) BatchMatchingService {
	return &batchMatchingService{
		jobRepo:         jobRepo,
		engineerRepo:    engineerRepo,
		batchRepo:       batchRepo,
		matchingService: matchingService,
		prefilter:       prefilter,
		logger:          logger,
	}
}

func (s *batchMatchingService) Preview(ctx context.Context, req dto.BatchMatchingPreviewRequest) (*dto.BatchMatchingPreviewResponse, error) {
	jobs, err := s.jobRepo.ListActive(ctx, req.StartMonthFrom, req.StartMonthTo)
	if err != nil {
		return nil, fmt.Errorf("案件取得失敗: %w", err)
	}

	engineers, err := s.engineerRepo.ListActive(ctx, req.StartMonthFrom, req.StartMonthTo)
	if err != nil {
		return nil, fmt.Errorf("人材取得失敗: %w", err)
	}

	pairs := s.prefilter.FilterPairs(jobs, engineers)

	return &dto.BatchMatchingPreviewResponse{
		TotalJobs:        len(jobs),
		TotalEngineers:   len(engineers),
		PairsAfterFilter: len(pairs),
		EstimatedCost:    s.prefilter.EstimateCost(len(pairs)),
	}, nil
}

func (s *batchMatchingService) Execute(ctx context.Context, userID string, req dto.BatchMatchingExecuteRequest) (*dto.BatchMatchingResponse, error) {
	jobs, err := s.jobRepo.ListActive(ctx, req.StartMonthFrom, req.StartMonthTo)
	if err != nil {
		return nil, fmt.Errorf("案件取得失敗: %w", err)
	}

	engineers, err := s.engineerRepo.ListActive(ctx, req.StartMonthFrom, req.StartMonthTo)
	if err != nil {
		return nil, fmt.Errorf("人材取得失敗: %w", err)
	}

	pairs := s.prefilter.FilterPairs(jobs, engineers)

	batch := &model.BatchMatching{
		UserID:         userID,
		BatchType:      model.BatchTypeNtoN,
		StartMonthFrom: req.StartMonthFrom,
		StartMonthTo:   req.StartMonthTo,
		TotalPairs:     len(pairs),
		Status:         model.BatchMatchingStatusRunning,
		Results:        json.RawMessage("[]"),
	}

	if err := s.batchRepo.Create(ctx, batch); err != nil {
		return nil, fmt.Errorf("バッチ作成失敗: %w", err)
	}

	// 非同期実行
	go s.executeBatch(batch.ID, userID, pairs)

	return s.batchToResponse(batch), nil
}

func (s *batchMatchingService) GetStatus(ctx context.Context, id string) (*dto.BatchMatchingResponse, error) {
	batch, err := s.batchRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.batchToResponse(batch), nil
}

// MatchJobToEngineers 指定案件に対して蓄積人材からマッチ候補を探す
func (s *batchMatchingService) MatchJobToEngineers(ctx context.Context, userID string, jobID string, req dto.OneToNMatchRequest) (*dto.BatchMatchingResponse, error) {
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("案件取得失敗: %w", err)
	}

	engineers, err := s.engineerRepo.ListActive(ctx, req.StartMonthFrom, req.StartMonthTo)
	if err != nil {
		return nil, fmt.Errorf("人材取得失敗: %w", err)
	}

	// プレフィルタ: 1件の案件 × 全人材
	jobs := []model.Job{*job}
	pairs := s.prefilter.FilterPairs(jobs, engineers)

	startMonth := job.StartMonth
	if startMonth == "" {
		startMonth = "all"
	}

	batch := &model.BatchMatching{
		UserID:         userID,
		BatchType:      model.BatchTypeJobToEngineers,
		StartMonthFrom: startMonth,
		StartMonthTo:   startMonth,
		TotalPairs:     len(pairs),
		Status:         model.BatchMatchingStatusRunning,
		Results:        json.RawMessage("[]"),
	}

	if err := s.batchRepo.Create(ctx, batch); err != nil {
		return nil, fmt.Errorf("バッチ作成失敗: %w", err)
	}

	go s.executeBatch(batch.ID, userID, pairs)

	return s.batchToResponse(batch), nil
}

// MatchEngineerToJobs 指定人材に対して蓄積案件からマッチ候補を探す
func (s *batchMatchingService) MatchEngineerToJobs(ctx context.Context, userID string, engineerID string, req dto.OneToNMatchRequest) (*dto.BatchMatchingResponse, error) {
	engineer, err := s.engineerRepo.GetByID(ctx, engineerID)
	if err != nil {
		return nil, fmt.Errorf("人材取得失敗: %w", err)
	}

	jobs, err := s.jobRepo.ListActive(ctx, req.StartMonthFrom, req.StartMonthTo)
	if err != nil {
		return nil, fmt.Errorf("案件取得失敗: %w", err)
	}

	// プレフィルタ: 全案件 × 1件の人材
	engineers := []model.EngineerProfile{*engineer}
	pairs := s.prefilter.FilterPairs(jobs, engineers)

	startMonth := engineer.StartMonth
	if startMonth == "" {
		startMonth = "all"
	}

	batch := &model.BatchMatching{
		UserID:         userID,
		BatchType:      model.BatchTypeEngineerToJobs,
		StartMonthFrom: startMonth,
		StartMonthTo:   startMonth,
		TotalPairs:     len(pairs),
		Status:         model.BatchMatchingStatusRunning,
		Results:        json.RawMessage("[]"),
	}

	if err := s.batchRepo.Create(ctx, batch); err != nil {
		return nil, fmt.Errorf("バッチ作成失敗: %w", err)
	}

	go s.executeBatch(batch.ID, userID, pairs)

	return s.batchToResponse(batch), nil
}

func (s *batchMatchingService) executeBatch(batchID, userID string, pairs []MatchPair) {
	ctx := context.Background()
	sem := make(chan struct{}, 5) // 並列5

	var successCount int64
	var failureCount int64
	var mu sync.Mutex
	var results []model.BatchMatchingResultItem

	var wg sync.WaitGroup

	for _, pair := range pairs {
		wg.Add(1)
		sem <- struct{}{} // セマフォ取得

		go func(p MatchPair) {
			defer wg.Done()
			defer func() { <-sem }() // セマフォ解放

			matchReq := dto.MatchingRequest{
				JobText:      p.Job.RawText,
				EngineerText: p.EngineerProfile.RawText,
			}

			resp, err := s.matchingService.Execute(ctx, userID, matchReq)
			if err != nil {
				atomic.AddInt64(&failureCount, 1)
				s.logger.Warn("バッチマッチング個別失敗",
					zap.String("job_id", p.Job.ID),
					zap.String("engineer_id", p.EngineerProfile.ID),
					zap.Error(err),
				)
				return
			}

			atomic.AddInt64(&successCount, 1)

			jobName := ""
			engName := ""
			if p.JobParsed != nil {
				jobName = p.JobParsed.Name
			}
			if p.EngineerParsed != nil {
				engName = p.EngineerParsed.Initials
			}

			item := model.BatchMatchingResultItem{
				JobID:        p.Job.ID,
				EngineerID:   p.EngineerProfile.ID,
				JobName:      jobName,
				EngineerName: engName,
				TotalScore:   resp.TotalScore,
				Grade:        resp.Grade,
				GradeLabel:   resp.GradeLabel,
				MatchingID:   resp.ID,
			}

			mu.Lock()
			results = append(results, item)
			mu.Unlock()

			// 定期的に進捗更新
			sc := atomic.LoadInt64(&successCount)
			fc := atomic.LoadInt64(&failureCount)
			if (sc+fc)%10 == 0 {
				_ = s.batchRepo.UpdateProgress(ctx, batchID, int(sc), int(fc))
			}
		}(pair)
	}

	wg.Wait()

	// スコア順ソート（降順）
	sortResults(results)

	resultsJSON, err := json.Marshal(results)
	if err != nil {
		s.logger.Error("バッチ結果JSON変換失敗", zap.Error(err))
		_ = s.batchRepo.Fail(ctx, batchID)
		return
	}

	if err := s.batchRepo.Complete(ctx, batchID, resultsJSON); err != nil {
		s.logger.Error("バッチ完了更新失敗", zap.Error(err))
	}

	s.logger.Info("バッチマッチング完了",
		zap.String("batch_id", batchID),
		zap.Int64("success", successCount),
		zap.Int64("failure", failureCount),
	)
}

func sortResults(results []model.BatchMatchingResultItem) {
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].TotalScore > results[i].TotalScore {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
}

func (s *batchMatchingService) batchToResponse(batch *model.BatchMatching) *dto.BatchMatchingResponse {
	return &dto.BatchMatchingResponse{
		ID:             batch.ID,
		BatchType:      string(batch.BatchType),
		StartMonthFrom: batch.StartMonthFrom,
		StartMonthTo:   batch.StartMonthTo,
		TotalPairs:     batch.TotalPairs,
		SuccessCount:   batch.SuccessCount,
		FailureCount:   batch.FailureCount,
		Status:         string(batch.Status),
		Results:        batch.Results,
		CreatedAt:      batch.CreatedAt,
		CompletedAt:    batch.CompletedAt,
	}
}

// RecoverStaleRunning 起動時リカバリ
func RecoverStaleRunning(ctx context.Context, batchRepo repository.BatchMatchingRepository, logger *zap.Logger) {
	recovered, err := batchRepo.RecoverStaleRunning(ctx)
	if err != nil {
		logger.Error("staleバッチリカバリ失敗", zap.Error(err))
		return
	}
	if recovered > 0 {
		logger.Info("staleバッチをfailedに更新", zap.Int64("count", recovered))
	}
}
