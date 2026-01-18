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
- Domain dns command: endpoint `/v1/domains/{id}/dns` may not exist
- Domain quota command: endpoint `/v1/domains/{id}/quota` may not exist
- Email list: from/to fields empty (JSON response structure mismatch)
- Email get: id/from/to fields empty (JSON response structure mismatch)

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
