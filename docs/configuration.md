# Configuration Guide

Learn how to configure Forward Email CLI for your environment and workflow.

**Status**: Fully implemented with OS keyring integration ✅

## Configuration Overview

Forward Email CLI uses a hierarchical configuration system that prioritizes settings in this order:

1. **Command-line flags** (highest priority)
2. **Environment variables**
3. **OS keyring storage** (for API keys)
4. **Profile-specific settings**
5. **Global configuration file**
6. **Default values** (lowest priority)

### Authentication Priority
For API key authentication specifically:
1. **Environment variable** (`FORWARD_EMAIL_API_KEY`)
2. **OS keyring storage** (secure credential storage)
3. **Profile configuration** (config file - not recommended for production)

## Configuration File

### Location

The configuration file is stored in:

- **Linux/macOS**: `~/.config/forwardemail/config.yaml`
- **Windows**: `%APPDATA%\forwardemail\config.yaml`

### Structure

```yaml
# Current active profile
current_profile: default

# Profile configurations
profiles:
  default:
    base_url: "https://api.forwardemail.net"
    timeout: "30s"
    output: "table"
    
  production:
    base_url: "https://api.forwardemail.net"
    timeout: "30s"
    output: "json"
    
  staging:
    base_url: "https://staging-api.forwardemail.net"
    timeout: "45s"
    output: "table"
```

## Profiles

Profiles allow you to manage multiple Forward Email accounts or environments easily.

### Creating Profiles

```bash
# Create a new profile
forward-email profile create production

# Create with custom settings
forward-email profile create staging \
  --base-url https://staging-api.forwardemail.net \
  --timeout 45s \
  --output json
```

### Managing Profiles

```bash
# List all profiles
forward-email profile list

# Show current profile details
forward-email profile show

# Show specific profile
forward-email profile show production

# Switch to different profile
forward-email profile switch production

# Delete a profile
forward-email profile delete old-profile --force
```

### Profile Settings

Each profile can have these settings:

| Setting | Description | Default |
|---------|-------------|---------|
| `base_url` | API endpoint URL | `https://api.forwardemail.net` |
| `timeout` | Request timeout duration | `30s` |
| `output` | Default output format | `table` |

## Authentication

### Credential Storage

Forward Email CLI stores credentials securely using your operating system's keyring:

- **macOS**: Keychain
- **Windows**: Credential Manager  
- **Linux**: Secret Service (gnome-keyring, KWallet, etc.)

#### Credential Store Options

- CLI flags (choose at login/init):
  - `--store auto` (default): Try system keyring; fall back to config if unavailable.
  - `--store keyring`: Force system keyring (may prompt via OS UI).
  - `--store file`: Encrypted file under `~/.config/forwardemail/keyring` (use `--file-pass` or provide when prompted).
  - `--store config`: Store in `config.yaml` (not recommended).

- Environment overrides (non-interactive/CI):
  - `FORWARDEMAIL_KEYRING_BACKEND=none` disables keyring usage (uses config file).
  - `FORWARDEMAIL_KEYRING_BACKEND=file` forces encrypted file backend.
  - `FORWARDEMAIL_KEYRING_PASSWORD=<passphrase>` passphrase for file backend.

- Testing/CI: All `make test*` targets run with `FORWARDEMAIL_KEYRING_BACKEND=none` to avoid GUI prompts.

### Credential Hierarchy

Credentials are resolved in this order:

1. **Environment variables** (highest priority)
2. **OS keyring** (recommended for interactive use)
3. **Configuration file** (not recommended for API keys)
4. **Interactive prompt** (fallback)

### Environment Variables

#### Generic API Key

```bash
export FORWARDEMAIL_API_KEY="your-api-key-here"
```

#### Profile-Specific API Keys

```bash
export FORWARDEMAIL_PRODUCTION_API_KEY="prod-api-key"
export FORWARDEMAIL_STAGING_API_KEY="staging-api-key"
export FORWARDEMAIL_DEVELOPMENT_API_KEY="dev-api-key"
```

#### Profile Selection

```bash
export FORWARDEMAIL_PROFILE="production"
```

### Managing Credentials

```bash
# Interactive login (stores in keyring)
forward-email auth login

# Login for specific profile
forward-email auth login --profile production

# Verify credentials
forward-email auth verify

# Check authentication status
forward-email auth status

# Logout (clears keyring)
forward-email auth logout

# Logout all profiles
forward-email auth logout --all
```

## Environment Variables

### Complete List

| Variable | Description | Example |
|----------|-------------|---------|
| `FORWARDEMAIL_API_KEY` | Global API key | `your-api-key` |
| `FORWARDEMAIL_{PROFILE}_API_KEY` | Profile-specific API key | `FORWARDEMAIL_PROD_API_KEY` |
| `FORWARDEMAIL_PROFILE` | Active profile name | `production` |
| `FORWARDEMAIL_BASE_URL` | API base URL | `https://api.forwardemail.net` |
| `FORWARDEMAIL_TIMEOUT` | Request timeout | `30s` |
| `FORWARDEMAIL_OUTPUT` | Default output format | `table` |
| `FORWARDEMAIL_DEBUG` | Enable debug mode | `true` |

### CI/CD Usage

For continuous integration and deployment:

```bash
# Set API key for CI/CD
export FORWARDEMAIL_API_KEY="$CI_FORWARD_EMAIL_API_KEY"

# Use JSON output for parsing
export FORWARDEMAIL_OUTPUT="json"

# Run commands
forward-email domain list
forward-email alias create ci-test --domain example.com --recipients test@company.com
```

## Output Formats

Configure default output formats for different use cases:

### Available Formats

- **table**: Human-readable tables (default)
- **json**: Machine-readable JSON
- **yaml**: Human-readable YAML
- **csv**: Spreadsheet-compatible CSV

### Setting Default Format

```bash
# Per command
forward-email domain list --output json

# Per profile
forward-email profile create automation --output json

# Via environment variable
export FORWARDEMAIL_OUTPUT="json"
```

### Format Examples

#### Table Format (Default)
```
┌─────────────┬──────────┬─────────────┬─────────┐
│ NAME        │ VERIFIED │ PLAN        │ ALIASES │
├─────────────┼──────────┼─────────────┼─────────┤
│ example.com │ true     │ enhanced    │ 5       │
│ test.com    │ false    │ free        │ 2       │
└─────────────┴──────────┴─────────────┴─────────┘
```

#### JSON Format
```json
[
  {
    "id": "507f1f77bcf86cd799439011",
    "name": "example.com",
    "verified": true,
    "plan": "enhanced",
    "alias_count": 5
  }
]
```

#### YAML Format
```yaml
- id: 507f1f77bcf86cd799439011
  name: example.com
  verified: true
  plan: enhanced
  alias_count: 5
```

#### CSV Format
```csv
id,name,verified,plan,alias_count
507f1f77bcf86cd799439011,example.com,true,enhanced,5
```

## Advanced Configuration

### Timeout Settings

Configure timeouts for different scenarios:

```bash
# Short timeout for health checks
forward-email debug api --timeout 5s

# Long timeout for bulk operations
forward-email alias bulk-import aliases.csv --timeout 300s
```

### Debug Mode

Enable debug output for troubleshooting:

```bash
# Enable debug for single command
forward-email domain list --debug

# Enable debug via environment
export FORWARDEMAIL_DEBUG=true
forward-email domain list

# Debug specific operations
forward-email debug auth
forward-email debug api
forward-email debug keys
```

### Custom Base URLs

For testing or custom deployments:

```bash
# Create profile with custom URL
forward-email profile create testing --base-url https://test-api.forwardemail.net

# Use custom URL for single command
forward-email domain list --profile testing
```

## Multi-Environment Workflows

### Example: Development → Staging → Production

```bash
# Setup environments
forward-email profile create development
forward-email profile create staging
forward-email profile create production

# Configure each environment
forward-email profile switch development
forward-email auth login  # Use development API key

forward-email profile switch staging
forward-email auth login  # Use staging API key

forward-email profile switch production
forward-email auth login  # Use production API key

# Daily workflow
forward-email profile switch development
forward-email alias create new-feature --domain dev.example.com --recipients dev@company.com

forward-email profile switch staging
forward-email alias create new-feature --domain staging.example.com --recipients staging@company.com

forward-email profile switch production
forward-email alias create new-feature --domain example.com --recipients production@company.com
```

### Example: Team Collaboration

```bash
# Each team member sets up their own profile
forward-email profile create alice
forward-email profile create bob
forward-email profile create charlie

# Switch between team member contexts
forward-email profile switch alice
forward-email domain list  # Alice's domains

forward-email profile switch bob
forward-email domain list  # Bob's domains
```

## Security Best Practices

### API Key Management

1. **Use OS keyring** for interactive usage
2. **Use environment variables** for CI/CD
3. **Never commit API keys** to version control
4. **Rotate keys regularly** in your Forward Email account
5. **Use profile-specific keys** for different environments

### Credential Verification

```bash
# Regular credential checks
forward-email auth verify

# Verify all profiles
for profile in $(forward-email profile list --output json | jq -r '.[].name'); do
  echo "Checking $profile..."
  forward-email auth verify --profile "$profile"
done
```

### Access Control

```bash
# Limit API key permissions in Forward Email dashboard
# Use read-only keys for monitoring scripts
# Use full-access keys only when necessary
```

## Troubleshooting Configuration

### Common Issues

#### "Profile not found"
```bash
# List available profiles
forward-email profile list

# Create missing profile
forward-email profile create missing-profile
```

#### "Authentication failed"
```bash
# Check authentication status
forward-email auth status

# Re-authenticate
forward-email auth logout
forward-email auth login
```

#### "Keyring access denied"
```bash
# Debug keyring access
forward-email debug keys

# Use environment variable as fallback
export FORWARDEMAIL_API_KEY="your-api-key"
```

### Configuration Validation

```bash
# Validate current configuration
forward-email debug auth

# Test API connectivity
forward-email debug api

# Check profile settings
forward-email profile show
```

## Configuration Examples

### Minimal Configuration

```yaml
current_profile: default
profiles:
  default:
    base_url: "https://api.forwardemail.net"
    timeout: "30s"
    output: "table"
```

### Multi-Environment Configuration

```yaml
current_profile: development
profiles:
  development:
    base_url: "https://api.forwardemail.net"
    timeout: "30s"
    output: "table"
  staging:
    base_url: "https://api.forwardemail.net"
    timeout: "45s"
    output: "json"
  production:
    base_url: "https://api.forwardemail.net"
    timeout: "60s"
    output: "json"
```

### CI/CD Configuration

```yaml
current_profile: ci
profiles:
  ci:
    base_url: "https://api.forwardemail.net"
    timeout: "120s"
    output: "json"
```

## See Also

For more information on specific commands and troubleshooting, see:
- [Command Reference](commands.md) - Complete command documentation
- [Quick Start Guide](quick-start.md) - Getting started tutorial
- [Troubleshooting Guide](troubleshooting.md) - Problem resolution

---

*Last Updated: 2025-08-27 | Configuration system fully implemented and tested*

---

Docs navigation: [Prev: Command Reference](commands.md) | [Next: Troubleshooting](troubleshooting.md)
