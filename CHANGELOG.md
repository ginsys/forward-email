## v0.1.0-alpha.0 - 2025-08-31

- Summary: (fill in)

## v0.1.0-alpha.2 - 2025-08-31

- Summary: (fill in)

# Changelog

All notable changes to this project will be documented in this file.

The format is based on Keep a Changelog, and this project adheres to Semantic Versioning after v1.0.0.

## Unreleased
### Added
- Version subcommand with `--json`, `--verbose`, `--license`, `--check-update` flags.
- Shell completion command for bash, zsh, fish, PowerShell.
- Interactive `init` setup wizard (profile + keyring + config).
- Documentation index and navigation across user and developer docs.
- Email service unit tests (send, list, get, delete).
- GoReleaser config and GitHub release workflow.

### Changed
- Root command now sources version metadata from `internal/version`.
- README Quick Start includes version check example.

### Fixed
- Tests no longer depend on removed root version variable.

## v0.1.0-alpha.1 - 2025-08-31
### Added
- Alias sync (merge/replace/preserve), dry-run, interactive conflicts, `--yes` for non-interactive.
- CSV import/export for aliases; `import --dry-run` preview.
- Credential store selector: `auth login` and `init` support keyring/file/config.
- Docs: releasing guide; command docs updated with sync and CSV details.

### Changed
- Version handling via `version` subcommand; root `--version` flag removed.
- Tests ensure keyring is disabled; stable CI flow.
