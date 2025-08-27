# Command Reference

Complete documentation for all Forward Email CLI commands with current implementation status.

**Status**: All commands fully implemented and tested ✅

## Global Flags

These flags are available for all commands:

```bash
--debug               Enable debug output
--help, -h            Help for any command
--output, -o string   Output format (table|json|yaml|csv) (default "table")
--profile, -p string  Configuration profile to use
--timeout duration    Request timeout duration
--verbose, -v         Enable verbose output
```

## Authentication Commands (`auth`)

Manage authentication credentials for Forward Email API.

### Available Subcommands
- `login` - Interactive API key setup with secure input
- `logout` - Clear stored credentials from keyring
- `status` - Show current authentication status
- `verify` - Validate credentials against the API

```bash
# Interactive login (recommended)
forward-email auth login

# Login for specific profile
forward-email auth login --profile production

# Check authentication status
forward-email auth status

# Verify credentials
forward-email auth verify

# Logout (clear stored credentials)
forward-email auth logout
```

## Profile Commands (`profile`)

Manage configuration profiles for different environments.

### Available Subcommands
- `create` - Create a new profile
- `delete` - Delete a profile
- `list` - List all profiles
- `show` - Show profile details
- `switch` - Switch to a different profile

```bash
# List all profiles
forward-email profile list

# Create production profile
forward-email profile create production

# Switch to production profile
forward-email profile switch production

# Show current profile details
forward-email profile show

# Delete a profile
forward-email profile delete staging
```

## Domain Commands (`domain`)

Complete domain lifecycle management.

### Available Subcommands
- `create` - Create a new domain
- `delete` - Delete a domain
- `get` - Get domain details
- `list` - List domains
- `members` - Manage domain members
- `update` - Update domain settings
- `verify` - DNS/SMTP verification

```bash
# List all domains
forward-email domain list

# Create new domain
forward-email domain create example.com

# Get domain details
forward-email domain get example.com

# Verify domain DNS settings
forward-email domain verify example.com

# List domain members
forward-email domain members example.com

# Update domain settings
forward-email domain update example.com --max-recipients 5
```

**Output Formats**: All commands support `--output table|json|yaml|csv`

## Alias Commands (`alias`)

Comprehensive alias management with all Forward Email features.

### Available Subcommands
- `create` - Create a new alias
- `delete` - Delete an alias
- `disable` - Disable an alias
- `enable` - Enable an alias
- `get` - Get alias details
- `list` - List aliases
- `password` - Generate IMAP password
- `quota` - Show alias quota
- `recipients` - Update alias recipients
- `stats` - Show alias statistics
- `update` - Update alias settings

```bash
# List aliases for domain
forward-email alias list --domain example.com

# Create new alias
forward-email alias create info@example.com --domain example.com --recipients team@company.com

# Get alias details
forward-email alias get info@example.com --domain example.com

# Update recipients
forward-email alias recipients info@example.com --domain example.com --recipients new@company.com

# Enable/disable alias
forward-email alias disable info@example.com --domain example.com
forward-email alias enable info@example.com --domain example.com

# Generate IMAP password
forward-email alias password info@example.com --domain example.com

# Show alias statistics
forward-email alias stats --domain example.com
```

**Domain Flag**: Most alias commands require `--domain` flag to specify the domain.

## Email Commands (`email`)

Send and manage emails with attachment support.

### Available Subcommands
- `delete` - Delete sent emails
- `get` - Get email details
- `list` - List sent emails
- `quota` - Show email quota
- `send` - Send emails (interactive or command-line)

```bash
# Interactive email composition
forward-email email send

# Command-line email sending
forward-email email send --to recipient@example.com --subject "Hello" --body "Message"

# Send with attachment
forward-email email send --to recipient@example.com --subject "Files" --attach /path/to/file.pdf

# List sent emails
forward-email email list

# Get email details
forward-email email get <email-id>

# Check email quota
forward-email email quota

# Delete sent email
forward-email email delete <email-id>
```

**Features**: Interactive composition wizard, attachment support, dry-run mode, custom headers.

## Debug Commands (`debug`)

Troubleshooting utilities for system diagnostics.

### Available Subcommands
- `auth` - Debug authentication system
- `keyring` - Debug keyring operations
- `api` - Test API connectivity

```bash
# Debug authentication
forward-email debug auth

# Debug keyring operations
forward-email debug keyring

# Test API connectivity
forward-email debug api
```

## Completion Commands (`completion`)

Generate shell completion scripts.

### Available Shells
- `bash` - Bash completion script
- `fish` - Fish completion script
- `powershell` - PowerShell completion script
- `zsh` - Zsh completion script

```bash
# Generate bash completion
forward-email completion bash > /usr/local/etc/bash_completion.d/forward-email

# Generate zsh completion
forward-email completion zsh > /usr/local/share/zsh/site-functions/_forward-email

# Generate fish completion
forward-email completion fish > ~/.config/fish/completions/forward-email.fish
```

## Error Handling

All commands implement consistent error handling:

- **HTTP Errors**: Mapped to user-friendly messages
- **Validation Errors**: Clear indication of invalid inputs
- **Network Errors**: Timeout and connectivity guidance
- **Authentication Errors**: Clear credential resolution steps

## Output Formats

All commands support multiple output formats:

- **table** (default): Human-readable tabular format
- **json**: Machine-readable JSON format
- **yaml**: YAML format for configuration files
- **csv**: Comma-separated values for data processing

```bash
# JSON output for scripting
forward-email domain list --output json

# CSV output for spreadsheets
forward-email alias list --domain example.com --output csv

# YAML output for configuration
forward-email profile show --output yaml
```

---

*Last Updated: 2025-08-27 | All commands tested and fully functional*