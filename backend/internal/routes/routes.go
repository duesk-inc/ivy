package routes

import (
	"github.com/duesk/ivy/internal/handler"
	"github.com/duesk/ivy/internal/middleware"
	"github.com/gin-gonic/gin"
)

// SetupRoutes ルーティングを設定
func SetupRoutes(
	router *gin.Engine,
	authMiddleware *middleware.CognitoAuthMiddleware,
	healthHandler *handler.HealthHandler,
	authHandler *handler.AuthHandler,
	matchingHandler *handler.MatchingHandler,
	fileHandler *handler.FileHandler,
	settingsHandler *handler.SettingsHandler,
	jobGroupHandler *handler.JobGroupHandler,
) {
	// ヘルスチェック（認証不要）
	router.GET("/health", healthHandler.HealthCheck)

	api := router.Group("/api/v1")

	// 認証エンドポイント（認証不要）
	auth := api.Group("/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.Refresh)
		auth.POST("/logout", authHandler.Logout)
	}

	// 認証が必要なエンドポイント
	authorized := api.Group("")
	authorized.Use(authMiddleware.AuthRequired())
	{
		// ユーザー情報
		authorized.GET("/me", authHandler.Me)

		// マッチング
		matchings := authorized.Group("/matchings")
		{
			matchings.POST("", matchingHandler.Execute)
			matchings.GET("", matchingHandler.List)
			matchings.GET("/:id", matchingHandler.GetByID)
			matchings.DELETE("/:id", matchingHandler.Delete)
			matchings.PUT("/:id/job-group", matchingHandler.LinkToJobGroup)
			matchings.DELETE("/:id/job-group", matchingHandler.UnlinkFromJobGroup)
		}

		// 案件グループ
		jobGroups := authorized.Group("/job-groups")
		{
			jobGroups.POST("", jobGroupHandler.Create)
			jobGroups.GET("", jobGroupHandler.List)
			jobGroups.GET("/:id", jobGroupHandler.Get)
			jobGroups.DELETE("/:id", jobGroupHandler.Delete)
		}

		// ファイル
		files := authorized.Group("/files")
		{
			files.POST("/parse", fileHandler.Parse)
		}

		// 設定（閲覧: admin/sales、更新: adminのみ）
		settings := authorized.Group("/settings")
		{
			settings.GET("", settingsHandler.GetAll)
			settings.PUT("/:key", middleware.AdminRequired(), settingsHandler.Update)
		}
	}
}
