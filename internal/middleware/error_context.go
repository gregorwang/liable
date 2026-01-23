package middleware

import "github.com/gin-gonic/gin"

const (
	errorMessageKey   = "error_message"
	errorStackKey     = "error_stack"
	panicRecoveredKey = "panic_recovered"
	errorCodeKey      = "error_code"
	errorTypeKey      = "error_type"
	errorDescKey      = "error_description"
)

func setErrorDetails(c *gin.Context, message, stack string, recovered bool) {
	if message != "" {
		c.Set(errorMessageKey, message)
	}
	if stack != "" {
		c.Set(errorStackKey, stack)
	}
	if recovered {
		c.Set(panicRecoveredKey, true)
	}
}

// SetErrorMetadata stores structured error metadata for later logging or response enrichment.
func SetErrorMetadata(c *gin.Context, code, errorType, description string) {
	if code != "" {
		c.Set(errorCodeKey, code)
	}
	if errorType != "" {
		c.Set(errorTypeKey, errorType)
	}
	if description != "" {
		c.Set(errorDescKey, description)
	}
}

func getErrorMessage(c *gin.Context) string {
	if value, ok := c.Get(errorMessageKey); ok {
		if msg, ok := value.(string); ok {
			return msg
		}
	}
	return ""
}

func getErrorStack(c *gin.Context) string {
	if value, ok := c.Get(errorStackKey); ok {
		if stack, ok := value.(string); ok {
			return stack
		}
	}
	return ""
}

func GetErrorCode(c *gin.Context) string {
	if value, ok := c.Get(errorCodeKey); ok {
		if code, ok := value.(string); ok {
			return code
		}
	}
	return ""
}

func GetErrorType(c *gin.Context) string {
	if value, ok := c.Get(errorTypeKey); ok {
		if errType, ok := value.(string); ok {
			return errType
		}
	}
	return ""
}

func GetErrorDescription(c *gin.Context) string {
	if value, ok := c.Get(errorDescKey); ok {
		if desc, ok := value.(string); ok {
			return desc
		}
	}
	return ""
}

func isPanicRecovered(c *gin.Context) bool {
	if value, ok := c.Get(panicRecoveredKey); ok {
		if recovered, ok := value.(bool); ok {
			return recovered
		}
	}
	return false
}
