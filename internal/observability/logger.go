package observability

import (
	"log"
	"runtime/debug"
)

// Infof logs an info message with trace context.
func Infof(traceID string, format string, args ...interface{}) {
	if traceID == "" {
		log.Printf(format, args...)
		return
	}
	log.Printf("trace_id=%s "+format, append([]interface{}{traceID}, args...)...)
}

// ErrorWithStack logs an error with stack trace and trace context.
func ErrorWithStack(traceID string, err error) string {
	stack := string(debug.Stack())
	if err == nil {
		Infof(traceID, "error stack captured")
		return stack
	}
	Infof(traceID, "error=%v stack=%s", err, stack)
	return stack
}
