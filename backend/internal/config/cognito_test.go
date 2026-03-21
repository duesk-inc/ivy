package config

import (
	"strings"
	"testing"
	"time"
)

func TestCognitoConfig_GetJWKURL_WithEndpoint(t *testing.T) {
	cfg := &CognitoConfig{
		Endpoint:   "http://cognito-local:9229",
		UserPoolID: "local_pool_123",
		Region:     "ap-northeast-1",
	}

	jwkURL := cfg.GetJWKURL()
	expected := "http://cognito-local:9229/local_pool_123/.well-known/jwks.json"

	if jwkURL != expected {
		t.Errorf("GetJWKURL() = %q, want %q", jwkURL, expected)
	}
}

func TestCognitoConfig_GetJWKURL_WithoutEndpoint(t *testing.T) {
	cfg := &CognitoConfig{
		Endpoint:   "",
		UserPoolID: "ap-northeast-1_AbCdEfG",
		Region:     "ap-northeast-1",
	}

	jwkURL := cfg.GetJWKURL()
	expected := "https://cognito-idp.ap-northeast-1.amazonaws.com/ap-northeast-1_AbCdEfG/.well-known/jwks.json"

	if jwkURL != expected {
		t.Errorf("GetJWKURL() = %q, want %q", jwkURL, expected)
	}
}

func TestCognitoConfig_GetIssuer_WithEndpoint(t *testing.T) {
	cfg := &CognitoConfig{
		Endpoint:   "http://cognito-local:9229",
		UserPoolID: "local_pool_123",
		Region:     "ap-northeast-1",
	}

	issuer := cfg.GetIssuer()

	// cognito-local hostname should be normalized to 0.0.0.0
	expected := "http://0.0.0.0:9229/local_pool_123"
	if issuer != expected {
		t.Errorf("GetIssuer() = %q, want %q", issuer, expected)
	}
}

func TestCognitoConfig_GetIssuer_WithoutEndpoint(t *testing.T) {
	cfg := &CognitoConfig{
		Endpoint:   "",
		UserPoolID: "ap-northeast-1_AbCdEfG",
		Region:     "ap-northeast-1",
	}

	issuer := cfg.GetIssuer()
	expected := "https://cognito-idp.ap-northeast-1.amazonaws.com/ap-northeast-1_AbCdEfG"

	if issuer != expected {
		t.Errorf("GetIssuer() = %q, want %q", issuer, expected)
	}
}

func TestCognitoConfig_GetIssuer_Localhost(t *testing.T) {
	cfg := &CognitoConfig{
		Endpoint:   "http://localhost:9229",
		UserPoolID: "local_pool_123",
		Region:     "ap-northeast-1",
	}

	issuer := cfg.GetIssuer()

	// localhost should be normalized to 0.0.0.0
	if !strings.HasPrefix(issuer, "http://0.0.0.0:9229/") {
		t.Errorf("GetIssuer() = %q, expected localhost to be normalized to 0.0.0.0", issuer)
	}

	expected := "http://0.0.0.0:9229/local_pool_123"
	if issuer != expected {
		t.Errorf("GetIssuer() = %q, want %q", issuer, expected)
	}
}

func TestCognitoConfig_GetIssuer_InvalidURL(t *testing.T) {
	cfg := &CognitoConfig{
		Endpoint:   "://invalid-url",
		UserPoolID: "local_pool_123",
		Region:     "ap-northeast-1",
	}

	issuer := cfg.GetIssuer()

	// Should fallback to default
	expected := "http://0.0.0.0:9229/local_pool_123"
	if issuer != expected {
		t.Errorf("GetIssuer() = %q, want fallback %q", issuer, expected)
	}
}

func TestCognitoConfig_SetDefaults(t *testing.T) {
	cfg := &CognitoConfig{}

	cfg.SetDefaults()

	if cfg.TokenExpiration != time.Hour {
		t.Errorf("TokenExpiration = %v, want %v", cfg.TokenExpiration, time.Hour)
	}
	if cfg.JWKCacheDuration != time.Hour {
		t.Errorf("JWKCacheDuration = %v, want %v", cfg.JWKCacheDuration, time.Hour)
	}
}

func TestCognitoConfig_SetDefaults_AlreadySet(t *testing.T) {
	cfg := &CognitoConfig{
		TokenExpiration:  30 * time.Minute,
		JWKCacheDuration: 2 * time.Hour,
	}

	cfg.SetDefaults()

	if cfg.TokenExpiration != 30*time.Minute {
		t.Errorf("TokenExpiration = %v, want %v (should be preserved)", cfg.TokenExpiration, 30*time.Minute)
	}
	if cfg.JWKCacheDuration != 2*time.Hour {
		t.Errorf("JWKCacheDuration = %v, want %v (should be preserved)", cfg.JWKCacheDuration, 2*time.Hour)
	}
}

func TestCognitoConfig_IsDevelopment(t *testing.T) {
	t.Run("with endpoint", func(t *testing.T) {
		cfg := &CognitoConfig{
			Endpoint: "http://cognito-local:9229",
		}
		if !cfg.IsDevelopment() {
			t.Error("IsDevelopment() should return true when Endpoint is set")
		}
	})

	t.Run("without endpoint", func(t *testing.T) {
		cfg := &CognitoConfig{
			Endpoint: "",
		}
		if cfg.IsDevelopment() {
			t.Error("IsDevelopment() should return false when Endpoint is empty")
		}
	})
}
