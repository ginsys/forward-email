# Forward Email CLI - Tasks

## Implemented Features
- Authentication (login/verify/status/logout, multi-source: env → keyring → config)
- Profiles (list/show/create/switch/delete, multi-environment)
- Domains (CRUD, DNS verification, member management, verification)
- Aliases (CRUD, enable/disable, recipients, PGP, vacation responder)
- Email (send with attachments, list/get/delete, interactive mode)
- Output formats (table/JSON/YAML/CSV/plain)
- Cross-platform support (Linux/macOS/Windows)
- Debug utilities

## Known Issues
None currently identified.

## API Implementation Tasks

> **Note**: Keep this section in sync with `docs/development/api-reference.md`.
> When completing tasks, update both files and mark items as complete.

### Complete Partial Implementations
- [x] Add `--smtp` flag to `domain verify` command (endpoint exists, flag missing)
- [x] Expose `alias password` command (client method exists, command not wired)
- [x] Verify/fix `email quota` endpoint path (may use `/quota` instead of `/limit`)

### Fix Implementation Errors
- [x] Verify `domain dns` endpoint exists, remove command if not (`/v1/domains/:id/dns`) - **REMOVED: endpoint does not exist**
- [x] Verify `domain quota` endpoint exists, remove command if not (`/v1/domains/:id/quota`) - **REMOVED: endpoint does not exist**
- [x] Verify `domain stats` endpoint exists, remove command if not (`/v1/domains/:id/stats`) - **REMOVED: endpoint does not exist**
- [x] Fix `email list` from/to field mapping - **FIXED: updated to use headers map**
- [x] Fix `email get` id/from/to field mapping - **FIXED: updated to use headers map**

### New Commands - Account Management
- [ ] `account show` - Get account details (`GET /v1/account`)
- [ ] `account update` - Update account settings (`PUT /v1/account`)

### New Commands - Domain Management
- [ ] `domain members update` - Update member role (`PUT /v1/domains/:id/members/:member_id`)
- [ ] `domain invites list` - List pending invites (`GET /v1/domains/:id/invites`)
- [ ] `domain invites send` - Send invitation (`POST /v1/domains/:id/invites`)
- [ ] `domain invites cancel` - Cancel invitation (`DELETE /v1/domains/:id/invites`)

### New Commands - Alias Management
- [ ] `alias catch-all list` - List catch-all passwords (`GET /v1/domains/:id/catch-all-passwords`)
- [ ] `alias catch-all create` - Generate catch-all password (`POST /v1/domains/:id/catch-all-passwords`)
- [ ] `alias catch-all delete` - Delete catch-all password (`DELETE /v1/domains/:id/catch-all-passwords/:token_id`)

### New Commands - Logs
- [ ] `logs download` - Download logs with rate limits (`GET /v1/logs/download`)

### New Commands - Debug/Utility
- [ ] `debug lookup` - Email address lookup (`GET /v1/lookup`)
- [ ] `debug port` - Port availability check (`GET /v1/port`)
- [ ] `debug self-test` - Run self-test (`POST /v1/self-test`)
- [ ] `debug settings` - Get settings (`GET /v1/settings`)
- [ ] `debug max-forwarded` - Get forwarding limits (`GET /v1/max-forwarded-addresses`)

### Deferred (Low Priority)
- Contacts API (5 endpoints) - CardDAV, better via native apps
- Calendars API (5 endpoints) - CalDAV, better via native apps
- Calendar Events API (5 endpoints) - CalDAV, better via native apps
- Messages API (5 endpoints) - IMAP, better via protocol
- Folders API (5 endpoints) - IMAP, better via protocol

## Planned Features

### Testing
- Email service test coverage
- Alias service test coverage
- Integration tests with mock API server
- Performance benchmarks

### User Experience
- Interactive setup wizard (`forward-email init`)
- Shell completion (bash/zsh/fish)
- Enhanced help text with examples
- Better error messages with suggested fixes

### Bulk Operations
- Domain alias synchronization
  - Notes: Specification complete in `docs/development/domain-alias-sync-specification.md`
- CSV import/export for aliases
- Concurrent processing with progress tracking

### CI/CD & Release
- GitHub Actions improvements
- GoReleaser automation
- Package distribution (Homebrew, Chocolatey)

### Future (Phase 2+)
- Template system for emails
- Real-time log monitoring
- Webhook management
- Log download (respecting API limits)
- Plugin architecture

---
*Last Updated: 2026-01-18*
