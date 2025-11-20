package errors

import (
	"fmt"
	"strings"
)

// 错误类型常量
const (
	ErrTypeRequest    = "REQUEST_ERROR"    // 请求错误
	ErrTypeResponse   = "RESPONSE_ERROR"   // 响应错误
	ErrTypeParse      = "PARSE_ERROR"      // 解析错误
	ErrTypeValidation = "VALIDATION_ERROR" // 验证错误
	ErrTypeTimeout    = "TIMEOUT_ERROR"    // 超时错误
	ErrTypeInternal   = "INTERNAL_ERROR"   // 内部错误
	ErrTypeNetwork    = "NETWORK_ERROR"    // 网络错误
	ErrTypeAPI        = "API_ERROR"
	ErrTypeStatusCode = "STATUS_CODE_ERROR" // 状态码错误
)

// AppError 自定义应用错误类型
type AppError struct {
	Type       string                 // 错误类型
	Message    string                 // 错误消息
	Err        error                  // 原始错误
	StatusCode int                    // HTTP状态码（如果适用）
	ErrorCode  string                 // 错误代码（如果适用）
	Details    map[string]interface{} // 错误详情
}

// Error 实现error接口
func (q *AppError) Error() string {
	details := []string{fmt.Sprintf("%s: %s", q.Type, q.Message)}

	if q.StatusCode > 0 {
		details = append(details, fmt.Sprintf("status=%d", q.StatusCode))
	}

	if q.ErrorCode != "" {
		details = append(details, fmt.Sprintf("code=%s", q.ErrorCode))
	}

	if q.Err != nil {
		details = append(details, fmt.Sprintf("error=%v", q.Err))
	}

	return strings.Join(details, ", ")
}

// Unwrap 实现errors.Unwrap接口
func (q *AppError) Unwrap() error {
	return q.Err
}

// NewRequestError 创建请求错误
func NewRequestError(message string, err error) *AppError {
	return &AppError{
		Type:    ErrTypeRequest,
		Message: message,
		Err:     err,
	}
}

// NewResponseError 创建响应错误
func NewResponseError(message string, err error) *AppError {
	return &AppError{
		Type:    ErrTypeResponse,
		Message: message,
		Err:     err,
	}
}

// NewResponseErrorWithStatus 创建带状态码的响应错误
func NewResponseErrorWithStatus(message string, statusCode int, err error) *AppError {
	return &AppError{
		Type:       ErrTypeResponse,
		Message:    message,
		Err:        err,
		StatusCode: statusCode,
	}
}

// NewNetworkError 创建网络错误
func NewNetworkError(message string, err error) *AppError {
	return &AppError{
		Type:    ErrTypeNetwork,
		Message: message,
		Err:     err,
	}
}

// NewAPIError 创建API错误
func NewAPIError(message string, errorCode string, err error) *AppError {
	return &AppError{
		Type:      ErrTypeAPI,
		Message:   message,
		ErrorCode: errorCode,
		Err:       err,
	}
}

// NewParseError 创建解析错误
func NewParseError(message string, err error) *AppError {
	return &AppError{
		Type:    ErrTypeParse,
		Message: message,
		Err:     err,
	}
}

// NewValidationError 创建验证错误
func NewValidationError(message string) *AppError {
	return &AppError{
		Type:    ErrTypeValidation,
		Message: message,
		Err:     nil,
	}
}

// NewTimeoutError 创建超时错误
func NewTimeoutError(message string) *AppError {
	return &AppError{
		Type:    ErrTypeTimeout,
		Message: message,
		Err:     nil,
	}
}

// NewInternalError 创建内部错误
func NewInternalError(message string, err error) *AppError {
	return &AppError{
		Type:    ErrTypeInternal,
		Message: message,
		Err:     err,
	}
}

// NewStatusCodeError 响应状态码错误
func NewStatusCodeError(message string) *AppError {
	return &AppError{
		Type:    ErrTypeStatusCode,
		Message: message,
		Err:     nil,
	}
}

// IsTimeoutError 检查是否为超时错误
func IsTimeoutError(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Type == ErrTypeTimeout
	}
	return false
}

// IsNetworkError 检查是否为网络错误
func IsNetworkError(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Type == ErrTypeNetwork || appErr.Type == ErrTypeRequest
	}
	return false
}

// IsStatusCodeError 检查是否为状态码错误
func IsStatusCodeError(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Type == ErrTypeStatusCode
	}
	return false
}
