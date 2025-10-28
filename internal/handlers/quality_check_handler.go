package handlers

import (
	"comment-review-platform/internal/models"
	"comment-review-platform/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type QualityCheckHandler struct {
	qcService *services.QualityCheckService
}

func NewQualityCheckHandler() *QualityCheckHandler {
	return &QualityCheckHandler{
		qcService: services.NewQualityCheckService(),
	}
}

// ClaimQCTasks allows a reviewer to claim quality check tasks
func (h *QualityCheckHandler) ClaimQCTasks(c *gin.Context) {
	var req models.ClaimQCTasksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get reviewer ID from context (set by auth middleware)
	reviewerID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	tasks, err := h.qcService.ClaimQCTasks(reviewerID.(int), req.Count)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response := models.ClaimQCTasksResponse{
		Tasks: tasks,
		Count: len(tasks),
	}

	c.JSON(http.StatusOK, response)
}

// GetMyQCTasks gets the current user's quality check tasks
func (h *QualityCheckHandler) GetMyQCTasks(c *gin.Context) {
	// Get reviewer ID from context
	reviewerID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	tasks, err := h.qcService.GetMyQCTasks(reviewerID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := models.ClaimQCTasksResponse{
		Tasks: tasks,
		Count: len(tasks),
	}

	c.JSON(http.StatusOK, response)
}

// SubmitQCReview submits a single quality check result
func (h *QualityCheckHandler) SubmitQCReview(c *gin.Context) {
	var req models.SubmitQCRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get reviewer ID from context
	reviewerID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	err := h.qcService.SubmitQCReview(reviewerID.(int), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Quality check submitted successfully"})
}

// SubmitBatchQCReviews submits multiple quality check results
func (h *QualityCheckHandler) SubmitBatchQCReviews(c *gin.Context) {
	var req models.BatchSubmitQCRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get reviewer ID from context
	reviewerID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	err := h.qcService.SubmitBatchQCReviews(reviewerID.(int), req.Reviews)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Quality checks submitted successfully",
		"submitted": len(req.Reviews),
	})
}

// ReturnQCTasks allows a reviewer to return quality check tasks
func (h *QualityCheckHandler) ReturnQCTasks(c *gin.Context) {
	var req models.ReturnQCTasksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get reviewer ID from context
	reviewerID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	returnedCount, err := h.qcService.ReturnQCTasks(reviewerID.(int), req.TaskIDs)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "QC tasks returned successfully",
		"count":   returnedCount,
	})
}

// GetQCStats gets quality check statistics for the current user
func (h *QualityCheckHandler) GetQCStats(c *gin.Context) {
	// Get reviewer ID from context
	reviewerID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	stats, err := h.qcService.GetQCStats(reviewerID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}
