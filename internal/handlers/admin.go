package handlers

import (
	"comment-review-platform/internal/models"
	"comment-review-platform/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	adminService      *services.AdminService
	statsService      *services.StatsService
	permissionService *services.PermissionService
}

func NewAdminHandler() *AdminHandler {
	return &AdminHandler{
		adminService:      services.NewAdminService(),
		statsService:      services.NewStatsService(),
		permissionService: services.NewPermissionService(),
	}
}

// GetPendingUsers retrieves all users pending approval
func (h *AdminHandler) GetPendingUsers(c *gin.Context) {
	users, err := h.adminService.GetPendingUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users, "count": len(users)})
}

// GetAllUsers retrieves all users (for permission management)
func (h *AdminHandler) GetAllUsers(c *gin.Context) {
	users, err := h.adminService.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users, "count": len(users)})
}

// ApproveUser approves or rejects a user
func (h *AdminHandler) ApproveUser(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req models.ApproveUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.adminService.ApproveUser(userID, req.Status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User status updated successfully"})
}

// GetOverviewStats retrieves overall statistics
func (h *AdminHandler) GetOverviewStats(c *gin.Context) {
	stats, err := h.statsService.GetOverviewStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetTodayReviewStats retrieves same-day review counts
func (h *AdminHandler) GetTodayReviewStats(c *gin.Context) {
	stats, err := h.statsService.GetTodayReviewStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetHourlyStats retrieves hourly statistics
func (h *AdminHandler) GetHourlyStats(c *gin.Context) {
	date := c.DefaultQuery("date", "")
	if date == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Date parameter is required (YYYY-MM-DD)"})
		return
	}

	stats, err := h.statsService.GetHourlyStats(date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetTagStats retrieves tag statistics
func (h *AdminHandler) GetTagStats(c *gin.Context) {
	stats, err := h.statsService.GetTagStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tags": stats})
}

// GetReviewerPerformance retrieves reviewer performance statistics
func (h *AdminHandler) GetReviewerPerformance(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	performances, err := h.statsService.GetReviewerPerformance(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"reviewers": performances})
}

// GetVideoQualityTagStats returns video quality tag statistics
func (h *AdminHandler) GetVideoQualityTagStats(c *gin.Context) {
	stats, err := h.statsService.GetVideoQualityTagStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"stats": stats})
}

// GetVideoQualityAnalysis returns comprehensive video quality analysis
func (h *AdminHandler) GetVideoQualityAnalysis(c *gin.Context) {
	analysis, err := h.statsService.GetVideoQualityAnalysis()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, analysis)
}

// GetAllTags retrieves all tags (including inactive)
func (h *AdminHandler) GetAllTags(c *gin.Context) {
	tags, err := h.adminService.GetAllTags()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tags": tags})
}

// CreateTag creates a new tag
func (h *AdminHandler) CreateTag(c *gin.Context) {
	var req models.CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tag, err := h.adminService.CreateTag(req.Name, req.Description)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, tag)
}

// UpdateTag updates a tag
func (h *AdminHandler) UpdateTag(c *gin.Context) {
	tagID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tag ID"})
		return
	}

	var req models.UpdateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.adminService.UpdateTag(tagID, req.Name, req.Description, req.IsActive); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tag updated successfully"})
}

// DeleteTag deletes a tag
func (h *AdminHandler) DeleteTag(c *gin.Context) {
	tagID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tag ID"})
		return
	}

	if err := h.adminService.DeleteTag(tagID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tag deleted successfully"})
}

// ==================== Task Queue Management ====================

type TaskQueueHandler struct {
	queueService *services.TaskQueueService
}

func NewTaskQueueHandler() *TaskQueueHandler {
	return &TaskQueueHandler{
		queueService: services.NewTaskQueueService(),
	}
}

// CreateTaskQueue creates a new task queue
func (h *TaskQueueHandler) CreateTaskQueue(c *gin.Context) {
	userID := c.GetInt("user_id")

	var req models.CreateTaskQueueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queue, err := h.queueService.CreateTaskQueue(req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, queue)
}

// ListTaskQueues retrieves task queues with pagination
func (h *TaskQueueHandler) ListTaskQueues(c *gin.Context) {
	var req models.ListTaskQueuesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.queueService.ListTaskQueues(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetTaskQueue retrieves a single task queue by ID
func (h *TaskQueueHandler) GetTaskQueue(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid queue ID"})
		return
	}

	queue, err := h.queueService.GetTaskQueueByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task queue not found"})
		return
	}

	c.JSON(http.StatusOK, queue)
}

// UpdateTaskQueue updates a task queue
func (h *TaskQueueHandler) UpdateTaskQueue(c *gin.Context) {
	userID := c.GetInt("user_id")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid queue ID"})
		return
	}

	var req models.UpdateTaskQueueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	queue, err := h.queueService.UpdateTaskQueue(id, req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, queue)
}

// DeleteTaskQueue deletes a task queue
func (h *TaskQueueHandler) DeleteTaskQueue(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid queue ID"})
		return
	}

	if err := h.queueService.DeleteTaskQueue(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task queue deleted successfully"})
}

// GetAllTaskQueues retrieves all active task queues
func (h *TaskQueueHandler) GetAllTaskQueues(c *gin.Context) {
	queues, err := h.queueService.GetAllTaskQueues()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"queues": queues, "count": len(queues)})
}

// GetPublicQueues retrieves all task queues (public, no auth required)
func (h *TaskQueueHandler) GetPublicQueues(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("page_size", "10")
	search := c.DefaultQuery("search", "")

	pageNum := 1
	if p, err := strconv.Atoi(page); err == nil {
		pageNum = p
	}

	size := 10
	if ps, err := strconv.Atoi(pageSize); err == nil {
		size = ps
	}

	var req models.ListTaskQueuesRequest
	req.Page = pageNum
	req.PageSize = size
	req.Search = search

	response, err := h.queueService.ListTaskQueues(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetPublicQueue retrieves a single task queue by ID (public, no auth required)
func (h *TaskQueueHandler) GetPublicQueue(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid queue ID"})
		return
	}

	queue, err := h.queueService.GetTaskQueueByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task queue not found"})
		return
	}

	c.JSON(http.StatusOK, queue)
}

// ListTaskQueuesForReviewers retrieves task queues for reviewer users (public read-only)
func (h *TaskQueueHandler) ListTaskQueuesForReviewers(c *gin.Context) {
	var req models.ListTaskQueuesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.queueService.ListTaskQueues(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetTaskQueueForReviewers retrieves a single task queue by ID for reviewer users (public read-only)
func (h *TaskQueueHandler) GetTaskQueueForReviewers(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid queue ID"})
		return
	}

	queue, err := h.queueService.GetTaskQueueByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task queue not found"})
		return
	}

	c.JSON(http.StatusOK, queue)
}

// ==================== Permission Management ====================

// ListPermissions retrieves permissions with pagination and filtering
func (h *AdminHandler) ListPermissions(c *gin.Context) {
	var req models.ListPermissionsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set defaults
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}

	response, err := h.permissionService.ListPermissions(req.Resource, req.Category, req.Search, req.Page, req.PageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetAllPermissions retrieves all permissions (no pagination)
func (h *AdminHandler) GetAllPermissions(c *gin.Context) {
	permissions, err := h.permissionService.GetAllPermissions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"permissions": permissions})
}

// GetUserPermissions retrieves permissions for a specific user
func (h *AdminHandler) GetUserPermissions(c *gin.Context) {
	userID, err := strconv.Atoi(c.Query("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id parameter"})
		return
	}

	permissions, err := h.permissionService.GetUserPermissions(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.UserPermissionsResponse{
		UserID:      userID,
		Permissions: permissions,
	})
}

// GrantPermissions grants permissions to a user
func (h *AdminHandler) GrantPermissions(c *gin.Context) {
	// Get current admin user ID
	adminUserID := c.GetInt("user_id")

	var req models.GrantPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Grant permissions
	err := h.permissionService.GrantPermissions(req.UserID, req.PermissionKeys, adminUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Permissions granted successfully",
		"user_id":     req.UserID,
		"permissions": req.PermissionKeys,
	})
}

// RevokePermissions revokes permissions from a user
func (h *AdminHandler) RevokePermissions(c *gin.Context) {
	var req models.RevokePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Revoke permissions
	err := h.permissionService.RevokePermissions(req.UserID, req.PermissionKeys)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Permissions revoked successfully",
		"user_id":     req.UserID,
		"permissions": req.PermissionKeys,
	})
}
