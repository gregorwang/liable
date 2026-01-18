package handlers

import (
	"comment-review-platform/internal/handlers/base"
	"comment-review-platform/internal/middleware"
	"comment-review-platform/internal/models"
	"comment-review-platform/internal/services"

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
	reviewerID := middleware.GetUserID(c)
	if reviewerID == 0 {
		base.RespondUnauthorized(c, "User not authenticated")
		return
	}

	var req models.ClaimQCTasksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.RespondBadRequest(c, base.ErrCodeInvalidRequest, err.Error())
		return
	}

	tasks, err := h.qcService.ClaimQCTasks(reviewerID, req.Count)
	if err != nil {
		base.RespondBadRequest(c, base.ErrCodeClaimFailed, err.Error())
		return
	}

	base.RespondSuccess(c, models.ClaimQCTasksResponse{
		Tasks: tasks,
		Count: len(tasks),
	})
}

// GetMyQCTasks gets the current user's quality check tasks
func (h *QualityCheckHandler) GetMyQCTasks(c *gin.Context) {
	reviewerID := middleware.GetUserID(c)
	if reviewerID == 0 {
		base.RespondUnauthorized(c, "User not authenticated")
		return
	}

	tasks, err := h.qcService.GetMyQCTasks(reviewerID)
	if err != nil {
		base.RespondInternalError(c, base.ErrCodeFetchFailed, err.Error())
		return
	}

	base.RespondSuccess(c, models.ClaimQCTasksResponse{
		Tasks: tasks,
		Count: len(tasks),
	})
}

// SubmitQCReview submits a single quality check result
func (h *QualityCheckHandler) SubmitQCReview(c *gin.Context) {
	reviewerID := middleware.GetUserID(c)
	if reviewerID == 0 {
		base.RespondUnauthorized(c, "User not authenticated")
		return
	}

	var req models.SubmitQCRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.RespondBadRequest(c, base.ErrCodeInvalidRequest, err.Error())
		return
	}

	if err := h.qcService.SubmitQCReview(reviewerID, req); err != nil {
		base.RespondBadRequest(c, base.ErrCodeSubmitFailed, err.Error())
		return
	}

	base.RespondSuccess(c, gin.H{"message": "Quality check submitted successfully"})
}

// SubmitBatchQCReviews submits multiple quality check results
func (h *QualityCheckHandler) SubmitBatchQCReviews(c *gin.Context) {
	reviewerID := middleware.GetUserID(c)
	if reviewerID == 0 {
		base.RespondUnauthorized(c, "User not authenticated")
		return
	}

	var req models.BatchSubmitQCRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.RespondBadRequest(c, base.ErrCodeInvalidRequest, err.Error())
		return
	}

	if err := h.qcService.SubmitBatchQCReviews(reviewerID, req.Reviews); err != nil {
		base.RespondBadRequest(c, base.ErrCodeSubmitFailed, err.Error())
		return
	}

	base.RespondSuccess(c, gin.H{
		"message": "Quality checks submitted successfully",
		"count":   len(req.Reviews),
	})
}

// ReturnQCTasks allows a reviewer to return quality check tasks
func (h *QualityCheckHandler) ReturnQCTasks(c *gin.Context) {
	reviewerID := middleware.GetUserID(c)
	if reviewerID == 0 {
		base.RespondUnauthorized(c, "User not authenticated")
		return
	}

	var req models.ReturnQCTasksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.RespondBadRequest(c, base.ErrCodeInvalidRequest, err.Error())
		return
	}

	returnedCount, err := h.qcService.ReturnQCTasks(reviewerID, req.TaskIDs)
	if err != nil {
		base.RespondBadRequest(c, base.ErrCodeReturnFailed, err.Error())
		return
	}

	base.RespondSuccess(c, gin.H{
		"message": "QC tasks returned successfully",
		"count":   returnedCount,
	})
}

// GetQCStats gets quality check statistics for the current user
func (h *QualityCheckHandler) GetQCStats(c *gin.Context) {
	reviewerID := middleware.GetUserID(c)
	if reviewerID == 0 {
		base.RespondUnauthorized(c, "User not authenticated")
		return
	}

	stats, err := h.qcService.GetQCStats(reviewerID)
	if err != nil {
		base.RespondInternalError(c, base.ErrCodeFetchFailed, err.Error())
		return
	}

	base.RespondSuccess(c, stats)
}
