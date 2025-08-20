# API Client Package

The `pkg/api` package provides a comprehensive Go client library for the Forward Email API. This package handles all API interactions, authentication, and data models for Forward Email resources.

## Overview

This package implements a clean, idiomatic Go API client with the following features:

- **HTTP Basic Authentication** with API key management
- **Comprehensive Error Handling** with typed errors and retry logic
- **Complete Resource Coverage** for domains, aliases, and emails
- **Request/Response Validation** with proper data models
- **Configurable HTTP Client** with timeouts and custom transport

## Quick Start

### Basic Usage

```go
import (
    "context"
    "github.com/ginsys/forward-email/pkg/api"
    "github.com/ginsys/forward-email/pkg/auth"
)

// Create authentication provider
authProvider := &auth.AuthProvider{
    // Configure with API key
}

// Create API client
client := api.NewClient(api.ClientConfig{
    BaseURL:      "https://api.forwardemail.net",
    AuthProvider: authProvider,
    Timeout:      30 * time.Second,
})

// Use domain service
domains, err := client.Domains.List(context.Background(), api.DomainListOptions{
    Page:  1,
    Limit: 25,
})
if err != nil {
    log.Fatalf("Failed to list domains: %v", err)
}

fmt.Printf("Found %d domains\n", len(domains.Data))
```

## Client Configuration

### Creating a Client

```go
type ClientConfig struct {
    BaseURL      string              // API base URL
    AuthProvider auth.Provider       // Authentication provider
    HTTPClient   *http.Client       // Custom HTTP client (optional)
    UserAgent    string             // Custom user agent (optional)
    Timeout      time.Duration      // Request timeout
}

client := api.NewClient(api.ClientConfig{
    BaseURL:      "https://api.forwardemail.net",
    AuthProvider: authProvider,
    Timeout:      30 * time.Second,
})
```

### HTTP Client Customization

```go
// Custom HTTP client with proxy
transport := &http.Transport{
    Proxy: http.ProxyURL(proxyURL),
}

httpClient := &http.Client{
    Transport: transport,
    Timeout:   60 * time.Second,
}

client := api.NewClient(api.ClientConfig{
    BaseURL:      "https://api.forwardemail.net",
    AuthProvider: authProvider,
    HTTPClient:   httpClient,
})
```

## Services

### Domain Service

Manage Forward Email domains with complete CRUD operations.

```go
// List domains with filtering
domains, err := client.Domains.List(ctx, api.DomainListOptions{
    Page:     1,
    Limit:    25,
    Verified: &[]bool{true}[0], // Only verified domains
    Search:   "example",
})

// Get single domain
domain, err := client.Domains.Get(ctx, "domain-id")

// Create new domain
newDomain, err := client.Domains.Create(ctx, api.DomainCreateRequest{
    Name: "example.com",
    Plan: "enhanced",
})

// Update domain
updatedDomain, err := client.Domains.Update(ctx, "domain-id", api.DomainUpdateRequest{
    Plan: "enhanced",
})

// Delete domain
err = client.Domains.Delete(ctx, "domain-id")

// Verify DNS configuration
verification, err := client.Domains.Verify(ctx, "domain-id")

// Manage domain members
err = client.Domains.AddMember(ctx, "domain-id", api.DomainMemberRequest{
    Email: "user@example.com",
    Role:  "admin",
})
```

### Alias Service

Manage email aliases with complete lifecycle support.

```go
// List aliases for a domain
aliases, err := client.Aliases.List(ctx, "domain-id", api.AliasListOptions{
    Page:    1,
    Limit:   25,
    Enabled: &[]bool{true}[0],
})

// Get single alias
alias, err := client.Aliases.Get(ctx, "domain-id", "alias-id")

// Create new alias
newAlias, err := client.Aliases.Create(ctx, "domain-id", api.AliasCreateRequest{
    Name:        "info",
    Recipients:  []string{"team@company.com"},
    Description: "General inquiries",
})

// Update alias recipients
err = client.Aliases.UpdateRecipients(ctx, "domain-id", "alias-id", 
    []string{"team@company.com", "backup@company.com"})

// Enable/disable alias
err = client.Aliases.Enable(ctx, "domain-id", "alias-id")
err = client.Aliases.Disable(ctx, "domain-id", "alias-id")

// Generate IMAP password
password, err := client.Aliases.GeneratePassword(ctx, "domain-id", "alias-id")

// Delete alias
err = client.Aliases.Delete(ctx, "domain-id", "alias-id")
```

### Email Service

Send and manage emails with attachment support.

```go
// Send email
emailResponse, err := client.Emails.Send(ctx, api.EmailSendRequest{
    From:    "info@example.com",
    To:      []string{"user@example.com"},
    Subject: "Welcome!",
    Text:    "Welcome to our service!",
    Attachments: []api.EmailAttachment{
        {
            Filename:    "welcome.pdf",
            ContentType: "application/pdf",
            Content:     base64.StdEncoding.EncodeToString(pdfData),
        },
    },
})

// List sent emails
emails, err := client.Emails.List(ctx, api.EmailListOptions{
    Page:  1,
    Limit: 25,
    After: time.Now().AddDate(0, -1, 0), // Last month
})

// Get email details
email, err := client.Emails.Get(ctx, "email-id")

// Delete sent email
err = client.Emails.Delete(ctx, "email-id")

// Check sending quota
quota, err := client.Emails.GetQuota(ctx)
```

## Data Models

### Domain Models

```go
type Domain struct {
    ID                        string              `json:"id"`
    Name                      string              `json:"name"`
    HasAdultContentProtection bool                `json:"has_adult_content_protection"`
    HasExecutableProtection   bool                `json:"has_executable_protection"`
    HasPhishingProtection     bool                `json:"has_phishing_protection"`
    HasVirusProtection        bool                `json:"has_virus_protection"`
    IsGlobal                  bool                `json:"is_global"`
    MaxQuotaPerAlias          int                 `json:"max_quota_per_alias"`
    Plan                      string              `json:"plan"`
    RetentionDays             int                 `json:"retention_days"`
    SmtpPort                  int                 `json:"smtp_port"`
    CreatedAt                 time.Time           `json:"created_at"`
    UpdatedAt                 time.Time           `json:"updated_at"`
    HasMxRecord               bool                `json:"has_mx_record"`
    HasTxtRecord              bool                `json:"has_txt_record"`
    Members                   []DomainMember      `json:"members,omitempty"`
}

type DomainListOptions struct {
    Page     int    `json:"page,omitempty"`
    Limit    int    `json:"limit,omitempty"`
    Search   string `json:"search,omitempty"`
    Verified *bool  `json:"verified,omitempty"`
    Plan     string `json:"plan,omitempty"`
    Sort     string `json:"sort,omitempty"`
    Order    string `json:"order,omitempty"`
}

type DomainCreateRequest struct {
    Name string `json:"name"`
    Plan string `json:"plan,omitempty"`
}
```

### Alias Models

```go
type Alias struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    Recipients  []string  `json:"recipients"`
    IsEnabled   bool      `json:"is_enabled"`
    HasImap     bool      `json:"has_imap"`
    HasPgp      bool      `json:"has_pgp"`
    Labels      []string  `json:"labels"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type AliasCreateRequest struct {
    Name        string   `json:"name"`
    Description string   `json:"description,omitempty"`
    Recipients  []string `json:"recipients"`
    Labels      []string `json:"labels,omitempty"`
}

type AliasListOptions struct {
    Page    int    `json:"page,omitempty"`
    Limit   int    `json:"limit,omitempty"`
    Search  string `json:"search,omitempty"`
    Enabled *bool  `json:"enabled,omitempty"`
}
```

### Email Models

```go
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

type Email struct {
    ID        string            `json:"id"`
    MessageID string            `json:"message_id"`
    From      string            `json:"from"`
    To        []string          `json:"to"`
    Subject   string            `json:"subject"`
    Status    string            `json:"status"`
    CreatedAt time.Time         `json:"created_at"`
    Headers   map[string]string `json:"headers,omitempty"`
}
```

## Error Handling

### Error Types

The API client provides typed errors for different scenarios:

```go
import "github.com/ginsys/forward-email/pkg/errors"

// Check error types
if err != nil {
    switch {
    case errors.IsNotFound(err):
        fmt.Println("Resource not found")
    case errors.IsUnauthorized(err):
        fmt.Println("Authentication failed")
    case errors.IsRateLimit(err):
        fmt.Println("Rate limit exceeded")
    case errors.IsValidation(err):
        fmt.Println("Validation error:", err)
    default:
        fmt.Println("Unexpected error:", err)
    }
}
```

### Error Response Structure

```go
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
```

### Retry Logic

```go
// Configure retry behavior
client := api.NewClient(api.ClientConfig{
    BaseURL:      "https://api.forwardemail.net",
    AuthProvider: authProvider,
    RetryConfig: &api.RetryConfig{
        MaxRetries:    3,
        BackoffFactor: 2,
        MaxBackoff:    30 * time.Second,
    },
})
```

## Authentication Integration

### Auth Provider Interface

The API client expects an authentication provider that implements:

```go
type Provider interface {
    GetAPIKey() (string, error)
    ValidateAPIKey(ctx context.Context, apiKey string) error
    GetProfile() string
}
```

### Example Implementation

```go
type SimpleAuthProvider struct {
    apiKey string
}

func (p *SimpleAuthProvider) GetAPIKey() (string, error) {
    if p.apiKey == "" {
        return "", errors.New("API key not configured")
    }
    return p.apiKey, nil
}

func (p *SimpleAuthProvider) ValidateAPIKey(ctx context.Context, apiKey string) error {
    // Implement validation logic
    return nil
}

func (p *SimpleAuthProvider) GetProfile() string {
    return "default"
}
```

## Testing

### Unit Tests

```go
func TestDomainService_List(t *testing.T) {
    // Create mock server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(api.DomainListResponse{
            Data: []api.Domain{
                {ID: "1", Name: "example.com", Plan: "free"},
            },
            Page:  1,
            Limit: 25,
            Total: 1,
        })
    }))
    defer server.Close()
    
    // Create client
    client := api.NewClient(api.ClientConfig{
        BaseURL:      server.URL,
        AuthProvider: &SimpleAuthProvider{apiKey: "test-key"},
    })
    
    // Test list operation
    result, err := client.Domains.List(context.Background(), api.DomainListOptions{})
    if err != nil {
        t.Fatalf("List failed: %v", err)
    }
    
    if len(result.Data) != 1 {
        t.Errorf("Expected 1 domain, got %d", len(result.Data))
    }
}
```

### Integration Tests

```go
func TestIntegration_DomainOperations(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    
    // Use real API with test credentials
    client := api.NewClient(api.ClientConfig{
        BaseURL: "https://api.forwardemail.net",
        AuthProvider: &auth.AuthProvider{
            // Configure with test API key
        },
    })
    
    ctx := context.Background()
    
    // Test domain creation
    domain, err := client.Domains.Create(ctx, api.DomainCreateRequest{
        Name: "test-" + generateRandomString() + ".com",
        Plan: "free",
    })
    if err != nil {
        t.Fatalf("Create failed: %v", err)
    }
    
    // Cleanup
    defer client.Domains.Delete(ctx, domain.ID)
    
    // Test domain retrieval
    retrieved, err := client.Domains.Get(ctx, domain.ID)
    if err != nil {
        t.Fatalf("Get failed: %v", err)
    }
    
    if retrieved.ID != domain.ID {
        t.Errorf("Expected ID %s, got %s", domain.ID, retrieved.ID)
    }
}
```

## Advanced Usage

### Custom HTTP Transport

```go
// Configure proxy
proxyURL, _ := url.Parse("http://proxy.example.com:8080")
transport := &http.Transport{
    Proxy: http.ProxyURL(proxyURL),
    TLSClientConfig: &tls.Config{
        InsecureSkipVerify: false,
    },
}

client := api.NewClient(api.ClientConfig{
    BaseURL:      "https://api.forwardemail.net",
    AuthProvider: authProvider,
    HTTPClient: &http.Client{
        Transport: transport,
        Timeout:   60 * time.Second,
    },
})
```

### Request Middleware

```go
// Add request logging
type LoggingTransport struct {
    Transport http.RoundTripper
}

func (t *LoggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
    start := time.Now()
    resp, err := t.Transport.RoundTrip(req)
    duration := time.Since(start)
    
    log.Printf("%s %s %d %v", req.Method, req.URL.Path, resp.StatusCode, duration)
    return resp, err
}

client := api.NewClient(api.ClientConfig{
    BaseURL:      "https://api.forwardemail.net",
    AuthProvider: authProvider,
    HTTPClient: &http.Client{
        Transport: &LoggingTransport{
            Transport: http.DefaultTransport,
        },
    },
})
```

### Concurrent Operations

```go
// Process multiple domains concurrently
var wg sync.WaitGroup
domains := []string{"domain1.com", "domain2.com", "domain3.com"}

for _, domainName := range domains {
    wg.Add(1)
    go func(name string) {
        defer wg.Done()
        
        domain, err := client.Domains.Create(ctx, api.DomainCreateRequest{
            Name: name,
            Plan: "free",
        })
        if err != nil {
            log.Printf("Failed to create %s: %v", name, err)
            return
        }
        
        log.Printf("Created domain %s with ID %s", name, domain.ID)
    }(domainName)
}

wg.Wait()
```

## Best Practices

### Error Handling

```go
// Always check for specific error types
result, err := client.Domains.Get(ctx, domainID)
if err != nil {
    if errors.IsNotFound(err) {
        // Handle missing domain
        return nil, fmt.Errorf("domain %s does not exist", domainID)
    }
    if errors.IsUnauthorized(err) {
        // Handle auth issues
        return nil, fmt.Errorf("authentication failed: %w", err)
    }
    // Handle other errors
    return nil, fmt.Errorf("failed to get domain: %w", err)
}
```

### Context Usage

```go
// Use context for cancellation and timeouts
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

result, err := client.Domains.List(ctx, api.DomainListOptions{})
if err != nil {
    if ctx.Err() == context.DeadlineExceeded {
        return fmt.Errorf("request timed out")
    }
    return fmt.Errorf("request failed: %w", err)
}
```

### Resource Management

```go
// Always clean up resources
domain, err := client.Domains.Create(ctx, api.DomainCreateRequest{
    Name: "temporary-domain.com",
})
if err != nil {
    return err
}

// Ensure cleanup even if operation fails
defer func() {
    if err := client.Domains.Delete(ctx, domain.ID); err != nil {
        log.Printf("Failed to cleanup domain %s: %v", domain.ID, err)
    }
}()

// Use the domain...
```

## Package Structure

```
pkg/api/
├── client.go              # HTTP client and configuration
├── client_test.go         # Client tests
├── domain.go              # Domain data models
├── domain_service.go      # Domain service implementation
├── domain_service_test.go # Domain service tests
├── alias.go               # Alias data models
├── alias_service.go       # Alias service implementation
├── alias_service_test.go  # Alias service tests
├── email.go               # Email data models
├── email_service.go       # Email service implementation
├── email_service_test.go  # Email service tests
├── services.go            # Service interfaces
└── README.md              # This documentation
```

## Dependencies

The API package has minimal external dependencies:

- **Standard Library**: net/http, encoding/json, time, context
- **Internal Packages**: pkg/auth, pkg/errors
- **External**: None (pure standard library implementation)

## Performance Considerations

- **Connection Pooling**: HTTP client reuses connections automatically
- **Request Timeouts**: Configurable timeouts prevent hanging requests
- **Memory Efficiency**: Streaming JSON parsing for large responses
- **Concurrent Safety**: Client is safe for concurrent use

For more information on the Forward Email CLI architecture and development, see:
- [Architecture Overview](../../docs/development/architecture.md)
- [API Integration Guide](../../docs/development/api-integration.md)
- [Testing Strategy](../../docs/development/testing.md)