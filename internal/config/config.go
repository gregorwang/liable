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
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	RedisUseTLS   bool

	// PostgreSQL Configuration
	DatabaseURL string

	// Task Configuration
	TaskClaimSize      int
	TaskTimeoutMinutes int

	// Cloudflare R2 Configuration
	CloudflareAccountID string
	R2AccessKeyID       string
	R2SecretAccessKey   string
	R2BucketName        string
	R2Endpoint          string
	R2VideoPathPrefix   string
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
	taskClaimSize, _ := strconv.Atoi(getEnv("TASK_CLAIM_SIZE", "20"))
	taskTimeoutMinutes, _ := strconv.Atoi(getEnv("TASK_TIMEOUT_MINUTES", "30"))

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
		DatabaseURL:        databaseURL,
		TaskClaimSize:      taskClaimSize,
		TaskTimeoutMinutes: taskTimeoutMinutes,

		// Cloudflare R2 Configuration
		CloudflareAccountID: getEnv("CLOUDFLARE_ACCOUNT_ID", ""),
		R2AccessKeyID:       getEnv("R2_ACCESS_KEY_ID", ""),
		R2SecretAccessKey:   getEnv("R2_SECRET_ACCESS_KEY", ""),
		R2BucketName:        getEnv("R2_BUCKET_NAME", ""),
		R2Endpoint:          getEnv("R2_ENDPOINT", ""),
		R2VideoPathPrefix:   getEnv("R2_VIDEO_PATH_PREFIX", "gregorwang/douyin/Postman Agent/"),
	}

	return AppConfig
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
