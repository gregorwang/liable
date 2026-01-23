package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	traceIDKey    = "trace_id"
	requestIDKey  = "request_id"
	traceHeader   = "X-Trace-Id"
	requestHeader = "X-Request-Id"
)

func newTraceID() string {
	return uuid.New().String()
}

// TraceMiddleware ensures every request has a trace_id and exposes it in headers.
func TraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := strings.TrimSpace(c.GetHeader(traceHeader))
		if traceID == "" {
			traceID = strings.TrimSpace(c.GetHeader(requestHeader))
		}
		if traceID == "" {
			traceID = newTraceID()
		}

		c.Set(traceIDKey, traceID)
		c.Set(requestIDKey, traceID)
		c.Writer.Header().Set(traceHeader, traceID)
		c.Writer.Header().Set(requestHeader, traceID)

		c.Next()
	}
}

// GetTraceID retrieves trace_id from context.
func GetTraceID(c *gin.Context) string {
	if value, ok := c.Get(traceIDKey); ok {
		if traceID, ok := value.(string); ok {
			return traceID
		}
	}
	return ""
}
