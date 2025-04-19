package errors

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ErrorType string

const (
	InputError      ErrorType = "InputError"      // 输入相关错误
	ServiceError    ErrorType = "ServiceError"    // 服务调用错误
	PermissionError ErrorType = "PermissionError" // 权限错误
	TimeoutError    ErrorType = "TimeoutError"    // 超时错误
	UnknownError    ErrorType = "UnknownError"    // 未知错误
)

type AppError struct {
	Type    ErrorType // 错误的类别
	Message string    // 错误的描述
}

// 实现 Go 的 error 接口
func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

func (e *AppError) Status() *status.Status {
	var code codes.Code
	switch e.Type {
	case InputError:
		code = codes.InvalidArgument
	case ServiceError:
		code = codes.Unavailable
	case PermissionError:
		code = codes.PermissionDenied
	default:
		code = codes.Unknown
	}
	return status.New(code, e.Message)
}

func (e *AppError) As(target interface{}) bool {
	switch v := target.(type) {
	case **AppError:
		*v = e
		return true
	case *error:
		*v = e
		return true
	default:
		return false
	}
}

// 辅助方法：创建错误
func NewAppError(errorType ErrorType, message string) *AppError {
	return &AppError{
		Type:    errorType,
		Message: message,
	}
}

func NewInputError(message string) *AppError {
	return NewAppError(InputError, message)
}

func NewInputErrorf(format string, args ...interface{}) *AppError {
	return NewAppError(InputError, fmt.Sprintf(format, args...))
}

func NewServiceError(message string) *AppError {
	return NewAppError(ServiceError, message)
}

func NewServiceErrorf(format string, args ...interface{}) *AppError {
	return NewAppError(ServiceError, fmt.Sprintf(format, args...))
}

func NewPermissionError(message string) *AppError {
	return NewAppError(PermissionError, message)
}

func NewPermissionErrorf(format string, args ...interface{}) *AppError {
	return NewAppError(PermissionError, fmt.Sprintf(format, args...))
}
