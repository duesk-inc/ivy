package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
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
	preamble := "重要: あなたはテキスト分析専用です。ツール（ファイル操作、Web検索、コード実行等）は一切使用せず、テキストのみで応答してください。以下の指示に従ってJSON形式で分析結果を直接出力してください。\n\n"
	reminder := "\n\n---\n\n【リマインド】必須スキル(required_skills)と尚可スキル(optional_skills)の両方を案件テキストから個別に抽出し、全項目をJSON配列に列挙すること。尚可スキルが案件に記載されている場合、optional_skillsを空配列にしないこと。\n\n"
	fullPrompt := preamble + s.systemPrompt + reminder + userMessage

	// claude CLI実行: stdin経由でプロンプトを渡す（引数長制限を回避）
	cmd := exec.CommandContext(ctx, "claude", "-p", "-", "--output-format", "json")
	cmd.Stdin = strings.NewReader(fullPrompt)

	// ANTHROPIC_API_KEYを除外してMaxプランのOAuth認証を使用させる
	env := os.Environ()
	filteredEnv := make([]string, 0, len(env))
	for _, e := range env {
		if !strings.HasPrefix(e, "ANTHROPIC_API_KEY=") {
			filteredEnv = append(filteredEnv, e)
		}
	}
	cmd.Env = filteredEnv

	s.logger.Info("claude CLI実行開始", zap.Int("prompt_length", len(fullPrompt)))

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		stderrStr := stderr.String()
		stdoutStr := stdout.String()
		if len(stdoutStr) > 500 {
			stdoutStr = stdoutStr[:500]
		}
		s.logger.Error("claude CLI実行エラー",
			zap.Error(err),
			zap.String("stderr", stderrStr),
			zap.String("stdout", stdoutStr),
		)
		return nil, fmt.Errorf("claude CLI実行失敗: %w", err)
	}

	output := stdout.Bytes()
	s.logger.Info("claude CLI raw output", zap.Int("length", len(output)))

	// claude --output-format json は {"type":"result","result":"..."} 形式で返す
	var cliResponse struct {
		Type    string `json:"type"`
		Subtype string `json:"subtype"`
		IsError bool   `json:"is_error"`
		Result  string `json:"result"`
	}
	if err := json.Unmarshal(output, &cliResponse); err != nil {
		// JSON形式でない場合はそのまま使う
		cliResponse.Result = string(output)
	}

	// CLIエラーチェック
	if cliResponse.IsError || cliResponse.Subtype == "error" {
		s.logger.Error("claude CLIがエラーを返した",
			zap.String("result", cliResponse.Result),
			zap.String("subtype", cliResponse.Subtype),
		)
		return nil, fmt.Errorf("claude CLIエラー: %s", cliResponse.Result)
	}

	if cliResponse.Result == "" {
		rawStr := string(output)
		if len(rawStr) > 500 {
			rawStr = rawStr[:500]
		}
		s.logger.Error("claude CLIの結果が空", zap.String("raw_output", rawStr))
		return nil, fmt.Errorf("claude CLIの結果が空です")
	}

	// 結果テキストからJSON部分を抽出
	resultJSON, err := extractJSON(cliResponse.Result)
	if err != nil {
		rawStr := cliResponse.Result
		if len(rawStr) > 500 {
			rawStr = rawStr[:500]
		}
		s.logger.Error("claude CLIレスポンスのJSON抽出失敗",
			zap.Error(err),
			zap.String("raw", rawStr),
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
