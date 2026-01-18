package base

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestRespondError(t *testing.T) {
	tests := []struct {
		name           string
		status         int
		code           string
		message        string
		expectedStatus int
	}{
		{
			name:           "bad request error",
			status:         http.StatusBadRequest,
			code:           ErrCodeInvalidRequest,
			message:        "Invalid input",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "internal server error",
			status:         http.StatusInternalServerError,
			code:           ErrCodeInternalError,
			message:        "Database connection failed",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "unauthorized error",
			status:         http.StatusUnauthorized,
			code:           ErrCodeUnauthorized,
			message:        "Token expired",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			RespondError(c, tt.status, tt.code, tt.message)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var resp ErrorResponse
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			if resp.Error != tt.message {
				t.Errorf("expected error message %q, got %q", tt.message, resp.Error)
			}

			if resp.Code != tt.code {
				t.Errorf("expected error code %q, got %q", tt.code, resp.Code)
			}
		})
	}
}

func TestRespondSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	data := gin.H{"message": "success", "count": 5}
	RespondSuccess(c, data)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp["message"] != "success" {
		t.Errorf("expected message 'success', got %v", resp["message"])
	}

	if resp["count"] != float64(5) {
		t.Errorf("expected count 5, got %v", resp["count"])
	}
}

func TestRespondBadRequest(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	RespondBadRequest(c, ErrCodeInvalidRequest, "Missing required field")

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Code != ErrCodeInvalidRequest {
		t.Errorf("expected code %q, got %q", ErrCodeInvalidRequest, resp.Code)
	}
}

func TestRespondInternalError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	RespondInternalError(c, ErrCodeInternalError, "Database error")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Code != ErrCodeInternalError {
		t.Errorf("expected code %q, got %q", ErrCodeInternalError, resp.Code)
	}
}

func TestRespondUnauthorized(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	RespondUnauthorized(c, "User not authenticated")

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Code != ErrCodeUnauthorized {
		t.Errorf("expected code %q, got %q", ErrCodeUnauthorized, resp.Code)
	}
}

func TestRespondNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	RespondNotFound(c, "Resource not found")

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Code != ErrCodeNotFound {
		t.Errorf("expected code %q, got %q", ErrCodeNotFound, resp.Code)
	}
}

func TestRespondTooManyRequests(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	RespondTooManyRequests(c, "Rate limit exceeded")

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("expected status %d, got %d", http.StatusTooManyRequests, w.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Code != ErrCodeRateLimitExceeded {
		t.Errorf("expected code %q, got %q", ErrCodeRateLimitExceeded, resp.Code)
	}
}

// TestErrorResponseFormat 验证错误响应格式符合 Requirements 6.1
func TestErrorResponseFormat(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	RespondError(c, http.StatusBadRequest, "TEST_CODE", "Test message")

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// 验证响应包含 error 和 code 字段
	if _, ok := resp["error"]; !ok {
		t.Error("response should contain 'error' field")
	}

	if _, ok := resp["code"]; !ok {
		t.Error("response should contain 'code' field")
	}

	// 验证没有多余字段
	if len(resp) != 2 {
		t.Errorf("expected 2 fields in response, got %d", len(resp))
	}
}
