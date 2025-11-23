package handlers

import (
	"comment-review-platform/internal/models"
	"comment-review-platform/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type VideoTagHandler struct {
	service *services.VideoTagService
}

func NewVideoTagHandler() *VideoTagHandler {
	return &VideoTagHandler{
		service: services.NewVideoTagService(),
	}
}

// GetAllVideoTags retrieves all video quality tags
// GET /api/admin/video-tags
func (h *VideoTagHandler) GetAllVideoTags(c *gin.Context) {
	scope := c.Query("scope") // optional filter by scope

	var tags []models.VideoQualityTag
	var err error

	if scope != "" {
		tags, err = h.service.GetTagsByScope(scope)
	} else {
		tags, err = h.service.GetAllTags()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tags"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tags": tags})
}

// GetVideoTagsByQueue retrieves tags for a specific queue
// GET /api/video-tags/queue/:pool
func (h *VideoTagHandler) GetVideoTagsByQueue(c *gin.Context) {
	pool := c.Param("pool")

	tags, err := h.service.GetTagsByScopeAndQueue("video", pool)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tags"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tags": tags})
}

// CreateVideoTag creates a new video quality tag
// POST /api/admin/video-tags
func (h *VideoTagHandler) CreateVideoTag(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Category    string `json:"category" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tag := &models.VideoQualityTag{
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
	}

	if err := h.service.CreateTag(tag); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Tag created successfully",
		"tag":     tag,
	})
}

// UpdateVideoTag updates an existing video quality tag
// PUT /api/admin/video-tags/:id
func (h *VideoTagHandler) UpdateVideoTag(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tag ID"})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateTag(id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tag updated successfully"})
}

// DeleteVideoTag deletes a video quality tag
// DELETE /api/admin/video-tags/:id
func (h *VideoTagHandler) DeleteVideoTag(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tag ID"})
		return
	}

	if err := h.service.DeleteTag(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tag deleted successfully"})
}

// ToggleVideoTagActive toggles the active status of a tag
// PATCH /api/admin/video-tags/:id/toggle
func (h *VideoTagHandler) ToggleVideoTagActive(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tag ID"})
		return
	}

	if err := h.service.ToggleTagActive(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tag status toggled successfully"})
}
