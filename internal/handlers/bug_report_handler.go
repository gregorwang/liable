package handlers

import (
	"comment-review-platform/internal/handlers/base"
	"comment-review-platform/internal/middleware"
	"comment-review-platform/internal/models"
	"comment-review-platform/internal/services"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"
)

type BugReportHandler struct {
	bugReportService *services.BugReportService
}

func NewBugReportHandler() *BugReportHandler {
	return &BugReportHandler{
		bugReportService: services.NewBugReportService(),
	}
}

// Create handles bug report submissions with optional screenshots.
func (h *BugReportHandler) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		base.RespondUnauthorized(c, "User not authenticated")
		return
	}

	if err := c.Request.ParseMultipartForm(8 << 20); err != nil && !errors.Is(err, http.ErrNotMultipart) {
		base.RespondBadRequest(c, base.ErrCodeInvalidRequest, "表单解析失败")
		return
	}

	var files []*multipart.FileHeader
	if c.Request.MultipartForm != nil {
		files = c.Request.MultipartForm.File["screenshots"]
	}

	input := models.CreateBugReportInput{
		Title:        c.PostForm("title"),
		Description:  c.PostForm("description"),
		ErrorDetails: c.PostForm("error_details"),
		PageURL:      c.PostForm("page_url"),
		UserAgent:    c.GetHeader("User-Agent"),
	}
	if input.PageURL == "" {
		input.PageURL = c.GetHeader("X-Page-Url")
	}

	report, err := h.bugReportService.CreateBugReport(userID, input, files)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrBugReportLimitReached):
			base.RespondTooManyRequests(c, "每人最多提交3次错误反馈")
		case errors.Is(err, services.ErrTooManyScreenshots):
			base.RespondBadRequest(c, base.ErrCodeInvalidRequest, "每次最多上传2张截图")
		case errors.Is(err, services.ErrScreenshotTooLarge):
			base.RespondBadRequest(c, base.ErrCodeInvalidRequest, "每张截图大小需小于1MB")
		case errors.Is(err, services.ErrUnsupportedScreenshot):
			base.RespondBadRequest(c, base.ErrCodeInvalidRequest, "仅支持PNG/JPG/WEBP格式截图")
		case errors.Is(err, services.ErrBugReportStorageMissing):
			base.RespondInternalError(c, base.ErrCodeInternalError, "截图存储未配置")
		default:
			base.RespondInternalError(c, base.ErrCodeInternalError, err.Error())
		}
		return
	}

	base.RespondSuccess(c, gin.H{
		"message": "Bug反馈已提交",
		"report":  report,
	})
}

// List provides paginated bug reports for admin views.
func (h *BugReportHandler) List(c *gin.Context) {
	var req models.BugReportQueryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		base.RespondBadRequest(c, base.ErrCodeInvalidRequest, "Invalid query parameters: "+err.Error())
		return
	}

	response, err := h.bugReportService.ListBugReports(req)
	if err != nil {
		base.RespondBadRequest(c, base.ErrCodeInvalidRequest, err.Error())
		return
	}

	base.RespondSuccess(c, response)
}

// Export exports bug reports as a file (csv/json).
func (h *BugReportHandler) Export(c *gin.Context) {
	var req models.BugReportExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		base.RespondBadRequest(c, base.ErrCodeInvalidRequest, "Invalid request: "+err.Error())
		return
	}

	data, contentType, filename, err := h.bugReportService.ExportBugReports(req)
	if err != nil {
		base.RespondBadRequest(c, base.ErrCodeInvalidRequest, err.Error())
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(http.StatusOK, contentType, data)
}
