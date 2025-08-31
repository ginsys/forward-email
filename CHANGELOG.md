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
