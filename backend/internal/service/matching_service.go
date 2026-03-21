package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/duesk/ivy/internal/dto"
	"github.com/duesk/ivy/internal/model"
	"github.com/duesk/ivy/internal/repository"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// MatchingService マッチングサービスインターフェース
type MatchingService interface {
	Execute(ctx context.Context, userID string, req dto.MatchingRequest) (*dto.MatchingResponse, error)
	GetByID(ctx context.Context, id string) (*dto.MatchingDetailResponse, error)
	List(ctx context.Context, userID string, role model.Role, req dto.MatchingListRequest) (*dto.MatchingListResponse, error)
	Delete(ctx context.Context, id string, userID string) error
	CreateJobGroup(ctx context.Context, userID string, req dto.CreateJobGroupRequest) (*dto.JobGroupResponse, error)
	GetJobGroup(ctx context.Context, id string) (*dto.JobGroupResponse, error)
	ListJobGroups(ctx context.Context, userID string, role model.Role) ([]dto.JobGroupResponse, error)
	DeleteJobGroup(ctx context.Context, id string, userID string) error
	LinkToJobGroup(ctx context.Context, matchingID string, jobGroupID string, userID string) error
	UnlinkFromJobGroup(ctx context.Context, matchingID string, userID string) error
}

type matchingService struct {
	matchingRepo  repository.MatchingRepository
	jobGroupRepo  repository.JobGroupRepository
	settingsRepo  repository.SettingsRepository
	aiService     AIService
	logger        *zap.Logger
}

// NewMatchingService マッチングサービスを作成
func NewMatchingService(
	matchingRepo repository.MatchingRepository,
	jobGroupRepo repository.JobGroupRepository,
	settingsRepo repository.SettingsRepository,
	aiService AIService,
	logger *zap.Logger,
) MatchingService {
	return &matchingService{
		matchingRepo:  matchingRepo,
		jobGroupRepo:  jobGroupRepo,
		settingsRepo:  settingsRepo,
		aiService:     aiService,
		logger:        logger,
	}
}

// Execute マッチング実行
func (s *matchingService) Execute(ctx context.Context, userID string, req dto.MatchingRequest) (*dto.MatchingResponse, error) {
	// 入力バリデーション
	if req.EngineerText == "" && req.EngineerFileKey == "" {
		return nil, fmt.Errorf("エンジニア情報（テキストまたはファイル）が必要です")
	}

	// engineer_file_keyとengineer_textの両方がある場合:
	// ファイルからの抽出テキストが主、engineer_textは補足として結合（設計書セクション6）
	// ※ファイルのテキスト抽出はhandler/fileで事前に行い、結果をengineer_textに格納する前提
	// ここでは両方あれば結合する
	engineerText := req.EngineerText

	// マージン設定を取得
	marginAmount := 50000
	marginType := "fixed"
	marginSetting, err := s.settingsRepo.GetByKey(ctx, model.SettingKeyMargin)
	if err == nil {
		ms, parseErr := model.ParseMarginSetting(marginSetting.Value)
		if parseErr == nil {
			marginAmount = ms.Amount
			marginType = ms.Type
		}
	}

	// AIモデル設定を取得
	modelName := "claude-haiku-4-5-20251001"
	aiModelSetting, err := s.settingsRepo.GetByKey(ctx, model.SettingKeyAIModel)
	if err == nil {
		ams, parseErr := model.ParseAIModelSetting(aiModelSetting.Value)
		if parseErr == nil {
			modelName = ams.Model
		}
	}

	// 補足情報をmapに変換
	supplement := make(map[string]interface{})
	if req.Supplement != nil {
		if req.Supplement.AffiliationType != "" {
			supplement["affiliation_type"] = req.Supplement.AffiliationType
		}
		if req.Supplement.AffiliationName != "" {
			supplement["affiliation_name"] = req.Supplement.AffiliationName
		}
		if req.Supplement.Rate > 0 {
			supplement["rate"] = req.Supplement.Rate
		}
		if req.Supplement.Nationality != "" {
			supplement["nationality"] = req.Supplement.Nationality
		}
		if req.Supplement.EmploymentType != "" {
			supplement["employment_type"] = req.Supplement.EmploymentType
		}
		if req.Supplement.AvailableFrom != "" {
			supplement["available_from"] = req.Supplement.AvailableFrom
		}
		if req.Supplement.SupplyChainLevel > 0 {
			supplement["supply_chain_level"] = req.Supplement.SupplyChainLevel
		}
		if req.Supplement.SupplyChainSource != "" {
			supplement["supply_chain_source"] = req.Supplement.SupplyChainSource
		}
	}

	// AIマッチング実行
	matchReq := MatchRequest{
		JobText:      req.JobText,
		EngineerText: engineerText,
		Supplement:   supplement,
		MarginAmount: marginAmount,
		MarginType:   marginType,
	}

	matchResp, err := s.aiService.Match(ctx, matchReq)
	if err != nil {
		s.logger.Error("AIマッチング失敗", zap.Error(err))
		return nil, fmt.Errorf("マッチング処理に失敗しました: %w", err)
	}

	// 補足情報をJSON化
	supplementJSON, _ := json.Marshal(supplement)

	// 商流レベルをマッピング
	var supplyChainLevel model.SupplyChainLevel
	var supplyChainSource string
	if req.Supplement != nil {
		supplyChainLevel = model.SupplyChainLevel(req.Supplement.SupplyChainLevel)
		supplyChainSource = req.Supplement.SupplyChainSource
	}

	// DBに保存
	matching := &model.Matching{
		UserID:            userID,
		JobText:           req.JobText,
		EngineerText:      engineerText,
		EngineerFileKey:   req.EngineerFileKey,
		Supplement:        supplementJSON,
		SupplyChainLevel:  supplyChainLevel,
		SupplyChainSource: supplyChainSource,
		TotalScore:        matchResp.TotalScore,
		Grade:             matchResp.Grade,
		Result:            matchResp.Result,
		ModelUsed:         modelName,
		TokensUsed:        matchResp.TokensUsed,
	}

	if err := s.matchingRepo.Create(ctx, matching); err != nil {
		s.logger.Error("マッチング結果保存失敗", zap.Error(err))
		return nil, fmt.Errorf("結果の保存に失敗しました: %w", err)
	}

	return &dto.MatchingResponse{
		ID:                matching.ID,
		TotalScore:        matching.TotalScore,
		Grade:             matching.Grade,
		GradeLabel:        matching.GradeLabel(),
		Result:            matching.Result,
		ModelUsed:         matching.ModelUsed,
		TokensUsed:        matching.TokensUsed,
		JobGroupID:        matching.JobGroupID,
		SupplyChainLevel:  int(matching.SupplyChainLevel),
		SupplyChainSource: matching.SupplyChainSource,
		CreatedAt:         matching.CreatedAt,
	}, nil
}

// GetByID マッチング詳細を取得
func (s *matchingService) GetByID(ctx context.Context, id string) (*dto.MatchingDetailResponse, error) {
	matching, err := s.matchingRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &dto.MatchingDetailResponse{
		ID:                matching.ID,
		JobText:           matching.JobText,
		EngineerText:      matching.EngineerText,
		EngineerFileKey:   matching.EngineerFileKey,
		Supplement:        matching.Supplement,
		TotalScore:        matching.TotalScore,
		Grade:             matching.Grade,
		GradeLabel:        matching.GradeLabel(),
		Result:            matching.Result,
		ModelUsed:         matching.ModelUsed,
		TokensUsed:        matching.TokensUsed,
		JobGroupID:        matching.JobGroupID,
		SupplyChainLevel:  int(matching.SupplyChainLevel),
		SupplyChainSource: matching.SupplyChainSource,
		CreatedAt:         matching.CreatedAt,
	}, nil
}

// List マッチング一覧を取得
func (s *matchingService) List(ctx context.Context, userID string, role model.Role, req dto.MatchingListRequest) (*dto.MatchingListResponse, error) {
	// adminは全件、salesは自分のみ
	queryUserID := userID
	if role == model.RoleAdmin {
		queryUserID = ""
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	matchings, total, err := s.matchingRepo.List(ctx, queryUserID, repository.ListMatchingParams{
		Page:     req.Page,
		PageSize: req.PageSize,
		Grade:    req.Grade,
	})
	if err != nil {
		return nil, err
	}

	items := make([]dto.MatchingListItem, len(matchings))
	for i, m := range matchings {
		// resultからjob_summaryを抽出
		var jobSummary json.RawMessage
		var result map[string]json.RawMessage
		if err := json.Unmarshal(m.Result, &result); err == nil {
			if js, ok := result["job_summary"]; ok {
				jobSummary = js
			}
		}

		items[i] = dto.MatchingListItem{
			ID:                m.ID,
			TotalScore:        m.TotalScore,
			Grade:             m.Grade,
			GradeLabel:        m.GradeLabel(),
			JobSummary:        jobSummary,
			ModelUsed:         m.ModelUsed,
			JobGroupID:        m.JobGroupID,
			SupplyChainLevel:  int(m.SupplyChainLevel),
			SupplyChainSource: m.SupplyChainSource,
			CreatedAt:         m.CreatedAt,
		}
	}

	totalPages := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPages++
	}

	return &dto.MatchingListResponse{
		Items:      items,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

// Delete マッチング結果を削除
func (s *matchingService) Delete(ctx context.Context, id string, userID string) error {
	return s.matchingRepo.Delete(ctx, id, userID)
}

// CreateJobGroup 案件グループを作成
func (s *matchingService) CreateJobGroup(ctx context.Context, userID string, req dto.CreateJobGroupRequest) (*dto.JobGroupResponse, error) {
	// マッチング結果の存在確認
	matching, err := s.matchingRepo.GetByID(ctx, req.MatchingID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("マッチング結果が見つかりません: %w", err)
		}
		return nil, fmt.Errorf("マッチング結果取得失敗: %w", err)
	}

	// 自分のマッチング結果かチェック
	if matching.UserID != userID {
		return nil, fmt.Errorf("このマッチング結果へのアクセス権限がありません")
	}

	jobGroup := &model.JobGroup{
		Name:   req.Name,
		UserID: userID,
	}

	if err := s.jobGroupRepo.Create(ctx, jobGroup); err != nil {
		s.logger.Error("案件グループ作成失敗", zap.Error(err))
		return nil, fmt.Errorf("案件グループの作成に失敗しました: %w", err)
	}

	// マッチングをグループに紐付け
	if err := s.matchingRepo.UpdateJobGroup(ctx, req.MatchingID, &jobGroup.ID); err != nil {
		s.logger.Error("マッチング紐付け失敗", zap.Error(err))
		return nil, fmt.Errorf("マッチングの紐付けに失敗しました: %w", err)
	}

	return s.buildJobGroupResponse(ctx, jobGroup)
}

// GetJobGroup 案件グループを取得
func (s *matchingService) GetJobGroup(ctx context.Context, id string) (*dto.JobGroupResponse, error) {
	jobGroup, err := s.jobGroupRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.buildJobGroupResponse(ctx, jobGroup)
}

// ListJobGroups 案件グループ一覧を取得
func (s *matchingService) ListJobGroups(ctx context.Context, userID string, role model.Role) ([]dto.JobGroupResponse, error) {
	queryUserID := userID
	if role == model.RoleAdmin {
		queryUserID = ""
	}

	jobGroups, err := s.jobGroupRepo.List(ctx, queryUserID)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.JobGroupResponse, 0, len(jobGroups))
	for _, jg := range jobGroups {
		resp, err := s.buildJobGroupResponse(ctx, &jg)
		if err != nil {
			s.logger.Error("案件グループレスポンス構築失敗", zap.Error(err), zap.String("job_group_id", jg.ID))
			continue
		}
		responses = append(responses, *resp)
	}

	return responses, nil
}

// DeleteJobGroup 案件グループを削除
func (s *matchingService) DeleteJobGroup(ctx context.Context, id string, userID string) error {
	jobGroup, err := s.jobGroupRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if jobGroup.UserID != userID {
		return fmt.Errorf("この案件グループへのアクセス権限がありません")
	}

	// グループに属するマッチングの紐付けを解除
	matchings, err := s.matchingRepo.GetByJobGroupID(ctx, id)
	if err != nil {
		return fmt.Errorf("グループ内マッチング取得失敗: %w", err)
	}
	for _, m := range matchings {
		if err := s.matchingRepo.UpdateJobGroup(ctx, m.ID, nil); err != nil {
			s.logger.Error("マッチング紐付け解除失敗", zap.Error(err), zap.String("matching_id", m.ID))
		}
	}

	return s.jobGroupRepo.Delete(ctx, id)
}

// LinkToJobGroup マッチングを案件グループに紐付け
func (s *matchingService) LinkToJobGroup(ctx context.Context, matchingID string, jobGroupID string, userID string) error {
	matching, err := s.matchingRepo.GetByID(ctx, matchingID)
	if err != nil {
		return err
	}

	if matching.UserID != userID {
		return fmt.Errorf("このマッチング結果へのアクセス権限がありません")
	}

	// グループの存在確認
	jobGroup, err := s.jobGroupRepo.GetByID(ctx, jobGroupID)
	if err != nil {
		return err
	}

	if jobGroup.UserID != userID {
		return fmt.Errorf("この案件グループへのアクセス権限がありません")
	}

	return s.matchingRepo.UpdateJobGroup(ctx, matchingID, &jobGroupID)
}

// UnlinkFromJobGroup マッチングの案件グループ紐付けを解除
func (s *matchingService) UnlinkFromJobGroup(ctx context.Context, matchingID string, userID string) error {
	matching, err := s.matchingRepo.GetByID(ctx, matchingID)
	if err != nil {
		return err
	}

	if matching.UserID != userID {
		return fmt.Errorf("このマッチング結果へのアクセス権限がありません")
	}

	return s.matchingRepo.UpdateJobGroup(ctx, matchingID, nil)
}

// buildJobGroupResponse 案件グループレスポンスを構築
func (s *matchingService) buildJobGroupResponse(ctx context.Context, jobGroup *model.JobGroup) (*dto.JobGroupResponse, error) {
	matchings, err := s.matchingRepo.GetByJobGroupID(ctx, jobGroup.ID)
	if err != nil {
		return nil, fmt.Errorf("グループ内マッチング取得失敗: %w", err)
	}

	items := make([]dto.MatchingListItem, len(matchings))
	for i, m := range matchings {
		var jobSummary json.RawMessage
		var result map[string]json.RawMessage
		if err := json.Unmarshal(m.Result, &result); err == nil {
			if js, ok := result["job_summary"]; ok {
				jobSummary = js
			}
		}

		items[i] = dto.MatchingListItem{
			ID:                m.ID,
			TotalScore:        m.TotalScore,
			Grade:             m.Grade,
			GradeLabel:        m.GradeLabel(),
			JobSummary:        jobSummary,
			ModelUsed:         m.ModelUsed,
			JobGroupID:        m.JobGroupID,
			SupplyChainLevel:  int(m.SupplyChainLevel),
			SupplyChainSource: m.SupplyChainSource,
			CreatedAt:         m.CreatedAt,
		}
	}

	// ベストルートを決定: supply_chain_level ASC (0は最後に), rate ASC
	var bestRoute *dto.MatchingListItem
	if len(items) > 0 {
		sorted := make([]dto.MatchingListItem, len(items))
		copy(sorted, items)
		sort.Slice(sorted, func(i, j int) bool {
			li := sorted[i].SupplyChainLevel
			lj := sorted[j].SupplyChainLevel
			// 0(不明)は最後に
			if li == 0 && lj != 0 {
				return false
			}
			if li != 0 && lj == 0 {
				return true
			}
			if li != lj {
				return li < lj
			}
			// 同じ商流レベルならrateで比較（低い方が良い）
			ri := extractRate(sorted[i].JobSummary)
			rj := extractRate(sorted[j].JobSummary)
			return ri < rj
		})
		bestRoute = &sorted[0]
	}

	return &dto.JobGroupResponse{
		ID:        jobGroup.ID,
		Name:      jobGroup.Name,
		Matchings: items,
		BestRoute: bestRoute,
		CreatedAt: jobGroup.CreatedAt,
	}, nil
}

// extractRate job_summaryからrateを抽出（数値として）
func extractRate(jobSummary json.RawMessage) int {
	if jobSummary == nil {
		return 0
	}
	var summary map[string]interface{}
	if err := json.Unmarshal(jobSummary, &summary); err != nil {
		return 0
	}
	rateVal, ok := summary["rate"]
	if !ok {
		return 0
	}
	switch v := rateVal.(type) {
	case float64:
		return int(v)
	case string:
		cleaned := strings.ReplaceAll(v, ",", "")
		cleaned = strings.ReplaceAll(cleaned, "万", "")
		cleaned = strings.ReplaceAll(cleaned, "円", "")
		cleaned = strings.TrimSpace(cleaned)
		n, err := strconv.Atoi(cleaned)
		if err != nil {
			return 0
		}
		return n
	default:
		return 0
	}
}
