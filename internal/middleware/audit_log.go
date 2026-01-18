package middleware

import (
	"log"
	"time"

	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuditLog represents an audit log entry
type AuditLog struct {
	ID                int64     `json:"id"`
	RequestID         string    `json:"request_id"`
	Timestamp         time.Time `json:"timestamp"`

	// User information
	UserID            *int      `json:"user_id,omitempty"`
	Username          string    `json:"username,omitempty"`
	Role              string    `json:"role,omitempty"`

	// Request information
	IPAddress         string    `json:"ip_address"`
	UserAgent         string    `json:"user_agent"`
	Method            string    `json:"method"`
	Path              string    `json:"path"`
	QueryParams       string    `json:"query_params"`

	// Permission check
	PermissionChecked string    `json:"permission_checked,omitempty"`
	PermissionGranted bool      `json:"permission_granted"`

	// Response information
	StatusCode       int       `json:"status_code"`
	ResponseTimeMs   int       `json:"response_time_ms"`
	ErrorMessage     string    `json:"error_message,omitempty"`
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

// AuditLogMiddleware returns a middleware that logs audit information
func AuditLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		requestID := uuid.New().String()
		c.Set("request_id", requestID)

		// Process request
		c.Next()

		// Only log if audit logger is initialized
		if auditLogger == nil {
			return
		}

		// Build audit log entry
		auditEntry := buildAuditLogEntry(c, startTime, requestID)

		// Async write to avoid blocking response
		go func(logEntry AuditLog) {
			if err := auditLogger.Save(logEntry); err != nil {
				log.Printf("Failed to save audit log: %v", err)
			}
		}(auditEntry)
	}
}

// buildAuditLogEntry constructs an audit log entry from Gin context
func buildAuditLogEntry(c *gin.Context, startTime time.Time, requestID string) AuditLog {
	// Get user information from context
	var userID *int
	if uid := GetUserID(c); uid > 0 {
		userID = &uid
	}
	username := GetUsername(c)
	role := GetRole(c)

	// Get permission check info
	permissionChecked := ""
	permissionGranted := c.Writer.Status() != 403
	if perm, exists := c.Get("checked_permission"); exists {
		permissionChecked = perm.(string)
	}

	// Get query params
	queryParams := ""
	if len(c.Request.URL.Query()) > 0 {
		queryParams = c.Request.URL.Query().Encode()
	}

	// Sanitize user agent
	userAgent := c.GetHeader("User-Agent")
	if len(userAgent) > 500 {
		userAgent = userAgent[:500]
	}

	// Extract error message if request failed
	errorMessage := ""
	if c.Writer.Status() >= 400 {
		if errors, exists := c.Get("errors"); exists {
			if errs, ok := errors.([]string); ok && len(errs) > 0 {
				for i, err := range errs {
					if i > 0 {
						errorMessage += "; "
					}
					errorMessage += err
				}
			}
		}
	}

	responseTimeMs := int(time.Since(startTime).Milliseconds())

	return AuditLog{
		RequestID:         requestID,
		Timestamp:         startTime,
		UserID:           userID,
		Username:         username,
		Role:             role,
		IPAddress:        c.ClientIP(),
		UserAgent:        userAgent,
		Method:           c.Request.Method,
		Path:             c.Request.URL.Path,
		QueryParams:      queryParams,
		PermissionChecked: permissionChecked,
		PermissionGranted: permissionGranted,
		StatusCode:       c.Writer.Status(),
		ResponseTimeMs:   responseTimeMs,
		ErrorMessage:     errorMessage,
	}
}

// Save stores an audit log entry to the database
func (a *AuditLogger) Save(entry AuditLog) error {
	if a.db == nil {
		return nil // Database not initialized, skip logging
	}

	query := `
		INSERT INTO audit_logs (
			request_id, timestamp, user_id, username, role,
			ip_address, user_agent, method, path, query_params,
			permission_checked, permission_granted,
			status_code, response_time_ms, error_message
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	_, err := a.db.Exec(
		query,
		entry.RequestID,
		entry.Timestamp,
		entry.UserID,
		entry.Username,
		entry.Role,
		entry.IPAddress,
		entry.UserAgent,
		entry.Method,
		entry.Path,
		entry.QueryParams,
		entry.PermissionChecked,
		entry.PermissionGranted,
		entry.StatusCode,
		entry.ResponseTimeMs,
		entry.ErrorMessage,
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
		WHERE timestamp < NOW() - INTERVAL '1 day' * $1
	`

	result, err := a.db.Exec(query, retentionDays)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	log.Printf("Cleaned up %d audit logs older than %d days", rowsAffected, retentionDays)
	return nil
}
