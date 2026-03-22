package service

import (
	"context"
	"fmt"

	"github.com/duesk/ivy/internal/dto"
	"github.com/duesk/ivy/internal/repository"
	"go.uber.org/zap"
)

// EngineerProfileService 人材プロファイルサービスインターフェース
type EngineerProfileService interface {
	List(ctx context.Context, req dto.EngineerProfileListRequest) (*dto.EngineerProfileListResponse, error)
}

type engineerProfileService struct {
	engineerProfileRepo repository.EngineerProfileRepository
	logger              *zap.Logger
}

// NewEngineerProfileService 人材プロファイルサービスを作成
func NewEngineerProfileService(engineerProfileRepo repository.EngineerProfileRepository, logger *zap.Logger) EngineerProfileService {
	return &engineerProfileService{engineerProfileRepo: engineerProfileRepo, logger: logger}
}

func (s *engineerProfileService) List(ctx context.Context, req dto.EngineerProfileListRequest) (*dto.EngineerProfileListResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	profiles, total, err := s.engineerProfileRepo.List(ctx, repository.ListEngineerProfileParams{
		Page:       req.Page,
		PageSize:   req.PageSize,
		StartMonth: req.StartMonth,
		Status:     req.Status,
	})
	if err != nil {
		return nil, fmt.Errorf("人材一覧取得失敗: %w", err)
	}

	items := make([]dto.EngineerProfileResponse, len(profiles))
	for i, p := range profiles {
		items[i] = dto.EngineerProfileResponse{
			ID:         p.ID,
			RawText:    p.RawText,
			Parsed:     p.Parsed,
			FileKey:    p.FileKey,
			StartMonth: p.StartMonth,
			Status:     string(p.Status),
			CreatedAt:  p.CreatedAt,
			ExpiresAt:  p.ExpiresAt,
		}
	}

	totalPages := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPages++
	}

	return &dto.EngineerProfileListResponse{
		Items:      items,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}
