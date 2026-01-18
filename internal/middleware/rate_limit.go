package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// GlobalRateLimiter returns a middleware that limits requests based on IP address
func GlobalRateLimiter() gin.HandlerFunc {
	// Simple in-memory rate limiting per IP
	// Limit: 100 requests per second per IP
	limiter := map[string][]time.Time{}
	var mutex sync.Mutex

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		now := time.Now()

		mutex.Lock()

		// Clean up old entries (older than 1 second)
		cutoff := now.Add(-1 * time.Second)
		var validTimes []time.Time
		for _, t := range limiter[clientIP] {
			if t.After(cutoff) {
				validTimes = append(validTimes, t)
			}
		}
		limiter[clientIP] = validTimes

		// Check limit
		if len(limiter[clientIP]) >= 100 {
			mutex.Unlock()
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded", "retry_after": "1s"})
			c.Abort()
			return
		}

		// Add current request
		limiter[clientIP] = append(limiter[clientIP], now)
		mutex.Unlock()
		c.Next()
	}
}

// EndpointRateLimiter returns a middleware that limits requests for specific endpoints
func EndpointRateLimiter(limit int, window time.Duration) gin.HandlerFunc {
	// Store request timestamps per endpoint per IP
	limiter := map[string][]time.Time{}
	var mutex sync.Mutex

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		path := c.FullPath()
		key := clientIP + ":" + path
		now := time.Now()

		mutex.Lock()

		// Clean up old entries
		cutoff := now.Add(-window)
		var validTimes []time.Time
		for _, t := range limiter[key] {
			if t.After(cutoff) {
				validTimes = append(validTimes, t)
			}
		}
		limiter[key] = validTimes

		// Check limit
		if len(limiter[key]) >= limit {
			mutex.Unlock()
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests", "retry_after": window.String()})
			c.Abort()
			return
		}

		// Add current request
		limiter[key] = append(limiter[key], now)
		mutex.Unlock()
		c.Next()
	}
}

// UserRateLimiter returns a middleware that limits requests per authenticated user
func UserRateLimiter(limit int, window time.Duration) gin.HandlerFunc {
	// Store request timestamps per user
	limiter := map[string][]time.Time{}
	var mutex sync.Mutex

	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == 0 {
			c.Next()
			return
		}

		key := fmt.Sprintf("user:%d", userID)
		now := time.Now()

		mutex.Lock()

		// Clean up old entries
		cutoff := now.Add(-window)
		var validTimes []time.Time
		for _, t := range limiter[key] {
			if t.After(cutoff) {
				validTimes = append(validTimes, t)
			}
		}
		limiter[key] = validTimes

		// Check limit
		if len(limiter[key]) >= limit {
			mutex.Unlock()
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "User rate limit exceeded", "retry_after": window.String()})
			c.Abort()
			return
		}

		// Add current request
		limiter[key] = append(limiter[key], now)
		mutex.Unlock()
		c.Next()
	}
}
