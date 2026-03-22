package service

import (
	"context"
	"fmt"

	"github.com/duesk/ivy/internal/dto"
	"github.com/duesk/ivy/internal/repository"
	"go.uber.org/zap"
)

// JobService 案件サービスインターフェース
type JobService interface {
	List(ctx context.Context, req dto.JobListRequest) (*dto.JobListResponse, error)
}

type jobService struct {
	jobRepo repository.JobRepository
	logger  *zap.Logger
}

// NewJobService 案件サービスを作成
func NewJobService(jobRepo repository.JobRepository, logger *zap.Logger) JobService {
	return &jobService{jobRepo: jobRepo, logger: logger}
}

func (s *jobService) List(ctx context.Context, req dto.JobListRequest) (*dto.JobListResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	jobs, total, err := s.jobRepo.List(ctx, repository.ListJobParams{
		Page:       req.Page,
		PageSize:   req.PageSize,
		StartMonth: req.StartMonth,
		Status:     req.Status,
	})
	if err != nil {
		return nil, fmt.Errorf("案件一覧取得失敗: %w", err)
	}

	items := make([]dto.JobResponse, len(jobs))
	for i, j := range jobs {
		items[i] = dto.JobResponse{
			ID:         j.ID,
			RawText:    j.RawText,
			Parsed:     j.Parsed,
			StartMonth: j.StartMonth,
			Status:     string(j.Status),
			CreatedAt:  j.CreatedAt,
			ExpiresAt:  j.ExpiresAt,
		}
	}

	totalPages := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPages++
	}

	return &dto.JobListResponse{
		Items:      items,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}
