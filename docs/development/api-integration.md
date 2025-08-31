# Forward Email API Integration

Complete guide to Forward Email API integration patterns, authentication, and service implementation.

## API Overview

### Forward Email API Details

**Base URL**: `https://api.forwardemail.net/v1/`  
**Authentication**: HTTP Basic Authentication  
**API Documentation**: Limited official docs, reverse-engineered from Auth.js examples  
**Rate Limiting**: Respectful usage required, 10 requests/day limit for logs  

### Current API Coverage Status

| Resource | Status | Operations | Notes |
|----------|--------|------------|--------|
| **Domains** | ✅ Implemented | CRUD, DNS/SMTP verification, members | Complete |
| **Aliases** | ✅ Implemented | Complete lifecycle, recipients, settings | Complete |
| **Emails** | ✅ Implemented | Send operations, attachment support | Complete |
| **Account** | ⏳ Planned | Profile management, quota monitoring | Phase 2 |
| **Logs** | ⏳ Planned | Download with rate limit respect (10/day) | Phase 2 |

## Authentication System

### HTTP Basic Authentication

Forward Email API uses HTTP Basic authentication with the API key as the username:

```go
// Authentication format
Authorization: Basic base64(api_key + ":")
```

### Implementation Pattern

```go
// pkg/api/client.go
func (c *Client) setAuth(req *http.Request) {
    if c.authProvider != nil {
        apiKey, err := c.authProvider.GetAPIKey()
        if err == nil && apiKey != "" {
            // HTTP Basic: API key as username, empty password
            req.SetBasicAuth(apiKey, "")
        }
    }
}
```

### Authentication Provider Interface

```go
// pkg/auth/provider.go
type Provider interface {
    GetAPIKey() (string, error)
    ValidateAPIKey(ctx context.Context, apiKey string) error
    GetProfile() string
}

type AuthProvider struct {
    config   *config.Config
    keyring  keyring.Keyring
    profile  string
}
```

### Credential Hierarchy

1. **Environment Variables** (highest priority)
   ```bash
   FORWARDEMAIL_API_KEY="your-api-key"
   FORWARDEMAIL_PRODUCTION_API_KEY="prod-key"
   ```

2. **OS Keyring** (recommended)
   - macOS: Keychain Services
   - Windows: Credential Manager
   - Linux: Secret Service

3. **Configuration File** (fallback)
   ```yaml
   profiles:
     production:
       api_key: "stored-in-config"  # Not recommended
   ```

4. **Interactive Prompt** (last resort)

## Service Layer Architecture

### Base Service Pattern

All API services follow a consistent pattern:

```go
// Service interface
type Service interface {
    List(ctx context.Context, options ListOptions) (*ListResponse, error)
    Get(ctx context.Context, id string) (*Resource, error)
    Create(ctx context.Context, req CreateRequest) (*Resource, error)
    Update(ctx context.Context, id string, req UpdateRequest) (*Resource, error)
    Delete(ctx context.Context, id string) error
}

// Service implementation
type DomainService struct {
    client *Client
}

func (s *DomainService) List(ctx context.Context, options DomainListOptions) (*DomainListResponse, error) {
    // Build request with pagination, filtering
    req, err := s.buildListRequest(options)
    if err != nil {
        return nil, fmt.Errorf("failed to build request: %w", err)
    }
    
    // Execute request with error handling
    resp, err := s.client.Do(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("API request failed: %w", err)
    }
    
    // Parse and validate response
    var result DomainListResponse
    if err := s.client.parseResponse(resp, &result); err != nil {
        return nil, fmt.Errorf("failed to parse response: %w", err)
    }
    
    return &result, nil
}
```

### HTTP Client Implementation

```go
// pkg/api/client.go
type Client struct {
    baseURL      string
    httpClient   *http.Client
    authProvider auth.Provider
    userAgent    string
}

func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
    // Set authentication
    c.setAuth(req)
    
    // Set standard headers
    req.Header.Set("User-Agent", c.userAgent)
    req.Header.Set("Accept", "application/json")
    if req.Method != http.MethodGet && req.Header.Get("Content-Type") == "" {
        req.Header.Set("Content-Type", "application/json")
    }
    
    // Execute with context
    resp, err := c.httpClient.Do(req.WithContext(ctx))
    if err != nil {
        return nil, fmt.Errorf("HTTP request failed: %w", err)
    }
    
    // Handle API errors
    if err := c.checkResponse(resp); err != nil {
        resp.Body.Close()
        return nil, err
    }
    
    return resp, nil
}
```

## Domain Service Implementation

### Domain Operations

```go
// pkg/api/domain_service.go
type DomainService struct {
    client *Client
}

// List domains with filtering and pagination
func (s *DomainService) List(ctx context.Context, options DomainListOptions) (*DomainListResponse, error) {
    url := "/domains"
    if len(options.toQuery()) > 0 {
        url += "?" + options.toQuery().Encode()
    }
    
    req, err := s.client.NewRequest(http.MethodGet, url, nil)
    if err != nil {
        return nil, err
    }
    
    resp, err := s.client.Do(ctx, req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var result DomainListResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }
    
    return &result, nil
}

// Get single domain
func (s *DomainService) Get(ctx context.Context, domainID string) (*Domain, error) {
    req, err := s.client.NewRequest(http.MethodGet, "/domains/"+domainID, nil)
    if err != nil {
        return nil, err
    }
    
    resp, err := s.client.Do(ctx, req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var domain Domain
    if err := json.NewDecoder(resp.Body).Decode(&domain); err != nil {
        return nil, fmt.Errorf("failed to decode domain: %w", err)
    }
    
    return &domain, nil
}

// Create new domain
func (s *DomainService) Create(ctx context.Context, req DomainCreateRequest) (*Domain, error) {
    httpReq, err := s.client.NewRequest(http.MethodPost, "/domains", req)
    if err != nil {
        return nil, err
    }
    
    resp, err := s.client.Do(ctx, httpReq)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var domain Domain
    if err := json.NewDecoder(resp.Body).Decode(&domain); err != nil {
        return nil, fmt.Errorf("failed to decode created domain: %w", err)
    }
    
    return &domain, nil
}
```

### Domain Data Models

```go
// pkg/api/domain.go
type Domain struct {
    ID                  string    `json:"id"`
    Name                string    `json:"name"`
    HasAdultContentProtection bool `json:"has_adult_content_protection"`
    HasExecutableProtection   bool `json:"has_executable_protection"`
    HasPhishingProtection     bool `json:"has_phishing_protection"`
    HasVirusProtection        bool `json:"has_virus_protection"`
    IsGlobal                  bool `json:"is_global"`
    MaxQuotaPerAlias          int  `json:"max_quota_per_alias"`
    Plan                      string `json:"plan"`
    RetentionDays             int   `json:"retention_days"`
    SmtpPort                  int   `json:"smtp_port"`
    CreatedAt                 time.Time `json:"created_at"`
    UpdatedAt                 time.Time `json:"updated_at"`
    
    // Verification status
    HasMxRecord  bool `json:"has_mx_record"`
    HasTxtRecord bool `json:"has_txt_record"`
    
    // Members
    Members []DomainMember `json:"members,omitempty"`
}

type DomainMember struct {
    User DomainUser `json:"user"`
    Group string    `json:"group"`
}

type DomainUser struct {
    ID    string `json:"id"`
    Email string `json:"email"`
}

type DomainListOptions struct {
    Page      int    `json:"page,omitempty"`
    Limit     int    `json:"limit,omitempty"`
    Search    string `json:"search,omitempty"`
    Verified  *bool  `json:"verified,omitempty"`
    Plan      string `json:"plan,omitempty"`
    Sort      string `json:"sort,omitempty"`
    Order     string `json:"order,omitempty"`
}
```

## Alias Service Implementation

### Alias Operations

```go
// pkg/api/alias_service.go
type AliasService struct {
    client *Client
}

func (s *AliasService) List(ctx context.Context, domainID string, options AliasListOptions) (*AliasListResponse, error) {
    url := fmt.Sprintf("/domains/%s/aliases", domainID)
    if len(options.toQuery()) > 0 {
        url += "?" + options.toQuery().Encode()
    }
    
    req, err := s.client.NewRequest(http.MethodGet, url, nil)
    if err != nil {
        return nil, err
    }
    
    resp, err := s.client.Do(ctx, req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var result AliasListResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }
    
    return &result, nil
}
```

### Alias Data Models

```go
// pkg/api/alias.go
type Alias struct {
    ID          string   `json:"id"`
    Name        string   `json:"name"`
    Description string   `json:"description"`
    Recipients  []string `json:"recipients"`
    IsEnabled   bool     `json:"is_enabled"`
    HasImap     bool     `json:"has_imap"`
    HasPgp      bool     `json:"has_pgp"`
    Labels      []string `json:"labels"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type AliasCreateRequest struct {
    Name        string   `json:"name"`
    Description string   `json:"description,omitempty"`
    Recipients  []string `json:"recipients"`
    Labels      []string `json:"labels,omitempty"`
}
```

## Email Service Implementation

### Email Operations

```go
// pkg/api/email_service.go
type EmailService struct {
    client *Client
}

func (s *EmailService) Send(ctx context.Context, req EmailSendRequest) (*EmailSendResponse, error) {
    httpReq, err := s.client.NewRequest(http.MethodPost, "/emails", req)
    if err != nil {
        return nil, err
    }
    
    resp, err := s.client.Do(ctx, httpReq)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var result EmailSendResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }
    
    return &result, nil
}
```

### Email Data Models

```go
// pkg/api/email.go
type EmailSendRequest struct {
    From        string            `json:"from"`
    To          []string          `json:"to"`
    Cc          []string          `json:"cc,omitempty"`
    Bcc         []string          `json:"bcc,omitempty"`
    Subject     string            `json:"subject"`
    Text        string            `json:"text,omitempty"`
    Html        string            `json:"html,omitempty"`
    Attachments []EmailAttachment `json:"attachments,omitempty"`
    Headers     map[string]string `json:"headers,omitempty"`
}

type EmailAttachment struct {
    Filename    string `json:"filename"`
    ContentType string `json:"content_type"`
    Content     string `json:"content"` // base64 encoded
}

type EmailSendResponse struct {
    ID        string    `json:"id"`
    MessageID string    `json:"message_id"`
    Status    string    `json:"status"`
    CreatedAt time.Time `json:"created_at"`
}
```

## Error Handling

### API Error Response

```go
// pkg/errors/errors.go
type APIError struct {
    StatusCode int    `json:"status_code"`
    Code       string `json:"code"`
    Message    string `json:"message"`
    Details    string `json:"details,omitempty"`
}

func (e *APIError) Error() string {
    if e.Details != "" {
        return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Details)
    }
    return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Error type constants
const (
    ErrCodeNotFound     = "NOT_FOUND"
    ErrCodeUnauthorized = "UNAUTHORIZED"
    ErrCodeForbidden    = "FORBIDDEN"
    ErrCodeValidation   = "VALIDATION_ERROR"
    ErrCodeRateLimit    = "RATE_LIMIT_EXCEEDED"
    ErrCodeServerError  = "INTERNAL_SERVER_ERROR"
)
```

### Error Response Handling

```go
func (c *Client) checkResponse(resp *http.Response) error {
    if resp.StatusCode >= 200 && resp.StatusCode < 300 {
        return nil
    }
    
    var apiErr APIError
    if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
        return &APIError{
            StatusCode: resp.StatusCode,
            Code:       "UNKNOWN_ERROR",
            Message:    fmt.Sprintf("HTTP %d", resp.StatusCode),
        }
    }
    
    apiErr.StatusCode = resp.StatusCode
    return &apiErr
}
```

## Request/Response Patterns

### Standard Request Building

```go
func (c *Client) NewRequest(method, url string, body interface{}) (*http.Request, error) {
    fullURL := c.baseURL + url
    
    var buf io.Reader
    if body != nil {
        jsonBody, err := json.Marshal(body)
        if err != nil {
            return nil, fmt.Errorf("failed to marshal request body: %w", err)
        }
        buf = bytes.NewBuffer(jsonBody)
    }
    
    req, err := http.NewRequest(method, fullURL, buf)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }
    
    return req, nil
}
```

### Pagination Support

```go
type ListOptions struct {
    Page  int `json:"page,omitempty"`
    Limit int `json:"limit,omitempty"`
}

func (o ListOptions) toQuery() url.Values {
    v := url.Values{}
    if o.Page > 0 {
        v.Set("page", strconv.Itoa(o.Page))
    }
    if o.Limit > 0 {
        v.Set("limit", strconv.Itoa(o.Limit))
    }
    return v
}

type ListResponse struct {
    Data       interface{} `json:"data"`
    Page       int         `json:"page"`
    Limit      int         `json:"limit"`
    Total      int         `json:"total"`
    TotalPages int         `json:"total_pages"`
}
```

## Testing API Integration

### Mock API Server

```go
// Test setup
func setupTestServer() *httptest.Server {
    mux := http.NewServeMux()
    
    mux.HandleFunc("/domains", func(w http.ResponseWriter, r *http.Request) {
        // Mock domain list response
        domains := []Domain{
            {ID: "1", Name: "example.com", Plan: "free"},
            {ID: "2", Name: "test.com", Plan: "enhanced"},
        }
        json.NewEncoder(w).Encode(domains)
    })
    
    return httptest.NewServer(mux)
}
```

### Integration Tests

```go
func TestDomainService_List(t *testing.T) {
    server := setupTestServer()
    defer server.Close()
    
    client := &Client{
        baseURL:    server.URL,
        httpClient: server.Client(),
    }
    
    service := &DomainService{client: client}
    
    domains, err := service.List(context.Background(), DomainListOptions{})
    if err != nil {
        t.Fatalf("List failed: %v", err)
    }
    
    if len(domains.Data) != 2 {
        t.Errorf("Expected 2 domains, got %d", len(domains.Data))
    }
}
```

## Performance Considerations

### Rate Limiting

```go
// Implement rate limiting for API calls
type RateLimiter struct {
    limiter *rate.Limiter
}

func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
    // Wait for rate limiter
    if err := c.rateLimiter.Wait(ctx); err != nil {
        return nil, fmt.Errorf("rate limit wait failed: %w", err)
    }
    
    // Execute request
    return c.httpClient.Do(req.WithContext(ctx))
}
```

### Request Timeouts

```go
func NewClient(config ClientConfig) *Client {
    return &Client{
        httpClient: &http.Client{
            Timeout: config.Timeout, // Default: 30s
            Transport: &http.Transport{
                DialContext: (&net.Dialer{
                    Timeout: 10 * time.Second,
                }).DialContext,
                TLSHandshakeTimeout: 10 * time.Second,
            },
        },
    }
}
```

## API Evolution & Versioning

### Forward Compatibility

```go
// Use json:",omitempty" for optional fields
// Ignore unknown fields in responses
type Domain struct {
    ID   string `json:"id"`
    Name string `json:"name"`
    // New fields added gracefully
    NewField string `json:"new_field,omitempty"`
}
```

### Version Detection

```go
// API version detection
func (c *Client) detectAPIVersion(ctx context.Context) (string, error) {
    req, err := c.NewRequest(http.MethodGet, "/version", nil)
    if err != nil {
        return "", err
    }
    
    resp, err := c.Do(ctx, req)
    if err != nil {
        return "v1", nil // Default fallback
    }
    defer resp.Body.Close()
    
    var version struct {
        Version string `json:"version"`
    }
    json.NewDecoder(resp.Body).Decode(&version)
    return version.Version, nil
}
```

## Security Best Practices

### Request Security

```go
func (c *Client) setSecurityHeaders(req *http.Request) {
    // Prevent credential leakage
    req.Header.Set("Cache-Control", "no-store")
    req.Header.Set("Pragma", "no-cache")
    
    // Set secure user agent
    req.Header.Set("User-Agent", fmt.Sprintf("forward-email-cli/%s", version.Version))
}
```

### Credential Protection

```go
func (c *Client) sanitizeForLogging(req *http.Request) *http.Request {
    // Clone request for logging
    clone := req.Clone(req.Context())
    
    // Remove authorization header
    clone.Header.Del("Authorization")
    
    return clone
}
```

## Future API Enhancements

### Planned Features
- **Account Management**: Profile operations, quota monitoring
- **Log Download**: Respect 10/day limit with intelligent caching
- **Webhook Management**: Configure and test webhook endpoints
- **Real-time Events**: WebSocket/SSE for real-time updates

### API Client Evolution
- **Response Caching**: Intelligent caching with TTL
- **Retry Logic**: Exponential backoff with jitter
- **Circuit Breaker**: Prevent cascading failures
- **Metrics Collection**: Performance and usage metrics

For more information on testing and contributing, see:
- [Testing Strategy](testing.md)
- [Architecture Overview](architecture.md)
- [Contributing Guide](contributing.md)

---

Docs navigation (Dev): [Prev: Architecture](architecture.md) | [Next: Testing Strategy](testing.md) | [Back: Dev Index](README.md)
