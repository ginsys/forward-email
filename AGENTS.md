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
- `CLAUDE.md`: Project memory/context. Update when commands, flags, or architecture change.
- `TASKS.md`: Roadmap and task tracker. Keep phase/status in sync with README and PRs.
- `settings.json`: Allowed commands and memory path. Do not loosen permissions; prefer `make` targets over ad‑hoc commands.
- `VERSIONING_RELEASE_PLAN.md`: Follow for version cmd, CHANGELOG, tags, and GoReleaser setup; align with Makefile LDFLAGS and `internal/version`.
- Agents: run `make pre-commit` locally; use `make test-ci` for CI parity; avoid real API calls in tests—use mocks and env vars (`FORWARDEMAIL_*`).
