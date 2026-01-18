// Package base provides common handler utilities and response functions
// for the comment review platform.
package base

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorResponse 统一错误响应结构
// 所有 API 错误响应都应使用此格式，确保前端可以统一处理
type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

// SuccessResponse 统一成功响应结构（可选使用）
type SuccessResponse struct {
	Data interface{} `json:"data,omitempty"`
}

// 预定义错误码
const (
	ErrCodeInvalidRequest    = "INVALID_REQUEST"
	ErrCodeUnauthorized      = "UNAUTHORIZED"
	ErrCodeNotFound          = "NOT_FOUND"
	ErrCodeClaimFailed       = "CLAIM_FAILED"
	ErrCodeSubmitFailed      = "SUBMIT_FAILED"
	ErrCodeReturnFailed      = "RETURN_FAILED"
	ErrCodeFetchFailed       = "FETCH_FAILED"
	ErrCodeInternalError     = "INTERNAL_ERROR"
	ErrCodeRateLimitExceeded = "RATE_LIMIT_EXCEEDED"
)

// RespondError 返回错误响应
// status: HTTP 状态码
// code: 错误码，用于前端识别错误类型
// message: 错误消息，用于显示给用户
func RespondError(c *gin.Context, status int, code string, message string) {
	c.JSON(status, ErrorResponse{
		Error: message,
		Code:  code,
	})
}

// RespondSuccess 返回成功响应
// data: 响应数据，可以是任意类型
func RespondSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}

// RespondBadRequest 返回 400 错误（客户端请求错误）
// 用于参数验证失败、业务规则校验失败等场景
func RespondBadRequest(c *gin.Context, code string, message string) {
	RespondError(c, http.StatusBadRequest, code, message)
}

// RespondInternalError 返回 500 错误（服务器内部错误）
// 用于数据库错误、外部服务调用失败等场景
func RespondInternalError(c *gin.Context, code string, message string) {
	RespondError(c, http.StatusInternalServerError, code, message)
}

// RespondUnauthorized 返回 401 错误（未授权）
// 用于用户未登录或 token 无效的场景
func RespondUnauthorized(c *gin.Context, message string) {
	RespondError(c, http.StatusUnauthorized, ErrCodeUnauthorized, message)
}

// RespondNotFound 返回 404 错误（资源不存在）
// 用于请求的资源不存在的场景
func RespondNotFound(c *gin.Context, message string) {
	RespondError(c, http.StatusNotFound, ErrCodeNotFound, message)
}

// RespondTooManyRequests 返回 429 错误（请求过多）
// 用于限流场景
func RespondTooManyRequests(c *gin.Context, message string) {
	RespondError(c, http.StatusTooManyRequests, ErrCodeRateLimitExceeded, message)
}
