package main

import (
	"comment-review-platform/internal/config"
	"comment-review-platform/internal/handlers"
	"comment-review-platform/internal/middleware"
	"comment-review-platform/internal/services"
	"comment-review-platform/pkg/database"
	redispkg "comment-review-platform/pkg/redis"
	"database/sql"
	"log"
	"time"

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
	_, err = redispkg.InitRedis(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB, cfg.RedisUseTLS, cfg.RedisTLSSkipVerify)
	if err != nil {
		log.Fatalf("❌ Failed to connect to Redis: %v", err)
	}
	defer redispkg.Close()

	// Initialize audit logger (requires DB connection)
	middleware.InitAuditLogger(db)

	// Initialize alerting and metrics services
	alertService := services.NewAlertService(redispkg.Client)
	middleware.InitAlertService(alertService)
	metricsService := services.NewMetricsService(redispkg.Client, cfg.MetricsWindowMinutes)
	middleware.InitMetricsService(metricsService)

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

	// Start daily statistics aggregation scheduler
	go startDailyStatsAggregator()

	// Start AI review scheduler
	go startAIReviewScheduler()

	// Setup Gin router
	router := setupRouter(db, metricsService)

	// Start audit log cleanup (retention: 90 days, runs daily)
	go middleware.StartAuditLogCleanup(90, 24*time.Hour)

	// Start server
	log.Printf("🚀 Server starting on port %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}

func setupRouter(db interface{}, metricsService *services.MetricsService) *gin.Engine {
	router := gin.New()

	// Global middleware (executed in order)
	router.Use(gin.Logger())
	router.Use(middleware.RecoveryMiddleware())
	router.Use(middleware.TraceMiddleware())

	// Global rate limiting (first line of defense - 100 req/sec per IP)
	router.Use(middleware.GlobalRateLimiterV2())

	// Metrics & audit logging (async, non-blocking)
	router.Use(middleware.MetricsMiddleware())
	router.Use(middleware.AuditLogMiddleware())

	// 3. CORS middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Trace-Id, X-Page-Url")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "X-Trace-Id, X-Request-Id")

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

	// Type assert database connection
	sqlDB, ok := db.(*sql.DB)
	if !ok {
		panic("failed to assert database connection to *sql.DB")
	}

	// Health check
	monitoringHandler := handlers.NewMonitoringHandler(sqlDB, redispkg.Client, metricsService)
	router.GET("/health", monitoringHandler.Health)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler()
	taskHandler := handlers.NewTaskHandler()
	secondReviewHandler := handlers.NewSecondReviewHandler()
	qualityCheckHandler := handlers.NewQualityCheckHandler()
	aiHumanDiffHandler := handlers.NewAIHumanDiffHandler()
	adminHandler := handlers.NewAdminHandler()
	aiReviewHandler := handlers.NewAIReviewHandler()
	auditLogHandler := handlers.NewAuditLogHandler()
	documentHandler := handlers.NewDocumentHandler()
	bugReportHandler := handlers.NewBugReportHandler()

	// Initialize video handler
	videoHandler, err := handlers.NewVideoHandler()
	if err != nil {
		log.Printf("⚠️ Warning: Failed to initialize video handler: %v", err)
		// Create a nil handler to avoid nil pointer issues
		videoHandler = nil
	}

	// Initialize video queue handler
	videoQueueHandler := handlers.NewVideoQueueHandler()

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
			// Rate limit: 3 attempts per 5.5 minutes for registration
			auth.POST("/register", middleware.EndpointRateLimiterV2(3, 5*time.Minute+30*time.Second), authHandler.Register)
			// Rate limit: 5 attempts per 5 minutes for login
			auth.POST("/login", middleware.EndpointRateLimiterV2(5, 5*time.Minute), authHandler.Login)
			// Rate limit: 10 attempts per hour for send-code
			auth.POST("/send-code", middleware.EndpointRateLimiterV2(10, 1*time.Hour), authHandler.SendVerificationCode)
			// Rate limit: 5 attempts per 5 minutes for
			auth.POST("/login-with-code", middleware.EndpointRateLimiterV2(5, 5*time.Minute), authHandler.LoginWithCode)
			// Rate limit: 3 attempts per 5.5 minutes for registration
			auth.POST("/register-with-code", middleware.EndpointRateLimiterV2(3, 5*time.Minute+30*time.Second), authHandler.RegisterWithCode)
			// Rate limit: 10 per minute for email checking
			auth.GET("/check-email", middleware.EndpointRateLimiterV2(10, 1*time.Minute), authHandler.CheckEmail)
			auth.GET("/profile", middleware.AuthMiddleware(), authHandler.GetProfile)
			auth.PUT("/profile", middleware.AuthMiddleware(), authHandler.UpdateProfile)
			auth.PUT("/profile/system", middleware.AuthMiddleware(), middleware.RequirePermission("users:profile:update"), authHandler.UpdateSystemProfile)
			auth.POST("/profile/avatar", middleware.AuthMiddleware(), authHandler.UploadAvatar)
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

		// Shared statistics routes (accessible to authenticated users)
		api.GET("/stats/today", middleware.AuthMiddleware(), adminHandler.GetTodayReviewStats)

		// System documents (requires authentication)
		api.GET("/docs", middleware.AuthMiddleware(), documentHandler.ListDocuments)

		// Bug report submission (requires authentication)
		api.POST("/bug-reports", middleware.AuthMiddleware(), bugReportHandler.Create)

		// Task routes (requires specific queue permissions)
		tasks := api.Group("/tasks")
		tasks.Use(middleware.AuthMiddleware())
		{
			// First review (comment review) routes
			tasks.POST("/claim", middleware.UserRateLimiterV2(10, time.Minute), middleware.RequirePermission("tasks:first-review:claim"), taskHandler.ClaimTasks)
			tasks.GET("/my", middleware.RequireAnyPermission("tasks:first-review:claim", "tasks:second-review:claim", "tasks:quality-check:claim", "tasks:video-first-review:claim", "tasks:video-second-review:claim"), taskHandler.GetMyTasks)
			tasks.POST("/submit", middleware.RequirePermission("tasks:first-review:submit"), taskHandler.SubmitReview)
			tasks.POST("/submit-batch", middleware.RequirePermission("tasks:first-review:submit"), taskHandler.SubmitBatchReviews)
			tasks.POST("/return", middleware.RequirePermission("tasks:first-review:return"), taskHandler.ReturnTasks)

			// Second review routes
			tasks.POST("/second-review/claim", middleware.UserRateLimiterV2(10, time.Minute), middleware.RequirePermission("tasks:second-review:claim"), secondReviewHandler.ClaimSecondReviewTasks)
			tasks.GET("/second-review/my", middleware.RequirePermission("tasks:second-review:claim"), secondReviewHandler.GetMySecondReviewTasks)
			tasks.POST("/second-review/submit", middleware.RequirePermission("tasks:second-review:submit"), secondReviewHandler.SubmitSecondReview)
			tasks.POST("/second-review/submit-batch", middleware.RequirePermission("tasks:second-review:submit"), secondReviewHandler.SubmitBatchSecondReviews)
			tasks.POST("/second-review/return", middleware.RequirePermission("tasks:second-review:return"), secondReviewHandler.ReturnSecondReviewTasks)

			// Quality check routes
			tasks.POST("/quality-check/claim", middleware.UserRateLimiterV2(10, time.Minute), middleware.RequirePermission("tasks:quality-check:claim"), qualityCheckHandler.ClaimQCTasks)
			tasks.GET("/quality-check/my", middleware.RequirePermission("tasks:quality-check:claim"), qualityCheckHandler.GetMyQCTasks)
			tasks.POST("/quality-check/submit", middleware.RequirePermission("tasks:quality-check:submit"), qualityCheckHandler.SubmitQCReview)
			tasks.POST("/quality-check/submit-batch", middleware.RequirePermission("tasks:quality-check:submit"), qualityCheckHandler.SubmitBatchQCReviews)
			tasks.POST("/quality-check/return", middleware.RequirePermission("tasks:quality-check:return"), qualityCheckHandler.ReturnQCTasks)
			tasks.GET("/quality-check/stats", middleware.RequirePermission("tasks:quality-check:stats"), qualityCheckHandler.GetQCStats)

			// AI vs human diff routes
			tasks.POST("/ai-human-diff/claim", middleware.UserRateLimiterV2(10, time.Minute), middleware.RequirePermission("tasks:ai-human-diff:claim"), aiHumanDiffHandler.ClaimDiffTasks)
			tasks.GET("/ai-human-diff/my", middleware.RequirePermission("tasks:ai-human-diff:claim"), aiHumanDiffHandler.GetMyDiffTasks)
			tasks.POST("/ai-human-diff/submit", middleware.RequirePermission("tasks:ai-human-diff:submit"), aiHumanDiffHandler.SubmitDiffReview)
			tasks.POST("/ai-human-diff/submit-batch", middleware.RequirePermission("tasks:ai-human-diff:submit"), aiHumanDiffHandler.SubmitBatchDiffReviews)
			tasks.POST("/ai-human-diff/return", middleware.RequirePermission("tasks:ai-human-diff:return"), aiHumanDiffHandler.ReturnDiffTasks)

			// Video review routes (if video handler is available)
			if videoHandler != nil {
				// Video first review routes
				tasks.POST("/video-first-review/claim", middleware.UserRateLimiterV2(10, time.Minute), middleware.RequirePermission("tasks:video-first-review:claim"), videoHandler.ClaimVideoFirstReviewTasks)
				tasks.GET("/video-first-review/my", middleware.RequirePermission("tasks:video-first-review:claim"), videoHandler.GetMyVideoFirstReviewTasks)
				tasks.POST("/video-first-review/submit", middleware.RequirePermission("tasks:video-first-review:submit"), videoHandler.SubmitVideoFirstReview)
				tasks.POST("/video-first-review/submit-batch", middleware.RequirePermission("tasks:video-first-review:submit"), videoHandler.SubmitBatchVideoFirstReviews)
				tasks.POST("/video-first-review/return", middleware.RequirePermission("tasks:video-first-review:return"), videoHandler.ReturnVideoFirstReviewTasks)

				// Video second review routes
				tasks.POST("/video-second-review/claim", middleware.UserRateLimiterV2(10, time.Minute), middleware.RequirePermission("tasks:video-second-review:claim"), videoHandler.ClaimVideoSecondReviewTasks)
				tasks.GET("/video-second-review/my", middleware.RequirePermission("tasks:video-second-review:claim"), videoHandler.GetMyVideoSecondReviewTasks)
				tasks.POST("/video-second-review/submit", middleware.RequirePermission("tasks:video-second-review:submit"), videoHandler.SubmitVideoSecondReview)
				tasks.POST("/video-second-review/submit-batch", middleware.RequirePermission("tasks:video-second-review:submit"), videoHandler.SubmitBatchVideoSecondReviews)
				tasks.POST("/video-second-review/return", middleware.RequirePermission("tasks:video-second-review:return"), videoHandler.ReturnVideoSecondReviewTasks)
			}
		}

		// Video Queue Pool System routes (new single-stage queue system)
		video := api.Group("/video")
		video.Use(middleware.AuthMiddleware())
		{
			// Routes for each pool: 100k, 1m, 10m
			video.POST("/:pool/tasks/claim", middleware.UserRateLimiterV2(10, time.Minute), func(c *gin.Context) {
				pool := c.Param("pool")
				middleware.RequirePermission("queue.video." + pool + ".claim")(c)
				if c.IsAborted() {
					return
				}
				videoQueueHandler.ClaimTasks(c)
			})

			video.GET("/:pool/tasks/my", func(c *gin.Context) {
				pool := c.Param("pool")
				middleware.RequirePermission("queue.video." + pool + ".my")(c)
				if c.IsAborted() {
					return
				}
				videoQueueHandler.GetMyTasks(c)
			})

			video.POST("/:pool/tasks/submit", func(c *gin.Context) {
				pool := c.Param("pool")
				middleware.RequirePermission("queue.video." + pool + ".submit")(c)
				if c.IsAborted() {
					return
				}
				videoQueueHandler.SubmitReview(c)
			})

			video.POST("/:pool/tasks/submit-batch", func(c *gin.Context) {
				pool := c.Param("pool")
				middleware.RequirePermission("queue.video." + pool + ".submit")(c)
				if c.IsAborted() {
					return
				}
				videoQueueHandler.SubmitBatchReviews(c)
			})

			video.POST("/:pool/tasks/return", func(c *gin.Context) {
				pool := c.Param("pool")
				middleware.RequirePermission("queue.video." + pool + ".return")(c)
				if c.IsAborted() {
					return
				}
				videoQueueHandler.ReturnTasks(c)
			})

			// Get tags for a specific pool
			video.GET("/:pool/tags", videoQueueHandler.GetTags)
		}

		// Search route (requires search permission)
		api.GET("/tasks/search", middleware.AuthMiddleware(), middleware.RequirePermission("tasks:search"), taskHandler.SearchTasks)

		// Tag routes (public for reviewers - requires any task permission)
		api.GET("/tags", middleware.AuthMiddleware(), middleware.RequireAnyPermission("tasks:first-review:claim", "tasks:second-review:claim", "tasks:quality-check:claim", "tasks:ai-human-diff:claim", "tasks:video-first-review:claim", "tasks:video-second-review:claim"), taskHandler.GetActiveTags)

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
			admin.POST("/users", middleware.RequirePermission("users:approve"), adminHandler.CreateUser)
			admin.DELETE("/users/:id", middleware.RequirePermission("users:approve"), adminHandler.DeleteUser)

			// Statistics
			admin.GET("/stats/overview", middleware.RequirePermission("stats:overview"), adminHandler.GetOverviewStats)
			admin.GET("/stats/today", middleware.RequirePermission("stats:overview"), adminHandler.GetTodayReviewStats)
			admin.GET("/stats/hourly", middleware.RequirePermission("stats:hourly"), adminHandler.GetHourlyStats)
			admin.GET("/stats/tags", middleware.RequirePermission("stats:tags"), adminHandler.GetTagStats)
			admin.GET("/stats/reviewers", middleware.RequirePermission("stats:reviewers"), adminHandler.GetReviewerPerformance)
			admin.GET("/monitoring/metrics", middleware.RequirePermission("monitoring.read"), monitoringHandler.Metrics)
			admin.GET("/monitoring/summary", middleware.RequirePermission("monitoring.read"), monitoringHandler.DailySummary)
			admin.GET("/monitoring/endpoints", middleware.RequirePermission("monitoring.read"), monitoringHandler.DailyEndpointHealth)

			// Tag management (comment tags)
			admin.GET("/tags", middleware.RequirePermission("tags:list"), adminHandler.GetAllTags)
			admin.POST("/tags", middleware.RequirePermission("tags:create"), adminHandler.CreateTag)
			admin.PUT("/tags/:id", middleware.RequirePermission("tags:update"), adminHandler.UpdateTag)
			admin.DELETE("/tags/:id", middleware.RequirePermission("tags:delete"), adminHandler.DeleteTag)

			// Video Quality Tag management (video queue tags)
			videoTagHandler := handlers.NewVideoTagHandler()
			admin.GET("/video-tags", middleware.RequirePermission("tags:list"), videoTagHandler.GetAllVideoTags)
			admin.POST("/video-tags", middleware.RequirePermission("tags:create"), videoTagHandler.CreateVideoTag)
			admin.PUT("/video-tags/:id", middleware.RequirePermission("tags:update"), videoTagHandler.UpdateVideoTag)
			admin.DELETE("/video-tags/:id", middleware.RequirePermission("tags:delete"), videoTagHandler.DeleteVideoTag)
			admin.PATCH("/video-tags/:id/toggle", middleware.RequirePermission("tags:update"), videoTagHandler.ToggleVideoTagActive)

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

			// Video Queue Pool statistics (admin only)
			admin.GET("/video-queue/:pool/stats", middleware.RequirePermission("stats:overview"), videoQueueHandler.GetPoolStats)

			// AI review management
			admin.POST("/ai-review/jobs", middleware.RequirePermission("ai-review:jobs:create"), aiReviewHandler.CreateJob)
			admin.POST("/ai-review/jobs/:id/start", middleware.RequirePermission("ai-review:jobs:start"), aiReviewHandler.StartJob)
			admin.POST("/ai-review/jobs/:id/archive", middleware.RequirePermission("ai-review:jobs:archive"), aiReviewHandler.ArchiveJob)
			admin.POST("/ai-review/jobs/:id/unarchive", middleware.RequirePermission("ai-review:jobs:archive"), aiReviewHandler.UnarchiveJob)
			admin.GET("/ai-review/jobs", middleware.RequirePermission("ai-review:jobs:list"), aiReviewHandler.ListJobs)
			admin.GET("/ai-review/jobs/:id", middleware.RequirePermission("ai-review:jobs:read"), aiReviewHandler.GetJob)
			admin.GET("/ai-review/jobs/:id/tasks", middleware.RequirePermission("ai-review:jobs:read"), aiReviewHandler.ListJobTasks)
			admin.DELETE("/ai-review/jobs/:id/tasks", middleware.RequirePermission("ai-review:tasks:delete"), aiReviewHandler.DeleteJobTasks)
			admin.GET("/ai-review/compare", middleware.RequirePermission("ai-review:compare"), aiReviewHandler.GetComparison)

			// System documents (edit)
			admin.PUT("/docs/:key", middleware.RequirePermission("docs:edit"), documentHandler.UpdateDocument)

			// Audit log management
			admin.GET("/audit-logs", middleware.RequirePermission("audit.logs.read"), auditLogHandler.ListLogs)
			admin.GET("/audit-logs/:id", middleware.RequirePermission("audit.logs.read"), auditLogHandler.GetLog)
			admin.POST("/audit-logs/export", middleware.RequirePermission("audit.logs.export"), auditLogHandler.ExportLogs)
			admin.GET("/audit-logs/exports", middleware.RequirePermission("audit.logs.read"), auditLogHandler.ListExports)

			// Bug reports (admin only)
			admin.GET("/bug-reports", bugReportHandler.List)
			admin.POST("/bug-reports/export", bugReportHandler.Export)
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
	qcCService := services.NewQualityCheckService()
	aiHumanDiffService := services.NewAIHumanDiffService()
	videoFirstReviewService := services.NewVideoFirstReviewService()
	videoSecondReviewService := services.NewVideoSecondReviewService()
	videoQueueService := services.NewVideoQueueService()
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
		if err := qcCService.ReleaseExpiredQCTasks(); err != nil {
			log.Printf("⚠️ Error releasing expired QC tasks: %v", err)
		}
		if err := aiHumanDiffService.ReleaseExpiredDiffTasks(); err != nil {
			log.Printf("⚠️ Error releasing expired AI diff tasks: %v", err)
		}
		if err := videoFirstReviewService.ReleaseExpiredFirstReviewTasks(); err != nil {
			log.Printf("⚠️ Error releasing expired video first review tasks: %v", err)
		}
		if err := videoSecondReviewService.ReleaseExpiredSecondReviewTasks(); err != nil {
			log.Printf("⚠️ Error releasing expired video second review tasks: %v", err)
		}
		// Release expired video queue tasks (all pools)
		if err := videoQueueService.ReleaseAllExpiredTasks(); err != nil {
			log.Printf("⚠️ Error releasing expired video queue tasks: %v", err)
		}
	}
}

func startDailyStatsAggregator() {
	scheduledTasksService := services.NewScheduledTasksService(database.DB)
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	log.Println("✅ Daily stats aggregation scheduler started (runs every hour)")

	// Run initial aggregation after a short delay
	go func() {
		time.Sleep(5 * time.Minute)
		if err := scheduledTasksService.RunDailyAggregation(); err != nil {
			log.Printf("⚠️ Error in initial daily stats aggregation: %v", err)
		}
	}()

	for range ticker.C {
		if err := scheduledTasksService.RunDailyAggregation(); err != nil {
			log.Printf("⚠️ Error in daily stats aggregation: %v", err)
		}
	}
}

func startAIReviewScheduler() {
	aiReviewService := services.NewAIReviewService()
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	log.Println("✅ AI review scheduler started (runs every minute)")

	for range ticker.C {
		if err := aiReviewService.RunScheduledJobs(); err != nil {
			log.Printf("⚠️ Error in AI review scheduler: %v", err)
		}
	}
}
