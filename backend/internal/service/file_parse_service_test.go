package service

import (
	"testing"
)

func TestSanitizeText(t *testing.T) {
	t.Run("nullバイトを除去", func(t *testing.T) {
		text := sanitizeText("hello\x00world")
		if text != "helloworld" {
			t.Errorf("expected 'helloworld', got %q", text)
		}
	})

	t.Run("改行・タブは保持", func(t *testing.T) {
		text := sanitizeText("hello\nworld\ttab")
		if text != "hello\nworld\ttab" {
			t.Errorf("expected 'hello\\nworld\\ttab', got %q", text)
		}
	})

	t.Run("空文字列", func(t *testing.T) {
		text := sanitizeText("")
		if text != "" {
			t.Errorf("expected empty string, got %q", text)
		}
	})
}
