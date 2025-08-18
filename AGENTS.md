# Repository Guidelines

## Project Structure & Module Organization
- `cmd/forward-email/`: CLI entrypoint (Cobra).
- `internal/cmd/`: Command implementations (e.g., `domain.go`, `alias.go`, `auth.go`).
- `pkg/api/`: API client and services (domains, aliases, email).
- `pkg/config/`: Configuration loading/saving (`~/.config/forwardemail/config.yaml`).
- `internal/keyring/`: Secure API key storage via OS keyring.
- `internal/client/`, `pkg/output/`, `pkg/errors/`: HTTP client, formatting, error helpers.
- Tests co-located as `*_test.go` within respective packages.

## Build, Test, and Development Commands
```bash
make build        # Build binary to ./bin/forward-email
make run ARGS="domain list"  # Build and run with args
make test         # Run unit tests with race detector
make coverage     # Generate coverage.out + coverage.html
make lint         # Run golangci-lint (or basic checks)
make deps         # Download/tidy modules
make completions  # Generate shell completion scripts
```

## Coding Style & Naming Conventions
- Go 1.21+; format with `go fmt` and imports via `goimports` (local: `github.com/ginsys/forwardemail-cli`).
- Lint with `golangci-lint` per `.golangci.yml` (vet, staticcheck, revive, etc.).
- Packages: lowercase, short, no underscores. Exported types and funcs: `CamelCase`. Errors wrap with `fmt.Errorf("…: %w", err)`.
- CLI commands live in `internal/cmd/<topic>.go`; group subcommands by domain (e.g., `domain`, `alias`, `email`).

## Testing Guidelines
- Use Go’s standard testing (`testing`) with table tests where sensible.
- Name files `*_test.go`; test exported behavior and edge cases.
- Aim for meaningful coverage across `pkg/*` and command surfaces.
- Run `make test` locally; open `coverage.html` from `make coverage` when adjusting tests.

## Security & Configuration Tips
- API keys: prefer OS keyring (`internal/keyring`). Env overrides supported: `FORWARDEMAIL_API_KEY` or `FORWARDEMAIL_<PROFILE>_API_KEY`.
- Config path: `~/.config/forwardemail/config.yaml`; profiles managed via `pkg/config`.
- Never commit secrets; scrub logs and use redaction where applicable.

## Commit & Pull Request Guidelines
- Commits: Conventional Commits (e.g., `feat(api): add domain verify`).
- PRs: clear description, linked issue, tests updated, docs touched if CLI/API changes.
- CI expectations: lint and tests pass; run `make lint && make test` before opening.
- Include usage examples for new/changed commands (e.g., `forward-email domain verify example.com`).
