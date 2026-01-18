# Forward Email CLI - Project Context

**Status**: Phase 1.3 COMPLETE ✅ | **LOC**: 17,151 | **Packages**: 11 | **Files**: 47 Go files

## Overview
Comprehensive command-line interface for Forward Email API management. **First-mover advantage** - Forward Email has zero official CLI tools despite 20+ API endpoints.

## Current Implementation (100% Functional)

### ✅ Core Features
- **Authentication**: Multi-source (env → keyring → config) with HTTP Basic API key auth
- **Profiles**: Multi-environment workflows (dev/staging/prod)
- **Domains**: Full CRUD, DNS verification, member management
- **Aliases**: Complete lifecycle, recipients, PGP, vacation responder
- **Email**: Interactive/programmatic sending, attachments, history
- **Output**: Multiple formats (table/JSON/YAML/CSV) with filtering
- **Cross-Platform**: Linux/macOS/Windows support

### 🔧 Architecture
- **Framework**: Cobra + Viper CLI with structured commands
- **Security**: OS keyring integration (99designs/keyring)
- **API Integration**: `https://api.forwardemail.net/v1/` with HTTP Basic auth
- **Service Layer**: Complete separation between API client and CLI commands
- **Error Handling**: Centralized management with user-friendly messages

### 📊 Quality Metrics
- **Tests**: 100+ test cases across all packages - ALL PASSING ✅
- **Coverage**: Comprehensive unit tests with mock implementations
- **Build**: Cross-platform builds successful
- **Linting**: All linting issues resolved

## API Coverage Status
- **Domains**: CRUD, DNS/SMTP verification ✅ **COMPLETE**
- **Aliases**: Full lifecycle with all settings ✅ **COMPLETE**
- **Emails**: Send operations with attachments ✅ **COMPLETE**
- **Account**: Profile management (planned Phase 1.4)
- **Logs**: Download with rate limits (planned Phase 1.4)

## Key Technical Decisions
- **Auth Method**: HTTP Basic with `Authorization: Basic <base64(api_key + ":")>`
- **Config Management**: Profile-based with Viper supporting multiple environments
- **Output Strategy**: Stable ordering across all formats for consistency
- **Testing Strategy**: Mock implementations for reliable unit testing
- **CLI Structure**: Cobra framework with global flags and comprehensive help

## Available Commands
```
auth        - Manage authentication (login/verify/status/logout)
profile     - Configuration profiles (list/show/create/switch/delete)
domain      - Domain operations (list/get/create/update/delete/verify)
alias       - Alias management (list/get/create/update/delete/enable/disable)
email       - Email operations (send/list/get/delete)
debug       - Troubleshooting utilities
completion  - Shell completion scripts
```

## Planning Documents
- **[TASKS.md](./TASKS.md)**: Centralized task management and roadmap
- **[VERSIONING_RELEASE_PLAN.md](./docs/VERSIONING_RELEASE_PLAN.md)**: Release automation strategy

## Next Phase (1.4)
- Enhanced test coverage for email services
- Bulk operations and domain alias synchronization
- Interactive setup wizards
- Shell completion and CI/CD automation

---

# Repository Guidelines

## Project Structure & Module Organization
- `cmd/forward-email/`: CLI entrypoint (`main.go`) and integration tests.
- `internal/cmd/`: Cobra command tree (root, auth, domain, alias, email, profile, debug).
- `internal/client/`, `internal/keyring/`, `internal/version/`: API client, secure key storage, build info.
- `docs/`: User and developer docs; update when commands or flags change.
- `bin/`: Built binaries (ignored by Git).

## Build, Test, and Development Commands
- `make build`: Compile `bin/forward-email` with version metadata.
- `make test` / `go test ./...`: Run tests (race detector default).
- `make coverage`: Generate `coverage.out` and `coverage.html`.
- `make lint`: Run `golangci-lint` (fallback to `gofmt` + `go vet`).
- `make pre-commit` / `make install-hooks`: Fast checks and Git hook setup.
- `make dev-setup`: Dependencies + local tooling.

## Coding Style & Naming Conventions
- Language: Go 1.21+. Format with `gofmt`; lint with `golangci-lint`.
- Packages and dirs: lowercase, no underscores; small, cohesive modules under `internal/`.
- Commands: nouns/verbs consistent with existing groups (auth, domain, alias, email, profile, debug).
- Flags: kebab-case; bind to Viper; support `FORWARDEMAIL_*` env vars.

## Testing Guidelines
- Use Go testing + `testify` assertions. Place tests alongside code as `*_test.go`.
- Prefer table-driven tests; cover success, error, and edge cases.
- Integration tests live under `cmd/forward-email/` and exercise the built binary.
- Run `make test` locally; keep or improve coverage; use `make test-bench` when relevant.

## Commit & Pull Request Guidelines
- Commits: Conventional Commits (e.g., `feat(alias): add sync subcommand`, `fix(client): handle 401 retry`).
- PRs: clear description, linked issues, CLI output/screenshots when UX changes, and docs updates under `docs/`.
- Checklist before opening PR: `make pre-commit`, green tests, updated docs.

## Security & Configuration Tips
- Never commit secrets. API keys are stored via OS keyring.
- Config loads from `~/.config/forwardemail/config.yaml`, current dir, and `FORWARDEMAIL_*` env vars.
- Validate inputs and redact sensitive fields in logs; respect `--profile`, `--output`, `--timeout`.

## Agent-Specific Instructions (.claude/)
- `settings.json`: Allowed commands and memory path. Do not loosen permissions; prefer `make` targets over ad‑hoc commands.
- Agents: run `make pre-commit` locally; use `make test-ci` for CI parity; avoid real API calls in tests—use mocks and env vars (`FORWARDEMAIL_*`).
- See **Planning Documents** section above for TASKS.md (root) and VERSIONING_RELEASE_PLAN.md (docs/).

---
*Last Updated: 2026-01-18*
