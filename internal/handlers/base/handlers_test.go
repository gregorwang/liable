package base

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// Mock service for testing
type mockTaskService struct {
	claimTasksFunc       func(reviewerID, count int) ([]mockTask, error)
	getMyTasksFunc       func(reviewerID int) ([]mockTask, error)
	submitReviewFunc     func(reviewerID int, req mockSubmitRequest) error
	submitBatchFunc      func(reviewerID int, reviews []mockSubmitRequest) error
	returnTasksFunc      func(reviewerID int, taskIDs []int) (int, error)
}

type mockTask struct {
	ID     int    `json:"id"`
	Status string `json:"status"`
}

type mockSubmitRequest struct {
	TaskID   int    `json:"task_id"`
	Decision string `json:"decision"`
}

func (m *mockTaskService) ClaimTasks(reviewerID, count int) ([]mockTask, error) {
	if m.claimTasksFunc != nil {
		return m.claimTasksFunc(reviewerID, count)
	}
	return []mockTask{{ID: 1, Status: "in_progress"}}, nil
}

func (m *mockTaskService) GetMyTasks(reviewerID int) ([]mockTask, error) {
	if m.getMyTasksFunc != nil {
		return m.getMyTasksFunc(reviewerID)
	}
	return []mockTask{{ID: 1, Status: "in_progress"}}, nil
}

func (m *mockTaskService) SubmitReview(reviewerID int, req mockSubmitRequest) error {
	if m.submitReviewFunc != nil {
		return m.submitReviewFunc(reviewerID, req)
	}
	return nil
}

func (m *mockTaskService) SubmitBatchReviews(reviewerID int, reviews []mockSubmitRequest) error {
	if m.submitBatchFunc != nil {
		return m.submitBatchFunc(reviewerID, reviews)
	}
	return nil
}

func (m *mockTaskService) ReturnTasks(reviewerID int, taskIDs []int) (int, error) {
	if m.returnTasksFunc != nil {
		return m.returnTasksFunc(reviewerID, taskIDs)
	}
	return len(taskIDs), nil
}

// Helper to create test context with user_id
func createTestContext(w *httptest.ResponseRecorder, body []byte, userID int) *gin.Context {
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	if userID > 0 {
		c.Set("user_id", userID)
	}
	return c
}

func TestHandleClaimTasks_Success(t *testing.T) {
	w := httptest.NewRecorder()
	body, _ := json.Marshal(ClaimTasksRequest{Count: 5})
	c := createTestContext(w, body, 1)

	service := &mockTaskService{
		claimTasksFunc: func(reviewerID, count int) ([]mockTask, error) {
			if reviewerID != 1 {
				t.Errorf("expected reviewerID 1, got %d", reviewerID)
			}
			if count != 5 {
				t.Errorf("expected count 5, got %d", count)
			}
			return []mockTask{{ID: 1}, {ID: 2}}, nil
		},
	}

	config := DefaultTaskHandlerConfig("test")
	HandleClaimTasks(c, service, config)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["count"] != float64(2) {
		t.Errorf("expected count 2, got %v", resp["count"])
	}
}

func TestHandleClaimTasks_Unauthorized(t *testing.T) {
	w := httptest.NewRecorder()
	body, _ := json.Marshal(ClaimTasksRequest{Count: 5})
	c := createTestContext(w, body, 0) // No user_id

	service := &mockTaskService{}
	config := DefaultTaskHandlerConfig("test")
	HandleClaimTasks(c, service, config)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestHandleClaimTasks_InvalidRequest(t *testing.T) {
	w := httptest.NewRecorder()
	body := []byte(`{"count": "invalid"}`) // Invalid JSON
	c := createTestContext(w, body, 1)

	service := &mockTaskService{}
	config := DefaultTaskHandlerConfig("test")
	HandleClaimTasks(c, service, config)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Code != ErrCodeInvalidRequest {
		t.Errorf("expected code %q, got %q", ErrCodeInvalidRequest, resp.Code)
	}
}

func TestHandleClaimTasks_ServiceError(t *testing.T) {
	w := httptest.NewRecorder()
	body, _ := json.Marshal(ClaimTasksRequest{Count: 5})
	c := createTestContext(w, body, 1)

	service := &mockTaskService{
		claimTasksFunc: func(reviewerID, count int) ([]mockTask, error) {
			return nil, errors.New("no tasks available")
		},
	}

	config := DefaultTaskHandlerConfig("test")
	HandleClaimTasks(c, service, config)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var resp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Code != ErrCodeClaimFailed {
		t.Errorf("expected code %q, got %q", ErrCodeClaimFailed, resp.Code)
	}
}

func TestHandleGetMyTasks_Success(t *testing.T) {
	w := httptest.NewRecorder()
	c := createTestContext(w, nil, 1)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

	service := &mockTaskService{
		getMyTasksFunc: func(reviewerID int) ([]mockTask, error) {
			return []mockTask{{ID: 1}, {ID: 2}, {ID: 3}}, nil
		},
	}

	HandleGetMyTasks(c, service)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["count"] != float64(3) {
		t.Errorf("expected count 3, got %v", resp["count"])
	}
}

func TestHandleGetMyTasks_ServiceError(t *testing.T) {
	w := httptest.NewRecorder()
	c := createTestContext(w, nil, 1)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

	service := &mockTaskService{
		getMyTasksFunc: func(reviewerID int) ([]mockTask, error) {
			return nil, errors.New("database error")
		},
	}

	HandleGetMyTasks(c, service)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestHandleSubmitReview_Success(t *testing.T) {
	w := httptest.NewRecorder()
	body, _ := json.Marshal(mockSubmitRequest{TaskID: 1, Decision: "approve"})
	c := createTestContext(w, body, 1)

	service := &mockTaskService{
		submitReviewFunc: func(reviewerID int, req mockSubmitRequest) error {
			if req.TaskID != 1 {
				t.Errorf("expected taskID 1, got %d", req.TaskID)
			}
			return nil
		},
	}

	HandleSubmitReview(c, service, "Review submitted")

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["message"] != "Review submitted" {
		t.Errorf("expected message 'Review submitted', got %v", resp["message"])
	}
}

func TestHandleReturnTasks_Success(t *testing.T) {
	w := httptest.NewRecorder()
	body, _ := json.Marshal(ReturnTasksRequest{TaskIDs: []int{1, 2, 3}})
	c := createTestContext(w, body, 1)

	service := &mockTaskService{
		returnTasksFunc: func(reviewerID int, taskIDs []int) (int, error) {
			return len(taskIDs), nil
		},
	}

	HandleReturnTasks(c, service, "Tasks returned")

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["count"] != float64(3) {
		t.Errorf("expected count 3, got %v", resp["count"])
	}
}

func TestHandleBatchSubmit_Success(t *testing.T) {
	w := httptest.NewRecorder()
	body, _ := json.Marshal(map[string]interface{}{
		"reviews": []mockSubmitRequest{
			{TaskID: 1, Decision: "approve"},
			{TaskID: 2, Decision: "reject"},
		},
	})
	c := createTestContext(w, body, 1)

	service := &mockTaskService{
		submitBatchFunc: func(reviewerID int, reviews []mockSubmitRequest) error {
			if len(reviews) != 2 {
				t.Errorf("expected 2 reviews, got %d", len(reviews))
			}
			return nil
		},
	}

	HandleBatchSubmit(c, service, "Batch submitted")

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["count"] != float64(2) {
		t.Errorf("expected count 2, got %v", resp["count"])
	}
}

func TestDefaultTaskHandlerConfig(t *testing.T) {
	config := DefaultTaskHandlerConfig("video_review")

	if config.TaskTypeName != "video_review" {
		t.Errorf("expected TaskTypeName 'video_review', got %q", config.TaskTypeName)
	}

	if config.ClaimCountMin != 1 {
		t.Errorf("expected ClaimCountMin 1, got %d", config.ClaimCountMin)
	}

	if config.ClaimCountMax != 50 {
		t.Errorf("expected ClaimCountMax 50, got %d", config.ClaimCountMax)
	}
}
