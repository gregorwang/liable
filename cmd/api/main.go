package main

import (
	"comment-review-platform/internal/config"
	"comment-review-platform/internal/handlers"
	"comment-review-platform/internal/middleware"
	"comment-review-platform/internal/services"
	"comment-review-platform/pkg/database"
	redispkg "comment-review-platform/pkg/redis"
	"log"
	"time"

	"database/sql"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()
	log.Println("✅ Configuration loaded")

	// Initialize PostgreSQL
	db, err := database.InitPostgres(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("❌ Failed to connect to PostgreSQL: %v", err)
	}
	defer database.Close()

	// Initialize Redis
	_, err = redispkg.InitRedis(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB, cfg.RedisUseTLS)
	if err != nil {
		log.Fatalf("❌ Failed to connect to Redis: %v", err)
	}
	defer redispkg.Close()

	// Create database tables if they don't exist
	if err := createTables(db); err != nil {
		log.Fatalf("❌ Failed to create tables: %v", err)
	}

	// Initialize default data
	if err := initializeDefaultData(); err != nil {
		log.Printf("⚠️ Warning: Failed to initialize default data: %v", err)
	}

	// Start background task for releasing expired tasks
	go startTaskReleaseWorker()

	// Start daily sampling scheduler
	go startSamplingScheduler()

	// Setup Gin router
	router := setupRouter(db)

	// Start server
	log.Printf("🚀 Server starting on port %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}

func setupRouter(db interface{}) *gin.Engine {
	router := gin.Default()

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// SSE specific headers
		if c.Request.URL.Path == "/api/notifications/stream" {
			c.Writer.Header().Set("Cache-Control", "no-cache")
			c.Writer.Header().Set("Connection", "keep-alive")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	// Initialize handlers
	authHandler := handlers.NewAuthHandler()
	taskHandler := handlers.NewTaskHandler()
	secondReviewHandler := handlers.NewSecondReviewHandler()
	qualityCheckHandler := handlers.NewQualityCheckHandler()
	adminHandler := handlers.NewAdminHandler()

	// Initialize video handler
	videoHandler, err := handlers.NewVideoHandler()
	if err != nil {
		log.Printf("⚠️ Warning: Failed to initialize video handler: %v", err)
		// Create a nil handler to avoid nil pointer issues
		videoHandler = nil
	}

	// Type assert database connection
	sqlDB, ok := db.(*sql.DB)
	if !ok {
		panic("failed to assert database connection to *sql.DB")
	}
	moderationRulesHandler := handlers.NewModerationRulesHandler(sqlDB)

	// Initialize SSE manager and notification service
	sseManager := services.NewSSEManager()
	notificationService := services.NewNotificationService(sqlDB, sseManager)
	notificationHandler := handlers.NewNotificationHandler(notificationService)

	// API routes
	api := router.Group("/api")
	{
		// Auth routes (public)
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/send-code", authHandler.SendVerificationCode)
			auth.POST("/login-with-code", authHandler.LoginWithCode)
			auth.POST("/register-with-code", authHandler.RegisterWithCode)
			auth.GET("/check-email", authHandler.CheckEmail)
			auth.GET("/profile", middleware.AuthMiddleware(), authHandler.GetProfile)
		}

		// Moderation Rules routes (public - for viewing rules)
		modRules := api.Group("/moderation-rules")
		{
			modRules.GET("", moderationRulesHandler.ListRules)
			modRules.GET("/all", moderationRulesHandler.GetAllRules)
			modRules.GET("/:code", moderationRulesHandler.GetRuleByCode)
			modRules.GET("/categories", moderationRulesHandler.GetCategories)
			modRules.GET("/risk-levels", moderationRulesHandler.GetRiskLevels)
		}

		// Task routes (requires specific queue permissions)
		tasks := api.Group("/tasks")
		tasks.Use(middleware.AuthMiddleware())
		{
			// First review (comment review) routes
			tasks.POST("/claim", middleware.RequirePermission("tasks:first-review:claim"), taskHandler.ClaimTasks)
			tasks.GET("/my", middleware.RequireAnyPermission("tasks:first-review:claim", "tasks:second-review:claim", "tasks:quality-check:claim", "tasks:video-first-review:claim", "tasks:video-second-review:claim"), taskHandler.GetMyTasks)
			tasks.POST("/submit", middleware.RequirePermission("tasks:first-review:submit"), taskHandler.SubmitReview)
			tasks.POST("/submit-batch", middleware.RequirePermission("tasks:first-review:submit"), taskHandler.SubmitBatchReviews)
			tasks.POST("/return", middleware.RequirePermission("tasks:first-review:return"), taskHandler.ReturnTasks)

			// Second review routes
			tasks.POST("/second-review/claim", middleware.RequirePermission("tasks:second-review:claim"), secondReviewHandler.ClaimSecondReviewTasks)
			tasks.GET("/second-review/my", middleware.RequirePermission("tasks:second-review:claim"), secondReviewHandler.GetMySecondReviewTasks)
			tasks.POST("/second-review/submit", middleware.RequirePermission("tasks:second-review:submit"), secondReviewHandler.SubmitSecondReview)
			tasks.POST("/second-review/submit-batch", middleware.RequirePermission("tasks:second-review:submit"), secondReviewHandler.SubmitBatchSecondReviews)
			tasks.POST("/second-review/return", middleware.RequirePermission("tasks:second-review:return"), secondReviewHandler.ReturnSecondReviewTasks)

			// Quality check routes
			tasks.POST("/quality-check/claim", middleware.RequirePermission("tasks:quality-check:claim"), qualityCheckHandler.ClaimQCTasks)
			tasks.GET("/quality-check/my", middleware.RequirePermission("tasks:quality-check:claim"), qualityCheckHandler.GetMyQCTasks)
			tasks.POST("/quality-check/submit", middleware.RequirePermission("tasks:quality-check:submit"), qualityCheckHandler.SubmitQCReview)
			tasks.POST("/quality-check/submit-batch", middleware.RequirePermission("tasks:quality-check:submit"), qualityCheckHandler.SubmitBatchQCReviews)
			tasks.POST("/quality-check/return", middleware.RequirePermission("tasks:quality-check:return"), qualityCheckHandler.ReturnQCTasks)
			tasks.GET("/quality-check/stats", middleware.RequirePermission("tasks:quality-check:stats"), qualityCheckHandler.GetQCStats)

			// Video review routes (if video handler is available)
			if videoHandler != nil {
				// Video first review routes
				tasks.POST("/video-first-review/claim", middleware.RequirePermission("tasks:video-first-review:claim"), videoHandler.ClaimVideoFirstReviewTasks)
				tasks.GET("/video-first-review/my", middleware.RequirePermission("tasks:video-first-review:claim"), videoHandler.GetMyVideoFirstReviewTasks)
				tasks.POST("/video-first-review/submit", middleware.RequirePermission("tasks:video-first-review:submit"), videoHandler.SubmitVideoFirstReview)
				tasks.POST("/video-first-review/submit-batch", middleware.RequirePermission("tasks:video-first-review:submit"), videoHandler.SubmitBatchVideoFirstReviews)
				tasks.POST("/video-first-review/return", middleware.RequirePermission("tasks:video-first-review:return"), videoHandler.ReturnVideoFirstReviewTasks)

				// Video second review routes
				tasks.POST("/video-second-review/claim", middleware.RequirePermission("tasks:video-second-review:claim"), videoHandler.ClaimVideoSecondReviewTasks)
				tasks.GET("/video-second-review/my", middleware.RequirePermission("tasks:video-second-review:claim"), videoHandler.GetMyVideoSecondReviewTasks)
				tasks.POST("/video-second-review/submit", middleware.RequirePermission("tasks:video-second-review:submit"), videoHandler.SubmitVideoSecondReview)
				tasks.POST("/video-second-review/submit-batch", middleware.RequirePermission("tasks:video-second-review:submit"), videoHandler.SubmitBatchVideoSecondReviews)
				tasks.POST("/video-second-review/return", middleware.RequirePermission("tasks:video-second-review:return"), videoHandler.ReturnVideoSecondReviewTasks)
			}
		}

		// Search route (requires search permission)
		api.GET("/tasks/search", middleware.AuthMiddleware(), middleware.RequirePermission("tasks:search"), taskHandler.SearchTasks)

		// Tag routes (public for reviewers - requires any task permission)
		api.GET("/tags", middleware.AuthMiddleware(), middleware.RequireAnyPermission("tasks:first-review:claim", "tasks:second-review:claim", "tasks:quality-check:claim", "tasks:video-first-review:claim", "tasks:video-second-review:claim"), taskHandler.GetActiveTags)

		// Video quality tags route (if video handler is available)
		if videoHandler != nil {
			api.GET("/video-quality-tags", middleware.AuthMiddleware(), middleware.RequireAnyPermission("tasks:video-first-review:claim", "tasks:video-second-review:claim"), videoHandler.GetVideoQualityTags)
			// Video URL generation - available for users with video permissions
			api.POST("/admin/videos/generate-url", middleware.AuthMiddleware(), middleware.RequireAnyPermission("videos:generate-url", "tasks:video-first-review:claim", "tasks:video-second-review:claim"), videoHandler.GenerateVideoURL)
			// Test endpoint for data structure validation (no auth required)
			api.POST("/test/video-review-structure", videoHandler.TestVideoReviewDataStructure)
		}

		// Public Queue Read-Only API (no auth required)
		taskQueueHandler := handlers.NewTaskQueueHandler()
		api.GET("/queues", taskQueueHandler.GetPublicQueues)
		api.GET("/queues/:id", taskQueueHandler.GetPublicQueue)

		// Notification SSE endpoint (public, token in query param)
		api.GET("/notifications/stream", notificationHandler.StreamSSE)

		// Notification routes (requires authentication)
		notifications := api.Group("/notifications")
		notifications.Use(middleware.AuthMiddleware())
		{
			notifications.GET("/unread", notificationHandler.GetUnread)
			notifications.GET("/unread-count", notificationHandler.GetUnreadCount)
			notifications.PUT("/:id/read", notificationHandler.MarkAsRead)
			notifications.GET("/recent", notificationHandler.GetRecent)
		}

		// Admin routes (requires admin role)
		admin := api.Group("/admin")
		admin.Use(middleware.AuthMiddleware(), middleware.RequireAdmin())
		{
			// Permission management (requires permission management permissions)
			// Note: More specific routes must be registered before generic ones
			admin.GET("/permissions/all", middleware.RequirePermission("permissions:read"), adminHandler.GetAllPermissions)
			admin.GET("/permissions/user", middleware.RequirePermission("permissions:read"), adminHandler.GetUserPermissions)
			admin.GET("/permissions", middleware.RequirePermission("permissions:read"), adminHandler.ListPermissions)
			admin.POST("/permissions/grant", middleware.RequirePermission("permissions:grant"), adminHandler.GrantPermissions)
			admin.POST("/permissions/revoke", middleware.RequirePermission("permissions:revoke"), adminHandler.RevokePermissions)

			// User management
			admin.GET("/users", middleware.RequirePermission("users:list"), adminHandler.GetPendingUsers)
			admin.GET("/users/all", middleware.RequirePermission("users:list"), adminHandler.GetAllUsers)
			admin.PUT("/users/:id/approve", middleware.RequirePermission("users:approve"), adminHandler.ApproveUser)

			// Statistics
			admin.GET("/stats/overview", middleware.RequirePermission("stats:overview"), adminHandler.GetOverviewStats)
			admin.GET("/stats/hourly", middleware.RequirePermission("stats:hourly"), adminHandler.GetHourlyStats)
			admin.GET("/stats/tags", middleware.RequirePermission("stats:tags"), adminHandler.GetTagStats)
			admin.GET("/stats/reviewers", middleware.RequirePermission("stats:reviewers"), adminHandler.GetReviewerPerformance)

			// Tag management
			admin.GET("/tags", middleware.RequirePermission("tags:list"), adminHandler.GetAllTags)
			admin.POST("/tags", middleware.RequirePermission("tags:create"), adminHandler.CreateTag)
			admin.PUT("/tags/:id", middleware.RequirePermission("tags:update"), adminHandler.UpdateTag)
			admin.DELETE("/tags/:id", middleware.RequirePermission("tags:delete"), adminHandler.DeleteTag)

			// Moderation Rules management
			admin.POST("/moderation-rules", middleware.RequirePermission("moderation-rules:create"), moderationRulesHandler.CreateRule)
			admin.PUT("/moderation-rules/:id", middleware.RequirePermission("moderation-rules:update"), moderationRulesHandler.UpdateRule)
			admin.DELETE("/moderation-rules/:id", middleware.RequirePermission("moderation-rules:delete"), moderationRulesHandler.DeleteRule)

			// Task Queue management (队列配置)
			admin.POST("/task-queues", middleware.RequirePermission("task-queues:create"), taskQueueHandler.CreateTaskQueue)
			admin.GET("/task-queues", middleware.RequirePermission("task-queues:list"), taskQueueHandler.ListTaskQueues)
			admin.GET("/task-queues/:id", middleware.RequirePermission("task-queues:list"), taskQueueHandler.GetTaskQueue)
			admin.PUT("/task-queues/:id", middleware.RequirePermission("task-queues:update"), taskQueueHandler.UpdateTaskQueue)
			admin.DELETE("/task-queues/:id", middleware.RequirePermission("task-queues:delete"), taskQueueHandler.DeleteTaskQueue)
			admin.GET("/task-queues-all", middleware.RequirePermission("task-queues:list"), taskQueueHandler.GetAllTaskQueues)

			// Notification management (admin only)
			admin.POST("/notifications", middleware.RequirePermission("notifications:create"), notificationHandler.CreateNotification)

			// Video management (if video handler is available)
			if videoHandler != nil {
				admin.POST("/videos/import", middleware.RequirePermission("videos:import"), videoHandler.ImportVideos)
				admin.GET("/videos", middleware.RequirePermission("videos:list"), videoHandler.ListVideos)
				admin.GET("/videos/:id", middleware.RequirePermission("videos:read"), videoHandler.GetVideo)
			}
		}
	}

	return router
}

func createTables(db interface{}) error {
	// This will be called via MCP
	log.Println("ℹ️  Database tables should be created via Supabase MCP")
	return nil
}

func initializeDefaultData() error {
	// This will be called via MCP
	log.Println("ℹ️  Default data should be initialized via Supabase MCP")
	return nil
}

func startSamplingScheduler() {
	samplingService := services.NewSamplingService()
	log.Println("✅ Daily sampling scheduler started")
	samplingService.StartDailySamplingScheduler()
}

func startTaskReleaseWorker() {
	taskService := services.NewTaskService()
	secondReviewService := services.NewSecondReviewService()
	qcService := services.NewQualityCheckService()
	videoFirstReviewService := services.NewVideoFirstReviewService()
	videoSecondReviewService := services.NewVideoSecondReviewService()
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	log.Println("✅ Background task release worker started")

	for range ticker.C {
		if err := taskService.ReleaseExpiredTasks(); err != nil {
			log.Printf("⚠️ Error releasing expired tasks: %v", err)
		}
		if err := secondReviewService.ReleaseExpiredSecondReviewTasks(); err != nil {
			log.Printf("⚠️ Error releasing expired second review tasks: %v", err)
		}
		if err := qcService.ReleaseExpiredQCTasks(); err != nil {
			log.Printf("⚠️ Error releasing expired QC tasks: %v", err)
		}
		if err := videoFirstReviewService.ReleaseExpiredFirstReviewTasks(); err != nil {
			log.Printf("⚠️ Error releasing expired video first review tasks: %v", err)
		}
		if err := videoSecondReviewService.ReleaseExpiredSecondReviewTasks(); err != nil {
			log.Printf("⚠️ Error releasing expired video second review tasks: %v", err)
		}
	}
}
