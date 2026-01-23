package services

import (
	"bytes"
	"comment-review-platform/internal/config"
	"comment-review-platform/internal/models"
	"comment-review-platform/internal/repository"
	"comment-review-platform/pkg/r2"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	bugReportMaxPerUser        = 3
	bugReportMaxScreenshots    = 2
	bugReportMaxScreenshotSize = 1 << 20 // 1MB
	bugReportPreviewExpiration = 24 * time.Hour
	bugReportExportMaxRows     = 5000
)

var (
	ErrBugReportLimitReached   = errors.New("bug report limit reached")
	ErrTooManyScreenshots      = errors.New("too many screenshots")
	ErrScreenshotTooLarge      = errors.New("screenshot too large")
	ErrUnsupportedScreenshot   = errors.New("unsupported screenshot type")
	ErrBugReportStorageMissing = errors.New("R2 storage is not configured")
)

type BugReportService struct {
	repo             *repository.BugReportRepository
	r2               *r2.R2Service
	screenshotPrefix string
}

func NewBugReportService() *BugReportService {
	r2Service, err := r2.NewR2Service()
	if err != nil {
		r2Service = nil
	}

	return &BugReportService{
		repo:             repository.NewBugReportRepository(),
		r2:               r2Service,
		screenshotPrefix: normalizeBugReportPrefix(config.AppConfig.R2BugReportPathPrefix),
	}
}

func (s *BugReportService) CreateBugReport(userID int, input models.CreateBugReportInput, files []*multipart.FileHeader) (*models.BugReport, error) {
	description := strings.TrimSpace(input.Description)
	if description == "" {
		return nil, errors.New("请填写问题描述")
	}

	if len(files) > bugReportMaxScreenshots {
		return nil, ErrTooManyScreenshots
	}

	count, err := s.repo.CountByUser(userID)
	if err != nil {
		return nil, err
	}
	if count >= bugReportMaxPerUser {
		return nil, ErrBugReportLimitReached
	}

	if len(files) > 0 && s.r2 == nil {
		return nil, ErrBugReportStorageMissing
	}

	screenshots := make([]models.BugReportScreenshot, 0, len(files))
	for _, fileHeader := range files {
		screenshot, err := s.uploadScreenshot(userID, fileHeader)
		if err != nil {
			return nil, err
		}
		screenshots = append(screenshots, screenshot)
	}

	report := models.BugReport{
		UserID:       userID,
		Title:        strings.TrimSpace(input.Title),
		Description:  description,
		ErrorDetails: strings.TrimSpace(input.ErrorDetails),
		PageURL:      strings.TrimSpace(input.PageURL),
		UserAgent:    strings.TrimSpace(input.UserAgent),
		Screenshots:  screenshots,
	}

	if err := s.repo.Create(&report); err != nil {
		return nil, err
	}

	return &report, nil
}

func (s *BugReportService) ListBugReports(req models.BugReportQueryRequest) (*models.BugReportListResponse, error) {
	page := req.Page
	pageSize := req.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	filters, err := parseBugReportFilters(req.StartTime, req.EndTime, req.UserID, req.Username, req.Keyword)
	if err != nil {
		return nil, err
	}

	records, total, err := s.repo.ListWithFilters(filters, page, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]models.BugReportAdminItem, 0, len(records))
	for _, record := range records {
		shots := make([]models.BugReportScreenshotView, 0, len(record.Screenshots))
		for _, shot := range record.Screenshots {
			view := models.BugReportScreenshotView{
				Key:         shot.Key,
				Filename:    shot.Filename,
				Size:        shot.Size,
				ContentType: shot.ContentType,
			}
			if s.r2 != nil && shot.Key != "" {
				if url, err := s.r2.GeneratePresignedURL(shot.Key, bugReportPreviewExpiration); err == nil {
					view.URL = url
				}
			}
			shots = append(shots, view)
		}

		items = append(items, models.BugReportAdminItem{
			ID:           record.ID,
			UserID:       record.UserID,
			Username:     record.Username,
			Title:        record.Title,
			Description:  record.Description,
			ErrorDetails: record.ErrorDetails,
			PageURL:      record.PageURL,
			UserAgent:    record.UserAgent,
			Screenshots:  shots,
			CreatedAt:    record.CreatedAt,
		})
	}

	totalPages := (total + pageSize - 1) / pageSize
	return &models.BugReportListResponse{
		Data:       items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *BugReportService) ExportBugReports(req models.BugReportExportRequest) ([]byte, string, string, error) {
	format := strings.ToLower(strings.TrimSpace(req.Format))
	if format == "" {
		format = "csv"
	}
	if format != "csv" && format != "json" {
		return nil, "", "", errors.New("仅支持 csv 或 json 格式导出")
	}

	filters, err := parseBugReportFilters(req.StartTime, req.EndTime, req.UserID, req.Username, req.Keyword)
	if err != nil {
		return nil, "", "", err
	}

	records, total, err := s.repo.ListWithFilters(filters, 1, bugReportExportMaxRows)
	if err != nil {
		return nil, "", "", err
	}
	if total > bugReportExportMaxRows {
		return nil, "", "", fmt.Errorf("导出数据超过上限（%d条）", bugReportExportMaxRows)
	}

	switch format {
	case "json":
		payload, err := json.Marshal(records)
		if err != nil {
			return nil, "", "", err
		}
		filename := fmt.Sprintf("bug-reports_%s.json", time.Now().Format("20060102_150405"))
		return payload, "application/json", filename, nil
	default:
		data, err := buildBugReportCSV(records)
		if err != nil {
			return nil, "", "", err
		}
		filename := fmt.Sprintf("bug-reports_%s.csv", time.Now().Format("20060102_150405"))
		return data, "text/csv", filename, nil
	}
}

func buildBugReportCSV(records []models.BugReportAdminRecord) ([]byte, error) {
	buffer := &bytes.Buffer{}
	writer := csv.NewWriter(buffer)

	header := []string{
		"id", "user_id", "username", "title", "description", "error_details",
		"page_url", "user_agent", "screenshots", "created_at",
	}
	if err := writer.Write(header); err != nil {
		return nil, err
	}

	for _, record := range records {
		screenshotsJSON, _ := json.Marshal(record.Screenshots)
		row := []string{
			fmt.Sprintf("%d", record.ID),
			fmt.Sprintf("%d", record.UserID),
			record.Username,
			record.Title,
			record.Description,
			record.ErrorDetails,
			record.PageURL,
			record.UserAgent,
			string(screenshotsJSON),
			record.CreatedAt.Format(time.RFC3339),
		}
		if err := writer.Write(row); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func parseBugReportFilters(startTime, endTime string, userID int, username, keyword string) (models.BugReportQueryFilters, error) {
	filters := models.BugReportQueryFilters{
		Username: strings.TrimSpace(username),
		Keyword:  strings.TrimSpace(keyword),
	}

	if userID > 0 {
		filters.UserID = &userID
	}

	start, err := parseBugReportTime(startTime)
	if err != nil {
		return filters, err
	}
	end, err := parseBugReportTime(endTime)
	if err != nil {
		return filters, err
	}
	if start != nil && end != nil && start.After(*end) {
		return filters, errors.New("开始时间不能晚于结束时间")
	}
	filters.StartTime = start
	filters.EndTime = end

	return filters, nil
}

func parseBugReportTime(value string) (*time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil, nil
	}

	layouts := []string{
		"2006-01-02 15:04:05",
		time.RFC3339,
	}

	for _, layout := range layouts {
		if parsed, err := time.ParseInLocation(layout, trimmed, time.Local); err == nil {
			return &parsed, nil
		}
	}
	return nil, fmt.Errorf("时间格式错误：%s", value)
}

func (s *BugReportService) uploadScreenshot(userID int, fileHeader *multipart.FileHeader) (models.BugReportScreenshot, error) {
	if fileHeader == nil {
		return models.BugReportScreenshot{}, errors.New("截图文件无效")
	}
	if fileHeader.Size > bugReportMaxScreenshotSize {
		return models.BugReportScreenshot{}, ErrScreenshotTooLarge
	}

	file, err := fileHeader.Open()
	if err != nil {
		return models.BugReportScreenshot{}, err
	}
	defer file.Close()

	data, err := readLimited(file, bugReportMaxScreenshotSize)
	if err != nil {
		return models.BugReportScreenshot{}, err
	}

	contentType := http.DetectContentType(data)
	if !isAllowedScreenshotType(contentType) {
		return models.BugReportScreenshot{}, ErrUnsupportedScreenshot
	}

	ext := strings.ToLower(path.Ext(fileHeader.Filename))
	if ext == "" {
		ext = extensionForContentType(contentType)
	}
	if ext == "" {
		ext = ".png"
	}
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	key := fmt.Sprintf("%s%d/%s%s", s.screenshotPrefix, userID, uuid.NewString(), ext)
	if err := s.r2.UploadObject(key, data, contentType); err != nil {
		return models.BugReportScreenshot{}, err
	}

	return models.BugReportScreenshot{
		Key:         key,
		Filename:    fileHeader.Filename,
		Size:        int64(len(data)),
		ContentType: contentType,
	}, nil
}

func readLimited(reader io.Reader, limit int64) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(reader, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, ErrScreenshotTooLarge
	}
	return data, nil
}

func isAllowedScreenshotType(contentType string) bool {
	switch contentType {
	case "image/png", "image/jpeg", "image/webp":
		return true
	default:
		return false
	}
}

func extensionForContentType(contentType string) string {
	extensions, err := mime.ExtensionsByType(contentType)
	if err != nil || len(extensions) == 0 {
		return ""
	}
	return extensions[0]
}

func normalizeBugReportPrefix(prefix string) string {
	cleaned := strings.TrimSpace(prefix)
	if cleaned == "" {
		cleaned = "bug-reports/screenshots/"
	}
	cleaned = strings.TrimPrefix(cleaned, "/")
	if !strings.HasSuffix(cleaned, "/") {
		cleaned += "/"
	}
	return cleaned
}
