# Forward Email CLI

A comprehensive command-line interface for managing [Forward Email](https://forwardemail.net/) accounts and resources through their public REST API. This CLI provides a powerful interface to manage your domains, aliases, and email operations programmatically.

## ✨ Features

- **Complete API Coverage**: All Forward Email endpoints supported
- **Multi-Profile Support**: Development, staging, and production environments  
- **Security First**: OS keyring integration with secure credential storage
- **Developer Experience**: Shell completion, interactive wizards, comprehensive help
- **Enterprise Ready**: Audit logging, CI/CD integration, bulk operations
- **Multiple Output Formats**: Table, JSON, YAML, CSV with filtering and sorting

## 🚀 Quick Start

```bash
# Build from source
git clone https://github.com/ginsys/forward-email.git
cd forward-email
go build -o bin/forward-email ./cmd/forward-email

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

# Email Operations
forward-email email send             # Interactive email composition
forward-email email list             # View sent email history
```

## 📌 Version Policy

**Current Status**: Pre-release development (v0.x.x)

This project is under active development towards v1.0.0. Until then:
- ⚠️ **Breaking changes** may occur in any release
- ⚠️ **No backwards compatibility** guaranteed  
- ⚠️ **API interfaces** may change without deprecation notices
- ✅ After v1.0.0: Semantic versioning with proper deprecation cycles

The CLI targets Forward Email API v1. While we strive for stability, please pin to specific versions in production use.

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

- **[Quick Start Guide](docs/quick-start.md)** - Get up and running quickly
- **[Command Reference](docs/commands.md)** - Complete command documentation
- **[Configuration Guide](docs/configuration.md)** - Profiles, environments, and settings
- **[Troubleshooting](docs/troubleshooting.md)** - Common issues and solutions

### For Developers
- **[Contributing Guide](docs/development/contributing.md)** - How to contribute to the project
- **[Architecture Overview](docs/development/architecture.md)** - System design and structure
- **[API Integration](docs/development/api-integration.md)** - Forward Email API details
- **[Testing Strategy](docs/development/testing.md)** - Testing approach and standards

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](docs/development/contributing.md) for details.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Forward Email](https://forwardemail.net) for providing the comprehensive API
- [Cobra](https://github.com/spf13/cobra) for the excellent CLI framework
- The Go community for outstanding tooling and libraries
