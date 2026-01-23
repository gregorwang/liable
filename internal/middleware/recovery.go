package middleware

import (
	"fmt"
	"net/http"

	"comment-review-platform/internal/observability"

	"github.com/gin-gonic/gin"
)

// RecoveryMiddleware captures panics, records stack traces, and returns a safe error response.
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				traceID := GetTraceID(c)
				err := fmt.Errorf("%v", recovered)
				errMsg := err.Error()
				stack := observability.ErrorWithStack(traceID, err)

				setErrorDetails(c, errMsg, stack, true)

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error":    "internal server error",
					"trace_id": traceID,
				})
			}
		}()

		c.Next()
	}
}
