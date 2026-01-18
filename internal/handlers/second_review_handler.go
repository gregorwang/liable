package handlers

import (
	"comment-review-platform/internal/handlers/base"
	"comment-review-platform/internal/middleware"
	"comment-review-platform/internal/models"
	"comment-review-platform/internal/services"

	"github.com/gin-gonic/gin"
)

type SecondReviewHandler struct {
	secondReviewService *services.SecondReviewService
}

func NewSecondReviewHandler() *SecondReviewHandler {
	return &SecondReviewHandler{
		secondReviewService: services.NewSecondReviewService(),
	}
}

// ClaimSecondReviewTasks allows a reviewer to claim second review tasks with custom count
func (h *SecondReviewHandler) ClaimSecondReviewTasks(c *gin.Context) {
	reviewerID := middleware.GetUserID(c)
	if reviewerID == 0 {
		base.RespondUnauthorized(c, "User not authenticated")
		return
	}

	var req models.ClaimSecondReviewTasksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.RespondBadRequest(c, base.ErrCodeInvalidRequest, "Invalid request: "+err.Error())
		return
	}

	tasks, err := h.secondReviewService.ClaimSecondReviewTasks(reviewerID, req.Count)
	if err != nil {
		base.RespondBadRequest(c, base.ErrCodeClaimFailed, err.Error())
		return
	}

	base.RespondSuccess(c, models.ClaimSecondReviewTasksResponse{
		Tasks: tasks,
		Count: len(tasks),
	})
}

// GetMySecondReviewTasks retrieves the current user's second review tasks
func (h *SecondReviewHandler) GetMySecondReviewTasks(c *gin.Context) {
	reviewerID := middleware.GetUserID(c)
	if reviewerID == 0 {
		base.RespondUnauthorized(c, "User not authenticated")
		return
	}

	tasks, err := h.secondReviewService.GetMySecondReviewTasks(reviewerID)
	if err != nil {
		base.RespondInternalError(c, base.ErrCodeFetchFailed, err.Error())
		return
	}

	base.RespondSuccess(c, gin.H{
		"tasks": tasks,
		"count": len(tasks),
	})
}

// SubmitSecondReview submits a single second review
func (h *SecondReviewHandler) SubmitSecondReview(c *gin.Context) {
	reviewerID := middleware.GetUserID(c)
	if reviewerID == 0 {
		base.RespondUnauthorized(c, "User not authenticated")
		return
	}

	var req models.SubmitSecondReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.RespondBadRequest(c, base.ErrCodeInvalidRequest, err.Error())
		return
	}

	if err := h.secondReviewService.SubmitSecondReview(reviewerID, req); err != nil {
		base.RespondBadRequest(c, base.ErrCodeSubmitFailed, err.Error())
		return
	}

	base.RespondSuccess(c, gin.H{"message": "Second review submitted successfully"})
}

// SubmitBatchSecondReviews submits multiple second reviews at once
func (h *SecondReviewHandler) SubmitBatchSecondReviews(c *gin.Context) {
	reviewerID := middleware.GetUserID(c)
	if reviewerID == 0 {
		base.RespondUnauthorized(c, "User not authenticated")
		return
	}

	var req models.BatchSubmitSecondReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.RespondBadRequest(c, base.ErrCodeInvalidRequest, err.Error())
		return
	}

	if err := h.secondReviewService.SubmitBatchSecondReviews(reviewerID, req.Reviews); err != nil {
		base.RespondBadRequest(c, base.ErrCodeSubmitFailed, err.Error())
		return
	}

	base.RespondSuccess(c, gin.H{
		"message": "All second reviews submitted successfully",
		"count":   len(req.Reviews),
	})
}

// ReturnSecondReviewTasks allows a reviewer to return second review tasks back to the pool
func (h *SecondReviewHandler) ReturnSecondReviewTasks(c *gin.Context) {
	reviewerID := middleware.GetUserID(c)
	if reviewerID == 0 {
		base.RespondUnauthorized(c, "User not authenticated")
		return
	}

	var req models.ReturnSecondReviewTasksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.RespondBadRequest(c, base.ErrCodeInvalidRequest, "Invalid request: "+err.Error())
		return
	}

	returnedCount, err := h.secondReviewService.ReturnSecondReviewTasks(reviewerID, req.TaskIDs)
	if err != nil {
		base.RespondBadRequest(c, base.ErrCodeReturnFailed, err.Error())
		return
	}

	base.RespondSuccess(c, gin.H{
		"message": "Second review tasks returned successfully",
		"count":   returnedCount,
	})
}
