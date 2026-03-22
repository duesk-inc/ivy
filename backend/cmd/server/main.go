package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/duesk/ivy/internal/common/logger"
	"github.com/duesk/ivy/internal/config"
	"github.com/duesk/ivy/internal/handler"
	"github.com/duesk/ivy/internal/middleware"
	"github.com/duesk/ivy/internal/repository"
	"github.com/duesk/ivy/internal/routes"
	"github.com/duesk/ivy/internal/service"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func main() {
	// 1. 設定読み込み
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("設定読み込みエラー: %v\n", err)
		os.Exit(1)
	}

	// 2. ロガー初期化
	zapLogger, err := logger.InitLogger(cfg.IsProduction())
	if err != nil {
		fmt.Printf("ロガー初期化エラー: %v\n", err)
		os.Exit(1)
	}
	defer zapLogger.Sync()

	zapLogger.Info("Ivy backend starting",
		zap.String("env", cfg.AppEnv),
		zap.String("port", cfg.Server.Port),
	)

	// 3. データベース接続
	db, err := config.InitDatabase(cfg, zapLogger)
	if err != nil {
		zapLogger.Fatal("データベース接続失敗", zap.Error(err))
	}
	zapLogger.Info("データベース接続成功")

	// 4. Redis接続
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		zapLogger.Warn("Redis接続失敗（続行）", zap.Error(err))
	} else {
		zapLogger.Info("Redis接続成功")
	}

	// 5. リポジトリ作成（Phase 1）
	userRepo := repository.NewUserRepository(db, zapLogger)
	matchingRepo := repository.NewMatchingRepository(db, zapLogger)
	jobGroupRepo := repository.NewJobGroupRepository(db, zapLogger)
	settingsRepo := repository.NewSettingsRepository(db, zapLogger)

	// 5b. リポジトリ作成（Phase 2）
	jobRepo := repository.NewJobRepository(db, zapLogger)
	engineerProfileRepo := repository.NewEngineerProfileRepository(db, zapLogger)
	processedEmailRepo := repository.NewProcessedEmailRepository(db, zapLogger)
	gmailSyncStateRepo := repository.NewGmailSyncStateRepository(db, zapLogger)
	batchMatchingRepo := repository.NewBatchMatchingRepository(db, zapLogger)

	// 5c. 起動時リカバリ: staleなバッチをfailedに更新
	service.RecoverStaleRunning(context.Background(), batchMatchingRepo, zapLogger)

	// 6. AIサービス初期化
	var aiService service.AIService
	switch cfg.AI.AIMode {
	case "cli":
		zapLogger.Info("ClaudeCliAIService を使用（ローカル claude CLI）")
		systemPrompt := loadSystemPrompt()
		zapLogger.Info("システムプロンプト読み込み完了", zap.Int("length", len(systemPrompt)))
		aiService = service.NewClaudeCliAIService(systemPrompt, zapLogger)
	case "api":
		zapLogger.Info("ClaudeAIService を使用（API）")
		modelName := "claude-haiku-4-5-20251001"
		aiModelSetting, err := settingsRepo.GetByKey(context.Background(), "ai_model")
		if err == nil {
			parsedModel, parseErr := parseAIModelSetting(aiModelSetting.Value)
			if parseErr == nil {
				modelName = parsedModel
			}
		}
		systemPrompt := loadSystemPrompt()
		aiService = service.NewClaudeAIService(cfg, modelName, systemPrompt, zapLogger)
	default:
		zapLogger.Info("MockAIService を使用")
		aiService = service.NewMockAIService(zapLogger)
	}

	// 6b. EmailAIサービス初期化（Phase 2）
	var emailAIService service.EmailAIService
	if cfg.AI.AIMode == "api" {
		emailAIService = service.NewClaudeEmailAIService(cfg.AI.APIKey, "claude-haiku-4-5-20251001", zapLogger)
	} else {
		emailAIService = service.NewMockEmailAIService()
	}

	// 7. サービス初期化（Phase 1）
	authService := service.NewAuthService(cfg, userRepo, zapLogger)
	matchingService := service.NewMatchingService(matchingRepo, jobGroupRepo, settingsRepo, aiService, zapLogger)
	fileParseService := service.NewFileParseService(zapLogger)
	s3Service := service.NewMockS3Service(zapLogger)

	// 7b. サービス初期化（Phase 2）
	var gmailService service.GmailService
	if cfg.Gmail.Enabled {
		gmailService = service.NewGmailService(
			gmailSyncStateRepo, processedEmailRepo, emailAIService,
			jobRepo, engineerProfileRepo, settingsRepo,
			fileParseService, cfg, zapLogger,
		)
	} else {
		gmailService = service.NewMockGmailService(zapLogger)
	}

	jobService := service.NewJobService(jobRepo, zapLogger)
	engineerProfileService := service.NewEngineerProfileService(engineerProfileRepo, zapLogger)
	prefilterService := service.NewPrefilterService(zapLogger)
	batchMatchingService := service.NewBatchMatchingService(
		jobRepo, engineerProfileRepo, batchMatchingRepo,
		matchingService, prefilterService, zapLogger,
	)
	retentionService := service.NewRetentionService(
		jobRepo, engineerProfileRepo, processedEmailRepo,
		settingsRepo, zapLogger,
	)

	// 8. ハンドラー作成
	handlers := &routes.Handlers{
		Health:          handler.NewHealthHandler(db, redisClient, zapLogger),
		Auth:            handler.NewAuthHandler(authService, zapLogger),
		Matching:        handler.NewMatchingHandler(matchingService, zapLogger),
		JobGroup:        handler.NewJobGroupHandler(matchingService, zapLogger),
		File:            handler.NewFileHandler(fileParseService, s3Service, zapLogger),
		Settings:        handler.NewSettingsHandler(settingsRepo, zapLogger),
		Job:             handler.NewJobHandler(jobService, zapLogger),
		EngineerProfile: handler.NewEngineerProfileHandler(engineerProfileService, zapLogger),
		Email:           handler.NewEmailHandler(gmailService, zapLogger),
		BatchMatching:   handler.NewBatchMatchingHandler(batchMatchingService, zapLogger),
		Admin:           handler.NewAdminHandler(retentionService, zapLogger),
	}

	// 9. ルーター設定
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.RequestLogger(zapLogger))
	router.Use(middleware.AuditLog(zapLogger))
	router.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.Cors.AllowOrigins,
		AllowMethods:     cfg.Cors.AllowMethods,
		AllowHeaders:     cfg.Cors.AllowHeaders,
		ExposeHeaders:    cfg.Cors.ExposeHeaders,
		AllowCredentials: cfg.Cors.AllowCredentials,
		MaxAge:           cfg.Cors.MaxAge,
	}))
	router.Use(middleware.SecurityHeaders())
	router.Use(middleware.HSTSHeader())

	rateLimiter := middleware.NewInMemoryRateLimiter(zapLogger)
	router.Use(middleware.RateLimitMiddleware(rateLimiter, 100, time.Minute))

	cognitoMiddleware := middleware.NewCognitoAuthMiddleware(cfg, userRepo, zapLogger)

	routes.SetupRoutes(router, cognitoMiddleware, handlers)

	// 9b. データ保持バッチ（週1回チェック）
	go func() {
		ticker := time.NewTicker(24 * time.Hour) // 日次チェック
		defer ticker.Stop()
		for range ticker.C {
			ctx := context.Background()
			shouldRun, err := retentionService.ShouldRun(ctx)
			if err != nil {
				zapLogger.Warn("保持期間チェック失敗", zap.Error(err))
				continue
			}
			if shouldRun {
				zapLogger.Info("データクリーンアップ実行開始")
				if _, err := retentionService.RunCleanup(ctx); err != nil {
					zapLogger.Error("データクリーンアップ失敗", zap.Error(err))
				} else {
					_ = retentionService.RecordLastCleanup(ctx)
				}
			}
		}
	}()

	// 10. サーバー起動
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		zapLogger.Info("Server starting", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zapLogger.Fatal("サーバー起動失敗", zap.Error(err))
		}
	}()

	// グレースフルシャットダウン
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zapLogger.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		zapLogger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	zapLogger.Info("Server exited")
}

// parseAIModelSetting 設定値からモデル名を取得
func parseAIModelSetting(data []byte) (string, error) {
	var setting struct {
		Model string `json:"model"`
	}
	if err := json.Unmarshal(data, &setting); err != nil {
		return "", err
	}
	return setting.Model, nil
}

// loadSystemPrompt システムプロンプトを読み込む
func loadSystemPrompt() string {
	promptPaths := []string{
		"/app/matching_prompt.md",
		"./matching_prompt.md",
		"../matching_prompt.md",
		"/Users/daichirouesaka/Documents/duesk-company/products/matching-tool/matching_prompt.md",
	}
	for _, path := range promptPaths {
		data, err := os.ReadFile(path)
		if err == nil {
			return string(data)
		}
	}
	return defaultSystemPrompt
}

const defaultSystemPrompt = `あなたはSES業界専門のマッチング判定AIです。
案件情報とエンジニア情報を受け取り、マッチ度を100点満点で評価してください。
必ずJSON形式のみで出力してください。`
