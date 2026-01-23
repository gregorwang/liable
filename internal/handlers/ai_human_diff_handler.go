package handlers

import (
	"comment-review-platform/internal/handlers/base"
	"comment-review-platform/internal/middleware"
	"comment-review-platform/internal/models"
	"comment-review-platform/internal/services"

	"github.com/gin-gonic/gin"
)

type AIHumanDiffHandler struct {
	diffService *services.AIHumanDiffService
}

func NewAIHumanDiffHandler() *AIHumanDiffHandler {
	return &AIHumanDiffHandler{
		diffService: services.NewAIHumanDiffService(),
	}
}

// ClaimDiffTasks allows a reviewer to claim AI diff tasks with custom count
func (h *AIHumanDiffHandler) ClaimDiffTasks(c *gin.Context) {
	reviewerID := middleware.GetUserID(c)
	if reviewerID == 0 {
		base.RespondUnauthorized(c, "User not authenticated")
		return
	}

	var req models.ClaimAIHumanDiffTasksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.RespondBadRequest(c, base.ErrCodeInvalidRequest, "Invalid request: "+err.Error())
		return
	}

	tasks, err := h.diffService.ClaimDiffTasks(reviewerID, req.Count)
	if err != nil {
		base.RespondBadRequest(c, base.ErrCodeClaimFailed, err.Error())
		return
	}

	base.RespondSuccess(c, models.ClaimAIHumanDiffTasksResponse{
		Tasks: tasks,
		Count: len(tasks),
	})
}

// GetMyDiffTasks retrieves the current user's AI diff tasks
func (h *AIHumanDiffHandler) GetMyDiffTasks(c *gin.Context) {
	reviewerID := middleware.GetUserID(c)
	if reviewerID == 0 {
		base.RespondUnauthorized(c, "User not authenticated")
		return
	}

	tasks, err := h.diffService.GetMyDiffTasks(reviewerID)
	if err != nil {
		base.RespondInternalError(c, base.ErrCodeFetchFailed, err.Error())
		return
	}

	base.RespondSuccess(c, gin.H{
		"tasks": tasks,
		"count": len(tasks),
	})
}

// SubmitDiffReview submits a single AI diff review
func (h *AIHumanDiffHandler) SubmitDiffReview(c *gin.Context) {
	reviewerID := middleware.GetUserID(c)
	if reviewerID == 0 {
		base.RespondUnauthorized(c, "User not authenticated")
		return
	}

	var req models.SubmitAIHumanDiffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.RespondBadRequest(c, base.ErrCodeInvalidRequest, err.Error())
		return
	}

	if err := h.diffService.SubmitDiffReview(reviewerID, req); err != nil {
		base.RespondBadRequest(c, base.ErrCodeSubmitFailed, err.Error())
		return
	}

	base.RespondSuccess(c, gin.H{"message": "AI diff review submitted successfully"})
}

// SubmitBatchDiffReviews submits multiple AI diff reviews at once
func (h *AIHumanDiffHandler) SubmitBatchDiffReviews(c *gin.Context) {
	reviewerID := middleware.GetUserID(c)
	if reviewerID == 0 {
		base.RespondUnauthorized(c, "User not authenticated")
		return
	}

	var req models.BatchSubmitAIHumanDiffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.RespondBadRequest(c, base.ErrCodeInvalidRequest, err.Error())
		return
	}

	if err := h.diffService.SubmitBatchDiffReviews(reviewerID, req.Reviews); err != nil {
		base.RespondBadRequest(c, base.ErrCodeSubmitFailed, err.Error())
		return
	}

	base.RespondSuccess(c, gin.H{
		"message": "All AI diff reviews submitted successfully",
		"count":   len(req.Reviews),
	})
}

// ReturnDiffTasks allows a reviewer to return AI diff tasks back to the pool
func (h *AIHumanDiffHandler) ReturnDiffTasks(c *gin.Context) {
	reviewerID := middleware.GetUserID(c)
	if reviewerID == 0 {
		base.RespondUnauthorized(c, "User not authenticated")
		return
	}

	var req models.ReturnAIHumanDiffTasksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.RespondBadRequest(c, base.ErrCodeInvalidRequest, "Invalid request: "+err.Error())
		return
	}

	returnedCount, err := h.diffService.ReturnDiffTasks(reviewerID, req.TaskIDs)
	if err != nil {
		base.RespondBadRequest(c, base.ErrCodeReturnFailed, err.Error())
		return
	}

	base.RespondSuccess(c, gin.H{
		"message": "AI diff tasks returned successfully",
		"count":   returnedCount,
	})
}
