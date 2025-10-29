package handlers

import (
	"comment-review-platform/internal/middleware"
	"comment-review-platform/internal/models"
	"comment-review-platform/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type VideoHandler struct {
	videoService        *services.VideoService
	firstReviewService  *services.VideoFirstReviewService
	secondReviewService *services.VideoSecondReviewService
}

func NewVideoHandler() (*VideoHandler, error) {
	videoService, err := services.NewVideoService()
	if err != nil {
		return nil, err
	}

	return &VideoHandler{
		videoService:        videoService,
		firstReviewService:  services.NewVideoFirstReviewService(),
		secondReviewService: services.NewVideoSecondReviewService(),
	}, nil
}

// Admin endpoints

// ImportVideos imports videos from R2 bucket path
func (h *VideoHandler) ImportVideos(c *gin.Context) {
	var req models.ImportVideosRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	response, err := h.videoService.ImportVideosFromR2(req.R2PathPrefix)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ListVideos lists all videos with pagination
func (h *VideoHandler) ListVideos(c *gin.Context) {
	var req models.ListVideosRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters: " + err.Error()})
		return
	}

	// Set default pagination values
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	videos, total, err := h.videoService.ListVideos(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Calculate total pages
	totalPages := total / req.PageSize
	if total%req.PageSize > 0 {
		totalPages++
	}

	response := models.ListVideosResponse{
		Data:       videos,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}

	c.JSON(http.StatusOK, response)
}

// GetVideo gets video details by ID
func (h *VideoHandler) GetVideo(c *gin.Context) {
	videoID, err := getIntParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID"})
		return
	}

	video, err := h.videoService.GetVideoByID(videoID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
		return
	}

	c.JSON(http.StatusOK, video)
}

// GenerateVideoURL generates a pre-signed URL for video access
func (h *VideoHandler) GenerateVideoURL(c *gin.Context) {
	var req models.GenerateVideoURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	response, err := h.videoService.GenerateVideoURL(req.VideoID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// Reviewer endpoints

// ClaimVideoFirstReviewTasks allows a reviewer to claim first review tasks
func (h *VideoHandler) ClaimVideoFirstReviewTasks(c *gin.Context) {
	reviewerID := middleware.GetUserID(c)

	var req models.ClaimVideoFirstReviewTasksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	tasks, err := h.firstReviewService.ClaimFirstReviewTasks(reviewerID, req.Count)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.ClaimVideoFirstReviewTasksResponse{
		Tasks: tasks,
		Count: len(tasks),
	})
}

// GetMyVideoFirstReviewTasks retrieves the current user's first review tasks
func (h *VideoHandler) GetMyVideoFirstReviewTasks(c *gin.Context) {
	reviewerID := middleware.GetUserID(c)

	tasks, err := h.firstReviewService.GetMyFirstReviewTasks(reviewerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks": tasks,
		"count": len(tasks),
	})
}

// SubmitVideoFirstReview submits a single first review
func (h *VideoHandler) SubmitVideoFirstReview(c *gin.Context) {
	reviewerID := middleware.GetUserID(c)

	var req models.SubmitVideoFirstReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.firstReviewService.SubmitFirstReview(reviewerID, req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Video first review submitted successfully"})
}

// SubmitBatchVideoFirstReviews submits multiple first reviews at once
func (h *VideoHandler) SubmitBatchVideoFirstReviews(c *gin.Context) {
	reviewerID := middleware.GetUserID(c)

	var req models.BatchSubmitVideoFirstReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.firstReviewService.SubmitBatchFirstReviews(reviewerID, req.Reviews); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "All video first reviews submitted successfully",
		"count":   len(req.Reviews),
	})
}

// ReturnVideoFirstReviewTasks allows a reviewer to return first review tasks back to the pool
func (h *VideoHandler) ReturnVideoFirstReviewTasks(c *gin.Context) {
	reviewerID := middleware.GetUserID(c)

	var req models.ReturnVideoFirstReviewTasksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	returnedCount, err := h.firstReviewService.ReturnFirstReviewTasks(reviewerID, req.TaskIDs)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Video first review tasks returned successfully",
		"count":   returnedCount,
	})
}

// ClaimVideoSecondReviewTasks allows a reviewer to claim second review tasks
func (h *VideoHandler) ClaimVideoSecondReviewTasks(c *gin.Context) {
	reviewerID := middleware.GetUserID(c)

	var req models.ClaimVideoSecondReviewTasksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	tasks, err := h.secondReviewService.ClaimSecondReviewTasks(reviewerID, req.Count)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.ClaimVideoSecondReviewTasksResponse{
		Tasks: tasks,
		Count: len(tasks),
	})
}

// GetMyVideoSecondReviewTasks retrieves the current user's second review tasks
func (h *VideoHandler) GetMyVideoSecondReviewTasks(c *gin.Context) {
	reviewerID := middleware.GetUserID(c)

	tasks, err := h.secondReviewService.GetMySecondReviewTasks(reviewerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks": tasks,
		"count": len(tasks),
	})
}

// SubmitVideoSecondReview submits a single second review
func (h *VideoHandler) SubmitVideoSecondReview(c *gin.Context) {
	reviewerID := middleware.GetUserID(c)

	var req models.SubmitVideoSecondReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.secondReviewService.SubmitSecondReview(reviewerID, req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Video second review submitted successfully"})
}

// SubmitBatchVideoSecondReviews submits multiple second reviews at once
func (h *VideoHandler) SubmitBatchVideoSecondReviews(c *gin.Context) {
	reviewerID := middleware.GetUserID(c)

	var req models.BatchSubmitVideoSecondReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.secondReviewService.SubmitBatchSecondReviews(reviewerID, req.Reviews); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "All video second reviews submitted successfully",
		"count":   len(req.Reviews),
	})
}

// ReturnVideoSecondReviewTasks allows a reviewer to return second review tasks back to the pool
func (h *VideoHandler) ReturnVideoSecondReviewTasks(c *gin.Context) {
	reviewerID := middleware.GetUserID(c)

	var req models.ReturnVideoSecondReviewTasksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	returnedCount, err := h.secondReviewService.ReturnSecondReviewTasks(reviewerID, req.TaskIDs)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Video second review tasks returned successfully",
		"count":   returnedCount,
	})
}

// GetVideoQualityTags retrieves video quality tags by category
func (h *VideoHandler) GetVideoQualityTags(c *gin.Context) {
	var req models.GetVideoQualityTagsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters: " + err.Error()})
		return
	}

	tags, err := h.videoService.GetVideoQualityTags(req.Category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.GetVideoQualityTagsResponse{Tags: tags})
}

// TestVideoReviewDataStructure tests the video review data structure (for debugging)
func (h *VideoHandler) TestVideoReviewDataStructure(c *gin.Context) {
	var req models.SubmitVideoFirstReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Just return the parsed data to verify structure
	c.JSON(http.StatusOK, gin.H{
		"message": "Data structure parsed successfully",
		"data":    req,
	})
}

// Helper function to get integer parameter from URL
func getIntParam(c *gin.Context, param string) (int, error) {
	return strconv.Atoi(c.Param(param))
}
