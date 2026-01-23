package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"comment-review-platform/internal/config"
	"comment-review-platform/internal/models"
	"comment-review-platform/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type MonitoringHandler struct {
	db      *sql.DB
	redis   *redis.Client
	metrics *services.MetricsService
}

func NewMonitoringHandler(db *sql.DB, redisClient *redis.Client, metrics *services.MetricsService) *MonitoringHandler {
	return &MonitoringHandler{
		db:      db,
		redis:   redisClient,
		metrics: metrics,
	}
}

func (h *MonitoringHandler) Health(c *gin.Context) {
	deps := map[string]models.HealthDependency{}
	overall := "healthy"
	statusCode := http.StatusOK

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if h.db != nil {
		start := time.Now()
		err := h.db.PingContext(ctx)
		deps["database"] = healthDependency(err, start)
		if err != nil {
			overall = "unhealthy"
			statusCode = http.StatusServiceUnavailable
		}
	} else {
		deps["database"] = models.HealthDependency{Status: "unknown", Error: "not initialized"}
		overall = "degraded"
	}

	if h.redis != nil {
		start := time.Now()
		err := h.redis.Ping(ctx).Err()
		deps["redis"] = healthDependency(err, start)
		if err != nil {
			overall = "unhealthy"
			statusCode = http.StatusServiceUnavailable
		}
	} else {
		deps["redis"] = models.HealthDependency{Status: "unknown", Error: "not initialized"}
		overall = "degraded"
	}

	deps["resend"] = configDependency(config.AppConfig.ResendAPIKey != "")
	deps["r2"] = configDependency(config.AppConfig.R2AccessKeyID != "" && config.AppConfig.R2SecretAccessKey != "")
	deps["ai_service"] = configDependency(config.AppConfig.AIBaseURL != "")

	if overall == "healthy" {
		if deps["resend"].Status != "healthy" || deps["r2"].Status != "healthy" || deps["ai_service"].Status != "healthy" {
			overall = "degraded"
		}
	}

	c.JSON(statusCode, models.HealthResponse{
		Status:       overall,
		Timestamp:    time.Now().UTC(),
		Dependencies: deps,
	})
}

func (h *MonitoringHandler) Metrics(c *gin.Context) {
	if h.metrics == nil {
		c.JSON(http.StatusOK, models.MetricsResponse{WindowMinutes: 0, Endpoints: []models.EndpointMetrics{}})
		return
	}

	window := config.AppConfig.MetricsWindowMinutes
	if windowStr := c.Query("window_minutes"); windowStr != "" {
		if parsed, err := strconv.Atoi(windowStr); err == nil && parsed > 0 {
			window = parsed
		}
	}

	metrics, err := h.metrics.GetMetrics(window)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.MetricsResponse{
		WindowMinutes: window,
		Endpoints:     metrics,
	})
}

func (h *MonitoringHandler) DailySummary(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database not initialized"})
		return
	}

	start, end, dateStr, err := parseDateRange(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var total, errors, clientErrors, serverErrors int
	query := `
		SELECT
			COUNT(*) AS total_requests,
			COALESCE(SUM(CASE WHEN status_code >= 400 THEN 1 ELSE 0 END), 0) AS error_requests,
			COALESCE(SUM(CASE WHEN status_code BETWEEN 400 AND 499 THEN 1 ELSE 0 END), 0) AS client_errors,
			COALESCE(SUM(CASE WHEN status_code >= 500 THEN 1 ELSE 0 END), 0) AS server_errors
		FROM audit_logs
		WHERE created_at >= $1 AND created_at < $2
	`

	if err := h.db.QueryRow(query, start, end).Scan(&total, &errors, &clientErrors, &serverErrors); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.MonitoringSummary{
		Date:          dateStr,
		TotalRequests: total,
		ErrorRequests: errors,
		ClientErrors:  clientErrors,
		ServerErrors:  serverErrors,
	})
}

func (h *MonitoringHandler) DailyEndpointHealth(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database not initialized"})
		return
	}

	start, end, dateStr, err := parseDateRange(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	limit := 100
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	query := `
		SELECT
			endpoint,
			http_method,
			COUNT(*) AS total_requests,
			COALESCE(SUM(CASE WHEN status_code < 400 THEN 1 ELSE 0 END), 0) AS success_requests,
			COALESCE(SUM(CASE WHEN status_code BETWEEN 400 AND 499 THEN 1 ELSE 0 END), 0) AS client_errors,
			COALESCE(SUM(CASE WHEN status_code >= 500 THEN 1 ELSE 0 END), 0) AS server_errors,
			COALESCE(AVG(duration_ms), 0) AS avg_latency_ms,
			COALESCE(PERCENTILE_CONT(0.99) WITHIN GROUP (ORDER BY duration_ms), 0) AS p99_latency_ms
		FROM audit_logs
		WHERE created_at >= $1 AND created_at < $2
		GROUP BY endpoint, http_method
		ORDER BY total_requests DESC
		LIMIT $3
	`

	rows, err := h.db.Query(query, start, end, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	endpoints := []models.EndpointMetrics{}
	for rows.Next() {
		var endpoint, method string
		var total, success, clientErrors, serverErrors int
		var avgLatency float64
		var p99Latency float64

		if err := rows.Scan(&endpoint, &method, &total, &success, &clientErrors, &serverErrors, &avgLatency, &p99Latency); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		successRate := 0.0
		if total > 0 {
			successRate = float64(success) / float64(total)
		}

		endpoints = append(endpoints, models.EndpointMetrics{
			Method:       method,
			Path:         endpoint,
			Total:        total,
			Success:      success,
			ClientError:  clientErrors,
			ServerError:  serverErrors,
			SuccessRate:  successRate,
			AvgLatencyMs: avgLatency,
			P99LatencyMs: int(p99Latency),
		})
	}

	c.JSON(http.StatusOK, models.EndpointHealthResponse{
		Date:      dateStr,
		Endpoints: endpoints,
	})
}

func healthDependency(err error, start time.Time) models.HealthDependency {
	if err != nil {
		return models.HealthDependency{
			Status: "unhealthy",
			Error:  err.Error(),
		}
	}
	return models.HealthDependency{
		Status:    "healthy",
		LatencyMs: time.Since(start).Milliseconds(),
	}
}

func configDependency(ok bool) models.HealthDependency {
	if ok {
		return models.HealthDependency{Status: "healthy"}
	}
	return models.HealthDependency{Status: "degraded", Error: "not configured"}
}

func parseDateRange(c *gin.Context) (time.Time, time.Time, string, error) {
	dateStr := c.Query("date")
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	start, err := time.ParseInLocation("2006-01-02", dateStr, time.Local)
	if err != nil {
		return time.Time{}, time.Time{}, "", err
	}
	end := start.Add(24 * time.Hour)
	return start, end, dateStr, nil
}
