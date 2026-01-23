package repository

import (
	"comment-review-platform/internal/models"
	"comment-review-platform/pkg/database"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
)

type AuditLogRepository struct {
	db *sql.DB
}

func NewAuditLogRepository() *AuditLogRepository {
	return &AuditLogRepository{db: database.DB}
}

func (r *AuditLogRepository) ListLogs(filters models.AuditLogQueryFilters, page, pageSize int, sortBy, sortOrder string) ([]models.AuditLogEntry, int, error) {
	whereClause, args := buildAuditLogWhere(filters)

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM audit_logs %s", whereClause)
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	orderColumn := "created_at"
	if sortBy == "duration_ms" {
		orderColumn = "duration_ms"
	}
	orderDirection := "DESC"
	if strings.ToLower(sortOrder) == "asc" {
		orderDirection = "ASC"
	}

	offset := (page - 1) * pageSize

	dataQuery := fmt.Sprintf(`
		SELECT
			id, created_at, user_id, username, user_role,
			action_type, action_category, action_description, result,
			endpoint, http_method, status_code, request_id, session_id,
			request_body, request_params, response_body, ip_address, user_agent,
			geo_location, device_type, browser, os,
			resource_type, resource_id, resource_ids, changes,
			error_code, error_type, error_description, error_message, error_stack, duration_ms,
			module_name, method_name, server_ip, server_port, page_url
		FROM audit_logs
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderColumn, orderDirection, len(args)+1, len(args)+2)

	args = append(args, pageSize, offset)

	rows, err := r.db.Query(dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	entries := []models.AuditLogEntry{}
	for rows.Next() {
		entry, err := scanAuditLogRow(rows)
		if err != nil {
			return nil, 0, err
		}
		entries = append(entries, entry)
	}

	return entries, total, nil
}

func (r *AuditLogRepository) GetLogByID(id string) (*models.AuditLogEntry, error) {
	query := `
		SELECT
			id, created_at, user_id, username, user_role,
			action_type, action_category, action_description, result,
			endpoint, http_method, status_code, request_id, session_id,
			request_body, request_params, response_body, ip_address, user_agent,
			geo_location, device_type, browser, os,
			resource_type, resource_id, resource_ids, changes,
			error_code, error_type, error_description, error_message, error_stack, duration_ms,
			module_name, method_name, server_ip, server_port, page_url
		FROM audit_logs
		WHERE id = $1
	`

	row := r.db.QueryRow(query, id)
	entry, err := scanAuditLogRow(row)
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func (r *AuditLogRepository) CreateExportRecord(userID int, username, format string, filters json.RawMessage, fields []string) (string, error) {
	query := `
		INSERT INTO audit_log_exports (user_id, username, export_format, filters, fields)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	var id string
	err := r.db.QueryRow(query, userID, username, format, nullableJSON(filters), pq.Array(fields)).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (r *AuditLogRepository) UpdateExportRecord(id string, status string, rowCount int, fileKey string, expiresAt time.Time, errorMessage string) error {
	query := `
		UPDATE audit_log_exports
		SET status = $1, row_count = $2, file_key = $3, expires_at = $4, error_message = $5
		WHERE id = $6
	`

	_, err := r.db.Exec(query, status, nullableInt(rowCount), nullableString(fileKey), nullableTime(expiresAt), nullableString(errorMessage), id)
	return err
}

func (r *AuditLogRepository) ListExports(userID int, page, pageSize int) ([]models.AuditLogExportRecord, int, error) {
	offset := (page - 1) * pageSize

	countQuery := `
		SELECT COUNT(*)
		FROM audit_log_exports
		WHERE user_id = $1
	`
	var total int
	if err := r.db.QueryRow(countQuery, userID).Scan(&total); err != nil {
		return nil, 0, err
	}

	dataQuery := `
		SELECT id, user_id, username, export_format, filters, fields, status, row_count, file_key, expires_at, error_message, created_at
		FROM audit_log_exports
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(dataQuery, userID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	records := []models.AuditLogExportRecord{}
	for rows.Next() {
		var record models.AuditLogExportRecord
		var filtersBytes []byte
		var fields []string
		var rowCount sql.NullInt64
		var fileKey sql.NullString
		var expiresAt sql.NullTime
		var errorMessage sql.NullString

		err := rows.Scan(
			&record.ID,
			&record.UserID,
			&record.Username,
			&record.ExportFormat,
			&filtersBytes,
			pq.Array(&fields),
			&record.Status,
			&rowCount,
			&fileKey,
			&expiresAt,
			&errorMessage,
			&record.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		record.Filters = rawFromBytes(filtersBytes)
		record.Fields = fields
		if rowCount.Valid {
			record.RowCount = int(rowCount.Int64)
		}
		if fileKey.Valid {
			record.FileKey = &fileKey.String
		}
		if expiresAt.Valid {
			record.ExpiresAt = &expiresAt.Time
		}
		if errorMessage.Valid {
			record.ErrorMessage = &errorMessage.String
		}

		records = append(records, record)
	}

	return records, total, nil
}

func (r *AuditLogRepository) CountExportsSince(userID int, since time.Time) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM audit_log_exports
		WHERE user_id = $1 AND created_at >= $2
	`
	var count int
	err := r.db.QueryRow(query, userID, since).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func buildAuditLogWhere(filters models.AuditLogQueryFilters) (string, []interface{}) {
	conditions := []string{"created_at >= $1", "created_at <= $2"}
	args := []interface{}{filters.StartTime, filters.EndTime}
	argPos := 3

	if filters.UserID != nil {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argPos))
		args = append(args, *filters.UserID)
		argPos++
	}

	if filters.Username != "" {
		conditions = append(conditions, fmt.Sprintf("username ILIKE $%d", argPos))
		args = append(args, "%"+filters.Username+"%")
		argPos++
	}

	if filters.UserRole != "" {
		conditions = append(conditions, fmt.Sprintf("user_role = $%d", argPos))
		args = append(args, filters.UserRole)
		argPos++
	}

	if len(filters.ActionTypes) > 0 {
		conditions = append(conditions, fmt.Sprintf("action_type = ANY($%d)", argPos))
		args = append(args, pq.Array(filters.ActionTypes))
		argPos++
	}

	if len(filters.ActionCategories) > 0 {
		conditions = append(conditions, fmt.Sprintf("action_category = ANY($%d)", argPos))
		args = append(args, pq.Array(filters.ActionCategories))
		argPos++
	}

	if filters.Result != "" {
		conditions = append(conditions, fmt.Sprintf("result = $%d", argPos))
		args = append(args, filters.Result)
		argPos++
	}

	if filters.IPAddress != "" {
		pattern := strings.ReplaceAll(filters.IPAddress, "*", "%")
		if strings.Contains(pattern, "%") {
			conditions = append(conditions, fmt.Sprintf("ip_address ILIKE $%d", argPos))
		} else {
			conditions = append(conditions, fmt.Sprintf("ip_address = $%d", argPos))
		}
		args = append(args, pattern)
		argPos++
	}

	if filters.Endpoint != "" {
		pattern := strings.ReplaceAll(filters.Endpoint, "*", "%")
		if strings.Contains(pattern, "%") {
			conditions = append(conditions, fmt.Sprintf("endpoint ILIKE $%d", argPos))
		} else {
			conditions = append(conditions, fmt.Sprintf("endpoint = $%d", argPos))
		}
		args = append(args, pattern)
		argPos++
	}

	if filters.HTTPMethod != "" {
		conditions = append(conditions, fmt.Sprintf("http_method = $%d", argPos))
		args = append(args, strings.ToUpper(filters.HTTPMethod))
		argPos++
	}

	if filters.StatusCode != nil {
		conditions = append(conditions, fmt.Sprintf("status_code = $%d", argPos))
		args = append(args, *filters.StatusCode)
		argPos++
	}

	if filters.ResourceType != "" {
		conditions = append(conditions, fmt.Sprintf("resource_type = $%d", argPos))
		args = append(args, filters.ResourceType)
		argPos++
	}

	if filters.ResourceID != "" {
		conditions = append(conditions, fmt.Sprintf("resource_id = $%d", argPos))
		args = append(args, filters.ResourceID)
		argPos++
	}

	if filters.Keyword != "" {
		conditions = append(conditions, fmt.Sprintf("to_tsvector('simple', coalesce(action_description, '') || ' ' || coalesce(error_message, '')) @@ plainto_tsquery('simple', $%d)", argPos))
		args = append(args, filters.Keyword)
		argPos++
	}

	if filters.MinDurationMs != nil {
		conditions = append(conditions, fmt.Sprintf("duration_ms >= $%d", argPos))
		args = append(args, *filters.MinDurationMs)
		argPos++
	}

	if filters.MaxDurationMs != nil {
		conditions = append(conditions, fmt.Sprintf("duration_ms <= $%d", argPos))
		args = append(args, *filters.MaxDurationMs)
		argPos++
	}

	if filters.GeoLocation != "" {
		conditions = append(conditions, fmt.Sprintf("geo_location = $%d", argPos))
		args = append(args, filters.GeoLocation)
		argPos++
	}

	if filters.DeviceType != "" {
		conditions = append(conditions, fmt.Sprintf("device_type = $%d", argPos))
		args = append(args, filters.DeviceType)
		argPos++
	}

	return "WHERE " + strings.Join(conditions, " AND "), args
}

type auditLogScanner interface {
	Scan(dest ...interface{}) error
}

func scanAuditLogRow(scanner auditLogScanner) (models.AuditLogEntry, error) {
	var entry models.AuditLogEntry
	var userID sql.NullInt64
	var username sql.NullString
	var userRole sql.NullString
	var actionType sql.NullString
	var actionCategory sql.NullString
	var actionDescription sql.NullString
	var result sql.NullString
	var endpoint sql.NullString
	var httpMethod sql.NullString
	var statusCode sql.NullInt64
	var requestID sql.NullString
	var sessionID sql.NullString
	var requestBody []byte
	var requestParams []byte
	var responseBody []byte
	var ipAddress sql.NullString
	var userAgent sql.NullString
	var geoLocation sql.NullString
	var deviceType sql.NullString
	var browser sql.NullString
	var osValue sql.NullString
	var resourceType sql.NullString
	var resourceID sql.NullString
	var resourceIDs []byte
	var changes []byte
	var errorCode sql.NullString
	var errorType sql.NullString
	var errorDescription sql.NullString
	var errorMessage sql.NullString
	var errorStack sql.NullString
	var durationMs sql.NullInt64
	var moduleName sql.NullString
	var methodName sql.NullString
	var serverIP sql.NullString
	var serverPort sql.NullString
	var pageURL sql.NullString

	err := scanner.Scan(
		&entry.ID,
		&entry.CreatedAt,
		&userID,
		&username,
		&userRole,
		&actionType,
		&actionCategory,
		&actionDescription,
		&result,
		&endpoint,
		&httpMethod,
		&statusCode,
		&requestID,
		&sessionID,
		&requestBody,
		&requestParams,
		&responseBody,
		&ipAddress,
		&userAgent,
		&geoLocation,
		&deviceType,
		&browser,
		&osValue,
		&resourceType,
		&resourceID,
		&resourceIDs,
		&changes,
		&errorCode,
		&errorType,
		&errorDescription,
		&errorMessage,
		&errorStack,
		&durationMs,
		&moduleName,
		&methodName,
		&serverIP,
		&serverPort,
		&pageURL,
	)
	if err != nil {
		return entry, err
	}

	entry.UserID = intPointer(userID)
	entry.Username = username.String
	entry.UserRole = userRole.String
	entry.ActionType = actionType.String
	entry.ActionCategory = actionCategory.String
	entry.ActionDescription = actionDescription.String
	entry.Result = result.String
	entry.Endpoint = endpoint.String
	entry.HTTPMethod = httpMethod.String
	if statusCode.Valid {
		entry.StatusCode = int(statusCode.Int64)
	}
	entry.RequestID = requestID.String
	entry.SessionID = sessionID.String
	entry.RequestBody = rawFromBytes(requestBody)
	entry.RequestParams = rawFromBytes(requestParams)
	entry.ResponseBody = rawFromBytes(responseBody)
	entry.IPAddress = ipAddress.String
	entry.UserAgent = userAgent.String
	entry.GeoLocation = geoLocation.String
	entry.DeviceType = deviceType.String
	entry.Browser = browser.String
	entry.OS = osValue.String
	entry.ResourceType = resourceType.String
	entry.ResourceID = resourceID.String
	entry.ResourceIDs = rawFromBytes(resourceIDs)
	entry.Changes = rawFromBytes(changes)
	entry.ErrorCode = errorCode.String
	entry.ErrorType = errorType.String
	entry.ErrorDescription = errorDescription.String
	entry.ErrorMessage = errorMessage.String
	entry.ErrorStack = errorStack.String
	if durationMs.Valid {
		entry.DurationMs = int(durationMs.Int64)
	}
	entry.ModuleName = moduleName.String
	entry.MethodName = methodName.String
	entry.ServerIP = serverIP.String
	entry.ServerPort = serverPort.String
	entry.PageURL = pageURL.String

	return entry, nil
}

func intPointer(value sql.NullInt64) *int {
	if !value.Valid {
		return nil
	}
	v := int(value.Int64)
	return &v
}

func rawFromBytes(value []byte) json.RawMessage {
	if len(value) == 0 {
		return nil
	}
	return json.RawMessage(value)
}

func nullableString(value string) interface{} {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func nullableInt(value int) interface{} {
	if value == 0 {
		return nil
	}
	return value
}

func nullableJSON(value json.RawMessage) interface{} {
	if len(value) == 0 {
		return nil
	}
	return value
}

func nullableTime(value time.Time) interface{} {
	if value.IsZero() {
		return nil
	}
	return value
}
