# forwardemail-cli --- Enhanced Architecture Document

*Status: Updated v0.2 - Incorporating API Research & Competitive Analysis*

## 1) Purpose & Goals

`forwardemail-cli` is a Go-based command-line interface to manage Forward Email accounts and resources through their public REST API. The tool targets:

- **Operators / DevOps**: automate domain and alias management, download logs, verify DNS/SMTP, send or inspect outbound emails.
- **Developers**: scriptable, predictable CLI with stable output formats for CI/CD, GitOps, and provisioning flows.
- **Enterprise Users**: multi-profile management, audit logging, and compliance features.

**Design principles**

- Minimal dependencies; prefer stdlib.
- Clear separation between **SDK (api client)** and **CLI commands**.
- Consistent UX: idempotent operations where possible, safe-by-default flags, human and machine-friendly output.
- Extensibility for upcoming API surface and community plugins.
- **Developer-first experience** with comprehensive tooling and automation.

**Strategic positioning**

Forward Email currently has **zero official CLI tools** despite a comprehensive API with 20+ endpoints. This represents a significant competitive advantage opportunity, especially given the service's cost-effectiveness ($3/month with full API access) and open-source alignment with developer values.

---

## 2) Enhanced Scope (v1)

### Core Operations
- **Accounts**: get/update profile, quota monitoring, API key rotation.
- **Domains**: CRUD, verify DNS/SMTP, catch‑all passwords, invites, members, settings (protections, quotas, webhooks, retention).
- **Aliases**: CRUD, recipients (emails/FQDN/IP/webhook URLs), IMAP+PGP flags, quotas, vacation responder, password generation.
- **Emails** (outbound): list/get/delete, send via structured fields or RFC822 `--raw`, check daily limit, template support.
- **Logs**: request deliverability logs (csv.gz by email delivery); respect 10 req/day limit with intelligent caching.
- **Encrypt**: helper to encrypt plaintext TXT strings for DNS.

### Enhanced Features
- **Interactive Setup**: `forward-email init` wizard for first-time configuration.
- **Bulk Operations**: Batch processing with progress tracking and rollback capabilities.
- **Template System**: YAML-based email templates with variable substitution.
- **Real-time Monitoring**: Log streaming with `--follow` flag and health checks.
- **CI/CD Integration**: GitHub Actions, Docker containers, and configuration validation.
- **Multi-profile Management**: Development/staging/production environment support.

### Developer Experience
- **Shell Completion**: Auto-generated completions for bash/zsh/fish.
- **Interactive Mode**: Guided workflows for complex operations.
- **Dry-run Support**: Preview operations before execution.
- **Audit Logging**: Local operation history for debugging and compliance.
- **Plugin Architecture**: Community extensibility framework.

**Out of scope (v1)**

- Interactive TUI; IMAP/CardDAV/CalDAV/Messages/Folders (planned in roadmap).
- Account sign-up flow.
- Visual configuration interface (potential v2 feature).

---

## 3) Enhanced Non‑Functional Requirements

### Performance & Scalability
- **Portability**: Linux/macOS/Windows (amd64/arm64).
- **Performance**: async pagination with intelligent caching; streaming writers for large payloads; concurrent processing with configurable limits.
- **Caching Strategy**: 5-minute TTL for domain/alias lists; session-based auth caching.
- **Bulk Processing**: Client-side batching with parallel execution (default concurrency: 5).

### Reliability & Error Handling
- **Reliability**: exponential backoff with jitter; context timeouts; retries on idempotent operations.
- **Structured Error Handling**: Mapped API errors to actionable CLI messages with suggested fixes.
- **Transaction Support**: Rollback capabilities for failed bulk operations.
- **Health Monitoring**: API connectivity checks and quota tracking.

### Security & Compliance
- **Security**: never print secrets; on-disk secrets 0600; mandatory OS keyring integration.
- **Credential Rotation**: Built-in API key management and rotation commands.
- **Audit Trail**: Local operation logging with tamper-evident timestamps.
- **Multi-factor Auth**: Support for future MFA API endpoints.

### Observability & Debugging
- **Observability**: structured logs (stderr), trace-friendly request IDs, `--verbose` and `--debug`.
- **HTTP Debugging**: `--debug-http` for request/response logging.
- **Telemetry**: Anonymous usage metrics (opt-in) for product improvement.
- **Performance Metrics**: Command execution time tracking and reporting.

### Output & Integration
- **Deterministic output**: stable key order for JSON, predictable tables/CSV.
- **Multiple Formats**: table (default), JSON, YAML, CSV, and custom Go templates.
- **Filter & Sort**: Client-side filtering and sorting with SQL-like syntax.
- **Streaming Output**: Progress indicators and real-time updates.

---

## 4) Enhanced Security Model

### Authentication Architecture
- **Primary Auth**: HTTP Basic with **API key as username**, empty password (service-wide).
- **Session Management**: 1-hour token caching to reduce API calls.
- **Credential Validation**: Pre-flight auth verification before operations.
- **Future-ready**: Support for alias-scoped authentication (Messages/Contacts/Calendars).

### Multi-Profile Support
```yaml
# ~/.config/forwardemail/config.yaml
profiles:
  production:
    api_key: "prod_key_from_keyring"
    base_url: "https://api.forwardemail.net"
  staging:
    api_key: "staging_key_from_keyring"
    base_url: "https://staging-api.forwardemail.net"
  development:
    api_key: "dev_key_from_keyring"
    base_url: "https://dev-api.forwardemail.net"

default_profile: development
```

### Secret Management Hierarchy
1. **Command-line flags** (highest priority) - for CI/CD scenarios
2. **Environment variables** - `FORWARDEMAIL_API_KEY`, `FORWARDEMAIL_PROFILE`
3. **OS Keyring** - secure credential storage (mandatory for interactive use)
4. **Configuration file** - profile management and preferences
5. **Interactive prompts** - fallback with secure input

### Security Features
- **Credential Redaction**: Mask tokens in logs and panic reports
- **Profile Indicators**: Visual cues for active environment (`[prod] $`)
- **Audit Logging**: Tamper-evident operation history
- **Secure Defaults**: Force keyring usage, warn on insecure configurations

---

## 5) Command Structure & User Experience

### Command Hierarchy
```bash
# Authentication & Configuration
forward-email init                           # Interactive setup wizard
forward-email config profile add production # Profile management
forward-email auth verify                    # Validate credentials
forward-email auth rotate                    # API key rotation

# Domain Management
forward-email domain list                    # List all domains
forward-email domain add example.com        # Add new domain
forward-email domain verify example.com     # DNS/SMTP verification
forward-email domain delete example.com --confirm

# Alias Operations
forward-email alias list --domain=example.com
forward-email alias create info@example.com --forward-to=team@company.com
forward-email alias bulk-import aliases.csv --dry-run
forward-email alias export --domain=example.com --format=csv

# Email Operations
forward-email send --to=user@example.com --subject="Welcome" --template=welcome.yaml
forward-email send --raw=message.eml       # RFC822 format
forward-email emails list --status=sent    # Outbound email history
forward-email quota check                  # Daily sending limits

# Monitoring & Logs
forward-email logs stream --domain=example.com --follow
forward-email logs download --date=2025-01-15 --format=csv
forward-email health check                 # API and service status

# Utilities
forward-email encrypt "plaintext TXT record"
forward-email completion bash > /etc/bash_completion.d/forward-email
forward-email version --check-updates
```

### Interactive Features
- **Setup Wizard**: Guided configuration for new users
- **Interactive Mode**: `forward-email interactive` for complex workflows
- **Confirmation Prompts**: Safe-by-default for destructive operations
- **Smart Suggestions**: Auto-complete and command suggestions

### Output Formats
```bash
# Standard output options
--output=table|json|yaml|csv|template
--template="{{.Name}}: {{.Recipients}}"
--filter="domain=example.com,verified=true"
--sort="name,created_at"
--no-header                                 # For scripting
--quiet                                     # Minimal output
```

---

## 6) Technical Implementation

### Core Architecture
```
cmd/
├── root.go                    # Cobra root command setup
├── auth/                      # Authentication commands
├── domain/                    # Domain management
├── alias/                     # Alias operations
├── email/                     # Email sending/management
├── logs/                      # Log retrieval
└── config/                    # Configuration management

pkg/
├── api/                       # Forward Email API client
├── config/                    # Configuration management
├── auth/                      # Authentication handling
├── cache/                     # Intelligent caching
├── bulk/                      # Batch operations
├── template/                  # Email templates
└── output/                    # Formatting and display

internal/
├── telemetry/                 # Usage analytics
├── update/                    # Version checking
└── keyring/                   # Secure storage
```

### Key Dependencies
- **Cobra**: Command-line interface framework
- **Viper**: Configuration management
- **Keyring**: Secure credential storage
- **Survey**: Interactive prompts
- **Spinner**: Progress indicators
- **Tablewriter**: Formatted output

### Error Handling Strategy
```go
type CLIError struct {
    Code        string                 // Structured error code
    Message     string                 // Human-readable message
    Suggestion  string                 // Actionable fix suggestion
    Context     map[string]interface{} // Operation context
    Retryable   bool                   // Can operation be retried
}
```

### Caching Implementation
- **API Response Caching**: 5-minute TTL for list operations
- **Authentication Caching**: 1-hour session tokens
- **Configuration Caching**: In-memory profile data
- **Log Caching**: Minimize daily download limit usage

---

## 7) Advanced Features

### Bulk Operations
```bash
# CSV format for bulk imports
# name,recipients,description
info,team@company.com,General inquiries
support,support@company.com,Customer support
sales,sales@company.com,Sales inquiries

forward-email alias bulk-import aliases.csv \
  --domain=example.com \
  --concurrency=5 \
  --rollback-on-error \
  --progress
```

### Template System
```yaml
# welcome.yaml
subject: Welcome to {{.CompanyName}}!
body: |
  Hello {{.UserName}},
  
  Welcome to {{.CompanyName}}! Your account has been created.
  
  Best regards,
  The {{.CompanyName}} Team
variables:
  - name: UserName
    required: true
  - name: CompanyName
    default: "Our Company"
```

### CI/CD Integration
```yaml
# .github/workflows/email-setup.yml
name: Email Configuration
on: [push]
jobs:
  setup:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Setup Forward Email CLI
        uses: forwardemail/setup-cli@v1
        with:
          api-key: ${{ secrets.FORWARD_EMAIL_API_KEY }}
      - name: Validate Configuration
        run: forward-email config validate email-config.yaml
      - name: Deploy Aliases
        run: forward-email alias bulk-import aliases.csv --confirm
```

### Plugin Architecture
```bash
# Plugin discovery and management
forward-email plugin list                   # Available plugins
forward-email plugin install terraform      # Install Terraform provider
forward-email plugin run webhook-manager    # Execute community plugin
```

---

## 8) Quality Assurance & Testing

### Testing Strategy
- **Unit Tests**: 90%+ coverage for critical paths
- **Integration Tests**: Against mock API and staging environment
- **End-to-End Tests**: Real workflow testing with testscript
- **Fuzzing**: Security testing with malformed inputs
- **Performance Tests**: Benchmark CLI operations

### Cross-Platform Support
- **Automated Testing**: GitHub Actions matrix for Windows/macOS/Linux
- **Distribution**: Multiple channels (Homebrew, apt/yum, Docker, GitHub releases)
- **Compatibility**: Go 1.19+ requirement for broad platform support

### Documentation Standards
- **Embedded Help**: Comprehensive `--help` with examples
- **Man Pages**: Generated documentation for Unix systems
- **Interactive Tutorials**: Built-in `forward-email learn` command
- **API Documentation**: Auto-generated from OpenAPI specs

---

## 9) Deployment & Distribution

### Release Pipeline
```bash
# Automated release process
1. Tag creation triggers release workflow
2. Cross-platform binary compilation
3. Package generation (deb, rpm, homebrew)
4. Docker image building and publishing
5. Documentation site update
6. Changelog generation
```

### Distribution Channels
- **GitHub Releases**: Primary distribution with checksums
- **Package Managers**: Homebrew, Chocolatey, Scoop
- **Container Registries**: Docker Hub, GitHub Container Registry
- **Linux Repositories**: apt/yum packages for major distributions

### Update Management
```bash
forward-email version --check              # Check for updates
forward-email update                        # Self-update capability
forward-email changelog                     # View release notes
```

---

## 10) Success Metrics & KPIs

### Technical Performance
- **Command Execution Time**: < 2 seconds for simple operations
- **Error Rate**: < 1% for stable API operations
- **Cache Hit Rate**: > 80% for repeated operations
- **API Quota Efficiency**: < 50% of daily limits under normal usage

### User Adoption
- **GitHub Stars**: 1000+ in first year
- **Community Contributions**: 10+ external contributors
- **Download Growth**: 50% month-over-month
- **User Retention**: 70% active after 30 days

### Quality Metrics
- **Test Coverage**: > 90% for critical functionality
- **Documentation Coverage**: 100% of public commands
- **Issue Resolution**: < 48 hours for critical bugs
- **Feature Request Turnaround**: < 30 days for prioritized features

---

## 11) Development Roadmap

### Phase 1: Foundation (Weeks 1-4)
- [ ] Core authentication and configuration management
- [ ] Basic CRUD operations for domains and aliases
- [ ] Essential error handling and retry logic
- [ ] Table and JSON output formats
- [ ] Shell completion generation

### Phase 2: Enhancement (Weeks 5-8)
- [ ] Bulk operations with progress tracking
- [ ] Interactive setup wizard and guided workflows
- [ ] Template system for email operations
- [ ] Log streaming and caching
- [ ] Advanced output formatting and filtering

### Phase 3: Ecosystem (Weeks 9-12)
- [ ] CI/CD integrations and GitHub Actions
- [ ] Plugin architecture and community features
- [ ] Docker containers and deployment tooling
- [ ] Performance optimizations and caching
- [ ] Comprehensive documentation and tutorials

### Phase 4: Enterprise (Weeks 13-16)
- [ ] Advanced security features and audit logging
- [ ] Enterprise integrations and compliance tools
- [ ] Professional support tooling and monitoring
- [ ] Advanced automation and webhook management
- [ ] Terraform provider and Infrastructure as Code

---

## 12) Risk Mitigation

### Technical Risks
- **API Changes**: Version compatibility checking and graceful degradation
- **Rate Limiting**: Intelligent backoff and quota management
- **Dependency Conflicts**: Minimal dependency tree and vendoring
- **Cross-platform Issues**: Comprehensive testing matrix

### Business Risks
- **Competitive Response**: Focus on superior UX and community building
- **API Sunset**: Prepare for potential API versioning changes
- **User Adoption**: Extensive documentation and migration tools
- **Support Burden**: Self-service tools and community support

---

This enhanced architecture incorporates comprehensive API research, competitive analysis, and industry best practices to position the Forward Email CLI as the definitive developer tool for email automation and management.