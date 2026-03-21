package config

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestLoad_Defaults(t *testing.T) {
	os.Clearenv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// AppEnv
	if cfg.AppEnv != "development" {
		t.Errorf("AppEnv = %q, want %q", cfg.AppEnv, "development")
	}

	// Server
	if cfg.Server.Port != "8081" {
		t.Errorf("Server.Port = %q, want %q", cfg.Server.Port, "8081")
	}
	if cfg.Server.ReadTimeout != 30*time.Second {
		t.Errorf("Server.ReadTimeout = %v, want %v", cfg.Server.ReadTimeout, 30*time.Second)
	}
	if cfg.Server.WriteTimeout != 120*time.Second {
		t.Errorf("Server.WriteTimeout = %v, want %v", cfg.Server.WriteTimeout, 120*time.Second)
	}

	// Database
	if cfg.Database.Host != "localhost" {
		t.Errorf("Database.Host = %q, want %q", cfg.Database.Host, "localhost")
	}
	if cfg.Database.Port != "5432" {
		t.Errorf("Database.Port = %q, want %q", cfg.Database.Port, "5432")
	}
	if cfg.Database.User != "ivy_user" {
		t.Errorf("Database.User = %q, want %q", cfg.Database.User, "ivy_user")
	}
	if cfg.Database.Password != "ivy_password" {
		t.Errorf("Database.Password = %q, want %q", cfg.Database.Password, "ivy_password")
	}
	if cfg.Database.DBName != "ivy" {
		t.Errorf("Database.DBName = %q, want %q", cfg.Database.DBName, "ivy")
	}
	if cfg.Database.SSLMode != "disable" {
		t.Errorf("Database.SSLMode = %q, want %q", cfg.Database.SSLMode, "disable")
	}

	// AI
	if cfg.AI.APIKey != "" {
		t.Errorf("AI.APIKey = %q, want empty", cfg.AI.APIKey)
	}
	if cfg.AI.AIMode != "mock" {
		t.Errorf("AI.AIMode = %q, want mock", cfg.AI.AIMode)
	}
	if cfg.AI.Timeout != 60*time.Second {
		t.Errorf("AI.Timeout = %v, want %v", cfg.AI.Timeout, 60*time.Second)
	}

	// Redis
	if cfg.Redis.Host != "localhost" {
		t.Errorf("Redis.Host = %q, want %q", cfg.Redis.Host, "localhost")
	}
	if cfg.Redis.Port != "6379" {
		t.Errorf("Redis.Port = %q, want %q", cfg.Redis.Port, "6379")
	}
	if cfg.Redis.Password != "" {
		t.Errorf("Redis.Password = %q, want empty", cfg.Redis.Password)
	}
	if cfg.Redis.DB != 0 {
		t.Errorf("Redis.DB = %d, want 0", cfg.Redis.DB)
	}

	// CORS
	if len(cfg.Cors.AllowOrigins) != 1 || cfg.Cors.AllowOrigins[0] != "http://localhost:5173" {
		t.Errorf("Cors.AllowOrigins = %v, want [http://localhost:5173]", cfg.Cors.AllowOrigins)
	}
	if cfg.Cors.AllowCredentials != true {
		t.Error("Cors.AllowCredentials should default to true")
	}
	if cfg.Cors.MaxAge != 300*time.Second {
		t.Errorf("Cors.MaxAge = %v, want %v", cfg.Cors.MaxAge, 300*time.Second)
	}

	// S3
	if cfg.S3.BucketName != "ivy-files-dev" {
		t.Errorf("S3.BucketName = %q, want %q", cfg.S3.BucketName, "ivy-files-dev")
	}
	if cfg.S3.Region != "ap-northeast-1" {
		t.Errorf("S3.Region = %q, want %q", cfg.S3.Region, "ap-northeast-1")
	}
	if cfg.S3.UseMockS3 != true {
		t.Error("S3.UseMockS3 should default to true")
	}

	// Cognito
	if cfg.Cognito.Enabled != false {
		t.Error("Cognito.Enabled should default to false")
	}
	if cfg.Cognito.Region != "ap-northeast-1" {
		t.Errorf("Cognito.Region = %q, want %q", cfg.Cognito.Region, "ap-northeast-1")
	}
}

func TestLoad_EnvOverride(t *testing.T) {
	os.Clearenv()
	t.Setenv("APP_ENV", "production")
	t.Setenv("PORT", "9090")
	t.Setenv("DB_HOST", "db.example.com")
	t.Setenv("DB_PORT", "5433")
	t.Setenv("DB_USER", "prod_user")
	t.Setenv("DB_PASSWORD", "prod_pass")
	t.Setenv("DB_NAME", "ivy_prod")
	t.Setenv("DB_SSLMODE", "require")
	t.Setenv("USE_MOCK_AI", "false")
	t.Setenv("ANTHROPIC_API_KEY", "sk-test-key")
	t.Setenv("REDIS_HOST", "redis.example.com")
	t.Setenv("REDIS_PORT", "6380")
	t.Setenv("REDIS_PASSWORD", "redis_secret")
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://app.example.com")
	t.Setenv("COGNITO_ENABLED", "true")
	t.Setenv("COGNITO_USER_POOL_ID", "ap-northeast-1_TestPool")
	t.Setenv("COGNITO_CLIENT_ID", "test-client-id")
	t.Setenv("AWS_S3_BUCKET", "ivy-files-prod")
	t.Setenv("USE_MOCK_S3", "false")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.AppEnv != "production" {
		t.Errorf("AppEnv = %q, want %q", cfg.AppEnv, "production")
	}
	if cfg.Server.Port != "9090" {
		t.Errorf("Server.Port = %q, want %q", cfg.Server.Port, "9090")
	}
	if cfg.Database.Host != "db.example.com" {
		t.Errorf("Database.Host = %q, want %q", cfg.Database.Host, "db.example.com")
	}
	if cfg.Database.Port != "5433" {
		t.Errorf("Database.Port = %q, want %q", cfg.Database.Port, "5433")
	}
	if cfg.Database.User != "prod_user" {
		t.Errorf("Database.User = %q, want %q", cfg.Database.User, "prod_user")
	}
	if cfg.Database.Password != "prod_pass" {
		t.Errorf("Database.Password = %q, want %q", cfg.Database.Password, "prod_pass")
	}
	if cfg.Database.DBName != "ivy_prod" {
		t.Errorf("Database.DBName = %q, want %q", cfg.Database.DBName, "ivy_prod")
	}
	if cfg.Database.SSLMode != "require" {
		t.Errorf("Database.SSLMode = %q, want %q", cfg.Database.SSLMode, "require")
	}
	if cfg.AI.AIMode != "api" {
		t.Errorf("AI.AIMode = %q, want api", cfg.AI.AIMode)
	}
	if cfg.AI.APIKey != "sk-test-key" {
		t.Errorf("AI.APIKey = %q, want %q", cfg.AI.APIKey, "sk-test-key")
	}
	if cfg.Redis.Host != "redis.example.com" {
		t.Errorf("Redis.Host = %q, want %q", cfg.Redis.Host, "redis.example.com")
	}
	if cfg.Redis.Port != "6380" {
		t.Errorf("Redis.Port = %q, want %q", cfg.Redis.Port, "6380")
	}
	if cfg.Redis.Password != "redis_secret" {
		t.Errorf("Redis.Password = %q, want %q", cfg.Redis.Password, "redis_secret")
	}
	if len(cfg.Cors.AllowOrigins) != 1 || cfg.Cors.AllowOrigins[0] != "https://app.example.com" {
		t.Errorf("Cors.AllowOrigins = %v, want [https://app.example.com]", cfg.Cors.AllowOrigins)
	}
	if cfg.Cognito.Enabled != true {
		t.Error("Cognito.Enabled should be true")
	}
	if cfg.Cognito.UserPoolID != "ap-northeast-1_TestPool" {
		t.Errorf("Cognito.UserPoolID = %q, want %q", cfg.Cognito.UserPoolID, "ap-northeast-1_TestPool")
	}
	if cfg.Cognito.ClientID != "test-client-id" {
		t.Errorf("Cognito.ClientID = %q, want %q", cfg.Cognito.ClientID, "test-client-id")
	}
	if cfg.S3.BucketName != "ivy-files-prod" {
		t.Errorf("S3.BucketName = %q, want %q", cfg.S3.BucketName, "ivy-files-prod")
	}
	if cfg.S3.UseMockS3 != false {
		t.Error("S3.UseMockS3 should be false")
	}
}

func TestLoad_WithEnvFile(t *testing.T) {
	os.Clearenv()

	_, err := Load("/nonexistent/path/.env")
	if err == nil {
		t.Error("Load with invalid env file path should return an error")
	}
	if !strings.Contains(err.Error(), "error loading .env file") {
		t.Errorf("error message = %q, want it to contain %q", err.Error(), "error loading .env file")
	}
}

func TestConfig_IsProduction(t *testing.T) {
	tests := []struct {
		name     string
		env      string
		expected bool
	}{
		{"production lowercase", "production", true},
		{"Production capitalized", "Production", true},
		{"development", "development", false},
		{"staging", "staging", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{AppEnv: tt.env}
			if got := cfg.IsProduction(); got != tt.expected {
				t.Errorf("IsProduction() for %q = %v, want %v", tt.env, got, tt.expected)
			}
		})
	}
}

func TestDatabaseConfig_GetPostgreSQLDSN(t *testing.T) {
	dbCfg := &DatabaseConfig{
		Host:     "db.example.com",
		Port:     "5433",
		User:     "test_user",
		Password: "test_pass",
		DBName:   "test_db",
		SSLMode:  "require",
	}

	dsn := dbCfg.GetPostgreSQLDSN()

	requiredParams := []string{
		"host=db.example.com",
		"port=5433",
		"user=test_user",
		"password=test_pass",
		"dbname=test_db",
		"sslmode=require",
		"timezone=Asia/Tokyo",
	}

	for _, param := range requiredParams {
		if !strings.Contains(dsn, param) {
			t.Errorf("DSN missing %q: got %q", param, dsn)
		}
	}
}

func TestDatabaseConfig_GetPostgreSQLDSN_EmptySSLMode(t *testing.T) {
	dbCfg := &DatabaseConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "user",
		Password: "pass",
		DBName:   "db",
		SSLMode:  "",
	}

	dsn := dbCfg.GetPostgreSQLDSN()

	if !strings.Contains(dsn, "sslmode=disable") {
		t.Errorf("DSN should default to sslmode=disable when SSLMode is empty: got %q", dsn)
	}
}

func TestLoad_CORSMultipleOrigins(t *testing.T) {
	os.Clearenv()
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:5173,https://app.example.com,https://admin.example.com")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	expectedOrigins := []string{
		"http://localhost:5173",
		"https://app.example.com",
		"https://admin.example.com",
	}

	if len(cfg.Cors.AllowOrigins) != len(expectedOrigins) {
		t.Fatalf("Cors.AllowOrigins length = %d, want %d", len(cfg.Cors.AllowOrigins), len(expectedOrigins))
	}

	for i, expected := range expectedOrigins {
		if cfg.Cors.AllowOrigins[i] != expected {
			t.Errorf("Cors.AllowOrigins[%d] = %q, want %q", i, cfg.Cors.AllowOrigins[i], expected)
		}
	}
}

func TestLoad_AITimeout(t *testing.T) {
	t.Run("default 60s", func(t *testing.T) {
		os.Clearenv()

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		if cfg.AI.Timeout != 60*time.Second {
			t.Errorf("AI.Timeout = %v, want %v", cfg.AI.Timeout, 60*time.Second)
		}
	})

	t.Run("custom override 120s", func(t *testing.T) {
		os.Clearenv()
		t.Setenv("AI_TIMEOUT_SECONDS", "120")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		if cfg.AI.Timeout != 120*time.Second {
			t.Errorf("AI.Timeout = %v, want %v", cfg.AI.Timeout, 120*time.Second)
		}
	})
}
