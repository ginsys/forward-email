# Documentation Index

A quick map to user guides, command help, configuration, troubleshooting, and developer docs for the Forward Email CLI.

## User Guides
- Quick Start: `docs/quick-start.md` — install, auth, first commands.
- Command Reference: `docs/commands.md` — full CLI command list with examples.
- Configuration Guide: `docs/configuration.md` — profiles, env vars, config files.
- Troubleshooting: `docs/troubleshooting.md` — common issues and fixes.

## Common Tasks
- Build: `make build` → `bin/forward-email`
- Test: `make test` (race) or `make test-ci` (coverage)
- Lint: `make lint`
- Version: `forward-email version --verbose` or `--json`

## Developer Docs
- Contributing Guide: `docs/development/contributing.md`
- Architecture Overview: `docs/development/architecture.md`
- API Integration: `docs/development/api-integration.md`
- Testing Strategy: `docs/development/testing.md`
- Domain Alias Sync Spec: `docs/development/domain-alias-sync-specification.md`

## Release & Versioning
- Plan and checklist: `VERSIONING_RELEASE_PLAN.md`

## Tips
- Global help: `forward-email --help`
- Command help: `forward-email <cmd> --help` (e.g., `forward-email domain --help`)
- Output formats: `--output table|json|yaml|csv`

---

Docs navigation: [Next: Quick Start](quick-start.md)
