package services

import (
	"bytes"
	"comment-review-platform/internal/models"
	"comment-review-platform/internal/repository"
	"comment-review-platform/pkg/r2"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

const (
	auditExportMaxRows    = 100000
	auditExportDailyLimit = 20
	auditExportExpiration = 7 * 24 * time.Hour
)

type AuditLogService struct {
	repo *repository.AuditLogRepository
	r2   *r2.R2Service
}

func NewAuditLogService() *AuditLogService {
	r2Service, err := r2.NewR2Service()
	if err != nil {
		r2Service = nil
	}
	return &AuditLogService{
		repo: repository.NewAuditLogRepository(),
		r2:   r2Service,
	}
}

func (s *AuditLogService) QueryLogs(req models.AuditLogQueryRequest) (*models.AuditLogQueryResponse, error) {
	filters, err := buildFiltersFromQuery(req)
	if err != nil {
		return nil, err
	}

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

	sortBy := strings.TrimSpace(req.SortBy)
	sortOrder := strings.TrimSpace(req.SortOrder)

	entries, total, err := s.repo.ListLogs(filters, page, pageSize, sortBy, sortOrder)
	if err != nil {
		return nil, err
	}

	totalPages := (total + pageSize - 1) / pageSize
	return &models.AuditLogQueryResponse{
		Data:       entries,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *AuditLogService) GetLogByID(id string) (*models.AuditLogEntry, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("invalid audit log id")
	}
	return s.repo.GetLogByID(id)
}

func (s *AuditLogService) ExportLogs(userID int, username, role string, req models.AuditLogExportRequest) (*models.AuditLogExportResponse, error) {
	if s.r2 == nil {
		return nil, errors.New("R2 storage is not configured")
	}

	filters, err := buildFiltersFromExport(req)
	if err != nil {
		return nil, err
	}

	if !isExportUnlimited(role, username) {
		startOfDay := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Local)
		count, err := s.repo.CountExportsSince(userID, startOfDay)
		if err != nil {
			return nil, err
		}
		if count >= auditExportDailyLimit {
			return nil, fmt.Errorf("daily export limit reached (%d)", auditExportDailyLimit)
		}
	}

	filtersJSON, _ := json.Marshal(req)
	fields := normalizeExportFields(req.Fields)

	exportID, err := s.repo.CreateExportRecord(userID, username, req.Format, filtersJSON, fields)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			_ = s.repo.UpdateExportRecord(exportID, "failed", 0, "", time.Time{}, err.Error())
		}
	}()

	entries, total, err := s.repo.ListLogs(filters, 1, auditExportMaxRows, "created_at", "asc")
	if err != nil {
		return nil, err
	}
	if total > auditExportMaxRows {
		err = fmt.Errorf("export rows exceed limit (%d)", auditExportMaxRows)
		return nil, err
	}

	data, contentType, err := buildExportFile(req.Format, fields, entries)
	if err != nil {
		return nil, err
	}

	ext := strings.ToLower(req.Format)
	key := fmt.Sprintf("audit-exports/%d/%s.%s", userID, exportID, ext)
	if err := s.r2.UploadObject(key, data, contentType); err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(auditExportExpiration)
	downloadURL, err := s.r2.GeneratePresignedURL(key, auditExportExpiration)
	if err != nil {
		return nil, err
	}

	if err := s.repo.UpdateExportRecord(exportID, "completed", len(entries), key, expiresAt, ""); err != nil {
		return nil, err
	}

	return &models.AuditLogExportResponse{
		ExportID:    exportID,
		DownloadURL: downloadURL,
		ExpiresAt:   expiresAt,
		RowCount:    len(entries),
	}, nil
}

func (s *AuditLogService) ListExports(userID int, page, pageSize int) (*models.AuditLogExportListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	records, total, err := s.repo.ListExports(userID, page, pageSize)
	if err != nil {
		return nil, err
	}

	if s.r2 != nil {
		for i := range records {
			if records[i].FileKey == nil {
				continue
			}
			if records[i].ExpiresAt != nil && records[i].ExpiresAt.Before(time.Now()) {
				continue
			}
			url, err := s.r2.GeneratePresignedURL(*records[i].FileKey, auditExportExpiration)
			if err != nil {
				continue
			}
			records[i].DownloadURL = &url
		}
	}

	totalPages := (total + pageSize - 1) / pageSize
	return &models.AuditLogExportListResponse{
		Data:       records,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func buildFiltersFromQuery(req models.AuditLogQueryRequest) (models.AuditLogQueryFilters, error) {
	start, err := parseTime(req.StartTime)
	if err != nil {
		return models.AuditLogQueryFilters{}, err
	}
	end, err := parseTime(req.EndTime)
	if err != nil {
		return models.AuditLogQueryFilters{}, err
	}
	if end.Before(start) {
		return models.AuditLogQueryFilters{}, errors.New("end_time must be after start_time")
	}

	return models.AuditLogQueryFilters{
		StartTime:        start,
		EndTime:          end,
		UserID:           req.UserID,
		Username:         strings.TrimSpace(req.Username),
		UserRole:         strings.TrimSpace(req.UserRole),
		ActionTypes:      splitCommaValues(req.ActionTypes),
		ActionCategories: splitCommaValues(req.ActionCategories),
		Result:           strings.TrimSpace(req.Result),
		IPAddress:        strings.TrimSpace(req.IPAddress),
		Endpoint:         strings.TrimSpace(req.Endpoint),
		HTTPMethod:       strings.TrimSpace(req.HTTPMethod),
		StatusCode:       req.StatusCode,
		ResourceType:     strings.TrimSpace(req.ResourceType),
		ResourceID:       strings.TrimSpace(req.ResourceID),
		Keyword:          strings.TrimSpace(req.Keyword),
		MinDurationMs:    req.MinDurationMs,
		MaxDurationMs:    req.MaxDurationMs,
		GeoLocation:      strings.TrimSpace(req.GeoLocation),
		DeviceType:       strings.TrimSpace(req.DeviceType),
	}, nil
}

func buildFiltersFromExport(req models.AuditLogExportRequest) (models.AuditLogQueryFilters, error) {
	start, err := parseTime(req.StartTime)
	if err != nil {
		return models.AuditLogQueryFilters{}, err
	}
	end, err := parseTime(req.EndTime)
	if err != nil {
		return models.AuditLogQueryFilters{}, err
	}
	if end.Before(start) {
		return models.AuditLogQueryFilters{}, errors.New("end_time must be after start_time")
	}

	return models.AuditLogQueryFilters{
		StartTime:        start,
		EndTime:          end,
		UserID:           req.UserID,
		Username:         strings.TrimSpace(req.Username),
		UserRole:         strings.TrimSpace(req.UserRole),
		ActionTypes:      req.ActionTypes,
		ActionCategories: req.ActionCategories,
		Result:           strings.TrimSpace(req.Result),
		IPAddress:        strings.TrimSpace(req.IPAddress),
		Endpoint:         strings.TrimSpace(req.Endpoint),
		HTTPMethod:       strings.TrimSpace(req.HTTPMethod),
		StatusCode:       req.StatusCode,
		ResourceType:     strings.TrimSpace(req.ResourceType),
		ResourceID:       strings.TrimSpace(req.ResourceID),
		Keyword:          strings.TrimSpace(req.Keyword),
		MinDurationMs:    req.MinDurationMs,
		MaxDurationMs:    req.MaxDurationMs,
		GeoLocation:      strings.TrimSpace(req.GeoLocation),
		DeviceType:       strings.TrimSpace(req.DeviceType),
	}, nil
}

func normalizeExportFields(fields []string) []string {
	clean := []string{}
	seen := map[string]bool{}
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field == "" || seen[field] {
			continue
		}
		if _, ok := exportFieldResolvers()[field]; ok {
			clean = append(clean, field)
			seen[field] = true
		}
	}
	if len(clean) == 0 {
		clean = []string{
			"created_at",
			"username",
			"user_role",
			"action_type",
			"action_category",
			"result",
			"endpoint",
			"http_method",
			"status_code",
			"ip_address",
			"request_id",
			"duration_ms",
		}
	}
	return clean
}

func buildExportFile(format string, fields []string, entries []models.AuditLogEntry) ([]byte, string, error) {
	switch strings.ToLower(format) {
	case "json":
		return buildJSONExport(fields, entries)
	case "xlsx":
		return buildXLSXExport(fields, entries)
	case "csv":
		fallthrough
	default:
		return buildCSVExport(fields, entries)
	}
}

func buildCSVExport(fields []string, entries []models.AuditLogEntry) ([]byte, string, error) {
	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)

	if err := writer.Write(fields); err != nil {
		return nil, "", err
	}

	resolvers := exportFieldResolvers()
	for _, entry := range entries {
		row := make([]string, len(fields))
		for i, field := range fields {
			row[i] = formatExportValue(resolvers[field](entry))
		}
		if err := writer.Write(row); err != nil {
			return nil, "", err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, "", err
	}

	return buffer.Bytes(), "text/csv", nil
}

func buildJSONExport(fields []string, entries []models.AuditLogEntry) ([]byte, string, error) {
	resolvers := exportFieldResolvers()
	records := make([]map[string]interface{}, 0, len(entries))
	for _, entry := range entries {
		record := make(map[string]interface{}, len(fields))
		for _, field := range fields {
			record[field] = resolvers[field](entry)
		}
		records = append(records, record)
	}

	data, err := json.Marshal(records)
	if err != nil {
		return nil, "", err
	}
	return data, "application/json", nil
}

func buildXLSXExport(fields []string, entries []models.AuditLogEntry) ([]byte, string, error) {
	file := excelize.NewFile()
	defer func() {
		_ = file.Close()
	}()

	sheet := "Sheet1"
	stream, err := file.NewStreamWriter(sheet)
	if err != nil {
		return nil, "", err
	}

	header := make([]interface{}, len(fields))
	for i, field := range fields {
		header[i] = field
	}
	if err := stream.SetRow("A1", header); err != nil {
		return nil, "", err
	}

	resolvers := exportFieldResolvers()
	for idx, entry := range entries {
		row := make([]interface{}, len(fields))
		for i, field := range fields {
			row[i] = formatExportValue(resolvers[field](entry))
		}
		cell, err := excelize.CoordinatesToCellName(1, idx+2)
		if err != nil {
			return nil, "", err
		}
		if err := stream.SetRow(cell, row); err != nil {
			return nil, "", err
		}
	}

	if err := stream.Flush(); err != nil {
		return nil, "", err
	}

	buffer, err := file.WriteToBuffer()
	if err != nil {
		return nil, "", err
	}

	return buffer.Bytes(), "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", nil
}

func exportFieldResolvers() map[string]func(models.AuditLogEntry) interface{} {
	return map[string]func(models.AuditLogEntry) interface{}{
		"id":         func(e models.AuditLogEntry) interface{} { return e.ID },
		"created_at": func(e models.AuditLogEntry) interface{} { return e.CreatedAt.Format(time.RFC3339) },
		"user_id": func(e models.AuditLogEntry) interface{} {
			if e.UserID == nil {
				return ""
			}
			return *e.UserID
		},
		"username":           func(e models.AuditLogEntry) interface{} { return e.Username },
		"user_role":          func(e models.AuditLogEntry) interface{} { return e.UserRole },
		"action_type":        func(e models.AuditLogEntry) interface{} { return e.ActionType },
		"action_category":    func(e models.AuditLogEntry) interface{} { return e.ActionCategory },
		"action_description": func(e models.AuditLogEntry) interface{} { return e.ActionDescription },
		"result":             func(e models.AuditLogEntry) interface{} { return e.Result },
		"endpoint":           func(e models.AuditLogEntry) interface{} { return e.Endpoint },
		"http_method":        func(e models.AuditLogEntry) interface{} { return e.HTTPMethod },
		"status_code":        func(e models.AuditLogEntry) interface{} { return e.StatusCode },
		"request_id":         func(e models.AuditLogEntry) interface{} { return e.RequestID },
		"session_id":         func(e models.AuditLogEntry) interface{} { return e.SessionID },
		"ip_address":         func(e models.AuditLogEntry) interface{} { return e.IPAddress },
		"user_agent":         func(e models.AuditLogEntry) interface{} { return e.UserAgent },
		"resource_type":      func(e models.AuditLogEntry) interface{} { return e.ResourceType },
		"resource_id":        func(e models.AuditLogEntry) interface{} { return e.ResourceID },
		"error_code":         func(e models.AuditLogEntry) interface{} { return e.ErrorCode },
		"error_type":         func(e models.AuditLogEntry) interface{} { return e.ErrorType },
		"error_description":  func(e models.AuditLogEntry) interface{} { return e.ErrorDescription },
		"error_message":      func(e models.AuditLogEntry) interface{} { return e.ErrorMessage },
		"error_stack":        func(e models.AuditLogEntry) interface{} { return e.ErrorStack },
		"duration_ms":        func(e models.AuditLogEntry) interface{} { return e.DurationMs },
		"module_name":        func(e models.AuditLogEntry) interface{} { return e.ModuleName },
		"method_name":        func(e models.AuditLogEntry) interface{} { return e.MethodName },
		"server_ip":          func(e models.AuditLogEntry) interface{} { return e.ServerIP },
		"server_port":        func(e models.AuditLogEntry) interface{} { return e.ServerPort },
		"page_url":           func(e models.AuditLogEntry) interface{} { return e.PageURL },
		"request_body":       func(e models.AuditLogEntry) interface{} { return jsonValue(e.RequestBody) },
		"request_params":     func(e models.AuditLogEntry) interface{} { return jsonValue(e.RequestParams) },
		"response_body":      func(e models.AuditLogEntry) interface{} { return jsonValue(e.ResponseBody) },
		"resource_ids":       func(e models.AuditLogEntry) interface{} { return jsonValue(e.ResourceIDs) },
		"changes":            func(e models.AuditLogEntry) interface{} { return jsonValue(e.Changes) },
	}
}

func jsonValue(raw json.RawMessage) interface{} {
	if len(raw) == 0 {
		return ""
	}
	var value interface{}
	if err := json.Unmarshal(raw, &value); err != nil {
		return string(raw)
	}
	return value
}

func formatExportValue(value interface{}) string {
	switch typed := value.(type) {
	case string:
		return typed
	case int:
		return fmt.Sprintf("%d", typed)
	case int64:
		return fmt.Sprintf("%d", typed)
	case float64:
		return fmt.Sprintf("%v", typed)
	case bool:
		return fmt.Sprintf("%t", typed)
	default:
		if value == nil {
			return ""
		}
		return fmt.Sprintf("%v", value)
	}
}

func parseTime(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, errors.New("time is required")
	}

	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}

	for _, layout := range layouts {
		if parsed, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid time format: %s", value)
}

func splitCommaValues(value string) []string {
	parts := strings.Split(value, ",")
	items := []string{}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			items = append(items, part)
		}
	}
	return items
}

func isExportUnlimited(role, username string) bool {
	if strings.ToLower(role) != "admin" {
		return false
	}
	return strings.ToLower(username) == "admin"
}
