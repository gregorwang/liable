package handlers

import (
	"comment-review-platform/internal/handlers/base"
	"comment-review-platform/internal/middleware"
	"comment-review-platform/internal/models"
	"comment-review-platform/internal/services"
	"database/sql"

	"github.com/gin-gonic/gin"
)

type AuditLogHandler struct {
	service *services.AuditLogService
}

func NewAuditLogHandler() *AuditLogHandler {
	return &AuditLogHandler{
		service: services.NewAuditLogService(),
	}
}

func (h *AuditLogHandler) ListLogs(c *gin.Context) {
	var req models.AuditLogQueryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		base.RespondBadRequest(c, base.ErrCodeInvalidRequest, "Invalid query parameters: "+err.Error())
		return
	}

	response, err := h.service.QueryLogs(req)
	if err != nil {
		base.RespondInternalError(c, base.ErrCodeFetchFailed, err.Error())
		return
	}

	base.RespondSuccess(c, response)
}

func (h *AuditLogHandler) GetLog(c *gin.Context) {
	id := c.Param("id")
	entry, err := h.service.GetLogByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			base.RespondNotFound(c, "Audit log not found")
			return
		}
		base.RespondInternalError(c, base.ErrCodeFetchFailed, err.Error())
		return
	}

	base.RespondSuccess(c, entry)
}

func (h *AuditLogHandler) ExportLogs(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		base.RespondUnauthorized(c, "User not authenticated")
		return
	}

	var req models.AuditLogExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.RespondBadRequest(c, base.ErrCodeInvalidRequest, "Invalid request: "+err.Error())
		return
	}

	username := middleware.GetUsername(c)
	role := middleware.GetRole(c)
	response, err := h.service.ExportLogs(userID, username, role, req)
	if err != nil {
		base.RespondBadRequest(c, base.ErrCodeFetchFailed, err.Error())
		return
	}

	base.RespondSuccess(c, response)
}

func (h *AuditLogHandler) ListExports(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		base.RespondUnauthorized(c, "User not authenticated")
		return
	}

	var req struct {
		Page     int `form:"page"`
		PageSize int `form:"page_size"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		base.RespondBadRequest(c, base.ErrCodeInvalidRequest, "Invalid query parameters: "+err.Error())
		return
	}

	response, err := h.service.ListExports(userID, req.Page, req.PageSize)
	if err != nil {
		base.RespondInternalError(c, base.ErrCodeFetchFailed, err.Error())
		return
	}

	base.RespondSuccess(c, response)
}
