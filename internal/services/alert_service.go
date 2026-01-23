package services

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"strings"
	"time"

	"comment-review-platform/internal/config"
	redispkg "comment-review-platform/pkg/redis"

	"github.com/redis/go-redis/v9"
)

type AlertService struct {
	emailService    *EmailService
	rdb             *redis.Client
	webhookURLs     []string
	emailRecipients []string
	silenceDuration time.Duration
	thresholdWindow time.Duration
	errorThreshold  int64
	detailsBaseURL  string
}

type AlertEvent struct {
	TraceID          string
	HTTPMethod       string
	Endpoint         string
	StatusCode       int
	OccurredAt       time.Time
	ErrorCode        string
	ErrorType        string
	ErrorDescription string
	ErrorMessage     string
	ErrorStack       string
	ModuleName       string
	MethodName       string
	ServerIP         string
	ServerPort       string
	PageURL          string
	UserID           *int
	Username         string
	UserAgent        string
	RequestBody      []byte
	RequestParams    []byte
	ResponseBody     []byte
}

type webhookPayload struct {
	AlertLevel  string `json:"alert_level"`
	TraceID     string `json:"trace_id"`
	ErrorMsg    string `json:"error_msg"`
	APIEndpoint string `json:"api_endpoint"`
	HTTPMethod  string `json:"http_method,omitempty"`
	StatusCode  int    `json:"status_code,omitempty"`
	OccurredAt  string `json:"occurred_at"`
	DetailsLink string `json:"details_link,omitempty"`
	UserID      *int   `json:"user_id,omitempty"`
	Username    string `json:"username,omitempty"`
}

func NewAlertService(rdb *redis.Client) *AlertService {
	cfg := config.AppConfig
	recipients := splitAndTrim(cfg.AlertEmailRecipients)
	webhooks := splitAndTrim(cfg.AlertWebhookURLs)
	silenceSeconds := cfg.AlertSilenceSeconds
	if silenceSeconds <= 0 {
		silenceSeconds = 300
	}
	thresholdWindowSeconds := cfg.AlertThresholdWindowSeconds
	if thresholdWindowSeconds <= 0 {
		thresholdWindowSeconds = 60
	}
	threshold := cfg.AlertErrorThreshold
	if threshold <= 0 {
		threshold = 1
	}

	var emailService *EmailService
	if cfg.ResendAPIKey != "" && len(recipients) > 0 {
		emailService = NewEmailService()
	}

	if rdb == nil {
		rdb = redispkg.Client
	}

	return &AlertService{
		emailService:    emailService,
		rdb:             rdb,
		webhookURLs:     webhooks,
		emailRecipients: recipients,
		silenceDuration: time.Duration(silenceSeconds) * time.Second,
		thresholdWindow: time.Duration(thresholdWindowSeconds) * time.Second,
		errorThreshold:  int64(threshold),
		detailsBaseURL:  strings.TrimRight(cfg.AlertDetailsBaseURL, "/"),
	}
}

func (s *AlertService) NotifyFromAuditLog(entry AlertEvent) error {
	if s == nil || entry.StatusCode < 400 {
		return nil
	}
	if shouldSkipAlert(entry) {
		return nil
	}

	level := "ERROR"
	if entry.ErrorStack != "" {
		level = "CRITICAL"
	}

	signature := s.signature(entry, level)
	if s.isSilenced(signature) {
		return nil
	}

	if shouldApplyThreshold(entry.StatusCode, level) {
		count := s.incrementErrorCount(signature)
		if count < s.errorThreshold {
			return nil
		}
	}

	s.setSilence(signature)

	payload := s.buildWebhookPayload(entry, level)
	if err := s.sendEmail(entry, level); err != nil {
		return err
	}
	if level == "CRITICAL" {
		if err := s.sendWebhooks(payload); err != nil {
			return err
		}
	}
	return nil
}

func shouldApplyThreshold(statusCode int, level string) bool {
	if level == "CRITICAL" {
		return false
	}
	return statusCode >= 400 && statusCode < 500
}

func shouldSkipAlert(entry AlertEvent) bool {
	if entry.StatusCode == http.StatusForbidden {
		return true
	}
	if strings.EqualFold(entry.ErrorType, "permission_denied") {
		return true
	}
	if strings.EqualFold(entry.ErrorCode, "PERMISSION_DENIED") {
		return true
	}
	return false
}

func (s *AlertService) signature(entry AlertEvent, level string) string {
	raw := fmt.Sprintf("%s|%s|%d|%s|%s", level, entry.HTTPMethod, entry.StatusCode, entry.Endpoint, entry.ErrorMessage)
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func (s *AlertService) incrementErrorCount(signature string) int64 {
	if s.rdb == nil {
		return s.errorThreshold
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	key := fmt.Sprintf("alert:count:%s", signature)
	count, err := s.rdb.Incr(ctx, key).Result()
	if err != nil {
		return s.errorThreshold
	}
	_ = s.rdb.Expire(ctx, key, s.thresholdWindow).Err()
	return count
}

func (s *AlertService) isSilenced(signature string) bool {
	if s.rdb == nil {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	key := fmt.Sprintf("alert:silence:%s", signature)
	exists, err := s.rdb.Exists(ctx, key).Result()
	return err == nil && exists > 0
}

func (s *AlertService) setSilence(signature string) {
	if s.rdb == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	key := fmt.Sprintf("alert:silence:%s", signature)
	_ = s.rdb.Set(ctx, key, "1", s.silenceDuration).Err()
}

func (s *AlertService) buildWebhookPayload(entry AlertEvent, level string) webhookPayload {
	detailsLink := ""
	if s.detailsBaseURL != "" && entry.TraceID != "" {
		detailsLink = fmt.Sprintf("%s/trace/%s", s.detailsBaseURL, entry.TraceID)
	}

	occurredAt := entry.OccurredAt.Format("2006-01-02 15:04:05")
	return webhookPayload{
		AlertLevel:  level,
		TraceID:     entry.TraceID,
		ErrorMsg:    entry.ErrorMessage,
		APIEndpoint: entry.Endpoint,
		HTTPMethod:  entry.HTTPMethod,
		StatusCode:  entry.StatusCode,
		OccurredAt:  occurredAt,
		DetailsLink: detailsLink,
		UserID:      entry.UserID,
		Username:    entry.Username,
	}
}

func (s *AlertService) sendEmail(entry AlertEvent, level string) error {
	if s.emailService == nil || len(s.emailRecipients) == 0 {
		return nil
	}

	subject := fmt.Sprintf("[%s] %s %s (%d)", level, entry.HTTPMethod, entry.Endpoint, entry.StatusCode)
	traceID := entry.TraceID
	body := buildAlertEmailBody(entry, level, traceID)
	return s.emailService.SendErrorEmail(s.emailRecipients, subject, body)
}

func (s *AlertService) sendWebhooks(payload webhookPayload) error {
	if len(s.webhookURLs) == 0 {
		return nil
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 5 * time.Second}
	for _, url := range s.webhookURLs {
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(payloadBytes))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		_ = resp.Body.Close()
	}
	return nil
}

func buildAlertEmailBody(entry AlertEvent, level, traceID string) string {
	requestBody := html.EscapeString(string(entry.RequestBody))
	requestParams := html.EscapeString(string(entry.RequestParams))
	responseBody := html.EscapeString(string(entry.ResponseBody))
	errorStack := html.EscapeString(entry.ErrorStack)
	userID := formatUserID(entry.UserID)

	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
		.container { max-width: 720px; margin: 0 auto; padding: 20px; }
		.header { font-size: 18px; font-weight: bold; margin-bottom: 12px; }
		.section { margin: 12px 0; }
		pre { background: #f6f6f6; padding: 12px; border-radius: 6px; overflow-x: auto; }
		.meta { color: #555; font-size: 13px; }
	</style>
	</head>
	<body>
		<div class="container">
			<div class="header">[%s] %s %s (%d)</div>
			<div class="meta">Occurred at: %s</div>
			<div class="section"><strong>Trace ID:</strong> %s</div>
	<div class="section"><strong>User:</strong> %s (ID: %s)</div>
			<div class="section"><strong>Client:</strong> %s</div>
	<div class="section"><strong>Module:</strong> %s</div>
	<div class="section"><strong>Method:</strong> %s</div>
	<div class="section"><strong>Server:</strong> %s:%s</div>
	<div class="section"><strong>Page URL:</strong> %s</div>
	<div class="section"><strong>Error Code:</strong> %s</div>
	<div class="section"><strong>Error Type:</strong> %s</div>
	<div class="section"><strong>Error Description:</strong> %s</div>
			<div class="section"><strong>Error:</strong> %s</div>
			<div class="section"><strong>Request Body:</strong><pre>%s</pre></div>
			<div class="section"><strong>Request Params:</strong><pre>%s</pre></div>
			<div class="section"><strong>Response Body:</strong><pre>%s</pre></div>
			<div class="section"><strong>Stack Trace:</strong><pre>%s</pre></div>
		</div>
	</body>
	</html>
	`,
		level,
		entry.HTTPMethod,
		entry.Endpoint,
		entry.StatusCode,
		entry.OccurredAt.Format("2006-01-02 15:04:05"),
		traceID,
		entry.Username,
		userID,
		entry.UserAgent,
		emptyFallback(entry.ModuleName),
		emptyFallback(entry.MethodName),
		emptyFallback(entry.ServerIP),
		emptyFallback(entry.ServerPort),
		emptyFallback(entry.PageURL),
		emptyFallback(entry.ErrorCode),
		emptyFallback(entry.ErrorType),
		emptyFallback(entry.ErrorDescription),
		entry.ErrorMessage,
		truncateText(requestBody, 4000),
		truncateText(requestParams, 4000),
		truncateText(responseBody, 4000),
		truncateText(errorStack, 4000),
	)
}

func truncateText(value string, max int) string {
	if len(value) <= max {
		return value
	}
	return value[:max] + "..."
}

func formatUserID(userID *int) string {
	if userID == nil {
		return "-"
	}
	return fmt.Sprintf("%d", *userID)
}

func emptyFallback(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	return value
}

func splitAndTrim(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	results := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			results = append(results, trimmed)
		}
	}
	return results
}
