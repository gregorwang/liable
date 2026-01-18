// Package base provides generic handler functions for task operations.
package base

import (
	"comment-review-platform/internal/middleware"

	"github.com/gin-gonic/gin"
)

// ClaimTasksRequest 领取任务请求
type ClaimTasksRequest struct {
	Count int `json:"count" binding:"required,min=1,max=50"`
}

// ReturnTasksRequest 退回任务请求
type ReturnTasksRequest struct {
	TaskIDs []int `json:"task_ids" binding:"required,min=1"`
}

// HandleClaimTasks 通用任务领取处理
// T: 任务类型
// service: 必须实现 ClaimTasks(reviewerID int, count int) ([]T, error) 方法
func HandleClaimTasks[T any](
	c *gin.Context,
	service interface{ ClaimTasks(int, int) ([]T, error) },
	config TaskHandlerConfig,
) {
	reviewerID := middleware.GetUserID(c)
	if reviewerID == 0 {
		RespondUnauthorized(c, "User not authenticated")
		return
	}

	var req ClaimTasksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondBadRequest(c, ErrCodeInvalidRequest, "Invalid request: "+err.Error())
		return
	}

	// 验证领取数量范围
	if req.Count < config.ClaimCountMin || req.Count > config.ClaimCountMax {
		RespondBadRequest(c, ErrCodeInvalidRequest,
			"Claim count must be between "+string(rune(config.ClaimCountMin+'0'))+" and 50")
		return
	}

	tasks, err := service.ClaimTasks(reviewerID, req.Count)
	if err != nil {
		RespondBadRequest(c, ErrCodeClaimFailed, err.Error())
		return
	}

	RespondSuccess(c, gin.H{"tasks": tasks, "count": len(tasks)})
}

// HandleGetMyTasks 通用获取我的任务处理
// T: 任务类型
// service: 必须实现 GetMyTasks(reviewerID int) ([]T, error) 方法
func HandleGetMyTasks[T any](
	c *gin.Context,
	service interface{ GetMyTasks(int) ([]T, error) },
) {
	reviewerID := middleware.GetUserID(c)
	if reviewerID == 0 {
		RespondUnauthorized(c, "User not authenticated")
		return
	}

	tasks, err := service.GetMyTasks(reviewerID)
	if err != nil {
		RespondInternalError(c, ErrCodeFetchFailed, err.Error())
		return
	}

	RespondSuccess(c, gin.H{"tasks": tasks, "count": len(tasks)})
}

// HandleSubmitReview 通用提交审核处理
// R: 提交请求类型
// service: 必须实现 SubmitReview(reviewerID int, req R) error 方法
func HandleSubmitReview[R any](
	c *gin.Context,
	service interface{ SubmitReview(int, R) error },
	successMsg string,
) {
	reviewerID := middleware.GetUserID(c)
	if reviewerID == 0 {
		RespondUnauthorized(c, "User not authenticated")
		return
	}

	var req R
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondBadRequest(c, ErrCodeInvalidRequest, "Invalid request: "+err.Error())
		return
	}

	if err := service.SubmitReview(reviewerID, req); err != nil {
		RespondBadRequest(c, ErrCodeSubmitFailed, err.Error())
		return
	}

	RespondSuccess(c, gin.H{"message": successMsg})
}

// HandleBatchSubmit 通用批量提交处理
// R: 单个提交请求类型
// service: 必须实现 SubmitBatchReviews(reviewerID int, reviews []R) error 方法
func HandleBatchSubmit[R any](
	c *gin.Context,
	service interface{ SubmitBatchReviews(int, []R) error },
	successMsg string,
) {
	reviewerID := middleware.GetUserID(c)
	if reviewerID == 0 {
		RespondUnauthorized(c, "User not authenticated")
		return
	}

	var req struct {
		Reviews []R `json:"reviews" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondBadRequest(c, ErrCodeInvalidRequest, "Invalid request: "+err.Error())
		return
	}

	if err := service.SubmitBatchReviews(reviewerID, req.Reviews); err != nil {
		RespondBadRequest(c, ErrCodeSubmitFailed, err.Error())
		return
	}

	RespondSuccess(c, gin.H{"message": successMsg, "count": len(req.Reviews)})
}

// HandleReturnTasks 通用退回任务处理
// service: 必须实现 ReturnTasks(reviewerID int, taskIDs []int) (int, error) 方法
func HandleReturnTasks(
	c *gin.Context,
	service interface{ ReturnTasks(int, []int) (int, error) },
	successMsg string,
) {
	reviewerID := middleware.GetUserID(c)
	if reviewerID == 0 {
		RespondUnauthorized(c, "User not authenticated")
		return
	}

	var req ReturnTasksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondBadRequest(c, ErrCodeInvalidRequest, "Invalid request: "+err.Error())
		return
	}

	count, err := service.ReturnTasks(reviewerID, req.TaskIDs)
	if err != nil {
		RespondBadRequest(c, ErrCodeReturnFailed, err.Error())
		return
	}

	RespondSuccess(c, gin.H{"message": successMsg, "count": count})
}
