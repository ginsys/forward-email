# Quick Start Guide

Get up and running with Forward Email CLI in minutes.

## Prerequisites

- Go 1.24+ (for building from source)
- A [Forward Email](https://forwardemail.net/) account with API access
- Your Forward Email API key

## Installation

### Option 1: Build from Source (Recommended)

```bash
# Clone the repository
git clone https://github.com/ginsys/forward-email.git
cd forward-email

# Build the CLI
go build -o bin/forward-email ./cmd/forward-email

# Optionally install to PATH
sudo cp bin/forward-email /usr/local/bin/
```

### Option 2: Download Binary (Coming Soon)

```bash
# Download latest release
curl -sSL https://github.com/ginsys/forward-email/releases/latest/download/install.sh | bash
```

## Initial Setup

### 1. Get Your API Key

1. Log in to your [Forward Email](https://forwardemail.net/) account
2. Navigate to **Settings** → **API**
3. Generate a new API key
4. Copy the API key (you'll need it in the next step)

### 2. Configure Authentication

```bash
# Interactive login (recommended)
forward-email auth login

# The CLI will prompt you to enter your API key securely
# Your key will be stored in your OS keyring (Keychain on macOS, Credential Manager on Windows, Secret Service on Linux)
```

### 3. Verify Your Setup

```bash
# Check authentication status
forward-email auth status

# Test API connectivity
forward-email debug api
```

## First Steps

### List Your Domains

```bash
# View all your domains
forward-email domain list

# Get detailed information about a specific domain
forward-email domain get example.com
```

### Create Your First Alias

```bash
# Create a simple alias
forward-email alias create info@example.com \
  --domain example.com \
  --recipients team@yourcompany.com

# List aliases for a domain
forward-email alias list --domain example.com
```

### Send Your First Email

```bash
# Interactive email composition
forward-email email send

# Or send directly with flags
forward-email email send \
  --from info@example.com \
  --to user@example.com \
  --subject "Welcome!" \
  --body "Hello from Forward Email CLI"
```

## Multi-Profile Setup (Optional)

If you work with multiple Forward Email accounts or environments:

```bash
# Create a production profile
forward-email profile create production
forward-email profile switch production
forward-email auth login  # Login with production API key

# Create a development profile
forward-email profile create development
forward-email profile switch development
forward-email auth login  # Login with development API key

# Switch between profiles
forward-email profile switch production
forward-email profile switch development

# List all profiles
forward-email profile list
```

## Basic Workflow Examples

### Domain Management

```bash
# Add a new domain
forward-email domain create newdomain.com

# Verify DNS setup
forward-email domain verify newdomain.com

# Update domain settings
forward-email domain update newdomain.com --plan enhanced
```

### Alias Management

```bash
# Create multiple aliases
forward-email alias create support@example.com --domain example.com --recipients support@company.com
forward-email alias create sales@example.com --domain example.com --recipients sales@company.com

# Enable/disable aliases
forward-email alias disable support@example.com --domain example.com
forward-email alias enable support@example.com --domain example.com

# Update alias recipients
forward-email alias recipients support@example.com --domain example.com --recipients support@company.com,backup@company.com
```

### Email Operations

```bash
# Check sending quota
forward-email email quota

# List sent emails
forward-email email list

# Get email details
forward-email email get <email-id>
```

## Output Formats

The CLI supports multiple output formats for automation and integration:

```bash
# Table format (default)
forward-email domain list

# JSON format
forward-email domain list --output json

# YAML format
forward-email domain list --output yaml

# CSV format for spreadsheets
forward-email domain list --output csv
```

## Getting Help

### Command Help

```bash
# General help
forward-email --help

# Command-specific help
forward-email domain --help
forward-email alias create --help
```

### Check CLI Version

```bash
# Short version
forward-email version

# Detailed info
forward-email version --verbose

# JSON (great for scripts)
forward-email version --json | jq
```

### Debug Information

```bash
# Debug authentication
forward-email debug auth

# Debug API connectivity
forward-email debug api

# Debug keyring access
forward-email debug keys
```

## What's Next?

- **[Command Reference](commands.md)** - Complete documentation of all commands
- **[Configuration Guide](configuration.md)** - Advanced configuration options
- **[Troubleshooting](troubleshooting.md)** - Solutions to common issues

## Common Issues

### Authentication Problems

```bash
# If login fails, verify your API key
forward-email auth verify

# Clear stored credentials and re-login
forward-email auth logout
forward-email auth login
```

### Permission Issues

```bash
# If you get permission errors, check your API key permissions
forward-email debug api

# Ensure your API key has the necessary permissions in your Forward Email account
```

### Profile Issues

```bash
# If commands fail, check your current profile
forward-email profile show

# Switch to the correct profile
forward-email profile switch <profile-name>
```

For more detailed troubleshooting, see the [Troubleshooting Guide](troubleshooting.md).

---

*Last Updated: 2025-08-27 | All features implemented and tested*

---

Docs navigation: [Prev: Docs Index](README.md) | [Next: Command Reference](commands.md)
# Optional: choose where to store credentials
# - keyring: OS keychain/credential manager
# - file: encrypted file under config dir (requires passphrase)
# - config: plain config file (not recommended)
# forward-email auth login --store keyring
