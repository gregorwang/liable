package handlers

import (
	"comment-review-platform/internal/middleware"
	"comment-review-platform/internal/models"
	"comment-review-platform/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DocumentHandler struct {
	service           *services.DocumentService
	permissionService *services.PermissionService
}

func NewDocumentHandler() *DocumentHandler {
	return &DocumentHandler{
		service:           services.NewDocumentService(),
		permissionService: services.NewPermissionService(),
	}
}

func (h *DocumentHandler) ListDocuments(c *gin.Context) {
	userID := middleware.GetUserID(c)
	documents, err := h.service.ListDocuments()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	canEdit := false
	if userID > 0 {
		if allowed, err := h.permissionService.HasPermission(userID, "docs:edit"); err == nil {
			canEdit = allowed
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data":     documents,
		"can_edit": canEdit,
	})
}

func (h *DocumentHandler) UpdateDocument(c *gin.Context) {
	key := c.Param("key")
	var req models.UpdateSystemDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := middleware.GetUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	doc, err := h.service.UpdateDocument(key, req.Content, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, doc)
}
