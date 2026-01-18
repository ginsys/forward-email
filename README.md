# Forward Email CLI

A comprehensive command-line interface for managing [Forward Email](https://forwardemail.net/) accounts and resources through their public REST API. This CLI provides a powerful interface to manage your domains, aliases, and email operations programmatically.

**Status**: Production Ready ✅ | **LOC**: 17,151 | **Tests**: 100+ passing | **Platform**: Linux/macOS/Windows

## ✨ Features

- **Complete API Coverage**: Full domain, alias, and email management
- **Multi-Profile Support**: Development, staging, and production environments  
- **Security First**: OS keyring integration with secure credential storage
- **Developer Experience**: Comprehensive help system and interactive operations
- **Multiple Output Formats**: Table, JSON, YAML, CSV with filtering and sorting
- **Cross-Platform**: Native support for Linux, macOS, and Windows
- **First-Mover Advantage**: Forward Email's first official CLI tool

## 🚀 Quick Start

```bash
# Build from source
git clone https://github.com/ginsys/forward-email.git
cd forward-email
go build -o bin/forward-email ./cmd/forward-email

# Check version (optional)
./bin/forward-email version --verbose

# Set up authentication
./bin/forward-email auth login

# List your domains
./bin/forward-email domain list

# Create an alias
./bin/forward-email alias create info@example.com --domain example.com --recipients team@company.com
```

## 📋 Core Commands

```bash
# Authentication & Profiles
forward-email auth login              # Interactive API key setup
forward-email profile create prod    # Create production profile
forward-email profile switch prod    # Switch to production profile

# Domain Management  
forward-email domain list            # List all domains
forward-email domain create example.com    # Add new domain
forward-email domain verify example.com    # DNS/SMTP verification

# Alias Operations
forward-email alias list --domain example.com
forward-email alias create info@example.com --domain example.com --recipients team@company.com
forward-email alias sync merge domain1.com domain2.com    # Sync aliases between domains

# Email Operations
forward-email email send             # Interactive email composition
forward-email email list             # View sent email history
```

## 📊 Current Status

### ✅ Fully Implemented
- **Authentication**: Multi-source credential management (env → keyring → config)
- **Profiles**: Multi-environment configuration management  
- **Domains**: Complete CRUD operations with DNS verification
- **Aliases**: Full lifecycle management with all settings
- **Email**: Interactive and programmatic sending with attachments
- **Output**: Multiple formats with filtering and sorting
- **Testing**: 100+ test cases across all packages

### 🔄 In Development (Phase 1.4)
- Enhanced test coverage for email services
- Interactive setup wizards
- Shell completion scripts  
- Bulk operations and CSV import/export
- Automated CI/CD pipeline

## 📌 Version Policy

**Current Status**: Pre-release development (v0.x.x)

This project is under active development towards v1.0.0. Until then:
- ⚠️ **Breaking changes** may occur in any release
- ⚠️ **No backwards compatibility** guaranteed  
- ⚠️ **API interfaces** may change without deprecation notices
- ✅ After v1.0.0: Semantic versioning with proper deprecation cycles

The CLI targets Forward Email API v1. While we strive for stability, please pin to specific versions in production use.

### Release Process
See the Versioning & Release Plan for details on semantic versioning, tagging, and CI/CD automation: `docs/VERSIONING_RELEASE_PLAN.md`.

## 🛠️ Development

```bash
# Install dependencies
go mod download

# Run tests
go test ./...

# Build binary
make build

# Run linter
make lint
```

## 📚 Documentation

- **[Documentation Index](docs/README.md)** - Central entry point for all docs
- **[Releasing](docs/RELEASING.md)** - Tagging and publishing guide
- **[Quick Start Guide](docs/quick-start.md)** - Get up and running quickly
- **[Command Reference](docs/commands.md)** - Complete command documentation
- **[Configuration Guide](docs/configuration.md)** - Profiles, environments, and settings
- **[Troubleshooting](docs/troubleshooting.md)** - Common issues and solutions

### For Developers
- **[Contributing Guide](docs/development/contributing.md)** - How to contribute to the project
- **[Architecture Overview](docs/development/architecture.md)** - System design and structure
- **[API Integration](docs/development/api-integration.md)** - Forward Email API details
- **[Testing Strategy](docs/development/testing.md)** - Testing approach and standards
- **[Domain Alias Sync Specification](docs/development/domain-alias-sync-specification.md)** - Bulk alias synchronization feature
- **Versioning & Release Plan** - See `docs/VERSIONING_RELEASE_PLAN.md`

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](docs/development/contributing.md) for details.

## 🏗️ Architecture

- **Framework**: Cobra CLI with Viper configuration
- **Authentication**: HTTP Basic with Forward Email API
- **Storage**: OS keyring integration for secure credentials
- **Testing**: Comprehensive mock-based testing strategy
- **Output**: Consistent formatting across all operations

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Forward Email](https://forwardemail.net) for providing the comprehensive API
- [Cobra](https://github.com/spf13/cobra) for the excellent CLI framework
- The Go community for outstanding tooling and libraries

---

**Built with ❤️ for the Forward Email community** | *Last Updated: 2025-08-27*
