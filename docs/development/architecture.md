# Forward Email CLI Architecture

## Overview

Forward Email CLI is a Go-based command-line interface that provides comprehensive management of Forward Email accounts and resources through their public REST API. The architecture follows modular design principles with centralized authentication, consistent patterns, and enterprise-grade features.

## Design Principles

### Core Principles
- **Clean Separation**: SDK (pkg/api) → CLI commands (internal/cmd) → User interface
- **Security First**: OS keyring integration, credential redaction, secure defaults
- **Developer Experience**: Shell completion, interactive wizards, comprehensive help
- **Enterprise Ready**: Multi-profile management, audit logging, CI/CD integration
- **Minimal Dependencies**: Prefer standard library when possible
- **Consistent UX**: Idempotent operations, safe-by-default flags, human and machine-friendly output

### Strategic Positioning

Forward Email currently has **zero official CLI tools** despite a comprehensive API with 20+ endpoints. This represents a significant competitive advantage opportunity, especially given:
- Cost-effectiveness: $3/month with full API access
- Developer-aligned values and open-source ecosystem
- Growing demand for CLI automation tools

## Architecture Overview

### Directory Structure

```
forward-email/
├── cmd/forward-email/          # Main CLI entry point
├── internal/                   # Internal packages
│   ├── client/                 # Centralized API client creation
│   ├── cmd/                    # Command implementations
│   ├── keyring/                # OS keyring integration
│   ├── telemetry/              # Usage analytics (planned)
│   ├── update/                 # Version checking (planned)
│   └── version/                # Version information
├── pkg/                        # Public packages
│   ├── api/                    # API service layer
│   ├── auth/                   # Authentication provider
│   ├── config/                 # Configuration management
│   ├── errors/                 # Centralized error handling
│   └── output/                 # Output formatting
├── docs/                       # User documentation
├── docs/development/           # Developer documentation
└── .claude/                    # Claude Code specific files
```

### Component Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   CLI Commands  │────│  API Services   │────│ HTTP Transport  │
│  (internal/cmd) │    │   (pkg/api)     │    │   (net/http)    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ Output Format   │    │ Authentication  │    │ Configuration   │
│  (pkg/output)   │    │   (pkg/auth)    │    │  (pkg/config)   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ Error Handling  │    │    Keyring      │    │ Profile Mgmt    │
│  (pkg/errors)   │    │(internal/keyring)│    │ (multi-env)     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Core Systems

### 1. Centralized Authentication

**Problem**: Initially, each command duplicated authentication logic, creating maintenance overhead and potential inconsistencies.

**Solution**: All API client creation is centralized in `internal/client/client.go`.

**Authentication Flow**:
1. **Profile Resolution**: 
   - Check `--profile` flag
   - Fall back to config `current_profile`
   - Final fallback to "default"

2. **Credential Loading**:
   - Environment variables (highest priority)
   - OS keyring (recommended)
   - Configuration file (fallback)
   - Interactive prompt (last resort)

3. **Client Creation**:
   - Initialize keyring (graceful degradation if unavailable)
   - Create auth provider with config, keyring, and profile
   - Create API client with base URL and auth provider

**Usage Pattern**:
```go
import "github.com/ginsys/forward-email/internal/client"

func runSomeCommand(cmd *cobra.Command, args []string) error {
    apiClient, err := client.NewAPIClient()
    if err != nil {
        return err
    }
    
    // Use apiClient for API calls
    response, err := apiClient.SomeService.SomeMethod(ctx, params)
    if err != nil {
        return fmt.Errorf("operation failed: %w", err)
    }
    
    return nil
}
```

### 2. Multi-Profile Management

The CLI supports multiple profiles for different accounts/environments:

- **Current Profile**: Stored in config file (`~/.config/forwardemail/config.yaml`)
- **Profile Storage**: API keys stored securely in OS keyring, settings in config file
- **Profile Selection**: Via `--profile` flag or config file `current_profile` setting
- **Fallback Chain**: Flag → config current_profile → "default"

**Profile Configuration**:
```yaml
current_profile: production
profiles:
  development:
    base_url: "https://api.forwardemail.net"
    timeout: "30s"
    output: "table"
  production:
    base_url: "https://api.forwardemail.net"
    timeout: "60s"
    output: "json"
```

### 3. Output Formatting System

Consistent output formatting across all commands supporting multiple formats:

- **Supported Formats**: table (default), json, yaml, csv
- **Table Formatting**: Uses `github.com/olekukonko/tablewriter`
- **Type System**: `output.TableData` struct for tabular data
- **Format Selection**: Via `--output` flag

**Usage Pattern**:
```go
import "github.com/ginsys/forward-email/pkg/output"

return formatOutput(data, outputFormat, func(format output.Format) (interface{}, error) {
    if format == output.FormatTable || format == output.FormatCSV {
        return output.FormatSomeData(data, format)
    }
    return data, nil
})
```

### 4. Error Handling System

Centralized error management with user-friendly messages:

**Error Types**:
```go
// Core error types with proper HTTP status mapping
ErrNotFound           // 404 - Resource not found
ErrUnauthorized       // 401 - Authentication required
ErrForbidden          // 403 - Access denied
ErrValidation         // 400/422 - Input validation failed
ErrRateLimit          // 429 - Rate limit exceeded (with retry-after)
ErrServerError        // 500 - Internal server error
ErrServiceUnavailable // 503 - Service temporarily unavailable
ErrConflict           // 409 - Resource conflict
```

**Usage Pattern**:
```go
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}
```

## API Integration

### Forward Email API

**Authentication Method**:
- **Type**: HTTP Basic Authentication
- **Format**: `Authorization: Basic <base64(api_key + ":")>`
- **Endpoint**: `https://api.forwardemail.net/v1/`

**Service Architecture**:
- **DomainService**: Complete CRUD operations, DNS verification, member management
- **AliasService**: Full lifecycle management, recipients, IMAP/PGP settings
- **EmailService**: Send operations, attachment support, quota management

### Error Handling Strategy

Structured error handling with actionable suggestions:

```go
type CLIError struct {
    Code        string                 // Structured error code
    Message     string                 // Human-readable message
    Suggestion  string                 // Actionable fix suggestion
    Context     map[string]interface{} // Operation context
    Retryable   bool                   // Can operation be retried
}
```

## Security Model

### Authentication Architecture
- **Primary Auth**: HTTP Basic with API key as username, empty password
- **Session Management**: Credential caching to reduce API calls
- **Credential Validation**: Pre-flight auth verification before operations
- **Multi-source Priority**: Environment variables → OS keyring → config file

### Security Features
- **Secure Storage**: OS keyring integration (Keychain, Credential Manager, Secret Service)
- **Credential Redaction**: Never log or display API keys
- **File Permissions**: Config files created with 0600 permissions
- **Profile Isolation**: Separate credentials for different environments

### OS Keyring Integration

API keys are stored securely in the OS keyring:
- **Service**: `forward-email`
- **Account**: `{profile}` (e.g., "production", "development")

**Supported Systems**:
- **macOS**: Keychain Services
- **Windows**: Windows Credential Manager
- **Linux**: Secret Service (gnome-keyring, KWallet)

## Performance & Scalability

### Caching Strategy
- **Response Caching**: 5-minute TTL for domain/alias lists (planned)
- **Authentication Caching**: Session-based credential caching
- **Configuration Caching**: In-memory profile data

### Bulk Processing
- **Client-side Batching**: Parallel execution with configurable limits
- **Progress Tracking**: Real-time progress indicators for long operations
- **Transaction Support**: Rollback capabilities for failed operations

### Performance Targets
- **Command Execution**: <2 seconds for simple operations
- **API Response Time**: <500ms for standard operations
- **Concurrent Processing**: Default concurrency of 5 operations

## Implementation Patterns

### Adding New Commands

When adding commands that need API access:

1. **Import the centralized client**:
   ```go
   import "github.com/ginsys/forward-email/internal/client"
   ```

2. **Use the standard pattern**:
   ```go
   func runYourCommand(cmd *cobra.Command, args []string) error {
       apiClient, err := client.NewAPIClient()
       if err != nil {
           return err
       }
       
       // Your API calls here
       
       return nil
   }
   ```

3. **Add output formatting**:
   ```go
   return formatOutput(data, outputFormat, func(format output.Format) (interface{}, error) {
       if format == output.FormatTable || format == output.FormatCSV {
           return output.FormatYourData(data, format)
       }
       return data, nil
   })
   ```

### Service Layer Pattern

All API operations follow a consistent service layer pattern:

```go
type SomeService struct {
    client *Client
}

func (s *SomeService) List(ctx context.Context, options ListOptions) (*ListResponse, error) {
    // Implementation with error handling, validation, etc.
}

func (s *SomeService) Get(ctx context.Context, id string) (*Resource, error) {
    // Implementation
}

func (s *SomeService) Create(ctx context.Context, req CreateRequest) (*Resource, error) {
    // Implementation
}
```

## Testing Strategy

### Test Architecture
- **Unit Tests**: >90% coverage for critical paths
- **Integration Tests**: Mock API responses for service layer
- **End-to-End Tests**: Real workflow testing with test accounts
- **Cross-Platform Tests**: GitHub Actions matrix for Windows/macOS/Linux

### Testing Patterns
- Mock the `client.NewAPIClient()` function for unit tests
- Test profile switching and authentication flows
- Validate output formatting for all supported formats
- Test error handling and graceful degradation

### Current Test Coverage
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

## Quality Assurance

### Code Quality Standards
- **Linting**: golangci-lint with comprehensive rule set
- **Formatting**: gofmt and goimports for consistent style
- **Documentation**: 100% of exported functions documented
- **Error Handling**: Consistent error wrapping and context

### Build System
- **Cross-Platform**: Builds for Linux/macOS/Windows (amd64/arm64)
- **Dependencies**: Minimal external dependencies, prefer stdlib
- **Binary Size**: ~20MB target size
- **Performance**: Sub-2-second command execution for simple operations

## Future Enhancements

### Phase 2 Features (Planned)
- **Interactive Setup**: `forward-email init` wizard
- **Bulk Operations**: CSV import/export with progress tracking, domain alias synchronization
- **Template System**: YAML-based email templates
- **Shell Completion**: Auto-generated completions for bash/zsh/fish
- **Plugin Architecture**: Community extensibility framework

### Performance Improvements
- **Response Caching**: Intelligent caching with TTL
- **Concurrent Processing**: Parallel operations for bulk tasks
- **Streaming Output**: Real-time progress for long operations
- **Request Optimization**: Efficient API usage patterns

### Enterprise Features
- **Audit Logging**: Comprehensive operation history
- **Compliance**: GDPR/SOX support
- **Advanced Security**: MFA support, credential rotation
- **CI/CD Integration**: GitHub Actions, Docker containers

## Troubleshooting & Debugging

### Debug Utilities
- `forward-email debug auth`: Authentication flow debugging
- `forward-email debug api`: API connectivity testing
- `forward-email debug keys`: Keyring access verification

### Logging & Observability
- **Structured Logging**: JSON formatted logs for automation
- **Trace IDs**: Request correlation for debugging
- **Performance Metrics**: Command execution timing
- **Error Context**: Rich error information with suggestions

## Contributing Guidelines

### Development Setup
1. Go 1.21+ required
2. Run `go mod download` for dependencies
3. Use `make test` for running tests
4. Follow existing patterns for new features

### Architecture Decisions
- All architectural changes should be documented
- Breaking changes require major version bump
- New features should include comprehensive tests
- Security changes require thorough review

For more detailed information, see:
- [API Integration Guide](api-integration.md)
- [Testing Strategy](testing.md)
- [Contributing Guide](contributing.md)