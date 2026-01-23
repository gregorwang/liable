package middleware

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"comment-review-platform/internal/config"
	"comment-review-platform/internal/observability"
	"comment-review-platform/internal/services"

	"github.com/gin-gonic/gin"
)

const (
	auditRequestBodyKey    = "audit_request_body"
	auditRequestBodyRawKey = "audit_request_body_raw"
	auditContextKey        = "audit_context"
	maxAuditPayloadBytes   = 32 * 1024
	maxAuditTextFieldBytes = 64 * 1024
)

// AuditLog represents an audit log entry
type AuditLog struct {
	ID                string
	CreatedAt         time.Time
	UserID            *int
	Username          string
	UserRole          string
	ActionType        string
	ActionCategory    string
	ActionDescription string
	Result            string
	Endpoint          string
	HTTPMethod        string
	StatusCode        int
	RequestID         string
	SessionID         string
	RequestBody       json.RawMessage
	ResponseBody      json.RawMessage
	IPAddress         string
	UserAgent         string
	GeoLocation       string
	DeviceType        string
	Browser           string
	OS                string
	ResourceType      string
	ResourceID        string
	ResourceIDs       json.RawMessage
	Changes           json.RawMessage
	ErrorCode         string
	ErrorType         string
	ErrorDescription  string
	ErrorMessage      string
	ErrorStack        string
	DurationMs        int
	ModuleName        string
	MethodName        string
	ServerIP          string
	ServerPort        string
	PageURL           string
	RequestParams     json.RawMessage
}

// AuditContext allows handlers to override audit metadata
type AuditContext struct {
	ActionType        string
	ActionCategory    string
	ActionDescription string
	ResourceType      string
	ResourceID        string
	ResourceIDs       json.RawMessage
	Changes           json.RawMessage
}

// AuditLogger handles audit log operations
type AuditLogger struct {
	db *sql.DB
}

var auditLogger *AuditLogger

// NewAuditLogger creates a new audit logger instance
func NewAuditLogger(db *sql.DB) *AuditLogger {
	return &AuditLogger{db: db}
}

// InitAuditLogger initializes the global audit logger
func InitAuditLogger(db *sql.DB) {
	auditLogger = NewAuditLogger(db)
}

// SetAuditContext allows handlers to set explicit audit metadata
func SetAuditContext(c *gin.Context, ctx AuditContext) {
	c.Set(auditContextKey, ctx)
}

// AuditLogMiddleware returns a middleware that logs audit information
func AuditLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		requestID := GetTraceID(c)
		if requestID == "" {
			requestID = newTraceID()
			c.Set(traceIDKey, requestID)
			c.Set(requestIDKey, requestID)
		}

		captureRequestBody(c)

		bodyWriter := &bodyLogWriter{
			ResponseWriter: c.Writer,
			buffer:         newLimitedBuffer(maxAuditPayloadBytes),
		}
		c.Writer = bodyWriter

		// Process request
		c.Next()

		// Only log if audit logger is initialized
		if auditLogger == nil {
			return
		}

		// Build audit log entry
		auditEntry := buildAuditLogEntry(c, startTime, requestID, bodyWriter)

		// Async write to avoid blocking response
		go func(logEntry AuditLog) {
			if err := auditLogger.Save(logEntry); err != nil {
				observability.Infof(logEntry.RequestID, "Failed to save audit log: %v", err)
			}
			if logEntry.StatusCode >= 400 {
				if logEntry.StatusCode == http.StatusForbidden {
					return
				}
				if svc := getAlertService(); svc != nil {
					alertEvent := services.AlertEvent{
						TraceID:          logEntry.RequestID,
						HTTPMethod:       logEntry.HTTPMethod,
						Endpoint:         logEntry.Endpoint,
						StatusCode:       logEntry.StatusCode,
						OccurredAt:       logEntry.CreatedAt,
						ErrorCode:        logEntry.ErrorCode,
						ErrorType:        logEntry.ErrorType,
						ErrorDescription: logEntry.ErrorDescription,
						ErrorMessage:     logEntry.ErrorMessage,
						ErrorStack:       logEntry.ErrorStack,
						ModuleName:       logEntry.ModuleName,
						MethodName:       logEntry.MethodName,
						ServerIP:         logEntry.ServerIP,
						ServerPort:       logEntry.ServerPort,
						PageURL:          logEntry.PageURL,
						UserID:           logEntry.UserID,
						Username:         logEntry.Username,
						UserAgent:        logEntry.UserAgent,
						RequestBody:      logEntry.RequestBody,
						RequestParams:    logEntry.RequestParams,
						ResponseBody:     logEntry.ResponseBody,
					}
					if err := svc.NotifyFromAuditLog(alertEvent); err != nil {
						observability.Infof(logEntry.RequestID, "Failed to send alert: %v", err)
					}
				}
			}
		}(auditEntry)
	}
}

// captureRequestBody reads and sanitizes JSON request bodies
func captureRequestBody(c *gin.Context) {
	if !shouldCapturePayload(c.GetHeader("Content-Type")) {
		return
	}

	bodyBytes, err := readBodyWithLimit(c.Request.Body, maxAuditPayloadBytes)
	_ = c.Request.Body.Close()
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	if err != nil || len(bodyBytes) == 0 {
		return
	}

	sanitizedValue, sanitizedJSON := sanitizeJSONPayload(bodyBytes)
	if sanitizedJSON != nil {
		c.Set(auditRequestBodyKey, sanitizedValue)
		c.Set(auditRequestBodyRawKey, sanitizedJSON)
	}
}

// buildAuditLogEntry constructs an audit log entry from Gin context
func buildAuditLogEntry(c *gin.Context, startTime time.Time, requestID string, bodyWriter *bodyLogWriter) AuditLog {
	// Get user information from context
	var userID *int
	if uid := GetUserID(c); uid > 0 {
		userID = &uid
	}
	username := GetUsername(c)
	role := GetRole(c)

	// Get query params
	endpoint := c.Request.URL.Path
	if fullPath := c.FullPath(); fullPath != "" {
		endpoint = fullPath
	}

	durationMs := int(time.Since(startTime).Milliseconds())
	statusCode := c.Writer.Status()

	// Resolve audit metadata
	requestBodyValue, _ := c.Get(auditRequestBodyKey)
	requestBodyJSON := getRawJSONFromContext(c, auditRequestBodyRawKey)
	requestParamsJSON := buildRequestParams(c, requestBodyValue)
	metadata := resolveAuditMetadata(c, requestBodyValue)

	// Determine result
	result := deriveResult(statusCode)

	// Capture response body if needed
	responseBody := sanitizeResponseBody(c, bodyWriter)

	// Extract error message if request failed
	errorMessage := collectErrors(c, responseBody)
	errorStack := getErrorStack(c)
	errorCode, errorType, errorDescription := collectErrorMeta(c, responseBody, errorMessage, errorStack)

	// Session ID from header or cookie
	sessionID := strings.TrimSpace(c.GetHeader("X-Session-Id"))
	if sessionID == "" {
		sessionID = strings.TrimSpace(c.GetHeader("X-Session-ID"))
	}
	if sessionID == "" {
		if cookieValue, err := c.Cookie("session_id"); err == nil {
			sessionID = cookieValue
		}
	}

	return AuditLog{
		CreatedAt:         startTime,
		UserID:            userID,
		Username:          username,
		UserRole:          role,
		ActionType:        metadata.ActionType,
		ActionCategory:    metadata.ActionCategory,
		ActionDescription: metadata.ActionDescription,
		Result:            result,
		Endpoint:          endpoint,
		HTTPMethod:        c.Request.Method,
		StatusCode:        statusCode,
		RequestID:         requestID,
		SessionID:         sessionID,
		RequestBody:       requestBodyJSON,
		ResponseBody:      responseBody,
		IPAddress:         c.ClientIP(),
		UserAgent:         truncateString(c.GetHeader("User-Agent"), 500),
		GeoLocation:       "",
		DeviceType:        "",
		Browser:           "",
		OS:                "",
		ResourceType:      metadata.ResourceType,
		ResourceID:        metadata.ResourceID,
		ResourceIDs:       metadata.ResourceIDs,
		Changes:           metadata.Changes,
		ErrorCode:         errorCode,
		ErrorType:         errorType,
		ErrorDescription:  errorDescription,
		ErrorMessage:      errorMessage,
		ErrorStack:        errorStack,
		DurationMs:        durationMs,
		ModuleName:        resolveModuleName(c),
		MethodName:        resolveMethodName(c),
		ServerIP:          resolveServerIP(c),
		ServerPort:        resolveServerPort(c),
		PageURL:           resolvePageURL(c),
		RequestParams:     requestParamsJSON,
	}
}

// Save stores an audit log entry to the database
func (a *AuditLogger) Save(entry AuditLog) error {
	if a.db == nil {
		return nil // Database not initialized, skip logging
	}

	query := `
		INSERT INTO audit_logs (
			created_at, user_id, username, user_role,
			action_type, action_category, action_description, result,
			endpoint, http_method, status_code, request_id, session_id,
			request_body, request_params, response_body, ip_address, user_agent,
			geo_location, device_type, browser, os,
			resource_type, resource_id, resource_ids, changes,
			error_code, error_type, error_description, error_message, error_stack, duration_ms,
			module_name, method_name, server_ip, server_port, page_url
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7, $8,
			$9, $10, $11, $12, $13,
			$14, $15, $16, $17, $18,
			$19, $20, $21, $22,
			$23, $24, $25, $26,
			$27, $28, $29, $30, $31, $32,
			$33, $34, $35, $36, $37
		)
	`

	_, err := a.db.Exec(
		query,
		entry.CreatedAt,
		entry.UserID,
		nullableString(entry.Username),
		nullableString(entry.UserRole),
		nullableString(entry.ActionType),
		nullableString(entry.ActionCategory),
		nullableString(entry.ActionDescription),
		nullableString(entry.Result),
		nullableString(entry.Endpoint),
		nullableString(entry.HTTPMethod),
		entry.StatusCode,
		nullableString(entry.RequestID),
		nullableString(entry.SessionID),
		nullableJSON(entry.RequestBody),
		nullableJSON(entry.RequestParams),
		nullableJSON(entry.ResponseBody),
		nullableString(entry.IPAddress),
		nullableString(entry.UserAgent),
		nullableString(entry.GeoLocation),
		nullableString(entry.DeviceType),
		nullableString(entry.Browser),
		nullableString(entry.OS),
		nullableString(entry.ResourceType),
		nullableString(entry.ResourceID),
		nullableJSON(entry.ResourceIDs),
		nullableJSON(entry.Changes),
		nullableString(entry.ErrorCode),
		nullableString(entry.ErrorType),
		nullableString(entry.ErrorDescription),
		nullableString(entry.ErrorMessage),
		nullableString(entry.ErrorStack),
		nullableInt(entry.DurationMs),
		nullableString(entry.ModuleName),
		nullableString(entry.MethodName),
		nullableString(entry.ServerIP),
		nullableString(entry.ServerPort),
		nullableString(entry.PageURL),
	)

	return err
}

// SetCheckedPermission marks that a permission check was performed
// This can be called in permission middleware to track what permission was checked
func SetCheckedPermission(c *gin.Context, permissionKey string) {
	c.Set("checked_permission", permissionKey)
}

// StartAuditLogCleanup starts a background job to clean up old audit logs
func StartAuditLogCleanup(retentionDays int, interval time.Duration) {
	if auditLogger == nil {
		return
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			if err := auditLogger.CleanupOldLogs(retentionDays); err != nil {
				log.Printf("Error cleaning up audit logs: %v", err)
			}
		}
	}()

	log.Printf("Audit log cleanup started: retention=%d days, interval=%v", retentionDays, interval)
}

// CleanupOldLogs deletes audit logs older than specified retention period
func (a *AuditLogger) CleanupOldLogs(retentionDays int) error {
	if a.db == nil {
		return nil
	}

	query := `
		DELETE FROM audit_logs
		WHERE created_at < NOW() - INTERVAL '1 day' * $1
	`

	result, err := a.db.Exec(query, retentionDays)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	log.Printf("Cleaned up %d audit logs older than %d days", rowsAffected, retentionDays)
	return nil
}

type auditMetadata struct {
	ActionType        string
	ActionCategory    string
	ActionDescription string
	ResourceType      string
	ResourceID        string
	ResourceIDs       json.RawMessage
	Changes           json.RawMessage
}

func resolveAuditMetadata(c *gin.Context, requestBody interface{}) auditMetadata {
	if ctxValue, ok := c.Get(auditContextKey); ok {
		if ctx, okCast := ctxValue.(AuditContext); okCast {
			return auditMetadata{
				ActionType:        fallbackString(ctx.ActionType, "api.request"),
				ActionCategory:    fallbackString(ctx.ActionCategory, "system_operation"),
				ActionDescription: fallbackString(ctx.ActionDescription, "API request"),
				ResourceType:      ctx.ResourceType,
				ResourceID:        ctx.ResourceID,
				ResourceIDs:       ctx.ResourceIDs,
				Changes:           ctx.Changes,
			}
		}
	}

	path := c.FullPath()
	if path == "" {
		path = c.Request.URL.Path
	}

	resourceType, resourceID := resolveResource(path, c)
	resourceIDs := extractResourceIDs(requestBody)
	changes := buildChanges(c.Request.Method, requestBody)

	actionType, category, description := resolveAction(path, c.Request.Method, requestBody)

	return auditMetadata{
		ActionType:        actionType,
		ActionCategory:    category,
		ActionDescription: description,
		ResourceType:      resourceType,
		ResourceID:        resourceID,
		ResourceIDs:       resourceIDs,
		Changes:           changes,
	}
}

func resolveAction(path, method string, requestBody interface{}) (string, string, string) {
	if method == "POST" && path == "/api/auth/login" {
		return "auth.login", "authentication", "用户登录"
	}
	if method == "POST" && path == "/api/auth/login-with-code" {
		return "auth.login", "authentication", "用户登录"
	}
	if method == "POST" && path == "/api/auth/register" {
		return "auth.register", "authentication", "用户注册"
	}
	if method == "POST" && path == "/api/auth/register-with-code" {
		return "auth.register", "authentication", "用户注册"
	}

	if method == "POST" && path == "/api/admin/permissions/grant" {
		return "permission.grant", "authorization", "授予权限"
	}
	if method == "POST" && path == "/api/admin/permissions/revoke" {
		return "permission.revoke", "authorization", "撤销权限"
	}

	if method == "PUT" && path == "/api/admin/users/:id/approve" {
		if body, ok := requestBody.(map[string]interface{}); ok {
			if status, ok := body["status"].(string); ok && status == "rejected" {
				return "user.reject", "user_management", "拒绝用户"
			}
		}
		return "user.approve", "user_management", "审批用户"
	}
	if method == "POST" && path == "/api/admin/users" {
		return "user.create", "user_management", "创建用户"
	}
	if method == "DELETE" && path == "/api/admin/users/:id" {
		return "user.delete", "user_management", "删除用户"
	}
	if method == "GET" && strings.HasPrefix(path, "/api/admin/users") {
		return "user.list", "user_management", "查看用户列表"
	}

	if strings.HasPrefix(path, "/api/admin/tags") {
		if method == "POST" {
			return "config.tag_create", "configuration", "创建标签"
		}
		if method == "PUT" {
			return "config.tag_update", "configuration", "更新标签"
		}
		if method == "DELETE" {
			return "config.tag_delete", "configuration", "删除标签"
		}
	}

	if strings.HasPrefix(path, "/api/admin/video-tags") {
		if method == "POST" {
			return "config.tag_create", "configuration", "创建视频标签"
		}
		if method == "PUT" || method == "PATCH" {
			return "config.tag_update", "configuration", "更新视频标签"
		}
		if method == "DELETE" {
			return "config.tag_delete", "configuration", "删除视频标签"
		}
	}

	if strings.HasPrefix(path, "/api/admin/moderation-rules") {
		if method == "POST" {
			return "config.rule_create", "configuration", "创建规则"
		}
		if method == "PUT" {
			return "config.rule_update", "configuration", "更新规则"
		}
		if method == "DELETE" {
			return "config.rule_delete", "configuration", "删除规则"
		}
	}

	if method == "PUT" && path == "/api/admin/docs/:key" {
		return "docs.update", "configuration", "更新系统文档"
	}

	if method == "DELETE" && path == "/api/admin/ai-review/jobs/:id/tasks" {
		return "ai_review.tasks_delete", "ai_review", "清空AI审核任务"
	}

	if strings.HasPrefix(path, "/api/admin/task-queues") {
		if method == "POST" {
			return "config.queue_create", "configuration", "创建队列"
		}
		if method == "PUT" {
			return "config.queue_update", "configuration", "更新队列"
		}
		if method == "DELETE" {
			return "config.queue_delete", "configuration", "删除队列"
		}
	}

	if strings.HasPrefix(path, "/api/tasks/") {
		if method == "POST" && strings.HasSuffix(path, "/claim") {
			return "review.claim", "content_moderation", "领取审核任务"
		}
		if method == "POST" && strings.HasSuffix(path, "/submit-batch") {
			return "review.submit_batch", "content_moderation", "批量提交审核"
		}
		if method == "POST" && strings.HasSuffix(path, "/submit") {
			return "review.submit", "content_moderation", "提交审核结果"
		}
		if method == "POST" && strings.HasSuffix(path, "/return") {
			return "review.return", "content_moderation", "退回任务"
		}
		if strings.Contains(path, "/quality-check") {
			return "review.quality_check", "content_moderation", "质检操作"
		}
		if strings.Contains(path, "/ai-human-diff") {
			return "review.ai_human_diff", "content_moderation", "AI与人工差异审核"
		}
	}

	if strings.HasPrefix(path, "/api/video/") {
		if method == "POST" && strings.HasSuffix(path, "/tasks/claim") {
			return "video.queue_assign", "content_moderation", "领取视频队列任务"
		}
		if method == "POST" && (strings.HasSuffix(path, "/tasks/submit") || strings.HasSuffix(path, "/tasks/submit-batch")) {
			return "video.review", "content_moderation", "提交视频审核"
		}
		if method == "POST" && strings.HasSuffix(path, "/tasks/return") {
			return "review.return", "content_moderation", "退回视频任务"
		}
	}

	if path == "/api/admin/videos/import" && method == "POST" {
		return "video.import", "content_moderation", "导入视频"
	}

	if strings.HasPrefix(path, "/api/admin/ai-review/jobs") {
		if method == "POST" && strings.HasSuffix(path, "/start") {
			return "ai.job_start", "content_moderation", "启动AI审核"
		}
		if method == "POST" {
			return "ai.job_create", "content_moderation", "创建AI审核任务"
		}
		if method == "GET" {
			return "ai.job_list", "content_moderation", "查看AI审核任务"
		}
	}
	if strings.HasPrefix(path, "/api/admin/ai-review/compare") {
		return "ai.result_review", "content_moderation", "AI审核对比分析"
	}

	if strings.HasPrefix(path, "/api/admin/audit-logs/export") {
		return "data.export", "data_operation", "导出审计日志"
	}
	if strings.HasPrefix(path, "/api/admin/audit-logs") {
		return "audit.logs.read", "system_operation", "查询审计日志"
	}

	return "api.request", "system_operation", "API请求"
}

func resolveResource(path string, c *gin.Context) (string, string) {
	if strings.Contains(path, "/users") {
		return "user", resolveParamID(c, "id")
	}
	if strings.Contains(path, "/permissions") {
		return "permission", resolveParamID(c, "id")
	}
	if strings.Contains(path, "/moderation-rules") {
		if id := resolveParamID(c, "id"); id != "" {
			return "moderation_rule", id
		}
		if code := c.Param("code"); code != "" {
			return "moderation_rule", code
		}
		return "moderation_rule", ""
	}
	if strings.Contains(path, "/tags") {
		return "tag", resolveParamID(c, "id")
	}
	if strings.Contains(path, "/task-queues") || strings.Contains(path, "/queues") {
		return "queue", resolveParamID(c, "id")
	}
	if strings.Contains(path, "/tasks") {
		return "review_task", resolveParamID(c, "id")
	}
	if strings.Contains(path, "/videos") {
		return "video", resolveParamID(c, "id")
	}
	if strings.Contains(path, "/audit-logs") {
		return "audit_log", resolveParamID(c, "id")
	}

	return "", ""
}

func resolveParamID(c *gin.Context, key string) string {
	if value := strings.TrimSpace(c.Param(key)); value != "" {
		return value
	}
	return ""
}

func extractResourceIDs(body interface{}) json.RawMessage {
	bodyMap, ok := body.(map[string]interface{})
	if !ok {
		return nil
	}

	keys := []string{"task_ids", "ids", "permission_keys", "video_ids", "user_ids"}
	for _, key := range keys {
		if value, exists := bodyMap[key]; exists {
			raw, _ := json.Marshal(value)
			return raw
		}
	}

	return nil
}

func buildChanges(method string, body interface{}) json.RawMessage {
	if body == nil {
		return nil
	}
	if method != "PUT" && method != "PATCH" && method != "DELETE" {
		return nil
	}

	payload := map[string]interface{}{
		"after": body,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil
	}
	return raw
}

func deriveResult(statusCode int) string {
	if statusCode == 206 || statusCode == 207 {
		return "partial"
	}
	if statusCode >= 200 && statusCode < 400 {
		return "success"
	}
	return "failure"
}

func collectErrors(c *gin.Context, responseBody json.RawMessage) string {
	if c.Writer.Status() < 400 {
		return ""
	}

	if errMsg := getErrorMessage(c); errMsg != "" {
		return errMsg
	}

	if errorsValue, exists := c.Get("errors"); exists {
		if errs, ok := errorsValue.([]string); ok && len(errs) > 0 {
			return strings.Join(errs, "; ")
		}
	}

	if errText := extractErrorMessage(responseBody); errText != "" {
		return errText
	}

	return ""
}

func collectErrorMeta(c *gin.Context, responseBody json.RawMessage, errorMessage, errorStack string) (string, string, string) {
	code := GetErrorCode(c)
	if code == "" {
		code = extractErrorCode(responseBody)
	}
	if code == "" {
		code = deriveErrorCode(c.Writer.Status(), errorMessage, errorStack)
	}

	errType := GetErrorType(c)
	if errType == "" {
		errType = extractErrorType(responseBody)
	}
	if errType == "" {
		errType = deriveErrorType(c.Writer.Status(), code, errorMessage, errorStack)
	}

	description := GetErrorDescription(c)
	if description == "" {
		description = extractErrorDescription(responseBody)
	}
	if description == "" {
		description = errorMessage
	}

	return code, errType, description
}

func extractErrorMessage(responseBody json.RawMessage) string {
	if len(responseBody) == 0 {
		return ""
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(responseBody, &payload); err != nil {
		return ""
	}

	if errVal, ok := payload["error"]; ok {
		if errText, ok := errVal.(string); ok {
			return errText
		}
	}
	if msgVal, ok := payload["message"]; ok {
		if msgText, ok := msgVal.(string); ok {
			return msgText
		}
	}
	if descVal, ok := payload["error_description"]; ok {
		if descText, ok := descVal.(string); ok {
			return descText
		}
	}
	if descVal, ok := payload["description"]; ok {
		if descText, ok := descVal.(string); ok {
			return descText
		}
	}

	return ""
}

func extractErrorCode(responseBody json.RawMessage) string {
	if len(responseBody) == 0 {
		return ""
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(responseBody, &payload); err != nil {
		return ""
	}

	if codeVal, ok := payload["code"]; ok {
		switch typed := codeVal.(type) {
		case string:
			return typed
		case float64:
			return fmt.Sprintf("%.0f", typed)
		default:
			return fmt.Sprintf("%v", typed)
		}
	}

	return ""
}

func extractErrorType(responseBody json.RawMessage) string {
	if len(responseBody) == 0 {
		return ""
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(responseBody, &payload); err != nil {
		return ""
	}

	if typeVal, ok := payload["error_type"]; ok {
		if typeText, ok := typeVal.(string); ok {
			return typeText
		}
	}

	return ""
}

func extractErrorDescription(responseBody json.RawMessage) string {
	if len(responseBody) == 0 {
		return ""
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(responseBody, &payload); err != nil {
		return ""
	}

	if descVal, ok := payload["error_description"]; ok {
		if descText, ok := descVal.(string); ok {
			return descText
		}
	}

	if descVal, ok := payload["description"]; ok {
		if descText, ok := descVal.(string); ok {
			return descText
		}
	}

	return ""
}

func deriveErrorCode(status int, message, stack string) string {
	if hasTimeoutHint(message) || hasTimeoutHint(stack) {
		return "SQL_TIMEOUT"
	}

	switch status {
	case http.StatusBadRequest:
		return "INVALID_REQUEST"
	case http.StatusUnauthorized:
		return "UNAUTHORIZED"
	case http.StatusForbidden:
		return "PERMISSION_DENIED"
	case http.StatusNotFound:
		return "NOT_FOUND"
	case http.StatusTooManyRequests:
		return "RATE_LIMIT_EXCEEDED"
	case http.StatusGatewayTimeout, http.StatusRequestTimeout:
		return "REQUEST_TIMEOUT"
	default:
		if status >= http.StatusInternalServerError {
			return "INTERNAL_ERROR"
		}
	}

	return ""
}

func deriveErrorType(status int, code, message, stack string) string {
	switch {
	case code == "INVALID_REQUEST" || status == http.StatusBadRequest:
		return "invalid_request"
	case code == "UNAUTHORIZED" || status == http.StatusUnauthorized:
		return "unauthorized"
	case code == "PERMISSION_DENIED" || status == http.StatusForbidden:
		return "permission_denied"
	case code == "NOT_FOUND" || status == http.StatusNotFound:
		return "not_found"
	case code == "RATE_LIMIT_EXCEEDED" || status == http.StatusTooManyRequests:
		return "rate_limited"
	case code == "SQL_TIMEOUT" || hasTimeoutHint(message) || hasTimeoutHint(stack):
		return "sql_timeout"
	case status >= http.StatusInternalServerError:
		return "internal_error"
	default:
		return "error"
	}
}

func hasTimeoutHint(value string) bool {
	if value == "" {
		return false
	}
	lower := strings.ToLower(value)
	return strings.Contains(lower, "timeout") || strings.Contains(lower, "timed out") || strings.Contains(lower, "deadline") || strings.Contains(lower, "statement timeout")
}

func sanitizeResponseBody(c *gin.Context, bodyWriter *bodyLogWriter) json.RawMessage {
	contentType := c.Writer.Header().Get("Content-Type")
	if !shouldCapturePayload(contentType) {
		return nil
	}

	responseBytes := bodyWriter.buffer.Bytes()
	if len(responseBytes) == 0 {
		return nil
	}

	_, sanitized := sanitizeJSONPayload(responseBytes)
	return sanitized
}

func buildRequestParams(c *gin.Context, requestBody interface{}) json.RawMessage {
	params := map[string]interface{}{}

	if len(c.Params) > 0 {
		pathParams := map[string]interface{}{}
		for _, param := range c.Params {
			pathParams[param.Key] = param.Value
		}
		params["path"] = sanitizeValue(pathParams)
	}

	queryValues := c.Request.URL.Query()
	if len(queryValues) > 0 {
		queryParams := map[string]interface{}{}
		for key, values := range queryValues {
			if len(values) == 1 {
				queryParams[key] = values[0]
			} else if len(values) > 1 {
				queryParams[key] = values
			}
		}
		if len(queryParams) > 0 {
			params["query"] = sanitizeValue(queryParams)
		}
	}

	if requestBody != nil {
		params["body"] = requestBody
	}

	if len(params) == 0 {
		return nil
	}

	raw, err := json.Marshal(params)
	if err != nil {
		return nil
	}

	if len(raw) > maxAuditTextFieldBytes {
		truncated := map[string]interface{}{
			"truncated": true,
		}
		raw, _ = json.Marshal(truncated)
	}

	return raw
}

func shouldCapturePayload(contentType string) bool {
	if contentType == "" {
		return false
	}
	contentType = strings.ToLower(contentType)
	return strings.Contains(contentType, "application/json") || strings.Contains(contentType, "+json")
}

func readBodyWithLimit(reader io.ReadCloser, limit int) ([]byte, error) {
	limited := io.LimitReader(reader, int64(limit))
	return io.ReadAll(limited)
}

func sanitizeJSONPayload(payload []byte) (interface{}, json.RawMessage) {
	if len(payload) == 0 {
		return nil, nil
	}

	trimmed := bytes.TrimSpace(payload)
	if len(trimmed) == 0 {
		return nil, nil
	}

	var value interface{}
	if err := json.Unmarshal(trimmed, &value); err != nil {
		return nil, nil
	}

	sanitizedValue := sanitizeValue(value)
	sanitizedJSON, err := json.Marshal(sanitizedValue)
	if err != nil {
		return nil, nil
	}

	if len(sanitizedJSON) > maxAuditTextFieldBytes {
		truncated := map[string]interface{}{
			"truncated": true,
		}
		sanitizedJSON, _ = json.Marshal(truncated)
	}

	return sanitizedValue, sanitizedJSON
}

func sanitizeValue(value interface{}) interface{} {
	switch typed := value.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{}, len(typed))
		for key, val := range typed {
			result[key] = sanitizeField(key, val)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(typed))
		for i, val := range typed {
			result[i] = sanitizeValue(val)
		}
		return result
	default:
		return value
	}
}

func sanitizeField(key string, value interface{}) interface{} {
	lowerKey := strings.ToLower(key)
	switch {
	case strings.Contains(lowerKey, "password"),
		strings.Contains(lowerKey, "token"),
		strings.Contains(lowerKey, "secret"),
		strings.Contains(lowerKey, "authorization"):
		return "***"
	case strings.Contains(lowerKey, "email"):
		if str, ok := value.(string); ok {
			return maskEmail(str)
		}
	case strings.Contains(lowerKey, "phone"),
		strings.Contains(lowerKey, "mobile"):
		if str, ok := value.(string); ok {
			return maskPhone(str)
		}
	case strings.Contains(lowerKey, "idcard"),
		strings.Contains(lowerKey, "id_card"),
		strings.Contains(lowerKey, "identity"):
		if str, ok := value.(string); ok {
			return maskID(str)
		}
	}

	return sanitizeValue(value)
}

func maskEmail(value string) string {
	parts := strings.Split(value, "@")
	if len(parts) != 2 {
		return "***"
	}
	name := parts[0]
	if len(name) <= 3 {
		return name + "***@" + parts[1]
	}
	return name[:3] + "***@" + parts[1]
}

func maskPhone(value string) string {
	if len(value) <= 7 {
		return "***"
	}
	return value[:3] + "****" + value[len(value)-4:]
}

func maskID(value string) string {
	if len(value) <= 10 {
		return "***"
	}
	return value[:6] + "****" + value[len(value)-4:]
}

func truncateString(value string, max int) string {
	if len(value) <= max {
		return value
	}
	return value[:max]
}

func fallbackString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func resolveModuleName(c *gin.Context) string {
	handlerName := strings.TrimSpace(c.HandlerName())
	if handlerName == "" {
		return ""
	}
	module, _ := splitHandlerName(handlerName)
	return module
}

func resolveMethodName(c *gin.Context) string {
	handlerName := strings.TrimSpace(c.HandlerName())
	if handlerName == "" {
		return ""
	}
	_, method := splitHandlerName(handlerName)
	return method
}

func splitHandlerName(handlerName string) (string, string) {
	lastDot := strings.LastIndex(handlerName, ".")
	if lastDot <= 0 || lastDot == len(handlerName)-1 {
		return handlerName, ""
	}
	return handlerName[:lastDot], handlerName[lastDot+1:]
}

var serverIPOnce sync.Once
var cachedServerIP string

func resolveServerIP(c *gin.Context) string {
	host := strings.TrimSpace(c.Request.Host)
	if host != "" {
		if hostPart, _, err := net.SplitHostPort(host); err == nil {
			if ip := net.ParseIP(hostPart); ip != nil {
				return ip.String()
			}
		} else if ip := net.ParseIP(host); ip != nil {
			return ip.String()
		}
	}

	serverIPOnce.Do(func() {
		cachedServerIP = resolveLocalIP()
	})
	return cachedServerIP
}

func resolveServerPort(c *gin.Context) string {
	host := strings.TrimSpace(c.Request.Host)
	if host != "" {
		if _, port, err := net.SplitHostPort(host); err == nil {
			return port
		}
	}
	if config.AppConfig != nil && config.AppConfig.Port != "" {
		return config.AppConfig.Port
	}
	return ""
}

func resolvePageURL(c *gin.Context) string {
	pageURL := strings.TrimSpace(c.GetHeader("X-Page-Url"))
	if pageURL == "" {
		pageURL = strings.TrimSpace(c.GetHeader("X-Page-URL"))
	}
	if pageURL == "" {
		pageURL = strings.TrimSpace(c.GetHeader("Referer"))
	}
	return truncateString(pageURL, 2048)
}

func resolveLocalIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch typed := addr.(type) {
			case *net.IPNet:
				ip = typed.IP
			case *net.IPAddr:
				ip = typed.IP
			default:
				continue
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			if ipv4 := ip.To4(); ipv4 != nil {
				return ipv4.String()
			}
		}
	}
	return ""
}

func getRawJSONFromContext(c *gin.Context, key string) json.RawMessage {
	if value, ok := c.Get(key); ok {
		if raw, ok := value.(json.RawMessage); ok {
			return raw
		}
	}
	return nil
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

type limitedBuffer struct {
	buf      bytes.Buffer
	limit    int
	truncate bool
}

func newLimitedBuffer(limit int) *limitedBuffer {
	return &limitedBuffer{limit: limit}
}

func (l *limitedBuffer) Write(p []byte) (int, error) {
	if l.limit <= 0 {
		return len(p), nil
	}

	remaining := l.limit - l.buf.Len()
	if remaining <= 0 {
		l.truncate = true
		return len(p), nil
	}

	if len(p) > remaining {
		_, _ = l.buf.Write(p[:remaining])
		l.truncate = true
		return len(p), nil
	}

	_, _ = l.buf.Write(p)
	return len(p), nil
}

func (l *limitedBuffer) Bytes() []byte {
	return l.buf.Bytes()
}

type bodyLogWriter struct {
	gin.ResponseWriter
	buffer *limitedBuffer
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	_, _ = w.buffer.Write(b)
	return w.ResponseWriter.Write(b)
}
