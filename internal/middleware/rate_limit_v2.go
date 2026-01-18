package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimitConfig 限流器配置
type RateLimitConfig struct {
	Limit       int           // 时间窗口内允许的最大请求数
	Window      time.Duration // 时间窗口
	KeyFunc     KeyFunc       // 生成限流 key 的函数
	ErrorMsg    string        // 超限时的错误消息
	SkipOnError bool          // key 生成失败时是否跳过限流
}

// KeyFunc 生成限流 key 的函数类型
type KeyFunc func(c *gin.Context) (string, bool)

// rateLimiter 通用限流器
type rateLimiter struct {
	config  RateLimitConfig
	store   map[string][]time.Time
	mutex   sync.Mutex
}

// newRateLimiter 创建新的限流器
func newRateLimiter(config RateLimitConfig) *rateLimiter {
	return &rateLimiter{
		config: config,
		store:  make(map[string][]time.Time),
	}
}

// allow 检查是否允许请求
func (r *rateLimiter) allow(key string) bool {
	now := time.Now()
	cutoff := now.Add(-r.config.Window)

	r.mutex.Lock()
	defer r.mutex.Unlock()

	// 清理过期的时间戳
	var validTimes []time.Time
	for _, t := range r.store[key] {
		if t.After(cutoff) {
			validTimes = append(validTimes, t)
		}
	}
	r.store[key] = validTimes

	// 检查是否超限
	if len(r.store[key]) >= r.config.Limit {
		return false
	}

	// 记录当前请求
	r.store[key] = append(r.store[key], now)
	return true
}

// CreateRateLimiter 创建通用限流中间件
func CreateRateLimiter(config RateLimitConfig) gin.HandlerFunc {
	limiter := newRateLimiter(config)

	return func(c *gin.Context) {
		key, ok := config.KeyFunc(c)
		if !ok {
			if config.SkipOnError {
				c.Next()
				return
			}
			// 如果 key 生成失败且不跳过，使用默认 key
			key = "default"
		}

		if !limiter.allow(key) {
			errorMsg := config.ErrorMsg
			if errorMsg == "" {
				errorMsg = "Rate limit exceeded"
			}
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       errorMsg,
				"retry_after": config.Window.String(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// 预定义的 KeyFunc 函数

// IPKeyFunc 基于客户端 IP 生成 key
func IPKeyFunc(c *gin.Context) (string, bool) {
	return c.ClientIP(), true
}

// IPEndpointKeyFunc 基于客户端 IP 和端点生成 key
func IPEndpointKeyFunc(c *gin.Context) (string, bool) {
	return c.ClientIP() + ":" + c.FullPath(), true
}

// UserKeyFunc 基于用户 ID 生成 key
func UserKeyFunc(c *gin.Context) (string, bool) {
	userID := GetUserID(c)
	if userID == 0 {
		return "", false
	}
	return fmt.Sprintf("user:%d", userID), true
}

// 预定义的限流器配置

// GlobalRateLimitConfig 全局限流配置（100 请求/秒/IP）
func GlobalRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Limit:    100,
		Window:   time.Second,
		KeyFunc:  IPKeyFunc,
		ErrorMsg: "Rate limit exceeded",
	}
}

// EndpointRateLimitConfig 端点限流配置
func EndpointRateLimitConfig(limit int, window time.Duration) RateLimitConfig {
	return RateLimitConfig{
		Limit:    limit,
		Window:   window,
		KeyFunc:  IPEndpointKeyFunc,
		ErrorMsg: "Too many requests",
	}
}

// UserRateLimitConfig 用户限流配置
func UserRateLimitConfig(limit int, window time.Duration) RateLimitConfig {
	return RateLimitConfig{
		Limit:       limit,
		Window:      window,
		KeyFunc:     UserKeyFunc,
		ErrorMsg:    "User rate limit exceeded",
		SkipOnError: true, // 未登录用户跳过限流
	}
}

// 使用新工厂函数重新实现原有限流器（保持向后兼容）

// GlobalRateLimiterV2 使用工厂函数实现的全局限流器
func GlobalRateLimiterV2() gin.HandlerFunc {
	return CreateRateLimiter(GlobalRateLimitConfig())
}

// EndpointRateLimiterV2 使用工厂函数实现的端点限流器
func EndpointRateLimiterV2(limit int, window time.Duration) gin.HandlerFunc {
	return CreateRateLimiter(EndpointRateLimitConfig(limit, window))
}

// UserRateLimiterV2 使用工厂函数实现的用户限流器
func UserRateLimiterV2(limit int, window time.Duration) gin.HandlerFunc {
	return CreateRateLimiter(UserRateLimitConfig(limit, window))
}
