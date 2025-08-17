# Forward Email CLI - Project Memory

## 🎯 Project Overview

**Forward Email CLI** - A comprehensive command-line interface for managing Forward Email accounts and resources through their public REST API. This project represents a **first-mover advantage** as Forward Email currently has zero official CLI tools.

**Current Phase**: Phase 1.2 Domain Operations → **COMPLETED** ✅  
**Next Phase**: Phase 1.3 Alias & Email Operations → **PLANNED** ⏳

## 🏗️ Architecture

### Core Components
```
cmd/forward-email/           # Main CLI entry point
internal/cmd/               # CLI command implementations
├── root.go                 # Root command setup
├── auth.go                 # Authentication commands
├── profile.go              # Profile management commands
├── domain.go               # Domain CRUD operations
└── debug.go                # Debug and troubleshooting utilities
internal/keyring/           # OS keyring integration
├── keyring.go              # Keyring wrapper
└── keyring_test.go         # Keyring tests
internal/client/            # Client wrapper for API initialization
├── client.go               # Enhanced client factory
└── client_test.go          # Client wrapper tests
pkg/api/                    # API client library
├── client.go               # HTTP client with auth
├── client_test.go          # Client tests
├── domain.go               # Domain data models
├── domain_service.go       # Domain service implementation
└── domain_service_test.go  # Domain service tests
pkg/auth/                   # Authentication system
├── provider.go             # Auth provider implementation
└── provider_test.go        # Auth tests
pkg/config/                 # Configuration management
├── config.go               # Profile-based config
└── config_test.go          # Configuration tests
pkg/errors/                 # Centralized error handling
├── errors.go               # Forward Email API errors
└── errors_test.go          # Error handling tests
pkg/output/                 # Output formatting system
├── formatter.go            # Multi-format output (table/JSON/YAML/CSV)
├── formatter_test.go       # Formatter tests
├── domain.go               # Domain-specific output formatting
└── domain_test.go          # Domain output tests
```

## 🔐 Authentication System (COMPLETED)

### Features Implemented
- **Multi-source credential hierarchy**: Environment variables → OS keyring → config file
- **Profile-specific authentication**: Support for multiple environments (dev/staging/prod)
- **Secure storage**: OS keyring integration via 99designs/keyring library
- **Cross-platform support**: Windows Credential Manager, macOS Keychain, Linux Secret Service
- **Comprehensive testing**: 24 test cases with 100% auth coverage

### Authentication Commands
```bash
forward-email auth login     # Interactive API key setup with secure input
forward-email auth verify    # Validate current credentials against API
forward-email auth status    # Show auth status across all profiles
forward-email auth logout    # Clear stored credentials (single or all profiles)
```

### Profile Management Commands
```bash
forward-email profile list                    # List all configured profiles
forward-email profile show [profile]          # Show current or specific profile details
forward-email profile create <name>           # Create new profile configuration
forward-email profile switch <name>           # Switch to different profile
forward-email profile delete <name> --force   # Delete profile configuration
```

### Debug & Troubleshooting Commands
```bash
forward-email debug keys [profile]    # Show keyring information for debugging
forward-email debug auth [profile]    # Debug full authentication flow
forward-email debug api [profile]     # Test API call with current authentication
```

### Environment Variables
```bash
FORWARDEMAIL_API_KEY              # Generic API key
FORWARDEMAIL_<PROFILE>_API_KEY    # Profile-specific API key (e.g., FORWARDEMAIL_PROD_API_KEY)
```

## 🏢 Domain Management System (COMPLETED)

### Features Implemented
- **Complete CRUD Operations**: List, get, create, update, delete domains
- **DNS Record Management**: MX, TXT, DMARC, SPF, DKIM verification status
- **Domain Verification**: Verification record generation and validation
- **Quota & Statistics**: Usage monitoring and limit tracking
- **Member Management**: Add/remove domain members with role-based access
- **Flexible Output**: Table, JSON, YAML, CSV formatting with pagination and filtering

### Domain Commands
```bash
forward-email domain list                     # List all domains with filtering and pagination
forward-email domain get <domain-name-or-id>  # Get detailed domain information
forward-email domain create <name>            # Create new domain
forward-email domain update <domain-or-id>    # Update domain settings
forward-email domain delete <domain-or-id>    # Delete domain
forward-email domain verify <domain-or-id>    # Verify domain DNS configuration
forward-email domain dns <domain-or-id>       # Show required DNS records
forward-email domain quota <domain-or-id>     # Show domain quota and usage
forward-email domain stats <domain-or-id>     # Show domain statistics
forward-email domain members <domain-or-id>   # Manage domain members
```

### Domain List Filtering & Pagination
```bash
# Filtering options
--verified true|false        # Filter by verification status
--plan free|enhanced|team    # Filter by subscription plan
--search <query>             # Search domain names
--sort name|created|updated  # Sort criteria
--order asc|desc            # Sort order

# Pagination options
--page <number>              # Page number (default: 1)
--limit <number>             # Items per page (default: 25)

# Output formatting
--output table|json|yaml|csv # Output format (default: table)
```

### Output Formatting System
- **Table Format**: Clean, readable tables with proper column alignment
- **JSON Format**: Standard JSON output for API integration
- **YAML Format**: Human-readable YAML for configuration files
- **CSV Format**: Comma-separated values for spreadsheet import
- **Pagination Support**: Consistent pagination across all list commands
- **Stable Ordering**: Deterministic sort orders for reproducible output

## 🚨 Error Handling System (COMPLETED)

### Features Implemented
- **Centralized Error Management**: Unified error types and handling across all API operations
- **User-Friendly Messages**: Clear, actionable error messages with suggestions
- **HTTP Status Code Mapping**: Proper mapping of API responses to semantic error types
- **Retry Logic**: Intelligent retry recommendations for transient errors
- **Error Type Classification**: Structured error types (NotFound, Unauthorized, RateLimit, etc.)

### Error Types & Handling
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

### Error Context & Details
- **Status Code**: HTTP status code for programmatic handling
- **Error Code**: Forward Email specific error codes
- **User Message**: Human-readable error descriptions
- **Retry Information**: Rate limit retry-after headers
- **Suggestion Engine**: Actionable next steps for error resolution

## 🛠️ Dependencies

### Core Dependencies
```go
github.com/99designs/keyring v1.2.2     // Secure credential storage
github.com/spf13/cobra v1.8.0           // CLI framework
github.com/spf13/viper v1.18.2          // Configuration management
golang.org/x/term v0.3.0                // Secure password input
github.com/olekukonko/tablewriter v0.0.5 // Table formatting
gopkg.in/yaml.v3 v3.0.1                 // YAML output formatting
```

### Build & Test
```bash
go build -o bin/forward-email ./cmd/forward-email  # Build CLI
go test ./...                                       # Run all tests
make test                                          # Alternative test runner
```

## 🔧 Configuration

### Profile Structure
```yaml
current_profile: default
profiles:
  default:
    base_url: "https://api.forwardemail.net"
    timeout: "30s"
    output: "table"
    api_key: ""  # Optional, prefer keyring storage
  production:
    base_url: "https://api.forwardemail.net"
    timeout: "30s"
    output: "json"
```

### Config Locations
- **Linux**: `~/.config/forwardemail/config.yaml`
- **macOS**: `~/.config/forwardemail/config.yaml`
- **Windows**: `%APPDATA%\forwardemail\config.yaml`

## 🧪 Testing Strategy

### Current Test Coverage
- **pkg/auth**: 8 test functions covering auth provider, validation, credential hierarchy
- **internal/keyring**: 6 test functions covering keyring operations and profile management
- **internal/client**: 7 test functions covering client wrapper and API initialization
- **internal/cmd**: 15+ test functions covering all CLI commands (auth, profile, domain, debug)
- **pkg/api**: 17 test functions covering client, domain service, and CRUD operations
- **pkg/config**: 12 test functions covering configuration management and profiles
- **pkg/errors**: 25+ test functions covering all error types and handling scenarios
- **pkg/output**: 15+ test functions covering all output formats and domain formatting

### Test Execution
```bash
go test ./...                    # Run all tests (100+ test cases)
go test ./pkg/... -v             # Run package tests with verbose output
go test -race ./...              # Run tests with race condition detection
go test -cover ./...             # Run tests with coverage reporting
```

### Test Statistics (Latest)
```
Total Packages: 10
Total Test Cases: 100+
Coverage: Comprehensive across all components
All Tests: PASSING ✅
```

## 📋 Current Implementation Status

### ✅ Completed (Phase 1.1 - Core Infrastructure)
- [x] **Authentication System**: Complete multi-source auth provider
- [x] **Keyring Integration**: Secure OS credential storage
- [x] **CLI Framework**: Cobra-based command structure
- [x] **Configuration Management**: Profile-based config with Viper
- [x] **HTTP Client**: Enhanced API client with auth validation
- [x] **Auth Commands**: Full authentication workflow (login/verify/status/logout)
- [x] **Profile Management**: Complete profile CRUD operations
- [x] **Cross-Platform Build**: Successful builds on Linux/macOS/Windows

### ✅ Completed (Phase 1.2 - Domain Operations)
- [x] **Domain Data Models**: Complete Go structs for Forward Email domain API
- [x] **Domain Service Implementation**: Full DomainService with all CRUD operations
- [x] **Domain Commands**: Complete domain lifecycle (list/get/create/update/delete/verify)
- [x] **DNS Management**: Domain verification and DNS record management
- [x] **Member Management**: Add/remove domain members with role-based access
- [x] **Error Handling System**: Centralized error management with user-friendly messages
- [x] **Output Formatting**: Multi-format output (table/JSON/YAML/CSV) with pagination
- [x] **Debug Utilities**: Comprehensive troubleshooting tools for auth and API issues
- [x] **Comprehensive Testing**: 100+ test cases covering all components
- [x] **API Integration**: Full Forward Email API integration with proper error handling

### ⏳ Planned (Phase 1.3 - Alias & Email Operations)
- [ ] **Alias Management**: Complete alias lifecycle (list/create/update/delete)
- [ ] **Email Operations**: Send, list, delete with quota management
- [ ] **Bulk Operations**: Batch processing for multiple resources
- [ ] **Interactive Wizards**: Setup and configuration wizards
- [ ] **Shell Completion**: Bash/Zsh/Fish completion scripts

### ⏳ Planned (Phase 1.4+ - Enhanced Features)
- [ ] **CI/CD Pipeline**: Automated testing and release process
- [ ] **Webhook Management**: Configure and test webhook endpoints
- [ ] **Log Management**: Download and analyze email logs
- [ ] **Template System**: Email templates and bulk sending
- [ ] **Monitoring Integration**: Health checks and status monitoring

## 🚀 Forward Email API Integration

### Authentication Method
- **Type**: HTTP Basic Authentication
- **Format**: `Authorization: Basic <base64(api_key + ":")>`
- **Endpoint**: `https://api.forwardemail.net/v1/`

### API Coverage Planned
- **Account**: Profile management, quota monitoring
- **Domains**: CRUD operations, DNS/SMTP verification
- **Aliases**: Complete lifecycle with recipients and settings
- **Emails**: Send operations with templates and tracking
- **Logs**: Download with rate limit respect (10/day)

## 🔍 Development Notes

### Next Immediate Steps (Phase 1.3)
1. **Alias Data Models**: Define API response structures for aliases
2. **Alias Service**: Implement complete alias CRUD operations
3. **Email Operations**: Send, list, delete functionality
4. **Bulk Operations**: Batch processing capabilities
5. **Interactive Wizards**: User-friendly setup and configuration

### Technical Decisions Made
- **Keyring Library**: Chose 99designs/keyring for mature cross-platform support
- **Auth Hierarchy**: Environment → keyring → config for maximum flexibility
- **Profile Architecture**: Multi-environment support for dev/staging/prod workflows
- **Testing Approach**: Comprehensive unit tests with mock implementations
- **Security Model**: Never log credentials, secure file permissions, OS keyring priority
- **Output System**: Flexible multi-format output with table formatting as default
- **Error Architecture**: Centralized error handling with HTTP status mapping and user-friendly messages
- **Domain Architecture**: Complete service layer with proper separation of concerns

### Known Limitations
- **API Documentation**: Limited Forward Email API docs, reverse-engineering from Auth.js examples
- **Rate Limiting**: Must implement respectful API usage patterns (partially implemented in error handling)
- **Offline Mode**: No offline capabilities planned for MVP
- **Alias Operations**: Not yet implemented (planned for Phase 1.3)
- **Email Operations**: Send/receive functionality not yet implemented

## 📊 Metrics & Quality

### Test Results (Latest)
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

### Build Status
```
Platform: Linux/macOS/Windows ✅
Binary Size: ~20MB (estimated)
Dependencies: 6 direct, 20+ transitive
Go Version: 1.21+
Test Coverage: Comprehensive across all components
```

### Current Capabilities
- ✅ **Full Authentication System** with multi-source credential hierarchy
- ✅ **Complete Domain Management** with CRUD operations and DNS verification
- ✅ **Profile Management** for multi-environment support
- ✅ **Flexible Output Formatting** (table/JSON/YAML/CSV) with pagination
- ✅ **Robust Error Handling** with user-friendly messages and retry logic
- ✅ **Debug Utilities** for troubleshooting authentication and API issues
- ✅ **Comprehensive Testing** with 100+ test cases across all components

This documentation reflects the current state as of **Phase 1.2 completion** with comprehensive domain management system fully implemented and ready for alias/email operations development.