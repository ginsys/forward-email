# Forward Email CLI Architecture

## Overview

This CLI follows a modular architecture with centralized authentication and consistent patterns across all commands.

## Key Principles

### 1. Centralized Authentication

**Problem**: Initially, each command duplicated authentication logic, creating maintenance overhead and potential inconsistencies.

**Solution**: All API client creation is now centralized in `internal/client/client.go`.

**Usage Pattern**:
```go
import "github.com/ginsys/forwardemail-cli/internal/client"

func runSomeCommand(cmd *cobra.Command, args []string) error {
    apiClient, err := client.NewAPIClient()
    if err != nil {
        return err
    }
    
    // Use apiClient for API calls
    response, err := apiClient.SomeService.SomeMethod(ctx, params)
    if err != nil {
        return fmt.Errorf("operation failed: %w", err)
    }
    
    return nil
}
```

### 2. Profile Management

The CLI supports multiple profiles for different accounts/environments:

- **Current Profile**: Stored in config file (`~/.config/forwardemail/config.yaml`)
- **Profile Storage**: API keys stored securely in OS keyring, settings in config file
- **Profile Selection**: Via `--profile` flag or config file `current_profile` setting
- **Fallback Chain**: Flag → config current_profile → "default"

### 3. Output Formatting

Consistent output formatting across all commands:

- **Supported Formats**: table (default), json, yaml, csv
- **Table Formatting**: Uses `github.com/olekukonko/tablewriter`
- **Type System**: `output.TableData` struct for tabular data
- **Format Selection**: Via `--output` flag

**Usage Pattern**:
```go
import "github.com/ginsys/forwardemail-cli/pkg/output"

return formatOutput(data, outputFormat, func(format output.Format) (interface{}, error) {
    if format == output.FormatTable || format == output.FormatCSV {
        return output.FormatSomeData(data, format)
    }
    return data, nil
})
```

## Directory Structure

```
├── cmd/forward-email/          # Main CLI entry point
├── internal/
│   ├── client/                 # Centralized API client creation
│   ├── cmd/                    # Command implementations
│   └── keyring/                # OS keyring integration
├── pkg/
│   ├── api/                    # API service layer
│   ├── auth/                   # Authentication provider
│   ├── config/                 # Configuration management
│   └── output/                 # Output formatting
```

## Adding New Commands

When adding new commands that need API access:

1. **Import the centralized client**:
   ```go
   import "github.com/ginsys/forwardemail-cli/internal/client"
   ```

2. **Use the standard pattern**:
   ```go
   func runYourCommand(cmd *cobra.Command, args []string) error {
       apiClient, err := client.NewAPIClient()
       if err != nil {
           return err
       }
       
       // Your API calls here
       
       return nil
   }
   ```

3. **Add output formatting if needed**:
   ```go
   return formatOutput(data, outputFormat, func(format output.Format) (interface{}, error) {
       if format == output.FormatTable || format == output.FormatCSV {
           return output.FormatYourData(data, format)
       }
       return data, nil
   })
   ```

## Authentication Flow

1. **Profile Resolution**: 
   - Check `--profile` flag
   - Fall back to config `current_profile`
   - Final fallback to "default"

2. **Credential Loading**:
   - Try OS keyring for API key
   - Fall back to config file
   - Environment variables override both

3. **Provider Creation**:
   - Initialize keyring (graceful degradation if unavailable)
   - Create auth provider with config, keyring, and profile
   - Create API client with base URL and auth provider

## Configuration Files

### Main Config: `~/.config/forwardemail/config.yaml`

```yaml
current_profile: ginsys
profiles:
  default:
    base_url: https://api.forwardemail.net
    api_key: ""          # Usually stored in keyring instead
    timeout: 30s
    output: table
  ginsys:
    base_url: https://api.forwardemail.net
    api_key: ""
    timeout: 30s
    output: table
```

### OS Keyring Storage

API keys are stored securely in the OS keyring with the pattern:
- **Service**: `forwardemail-cli`
- **Account**: `{profile}` (e.g., "ginsys", "default")

## Error Handling

Consistent error handling patterns:

```go
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}
```

API errors are automatically wrapped with context by the client layer.

## Testing Considerations

- Mock the `client.NewAPIClient()` function for unit tests
- Test profile switching and authentication flows
- Validate output formatting for all supported formats
- Test error handling and graceful degradation

## Profile Management Commands

The CLI now includes comprehensive profile management:

### Available Commands

```bash
# List all profiles
forward-email profile list [--output table|json|yaml]

# Show profile details  
forward-email profile show [profile-name]

# Switch current profile
forward-email profile switch <profile-name>

# Create new profile
forward-email profile create <profile-name>

# Delete profile (with confirmation)
forward-email profile delete <profile-name> [--force]
```

### Examples

```bash
# View all profiles with their status
$ forward-email profile list
┌─────────┬─────────┬──────────────────────────────┬─────────────┬────────┬─────────┐
│ PROFILE │ CURRENT │           BASE URL           │ HAS API KEY │ OUTPUT │ TIMEOUT │
├─────────┼─────────┼──────────────────────────────┼─────────────┼────────┼─────────┤
│ ginsys  │ ✓       │ https://api.forwardemail.net │ ✓ (keyring) │ table  │ 30s     │
│ default │         │ https://api.forwardemail.net │ ✗           │ table  │ 30s     │
└─────────┴─────────┴──────────────────────────────┴─────────────┴────────┴─────────┘

# Switch to different profile
$ forward-email profile switch work
Switched to profile 'work'

# Create new profile for development
$ forward-email profile create dev
Profile 'dev' created successfully
Use 'forward-email auth login --profile dev' to add API credentials
```

## Future Improvements

1. **Configuration Validation**: Add config file validation and repair capabilities  
2. **Caching**: Implement response caching for frequently accessed data
3. **Batch Operations**: Add support for bulk operations where the API supports them
4. **Profile Templates**: Add predefined profile templates for common configurations