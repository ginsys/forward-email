# CLI Command Implementation Guide

The `internal/cmd` package contains all CLI command implementations for the Forward Email CLI. This package provides the user interface layer that bridges between the Cobra CLI framework and the API service layer.

## Overview

This package implements the command-line interface using the following patterns:

- **Consistent Command Structure** across all operations
- **Centralized Authentication** through the client wrapper
- **Standardized Output Formatting** for all data types
- **Comprehensive Error Handling** with user-friendly messages
- **Modular Design** with one file per command group

## Package Structure

```
internal/cmd/
├── root.go         # Root command and global flags
├── auth.go         # Authentication commands (login, verify, status, logout)
├── profile.go      # Profile management commands (list, show, create, switch, delete)
├── domain.go       # Domain management commands (list, get, create, update, delete, verify)
├── alias.go        # Alias management commands (list, get, create, update, delete, enable, disable)
├── email.go        # Email operations commands (send, list, get, delete, quota, stats)
├── debug.go        # Debug and troubleshooting commands (keys, auth, api)
└── README.md       # This documentation
```

## Command Implementation Patterns

### Basic Command Structure

All commands follow a consistent structure:

```go
func NewDomainCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "domain",
        Short: "Manage Forward Email domains",
        Long:  `Manage Forward Email domains with CRUD operations and DNS verification.`,
    }
    
    // Add subcommands
    cmd.AddCommand(NewDomainListCommand())
    cmd.AddCommand(NewDomainGetCommand())
    cmd.AddCommand(NewDomainCreateCommand())
    // ... more subcommands
    
    return cmd
}

func NewDomainListCommand() *cobra.Command {
    var options struct {
        Page     int
        Limit    int
        Search   string
        Verified *bool
        Plan     string
        Sort     string
        Order    string
        Output   string
    }
    
    cmd := &cobra.Command{
        Use:   "list",
        Short: "List all domains",
        Long:  `List all domains with filtering, pagination, and sorting options.`,
        Example: `  forward-email domain list
  forward-email domain list --verified true
  forward-email domain list --search example --output json`,
        RunE: func(cmd *cobra.Command, args []string) error {
            return runDomainListCommand(cmd, args, options)
        },
    }
    
    // Add flags
    cmd.Flags().IntVar(&options.Page, "page", 1, "Page number")
    cmd.Flags().IntVar(&options.Limit, "limit", 25, "Items per page")
    cmd.Flags().StringVar(&options.Search, "search", "", "Search domain names")
    cmd.Flags().BoolVar(options.Verified, "verified", nil, "Filter by verification status")
    cmd.Flags().StringVar(&options.Plan, "plan", "", "Filter by plan (free|enhanced|team)")
    cmd.Flags().StringVar(&options.Sort, "sort", "", "Sort by field (name|created|updated)")
    cmd.Flags().StringVar(&options.Order, "order", "", "Sort order (asc|desc)")
    
    return cmd
}
```

### Command Execution Pattern

All command execution follows this pattern:

```go
func runDomainListCommand(cmd *cobra.Command, args []string, options domainListOptions) error {
    // 1. Get API client through centralized client wrapper
    apiClient, err := client.NewAPIClient()
    if err != nil {
        return fmt.Errorf("failed to create API client: %w", err)
    }
    
    // 2. Build API request from command options
    listOptions := api.DomainListOptions{
        Page:     options.Page,
        Limit:    options.Limit,
        Search:   options.Search,
        Verified: options.Verified,
        Plan:     options.Plan,
        Sort:     options.Sort,
        Order:    options.Order,
    }
    
    // 3. Execute API call with context
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    result, err := apiClient.Domains.List(ctx, listOptions)
    if err != nil {
        return fmt.Errorf("failed to list domains: %w", err)
    }
    
    // 4. Format and display output
    outputFormat, _ := cmd.Flags().GetString("output")
    return formatOutput(cmd, result.Data, outputFormat, func(format output.Format) (interface{}, error) {
        if format == output.FormatTable || format == output.FormatCSV {
            return output.FormatDomains(result.Data, format)
        }
        return result.Data, nil
    })
}
```

## Authentication Commands

### Auth Command Structure

The auth commands handle API key management and validation:

```go
// internal/cmd/auth.go
func NewAuthCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "auth",
        Short: "Manage authentication credentials",
        Long:  `Manage API keys and authentication for Forward Email.`,
    }
    
    cmd.AddCommand(NewAuthLoginCommand())
    cmd.AddCommand(NewAuthVerifyCommand()) 
    cmd.AddCommand(NewAuthStatusCommand())
    cmd.AddCommand(NewAuthLogoutCommand())
    
    return cmd
}
```

### Interactive Login Implementation

```go
func runAuthLoginCommand(cmd *cobra.Command, args []string, options authLoginOptions) error {
    // Get config and keyring
    cfg, err := config.Load()
    if err != nil {
        return fmt.Errorf("failed to load config: %w", err)
    }
    
    kr, err := keyring.GetKeyring()
    if err != nil {
        return fmt.Errorf("failed to access keyring: %w", err)
    }
    
    // Determine profile
    profile := options.Profile
    if profile == "" {
        profile = cfg.CurrentProfile
        if profile == "" {
            profile = "default"
        }
    }
    
    // Get API key securely
    fmt.Printf("Enter your Forward Email API key for profile '%s': ", profile)
    apiKeyBytes, err := term.ReadPassword(int(syscall.Stdin))
    if err != nil {
        return fmt.Errorf("failed to read API key: %w", err)
    }
    fmt.Println() // New line after password input
    
    apiKey := strings.TrimSpace(string(apiKeyBytes))
    if apiKey == "" {
        return errors.New("API key cannot be empty")
    }
    
    // Validate API key
    authProvider := &auth.AuthProvider{
        Config:  cfg,
        Keyring: kr,
        Profile: profile,
    }
    
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := authProvider.ValidateAPIKey(ctx, apiKey); err != nil {
        return fmt.Errorf("API key validation failed: %w", err)
    }
    
    // Store in keyring
    if err := kr.Set("forward-email", profile, apiKey); err != nil {
        return fmt.Errorf("failed to store API key: %w", err)
    }
    
    fmt.Printf("✓ Login successful for profile '%s'\n", profile)
    return nil
}
```

## Profile Management Commands

### Profile Operations

Profile commands manage multiple environment configurations:

```go
func runProfileListCommand(cmd *cobra.Command, args []string, options profileListOptions) error {
    cfg, err := config.Load()
    if err != nil {
        return fmt.Errorf("failed to load config: %w", err)
    }
    
    kr, err := keyring.GetKeyring()
    if err != nil {
        // Keyring access is optional for listing
        kr = nil
    }
    
    // Build profile info
    var profiles []ProfileInfo
    for name, profile := range cfg.Profiles {
        info := ProfileInfo{
            Name:     name,
            Current:  name == cfg.CurrentProfile,
            BaseURL:  profile.BaseURL,
            Output:   profile.Output,
            Timeout:  profile.Timeout,
        }
        
        // Check for API key
        if kr != nil {
            if _, err := kr.Get("forward-email", name); err == nil {
                info.HasAPIKey = true
                info.APIKeySource = "keyring"
            }
        }
        
        // Check environment variable
        envKey := fmt.Sprintf("FORWARDEMAIL_%s_API_KEY", strings.ToUpper(name))
        if os.Getenv(envKey) != "" {
            info.HasAPIKey = true
            info.APIKeySource = "environment"
        }
        
        profiles = append(profiles, info)
    }
    
    // Sort profiles
    sort.Slice(profiles, func(i, j int) bool {
        // Current profile first, then alphabetical
        if profiles[i].Current != profiles[j].Current {
            return profiles[i].Current
        }
        return profiles[i].Name < profiles[j].Name
    })
    
    // Format output
    outputFormat, _ := cmd.Flags().GetString("output")
    return formatOutput(cmd, profiles, outputFormat, func(format output.Format) (interface{}, error) {
        if format == output.FormatTable || format == output.FormatCSV {
            return output.FormatProfiles(profiles, format)
        }
        return profiles, nil
    })
}
```

## Domain Management Commands

### Domain CRUD Operations

Domain commands provide complete lifecycle management:

```go
func runDomainCreateCommand(cmd *cobra.Command, args []string, options domainCreateOptions) error {
    if len(args) != 1 {
        return errors.New("domain name is required")
    }
    domainName := args[0]
    
    // Validate domain name
    if err := validateDomainName(domainName); err != nil {
        return fmt.Errorf("invalid domain name: %w", err)
    }
    
    // Get API client
    apiClient, err := client.NewAPIClient()
    if err != nil {
        return fmt.Errorf("failed to create API client: %w", err)
    }
    
    // Create domain
    createReq := api.DomainCreateRequest{
        Name: domainName,
        Plan: options.Plan,
    }
    
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    domain, err := apiClient.Domains.Create(ctx, createReq)
    if err != nil {
        return fmt.Errorf("failed to create domain: %w", err)
    }
    
    // Display result
    fmt.Printf("✓ Created domain '%s' with ID %s\n", domain.Name, domain.ID)
    
    // Show DNS setup information
    fmt.Println("\nNext steps:")
    fmt.Printf("1. Add MX record: %s\n", domain.MXRecord)
    fmt.Printf("2. Add TXT record: %s\n", domain.TXTRecord)
    fmt.Printf("3. Run 'forward-email domain verify %s' to verify DNS setup\n", domain.Name)
    
    return nil
}
```

### Domain Verification

```go
func runDomainVerifyCommand(cmd *cobra.Command, args []string, options domainVerifyOptions) error {
    if len(args) != 1 {
        return errors.New("domain name or ID is required")
    }
    domainID := args[0]
    
    apiClient, err := client.NewAPIClient()
    if err != nil {
        return fmt.Errorf("failed to create API client: %w", err)
    }
    
    ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
    defer cancel()
    
    // Run verification
    verification, err := apiClient.Domains.Verify(ctx, domainID)
    if err != nil {
        return fmt.Errorf("domain verification failed: %w", err)
    }
    
    // Display results
    fmt.Printf("Domain verification results for '%s':\n\n", verification.DomainName)
    
    fmt.Printf("MX Record: %s\n", formatVerificationStatus(verification.MXRecordValid))
    if !verification.MXRecordValid {
        fmt.Printf("  Expected: %s\n", verification.ExpectedMXRecord)
        fmt.Printf("  Found: %s\n", verification.ActualMXRecord)
    }
    
    fmt.Printf("TXT Record: %s\n", formatVerificationStatus(verification.TXTRecordValid))
    if !verification.TXTRecordValid {
        fmt.Printf("  Expected: %s\n", verification.ExpectedTXTRecord)
        fmt.Printf("  Found: %s\n", verification.ActualTXTRecord)
    }
    
    if verification.MXRecordValid && verification.TXTRecordValid {
        fmt.Println("\n✓ Domain is fully verified and ready to use!")
    } else {
        fmt.Println("\n⚠ Domain verification incomplete. Please update DNS records and try again.")
        fmt.Println("Note: DNS propagation can take up to 48 hours.")
    }
    
    return nil
}

func formatVerificationStatus(valid bool) string {
    if valid {
        return "✓ Valid"
    }
    return "✗ Invalid"
}
```

## Alias Management Commands

### Alias Operations

Alias commands handle email forwarding configuration:

```go
func runAliasCreateCommand(cmd *cobra.Command, args []string, options aliasCreateOptions) error {
    if len(args) != 1 {
        return errors.New("alias name is required")
    }
    aliasName := args[0]
    
    if options.Domain == "" {
        return errors.New("--domain flag is required")
    }
    
    if len(options.Recipients) == 0 {
        return errors.New("--recipients flag is required")
    }
    
    // Validate recipients
    for _, recipient := range options.Recipients {
        if err := validateEmailAddress(recipient); err != nil {
            return fmt.Errorf("invalid recipient '%s': %w", recipient, err)
        }
    }
    
    // Get API client
    apiClient, err := client.NewAPIClient()
    if err != nil {
        return fmt.Errorf("failed to create API client: %w", err)
    }
    
    // Create alias
    createReq := api.AliasCreateRequest{
        Name:        aliasName,
        Description: options.Description,
        Recipients:  options.Recipients,
        Labels:      options.Labels,
    }
    
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    alias, err := apiClient.Aliases.Create(ctx, options.Domain, createReq)
    if err != nil {
        return fmt.Errorf("failed to create alias: %w", err)
    }
    
    // Display result
    fmt.Printf("✓ Created alias '%s@%s' with ID %s\n", alias.Name, options.Domain, alias.ID)
    fmt.Printf("Recipients: %s\n", strings.Join(alias.Recipients, ", "))
    
    if alias.Description != "" {
        fmt.Printf("Description: %s\n", alias.Description)
    }
    
    return nil
}
```

## Email Operations Commands

### Email Sending

Email commands support both interactive and command-line composition:

```go
func runEmailSendCommand(cmd *cobra.Command, args []string, options emailSendOptions) error {
    // Check if interactive mode (no flags provided)
    if options.From == "" && len(options.To) == 0 && options.Subject == "" {
        return runInteractiveEmailSend(cmd, options)
    }
    
    // Command-line mode
    return runDirectEmailSend(cmd, options)
}

func runInteractiveEmailSend(cmd *cobra.Command, options emailSendOptions) error {
    var req api.EmailSendRequest
    
    // Interactive prompts
    fmt.Print("From address: ")
    if _, err := fmt.Scanln(&req.From); err != nil {
        return fmt.Errorf("failed to read from address: %w", err)
    }
    
    fmt.Print("To addresses (comma-separated): ")
    var toStr string
    if _, err := fmt.Scanln(&toStr); err != nil {
        return fmt.Errorf("failed to read to addresses: %w", err)
    }
    req.To = strings.Split(toStr, ",")
    for i, addr := range req.To {
        req.To[i] = strings.TrimSpace(addr)
    }
    
    fmt.Print("Subject: ")
    if _, err := fmt.Scanln(&req.Subject); err != nil {
        return fmt.Errorf("failed to read subject: %w", err)
    }
    
    fmt.Println("Message body (enter empty line to finish):")
    var bodyLines []string
    scanner := bufio.NewScanner(os.Stdin)
    for scanner.Scan() {
        line := scanner.Text()
        if line == "" {
            break
        }
        bodyLines = append(bodyLines, line)
    }
    req.Text = strings.Join(bodyLines, "\n")
    
    // Confirmation
    fmt.Println("\nEmail preview:")
    fmt.Printf("From: %s\n", req.From)
    fmt.Printf("To: %s\n", strings.Join(req.To, ", "))
    fmt.Printf("Subject: %s\n", req.Subject)
    fmt.Printf("Body: %s\n", req.Text)
    
    fmt.Print("\nSend this email? (y/N): ")
    var confirm string
    fmt.Scanln(&confirm)
    
    if strings.ToLower(confirm) != "y" && strings.ToLower(confirm) != "yes" {
        fmt.Println("Email sending cancelled")
        return nil
    }
    
    return sendEmail(req)
}

func runDirectEmailSend(cmd *cobra.Command, options emailSendOptions) error {
    // Validate required fields
    if options.From == "" {
        return errors.New("--from flag is required")
    }
    if len(options.To) == 0 {
        return errors.New("--to flag is required")
    }
    if options.Subject == "" {
        return errors.New("--subject flag is required")
    }
    
    // Build request
    req := api.EmailSendRequest{
        From:    options.From,
        To:      options.To,
        Cc:      options.Cc,
        Bcc:     options.Bcc,
        Subject: options.Subject,
        Text:    options.Body,
        Html:    options.HTML,
        Headers: options.Headers,
    }
    
    // Handle attachments
    if len(options.Attachments) > 0 {
        attachments, err := processAttachments(options.Attachments)
        if err != nil {
            return fmt.Errorf("failed to process attachments: %w", err)
        }
        req.Attachments = attachments
    }
    
    // Dry run mode
    if options.DryRun {
        fmt.Println("Dry run mode - email preview:")
        return displayEmailPreview(req)
    }
    
    return sendEmail(req)
}

func sendEmail(req api.EmailSendRequest) error {
    apiClient, err := client.NewAPIClient()
    if err != nil {
        return fmt.Errorf("failed to create API client: %w", err)
    }
    
    ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
    defer cancel()
    
    response, err := apiClient.Emails.Send(ctx, req)
    if err != nil {
        return fmt.Errorf("failed to send email: %w", err)
    }
    
    fmt.Printf("✓ Email sent successfully\n")
    fmt.Printf("Message ID: %s\n", response.MessageID)
    fmt.Printf("Email ID: %s\n", response.ID)
    
    return nil
}
```

## Output Formatting

### Unified Output Function

All commands use a consistent output formatting pattern:

```go
func formatOutput(cmd *cobra.Command, data interface{}, outputFormat string, formatter func(output.Format) (interface{}, error)) error {
    // Determine output format
    format := output.ParseFormat(outputFormat)
    if format == output.FormatUnknown {
        format = output.FormatTable // Default
    }
    
    // Apply formatter
    result, err := formatter(format)
    if err != nil {
        return fmt.Errorf("failed to format output: %w", err)
    }
    
    // Write output
    switch format {
    case output.FormatTable:
        fmt.Print(result)
    case output.FormatJSON:
        encoder := json.NewEncoder(cmd.OutOrStdout())
        encoder.SetIndent("", "  ")
        return encoder.Encode(result)
    case output.FormatYAML:
        yamlData, err := yaml.Marshal(result)
        if err != nil {
            return fmt.Errorf("failed to marshal YAML: %w", err)
        }
        fmt.Print(string(yamlData))
    case output.FormatCSV:
        fmt.Print(result)
    }
    
    return nil
}
```

## Error Handling

### User-Friendly Error Messages

All commands implement consistent error handling:

```go
func handleAPIError(err error) error {
    // Check for specific error types
    if errors.IsNotFound(err) {
        return fmt.Errorf("resource not found - please check the ID or name and try again")
    }
    
    if errors.IsUnauthorized(err) {
        return fmt.Errorf("authentication failed - please run 'forward-email auth login' to authenticate")
    }
    
    if errors.IsRateLimit(err) {
        return fmt.Errorf("rate limit exceeded - please wait a few minutes and try again")
    }
    
    if errors.IsValidation(err) {
        return fmt.Errorf("validation error: %w", err)
    }
    
    // Generic error with suggestion
    return fmt.Errorf("operation failed: %w\n\nTry running 'forward-email debug api' to test connectivity", err)
}
```

## Testing Commands

### Command Testing Pattern

```go
func TestDomainListCommand(t *testing.T) {
    tests := []struct {
        name       string
        args       []string
        flags      map[string]string
        mockSetup  func(*httptest.Server)
        wantOutput string
        wantErr    bool
    }{
        {
            name: "successful list",
            args: []string{"domain", "list"},
            flags: map[string]string{
                "output": "json",
            },
            mockSetup: func(server *httptest.Server) {
                // Setup mock responses
            },
            wantOutput: "example.com",
            wantErr:    false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup mock server
            server := createMockServer()
            defer server.Close()
            
            if tt.mockSetup != nil {
                tt.mockSetup(server)
            }
            
            // Setup command
            cmd := NewRootCommand()
            cmd.SetArgs(tt.args)
            
            // Set flags
            for key, value := range tt.flags {
                cmd.Flags().Set(key, value)
            }
            
            // Capture output
            var output bytes.Buffer
            cmd.SetOut(&output)
            cmd.SetErr(&output)
            
            // Execute
            err := cmd.Execute()
            
            // Verify results
            if (err != nil) != tt.wantErr {
                t.Errorf("Command error = %v, wantErr %v", err, tt.wantErr)
            }
            
            if tt.wantOutput != "" && !strings.Contains(output.String(), tt.wantOutput) {
                t.Errorf("Expected output to contain '%s', got '%s'", tt.wantOutput, output.String())
            }
        })
    }
}
```

## Best Practices

### Command Design

1. **Consistent Naming**: Use verb-noun pattern (list, get, create, update, delete)
2. **Required Arguments**: Use positional arguments for required values
3. **Optional Flags**: Use flags for optional parameters and filters
4. **Output Flexibility**: Support multiple output formats via --output flag
5. **Help Documentation**: Provide clear descriptions and examples

### Error Handling

1. **User-Friendly Messages**: Convert API errors to helpful user messages
2. **Actionable Suggestions**: Include next steps in error messages
3. **Context Preservation**: Wrap errors with relevant context
4. **Graceful Degradation**: Handle missing dependencies gracefully

### Authentication

1. **Centralized Client**: Always use client.NewAPIClient() for consistency
2. **Profile Awareness**: Respect --profile flag across all commands
3. **Error Context**: Provide helpful auth error messages with login suggestions

### Testing

1. **Mock API Responses**: Use httptest.Server for testing API interactions
2. **Command Isolation**: Test commands independently with mocked dependencies
3. **Output Validation**: Verify both successful and error output formats
4. **Cross-Platform**: Ensure commands work across all supported platforms

For more information on the Forward Email CLI development, see:
- [Architecture Overview](../../docs/development/architecture.md)
- [API Integration Guide](../../docs/development/api-integration.md)
- [Testing Strategy](../../docs/development/testing.md)