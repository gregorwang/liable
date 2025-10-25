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
	adminHandler := handlers.NewAdminHandler()

	// Type assert database connection
	sqlDB, ok := db.(*sql.DB)
	if !ok {
		panic("failed to assert database connection to *sql.DB")
	}
	moderationRulesHandler := handlers.NewModerationRulesHandler(sqlDB)

	// API routes
	api := router.Group("/api")
	{
		// Auth routes (public)
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.GET("/profile", middleware.AuthMiddleware(), authHandler.GetProfile)
		}

		// Moderation Rules routes (public - for viewing rules)
		modRules := api.Group("/moderation-rules")
		{
			modRules.GET("", moderationRulesHandler.ListRules)
			modRules.GET("/:code", moderationRulesHandler.GetRuleByCode)
			modRules.GET("/categories", moderationRulesHandler.GetCategories)
			modRules.GET("/risk-levels", moderationRulesHandler.GetRiskLevels)
		}

		// Task routes (requires reviewer role)
		tasks := api.Group("/tasks")
		tasks.Use(middleware.AuthMiddleware(), middleware.RequireReviewer())
		{
			tasks.POST("/claim", taskHandler.ClaimTasks)
			tasks.GET("/my", taskHandler.GetMyTasks)
			tasks.POST("/submit", taskHandler.SubmitReview)
			tasks.POST("/submit-batch", taskHandler.SubmitBatchReviews)
			tasks.POST("/return", taskHandler.ReturnTasks)
		}

		// Search route (requires login, available for both admin and reviewer)
		api.GET("/tasks/search", middleware.AuthMiddleware(), taskHandler.SearchTasks)

		// Tag routes (public for reviewers)
		api.GET("/tags", middleware.AuthMiddleware(), taskHandler.GetActiveTags)

		// Admin routes (requires admin role)
		admin := api.Group("/admin")
		admin.Use(middleware.AuthMiddleware(), middleware.RequireAdmin())
		{
			// User management
			admin.GET("/users", adminHandler.GetPendingUsers)
			admin.PUT("/users/:id/approve", adminHandler.ApproveUser)

			// Statistics
			admin.GET("/stats/overview", adminHandler.GetOverviewStats)
			admin.GET("/stats/hourly", adminHandler.GetHourlyStats)
			admin.GET("/stats/tags", adminHandler.GetTagStats)
			admin.GET("/stats/reviewers", adminHandler.GetReviewerPerformance)

			// Tag management
			admin.GET("/tags", adminHandler.GetAllTags)
			admin.POST("/tags", adminHandler.CreateTag)
			admin.PUT("/tags/:id", adminHandler.UpdateTag)
			admin.DELETE("/tags/:id", adminHandler.DeleteTag)
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

func startTaskReleaseWorker() {
	taskService := services.NewTaskService()
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	log.Println("✅ Background task release worker started")

	for range ticker.C {
		if err := taskService.ReleaseExpiredTasks(); err != nil {
			log.Printf("⚠️ Error releasing expired tasks: %v", err)
		}
	}
}
