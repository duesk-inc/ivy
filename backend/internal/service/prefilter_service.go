package service

import (
	"math"
	"strings"

	"github.com/duesk/ivy/internal/model"
	"go.uber.org/zap"
)

// MatchPair プレフィルタ後のマッチングペア
type MatchPair struct {
	Job             *model.Job
	EngineerProfile *model.EngineerProfile
	JobParsed       *model.ParsedJobData
	EngineerParsed  *model.ParsedEngineerData
}

// CostPerPair Claude API呼び出し1回あたりの推定コスト（USD）
const CostPerPair = 0.02

// PrefilterService プレフィルタサービスインターフェース
type PrefilterService interface {
	FilterPairs(jobs []model.Job, engineers []model.EngineerProfile) []MatchPair
	EstimateCost(pairCount int) float64
}

type prefilterService struct {
	logger *zap.Logger
}

// NewPrefilterService プレフィルタサービスを作成
func NewPrefilterService(logger *zap.Logger) PrefilterService {
	return &prefilterService{logger: logger}
}

func (s *prefilterService) FilterPairs(jobs []model.Job, engineers []model.EngineerProfile) []MatchPair {
	var pairs []MatchPair

	for i := range jobs {
		jobParsed, err := jobs[i].GetParsedData()
		if err != nil {
			s.logger.Warn("案件パースデータ取得失敗", zap.String("job_id", jobs[i].ID), zap.Error(err))
			continue
		}

		for j := range engineers {
			engParsed, err := engineers[j].GetParsedData()
			if err != nil {
				s.logger.Warn("人材パースデータ取得失敗", zap.String("engineer_id", engineers[j].ID), zap.Error(err))
				continue
			}

			if s.shouldExclude(jobParsed, engParsed) {
				continue
			}

			pairs = append(pairs, MatchPair{
				Job:             &jobs[i],
				EngineerProfile: &engineers[j],
				JobParsed:       jobParsed,
				EngineerParsed:  engParsed,
			})
		}
	}

	s.logger.Info("プレフィルタ完了",
		zap.Int("total_combinations", len(jobs)*len(engineers)),
		zap.Int("after_filter", len(pairs)),
	)

	return pairs
}

func (s *prefilterService) EstimateCost(pairCount int) float64 {
	return float64(pairCount) * CostPerPair
}

// shouldExclude ルールベースの除外判定
func (s *prefilterService) shouldExclude(job *model.ParsedJobData, eng *model.ParsedEngineerData) bool {
	// 1. 稼働時期不一致
	if job.StartMonth != "" && eng.StartMonth != "" && job.StartMonth != eng.StartMonth {
		return true
	}

	// 2. 外国籍NG × 外国籍エンジニア
	if job.NationalityOK != nil && !*job.NationalityOK {
		if eng.Nationality != "" && !isJapanese(eng.Nationality) {
			return true
		}
	}

	// 3. フリーランスNG × フリーランス
	if job.FreelanceOK != nil && !*job.FreelanceOK {
		if isFreelance(eng.EmploymentType) {
			return true
		}
	}

	// 4. 単価乖離20万以上
	if eng.Rate != nil && *eng.Rate > 0 {
		if job.RateMax != nil && *job.RateMax > 0 {
			if math.Abs(float64(*eng.Rate-*job.RateMax)) > 20 {
				return true
			}
		}
	}

	// 5. 年齢制限超え
	if job.AgeLimit != nil && *job.AgeLimit > 0 && eng.Age != nil && *eng.Age > 0 {
		if *eng.Age > *job.AgeLimit {
			return true
		}
	}

	return false
}

func isJapanese(nationality string) bool {
	n := strings.ToLower(nationality)
	return n == "japanese" || n == "日本" || n == "日本人" || n == "jp"
}

func isFreelance(employmentType string) bool {
	e := strings.ToLower(employmentType)
	return e == "freelance" || e == "フリーランス" || e == "個人事業主"
}
