package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// Error types
var (
	ErrNotFound           = errors.New("resource not found")
	ErrUnauthorized       = errors.New("unauthorized access")
	ErrForbidden          = errors.New("access forbidden")
	ErrValidation         = errors.New("validation failed")
	ErrRateLimit          = errors.New("rate limit exceeded")
	ErrServerError        = errors.New("internal server error")
	ErrBadRequest         = errors.New("bad request")
	ErrConflict           = errors.New("resource conflict")
	ErrServiceUnavailable = errors.New("service unavailable")
)

// ForwardEmailError represents a Forward Email API error
type ForwardEmailError struct {
	Type       string `json:"type"`
	Message    string `json:"message"`
	Code       string `json:"code,omitempty"`
	StatusCode int    `json:"status_code"`
	Details    string `json:"details,omitempty"`
}

func (e *ForwardEmailError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("%s (%s): %s", e.Type, e.Code, e.Message)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the underlying error type
func (e *ForwardEmailError) Unwrap() error {
	switch e.StatusCode {
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusForbidden:
		return ErrForbidden
	case http.StatusBadRequest:
		return ErrBadRequest
	case http.StatusConflict:
		return ErrConflict
	case http.StatusTooManyRequests:
		return ErrRateLimit
	case http.StatusServiceUnavailable:
		return ErrServiceUnavailable
	case http.StatusInternalServerError:
		return ErrServerError
	default:
		if e.StatusCode >= 400 && e.StatusCode < 500 {
			return ErrValidation
		}
		return ErrServerError
	}
}

// NewForwardEmailError creates a new Forward Email error
func NewForwardEmailError(statusCode int, message, code string) *ForwardEmailError {
	errorType := getErrorType(statusCode)
	return &ForwardEmailError{
		Type:       errorType,
		Message:    message,
		Code:       code,
		StatusCode: statusCode,
	}
}

// NewNotFoundError creates a resource not found error
func NewNotFoundError(resource string) *ForwardEmailError {
	return &ForwardEmailError{
		Type:       "NotFound",
		Message:    fmt.Sprintf("%s not found", resource),
		StatusCode: http.StatusNotFound,
	}
}

// NewValidationError creates a validation error
func NewValidationError(message string) *ForwardEmailError {
	return &ForwardEmailError{
		Type:       "ValidationError",
		Message:    message,
		StatusCode: http.StatusBadRequest,
	}
}

// NewUnauthorizedError creates an unauthorized error
func NewUnauthorizedError(message string) *ForwardEmailError {
	if message == "" {
		message = "Authentication required"
	}
	return &ForwardEmailError{
		Type:       "Unauthorized",
		Message:    message,
		StatusCode: http.StatusUnauthorized,
	}
}

// NewForbiddenError creates a forbidden access error
func NewForbiddenError(message string) *ForwardEmailError {
	if message == "" {
		message = "Access forbidden"
	}
	return &ForwardEmailError{
		Type:       "Forbidden",
		Message:    message,
		StatusCode: http.StatusForbidden,
	}
}

// NewRateLimitError creates a rate limit error
func NewRateLimitError(retryAfter string) *ForwardEmailError {
	message := "Rate limit exceeded"
	if retryAfter != "" {
		message = fmt.Sprintf("Rate limit exceeded. Retry after %s", retryAfter)
	}
	return &ForwardEmailError{
		Type:       "RateLimit",
		Message:    message,
		StatusCode: http.StatusTooManyRequests,
		Details:    retryAfter,
	}
}

// NewServerError creates a server error
func NewServerError(message string) *ForwardEmailError {
	if message == "" {
		message = "Internal server error"
	}
	return &ForwardEmailError{
		Type:       "ServerError",
		Message:    message,
		StatusCode: http.StatusInternalServerError,
	}
}

// NewServiceUnavailableError creates a service unavailable error
func NewServiceUnavailableError(message string) *ForwardEmailError {
	if message == "" {
		message = "Service temporarily unavailable"
	}
	return &ForwardEmailError{
		Type:       "ServiceUnavailable",
		Message:    message,
		StatusCode: http.StatusServiceUnavailable,
	}
}

// getErrorType returns the error type based on status code
func getErrorType(statusCode int) string {
	switch statusCode {
	case http.StatusBadRequest:
		return "BadRequest"
	case http.StatusUnauthorized:
		return "Unauthorized"
	case http.StatusForbidden:
		return "Forbidden"
	case http.StatusNotFound:
		return "NotFound"
	case http.StatusConflict:
		return "Conflict"
	case http.StatusTooManyRequests:
		return "RateLimit"
	case http.StatusInternalServerError:
		return "ServerError"
	case http.StatusServiceUnavailable:
		return "ServiceUnavailable"
	default:
		if statusCode >= 400 && statusCode < 500 {
			return "ClientError"
		}
		return "ServerError"
	}
}

// IsNotFound checks if an error is a not found error
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsUnauthorized checks if an error is an unauthorized error
func IsUnauthorized(err error) bool {
	return errors.Is(err, ErrUnauthorized)
}

// IsForbidden checks if an error is a forbidden error
func IsForbidden(err error) bool {
	return errors.Is(err, ErrForbidden)
}

// IsValidation checks if an error is a validation error
func IsValidation(err error) bool {
	return errors.Is(err, ErrValidation)
}

// IsRateLimit checks if an error is a rate limit error
func IsRateLimit(err error) bool {
	return errors.Is(err, ErrRateLimit)
}

// IsServerError checks if an error is a server error
func IsServerError(err error) bool {
	return errors.Is(err, ErrServerError)
}

// IsServiceUnavailable checks if an error is a service unavailable error
func IsServiceUnavailable(err error) bool {
	return errors.Is(err, ErrServiceUnavailable)
}

// IsRetryable checks if an error is retryable
func IsRetryable(err error) bool {
	return IsRateLimit(err) || IsServiceUnavailable(err) || IsServerError(err)
}

// GetStatusCode extracts the HTTP status code from an error
func GetStatusCode(err error) int {
	var feErr *ForwardEmailError
	if errors.As(err, &feErr) {
		return feErr.StatusCode
	}
	return 0
}

// GetErrorCode extracts the error code from an error
func GetErrorCode(err error) string {
	var feErr *ForwardEmailError
	if errors.As(err, &feErr) {
		return feErr.Code
	}
	return ""
}

// GetErrorDetails extracts error details from an error
func GetErrorDetails(err error) string {
	var feErr *ForwardEmailError
	if errors.As(err, &feErr) {
		return feErr.Details
	}
	return ""
}
