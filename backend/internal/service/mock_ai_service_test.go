package service

import (
	"context"
	"encoding/json"
	"testing"

	"go.uber.org/zap"
)

func TestMockAIService_Match(t *testing.T) {
	logger := zap.NewNop()
	svc := NewMockAIService(logger)

	req := MatchRequest{
		JobText:      "Java案件",
		EngineerText: "Java 4年",
	}

	resp, err := svc.Match(context.Background(), req)
	if err != nil {
		t.Fatalf("Match failed: %v", err)
	}

	if resp.TotalScore != 72 {
		t.Errorf("TotalScore = %d, want 72", resp.TotalScore)
	}
	if resp.Grade != "B" {
		t.Errorf("Grade = %s, want B", resp.Grade)
	}
	if resp.GradeLabel != "提案検討可" {
		t.Errorf("GradeLabel = %s, want 提案検討可", resp.GradeLabel)
	}
	if resp.TokensUsed != 3200 {
		t.Errorf("TokensUsed = %d, want 3200", resp.TokensUsed)
	}
	if resp.ModelUsed != "mock-ai" {
		t.Errorf("ModelUsed = %s, want mock-ai", resp.ModelUsed)
	}
}

func TestMockAIService_Match_ResponseFormat(t *testing.T) {
	logger := zap.NewNop()
	svc := NewMockAIService(logger)

	req := MatchRequest{
		JobText:      "テスト案件",
		EngineerText: "テストエンジニア",
	}

	resp, err := svc.Match(context.Background(), req)
	if err != nil {
		t.Fatalf("Match failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		t.Fatalf("Result JSON parse failed: %v", err)
	}

	requiredFields := []string{
		"job_summary",
		"engineer_summary",
		"total_score",
		"grade",
		"grade_label",
		"scores",
		"ng_flags",
		"negatives",
		"positives",
		"warnings",
		"advice",
		"confirmation_hints",
	}
	for _, field := range requiredFields {
		if _, ok := result[field]; !ok {
			t.Errorf("missing required field: %s", field)
		}
	}
}

func TestMockAIService_Match_ScoresFormat(t *testing.T) {
	logger := zap.NewNop()
	svc := NewMockAIService(logger)

	req := MatchRequest{
		JobText:      "テスト案件",
		EngineerText: "テストエンジニア",
	}

	resp, err := svc.Match(context.Background(), req)
	if err != nil {
		t.Fatalf("Match failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		t.Fatalf("Result JSON parse failed: %v", err)
	}

	scores, ok := result["scores"].(map[string]interface{})
	if !ok {
		t.Fatal("scores field is not a map")
	}

	requiredScoreFields := []string{
		"skill",
		"timing",
		"rate",
		"experience_years",
		"location",
		"industry",
	}
	for _, field := range requiredScoreFields {
		scoreObj, ok := scores[field]
		if !ok {
			t.Errorf("missing score field: %s", field)
			continue
		}
		scoreMap, ok := scoreObj.(map[string]interface{})
		if !ok {
			t.Errorf("score field %s is not a map", field)
			continue
		}
		if _, ok := scoreMap["score"]; !ok {
			t.Errorf("score field %s missing 'score' key", field)
		}
		if _, ok := scoreMap["max"]; !ok {
			t.Errorf("score field %s missing 'max' key", field)
		}
		if _, ok := scoreMap["reason"]; !ok {
			t.Errorf("score field %s missing 'reason' key", field)
		}
	}
}

func TestMockAIService_Match_SkillDetails(t *testing.T) {
	logger := zap.NewNop()
	svc := NewMockAIService(logger)

	req := MatchRequest{
		JobText:      "テスト案件",
		EngineerText: "テストエンジニア",
	}

	resp, err := svc.Match(context.Background(), req)
	if err != nil {
		t.Fatalf("Match failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		t.Fatalf("Result JSON parse failed: %v", err)
	}

	scores := result["scores"].(map[string]interface{})
	skill := scores["skill"].(map[string]interface{})

	requiredSkills, ok := skill["required_skills"].([]interface{})
	if !ok {
		t.Fatal("required_skills is not an array")
	}
	if len(requiredSkills) == 0 {
		t.Fatal("required_skills is empty")
	}

	validStatuses := map[string]bool{
		"met":     true,
		"partial": true,
		"unmet":   true,
	}

	for i, rs := range requiredSkills {
		skillItem, ok := rs.(map[string]interface{})
		if !ok {
			t.Errorf("required_skills[%d] is not a map", i)
			continue
		}

		status, ok := skillItem["status"].(string)
		if !ok {
			t.Errorf("required_skills[%d] missing 'status' field", i)
			continue
		}
		if !validStatuses[status] {
			t.Errorf("required_skills[%d] has invalid status %q, want one of: met, partial, not_met", i, status)
		}

		if _, ok := skillItem["skill"]; !ok {
			t.Errorf("required_skills[%d] missing 'skill' field", i)
		}
		if _, ok := skillItem["detail"]; !ok {
			t.Errorf("required_skills[%d] missing 'detail' field", i)
		}
	}
}
