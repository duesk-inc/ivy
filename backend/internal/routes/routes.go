package routes

import (
	"github.com/duesk/ivy/internal/handler"
	"github.com/duesk/ivy/internal/middleware"
	"github.com/gin-gonic/gin"
)

// Handlers 全ハンドラーをまとめる構造体
type Handlers struct {
	Health          *handler.HealthHandler
	Auth            *handler.AuthHandler
	Matching        *handler.MatchingHandler
	File            *handler.FileHandler
	Settings        *handler.SettingsHandler
	JobGroup        *handler.JobGroupHandler
	Job             *handler.JobHandler
	EngineerProfile *handler.EngineerProfileHandler
	Email           *handler.EmailHandler
	BatchMatching   *handler.BatchMatchingHandler
	Admin           *handler.AdminHandler
}

// SetupRoutes ルーティングを設定
func SetupRoutes(
	router *gin.Engine,
	authMiddleware *middleware.CognitoAuthMiddleware,
	h *Handlers,
) {
	// ヘルスチェック（認証不要）
	router.GET("/health", h.Health.HealthCheck)

	api := router.Group("/api/v1")

	// 認証エンドポイント（認証不要）
	auth := api.Group("/auth")
	{
		auth.POST("/login", h.Auth.Login)
		auth.POST("/refresh", h.Auth.Refresh)
		auth.POST("/logout", h.Auth.Logout)
	}

	// 認証が必要なエンドポイント
	authorized := api.Group("")
	authorized.Use(authMiddleware.AuthRequired())
	{
		// ユーザー情報
		authorized.GET("/me", h.Auth.Me)

		// マッチング
		matchings := authorized.Group("/matchings")
		{
			matchings.POST("", h.Matching.Execute)
			matchings.GET("", h.Matching.List)
			matchings.GET("/:id", h.Matching.GetByID)
			matchings.DELETE("/:id", h.Matching.Delete)
			matchings.PUT("/:id/job-group", h.Matching.LinkToJobGroup)
			matchings.DELETE("/:id/job-group", h.Matching.UnlinkFromJobGroup)
		}

		// バッチマッチング（Phase 2）
		if h.BatchMatching != nil {
			batchMatchings := authorized.Group("/matchings/batch")
			{
				batchMatchings.POST("/preview", h.BatchMatching.Preview)
				batchMatchings.POST("", h.BatchMatching.Execute)
				batchMatchings.GET("/:id", h.BatchMatching.GetStatus)
			}
		}

		// 案件グループ
		jobGroups := authorized.Group("/job-groups")
		{
			jobGroups.POST("", h.JobGroup.Create)
			jobGroups.GET("", h.JobGroup.List)
			jobGroups.GET("/:id", h.JobGroup.Get)
			jobGroups.DELETE("/:id", h.JobGroup.Delete)
		}

		// 案件一覧 + 1:Nマッチング（Phase 2）
		if h.Job != nil {
			jobs := authorized.Group("/jobs")
			{
				jobs.GET("", h.Job.List)
				if h.BatchMatching != nil {
					jobs.POST("/:id/match-engineers", h.BatchMatching.MatchJobToEngineers)
				}
			}
		}

		// 人材一覧 + 1:Nマッチング（Phase 2）
		if h.EngineerProfile != nil {
			engineers := authorized.Group("/engineers")
			{
				engineers.GET("", h.EngineerProfile.List)
				if h.BatchMatching != nil {
					engineers.POST("/:id/match-jobs", h.BatchMatching.MatchEngineerToJobs)
				}
			}
		}

		// ファイル
		files := authorized.Group("/files")
		{
			files.POST("/parse", h.File.Parse)
		}

		// メール同期（Phase 2、admin限定）
		if h.Email != nil {
			emails := authorized.Group("/emails")
			{
				emails.POST("/sync", middleware.AdminRequired(), h.Email.Sync)
				emails.GET("/sync/state", h.Email.GetSyncState)
			}
		}

		// 設定（閲覧: admin/sales、更新: adminのみ）
		settings := authorized.Group("/settings")
		{
			settings.GET("", h.Settings.GetAll)
			settings.PUT("/:key", middleware.AdminRequired(), h.Settings.Update)
		}

		// 管理者エンドポイント（Phase 2）
		if h.Admin != nil {
			admin := authorized.Group("/admin")
			admin.Use(middleware.AdminRequired())
			{
				admin.POST("/retention/run", h.Admin.RunRetention)
			}
		}
	}
}
