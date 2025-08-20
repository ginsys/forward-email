# Command Reference

Complete documentation for all Forward Email CLI commands.

## Global Flags

These flags are available for all commands:

```bash
--debug               Enable debug output
--help, -h            Help for any command
--output, -o string   Output format (table|json|yaml|csv) (default "table")
--profile, -p string  Configuration profile to use
--timeout duration    Request timeout duration
```

## Authentication Commands

### `forward-email auth`

Manage authentication credentials.

#### `auth login`

Interactive API key setup with secure input.

```bash
forward-email auth login [flags]

# Interactive login (recommended)
forward-email auth login

# Login for specific profile
forward-email auth login --profile production
```

**Flags:**
- `--profile, -p string`: Profile to configure (default: current profile)

#### `auth verify`

Validate current credentials against the API.

```bash
forward-email auth verify [flags]

# Verify current profile credentials
forward-email auth verify

# Verify specific profile
forward-email auth verify --profile production
```

#### `auth status`

Show authentication status across all profiles.

```bash
forward-email auth status

# Output shows:
# - Current profile
# - Authentication status for each profile
# - API key source (keyring, environment, etc.)
```

#### `auth logout`

Clear stored credentials (single or all profiles).

```bash
forward-email auth logout [flags]

# Logout current profile
forward-email auth logout

# Logout specific profile
forward-email auth logout --profile production

# Logout all profiles
forward-email auth logout --all
```

**Flags:**
- `--all`: Clear credentials for all profiles
- `--profile, -p string`: Profile to logout

## Profile Management Commands

### `forward-email profile`

Manage configuration profiles for different environments.

#### `profile list`

List all configured profiles.

```bash
forward-email profile list

# Shows:
# - Profile names
# - Current profile indicator
# - Authentication status
# - Base URL
```

#### `profile show`

Show current or specific profile details.

```bash
forward-email profile show [profile-name]

# Show current profile
forward-email profile show

# Show specific profile
forward-email profile show production
```

#### `profile create`

Create new profile configuration.

```bash
forward-email profile create <name> [flags]

# Create production profile
forward-email profile create production

# Create with custom base URL
forward-email profile create staging --base-url https://staging-api.forwardemail.net
```

**Flags:**
- `--base-url string`: API base URL (default: https://api.forwardemail.net)
- `--timeout duration`: Request timeout (default: 30s)
- `--output string`: Default output format

#### `profile switch`

Switch to different profile.

```bash
forward-email profile switch <name>

# Switch to production
forward-email profile switch production
```

#### `profile delete`

Delete profile configuration.

```bash
forward-email profile delete <name> [flags]

# Delete with confirmation
forward-email profile delete old-profile --force
```

**Flags:**
- `--force`: Skip confirmation prompt

## Domain Management Commands

### `forward-email domain`

Manage Forward Email domains.

#### `domain list`

List all domains with filtering and pagination.

```bash
forward-email domain list [flags]

# List all domains
forward-email domain list

# Filter by verification status
forward-email domain list --verified true

# Search by name
forward-email domain list --search example

# Sort by creation date
forward-email domain list --sort created --order desc

# Paginate results
forward-email domain list --page 2 --limit 50
```

**Flags:**
- `--verified bool`: Filter by verification status
- `--plan string`: Filter by subscription plan (free|enhanced|team)
- `--search string`: Search domain names
- `--sort string`: Sort criteria (name|created|updated)
- `--order string`: Sort order (asc|desc)
- `--page int`: Page number (default: 1)
- `--limit int`: Items per page (default: 25)

#### `domain get`

Get detailed domain information.

```bash
forward-email domain get <domain-name-or-id>

# Get by domain name
forward-email domain get example.com

# Get by domain ID
forward-email domain get 507f1f77bcf86cd799439011
```

#### `domain create`

Create new domain.

```bash
forward-email domain create <name> [flags]

# Create basic domain
forward-email domain create example.com

# Create with specific plan
forward-email domain create example.com --plan enhanced
```

**Flags:**
- `--plan string`: Subscription plan (free|enhanced|team)

#### `domain update`

Update domain settings.

```bash
forward-email domain update <domain-or-id> [flags]

# Update domain plan
forward-email domain update example.com --plan enhanced

# Update domain settings
forward-email domain update example.com --has-mx-record true
```

**Flags:**
- `--plan string`: Subscription plan
- `--has-mx-record bool`: Domain has MX record configured
- `--has-txt-record bool`: Domain has TXT record configured

#### `domain delete`

Delete domain.

```bash
forward-email domain delete <domain-or-id> [flags]

# Delete with confirmation
forward-email domain delete example.com --force
```

**Flags:**
- `--force`: Skip confirmation prompt

#### `domain verify`

Verify domain DNS configuration.

```bash
forward-email domain verify <domain-or-id>

# Verify DNS records
forward-email domain verify example.com
```

#### `domain dns`

Show required DNS records.

```bash
forward-email domain dns <domain-or-id>

# Show DNS records to configure
forward-email domain dns example.com
```

#### `domain quota`

Show domain quota and usage.

```bash
forward-email domain quota <domain-or-id>

# Show quota information
forward-email domain quota example.com
```

#### `domain stats`

Show domain statistics.

```bash
forward-email domain stats <domain-or-id>

# Show domain statistics
forward-email domain stats example.com
```

#### `domain members`

Manage domain members.

```bash
forward-email domain members <domain-or-id> [flags]

# List members
forward-email domain members example.com

# Add member
forward-email domain members example.com --add user@example.com --role admin

# Remove member
forward-email domain members example.com --remove user@example.com
```

**Flags:**
- `--add string`: Add member email
- `--remove string`: Remove member email
- `--role string`: Member role (admin|user)

## Alias Management Commands

### `forward-email alias`

Manage Forward Email aliases.

#### `alias list`

List all aliases for a domain.

```bash
forward-email alias list --domain <domain> [flags]

# List all aliases
forward-email alias list --domain example.com

# Filter by enabled status
forward-email alias list --domain example.com --enabled true

# Search aliases
forward-email alias list --domain example.com --search info
```

**Flags:**
- `--domain string`: Domain name (required)
- `--enabled bool`: Filter by enabled status
- `--search string`: Search alias names
- `--page int`: Page number
- `--limit int`: Items per page

#### `alias get`

Get detailed alias information.

```bash
forward-email alias get <alias-id> --domain <domain>

# Get alias details
forward-email alias get 507f1f77bcf86cd799439011 --domain example.com
```

#### `alias create`

Create new alias.

```bash
forward-email alias create <name> --domain <domain> --recipients <emails> [flags]

# Create simple alias
forward-email alias create info --domain example.com --recipients team@company.com

# Create with multiple recipients
forward-email alias create support --domain example.com --recipients support@company.com,backup@company.com

# Create with description
forward-email alias create sales --domain example.com --recipients sales@company.com --description "Sales inquiries"
```

**Flags:**
- `--domain string`: Domain name (required)
- `--recipients strings`: Recipient email addresses (required)
- `--description string`: Alias description
- `--labels strings`: Alias labels

#### `alias update`

Update alias settings.

```bash
forward-email alias update <alias-id> --domain <domain> [flags]

# Update description
forward-email alias update 507f1f77bcf86cd799439011 --domain example.com --description "Updated description"

# Update labels
forward-email alias update 507f1f77bcf86cd799439011 --domain example.com --labels support,urgent
```

#### `alias delete`

Delete alias.

```bash
forward-email alias delete <alias-id> --domain <domain> [flags]

# Delete with confirmation
forward-email alias delete 507f1f77bcf86cd799439011 --domain example.com --force
```

#### `alias enable` / `alias disable`

Enable or disable alias.

```bash
# Enable alias
forward-email alias enable <alias-id> --domain <domain>

# Disable alias  
forward-email alias disable <alias-id> --domain <domain>
```

#### `alias recipients`

Update alias recipients.

```bash
forward-email alias recipients <alias-id> --domain <domain> --recipients <emails>

# Update recipients
forward-email alias recipients 507f1f77bcf86cd799439011 --domain example.com --recipients new@company.com,backup@company.com
```

#### `alias password`

Generate IMAP password for alias.

```bash
forward-email alias password <alias-id> --domain <domain>

# Generate new IMAP password
forward-email alias password 507f1f77bcf86cd799439011 --domain example.com
```

#### `alias quota` / `alias stats`

Show alias quota and statistics.

```bash
# Show quota
forward-email alias quota <alias-id> --domain <domain>

# Show statistics
forward-email alias stats <alias-id> --domain <domain>
```

## Email Management Commands

### `forward-email email`

Send and manage emails.

#### `email send`

Send emails with interactive or command-line composition.

```bash
# Interactive email composition
forward-email email send

# Send with command-line flags
forward-email email send --from info@example.com --to user@example.com --subject "Welcome" --body "Hello!"

# Send with file attachment
forward-email email send --from info@example.com --to user@example.com --subject "Report" --attachments report.pdf

# Send HTML email
forward-email email send --from info@example.com --to user@example.com --subject "Newsletter" --html newsletter.html
```

**Flags:**
- `--from string`: Sender email address
- `--to strings`: Recipient email addresses
- `--cc strings`: CC recipients
- `--bcc strings`: BCC recipients
- `--subject string`: Email subject
- `--body string`: Plain text body
- `--html string`: HTML body (file path or content)
- `--attachments strings`: File attachments
- `--dry-run`: Preview email without sending

#### `email list`

List sent emails with filtering.

```bash
forward-email email list [flags]

# List recent emails
forward-email email list

# Filter by date range
forward-email email list --after 2024-01-01 --before 2024-12-31

# Search by subject
forward-email email list --search "Welcome"
```

**Flags:**
- `--after string`: Show emails after date (YYYY-MM-DD)
- `--before string`: Show emails before date (YYYY-MM-DD)
- `--search string`: Search subject and content
- `--page int`: Page number
- `--limit int`: Items per page

#### `email get`

Get detailed email information.

```bash
forward-email email get <email-id>

# Get email details
forward-email email get 507f1f77bcf86cd799439011
```

#### `email delete`

Delete sent email.

```bash
forward-email email delete <email-id> [flags]

# Delete with confirmation
forward-email email delete 507f1f77bcf86cd799439011 --force
```

#### `email quota` / `email stats`

Show email sending quota and statistics.

```bash
# Show sending quota
forward-email email quota

# Show email statistics
forward-email email stats
```

## Debug & Troubleshooting Commands

### `forward-email debug`

Debug utilities for troubleshooting.

#### `debug keys`

Show keyring information for debugging.

```bash
forward-email debug keys [profile]

# Debug current profile keyring
forward-email debug keys

# Debug specific profile
forward-email debug keys production
```

#### `debug auth`

Debug full authentication flow.

```bash
forward-email debug auth [profile]

# Debug current profile auth
forward-email debug auth

# Debug specific profile
forward-email debug auth production
```

#### `debug api`

Test API call with current authentication.

```bash
forward-email debug api [profile]

# Test API connectivity
forward-email debug api

# Test specific profile
forward-email debug api production
```

## Examples

### Complete Workflow Example

```bash
# Setup
forward-email auth login
forward-email domain create example.com
forward-email domain verify example.com

# Create aliases
forward-email alias create info --domain example.com --recipients team@company.com
forward-email alias create support --domain example.com --recipients support@company.com

# Send welcome email
forward-email email send --from info@example.com --to newuser@example.com --subject "Welcome!" --body "Welcome to our service!"

# Check quota
forward-email email quota
```

### Multi-Environment Setup

```bash
# Setup production
forward-email profile create production
forward-email profile switch production
forward-email auth login

# Setup staging
forward-email profile create staging  
forward-email profile switch staging
forward-email auth login

# Use different environments
forward-email profile switch production
forward-email domain list

forward-email profile switch staging
forward-email domain list
```

For more examples and troubleshooting, see the [Quick Start Guide](quick-start.md) and [Troubleshooting Guide](troubleshooting.md).