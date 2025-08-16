package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// Client represents the Forward Email API client
type Client struct {
	HTTPClient *http.Client
	BaseURL    *url.URL
	Auth       AuthProvider
	UserAgent  string

	// Services
	Account *AccountService
	Domains *DomainService
	Aliases *AliasService
	Emails  *EmailService
	Logs    *LogService
	Crypto  *CryptoService
}

// AuthProvider defines the interface for authentication
type AuthProvider interface {
	Apply(req *http.Request) error
}

// BasicAuth implements HTTP Basic Authentication
type BasicAuth struct {
	Username string
	Password string
}

// Apply sets the Authorization header for Basic Auth
func (b BasicAuth) Apply(req *http.Request) error {
	req.SetBasicAuth(b.Username, b.Password)
	return nil
}

// ClientOption defines options for configuring the client
type ClientOption func(*Client) error

// NewClient creates a new Forward Email API client
func NewClient(baseURL string, auth AuthProvider, opts ...ClientOption) (*Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	client := &Client{
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		BaseURL:   u,
		Auth:      auth,
		UserAgent: "forwardemail-cli/dev",
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

// Do performs an HTTP request
func (c *Client) Do(ctx context.Context, req *http.Request, v interface{}) error {
	// Apply authentication
	if err := c.Auth.Apply(req); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)

	// Execute request
	resp, err := c.HTTPClient.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Handle errors
	if resp.StatusCode >= 400 {
		return c.handleErrorResponse(resp)
	}

	// Decode response
	if v != nil {
		return json.NewDecoder(resp.Body).Decode(v)
	}

	return nil
}

// handleErrorResponse parses and returns API errors
func (c *Client) handleErrorResponse(resp *http.Response) error {
	var apiErr APIError
	if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	apiErr.StatusCode = resp.StatusCode
	return &apiErr
}

// APIError represents an API error response
type APIError struct {
	StatusCode int    `json:"-"`
	Message    string `json:"message"`
	Code       string `json:"code,omitempty"`
}

func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("API error %d (%s): %s", e.StatusCode, e.Code, e.Message)
	}
	return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Message)
}
