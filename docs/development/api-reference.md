# Forward Email API Reference

**Last Updated**: 2026-01-18
**CLI Version**: 1.3.x
**API Base URL**: `https://api.forwardemail.net/v1/`
**Authentication**: HTTP Basic (API key + `:`)

---

## Overview

This document provides a complete mapping between Forward Email API endpoints and CLI commands, tracking implementation status and identifying gaps.

### Quick Reference

| Category | Implemented | Total | Coverage |
|----------|-------------|-------|----------|
| Accounts | 0 | 3 | 0% |
| Domains | 6 | 7 | 86% |
| Domain Invites | 0 | 3 | 0% |
| Domain Members | 1 | 2 | 50% |
| Aliases | 5 | 9 | 56% |
| Emails | 4 | 5 | 80% |
| Logs | 0 | 2 | 0% |
| Contacts (CardDAV) | 0 | 5 | 0% |
| Calendars (CalDAV) | 0 | 5 | 0% |
| Calendar Events | 0 | 5 | 0% |
| Messages (IMAP) | 0 | 5 | 0% |
| Folders (IMAP) | 0 | 5 | 0% |
| Utility | 0 | 5 | 0% |
| **Total** | **16** | **61** | **26%** |

### Implementation Legend

- ✅ **Implemented**: Full CLI command exists
- ⚠️ **Partial**: Endpoint implemented but missing flags or full functionality
- ❌ **Not Implemented**: No CLI command exists

---

## API Categories

### 1. Account Management

Account operations for user profile and settings.

| Method | Endpoint | CLI Command | Status | Notes |
|--------|----------|-------------|--------|-------|
| `POST` | `/v1/account` | - | ❌ | Create account |
| `GET` | `/v1/account` | - | ❌ | Get account details |
| `PUT` | `/v1/account` | - | ❌ | Update account |

**Planned Phase**: 1.4

---

### 2. Domain Operations

Complete domain lifecycle management with DNS verification.

| Method | Endpoint | CLI Command | Status | Notes |
|--------|----------|-------------|--------|-------|
| `GET` | `/v1/domains` | `domain list` | ✅ | List all domains |
| `POST` | `/v1/domains` | `domain create <name>` | ✅ | Create new domain |
| `GET` | `/v1/domains/:domain_id` | `domain get <domain>` | ✅ | Get domain details |
| `PUT` | `/v1/domains/:domain_id` | `domain update <domain>` | ✅ | Update domain settings |
| `DELETE` | `/v1/domains/:domain_id` | `domain delete <domain>` | ✅ | Delete domain |
| `GET` | `/v1/domains/:domain_id/verify-records` | `domain verify <domain>` | ✅ | Verify DNS records |
| `GET` | `/v1/domains/:domain_id/verify-smtp` | `domain verify --smtp` | ⚠️ | SMTP verification (flag missing) |

**Implementation**: Core functionality complete (86%)
**Known Issue**: SMTP verification endpoint exists but `--smtp` flag not exposed in CLI

---

### 3. Domain Invites

Invite management for domain members.

| Method | Endpoint | CLI Command | Status | Notes |
|--------|----------|-------------|--------|-------|
| `GET` | `/v1/domains/:domain_id/invites` | - | ❌ | List invites |
| `POST` | `/v1/domains/:domain_id/invites` | - | ❌ | Send invite |
| `DELETE` | `/v1/domains/:domain_id/invites` | - | ❌ | Cancel invite |

**Planned Phase**: Future

---

### 4. Domain Members

Member management for domain access control.

| Method | Endpoint | CLI Command | Status | Notes |
|--------|----------|-------------|--------|-------|
| `PUT` | `/v1/domains/:domain_id/members/:member_id` | - | ❌ | Update member role |
| `DELETE` | `/v1/domains/:domain_id/members/:member_id` | `domain members remove` | ✅ | Remove member |

**Implementation**: Partial (50%)
**Note**: `domain members list` extracts members from `GET /v1/domains/:id` response

---

### 5. Alias Management

Complete alias lifecycle with recipient and password management.

| Method | Endpoint | CLI Command | Status | Notes |
|--------|----------|-------------|--------|-------|
| `GET` | `/v1/domains/:domain_id/aliases` | `alias list <domain>` | ✅ | List aliases |
| `POST` | `/v1/domains/:domain_id/aliases` | `alias create <domain> <name>` | ✅ | Create alias |
| `GET` | `/v1/domains/:domain_id/aliases/:alias_id` | `alias get <domain> <alias>` | ✅ | Get alias details |
| `PUT` | `/v1/domains/:domain_id/aliases/:alias_id` | `alias update <domain> <alias>` | ✅ | Update alias |
| `DELETE` | `/v1/domains/:domain_id/aliases/:alias_id` | `alias delete <domain> <alias>` | ✅ | Delete alias |
| `POST` | `/v1/domains/:domain_id/aliases/:alias_id/generate-password` | `alias password` | ⚠️ | Defined but not exposed |
| `GET` | `/v1/domains/:domain_id/catch-all-passwords` | - | ❌ | List catch-all passwords |
| `POST` | `/v1/domains/:domain_id/catch-all-passwords` | - | ❌ | Generate catch-all password |
| `DELETE` | `/v1/domains/:domain_id/catch-all-passwords/:token_id` | - | ❌ | Delete catch-all password |

**Implementation**: Core functionality complete (56%)
**Known Issue**: Password generation endpoint implemented but command not exposed
**Missing**: Catch-all password management

---

### 6. Email Operations

Email sending and management with attachment support.

| Method | Endpoint | CLI Command | Status | Notes |
|--------|----------|-------------|--------|-------|
| `POST` | `/v1/emails` | `email send` | ✅ | Send email with attachments |
| `GET` | `/v1/emails` | `email list` | ✅ | List sent emails |
| `GET` | `/v1/emails/limit` | `email quota` | ⚠️ | Uses `/quota` not `/limit` |
| `GET` | `/v1/emails/:id` | `email get <id>` | ✅ | Get email details |
| `DELETE` | `/v1/emails/:id` | `email delete <id>` | ✅ | Delete email |

**Implementation**: Mostly complete (80%)
**Known Issue**: `email quota` may be using incorrect endpoint path

---

### 7. Logs & Analytics

Log download and submission endpoints.

| Method | Endpoint | CLI Command | Status | Notes |
|--------|----------|-------------|--------|-------|
| `GET` | `/v1/logs/download` | - | ❌ | Download logs |
| `POST` | `/v1/log` | - | ❌ | Submit log entry |

**Planned Phase**: 1.4

---

### 8. Contacts (CardDAV)

CardDAV contact management endpoints.

| Method | Endpoint | CLI Command | Status | Notes |
|--------|----------|-------------|--------|-------|
| `GET` | `/v1/contacts` | - | ❌ | List contacts |
| `POST` | `/v1/contacts` | - | ❌ | Create contact |
| `GET` | `/v1/contacts/:id` | - | ❌ | Get contact |
| `PUT` | `/v1/contacts/:id` | - | ❌ | Update contact |
| `DELETE` | `/v1/contacts/:id` | - | ❌ | Delete contact |

**Planned Phase**: Future (low priority for CLI)

---

### 9. Calendars (CalDAV)

CalDAV calendar management endpoints.

| Method | Endpoint | CLI Command | Status | Notes |
|--------|----------|-------------|--------|-------|
| `GET` | `/v1/calendars` | - | ❌ | List calendars |
| `POST` | `/v1/calendars` | - | ❌ | Create calendar |
| `GET` | `/v1/calendars/:id` | - | ❌ | Get calendar |
| `PUT` | `/v1/calendars/:id` | - | ❌ | Update calendar |
| `DELETE` | `/v1/calendars/:id` | - | ❌ | Delete calendar |

**Planned Phase**: Future (low priority for CLI)

---

### 10. Calendar Events

CalDAV event management endpoints.

| Method | Endpoint | CLI Command | Status | Notes |
|--------|----------|-------------|--------|-------|
| `GET` | `/v1/calendar-events` | - | ❌ | List events |
| `POST` | `/v1/calendar-events` | - | ❌ | Create event |
| `GET` | `/v1/calendar-events/:id` | - | ❌ | Get event |
| `PUT` | `/v1/calendar-events/:id` | - | ❌ | Update event |
| `DELETE` | `/v1/calendar-events/:id` | - | ❌ | Delete event |

**Planned Phase**: Future (low priority for CLI)

---

### 11. Messages (IMAP)

IMAP message management endpoints.

| Method | Endpoint | CLI Command | Status | Notes |
|--------|----------|-------------|--------|-------|
| `GET` | `/v1/messages` | - | ❌ | List messages |
| `POST` | `/v1/messages` | - | ❌ | Create message |
| `GET` | `/v1/messages/:id` | - | ❌ | Get message |
| `PUT` | `/v1/messages/:id` | - | ❌ | Update message |
| `DELETE` | `/v1/messages/:id` | - | ❌ | Delete message |

**Planned Phase**: Future (IMAP operations typically done via protocol)

---

### 12. Folders (IMAP)

IMAP folder management endpoints.

| Method | Endpoint | CLI Command | Status | Notes |
|--------|----------|-------------|--------|-------|
| `GET` | `/v1/folders` | - | ❌ | List folders |
| `POST` | `/v1/folders` | - | ❌ | Create folder |
| `GET` | `/v1/folders/:id` | - | ❌ | Get folder |
| `PUT` | `/v1/folders/:id` | - | ❌ | Update folder |
| `DELETE` | `/v1/folders/:id` | - | ❌ | Delete folder |

**Planned Phase**: Future (IMAP operations typically done via protocol)

---

### 13. Utility Endpoints

Utility and configuration endpoints.

| Method | Endpoint | CLI Command | Status | Notes |
|--------|----------|-------------|--------|-------|
| `GET` | `/v1/lookup` | - | ❌ | Email address lookup |
| `GET` | `/v1/port` | - | ❌ | Port availability check |
| `GET` | `/v1/max-forwarded-addresses` | - | ❌ | Get forwarding limits |
| `POST` | `/v1/self-test` | - | ❌ | Run self-test |
| `GET` | `/v1/settings` | - | ❌ | Get settings |

**Planned Phase**: Future (consider for debug commands)

---

## Detailed Endpoint Reference

### Account Management

#### `POST /v1/account`
**Create Account**

- **CLI**: Not implemented
- **Status**: ❌
- **Priority**: Medium (Phase 1.4)

#### `GET /v1/account`
**Get Account Details**

- **CLI**: Not implemented
- **Status**: ❌
- **Priority**: High (Phase 1.4)
- **Proposed Command**: `account show` or `account info`

#### `PUT /v1/account`
**Update Account**

- **CLI**: Not implemented
- **Status**: ❌
- **Priority**: High (Phase 1.4)
- **Proposed Command**: `account update`

---

### Domain Operations

#### `GET /v1/domains`
**List Domains**

- **CLI**: `domain list [flags]`
- **Status**: ✅ Fully implemented
- **Flags**:
  - `--output, -o`: Output format (table/json/yaml/csv)
  - `--filter`: Filter domains by pattern
- **File**: `internal/cmd/domain/list.go:55`

#### `POST /v1/domains`
**Create Domain**

- **CLI**: `domain create <name> [flags]`
- **Status**: ✅ Fully implemented
- **Flags**:
  - `--plan`: Domain plan (free/enhanced_protection/team)
  - `--has-mx-record`: Set MX record status
  - `--has-txt-record`: Set TXT record status
- **File**: `internal/cmd/domain/create.go:43`

#### `GET /v1/domains/:domain_id`
**Get Domain Details**

- **CLI**: `domain get <domain> [flags]`
- **Status**: ✅ Fully implemented
- **Flags**:
  - `--output, -o`: Output format
- **File**: `internal/cmd/domain/get.go:37`

#### `PUT /v1/domains/:domain_id`
**Update Domain**

- **CLI**: `domain update <domain> [flags]`
- **Status**: ✅ Fully implemented
- **Flags**:
  - `--plan`: Update plan
  - `--smtp-port`: Custom SMTP port
  - Multiple configuration flags
- **File**: `internal/cmd/domain/update.go:51`

#### `DELETE /v1/domains/:domain_id`
**Delete Domain**

- **CLI**: `domain delete <domain> [flags]`
- **Status**: ✅ Fully implemented
- **Flags**:
  - `--force, -f`: Skip confirmation
- **File**: `internal/cmd/domain/delete.go:38`

#### `GET /v1/domains/:domain_id/verify-records`
**Verify DNS Records**

- **CLI**: `domain verify <domain> [flags]`
- **Status**: ✅ Fully implemented
- **File**: `internal/cmd/domain/verify.go:39`

#### `GET /v1/domains/:domain_id/verify-smtp`
**Verify SMTP Configuration**

- **CLI**: `domain verify --smtp <domain>` (flag not exposed)
- **Status**: ⚠️ Partial
- **Issue**: Endpoint exists but CLI flag missing
- **Fix Required**: Add `--smtp` flag to verify command

---

### Domain Invites

#### `GET /v1/domains/:domain_id/invites`
**List Domain Invites**

- **CLI**: Not implemented
- **Status**: ❌
- **Priority**: Low
- **Proposed Command**: `domain invites list <domain>`

#### `POST /v1/domains/:domain_id/invites`
**Send Domain Invite**

- **CLI**: Not implemented
- **Status**: ❌
- **Priority**: Low
- **Proposed Command**: `domain invites send <domain> <email>`

#### `DELETE /v1/domains/:domain_id/invites`
**Cancel Domain Invite**

- **CLI**: Not implemented
- **Status**: ❌
- **Priority**: Low
- **Proposed Command**: `domain invites cancel <domain> <invite_id>`

---

### Domain Members

#### `PUT /v1/domains/:domain_id/members/:member_id`
**Update Member Role**

- **CLI**: Not implemented
- **Status**: ❌
- **Priority**: Medium
- **Proposed Command**: `domain members update <domain> <member_id> --role <role>`

#### `DELETE /v1/domains/:domain_id/members/:member_id`
**Remove Domain Member**

- **CLI**: `domain members remove <domain> <member_id> [flags]`
- **Status**: ✅ Fully implemented
- **Flags**:
  - `--force, -f`: Skip confirmation
- **File**: `internal/cmd/domain/members.go:112`

#### Member Listing (Virtual Command)
**List Domain Members**

- **CLI**: `domain members list <domain>`
- **Status**: ✅ Implemented (extracts from domain response)
- **Method**: Calls `GET /v1/domains/:id` and filters members
- **File**: `internal/cmd/domain/members.go:54`

---

### Alias Management

#### `GET /v1/domains/:domain_id/aliases`
**List Aliases**

- **CLI**: `alias list <domain> [flags]`
- **Status**: ✅ Fully implemented
- **Flags**:
  - `--output, -o`: Output format
  - `--filter`: Filter aliases
- **File**: `internal/cmd/alias/list.go:51`

#### `POST /v1/domains/:domain_id/aliases`
**Create Alias**

- **CLI**: `alias create <domain> <name> [flags]`
- **Status**: ✅ Fully implemented
- **Flags**:
  - `--recipients`: Recipient addresses
  - `--description`: Alias description
  - `--labels`: Comma-separated labels
  - Multiple configuration flags
- **File**: `internal/cmd/alias/create.go:57`

#### `GET /v1/domains/:domain_id/aliases/:alias_id`
**Get Alias Details**

- **CLI**: `alias get <domain> <alias> [flags]`
- **Status**: ✅ Fully implemented
- **File**: `internal/cmd/alias/get.go:41`

#### `PUT /v1/domains/:domain_id/aliases/:alias_id`
**Update Alias**

- **CLI**: `alias update <domain> <alias> [flags]`
- **Status**: ✅ Fully implemented
- **File**: `internal/cmd/alias/update.go:57`

#### `DELETE /v1/domains/:domain_id/aliases/:alias_id`
**Delete Alias**

- **CLI**: `alias delete <domain> <alias> [flags]`
- **Status**: ✅ Fully implemented
- **Flags**:
  - `--force, -f`: Skip confirmation
- **File**: `internal/cmd/alias/delete.go:41`

#### `POST /v1/domains/:domain_id/aliases/:alias_id/generate-password`
**Generate Alias Password**

- **CLI**: `alias password <domain> <alias>` (not exposed)
- **Status**: ⚠️ Partial
- **Issue**: Client method exists but command not exposed
- **Fix Required**: Expose password generation command

#### `GET /v1/domains/:domain_id/catch-all-passwords`
**List Catch-All Passwords**

- **CLI**: Not implemented
- **Status**: ❌
- **Priority**: Low

#### `POST /v1/domains/:domain_id/catch-all-passwords`
**Generate Catch-All Password**

- **CLI**: Not implemented
- **Status**: ❌
- **Priority**: Low

#### `DELETE /v1/domains/:domain_id/catch-all-passwords/:token_id`
**Delete Catch-All Password**

- **CLI**: Not implemented
- **Status**: ❌
- **Priority**: Low

#### Alias Enable/Disable (Virtual Commands)
**Enable/Disable Alias**

- **CLI**: `alias enable/disable <domain> <alias>`
- **Status**: ✅ Implemented (calls PUT with `is_enabled` field)
- **Method**: Updates `is_enabled` field via `PUT /v1/domains/:id/aliases/:id`
- **File**: `internal/cmd/alias/enable.go:35`, `internal/cmd/alias/disable.go:35`

#### Alias Recipients (Virtual Command)
**Manage Recipients**

- **CLI**: `alias recipients <domain> <alias> <recipients>`
- **Status**: ✅ Implemented (calls PUT with recipients only)
- **Method**: Updates recipients via `PUT /v1/domains/:id/aliases/:id`
- **File**: `internal/cmd/alias/recipients.go:40`

---

### Email Operations

#### `POST /v1/emails`
**Send Email**

- **CLI**: `email send [flags]`
- **Status**: ✅ Fully implemented
- **Flags**:
  - `--from`: Sender address (required)
  - `--to`: Recipient addresses (required)
  - `--subject`: Email subject
  - `--text`: Plain text body
  - `--html`: HTML body
  - `--cc`, `--bcc`: Carbon copy recipients
  - `--attachments`: File attachments
  - `--interactive, -i`: Interactive mode
- **File**: `internal/cmd/email/send.go:59`

#### `GET /v1/emails`
**List Emails**

- **CLI**: `email list [flags]`
- **Status**: ✅ Fully implemented
- **Flags**:
  - `--output, -o`: Output format
  - `--limit`: Result limit
  - `--offset`: Result offset
- **File**: `internal/cmd/email/list.go:45`

#### `GET /v1/emails/limit`
**Get Email Quota**

- **CLI**: `email quota`
- **Status**: ⚠️ Potential issue
- **Issue**: CLI may use `/v1/emails/quota` instead of `/v1/emails/limit`
- **Fix Required**: Verify correct endpoint path
- **File**: `internal/client/client.go` (needs verification)

#### `GET /v1/emails/:id`
**Get Email Details**

- **CLI**: `email get <id> [flags]`
- **Status**: ✅ Fully implemented
- **File**: `internal/cmd/email/get.go:39`

#### `DELETE /v1/emails/:id`
**Delete Email**

- **CLI**: `email delete <id> [flags]`
- **Status**: ✅ Fully implemented
- **Flags**:
  - `--force, -f`: Skip confirmation
- **File**: `internal/cmd/email/delete.go:39`

---

### Logs & Analytics

#### `GET /v1/logs/download`
**Download Logs**

- **CLI**: Not implemented
- **Status**: ❌
- **Priority**: High (Phase 1.4)
- **Proposed Command**: `logs download [flags]`
- **Notes**: Rate-limited endpoint

#### `POST /v1/log`
**Submit Log Entry**

- **CLI**: Not implemented
- **Status**: ❌
- **Priority**: Low

---

## Multi-Endpoint Commands

Some CLI commands orchestrate multiple API calls or extract data from nested responses:

| Command | Primary Endpoint | Additional Operations | Description |
|---------|------------------|----------------------|-------------|
| `auth verify` | `GET /v1/domains` | None | Tests API connectivity by listing domains |
| `domain members list` | `GET /v1/domains/:id` | Extracts `members` field | Virtual command using domain response |
| `alias enable` | `PUT /v1/domains/:id/aliases/:id` | Sets `is_enabled: true` | Specialized update operation |
| `alias disable` | `PUT /v1/domains/:id/aliases/:id` | Sets `is_enabled: false` | Specialized update operation |
| `alias recipients` | `PUT /v1/domains/:id/aliases/:id` | Updates `recipients` only | Specialized update operation |

---

## Known Issues & Discrepancies

### Non-Existent Endpoints (Cleanup Required)

These endpoints were speculatively implemented but don't exist in the actual API:

| Endpoint | CLI Command | Status | Action Required |
|----------|-------------|--------|-----------------|
| ~~`/v1/emails/stats`~~ | ~~`email stats`~~ | ✅ | Already removed |
| `/v1/domains/:id/dns` | `domain dns` | ⚠️ | Verify existence or remove |
| `/v1/domains/:id/quota` | `domain quota` | ⚠️ | Verify existence or remove |
| `/v1/domains/:id/stats` | `domain stats` | ⚠️ | Verify existence or remove |
| `/v1/emails/quota` | `email quota` | ⚠️ | Should use `/v1/emails/limit` |

### Partial Implementations

| Endpoint | Issue | Fix Required |
|----------|-------|--------------|
| `/v1/domains/:id/verify-smtp` | `--smtp` flag not exposed | Add flag to `domain verify` command |
| `/v1/domains/:id/aliases/:id/generate-password` | Command not exposed | Expose `alias password` command |
| `/v1/emails/limit` | May be using wrong path | Verify endpoint path in client |

---

## Implementation Roadmap

### Phase 1.4 (Current - Q1 2026)

**Priority: Account Management**
- [ ] `account show` - Get account details (`GET /v1/account`)
- [ ] `account update` - Update account settings (`PUT /v1/account`)

**Priority: Logs & Analytics**
- [ ] `logs download` - Download logs (`GET /v1/logs/download`)

**Priority: Fixes**
- [ ] Add `--smtp` flag to `domain verify`
- [ ] Expose `alias password` command
- [ ] Verify/fix `email quota` endpoint path

### Phase 1.5 (Q2 2026)

**Priority: Domain Members**
- [ ] `domain members update` - Update member roles

**Priority: Invites**
- [ ] `domain invites list` - List pending invites
- [ ] `domain invites send` - Send invitation
- [ ] `domain invites cancel` - Cancel invitation

### Phase 2.0 (Future)

**Priority: Advanced Alias Management**
- [ ] Catch-all password commands (3 endpoints)

**Priority: Utility Commands**
- [ ] `debug lookup` - Email address lookup
- [ ] `debug self-test` - Run self-test
- [ ] Other utility endpoints as needed

**Low Priority: CalDAV/CardDAV/IMAP**
- Contacts (5 endpoints) - Consider web UI instead
- Calendars (5 endpoints) - Consider web UI instead
- Calendar Events (5 endpoints) - Consider web UI instead
- Messages (5 endpoints) - Use IMAP protocol
- Folders (5 endpoints) - Use IMAP protocol

---

## Testing Checklist

When implementing new endpoints:

1. **Client Method**: Add to `internal/client/client.go`
2. **Command**: Create Cobra command in `internal/cmd/<category>/`
3. **Tests**: Add unit tests with mocks
4. **Documentation**: Update this file and user docs
5. **Help Text**: Ensure `--help` is comprehensive
6. **Output Formats**: Support all formats (table/json/yaml/csv)
7. **Error Handling**: User-friendly error messages
8. **Integration Test**: Add to `cmd/forward-email/integration_test.go`

---

## References

- [Forward Email API Docs](https://forwardemail.net/en/email-api)
- [Forward Email Source Code](https://github.com/forwardemail/forwardemail.net)
- [CLI TASKS.md](../../TASKS.md)
- [CLI User Guide](../USER_GUIDE.md)

---

## Maintenance

This document should be updated when:
- New API endpoints are added to Forward Email
- New CLI commands are implemented
- Implementation status changes
- Endpoint paths or behavior changes
- Bugs or discrepancies are discovered

**Document Maintainer**: Development team
**Review Frequency**: Every minor version release
