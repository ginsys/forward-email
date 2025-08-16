# Forward Email CLI Project Memory

## Project Overview
Go-based CLI tool for Forward Email API management with enterprise-grade features and developer-first experience.

## Architecture Principles
- Clean separation: SDK (pkg/api) � CLI commands (cmd/) � User interface
- Security-first: OS keyring integration, credential redaction, secure defaults
- Developer Experience: Shell completion, interactive wizards, comprehensive help
- Enterprise Ready: Multi-profile, audit logging, CI/CD integration

## Current Status
- ✅ Architecture documents: v0.1 (foundation) + v0.2 (enhanced) analyzed
- ✅ Project structure: Complete Go module with pkg/, internal/, cmd/ organization
- ✅ Build system: Makefile with cross-platform builds, CI/CD pipeline configured
- ✅ Documentation: README, CONTRIBUTING, LICENSE, implementation plan established
- ✅ Core framework: Cobra CLI, Viper config, basic API client foundation
- 🎯 Next: Begin Phase 1.1 - Authentication system and domain operations
- Target: Go 1.21+, cross-platform (Linux/macOS/Windows)

## Key Technical Decisions
- Framework: Cobra + Viper for CLI and configuration
- Authentication: HTTP Basic with API key, OS keyring for secure storage
- Output: Multiple formats (table/JSON/YAML/CSV) with stable ordering
- Caching: 5-minute TTL for API responses, 1-hour auth sessions
- Error Handling: Structured errors with actionable suggestions

## Development Phases
1. Foundation (Phase 1): Core auth, CRUD operations, basic output
2. Enhancement (Phase 2): Bulk ops, templates, interactive features  
3. Ecosystem (Phase 3): CI/CD, plugins, documentation
4. Enterprise (Phase 4): Advanced security, compliance, automation

## Competitive Advantage
Forward Email has **zero official CLI tools** despite comprehensive API. First-mover advantage with $3/month cost-effectiveness and developer-aligned values.

## Implementation Plan
See [IMPLEMENTATION_PLAN.md](../IMPLEMENTATION_PLAN.md) for detailed development roadmap with phases:
1. Foundation (Phase 1): Core functionality and architecture
2. Enhancement (Phase 2): Professional features and developer experience  
3. Ecosystem (Phase 3): Community integration and plugin system
4. Enterprise (Phase 4): Advanced features and automation

## Development Rules
- **Status Updates**: README development status section MUST be updated with each significant changeset
- **Documentation Sync**: All architectural decisions and status changes must be reflected in both CLAUDE.md and README.md
- **Phase Tracking**: When moving between implementation phases, update current phase in both memory and README
- **Feature Completion**: Mark features as completed (✅) only when fully implemented with tests