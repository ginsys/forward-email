# Domain Alias Synchronization Feature Specification

## Overview

The domain alias synchronization feature provides bulk operations to synchronize alias configurations across multiple domains in a Forward Email account. This feature addresses the need to maintain consistent alias setups across related domains (e.g., primary domain and alternate TLDs, development/staging/production domains).

## Business Requirements

### Use Cases
1. **Multi-domain businesses** - Companies with multiple TLD variants (example.com, example.org, example.net)
2. **Environment management** - Synchronizing aliases between development, staging, and production domains
3. **Brand consolidation** - Merging alias configurations when acquiring domains or rebranding
4. **Backup domain management** - Keeping backup domains in sync with primary domains

### Goals
- Reduce manual effort in maintaining identical alias configurations across domains
- Prevent configuration drift between related domains
- Provide safe, reversible operations with conflict resolution
- Support both automated and interactive workflows

## Technical Specification

### Command Structure

The sync functionality is implemented as a subcommand under `alias`:

```bash
forward-email alias sync <mode> <domains...> [flags]
```

### Sync Modes

#### 1. Merge Sync
**Purpose**: Bidirectional synchronization where all domains receive all unique aliases from all domains.

**Syntax**:
```bash
forward-email alias sync merge domain1.com domain2.com [domain3.com ...]
```

**Behavior**:
- Collects all unique alias names from all specified domains
- Ensures each domain has all collected aliases
- Maintains recipient mappings from the source domain where each alias was found
- Handles conflicts through interactive or automated resolution

#### 2. One-way Sync (Replace Mode)
**Purpose**: Unidirectional synchronization from source domain to target domain(s), removing extra aliases from targets.

**Syntax**:
```bash
forward-email alias sync from source.com target1.com [target2.com ...]
```

**Behavior**:
- Copies all aliases from source domain to target domain(s)
- Removes aliases from target domains that don't exist in source domain
- Overwrites existing alias configurations in target domains
- Handles conflicts through interactive or automated resolution

#### 3. One-way Sync (Preserve Mode)
**Purpose**: Unidirectional synchronization from source domain to target domain(s), preserving extra aliases in targets.

**Syntax**:
```bash
forward-email alias sync from source.com target1.com [target2.com ...] --keep-extra
```

**Behavior**:
- Copies all aliases from source domain to target domain(s)
- Preserves aliases in target domains that don't exist in source domain
- Updates existing aliases to match source configuration
- Handles conflicts through interactive or automated resolution

### Flags and Options

#### Core Flags
- `--dry-run, -d` - Preview changes without applying them
- `--conflicts <strategy>, -c <strategy>` - Non-interactive conflict resolution strategy
- `--keep-extra, -k` - Preserve extra aliases in target domains (one-way sync only)

#### Conflict Resolution Strategies
- `overwrite` - Replace target alias configuration with source configuration
- `skip` - Keep target alias unchanged, skip synchronization for this alias
- `merge` - Combine recipient lists from both source and target aliases

#### Output Control
- `--output <format>, -o <format>` - Output format (table|json|yaml|csv)
- `--verbose, -v` - Enable verbose output
- `--quiet, -q` - Suppress non-essential output

### Interactive Conflict Resolution

When conflicts are detected and no `--conflicts` strategy is specified, the command presents an interactive prompt:

```
Conflict found for alias 'info':
  source.com → team@company.com, sales@company.com
  target.com → support@company.com (existing)

Choose action:
  [o] Overwrite (replace with source settings)
  [s] Skip (keep target unchanged) 
  [m] Merge recipients (forward to team@company.com, sales@company.com, support@company.com)
  [a] Apply to all remaining conflicts
Choice [o/s/m/a]:
```

### Error Handling and Validation

#### Pre-execution Validation
- Verify all specified domains exist and are accessible
- Validate API permissions for all domains
- Check for circular dependencies in merge operations

#### Runtime Error Handling
- API failures are retried with exponential backoff
- Partial failures are reported with detailed error messages
- Operations are atomic per domain (all aliases succeed or all fail)

#### Flag Validation Errors
- `--keep-extra` with `merge` mode → Error: "Cannot use --keep-extra with merge mode (merge naturally preserves all aliases)"
- Invalid conflict strategy → Error: "Invalid conflict strategy 'invalid'. Valid options: overwrite, skip, merge"
- Insufficient arguments → Clear usage message with examples

### Output Format

#### Dry Run Output
```
DRY RUN: Domain Alias Sync Plan
═══════════════════════════════════

Source Analysis:
  source.com: 5 aliases (info, support, sales, admin, noreply)
  target.com: 3 aliases (info, help, contact)

Planned Changes:
  target.com:
    + support → team@company.com (new)
    + sales → sales@company.com (new)  
    + admin → admin@company.com (new)
    + noreply → noreply@company.com (new)
    ! info: conflict detected (support@company.com → team@company.com)
    - help → contact@company.com (will be removed)
    - contact → support@company.com (will be removed)

Summary:
  4 aliases to add
  1 alias conflict to resolve
  2 aliases to remove

Use --conflicts <strategy> to resolve conflicts automatically.
```

#### Execution Output
```
Domain Alias Sync Results
════════════════════════

✓ source.com: 5 aliases analyzed
✓ target.com: 7 changes applied successfully

Summary:
  4 aliases added
  1 alias updated (conflict resolved: merge)
  2 aliases removed
  
Total execution time: 2.3s
```

## Implementation Plan

### Phase 1: Core Infrastructure
- [ ] Add sync command structure to `internal/cmd/alias.go`
- [ ] Implement domain alias collection and comparison logic
- [ ] Create conflict detection and resolution framework

### Phase 2: Sync Mode Implementation
- [ ] Implement merge sync logic
- [ ] Implement one-way sync (replace mode)
- [ ] Implement one-way sync (preserve mode)

### Phase 3: User Experience
- [ ] Add interactive conflict resolution
- [ ] Implement dry-run functionality
- [ ] Add comprehensive error handling and validation

### Phase 4: Testing and Documentation
- [ ] Create comprehensive test suite
- [ ] Add command documentation
- [ ] Update CLI help system

## API Integration

### Required Forward Email API Endpoints
- `GET /v1/domains/{domain}/aliases` - List aliases for a domain
- `POST /v1/domains/{domain}/aliases` - Create alias
- `PUT /v1/domains/{domain}/aliases/{alias}` - Update alias
- `DELETE /v1/domains/{domain}/aliases/{alias}` - Delete alias

### Rate Limiting Considerations
- Implement request batching to respect API rate limits
- Add progress indicators for long-running operations
- Provide options to control concurrency levels

## Security Considerations

- Validate user permissions for all domains before starting sync operations
- Log all sync operations for audit trails
- Implement confirmation prompts for destructive operations
- Ensure atomic operations to prevent partial sync states

## Future Enhancements

### Planned Features
- **Alias filtering** - Sync only specific aliases matching patterns
- **Mailbox settings sync** - Synchronize IMAP passwords, quotas, and other settings
- **Scheduled sync** - Automated periodic synchronization
- **Configuration templates** - Predefined sync configurations for common scenarios

### Integration Opportunities
- **CI/CD integration** - Automated sync in deployment pipelines
- **Webhook triggers** - React to domain changes automatically
- **Bulk import/export** - Integration with CSV/YAML configuration files

## Testing Strategy

### Unit Tests
- Conflict detection logic
- Sync mode implementations
- Flag validation and error handling

### Integration Tests
- Full sync workflows with mock API responses
- Error condition handling
- Output format validation

### End-to-End Tests
- Real API integration with test domains
- Performance testing with large alias sets
- Cross-platform compatibility validation

---

**Document Version**: 1.0  
**Created**: 2025-08-26  
**Author**: AI Assistant  
**Status**: Specification Complete - Ready for Implementation

---

Docs navigation (Dev): [Prev: Makefile Guide](makefile-guide.md) | [Next: Dev Index](README.md) | [Back: User Docs Index](../README.md)
