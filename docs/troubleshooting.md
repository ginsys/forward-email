# Troubleshooting Guide

Solutions to common issues and problems with Forward Email CLI.

## Quick Diagnostic Commands

Run these commands to quickly diagnose common issues:

```bash
# Check authentication status
forward-email auth status

# Test API connectivity
forward-email debug api

# Check keyring access
forward-email debug keys

# Verify current configuration
forward-email profile show
```

## Authentication Issues

### "Authentication failed" or "Invalid API key"

#### Symptoms
- Login fails with "authentication failed"
- Commands return "invalid API key" errors
- API calls return 401 Unauthorized

#### Solutions

1. **Verify your API key**:
   ```bash
   # Check if API key is set correctly
   forward-email auth verify
   
   # Re-authenticate
   forward-email auth logout
   forward-email auth login
   ```

2. **Check API key in Forward Email dashboard**:
   - Log in to [Forward Email](https://forwardemail.net/)
   - Go to Settings → API
   - Verify your API key is active and has correct permissions

3. **Clear and reset credentials**:
   ```bash
   # Clear all stored credentials
   forward-email auth logout --all
   
   # Re-login with correct API key
   forward-email auth login
   ```

### "Keyring access denied" or "Failed to access keyring"

#### Symptoms
- Cannot store credentials in keyring
- Keyring prompts fail or timeout
- Falls back to environment variables

#### Solutions

1. **Linux**: Install and configure keyring service
   ```bash
   # Ubuntu/Debian
   sudo apt install gnome-keyring
   
   # Fedora/RHEL
   sudo dnf install gnome-keyring
   
   # Start keyring service
   gnome-keyring-daemon --start
   ```

2. **macOS**: Check Keychain Access
   - Open Keychain Access app
   - Verify you can access "login" keychain
   - Unlock keychain if locked

3. **Windows**: Check Credential Manager
   - Open Control Panel → Credential Manager
   - Verify you can access Windows Credentials

4. **Fallback to environment variables**:
   ```bash
   export FORWARDEMAIL_API_KEY="your-api-key"
   forward-email domain list
   ```

### "Profile not found" or "No active profile"

#### Symptoms
- Commands fail with "profile not found"
- No current profile set

#### Solutions

1. **List and create profiles**:
   ```bash
   # List existing profiles
   forward-email profile list
   
   # Create default profile if missing
   forward-email profile create default
   
   # Switch to profile
   forward-email profile switch default
   ```

2. **Reset configuration**:
   ```bash
   # Create new default profile
   forward-email profile create default
   forward-email auth login
   ```

## Network and API Issues

### "Connection timeout" or "Network unreachable"

#### Symptoms
- Commands hang or timeout
- "connection refused" errors
- Network timeouts

#### Solutions

1. **Check internet connectivity**:
   ```bash
   # Test basic connectivity
   ping api.forwardemail.net
   curl -I https://api.forwardemail.net
   ```

2. **Increase timeout**:
   ```bash
   # Use longer timeout
   forward-email domain list --timeout 60s
   
   # Set default timeout in profile
   forward-email profile create slow-network --timeout 120s
   ```

3. **Check proxy settings**:
   ```bash
   # Set proxy if needed
   export HTTP_PROXY=http://proxy.company.com:8080
   export HTTPS_PROXY=http://proxy.company.com:8080
   ```

### "Rate limit exceeded" or "Too many requests"

#### Symptoms
- Commands fail with "rate limit" errors
- HTTP 429 responses
- Temporary API blocks

#### Solutions

1. **Wait and retry**:
   ```bash
   # Wait a few minutes and retry
   sleep 300  # 5 minutes
   forward-email domain list
   ```

2. **Reduce request frequency**:
   ```bash
   # Add delays between commands
   forward-email alias list --domain example.com
   sleep 5
   forward-email alias list --domain other.com
   ```

3. **Check quota usage**:
   ```bash
   forward-email email quota
   forward-email domain quota example.com
   ```

### "Invalid JSON response" or "Unexpected response"

#### Symptoms
- Malformed JSON errors
- Unexpected API responses
- Parsing failures

#### Solutions

1. **Enable debug mode**:
   ```bash
   # See raw API responses
   forward-email domain list --debug
   ```

2. **Check API status**:
   ```bash
   # Test basic API connectivity
   forward-email debug api
   
   # Check Forward Email status page
   curl -s https://status.forwardemail.net/api/v1/status
   ```

3. **Try different output format**:
   ```bash
   # Use table format instead of JSON
   forward-email domain list --output table
   ```

## Command-Specific Issues

### Domain Commands

#### "Domain not found"
```bash
# Verify domain name spelling
forward-email domain list | grep -i "example"

# Use domain ID instead of name
forward-email domain get 507f1f77bcf86cd799439011
```

#### "Domain verification failed"
```bash
# Check DNS records manually
dig MX example.com
dig TXT example.com

# Get required DNS records
forward-email domain dns example.com

# Wait for DNS propagation (up to 48 hours)
```

### Alias Commands

#### "Alias creation failed"
```bash
# Check domain ownership
forward-email domain get example.com

# Verify alias name format
forward-email alias create support --domain example.com --recipients support@company.com

# Check domain verification status
forward-email domain verify example.com
```

#### "Recipients validation failed"
```bash
# Use valid email format
forward-email alias create info --domain example.com --recipients team@company.com

# Multiple recipients (comma-separated)
forward-email alias create support --domain example.com --recipients primary@company.com,backup@company.com
```

### Email Commands

#### "Email sending failed"
```bash
# Check sending quota
forward-email email quota

# Verify sender address
forward-email alias list --domain example.com

# Use interactive mode
forward-email email send
```

#### "Attachment too large"
```bash
# Check file size
ls -lh attachment.pdf

# Forward Email has attachment size limits
# Compress or use cloud storage links instead
```

## Performance Issues

### Slow Command Execution

#### Symptoms
- Commands take longer than expected
- Frequent timeouts
- Poor responsiveness

#### Solutions

1. **Use caching output**:
   ```bash
   # Cache domain list
   forward-email domain list --output json > domains.json
   
   # Work with cached data
   cat domains.json | jq '.[] | select(.verified == true)'
   ```

2. **Optimize queries**:
   ```bash
   # Use specific filters
   forward-email domain list --verified true --limit 10
   
   # Paginate large results
   forward-email alias list --domain example.com --page 1 --limit 25
   ```

3. **Use appropriate output format**:
   ```bash
   # JSON is faster for large datasets
   forward-email domain list --output json
   
   # Table format is slower but more readable
   forward-email domain list --output table
   ```

## Output and Formatting Issues

### "Invalid output format" or "Formatting errors"

#### Symptoms
- Garbled output
- Missing columns
- JSON parsing errors

#### Solutions

1. **Try different output formats**:
   ```bash
   # If JSON fails, try table
   forward-email domain list --output table
   
   # If table is garbled, try JSON
   forward-email domain list --output json
   ```

2. **Check terminal width**:
   ```bash
   # Set wider terminal
   export COLUMNS=120
   forward-email domain list
   ```

3. **Use specific columns**:
   ```bash
   # Limit output fields (when available)
   forward-email domain list --output json | jq '.[] | {name, verified}'
   ```

### "Special characters not displaying"

#### Symptoms
- Unicode characters show as boxes
- Emoji not rendering
- Encoding issues

#### Solutions

1. **Check terminal encoding**:
   ```bash
   # Set UTF-8 encoding
   export LANG=en_US.UTF-8
   export LC_ALL=en_US.UTF-8
   ```

2. **Use ASCII-safe output**:
   ```bash
   # JSON output is ASCII-safe
   forward-email domain list --output json
   ```

## Configuration Issues

### "Config file not found" or "Permission denied"

#### Symptoms
- Cannot read/write config file
- Permission errors
- Config not persisting

#### Solutions

1. **Check config directory permissions**:
   ```bash
   # Linux/macOS
   ls -la ~/.config/forwardemail/
   chmod 755 ~/.config/forwardemail/
   chmod 644 ~/.config/forwardemail/config.yaml
   
   # Create directory if missing
   mkdir -p ~/.config/forwardemail/
   ```

2. **Manually create config**:
   ```bash
   # Create minimal config
   cat > ~/.config/forwardemail/config.yaml << EOF
   current_profile: default
   profiles:
     default:
       base_url: "https://api.forwardemail.net"
       timeout: "30s"
       output: "table"
   EOF
   ```

### "Invalid configuration format"

#### Symptoms
- YAML parsing errors
- Config validation fails
- Unexpected behavior

#### Solutions

1. **Validate YAML syntax**:
   ```bash
   # Check YAML syntax
   python -c "import yaml; yaml.safe_load(open('~/.config/forwardemail/config.yaml'))"
   ```

2. **Reset configuration**:
   ```bash
   # Backup current config
   cp ~/.config/forwardemail/config.yaml ~/.config/forwardemail/config.yaml.backup
   
   # Create fresh config
   forward-email profile create default
   ```

## Building and Installation Issues

### "Go build failed" or "Dependency errors"

#### Symptoms
- Build compilation errors
- Missing dependencies
- Version conflicts

#### Solutions

1. **Check Go version**:
   ```bash
   # Requires Go 1.24+
   go version
   
   # Update Go if needed
   # https://golang.org/doc/install
   ```

2. **Clean and rebuild**:
   ```bash
   # Clean module cache
   go clean -modcache
   
   # Download dependencies
   go mod download
   
   # Rebuild
   go build -o bin/forward-email ./cmd/forward-email
   ```

3. **Check dependencies**:
   ```bash
   # Verify modules
   go mod verify
   
   # Update dependencies
   go mod tidy
   ```

### "Binary not found" or "Command not found"

#### Symptoms
- `forward-email: command not found`
- Binary not in PATH
- Installation issues

#### Solutions

1. **Check binary location**:
   ```bash
   # Find binary
   find . -name "forward-email" -type f
   
   # Make executable
   chmod +x bin/forward-email
   
   # Test directly
   ./bin/forward-email --help
   ```

2. **Add to PATH**:
   ```bash
   # Add to PATH temporarily
   export PATH="$PWD/bin:$PATH"
   
   # Add to shell profile permanently
   echo 'export PATH="$PWD/bin:$PATH"' >> ~/.bashrc
   source ~/.bashrc
   ```

3. **Install globally**:
   ```bash
   # Copy to system location
   sudo cp bin/forward-email /usr/local/bin/
   
   # Or use go install
   go install ./cmd/forward-email
   ```

## Getting Help

### Debug Information

Collect debug information for support:

```bash
# System information
uname -a
go version

# CLI version and build info
forward-email version

# Configuration status
forward-email profile show
forward-email auth status

# Debug output
forward-email debug auth
forward-email debug api
forward-email debug keys
```

### Enable Verbose Logging

```bash
# Enable debug mode
export FORWARDEMAIL_DEBUG=true

# Run problematic command
forward-email domain list --debug

# Check specific authentication flow
forward-email debug auth --debug
```

### Common Debug Commands

```bash
# Test network connectivity
curl -v https://api.forwardemail.net/v1/domains

# Check DNS resolution
nslookup api.forwardemail.net

# Verify SSL certificates
openssl s_client -connect api.forwardemail.net:443

# Test with minimal configuration
FORWARDEMAIL_API_KEY="your-key" forward-email domain list
```

### Reporting Issues

When reporting issues, include:

1. **CLI version**: `forward-email version`
2. **Operating system**: `uname -a`
3. **Go version**: `go version`
4. **Command that failed**: Exact command and flags used
5. **Error output**: Complete error message
6. **Debug output**: Output with `--debug` flag
7. **Configuration**: Sanitized config (remove API keys)

### Additional Resources

- **GitHub Issues**: [https://github.com/ginsys/forward-email/issues](https://github.com/ginsys/forward-email/issues)
- **Forward Email API Status**: [https://status.forwardemail.net](https://status.forwardemail.net)
- **Forward Email Documentation**: [https://forwardemail.net/en/docs](https://forwardemail.net/en/docs)

## FAQ

### Q: Why is my API key not working?
A: Check that your API key is active in your Forward Email dashboard and has the necessary permissions for the operations you're trying to perform.

### Q: Can I use the CLI without storing credentials?
A: Yes, use environment variables: `export FORWARDEMAIL_API_KEY="your-key"`

### Q: How do I automate CLI usage in scripts?
A: Use JSON output format and environment variables for authentication:
```bash
export FORWARDEMAIL_API_KEY="your-key"
export FORWARDEMAIL_OUTPUT="json"
forward-email domain list | jq '.[] | .name'
```

### Q: Why are my commands slow?
A: Try using JSON output format, enable caching, or increase timeout values. The CLI may be slower on first run due to keyring access.

### Q: Can I use multiple Forward Email accounts?
A: Yes, create separate profiles for each account using `forward-email profile create`.

---

*Last Updated: 2025-08-27 | Comprehensive troubleshooting guide for current implementation*

---

Docs navigation: [Prev: Configuration](configuration.md) | [Next: Docs Index](README.md)
