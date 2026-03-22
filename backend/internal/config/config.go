package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// GmailConfig Gmail API設定
type GmailConfig struct {
	Enabled               bool
	ServiceAccountKeyFile string
	TargetEmail           string
}

// Config アプリケーション全体の設定
type Config struct {
	AppEnv   string
	Server   ServerConfig
	Database DatabaseConfig
	Cors     CorsConfig
	Redis    RedisConfig
	Cognito  CognitoConfig
	AI       AIConfig
	S3       S3Config
	Gmail    GmailConfig
}

// IsProduction 本番環境かどうかを判定
func (c *Config) IsProduction() bool {
	return strings.ToLower(c.AppEnv) == "production"
}

// ServerConfig サーバー設定
type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// DatabaseConfig データベース接続設定
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// GetPostgreSQLDSN PostgreSQL接続用のDSN文字列を生成
func (c *DatabaseConfig) GetPostgreSQLDSN() string {
	sslMode := c.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s timezone=Asia/Tokyo",
		c.Host, c.Port, c.User, c.Password, c.DBName, sslMode)
}

// CorsConfig CORS設定
type CorsConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           time.Duration
}

// RedisConfig Redis設定
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// AIConfig Claude API設定
type AIConfig struct {
	APIKey  string
	AIMode  string // "mock", "cli", "api"
	Timeout time.Duration
}

// S3Config S3設定
type S3Config struct {
	BucketName string
	Region     string
	UseMockS3  bool
}

// Load 設定を読み込む
func Load(envFile ...string) (*Config, error) {
	if len(envFile) > 0 && envFile[0] != "" {
		if err := godotenv.Load(envFile[0]); err != nil {
			return nil, fmt.Errorf("error loading .env file: %w", err)
		}
	} else {
		// 引数なしの場合、.envファイルを自動検索して読み込む（存在しなければスキップ）
		_ = godotenv.Load("../.env", ".env")
	}

	readTimeout, _ := strconv.Atoi(getEnv("SERVER_READ_TIMEOUT", "30"))
	writeTimeout, _ := strconv.Atoi(getEnv("SERVER_WRITE_TIMEOUT", "300"))
	corsMaxAge, _ := strconv.Atoi(getEnv("CORS_MAX_AGE", "300"))
	aiTimeout, _ := strconv.Atoi(getEnv("AI_TIMEOUT_SECONDS", "60"))

	config := &Config{
		AppEnv: getEnv("APP_ENV", "development"),

		Server: ServerConfig{
			Port:         getEnv("PORT", "8081"),
			ReadTimeout:  time.Duration(readTimeout) * time.Second,
			WriteTimeout: time.Duration(writeTimeout) * time.Second,
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "ivy_user"),
			Password: getEnv("DB_PASSWORD", "ivy_password"),
			DBName:   getEnv("DB_NAME", "ivy"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Cors: CorsConfig{
			AllowOrigins:     strings.Split(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173"), ","),
			AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
			MaxAge:           time.Duration(corsMaxAge) * time.Second,
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       0,
		},
		Cognito: CognitoConfig{
			Enabled:      getEnv("COGNITO_ENABLED", "false") == "true",
			Region:       getEnv("COGNITO_REGION", "ap-northeast-1"),
			UserPoolID:   getEnv("COGNITO_USER_POOL_ID", ""),
			ClientID:     getEnv("COGNITO_CLIENT_ID", ""),
			ClientSecret: getEnv("COGNITO_CLIENT_SECRET", ""),
			Endpoint:     getEnv("COGNITO_ENDPOINT", ""),
		},
		AI: AIConfig{
			APIKey:    getEnv("ANTHROPIC_API_KEY", ""),
			AIMode:  parseAIMode(getEnv("USE_MOCK_AI", "true")),
			Timeout:   time.Duration(aiTimeout) * time.Second,
		},
		S3: S3Config{
			BucketName: getEnv("AWS_S3_BUCKET", "ivy-files-dev"),
			Region:     getEnv("AWS_REGION", "ap-northeast-1"),
			UseMockS3:  getEnv("USE_MOCK_S3", "true") == "true",
		},
		Gmail: GmailConfig{
			Enabled:               getEnv("GMAIL_ENABLED", "false") == "true",
			ServiceAccountKeyFile: getEnv("GMAIL_SERVICE_ACCOUNT_KEY_FILE", ""),
			TargetEmail:           getEnv("GMAIL_TARGET_EMAIL", ""),
		},
	}

	config.Cognito.SetDefaults()

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// parseAIMode USE_MOCK_AI環境変数を解析
// "true" → "mock", "false" → "api", "cli" → "cli"
func parseAIMode(val string) string {
	switch strings.ToLower(val) {
	case "false":
		return "api"
	case "cli":
		return "cli"
	default:
		return "mock"
	}
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
