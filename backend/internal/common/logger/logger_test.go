package logger

import (
	"errors"
	"strings"
	"testing"

	"go.uber.org/zap"
)

func TestInitLogger_Development(t *testing.T) {
	l, err := InitLogger(false)
	if err != nil {
		t.Fatalf("InitLogger(false) returned error: %v", err)
	}
	if l == nil {
		t.Fatal("InitLogger(false) returned nil logger")
	}
}

func TestInitLogger_Production(t *testing.T) {
	l, err := InitLogger(true)
	if err != nil {
		t.Fatalf("InitLogger(true) returned error: %v", err)
	}
	if l == nil {
		t.Fatal("InitLogger(true) returned nil logger")
	}
}

func TestInitLogger_WithLogLevel(t *testing.T) {
	t.Setenv("LOG_LEVEL", "warn")

	l, err := InitLogger(false)
	if err != nil {
		t.Fatalf("InitLogger returned error: %v", err)
	}
	if l == nil {
		t.Fatal("InitLogger returned nil logger")
	}

	// The logger should be created successfully with warn level.
	// Verify by checking that the logger's core is enabled at warn but not at info.
	if l.Core().Enabled(zap.InfoLevel) {
		t.Error("logger should not be enabled at info level when LOG_LEVEL=warn")
	}
	if !l.Core().Enabled(zap.WarnLevel) {
		t.Error("logger should be enabled at warn level when LOG_LEVEL=warn")
	}
}

func TestInitLogger_WithInvalidLogLevel(t *testing.T) {
	t.Setenv("LOG_LEVEL", "not_a_valid_level")

	l, err := InitLogger(false)
	if err != nil {
		t.Fatalf("InitLogger returned error: %v", err)
	}
	if l == nil {
		t.Fatal("InitLogger returned nil logger")
	}

	// Invalid log level should be ignored, using default level.
	// Development default is debug level.
	if !l.Core().Enabled(zap.DebugLevel) {
		t.Error("logger should use default debug level when LOG_LEVEL is invalid (development mode)")
	}
}

func TestGetRequestLogger_WithUserID(t *testing.T) {
	base, _ := zap.NewDevelopment()
	reqLogger := GetRequestLogger(base, "req-123", "user-456")

	if reqLogger == nil {
		t.Fatal("GetRequestLogger returned nil")
	}

	// Verify the returned logger is not the same instance as the base logger
	// (it should be a new logger with additional fields).
	if reqLogger == base {
		t.Error("GetRequestLogger should return a new logger instance with fields")
	}
}

func TestGetRequestLogger_WithoutUserID(t *testing.T) {
	base, _ := zap.NewDevelopment()
	reqLogger := GetRequestLogger(base, "req-789", "")

	if reqLogger == nil {
		t.Fatal("GetRequestLogger returned nil")
	}

	if reqLogger == base {
		t.Error("GetRequestLogger should return a new logger instance with request_id field")
	}
}

func TestLogAndWrapError(t *testing.T) {
	l, _ := zap.NewDevelopment()
	originalErr := errors.New("original error")

	wrappedErr := LogAndWrapError(l, originalErr, "operation failed")

	if wrappedErr == nil {
		t.Fatal("LogAndWrapError returned nil")
	}

	// Verify the wrapped error message contains the message prefix
	if !strings.Contains(wrappedErr.Error(), "operation failed") {
		t.Errorf("wrapped error = %q, want it to contain %q", wrappedErr.Error(), "operation failed")
	}

	// Verify the original error is wrapped (can be unwrapped)
	if !errors.Is(wrappedErr, originalErr) {
		t.Error("wrapped error should wrap the original error (errors.Is should return true)")
	}

	// Verify the format is "message: original error"
	expected := "operation failed: original error"
	if wrappedErr.Error() != expected {
		t.Errorf("wrapped error = %q, want %q", wrappedErr.Error(), expected)
	}
}

func TestEnsureLogger_Nil(t *testing.T) {
	result := EnsureLogger(nil)

	if result == nil {
		t.Fatal("EnsureLogger(nil) should return a non-nil nop logger")
	}

	// A nop logger should not be enabled at any level
	if result.Core().Enabled(zap.DebugLevel) {
		t.Error("nop logger should not be enabled at debug level")
	}
}

func TestEnsureLogger_NonNil(t *testing.T) {
	original, _ := zap.NewDevelopment()
	result := EnsureLogger(original)

	if result != original {
		t.Error("EnsureLogger should return the same logger instance when non-nil")
	}
}
