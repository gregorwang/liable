package handlers

import (
	"comment-review-platform/internal/middleware"
	"comment-review-platform/internal/models"
	"comment-review-platform/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TaskHandler struct {
	taskService *services.TaskService
}

func NewTaskHandler() *TaskHandler {
	return &TaskHandler{
		taskService: services.NewTaskService(),
	}
}

// ClaimTasks allows a reviewer to claim tasks with custom count
func (h *TaskHandler) ClaimTasks(c *gin.Context) {
	reviewerID := middleware.GetUserID(c)

	var req models.ClaimTasksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	tasks, err := h.taskService.ClaimTasks(reviewerID, req.Count)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.ClaimTasksResponse{
		Tasks: tasks,
		Count: len(tasks),
	})
}

// GetMyTasks retrieves the current user's tasks
func (h *TaskHandler) GetMyTasks(c *gin.Context) {
	reviewerID := middleware.GetUserID(c)

	tasks, err := h.taskService.GetMyTasks(reviewerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks": tasks,
		"count": len(tasks),
	})
}

// SubmitReview submits a single review
func (h *TaskHandler) SubmitReview(c *gin.Context) {
	reviewerID := middleware.GetUserID(c)

	var req models.SubmitReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.taskService.SubmitReview(reviewerID, req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Review submitted successfully"})
}

// SubmitBatchReviews submits multiple reviews at once
func (h *TaskHandler) SubmitBatchReviews(c *gin.Context) {
	reviewerID := middleware.GetUserID(c)

	var req models.BatchSubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.taskService.SubmitBatchReviews(reviewerID, req.Reviews); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "All reviews submitted successfully",
		"count":   len(req.Reviews),
	})
}

// ReturnTasks allows a reviewer to return tasks back to the pool
func (h *TaskHandler) ReturnTasks(c *gin.Context) {
	reviewerID := middleware.GetUserID(c)

	var req models.ReturnTasksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	returnedCount, err := h.taskService.ReturnTasks(reviewerID, req.TaskIDs)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tasks returned successfully",
		"count":   returnedCount,
	})
}

// GetActiveTags retrieves all active tags
func (h *TaskHandler) GetActiveTags(c *gin.Context) {
	tags, err := h.taskService.GetActiveTags()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tags": tags})
}

// SearchTasks searches review tasks with filters
func (h *TaskHandler) SearchTasks(c *gin.Context) {
	var req models.SearchTasksRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters: " + err.Error()})
		return
	}

	response, err := h.taskService.SearchTasks(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
