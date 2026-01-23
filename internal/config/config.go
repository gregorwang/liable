package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server Configuration
	Port      string
	JWTSecret string

	// Redis Configuration
	RedisAddr          string
	RedisPassword      string
	RedisDB            int
	RedisUseTLS        bool
	RedisTLSSkipVerify bool

	// PostgreSQL Configuration
	DatabaseURL string

	// Task Configuration
	TaskClaimSize      int
	TaskTimeoutMinutes int

	// Cloudflare R2 Configuration
	CloudflareAccountID   string
	R2AccessKeyID         string
	R2SecretAccessKey     string
	R2BucketName          string
	R2Endpoint            string
	R2VideoPathPrefix     string
	R2BugReportPathPrefix string
	R2AvatarPathPrefix    string

	// Resend (Email) Configuration
	ResendAPIKey    string
	ResendFromEmail string

	// AI Review Configuration
	AIBaseURL        string
	AIAPIKey         string
	AIModel          string
	AITimeoutSeconds int
	AIConcurrency    int

	// Alerting Configuration
	AlertEmailRecipients        string
	AlertWebhookURLs            string
	AlertErrorThreshold         int
	AlertThresholdWindowSeconds int
	AlertSilenceSeconds         int
	AlertDetailsBaseURL         string

	// Metrics Configuration
	MetricsWindowMinutes int
}

var AppConfig *Config

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	// Load .env file if exists
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("⚠️  Warning: Could not load .env file: %v", err)
		log.Println("Using system environment variables instead")
	} else {
		log.Println("✅ .env file loaded successfully")
	}

	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))
	redisUseTLS := getEnv("REDIS_USE_TLS", "false") == "true"
	redisTLSSkipVerify := getEnv("REDIS_TLS_SKIP_VERIFY", "false") == "true"
	taskClaimSize, _ := strconv.Atoi(getEnv("TASK_CLAIM_SIZE", "20"))
	taskTimeoutMinutes, _ := strconv.Atoi(getEnv("TASK_TIMEOUT_MINUTES", "30"))
	aiTimeoutSeconds, _ := strconv.Atoi(getEnv("AI_TIMEOUT_SECONDS", "30"))
	aiConcurrency, _ := strconv.Atoi(getEnv("AI_CONCURRENCY", "5"))
	aiBaseURL := getEnv("AI_BASE_URL", getEnv("OPENAI_BASE_URL", ""))
	aiAPIKey := getEnv("AI_API_KEY", getEnv("OPENAI_API_KEY", ""))
	aiModel := getEnv("AI_MODEL", getEnv("OPENAI_MODEL", ""))
	alertErrorThreshold, _ := strconv.Atoi(getEnv("ALERT_ERROR_THRESHOLD", "1"))
	alertThresholdWindowSeconds, _ := strconv.Atoi(getEnv("ALERT_THRESHOLD_WINDOW_SECONDS", "60"))
	alertSilenceSeconds, _ := strconv.Atoi(getEnv("ALERT_SILENCE_SECONDS", "300"))
	metricsWindowMinutes, _ := strconv.Atoi(getEnv("METRICS_WINDOW_MINUTES", "5"))

	databaseURL := getEnv("DATABASE_URL", "")
	if databaseURL == "" {
		log.Println("❌ DATABASE_URL is not set!")
	} else {
		// 只显示 URL 的前 30 个字符以保护敏感信息
		preview := databaseURL
		if len(preview) > 30 {
			preview = preview[:30] + "..."
		}
		log.Printf("✅ DATABASE_URL loaded: %s", preview)
	}

	AppConfig = &Config{
		Port:               getEnv("PORT", "8080"),
		JWTSecret:          getEnv("JWT_SECRET", "your-secret-key-change-this"),
		RedisAddr:          getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:      getEnv("REDIS_PASSWORD", ""),
		RedisDB:            redisDB,
		RedisUseTLS:        redisUseTLS,
		RedisTLSSkipVerify: redisTLSSkipVerify,
		DatabaseURL:        databaseURL,
		ResendAPIKey:       getEnv("RESEND_API_KEY", ""),
		ResendFromEmail:    getEnv("RESEND_FROM_EMAIL", "noreply@wangjiajun.asia"),
		TaskClaimSize:      taskClaimSize,
		TaskTimeoutMinutes: taskTimeoutMinutes,

		// AI Review Configuration
		AIBaseURL:        aiBaseURL,
		AIAPIKey:         aiAPIKey,
		AIModel:          aiModel,
		AITimeoutSeconds: aiTimeoutSeconds,
		AIConcurrency:    aiConcurrency,

		// Alerting Configuration
		AlertEmailRecipients:        getEnv("ALERT_EMAIL_RECIPIENTS", ""),
		AlertWebhookURLs:            getEnv("ALERT_WEBHOOK_URLS", ""),
		AlertErrorThreshold:         alertErrorThreshold,
		AlertThresholdWindowSeconds: alertThresholdWindowSeconds,
		AlertSilenceSeconds:         alertSilenceSeconds,
		AlertDetailsBaseURL:         getEnv("ALERT_DETAILS_BASE_URL", ""),

		// Metrics Configuration
		MetricsWindowMinutes: metricsWindowMinutes,

		// Cloudflare R2 Configuration
		CloudflareAccountID:   getEnv("CLOUDFLARE_ACCOUNT_ID", ""),
		R2AccessKeyID:         getEnv("R2_ACCESS_KEY_ID", ""),
		R2SecretAccessKey:     getEnv("R2_SECRET_ACCESS_KEY", ""),
		R2BucketName:          getEnv("R2_BUCKET_NAME", ""),
		R2Endpoint:            getEnv("R2_ENDPOINT", ""),
		R2VideoPathPrefix:     getEnv("R2_VIDEO_PATH_PREFIX", "gregorwang/douyin/Postman Agent/"),
		R2BugReportPathPrefix: getEnv("R2_BUG_REPORT_PATH_PREFIX", "bug-reports/screenshots/"),
		R2AvatarPathPrefix:    getEnv("R2_AVATAR_PATH_PREFIX", "user-avatars/"),
	}

	return AppConfig
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
