package service

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/duesk/ivy/internal/config"
	"go.uber.org/zap"
)

// --- extractJSON tests ---

func TestExtractJSON_MarkdownBlock(t *testing.T) {
	input := "Here is the result:\n```json\n{\"total_score\": 72, \"grade\": \"B\"}\n```\nDone."
	result, err := extractJSON(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := `{"total_score": 72, "grade": "B"}`
	if string(result) != expected {
		t.Errorf("got %s, want %s", string(result), expected)
	}
}

func TestExtractJSON_RawJSON(t *testing.T) {
	input := `{"total_score": 85, "grade": "A"}`
	result, err := extractJSON(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(result) != input {
		t.Errorf("got %s, want %s", string(result), input)
	}
}

func TestExtractJSON_WithSurroundingText(t *testing.T) {
	input := "Analysis complete.\n{\"total_score\": 50, \"grade\": \"C\"}\nPlease review."
	result, err := extractJSON(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := `{"total_score": 50, "grade": "C"}`
	if string(result) != expected {
		t.Errorf("got %s, want %s", string(result), expected)
	}
}

func TestExtractJSON_NestedJSON(t *testing.T) {
	input := `{"total_score": 80, "scores": {"skill": {"score": 40, "max": 50}, "timing": {"score": 10, "max": 10}}, "grade": "A"}`
	result, err := extractJSON(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(result) != input {
		t.Errorf("got %s, want %s", string(result), input)
	}
}

func TestExtractJSON_NoJSON(t *testing.T) {
	input := "This is just plain text with no JSON."
	_, err := extractJSON(input)
	if err == nil {
		t.Error("expected error for non-JSON input")
	}
}

func TestExtractJSON_EmptyString(t *testing.T) {
	_, err := extractJSON("")
	if err == nil {
		t.Error("expected error for empty string")
	}
}

func TestExtractJSON_WhitespaceOnly(t *testing.T) {
	_, err := extractJSON("   \n\t  ")
	if err == nil {
		t.Error("expected error for whitespace-only string")
	}
}

// --- isRetryable tests ---

func TestIsRetryable_429(t *testing.T) {
	err := &apiError{StatusCode: 429, Message: "rate limited"}
	if !isRetryable(err) {
		t.Error("429 should be retryable")
	}
}

func TestIsRetryable_500(t *testing.T) {
	err := &apiError{StatusCode: 500, Message: "server error"}
	if !isRetryable(err) {
		t.Error("500 should be retryable")
	}
}

func TestIsRetryable_503(t *testing.T) {
	err := &apiError{StatusCode: 503, Message: "service unavailable"}
	if !isRetryable(err) {
		t.Error("503 should be retryable")
	}
}

func TestIsRetryable_401(t *testing.T) {
	err := &apiError{StatusCode: 401, Message: "unauthorized"}
	if isRetryable(err) {
		t.Error("401 should not be retryable")
	}
}

func TestIsRetryable_400(t *testing.T) {
	err := &apiError{StatusCode: 400, Message: "bad request"}
	if isRetryable(err) {
		t.Error("400 should not be retryable")
	}
}

func TestIsRetryable_NonAPIError(t *testing.T) {
	err := fmt.Errorf("generic network error")
	if isRetryable(err) {
		t.Error("non-apiError should not be retryable")
	}
}

// --- apiError.Error tests ---

func TestApiError_Error(t *testing.T) {
	err := &apiError{StatusCode: 429, Message: "too many requests"}
	expected := "Claude API error (status 429): too many requests"
	if err.Error() != expected {
		t.Errorf("got %q, want %q", err.Error(), expected)
	}
}

// --- buildUserMessage tests ---

func newTestClaudeService() *ClaudeAIService {
	logger := zap.NewNop()
	cfg := &config.Config{
		AI: config.AIConfig{
			APIKey:  "test-key",
			Timeout: 30 * time.Second,
		},
	}
	return NewClaudeAIService(cfg, "test-model", "test-prompt", logger)
}

func TestBuildUserMessage_Basic(t *testing.T) {
	svc := newTestClaudeService()
	req := MatchRequest{
		JobText:      "Java開発案件",
		EngineerText: "Java経験5年のエンジニア",
	}

	msg := svc.buildUserMessage(req)

	if !strings.Contains(msg, "## 案件情報") {
		t.Error("expected message to contain '## 案件情報'")
	}
	if !strings.Contains(msg, "Java開発案件") {
		t.Error("expected message to contain job text")
	}
	if !strings.Contains(msg, "## エンジニア情報") {
		t.Error("expected message to contain '## エンジニア情報'")
	}
	if !strings.Contains(msg, "Java経験5年のエンジニア") {
		t.Error("expected message to contain engineer text")
	}
	if strings.Contains(msg, "## 補足情報") {
		t.Error("should not contain supplement section when no supplement provided")
	}
	if strings.Contains(msg, "## マージン設定") {
		t.Error("should not contain margin section when margin is 0")
	}
}

func TestBuildUserMessage_WithSupplement(t *testing.T) {
	svc := newTestClaudeService()
	req := MatchRequest{
		JobText:      "案件テキスト",
		EngineerText: "エンジニアテキスト",
		Supplement: map[string]interface{}{
			"nationality": "日本",
			"rate":        60,
		},
	}

	msg := svc.buildUserMessage(req)

	if !strings.Contains(msg, "## 補足情報") {
		t.Error("expected message to contain '## 補足情報'")
	}
	if !strings.Contains(msg, "nationality") {
		t.Error("expected message to contain supplement key 'nationality'")
	}
	if !strings.Contains(msg, "日本") {
		t.Error("expected message to contain supplement value '日本'")
	}
}

func TestBuildUserMessage_WithMargin(t *testing.T) {
	svc := newTestClaudeService()
	req := MatchRequest{
		JobText:      "案件テキスト",
		EngineerText: "エンジニアテキスト",
		MarginAmount: 50000,
		MarginType:   "fixed",
	}

	msg := svc.buildUserMessage(req)

	if !strings.Contains(msg, "## マージン設定") {
		t.Error("expected message to contain '## マージン設定'")
	}
	if !strings.Contains(msg, "50000") {
		t.Error("expected message to contain margin amount '50000'")
	}
	if !strings.Contains(msg, "fixed") {
		t.Error("expected message to contain margin type 'fixed'")
	}
}
