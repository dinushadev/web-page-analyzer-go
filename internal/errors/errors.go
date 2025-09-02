package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorType represents the type of error
type ErrorType string

const (
	// Client errors (4xx)
	ErrorTypeValidation       ErrorType = "VALIDATION_ERROR"
	ErrorTypeBadRequest       ErrorType = "BAD_REQUEST"
	ErrorTypeUnauthorized     ErrorType = "UNAUTHORIZED"
	ErrorTypeForbidden        ErrorType = "FORBIDDEN"
	ErrorTypeNotFound         ErrorType = "NOT_FOUND"
	ErrorTypeMethodNotAllowed ErrorType = "METHOD_NOT_ALLOWED"
	ErrorTypeTimeout          ErrorType = "TIMEOUT"

	// Server errors (5xx)
	ErrorTypeInternal     ErrorType = "INTERNAL_ERROR"
	ErrorTypeUnavailable  ErrorType = "SERVICE_UNAVAILABLE"
	ErrorTypeUpstream     ErrorType = "UPSTREAM_ERROR"
	ErrorTypeParseFailure ErrorType = "PARSE_FAILURE"
)

// AppError represents a standardized application error
type AppError struct {
	Type       ErrorType `json:"type"`
	Message    string    `json:"message"`
	Details    string    `json:"details,omitempty"`
	StatusCode int       `json:"status_code"`
	Err        error     `json:"-"` // Original error, not exposed in JSON
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Type, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap allows errors.Is and errors.As to work
func (e *AppError) Unwrap() error {
	return e.Err
}

// HTTPResponse returns the error formatted for HTTP response
type HTTPResponse struct {
	Error      string    `json:"error"`
	Type       ErrorType `json:"type"`
	StatusCode int       `json:"status_code"`
	Details    string    `json:"details,omitempty"`
}

// ToHTTPResponse converts AppError to HTTP response format
func (e *AppError) ToHTTPResponse() HTTPResponse {
	return HTTPResponse{
		Error:      e.Message,
		Type:       e.Type,
		StatusCode: e.StatusCode,
		Details:    e.Details,
	}
}

// Common error constructors

// NewValidationError creates a new validation error
func NewValidationError(message string, details ...string) *AppError {
	det := ""
	if len(details) > 0 {
		det = details[0]
	}
	return &AppError{
		Type:       ErrorTypeValidation,
		Message:    message,
		Details:    det,
		StatusCode: http.StatusBadRequest,
	}
}

// NewBadRequestError creates a new bad request error
func NewBadRequestError(message string, err error) *AppError {
	return &AppError{
		Type:       ErrorTypeBadRequest,
		Message:    message,
		StatusCode: http.StatusBadRequest,
		Err:        err,
	}
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Type:       ErrorTypeNotFound,
		Message:    fmt.Sprintf("%s not found", resource),
		StatusCode: http.StatusNotFound,
	}
}

// NewMethodNotAllowedError creates a new method not allowed error
func NewMethodNotAllowedError(allowed []string) *AppError {
	return &AppError{
		Type:       ErrorTypeMethodNotAllowed,
		Message:    "Method not allowed",
		Details:    fmt.Sprintf("allowed methods: %v", allowed),
		StatusCode: http.StatusMethodNotAllowed,
	}
}

// NewTimeoutError creates a new timeout error
func NewTimeoutError(operation string, err error) *AppError {
	return &AppError{
		Type:       ErrorTypeTimeout,
		Message:    fmt.Sprintf("%s timed out", operation),
		StatusCode: http.StatusGatewayTimeout,
		Err:        err,
	}
}

// NewInternalError creates a new internal server error
func NewInternalError(message string, err error) *AppError {
	return &AppError{
		Type:       ErrorTypeInternal,
		Message:    message,
		StatusCode: http.StatusInternalServerError,
		Err:        err,
	}
}

// NewUpstreamError creates a new upstream error
func NewUpstreamError(service string, statusCode int, err error) *AppError {
	return &AppError{
		Type:       ErrorTypeUpstream,
		Message:    fmt.Sprintf("upstream service %s error", service),
		Details:    fmt.Sprintf("status code: %d", statusCode),
		StatusCode: http.StatusBadGateway,
		Err:        err,
	}
}

// NewParseError creates a new parse error
func NewParseError(format string, err error) *AppError {
	return &AppError{
		Type:       ErrorTypeParseFailure,
		Message:    fmt.Sprintf("failed to parse %s", format),
		StatusCode: http.StatusInternalServerError,
		Err:        err,
	}
}

// WrapError wraps an existing error with AppError
func WrapError(err error, errType ErrorType, message string) *AppError {
	if appErr, ok := err.(*AppError); ok {
		// If it's already an AppError, preserve the original type but update message
		appErr.Message = message
		return appErr
	}

	statusCode := http.StatusInternalServerError
	switch errType {
	case ErrorTypeValidation, ErrorTypeBadRequest:
		statusCode = http.StatusBadRequest
	case ErrorTypeUnauthorized:
		statusCode = http.StatusUnauthorized
	case ErrorTypeForbidden:
		statusCode = http.StatusForbidden
	case ErrorTypeNotFound:
		statusCode = http.StatusNotFound
	case ErrorTypeMethodNotAllowed:
		statusCode = http.StatusMethodNotAllowed
	case ErrorTypeTimeout:
		statusCode = http.StatusGatewayTimeout
	case ErrorTypeUnavailable:
		statusCode = http.StatusServiceUnavailable
	case ErrorTypeUpstream:
		statusCode = http.StatusBadGateway
	}

	return &AppError{
		Type:       errType,
		Message:    message,
		StatusCode: statusCode,
		Err:        err,
	}
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// GetAppError extracts AppError from an error
func GetAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

// Sentinel errors for backward compatibility
var (
	ErrUnreachable = NewUpstreamError("target", 0, errors.New("url is unreachable"))
	ErrInvalidURL  = NewValidationError("invalid url")
	ErrUpstream    = NewUpstreamError("http", 0, errors.New("upstream http error"))
	ErrParseHTML   = NewParseError("html", errors.New("failed to parse html"))
	ErrTimeout     = NewTimeoutError("request", errors.New("upstream timeout"))
)
