package handlers

import (
	"comment-review-platform/internal/middleware"
	"comment-review-platform/internal/models"
	"comment-review-platform/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type VideoQueueHandler struct {
	videoQueueService *services.VideoQueueService
}

func NewVideoQueueHandler() *VideoQueueHandler {
	return &VideoQueueHandler{
		videoQueueService: services.NewVideoQueueService(),
	}
}

// ClaimTasks allows a reviewer to claim tasks from a specific pool
// POST /api/video/{pool}/tasks/claim
func (h *VideoQueueHandler) ClaimTasks(c *gin.Context) {
	reviewerID := middleware.GetUserID(c)
	pool := c.Param("pool")

	var req models.ClaimVideoQueueTasksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	tasks, err := h.videoQueueService.ClaimTasks(pool, reviewerID, req.Count)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.ClaimVideoQueueTasksResponse{
		Tasks: tasks,
		Count: len(tasks),
	})
}

// GetMyTasks retrieves the current user's tasks in a pool
// GET /api/video/{pool}/tasks/my
func (h *VideoQueueHandler) GetMyTasks(c *gin.Context) {
	reviewerID := middleware.GetUserID(c)
	pool := c.Param("pool")

	tasks, err := h.videoQueueService.GetMyTasks(pool, reviewerID)
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
// POST /api/video/{pool}/tasks/submit
func (h *VideoQueueHandler) SubmitReview(c *gin.Context) {
	reviewerID := middleware.GetUserID(c)
	pool := c.Param("pool")

	var req models.SubmitVideoQueueReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.videoQueueService.SubmitReview(pool, reviewerID, req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Video queue review submitted successfully"})
}

// SubmitBatchReviews submits multiple reviews at once
// POST /api/video/{pool}/tasks/submit-batch
func (h *VideoQueueHandler) SubmitBatchReviews(c *gin.Context) {
	reviewerID := middleware.GetUserID(c)
	pool := c.Param("pool")

	var req models.BatchSubmitVideoQueueReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.videoQueueService.SubmitBatchReviews(pool, reviewerID, req.Reviews); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "All video queue reviews submitted successfully",
		"count":   len(req.Reviews),
	})
}

// ReturnTasks allows a reviewer to return tasks back to the pool
// POST /api/video/{pool}/tasks/return
func (h *VideoQueueHandler) ReturnTasks(c *gin.Context) {
	reviewerID := middleware.GetUserID(c)
	pool := c.Param("pool")

	var req models.ReturnVideoQueueTasksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	returnedCount, err := h.videoQueueService.ReturnTasks(pool, reviewerID, req.TaskIDs)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Video queue tasks returned successfully",
		"count":   returnedCount,
	})
}

// GetTags retrieves available tags for a specific pool
// GET /api/video/{pool}/tags
func (h *VideoQueueHandler) GetTags(c *gin.Context) {
	pool := c.Param("pool")

	tags, err := h.videoQueueService.GetTags(pool)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.GetVideoQueueTagsResponse{Tags: tags})
}

// GetPoolStats retrieves statistics for a specific pool (admin only)
// GET /api/admin/video-queue/{pool}/stats
func (h *VideoQueueHandler) GetPoolStats(c *gin.Context) {
	pool := c.Param("pool")

	stats, err := h.videoQueueService.GetPoolStats(pool)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}
