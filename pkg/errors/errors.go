package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// Standard error types for Forward Email API responses.
// These sentinel errors can be used with errors.Is() for error type checking
// and provide a consistent way to handle different classes of API errors.
var (
	ErrNotFound           = errors.New("resource not found")    // 404 - Resource does not exist
	ErrUnauthorized       = errors.New("unauthorized access")   // 401 - Authentication required
	ErrForbidden          = errors.New("access forbidden")      // 403 - Access denied
	ErrValidation         = errors.New("validation failed")     // 400/422 - Input validation errors
	ErrRateLimit          = errors.New("rate limit exceeded")   // 429 - Too many requests
	ErrServerError        = errors.New("internal server error") // 500 - Server-side error
	ErrBadRequest         = errors.New("bad request")           // 400 - Malformed request
	ErrConflict           = errors.New("resource conflict")     // 409 - Resource already exists
	ErrServiceUnavailable = errors.New("service unavailable")   // 503 - Service temporarily down
)

// ForwardEmailError represents a structured error from the Forward Email API.
// It captures both HTTP status information and API-specific error details,
// providing rich context for error handling and user-friendly error messages.
type ForwardEmailError struct {
	Type       string `json:"type"`              // Error type classification
	Message    string `json:"message"`           // Human-readable error message
	Code       string `json:"code,omitempty"`    // API-specific error code
	Details    string `json:"details,omitempty"` // Additional error context
	StatusCode int    `json:"status_code"`       // HTTP status code
}

// Error implements the error interface for ForwardEmailError.
// It formats the error message with type, code (if available), and message
// to provide clear, actionable error information to users.
func (e *ForwardEmailError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("%s (%s): %s", e.Type, e.Code, e.Message)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the corresponding sentinel error based on HTTP status code.
// This enables error type checking with errors.Is() and allows consumers
// to handle different error categories without checking status codes directly.
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
