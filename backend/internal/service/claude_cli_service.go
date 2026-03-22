package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"go.uber.org/zap"
)

// ClaudeCliAIService ローカル開発用: claude CLIコマンドでマッチングを実行
type ClaudeCliAIService struct {
	systemPrompt string
	logger       *zap.Logger
}

// NewClaudeCliAIService ClaudeCliAIServiceを作成
func NewClaudeCliAIService(systemPrompt string, logger *zap.Logger) *ClaudeCliAIService {
	return &ClaudeCliAIService{
		systemPrompt: systemPrompt,
		logger:       logger,
	}
}

// Match claude CLIを使ってマッチング実行
func (s *ClaudeCliAIService) Match(ctx context.Context, req MatchRequest) (*MatchResponse, error) {
	s.logger.Info("ClaudeCliAIService: claude CLIでマッチング実行")

	// ユーザーメッセージを構築
	var sb strings.Builder
	sb.WriteString("## 案件情報\n")
	sb.WriteString(req.JobText)
	sb.WriteString("\n\n## エンジニア情報\n")
	sb.WriteString(req.EngineerText)

	if len(req.Supplement) > 0 {
		sb.WriteString("\n\n## 補足情報\n")
		for k, v := range req.Supplement {
			sb.WriteString(fmt.Sprintf("- %s: %v\n", k, v))
		}
	}

	if req.MarginAmount > 0 {
		sb.WriteString(fmt.Sprintf("\n## マージン設定\n- 種別: %s\n- 金額: %d円\n", req.MarginType, req.MarginAmount))
	}

	userMessage := sb.String()

	// システムプロンプト + ユーザーメッセージを結合
	fullPrompt := s.systemPrompt + "\n\n---\n\n" + userMessage

	// claude CLI実行: stdin経由でプロンプトを渡す（引数長制限を回避）
	cmd := exec.CommandContext(ctx, "claude", "-p", "-", "--output-format", "json")
	cmd.Stdin = strings.NewReader(fullPrompt)

	s.logger.Info("claude CLI実行開始", zap.Int("prompt_length", len(fullPrompt)))

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			s.logger.Error("claude CLI実行エラー",
				zap.Error(err),
				zap.String("stderr", string(exitErr.Stderr)),
			)
		}
		return nil, fmt.Errorf("claude CLI実行失敗: %w", err)
	}

	s.logger.Debug("claude CLI raw output", zap.Int("length", len(output)))

	// claude --output-format json は {"type":"result","result":"..."} 形式で返す
	var cliResponse struct {
		Result string `json:"result"`
	}
	if err := json.Unmarshal(output, &cliResponse); err != nil {
		// JSON形式でない場合はそのまま使う
		cliResponse.Result = string(output)
	}

	// 結果テキストからJSON部分を抽出
	resultJSON, err := extractJSON(cliResponse.Result)
	if err != nil {
		s.logger.Error("claude CLIレスポンスのJSON抽出失敗",
			zap.Error(err),
			zap.String("raw", cliResponse.Result[:min(500, len(cliResponse.Result))]),
		)
		return nil, fmt.Errorf("レスポンスのJSON抽出失敗: %w", err)
	}

	// スコアとグレードを抽出
	var result map[string]interface{}
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("結果JSONの解析失敗: %w", err)
	}

	totalScore := int(result["total_score"].(float64))
	grade := result["grade"].(string)
	gradeLabel := result["grade_label"].(string)

	return &MatchResponse{
		TotalScore: totalScore,
		Grade:      grade,
		GradeLabel: gradeLabel,
		Result:     resultJSON,
		TokensUsed: 0, // CLIではトークン数不明
		ModelUsed:  "claude-cli",
	}, nil
}
