package errors

import (
	"errors"
	"net/http"
	"testing"
)

func TestForwardEmailError_Error(t *testing.T) {
	tests := []struct {
		name     string
		error    *ForwardEmailError
		expected string
	}{
		{
			name: "error with code",
			error: &ForwardEmailError{
				Type:    "ValidationError",
				Message: "Invalid email format",
				Code:    "INVALID_EMAIL",
			},
			expected: "ValidationError (INVALID_EMAIL): Invalid email format",
		},
		{
			name: "error without code",
			error: &ForwardEmailError{
				Type:    "NotFound",
				Message: "Domain not found",
			},
			expected: "NotFound: Domain not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.error.Error()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestForwardEmailError_Unwrap(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		expected   error
	}{
		{"NotFound", http.StatusNotFound, ErrNotFound},
		{"Unauthorized", http.StatusUnauthorized, ErrUnauthorized},
		{"Forbidden", http.StatusForbidden, ErrForbidden},
		{"BadRequest", http.StatusBadRequest, ErrBadRequest},
		{"Conflict", http.StatusConflict, ErrConflict},
		{"RateLimit", http.StatusTooManyRequests, ErrRateLimit},
		{"ServiceUnavailable", http.StatusServiceUnavailable, ErrServiceUnavailable},
		{"ServerError", http.StatusInternalServerError, ErrServerError},
		{"4xx validation error", http.StatusUnprocessableEntity, ErrValidation},
		{"5xx server error", http.StatusBadGateway, ErrServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &ForwardEmailError{StatusCode: tt.statusCode}
			result := err.Unwrap()
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestNewForwardEmailError(t *testing.T) {
	statusCode := http.StatusBadRequest
	message := "Invalid request"
	code := "INVALID_REQ"

	err := NewForwardEmailError(statusCode, message, code)

	if err.StatusCode != statusCode {
		t.Errorf("Expected status code %d, got %d", statusCode, err.StatusCode)
	}
	if err.Message != message {
		t.Errorf("Expected message %q, got %q", message, err.Message)
	}
	if err.Code != code {
		t.Errorf("Expected code %q, got %q", code, err.Code)
	}
	if err.Type != "BadRequest" {
		t.Errorf("Expected type 'BadRequest', got %q", err.Type)
	}
}

func TestNewNotFoundError(t *testing.T) {
	resource := "Domain"
	err := NewNotFoundError(resource)

	if err.Type != "NotFound" {
		t.Errorf("Expected type 'NotFound', got %q", err.Type)
	}
	if err.Message != "Domain not found" {
		t.Errorf("Expected message 'Domain not found', got %q", err.Message)
	}
	if err.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, err.StatusCode)
	}
}

func TestNewValidationError(t *testing.T) {
	message := "Email is required"
	err := NewValidationError(message)

	if err.Type != "ValidationError" {
		t.Errorf("Expected type 'ValidationError', got %q", err.Type)
	}
	if err.Message != message {
		t.Errorf("Expected message %q, got %q", message, err.Message)
	}
	if err.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, err.StatusCode)
	}
}

// testErrorCreator is a helper function to reduce code duplication in error creation tests
func testErrorCreator(
	t *testing.T,
	name string,
	creator func(string) *ForwardEmailError,
	expectedType string,
	expectedStatus int,
	customMessage, customExpected, defaultExpected string,
) {
	t.Helper()

	tests := []struct {
		name     string
		message  string
		expected string
	}{
		{
			name:     "with custom message",
			message:  customMessage,
			expected: customExpected,
		},
		{
			name:     "with empty message",
			message:  "",
			expected: defaultExpected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := creator(tt.message)

			if err.Type != expectedType {
				t.Errorf("Expected type %q, got %q", expectedType, err.Type)
			}
			if err.Message != tt.expected {
				t.Errorf("Expected message %q, got %q", tt.expected, err.Message)
			}
			if err.StatusCode != expectedStatus {
				t.Errorf("Expected status code %d, got %d", expectedStatus, err.StatusCode)
			}
		})
	}
}

func TestNewUnauthorizedError(t *testing.T) {
	testErrorCreator(t, "NewUnauthorizedError", NewUnauthorizedError, "Unauthorized", http.StatusUnauthorized,
		"Invalid API key", "Invalid API key", "Authentication required")
}

func TestNewForbiddenError(t *testing.T) {
	testErrorCreator(t, "NewForbiddenError", NewForbiddenError, "Forbidden", http.StatusForbidden,
		"Insufficient permissions", "Insufficient permissions", "Access forbidden")
}

func TestNewRateLimitError(t *testing.T) {
	tests := []struct {
		name       string
		retryAfter string
		expected   string
	}{
		{
			name:       "with retry after",
			retryAfter: "60",
			expected:   "Rate limit exceeded. Retry after 60",
		},
		{
			name:       "without retry after",
			retryAfter: "",
			expected:   "Rate limit exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewRateLimitError(tt.retryAfter)

			if err.Type != "RateLimit" {
				t.Errorf("Expected type 'RateLimit', got %q", err.Type)
			}
			if err.Message != tt.expected {
				t.Errorf("Expected message %q, got %q", tt.expected, err.Message)
			}
			if err.StatusCode != http.StatusTooManyRequests {
				t.Errorf("Expected status code %d, got %d", http.StatusTooManyRequests, err.StatusCode)
			}
			if err.Details != tt.retryAfter {
				t.Errorf("Expected details %q, got %q", tt.retryAfter, err.Details)
			}
		})
	}
}

func TestNewServerError(t *testing.T) {
	testErrorCreator(t, "NewServerError", NewServerError, "ServerError", http.StatusInternalServerError,
		"Database connection failed", "Database connection failed", "Internal server error")
}

func TestNewServiceUnavailableError(t *testing.T) {
	testErrorCreator(t, "NewServiceUnavailableError", NewServiceUnavailableError,
		"ServiceUnavailable", http.StatusServiceUnavailable,
		"Maintenance in progress", "Maintenance in progress", "Service temporarily unavailable")
}

func Test_getErrorType(t *testing.T) {
	tests := []struct {
		statusCode int
		expected   string
	}{
		{http.StatusBadRequest, "BadRequest"},
		{http.StatusUnauthorized, "Unauthorized"},
		{http.StatusForbidden, "Forbidden"},
		{http.StatusNotFound, "NotFound"},
		{http.StatusConflict, "Conflict"},
		{http.StatusTooManyRequests, "RateLimit"},
		{http.StatusInternalServerError, "ServerError"},
		{http.StatusServiceUnavailable, "ServiceUnavailable"},
		{http.StatusUnprocessableEntity, "ClientError"}, // 4xx fallback
		{http.StatusBadGateway, "ServerError"},          // 5xx fallback
		{300, "ServerError"},                            // default fallback
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := getErrorType(tt.statusCode)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestErrorTypeCheckers(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		checker  func(error) bool
		expected bool
	}{
		{"IsNotFound with NotFound error", NewNotFoundError("test"), IsNotFound, true},
		{"IsNotFound with other error", NewUnauthorizedError("test"), IsNotFound, false},
		{"IsUnauthorized with Unauthorized error", NewUnauthorizedError("test"), IsUnauthorized, true},
		{"IsUnauthorized with other error", NewNotFoundError("test"), IsUnauthorized, false},
		{"IsForbidden with Forbidden error", NewForbiddenError("test"), IsForbidden, true},
		{"IsForbidden with other error", NewNotFoundError("test"), IsForbidden, false},
		{"IsValidation with 422 error", NewForwardEmailError(http.StatusUnprocessableEntity, "test", ""), IsValidation, true},
		{"IsValidation with other error", NewNotFoundError("test"), IsValidation, false},
		{
			"NewValidationError actually creates BadRequest error",
			NewValidationError("test"),
			func(err error) bool { return errors.Is(err, ErrBadRequest) },
			true,
		},
		{"IsRateLimit with RateLimit error", NewRateLimitError("60"), IsRateLimit, true},
		{"IsRateLimit with other error", NewNotFoundError("test"), IsRateLimit, false},
		{"IsServerError with ServerError", NewServerError("test"), IsServerError, true},
		{"IsServerError with other error", NewNotFoundError("test"), IsServerError, false},
		{"IsServiceUnavailable with ServiceUnavailable error",
			NewServiceUnavailableError("test"), IsServiceUnavailable, true},
		{"IsServiceUnavailable with other error", NewNotFoundError("test"), IsServiceUnavailable, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.checker(tt.err)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"RateLimit error is retryable", NewRateLimitError("60"), true},
		{"ServiceUnavailable error is retryable", NewServiceUnavailableError("test"), true},
		{"ServerError is retryable", NewServerError("test"), true},
		{"NotFound error is not retryable", NewNotFoundError("test"), false},
		{"Unauthorized error is not retryable", NewUnauthorizedError("test"), false},
		{"Validation error is not retryable", NewValidationError("test"), false},
		{"Standard error is not retryable", errors.New("standard error"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRetryable(tt.err)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGetStatusCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected int
	}{
		{"ForwardEmailError", NewNotFoundError("test"), http.StatusNotFound},
		{"Standard error", errors.New("standard error"), 0},
		{"Nil error", nil, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetStatusCode(tt.err)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestGetErrorCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			"ForwardEmailError with code",
			NewForwardEmailError(http.StatusBadRequest, "test", "TEST_CODE"),
			"TEST_CODE",
		},
		{
			"ForwardEmailError without code",
			NewNotFoundError("test"),
			"",
		},
		{
			"Standard error",
			errors.New("standard error"),
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetErrorCode(tt.err)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestGetErrorDetails(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			"RateLimit error with details",
			NewRateLimitError("60"),
			"60",
		},
		{
			"RateLimit error without details",
			NewRateLimitError(""),
			"",
		},
		{
			"Other ForwardEmailError",
			NewNotFoundError("test"),
			"",
		},
		{
			"Standard error",
			errors.New("standard error"),
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetErrorDetails(tt.err)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestErrorIs(t *testing.T) {
	notFoundErr := NewNotFoundError("test")

	// Test that errors.Is works correctly with our custom error type
	if !errors.Is(notFoundErr, ErrNotFound) {
		t.Error("Expected NotFound error to match ErrNotFound")
	}

	if errors.Is(notFoundErr, ErrUnauthorized) {
		t.Error("Expected NotFound error not to match ErrUnauthorized")
	}
}

func TestErrorAs(t *testing.T) {
	originalErr := NewNotFoundError("test")

	var feErr *ForwardEmailError
	if !errors.As(originalErr, &feErr) {
		t.Fatal("Expected error to be assignable to ForwardEmailError")
	}

	if feErr.Type != "NotFound" {
		t.Errorf("Expected type 'NotFound', got %q", feErr.Type)
	}

	if feErr.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, feErr.StatusCode)
	}
}
