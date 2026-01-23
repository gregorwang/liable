package middleware

import (
	"time"

	"comment-review-platform/internal/services"

	"github.com/gin-gonic/gin"
)

var metricsService *services.MetricsService

// InitMetricsService initializes the global metrics service.
func InitMetricsService(service *services.MetricsService) {
	metricsService = service
}

// MetricsMiddleware records basic API metrics.
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		if metricsService == nil {
			return
		}

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		metricsService.Record(c.Request.Method, path, c.Writer.Status(), time.Since(start))
	}
}
