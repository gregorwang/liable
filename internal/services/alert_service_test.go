package services

import (
	"strings"
	"testing"
	"time"
)

func TestShouldApplyThreshold(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		level      string
		want       bool
	}{
		{name: "client error 404", statusCode: 404, level: "ERROR", want: true},
		{name: "client error 400", statusCode: 400, level: "ERROR", want: true},
		{name: "client error 499", statusCode: 499, level: "ERROR", want: true},
		{name: "server error 500", statusCode: 500, level: "ERROR", want: false},
		{name: "server error 503", statusCode: 503, level: "ERROR", want: false},
		{name: "critical bypass threshold", statusCode: 404, level: "CRITICAL", want: false},
		{name: "success bypass threshold", statusCode: 200, level: "ERROR", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldApplyThreshold(tt.statusCode, tt.level); got != tt.want {
				t.Fatalf("shouldApplyThreshold(%d, %q) = %v, want %v", tt.statusCode, tt.level, got, tt.want)
			}
		})
	}
}

func TestShouldSkipAlert(t *testing.T) {
	userID := 12
	tests := []struct {
		name  string
		entry AlertEvent
		want  bool
	}{
		{
			name: "skip forbidden status",
			entry: AlertEvent{
				StatusCode: 403,
			},
			want: true,
		},
		{
			name: "skip permission denied error type",
			entry: AlertEvent{
				StatusCode: 500,
				ErrorType:  "permission_denied",
			},
			want: true,
		},
		{
			name: "skip permission denied error code",
			entry: AlertEvent{
				StatusCode: 500,
				ErrorCode:  "PERMISSION_DENIED",
			},
			want: true,
		},
		{
			name: "do not skip other errors",
			entry: AlertEvent{
				StatusCode: 500,
				ErrorCode:  "INTERNAL_ERROR",
				UserID:     &userID,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldSkipAlert(tt.entry); got != tt.want {
				t.Fatalf("shouldSkipAlert() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildAlertEmailBodyIncludesContext(t *testing.T) {
	userID := 42
	entry := AlertEvent{
		HTTPMethod:       "GET",
		Endpoint:         "/api/admin/stats/today",
		StatusCode:       500,
		OccurredAt:       time.Date(2026, 1, 22, 20, 0, 0, 0, time.UTC),
		ErrorCode:        "INTERNAL_ERROR",
		ErrorType:        "internal_error",
		ErrorDescription: "boom",
		ErrorMessage:     "boom",
		ModuleName:       "internal/handlers/admin",
		MethodName:       "GetTodayReviewStats",
		ServerIP:         "10.0.0.1",
		ServerPort:       "8080",
		PageURL:          "https://example.com/admin",
		UserID:           &userID,
		Username:         "test@example.com",
		UserAgent:        "TestAgent",
	}

	body := buildAlertEmailBody(entry, "ERROR", "trace-123")

	if !strings.Contains(body, "ID: 42") {
		t.Fatalf("expected user ID to be rendered, got body: %s", body)
	}
	if !strings.Contains(body, "Module:</strong> internal/handlers/admin") {
		t.Fatalf("expected module name in email body, got body: %s", body)
	}
	if !strings.Contains(body, "Method:</strong> GetTodayReviewStats") {
		t.Fatalf("expected method name in email body, got body: %s", body)
	}
	if !strings.Contains(body, "Server:</strong> 10.0.0.1:8080") {
		t.Fatalf("expected server info in email body, got body: %s", body)
	}
	if !strings.Contains(body, "Page URL:</strong> https://example.com/admin") {
		t.Fatalf("expected page URL in email body, got body: %s", body)
	}
	if !strings.Contains(body, "Error Code:</strong> INTERNAL_ERROR") {
		t.Fatalf("expected error code in email body, got body: %s", body)
	}
	if !strings.Contains(body, "Error Type:</strong> internal_error") {
		t.Fatalf("expected error type in email body, got body: %s", body)
	}
	if !strings.Contains(body, "Error Description:</strong> boom") {
		t.Fatalf("expected error description in email body, got body: %s", body)
	}
}
