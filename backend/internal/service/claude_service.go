package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/duesk/ivy/internal/config"
	"go.uber.org/zap"
)

// ClaudeAIService Claude APIを使った実AIサービス
type ClaudeAIService struct {
	apiKey     string
	model      string
	timeout    time.Duration
	logger     *zap.Logger
	httpClient *http.Client
	prompt     string
}

// NewClaudeAIService ClaudeAIServiceを作成
func NewClaudeAIService(cfg *config.Config, model string, systemPrompt string, logger *zap.Logger) *ClaudeAIService {
	return &ClaudeAIService{
		apiKey:  cfg.AI.APIKey,
		model:   model,
		timeout: cfg.AI.Timeout,
		logger:  logger,
		httpClient: &http.Client{
			Timeout: cfg.AI.Timeout,
		},
		prompt: systemPrompt,
	}
}

// claudeRequest Claude APIリクエスト
type claudeRequest struct {
	Model     string           `json:"model"`
	MaxTokens int              `json:"max_tokens"`
	System    string           `json:"system,omitempty"`
	Messages  []claudeMessage  `json:"messages"`
}

// claudeMessage Claude APIメッセージ
type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// claudeResponse Claude APIレスポンス
type claudeResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	StopReason string `json:"stop_reason"`
}

// Match マッチング実行
func (s *ClaudeAIService) Match(ctx context.Context, req MatchRequest) (*MatchResponse, error) {
	userMessage := s.buildUserMessage(req)

	claudeReq := claudeRequest{
		Model:     s.model,
		MaxTokens: 4096,
		System:    s.prompt,
		Messages: []claudeMessage{
			{Role: "user", Content: userMessage},
		},
	}

	// リトライポリシー（設計書セクション4.7準拠）
	// 429: 最大3回、指数バックオフ（1s→2s→4s）
	// 500/503: 1回、3秒後
	// JSONパース失敗: 1回、即時
	// タイムアウト: 1回、即時
	// 401: リトライしない
	var lastErr error
	apiAttempt := 0
	maxAPIAttempts := 4 // 初回 + 最大3回リトライ

	for apiAttempt < maxAPIAttempts {
		resp, err := s.callAPI(ctx, claudeReq)
		if err != nil {
			lastErr = err
			apiAttempt++

			if apiErr, ok := err.(*apiError); ok {
				switch {
				case apiErr.StatusCode == 429:
					// レート制限: 最大3回、指数バックオフ
					if apiAttempt >= 4 {
						return nil, fmt.Errorf("Claude API呼び出し失敗（レート制限）: %w", err)
					}
					backoff := time.Duration(1<<uint(apiAttempt-1)) * time.Second
					s.logger.Warn("Claude API レート制限リトライ",
						zap.Int("attempt", apiAttempt),
						zap.Duration("backoff", backoff),
					)
					time.Sleep(backoff)
					continue
				case apiErr.StatusCode >= 500:
					// サーバーエラー: 1回、3秒後
					if apiAttempt >= 2 {
						return nil, fmt.Errorf("Claude API呼び出し失敗（サーバーエラー）: %w", err)
					}
					s.logger.Warn("Claude API サーバーエラーリトライ", zap.Int("status", apiErr.StatusCode))
					time.Sleep(3 * time.Second)
					continue
				default:
					// 401等: リトライしない
					return nil, fmt.Errorf("Claude API呼び出し失敗: %w", err)
				}
			}

			// ネットワークエラー等: 1回、即時リトライ
			if apiAttempt >= 2 {
				return nil, fmt.Errorf("Claude API呼び出し失敗: %w", err)
			}
			s.logger.Warn("Claude API リトライ（ネットワークエラー）", zap.Error(err))
			continue
		}

		// レスポンスからJSON抽出
		resultJSON, err := extractJSON(resp.Content[0].Text)
		if err != nil {
			lastErr = err
			apiAttempt++
			// JSONパース失敗: 1回、即時リトライ
			if apiAttempt >= 2 {
				return nil, fmt.Errorf("レスポンスのJSONパース失敗: %w", err)
			}
			s.logger.Warn("JSONパース失敗、即時リトライ", zap.Error(err))
			continue
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
			TokensUsed: resp.Usage.InputTokens + resp.Usage.OutputTokens,
			ModelUsed:  s.model,
		}, nil
	}

	return nil, fmt.Errorf("最大リトライ回数を超過: %w", lastErr)
}

// buildUserMessage ユーザーメッセージを構築
func (s *ClaudeAIService) buildUserMessage(req MatchRequest) string {
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

	return sb.String()
}

// callAPI Claude APIを呼び出す
func (s *ClaudeAIService) callAPI(ctx context.Context, req claudeRequest) (*claudeResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("リクエストのJSON化失敗: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("HTTPリクエスト作成失敗: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", s.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP通信失敗: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("レスポンス読み取り失敗: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &apiError{
			StatusCode: resp.StatusCode,
			Message:    string(respBody),
		}
	}

	var claudeResp claudeResponse
	if err := json.Unmarshal(respBody, &claudeResp); err != nil {
		return nil, fmt.Errorf("レスポンスのJSON解析失敗: %w", err)
	}

	if len(claudeResp.Content) == 0 {
		return nil, fmt.Errorf("空のレスポンス")
	}

	return &claudeResp, nil
}

// apiError API呼び出しエラー
type apiError struct {
	StatusCode int
	Message    string
}

func (e *apiError) Error() string {
	return fmt.Sprintf("Claude API error (status %d): %s", e.StatusCode, e.Message)
}

// isRetryable リトライ可能なエラーか判定
func isRetryable(err error) bool {
	if apiErr, ok := err.(*apiError); ok {
		return apiErr.StatusCode == 429 || apiErr.StatusCode >= 500
	}
	return false
}

// extractJSON テキストからJSON部分を抽出
func extractJSON(text string) ([]byte, error) {
	// ```json ... ``` ブロックを検出
	re := regexp.MustCompile("(?s)```json\\s*\\n?(.*?)\\n?```")
	matches := re.FindStringSubmatch(text)
	if len(matches) >= 2 {
		text = matches[1]
	}

	text = strings.TrimSpace(text)

	// JSONとしてパース可能か検証
	var result json.RawMessage
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		// { から始まるJSONブロックを探す
		start := strings.Index(text, "{")
		end := strings.LastIndex(text, "}")
		if start >= 0 && end > start {
			jsonStr := text[start : end+1]
			if err2 := json.Unmarshal([]byte(jsonStr), &result); err2 != nil {
				return nil, fmt.Errorf("JSONの抽出に失敗: %w", err2)
			}
			return []byte(jsonStr), nil
		}
		return nil, fmt.Errorf("JSONの抽出に失敗: %w", err)
	}

	return []byte(text), nil
}
