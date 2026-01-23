package models

import "time"

type EndpointMetrics struct {
	Method       string  `json:"method"`
	Path         string  `json:"path"`
	Total        int     `json:"total"`
	Success      int     `json:"success"`
	ClientError  int     `json:"client_error"`
	ServerError  int     `json:"server_error"`
	SuccessRate  float64 `json:"success_rate"`
	AvgLatencyMs float64 `json:"avg_latency_ms"`
	P99LatencyMs int     `json:"p99_latency_ms"`
}

type MetricsResponse struct {
	WindowMinutes int               `json:"window_minutes"`
	Endpoints     []EndpointMetrics `json:"endpoints"`
}

type MonitoringSummary struct {
	Date          string `json:"date"`
	TotalRequests int    `json:"total_requests"`
	ErrorRequests int    `json:"error_requests"`
	ClientErrors  int    `json:"client_errors"`
	ServerErrors  int    `json:"server_errors"`
}

type EndpointHealthResponse struct {
	Date      string            `json:"date"`
	Endpoints []EndpointMetrics `json:"endpoints"`
}

type HealthDependency struct {
	Status    string `json:"status"`
	Error     string `json:"error,omitempty"`
	LatencyMs int64  `json:"latency_ms,omitempty"`
}

type HealthResponse struct {
	Status       string                      `json:"status"`
	Timestamp    time.Time                   `json:"timestamp"`
	Dependencies map[string]HealthDependency `json:"dependencies"`
}
