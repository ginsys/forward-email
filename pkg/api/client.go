// Package api provides the Forward Email REST API client and data structures.
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/ginsys/forward-email/pkg/auth"
	"github.com/ginsys/forward-email/pkg/errors"
)

// Client represents the Forward Email API client
type Client struct {
	HTTPClient *http.Client
	BaseURL    *url.URL
	Auth       auth.Provider
	Account    *AccountService
	Domains    *DomainService
	Aliases    *AliasService
	Emails     *EmailService
	Logs       *LogService
	Crypto     *CryptoService
	UserAgent  string
}

// ClientOption defines options for configuring the client
type ClientOption func(*Client) error

// NewClient creates a new Forward Email API client
func NewClient(baseURL string, authProvider auth.Provider, opts ...ClientOption) (*Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	client := &Client{
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		BaseURL:   u,
		Auth:      authProvider,
		UserAgent: "forward-email/dev",
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(client); err != nil {
			return nil, err
		}
	}

	// Initialize services
	client.Account = &AccountService{client: client}
	client.Domains = &DomainService{client: client}
	client.Aliases = &AliasService{client: client}
	client.Emails = &EmailService{client: client}
	client.Logs = &LogService{client: client}
	client.Crypto = &CryptoService{client: client}

	return client, nil
}

// ValidateAuth validates the authentication credentials
func (c *Client) ValidateAuth(ctx context.Context) error {
	if c.Auth == nil {
		return fmt.Errorf("no authentication provider configured")
	}

	return c.Auth.Validate(ctx)
}

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) error {
		c.HTTPClient = httpClient
		return nil
	}
}

// WithUserAgent sets a custom user agent
func WithUserAgent(userAgent string) ClientOption {
	return func(c *Client) error {
		c.UserAgent = userAgent
		return nil
	}
}

// Do performs an HTTP request with authentication and error handling.
// The request context is used for cancellation and timeout control.
// If v is provided, the response body will be JSON decoded into it.
// API errors are automatically parsed and returned as typed errors.
func (c *Client) Do(ctx context.Context, req *http.Request, v interface{}) error {
	// Apply authentication using the configured auth provider
	if err := c.Auth.Apply(req); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Set standard headers expected by the Forward Email API
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)

	// Execute request with context for cancellation support
	resp, err := c.HTTPClient.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	// Handle HTTP error status codes (4xx, 5xx)
	if resp.StatusCode >= 400 {
		return c.handleErrorResponse(resp)
	}

	// Decode successful response body if destination provided
	if v != nil {
		return json.NewDecoder(resp.Body).Decode(v)
	}

	return nil
}

// handleErrorResponse parses HTTP error responses and returns typed errors.
// It attempts to parse the JSON error response from the Forward Email API,
// falling back to generic errors if parsing fails. Special handling is
// provided for rate limiting errors which include retry-after information.
func (c *Client) handleErrorResponse(resp *http.Response) error {
	var apiResponse struct {
		Message string `json:"message"`
		Code    string `json:"code,omitempty"`
		Error   string `json:"error,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		// Fallback to generic error if we can't parse the response body
		return errors.NewForwardEmailError(resp.StatusCode, resp.Status, "")
	}

	// Extract error message with fallback priority: message -> error -> status
	message := apiResponse.Message
	if message == "" {
		message = apiResponse.Error
	}
	if message == "" {
		message = resp.Status
	}

	code := apiResponse.Code

	// Handle rate limiting with retry-after header
	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfter := resp.Header.Get("Retry-After")
		return errors.NewRateLimitError(retryAfter)
	}

	return errors.NewForwardEmailError(resp.StatusCode, message, code)
}
