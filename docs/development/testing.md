# Testing Strategy

Comprehensive testing approach for Forward Email CLI with focus on reliability, coverage, and cross-platform compatibility.

## Testing Philosophy

### Core Principles
- **Test-Driven Development**: Write tests before implementation when possible
- **Comprehensive Coverage**: >90% coverage for critical paths, 100% for validators/formatters
- **Cross-Platform Testing**: All features must work on Linux/macOS/Windows
- **Integration Testing**: Real API interactions with proper mocking
- **Performance Testing**: Benchmark CLI operations and API calls

### Quality Standards
- **Test Coverage**: >90% for critical functionality
- **Performance**: <2 seconds for simple operations
- **Error Handling**: Comprehensive error scenarios tested
- **Documentation**: All test scenarios documented

## Current Test Status

### Test Coverage Overview
```
Total Packages: 10
Total Test Cases: 100+
All Tests: PASSING ✅

Package Breakdown:
- pkg/auth: 8 tests (authentication & credential management)
- internal/keyring: 6 tests (OS keyring integration)
- internal/client: 7 tests (API client wrapper)
- internal/cmd: 15+ tests (all CLI commands)
- pkg/api: 17 tests (HTTP client & domain service)
- pkg/config: 12 tests (configuration management)
- pkg/errors: 25+ tests (error handling)
- pkg/output: 15+ tests (output formatting)
```

### Test Execution

For detailed information on test commands and build automation, see [Makefile Guide](makefile-guide.md).

Quick reference:
```bash
make test               # Run tests with race detector
make test-ci            # CI execution with coverage
make pre-commit         # Pre-commit checks
```

## Test Architecture

### Test Organization

```
├── pkg/
│   ├── api/
│   │   ├── client_test.go           # HTTP client tests
│   │   ├── domain_service_test.go   # Domain service tests
│   │   ├── alias_service_test.go    # Alias service tests
│   │   └── testdata/                # Test fixtures
│   ├── auth/
│   │   └── provider_test.go         # Authentication tests
│   ├── config/
│   │   └── config_test.go           # Configuration tests
│   └── output/
│       └── formatter_test.go        # Output formatting tests
├── internal/
│   ├── cmd/
│   │   ├── auth_test.go             # Auth command tests
│   │   ├── domain_test.go           # Domain command tests
│   │   └── profile_test.go          # Profile command tests
│   ├── client/
│   │   └── client_test.go           # Client wrapper tests
│   └── keyring/
│       └── keyring_test.go          # Keyring integration tests
└── testdata/                        # Global test fixtures
```

### Test Categories

1. **Unit Tests**: Individual component testing
2. **Integration Tests**: API service interaction testing
3. **Command Tests**: CLI command functionality testing
4. **End-to-End Tests**: Complete workflow testing
5. **Cross-Platform Tests**: OS-specific functionality testing

## Unit Testing Patterns

### Basic Test Structure

```go
func TestDomainService_List(t *testing.T) {
    tests := []struct {
        name           string
        options        DomainListOptions
        mockResponse   string
        mockStatusCode int
        want           *DomainListResponse
        wantErr        bool
    }{
        {
            name: "successful list",
            options: DomainListOptions{
                Page:  1,
                Limit: 25,
            },
            mockResponse: `{
                "data": [
                    {"id": "1", "name": "example.com", "plan": "free"},
                    {"id": "2", "name": "test.com", "plan": "enhanced"}
                ],
                "page": 1,
                "limit": 25,
                "total": 2
            }`,
            mockStatusCode: 200,
            want: &DomainListResponse{
                Data: []Domain{
                    {ID: "1", Name: "example.com", Plan: "free"},
                    {ID: "2", Name: "test.com", Plan: "enhanced"},
                },
                Page:  1,
                Limit: 25,
                Total: 2,
            },
            wantErr: false,
        },
        // ... more test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Mock Server Setup

Tests use `httptest.Server` for API mocking. For detailed mock server patterns and API integration testing, see [API Integration](api-integration.md).

## Authentication Testing

### Auth Provider Tests

```go
func TestAuthProvider_GetAPIKey(t *testing.T) {
    tests := []struct {
        name        string
        setupEnv    map[string]string
        setupConfig *config.Config
        setupKeyring func(*mockKeyring)
        want        string
        wantErr     bool
    }{
        {
            name: "get from environment variable",
            setupEnv: map[string]string{
                "FORWARDEMAIL_API_KEY": "env-api-key",
            },
            want:    "env-api-key",
            wantErr: false,
        },
        {
            name: "get from keyring",
            setupKeyring: func(k *mockKeyring) {
                k.Set("forward-email", "default", "keyring-api-key")
            },
            want:    "keyring-api-key",
            wantErr: false,
        },
        // ... more test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup environment
            for key, value := range tt.setupEnv {
                os.Setenv(key, value)
                defer os.Unsetenv(key)
            }
            
            // Test implementation
        })
    }
}
```

### Keyring Mock

```go
type mockKeyring struct {
    items map[string]map[string]string
    err   error
}

func (m *mockKeyring) Get(service, user string) (keyring.Item, error) {
    if m.err != nil {
        return keyring.Item{}, m.err
    }
    
    if serviceItems, exists := m.items[service]; exists {
        if data, exists := serviceItems[user]; exists {
            return keyring.Item{
                Key:  user,
                Data: []byte(data),
            }, nil
        }
    }
    
    return keyring.Item{}, keyring.ErrKeyNotFound
}

func (m *mockKeyring) Set(service, user, data string) {
    if m.items == nil {
        m.items = make(map[string]map[string]string)
    }
    if m.items[service] == nil {
        m.items[service] = make(map[string]string)
    }
    m.items[service][user] = data
}
```

## API Service Testing

### HTTP Client Testing

```go
func TestClient_Do(t *testing.T) {
    tests := []struct {
        name           string
        setupAuth      func() auth.Provider
        requestMethod  string
        requestURL     string
        requestBody    interface{}
        mockStatusCode int
        mockResponse   string
        wantErr        bool
        wantAuthHeader string
    }{
        {
            name: "successful GET request",
            setupAuth: func() auth.Provider {
                return &mockAuthProvider{apiKey: "test-api-key"}
            },
            requestMethod:  "GET",
            requestURL:     "/domains",
            mockStatusCode: 200,
            mockResponse:   `{"data": []}`,
            wantErr:        false,
            wantAuthHeader: "Basic " + base64.StdEncoding.EncodeToString([]byte("test-api-key:")),
        },
        // ... more test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Service Integration Tests

```go
func TestDomainService_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }
    
    // Setup test server with realistic responses
    server := createMockServer(map[string]mockResponse{
        "/domains": {
            StatusCode: 200,
            Body: `{
                "data": [
                    {
                        "id": "507f1f77bcf86cd799439011",
                        "name": "example.com",
                        "plan": "free",
                        "verified": true,
                        "created_at": "2024-01-01T00:00:00Z"
                    }
                ],
                "page": 1,
                "limit": 25,
                "total": 1
            }`,
        },
    })
    defer server.Close()
    
    // Create client
    client := &Client{
        baseURL:    server.URL,
        httpClient: server.Client(),
        authProvider: &mockAuthProvider{apiKey: "test-key"},
    }
    
    service := &DomainService{client: client}
    
    // Test list operation
    result, err := service.List(context.Background(), DomainListOptions{})
    if err != nil {
        t.Fatalf("List failed: %v", err)
    }
    
    if len(result.Data) != 1 {
        t.Errorf("Expected 1 domain, got %d", len(result.Data))
    }
    
    domain := result.Data[0]
    if domain.Name != "example.com" {
        t.Errorf("Expected domain name 'example.com', got '%s'", domain.Name)
    }
}
```

## Command Testing

### CLI Command Tests

```go
func TestAuthLoginCommand(t *testing.T) {
    tests := []struct {
        name           string
        args           []string
        setupKeyring   func(*mockKeyring)
        inputResponses []string
        wantErr        bool
        wantOutput     string
    }{
        {
            name: "successful login",
            args: []string{"auth", "login"},
            setupKeyring: func(k *mockKeyring) {
                // Keyring starts empty
            },
            inputResponses: []string{"test-api-key"},
            wantErr:        false,
            wantOutput:     "Login successful",
        },
        // ... more test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup mock keyring
            mockKeyring := &mockKeyring{}
            if tt.setupKeyring != nil {
                tt.setupKeyring(mockKeyring)
            }
            
            // Setup command with mocked dependencies
            cmd := NewRootCommand()
            cmd.SetArgs(tt.args)
            
            // Capture output
            var output bytes.Buffer
            cmd.SetOut(&output)
            cmd.SetErr(&output)
            
            // Execute command
            err := cmd.Execute()
            
            if (err != nil) != tt.wantErr {
                t.Errorf("Command error = %v, wantErr %v", err, tt.wantErr)
            }
            
            if tt.wantOutput != "" && !strings.Contains(output.String(), tt.wantOutput) {
                t.Errorf("Expected output to contain '%s', got '%s'", tt.wantOutput, output.String())
            }
        })
    }
}
```

### Command Mock Setup

```go
// Test helper for command setup
func setupTestCommand() *cobra.Command {
    // Create root command with test configuration
    cmd := &cobra.Command{
        Use:   "forward-email",
        Short: "Forward Email CLI",
    }
    
    // Add subcommands
    cmd.AddCommand(NewAuthCommand())
    cmd.AddCommand(NewDomainCommand())
    cmd.AddCommand(NewAliasCommand())
    
    return cmd
}

// Mock input for interactive commands
func mockInput(responses []string) io.Reader {
    input := strings.Join(responses, "\n") + "\n"
    return strings.NewReader(input)
}
```

## Output Formatting Tests

### Formatter Tests

```go
func TestFormatDomains(t *testing.T) {
    domains := []Domain{
        {
            ID:       "1",
            Name:     "example.com",
            Plan:     "free",
            Verified: true,
        },
        {
            ID:       "2", 
            Name:     "test.com",
            Plan:     "enhanced",
            Verified: false,
        },
    }
    
    tests := []struct {
        name   string
        format output.Format
        want   string
    }{
        {
            name:   "table format",
            format: output.FormatTable,
            want:   "example.com",
        },
        {
            name:   "json format",
            format: output.FormatJSON,
            want:   `"name":"example.com"`,
        },
        {
            name:   "yaml format", 
            format: output.FormatYAML,
            want:   "name: example.com",
        },
        {
            name:   "csv format",
            format: output.FormatCSV,
            want:   "example.com,free,true",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := output.FormatDomains(domains, tt.format)
            if err != nil {
                t.Fatalf("FormatDomains failed: %v", err)
            }
            
            resultStr := ""
            switch v := result.(type) {
            case string:
                resultStr = v
            case []byte:
                resultStr = string(v)
            default:
                resultBytes, _ := json.Marshal(v)
                resultStr = string(resultBytes)
            }
            
            if !strings.Contains(resultStr, tt.want) {
                t.Errorf("Expected output to contain '%s', got '%s'", tt.want, resultStr)
            }
        })
    }
}
```

## Error Handling Tests

### Error Response Tests

```go
func TestAPIErrorHandling(t *testing.T) {
    tests := []struct {
        name           string
        statusCode     int
        responseBody   string
        wantErrorType  error
        wantErrorMsg   string
    }{
        {
            name:       "not found error",
            statusCode: 404,
            responseBody: `{
                "code": "NOT_FOUND",
                "message": "Domain not found"
            }`,
            wantErrorType: &errors.NotFoundError{},
            wantErrorMsg:  "Domain not found",
        },
        {
            name:       "unauthorized error",
            statusCode: 401,
            responseBody: `{
                "code": "UNAUTHORIZED", 
                "message": "Invalid API key"
            }`,
            wantErrorType: &errors.UnauthorizedError{},
            wantErrorMsg:  "Invalid API key",
        },
        // ... more test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test error parsing and type checking
        })
    }
}
```

## Configuration Testing

### Config File Tests

```go
func TestConfig_Load(t *testing.T) {
    tests := []struct {
        name       string
        configYAML string
        want       *Config
        wantErr    bool
    }{
        {
            name: "valid config",
            configYAML: `
current_profile: production
profiles:
  production:
    base_url: "https://api.forwardemail.net"
    timeout: "30s"
    output: "json"
`,
            want: &Config{
                CurrentProfile: "production",
                Profiles: map[string]Profile{
                    "production": {
                        BaseURL: "https://api.forwardemail.net",
                        Timeout: "30s",
                        Output:  "json",
                    },
                },
            },
            wantErr: false,
        },
        // ... more test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Create temporary config file
            tmpfile, err := os.CreateTemp("", "config_test_*.yaml")
            if err != nil {
                t.Fatalf("Failed to create temp file: %v", err)
            }
            defer os.Remove(tmpfile.Name())
            
            // Write test config
            if _, err := tmpfile.WriteString(tt.configYAML); err != nil {
                t.Fatalf("Failed to write config: %v", err)
            }
            tmpfile.Close()
            
            // Test config loading
            config, err := LoadConfig(tmpfile.Name())
            if (err != nil) != tt.wantErr {
                t.Errorf("LoadConfig error = %v, wantErr %v", err, tt.wantErr)
            }
            
            if !reflect.DeepEqual(config, tt.want) {
                t.Errorf("LoadConfig = %+v, want %+v", config, tt.want)
            }
        })
    }
}
```

## Cross-Platform Testing

### GitHub Actions Matrix

```yaml
# .github/workflows/test.yml
name: Test
on: [push, pull_request]

jobs:
  test:
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
        go-version: [1.21, 1.22]
    
    runs-on: ${{ matrix.os }}
    
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v3
      with:
        go-version: ${{ matrix.go-version }}
    
    - name: Run tests
      run: go test -race -coverprofile=coverage.out ./...
    
    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out
```

### OS-Specific Tests

```go
//go:build windows
// +build windows

func TestWindowsKeyring(t *testing.T) {
    // Windows-specific keyring tests
}

//go:build darwin
// +build darwin

func TestMacOSKeychain(t *testing.T) {
    // macOS-specific keychain tests
}

//go:build linux
// +build linux

func TestLinuxSecretService(t *testing.T) {
    // Linux-specific secret service tests
}
```

## Performance Testing

### Benchmark Tests

```go
func BenchmarkDomainList(b *testing.B) {
    // Setup mock server
    server := createMockServer(map[string]mockResponse{
        "/domains": {StatusCode: 200, Body: `{"data": []}`},
    })
    defer server.Close()
    
    client := &Client{
        baseURL:    server.URL,
        httpClient: server.Client(),
    }
    service := &DomainService{client: client}
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := service.List(context.Background(), DomainListOptions{})
        if err != nil {
            b.Fatalf("List failed: %v", err)
        }
    }
}

func BenchmarkOutputFormatting(b *testing.B) {
    domains := make([]Domain, 100)
    for i := range domains {
        domains[i] = Domain{
            ID:   fmt.Sprintf("domain-%d", i),
            Name: fmt.Sprintf("example%d.com", i),
            Plan: "free",
        }
    }
    
    b.Run("table", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            output.FormatDomains(domains, output.FormatTable)
        }
    })
    
    b.Run("json", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            output.FormatDomains(domains, output.FormatJSON)
        }
    })
}
```

## Test Data Management

### Test Fixtures

```go
// testdata/domains.json
var testDomains = []Domain{
    {
        ID:                        "507f1f77bcf86cd799439011",
        Name:                      "example.com",
        HasAdultContentProtection: true,
        HasExecutableProtection:   true,
        HasPhishingProtection:     true,
        HasVirusProtection:        true,
        Plan:                      "enhanced",
        CreatedAt:                 time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
    },
}

// Load test data
func loadTestDomains(t *testing.T) []Domain {
    data, err := os.ReadFile("testdata/domains.json")
    if err != nil {
        t.Fatalf("Failed to load test data: %v", err)
    }
    
    var domains []Domain
    if err := json.Unmarshal(data, &domains); err != nil {
        t.Fatalf("Failed to unmarshal test data: %v", err)
    }
    
    return domains
}
```

## Test Utilities

### Common Test Helpers

```go
// Test assertion helpers
func assertNoError(t *testing.T, err error) {
    t.Helper()
    if err != nil {
        t.Fatalf("Expected no error, got: %v", err)
    }
}

func assertEqual(t *testing.T, got, want interface{}) {
    t.Helper()
    if !reflect.DeepEqual(got, want) {
        t.Errorf("Got %+v, want %+v", got, want)
    }
}

func assertContains(t *testing.T, haystack, needle string) {
    t.Helper()
    if !strings.Contains(haystack, needle) {
        t.Errorf("Expected '%s' to contain '%s'", haystack, needle)
    }
}

// Setup and teardown helpers
func setupTest(t *testing.T) (context.Context, func()) {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    
    cleanup := func() {
        cancel()
    }
    
    return ctx, cleanup
}
```

## Test Execution Guidelines

### Running Tests

#### Makefile Commands (Aligned with CI)
```bash
# Standard workflows
make test               # Default: tests with race detector
make test-ci            # Exactly what CI runs (with coverage)
make test-quick         # Fast feedback loop

# Specific test types
make test-unit          # Unit tests only
make test-bench         # Benchmarks
make test-pkg PKG=api   # Test specific package

# Quality checks
make fmt-check          # Verify formatting
make lint-ci            # CI-compatible linting
make pre-commit         # Full pre-commit checks
```

#### Development Setup
```bash
# One-time setup
make dev-setup          # Install tools and git hooks

# Daily development
make check              # Quick checks before commit
make check-all          # Full validation
```

### Test Coverage Goals

- **Critical Functionality**: >90% coverage
- **Error Handling**: 100% coverage
- **Validators/Formatters**: 100% coverage
- **Integration Points**: >80% coverage

### Continuous Integration

#### CI Pipeline (GitHub Actions)
```bash
# CI test pipeline (matches local commands)
make deps           # Download dependencies
make fmt-check      # Verify formatting
make test-ci        # Tests with coverage
make lint-ci        # Linting
make build-all      # Multi-platform builds
```

#### Local CI Simulation
```bash
# Run exactly what CI runs
make deps
make fmt-check
make test-ci
make lint-ci
make build-all

# Or use the convenience command
make check-all      # Runs fmt-check, lint, test-ci
```

#### Pre-commit Integration
```bash
# Install pre-commit hooks (one-time)
make install-hooks

# Manual pre-commit check
make pre-commit     # Quick: fmt-check, lint-fast, test-quick

# Full pre-commit validation
make pre-commit-full  # Same as CI: fmt-check, lint, test-ci
```

For more information on development practices, see:
- [Architecture Overview](architecture.md)
- [API Integration](api-integration.md)
- [Contributing Guide](contributing.md)
- [Makefile Guide](makefile-guide.md)

---

*Last Updated: 2026-01-18*
