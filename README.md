# Forward Email CLI

A comprehensive command-line interface for managing [Forward Email](https://forwardemail.net/) accounts and resources through their public REST API.

[Forward Email](https://forwardemail.net/) is a free, encrypted, and open-source email forwarding service for custom domains. This CLI provides a powerful interface to manage your domains, aliases, and email operations programmatically.

## 🚧 Development Status

**Current Phase**: Foundation (Phase 1.1) - Architecture & Core Infrastructure

**Implemented**:
- ✅ Project structure with clean separation (pkg/api, internal/cmd, cmd/)
- ✅ Build system with Makefile and cross-platform support
- ✅ Core CLI framework (Cobra + Viper)
- ✅ Basic API client foundation with authentication interface
- ✅ Configuration management foundation

**In Progress**:
- 🔄 Authentication system (API key management, credential validation)
- 🔄 Domain operations (CRUD, verification, settings)
- 🔄 HTTP client with retry logic and error handling

**Next**: Alias operations, email operations, utility commands

For detailed roadmap see [Implementation Plan](IMPLEMENTATION_PLAN.md). This project represents a **first-mover advantage** as Forward Email currently has no official CLI tools.

## 🚀 Quick Start

```bash
# Install via Homebrew (coming soon)
brew install ginsys/tap/forwardemail-cli

# Or download from releases
curl -sSL https://github.com/ginsys/forwardemail-cli/releases/latest/download/install.sh | bash

# Initialize configuration
forward-email init

# Get started
forward-email domain list
```

## ✨ Features

- **Complete API Coverage**: All Forward Email endpoints supported
- **Multi-Profile Support**: Development, staging, and production environments
- **Security First**: OS keyring integration, credential redaction
- **Developer Experience**: Shell completion, interactive wizards, comprehensive help
- **Enterprise Ready**: Audit logging, CI/CD integration, bulk operations
- **Multiple Output Formats**: Table, JSON, YAML, CSV with filtering and sorting

## 📋 Command Overview

### Authentication & Configuration
```bash
forward-email init                           # Interactive setup wizard
forward-email auth verify                    # Validate credentials
forward-email config profile add production # Profile management
```

### Domain Management
```bash
forward-email domain list                    # List all domains
forward-email domain add example.com        # Add new domain
forward-email domain verify example.com     # DNS/SMTP verification
```

### Alias Operations
```bash
forward-email alias list --domain=example.com
forward-email alias create info@example.com --forward-to=team@company.com
forward-email alias bulk-import aliases.csv --dry-run
```

### Email Operations
```bash
forward-email send --to=user@example.com --subject="Welcome" --template=welcome.yaml
forward-email emails list --status=sent    # Outbound email history
forward-email quota check                  # Daily sending limits
```

### Monitoring & Logs
```bash
forward-email logs stream --domain=example.com --follow
forward-email logs download --date=2025-01-15 --format=csv
forward-email health check                 # API and service status
```

## 🏗️ Architecture

- **Clean Separation**: SDK (pkg/api) → CLI commands (cmd/) → User interface
- **Security First**: OS keyring integration, credential redaction, secure defaults
- **Developer Experience**: Shell completion, interactive wizards, comprehensive help
- **Enterprise Ready**: Multi-profile, audit logging, CI/CD integration

## 🛠️ Development

```bash
# Clone repository
git clone https://github.com/ginsys/forwardemail-cli.git
cd forwardemail-cli

# Install dependencies
go mod download

# Run tests
go test ./...

# Build local binary
go build -o bin/forward-email ./cmd/forwardemail-cli

# Install development version
go install ./cmd/forwardemail-cli
```

## 📚 Documentation

- [Architecture Overview](docs/forwardemail_cli_architecture_0.2.md)
- [API Reference](docs/api.md)
- [Configuration Guide](docs/configuration.md)
- [Developer Guide](docs/development.md)

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Forward Email](https://forwardemail.net) for providing the comprehensive API
- [Cobra](https://github.com/spf13/cobra) for the excellent CLI framework
- The Go community for outstanding tooling and libraries