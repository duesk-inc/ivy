package service

import (
	"context"
	"encoding/json"
	"strings"

	"go.uber.org/zap"
)

// MockAIService テスト/開発用のモックAIサービス（設計書セクション11: 3パターン対応）
type MockAIService struct {
	logger *zap.Logger
}

// NewMockAIService モックAIサービスを作成
func NewMockAIService(logger *zap.Logger) *MockAIService {
	return &MockAIService{logger: logger}
}

// Match モックマッチング実行（入力に応じて3パターンの固定JSONを返す）
func (s *MockAIService) Match(ctx context.Context, req MatchRequest) (*MatchResponse, error) {
	// 入力テキストからパターンを判定
	pattern := detectPattern(req.JobText, req.EngineerText)
	s.logger.Info("MockAIService: マッチング実行", zap.String("pattern", pattern))

	switch pattern {
	case "low":
		return buildLowMatchResponse()
	case "high":
		return buildHighMatchResponse()
	default:
		return buildMidMatchResponse()
	}
}

// detectPattern 入力テキストからマッチパターンを判定
func detectPattern(jobText, engineerText string) string {
	combined := strings.ToLower(jobText + engineerText)
	// インフラ系 × 開発系 → 低マッチ
	if (strings.Contains(combined, "インフラ") || strings.Contains(combined, "運用")) &&
		(strings.Contains(combined, "開発") || strings.Contains(combined, "cra")) {
		return "low"
	}
	// PM × PM → 高マッチ
	if strings.Contains(combined, "pm") || strings.Contains(combined, "プロジェクトマネージャ") {
		return "high"
	}
	return "mid"
}

// buildLowMatchResponse D判定（低マッチ）
func buildLowMatchResponse() (*MatchResponse, error) {
	result := map[string]any{
		"job_summary":      map[string]any{"name": "CRA準拠Webアプリ開発", "location": "東京都港区", "remote": "常駐", "rate": "55-65万円", "start": "即日", "conditions": "React, TypeScript, CRA準拠"},
		"engineer_summary": map[string]any{"initials": "K.S.", "age": 28, "gender": "男性", "nearest_station": "品川駅", "affiliation": "パートナー", "rate": "50万円", "available_from": "即日"},
		"total_score":      25,
		"grade":            "D",
		"grade_label":      "提案非推奨",
		"scores": map[string]any{
			"skill":            map[string]any{"score": 10, "max": 50, "reason": "インフラ運用経験が中心。React/TypeScript経験なし", "required_skills": []map[string]any{{"skill": "React", "status": "unmet", "detail": "経験なし"}, {"skill": "TypeScript", "status": "unmet", "detail": "経験なし"}}, "optional_skills": []map[string]any{}},
			"timing":           map[string]any{"score": 10, "max": 10, "reason": "即日稼働可能"},
			"rate":             map[string]any{"score": 5, "max": 10, "reason": "単価レンジは許容範囲内", "calculation": "案件55-65万 vs エンジニア50万+マージン5万=55万"},
			"experience_years": map[string]any{"score": 0, "max": 10, "reason": "開発経験なし"},
			"location":         map[string]any{"score": 7, "max": 10, "reason": "品川から港区は通勤圏内", "commute_time": "約15分"},
			"industry":         map[string]any{"score": 0, "max": 10, "reason": "Web開発業界の経験なし"},
		},
		"ng_flags": map[string]any{
			"nationality":  map[string]any{"status": "ok", "detail": "日本国籍"},
			"freelancer":   map[string]any{"status": "ok", "detail": "制限なし"},
			"supply_chain": map[string]any{"status": "ok", "detail": "制限なし"},
			"age":          map[string]any{"status": "ok", "detail": "制限なし"},
		},
		"negatives":          []string{"必須スキル（React, TypeScript）の経験が全くない", "開発経験がなくインフラ運用中心"},
		"positives":          []string{"稼働時期は合致", "通勤圏内"},
		"warnings":           []string{"スキルセットが根本的に異なります。提案は推奨しません"},
		"advice":             "スキルセットの不一致が大きいため、提案は見送りを推奨します。",
		"confirmation_hints": []map[string]any{},
	}
	resultJSON, _ := json.Marshal(result)
	return &MatchResponse{TotalScore: 25, Grade: "D", GradeLabel: "提案非推奨", Result: resultJSON, TokensUsed: 2800, ModelUsed: "mock-ai"}, nil
}

// buildMidMatchResponse B判定（中マッチ）
func buildMidMatchResponse() (*MatchResponse, error) {
	result := map[string]any{
		"job_summary":      map[string]any{"name": "Webアプリケーション開発", "location": "東京都渋谷区", "remote": "一部リモート", "rate": "65-75万円", "start": "2026年4月〜", "conditions": "Java, Spring Boot, 3年以上"},
		"engineer_summary": map[string]any{"initials": "T.Y.", "age": 30, "gender": "男性", "nearest_station": "池袋駅", "affiliation": "パートナー（株式会社SasaTech）", "rate": "60万円", "available_from": "2026年4月〜"},
		"total_score":      72,
		"grade":            "B",
		"grade_label":      "提案検討可",
		"scores": map[string]any{
			"skill":            map[string]any{"score": 35, "max": 50, "reason": "Java経験4年あり。Spring Boot経験は2年で必須要件の3年にやや不足", "required_skills": []map[string]any{{"skill": "Java", "status": "met", "detail": "4年の実務経験あり"}, {"skill": "Spring Boot", "status": "partial", "detail": "2年の経験。要件の3年にやや不足"}, {"skill": "PostgreSQL", "status": "met", "detail": "3年の経験あり"}}, "optional_skills": []map[string]any{{"skill": "Docker", "status": "met", "detail": "1年の経験あり"}}},
			"timing":           map[string]any{"score": 10, "max": 10, "reason": "案件開始2026年4月、エンジニア稼働可能2026年4月で一致"},
			"rate":             map[string]any{"score": 8, "max": 10, "reason": "案件65-75万、エンジニア60万＋マージン5万＝デュスク提示65万で下限ちょうど", "calculation": "案件65-75万 - マージン5万 = 上限70万、エンジニア60万 → OK"},
			"experience_years": map[string]any{"score": 7, "max": 10, "reason": "IT業界4年の経験。案件要件の3年以上を満たす"},
			"location":         map[string]any{"score": 7, "max": 10, "reason": "池袋駅から渋谷駅まで約20分。通勤圏内", "commute_time": "約20分"},
			"industry":         map[string]any{"score": 5, "max": 10, "reason": "Web系開発経験あり。業界特化の経験は見当たらない"},
		},
		"ng_flags": map[string]any{
			"nationality":  map[string]any{"status": "ok", "detail": "日本国籍"},
			"freelancer":   map[string]any{"status": "ok", "detail": "フリーランス可の案件"},
			"supply_chain": map[string]any{"status": "ok", "detail": "商流制限なし"},
			"age":          map[string]any{"status": "ok", "detail": "年齢制限なし"},
		},
		"negatives": []string{"Spring Boot経験が要件の3年に対して2年とやや不足"},
		"positives": []string{"Java経験4年で必須要件を十分に満たす", "稼働時期が案件開始と一致", "通勤圏内の立地"},
		"warnings":  []string{"Spring Boot経験について、案件側に2年でも可か確認をお勧めします"},
		"advice":    "Java経験の豊富さを訴求ポイントとして提案可能。Spring Boot経験の不足分は自己学習の意欲等でカバーできる可能性あり。",
		"confirmation_hints": []map[string]any{
			{"target": "案件側", "question": "Spring Boot経験について、2年の実務経験でもご検討いただけますでしょうか？Java経験は4年ございます。", "reason": "必須スキルの経験年数が部分的に不足しているため、柔軟性を確認する価値がある"},
			{"target": "パートナー（株式会社SasaTech）", "question": "T.Y.様のSpring Boot経験について、業務外での学習やプライベートプロジェクトでの経験はございますか？", "reason": "経歴書に記載がない追加経験がある可能性がある"},
		},
	}
	resultJSON, _ := json.Marshal(result)
	return &MatchResponse{TotalScore: 72, Grade: "B", GradeLabel: "提案検討可", Result: resultJSON, TokensUsed: 3200, ModelUsed: "mock-ai"}, nil
}

// buildHighMatchResponse A判定（高マッチ）
func buildHighMatchResponse() (*MatchResponse, error) {
	result := map[string]any{
		"job_summary":      map[string]any{"name": "ゲーム開発PMポジション", "location": "東京都新宿区", "remote": "フルリモート", "rate": "80-100万円", "start": "2026年5月〜", "conditions": "PM経験10年以上、ゲーム業界経験必須"},
		"engineer_summary": map[string]any{"initials": "Y.Y.", "age": 45, "gender": "男性", "nearest_station": "新宿駅", "affiliation": "デュスク社員", "rate": "-", "available_from": "2026年5月〜"},
		"total_score":      88,
		"grade":            "A",
		"grade_label":      "提案推奨",
		"scores": map[string]any{
			"skill":            map[string]any{"score": 45, "max": 50, "reason": "PM経験20年、ゲーム業界PM経験15年で必須要件を大幅に超過", "required_skills": []map[string]any{{"skill": "PM経験10年以上", "status": "met", "detail": "PM経験20年で大幅に超過"}, {"skill": "ゲーム業界経験", "status": "met", "detail": "ゲーム業界PM経験15年"}}, "optional_skills": []map[string]any{{"skill": "アジャイル", "status": "met", "detail": "スクラムマスター資格保有"}}},
			"timing":           map[string]any{"score": 10, "max": 10, "reason": "2026年5月開始で一致"},
			"rate":             map[string]any{"score": 10, "max": 10, "reason": "デュスク社員のため単価がそのまま売上", "calculation": "案件80-100万がそのまま売上"},
			"experience_years": map[string]any{"score": 10, "max": 10, "reason": "PM20年の豊富な経験"},
			"location":         map[string]any{"score": 10, "max": 10, "reason": "フルリモートのため立地は関係なし"},
			"industry":         map[string]any{"score": 3, "max": 10, "reason": "ゲーム業界PM経験15年で完全一致"},
		},
		"ng_flags": map[string]any{
			"nationality":  map[string]any{"status": "ok", "detail": "日本国籍"},
			"freelancer":   map[string]any{"status": "ok", "detail": "正社員"},
			"supply_chain": map[string]any{"status": "ok", "detail": "自社社員"},
			"age":          map[string]any{"status": "ok", "detail": "年齢制限なし"},
		},
		"negatives":          []string{},
		"positives":          []string{"PM経験20年で要件を大幅に超過", "ゲーム業界の深い知見", "フルリモートで勤務地の制約なし", "自社社員のため単価交渉不要"},
		"warnings":           []string{},
		"advice":             "スキル・経験ともに高いマッチ度。即座に提案を推奨します。ゲーム業界での長年のPM経験を訴求ポイントとして強調してください。",
		"confirmation_hints": []map[string]any{{"target": "案件側", "question": "Y.Y.は弊社社員でPM経験20年（ゲーム業界15年）のベテランです。直近のプロジェクト規模や体制について詳細をお伺いできますか？", "reason": "高マッチのためすぐに面談設定できる可能性が高い"}},
	}
	resultJSON, _ := json.Marshal(result)
	return &MatchResponse{TotalScore: 88, Grade: "A", GradeLabel: "提案推奨", Result: resultJSON, TokensUsed: 3500, ModelUsed: "mock-ai"}, nil
}
