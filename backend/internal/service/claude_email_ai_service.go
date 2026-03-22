package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// claudeEmailAIService Claude APIを使ったメール分類・パースサービス
type claudeEmailAIService struct {
	apiKey        string
	defaultModel  string
	classifyModel string
	httpClient    *http.Client
	logger        *zap.Logger
}

// NewClaudeEmailAIService ClaudeEmailAIServiceを作成
func NewClaudeEmailAIService(apiKey, defaultModel string, logger *zap.Logger) EmailAIService {
	return &claudeEmailAIService{
		apiKey:        apiKey,
		defaultModel:  defaultModel,
		classifyModel: "claude-haiku-4-5-20251001",
		httpClient:    &http.Client{Timeout: 60 * time.Second},
		logger:        logger,
	}
}

const (
	classifySystemPrompt = "あなたはSES業界のメール分類AIです。メールの件名と本文を分析し、以下の3つに分類してください: job(案件情報), engineer(人材情報/スキルシート), other(その他)。JSONで回答してください: {\"classification\": \"job\", \"confidence\": 0.95}"

	parseJobSystemPrompt = "あなたはSES案件メールの構造化データ抽出AIです。メール本文から以下の情報をJSONで抽出してください: name(案件名), skills(必要スキル配列), rate_min(単価下限/万円), rate_max(単価上限/万円), location(勤務地), remote(リモート可否), start_month(開始月/YYYY-MM), settlement(精算), nationality_ok(外国籍可/bool), freelance_ok(フリーランス可/bool), age_limit(年齢上限), conditions(その他条件)。不明な項目はnullとしてください。"

	parseEngineerSystemPrompt = "あなたはSES人材メールの構造化データ抽出AIです。メール本文とスキルシートテキストから以下の情報をJSONで抽出してください: initials(イニシャル), age(年齢), gender(性別), skills(スキル配列), rate(希望単価/万円), start_month(稼働開始月/YYYY-MM), nationality(国籍), employment_type(雇用形態), affiliation(所属), nearest_station(最寄り駅)。不明な項目はnullとしてください。"
)

// ClassifyEmail メールを分類する（Haikuモデル使用）
func (s *claudeEmailAIService) ClassifyEmail(ctx context.Context, subject, body, sender string) (*EmailClassification, error) {
	userMessage := fmt.Sprintf("件名: %s\n送信者: %s\n本文:\n%s", subject, sender, body)

	resultJSON, err := s.callWithRetry(ctx, s.classifyModel, classifySystemPrompt, userMessage, 1024)
	if err != nil {
		return nil, fmt.Errorf("メール分類失敗: %w", err)
	}

	var classification EmailClassification
	if err := json.Unmarshal(resultJSON, &classification); err != nil {
		return nil, fmt.Errorf("分類結果のJSONパース失敗: %w", err)
	}

	s.logger.Info("メール分類完了",
		zap.String("classification", classification.Classification),
		zap.Float64("confidence", classification.Confidence),
	)

	return &classification, nil
}

// ParseJobFromEmail メール本文から案件情報を抽出する
func (s *claudeEmailAIService) ParseJobFromEmail(ctx context.Context, emailText string) (json.RawMessage, error) {
	resultJSON, err := s.callWithRetry(ctx, s.defaultModel, parseJobSystemPrompt, emailText, 4096)
	if err != nil {
		return nil, fmt.Errorf("案件情報パース失敗: %w", err)
	}

	s.logger.Info("案件メールパース完了")

	return json.RawMessage(resultJSON), nil
}

// ParseEngineerFromEmail メール本文とスキルシートから人材情報を抽出する
func (s *claudeEmailAIService) ParseEngineerFromEmail(ctx context.Context, emailText, attachmentText string) (json.RawMessage, error) {
	var userMessage string
	if attachmentText != "" {
		userMessage = fmt.Sprintf("## メール本文\n%s\n\n## スキルシート/添付ファイル\n%s", emailText, attachmentText)
	} else {
		userMessage = fmt.Sprintf("## メール本文\n%s", emailText)
	}

	resultJSON, err := s.callWithRetry(ctx, s.defaultModel, parseEngineerSystemPrompt, userMessage, 4096)
	if err != nil {
		return nil, fmt.Errorf("人材情報パース失敗: %w", err)
	}

	s.logger.Info("人材メールパース完了")

	return json.RawMessage(resultJSON), nil
}

// callWithRetry Claude APIを呼び出し、リトライとJSON抽出を行う
func (s *claudeEmailAIService) callWithRetry(ctx context.Context, model, systemPrompt, userMessage string, maxTokens int) ([]byte, error) {
	req := claudeRequest{
		Model:     model,
		MaxTokens: maxTokens,
		System:    systemPrompt,
		Messages: []claudeMessage{
			{Role: "user", Content: userMessage},
		},
	}

	var lastErr error
	apiAttempt := 0
	maxAPIAttempts := 4

	for apiAttempt < maxAPIAttempts {
		resp, err := s.callEmailAPI(ctx, req)
		if err != nil {
			lastErr = err
			apiAttempt++

			if apiErr, ok := err.(*apiError); ok {
				switch {
				case apiErr.StatusCode == 429:
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
					if apiAttempt >= 2 {
						return nil, fmt.Errorf("Claude API呼び出し失敗（サーバーエラー）: %w", err)
					}
					s.logger.Warn("Claude API サーバーエラーリトライ", zap.Int("status", apiErr.StatusCode))
					time.Sleep(3 * time.Second)
					continue
				default:
					return nil, fmt.Errorf("Claude API呼び出し失敗: %w", err)
				}
			}

			if apiAttempt >= 2 {
				return nil, fmt.Errorf("Claude API呼び出し失敗: %w", err)
			}
			s.logger.Warn("Claude API リトライ（ネットワークエラー）", zap.Error(err))
			continue
		}

		resultJSON, err := extractJSON(resp.Content[0].Text)
		if err != nil {
			lastErr = err
			apiAttempt++
			if apiAttempt >= 2 {
				return nil, fmt.Errorf("レスポンスのJSONパース失敗: %w", err)
			}
			s.logger.Warn("JSONパース失敗、即時リトライ", zap.Error(err))
			continue
		}

		return resultJSON, nil
	}

	return nil, fmt.Errorf("最大リトライ回数を超過: %w", lastErr)
}

// callEmailAPI Claude APIを呼び出す
func (s *claudeEmailAIService) callEmailAPI(ctx context.Context, req claudeRequest) (*claudeResponse, error) {
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
