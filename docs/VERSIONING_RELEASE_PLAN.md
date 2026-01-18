# Comprehensive Versioning and Release Implementation Plan

**Status**: Phase 1 planning complete, implementation pending Phase 1.4 ⏳  
**Current Version**: dev (no versioning system implemented yet)

## Overview

This document outlines the complete strategy for implementing versioning and release management for the forward-email CLI project. The plan establishes a professional, automated release process supporting multiple distribution channels.

## Implementation Status

### ✅ Completed (Phase 1.1-1.3)
- Core CLI functionality fully implemented
- All major commands operational and tested
- Comprehensive test coverage with 100+ passing tests
- Cross-platform compatibility verified
- Documentation system established

### ⏳ Pending (Phase 1.4+)
- Version management system implementation
- GitHub Actions CI/CD pipeline
- Automated release process with GoReleaser
- Multi-platform binary distribution
- Package manager integration (Homebrew, Chocolatey)

## 1. Semantic Versioning Strategy

### Version Format
- **Standard**: `v{MAJOR}.{MINOR}.{PATCH}` (e.g., v0.1.0, v1.0.0)
- **Pre-release**: `v{MAJOR}.{MINOR}.{PATCH}-{PRERELEASE}` (e.g., v0.1.0-alpha.1, v0.1.0-beta.2)
- **Initial Version**: Start with v0.1.0 (since project is in active development)

### Version Bumping Rules
- **MAJOR**: Breaking API changes, major CLI command restructuring
- **MINOR**: New features, new commands, backward-compatible changes
- **PATCH**: Bug fixes, documentation updates, minor improvements

### Pre-release Strategy
- **alpha**: Early development, unstable features
- **beta**: Feature complete, testing phase
- **rc**: Release candidate, final testing

### Pre-1.0 Development Policy
- **Breaking Changes**: Allowed in any release without deprecation notices
- **API Stability**: No backwards compatibility guarantees
- **CLI Interface**: Commands, flags, and output formats may change
- **Configuration**: Profile and config structure may evolve
- **Migration**: Users must adapt to changes between versions
- **Timeline**: v1.0.0 target establishes stability commitment

### Post-1.0 Stability Commitment
- **Semantic Versioning**: Strict adherence to SemVer after v1.0.0
- **Deprecation Cycle**: Minimum 1 major version warning period
- **Breaking Changes**: Only in major version releases
- **Migration Guides**: Provided for all breaking changes
- **LTS Support**: Consider LTS policy for enterprise users

## 2. Forward Email API Compatibility Strategy

### API Version Targeting
- **Current Target**: Forward Email API v1 (`/v1/` endpoints)
- **Base URL**: `https://api.forwardemail.net/v1/`
- **Versioning**: Forward Email uses path-based versioning

### API Change Monitoring
- **No Official Policy**: Forward Email lacks public API versioning/deprecation policy
- **Defensive Approach**: Monitor for breaking changes proactively
- **Error Handling**: Robust error handling for API changes
- **Version Pinning**: Consider API version pinning in future releases

### Compatibility Testing
- **Integration Tests**: Test against live API endpoints (rate-limited)
- **Mock Tests**: Use recorded responses for unit testing
- **Breaking Change Detection**: Automated tests to detect API changes
- **Fallback Strategies**: Graceful degradation for unsupported features

### Release Impact Assessment
- **API Changes**: Each release notes Forward Email API compatibility
- **Breaking Changes**: Document when API changes affect CLI functionality
- **Version Matrices**: Maintain compatibility matrix with API versions
- **User Communication**: Clear communication about API-related breaking changes

### Risk Mitigation
- **Defensive Coding**: Handle unexpected API responses gracefully
- **Feature Flags**: Toggle features based on API availability
- **Backward Compatibility**: Maintain compatibility with older API responses when possible
- **User Warnings**: Warn users about potential API incompatibilities

## 3. Version Management Infrastructure

### A. Fix Version Package Integration
- Update `internal/cmd/root.go` to use the centralized version package
- Add a dedicated `version` command with detailed output options
- Include version info in error reports and debug output

### B. CHANGELOG Management
- Create `CHANGELOG.md` following Keep a Changelog format
- Sections: Added, Changed, Deprecated, Removed, Fixed, Security
- Automate changelog generation using git-chglog or conventional commits

### C. Git Tag Strategy
- Protected tags on main branch only
- Signed tags for releases (GPG signing)
- Annotated tags with release notes

## 3. Release Automation with GoReleaser

### A. GoReleaser Configuration (`.goreleaser.yml`)
- Multi-platform builds (Linux, macOS, Windows)
- Architecture support (amd64, arm64, 386)
- Binary compression with UPX
- Checksums and signatures
- Docker image generation
- Homebrew tap formula updates
- Snap package generation
- DEB/RPM package creation

### B. GitHub Actions Release Workflow
- Trigger on tag push (v* pattern)
- Build and test validation
- GoReleaser execution
- Asset upload to GitHub Releases
- Container registry push
- Package repository updates

## 4. Distribution Channels

### A. GitHub Releases (Primary)
- Pre-built binaries for all platforms
- Installation scripts (install.sh, install.ps1)
- Checksums and signatures
- Release notes with changelog

### B. Package Managers
- **Homebrew**: Create ginsys/homebrew-tap repository
- **Scoop** (Windows): Submit to scoop-extras bucket
- **AUR** (Arch Linux): Create forward-email-cli package
- **Snap**: Publish to Snapcraft store
- **APT/YUM**: Host repository on GitHub Pages

### C. Container Distribution
- Docker Hub: ginsys/forward-email
- GitHub Container Registry: ghcr.io/ginsys/forward-email
- Multi-arch images (linux/amd64, linux/arm64)

### D. Go Install Support
- Ensure compatibility with `go install github.com/ginsys/forward-email/cmd/forward-email@latest`

## 5. Version Command Enhancement

### Features to Implement
- Basic output: version number only
- Verbose output: version, commit, build date, Go version, OS/arch
- JSON output for automation
- Update check functionality (compare with latest GitHub release)
- License information display

### Command Examples
```bash
forward-email version                    # Basic version
forward-email version --verbose          # Detailed info
forward-email version --json             # JSON output
forward-email version --check-update     # Check for updates
forward-email version --license          # Show license
```

## 6. CI/CD Pipeline Updates

### A. Development Workflow
- Branch protection rules for main
- Required checks: tests, linting, security scanning
- Automated version bumping via conventional commits
- Nightly builds for development branch

### B. Release Workflow
- Manual trigger with version input
- Automated testing across all platforms
- Security scanning (gosec, vulnerability checks)
- SBOM generation (Software Bill of Materials)
- Release candidate support

## 7. Documentation Updates

### A. Release Documentation
- `RELEASING.md`: Step-by-step release process
- Version compatibility matrix
- Migration guides for breaking changes
- Release checklist template

### B. User Documentation
- Installation instructions per platform
- Upgrade procedures
- Version-specific documentation
- API compatibility notes

## 8. Implementation Phases

### Phase 1: Foundation (Week 1)
1. Fix version package integration in root.go
2. Create dedicated version command
3. Initialize CHANGELOG.md with current changes
4. Create first git tag (v0.1.0-alpha.1)

### Phase 2: GoReleaser Setup (Week 1-2)
1. Create .goreleaser.yml configuration
2. Test local releases with goreleaser
3. Create GitHub Actions release workflow
4. Set up GPG signing for releases

### Phase 3: Distribution Channels (Week 2-3)
1. Create Homebrew tap repository
2. Set up Docker build pipeline
3. Configure Snap packaging
4. Create installation scripts

### Phase 4: Automation (Week 3-4)
1. Implement conventional commits
2. Set up automated changelog generation
3. Configure automated version bumping
4. Add update notification system

### Phase 5: Documentation & Polish (Week 4)
1. Write comprehensive release documentation
2. Update README with installation methods
3. Create migration guides
4. Set up release announcement templates
5. Create docs/development/versioning.md with user-facing versioning policy

## 9. File Structure Changes

```
.github/
  workflows/
    release.yml         # New: Release automation
    nightly.yml        # New: Nightly builds
.goreleaser.yml        # New: GoReleaser config
.chglog/               # New: Changelog generation config
  config.yml
  CHANGELOG.tpl.md
CHANGELOG.md           # New: Version history
RELEASING.md           # New: Release process docs
VERSION                # New: Current version file
internal/
  cmd/
    version.go         # New: Version command
scripts/
  install.sh          # New: Unix installation
  install.ps1         # New: Windows installation
  version-bump.sh     # New: Version bumping script
docs/
  development/
    versioning.md     # New: User-facing versioning policy
```

## 10. Testing Strategy

### Build Testing
- Test version injection at build time
- Validate multi-platform builds
- Test installation scripts on fresh systems

### Distribution Testing
- Verify package manager installations
- Test upgrade procedures
- Validate signature verification

### Automation Testing
- Test GoReleaser configuration
- Validate GitHub Actions workflows
- Test changelog generation

## 11. Security Considerations

### Release Security
- GPG signing of releases
- Checksums for all artifacts
- Secure build environment
- SBOM generation

### Distribution Security
- Package signing
- Secure download URLs
- Vulnerability scanning
- Supply chain security

## 12. Monitoring and Metrics

### Release Metrics
- Download statistics
- Platform distribution
- Version adoption rates
- Error reporting

### Quality Metrics
- Build success rates
- Test coverage
- Security scan results
- Performance benchmarks

## 13. Rollback Strategy

### Release Rollback
- Ability to unpublish releases
- Revert tag procedures
- Package manager rollback
- Communication plan

### Version Rollback
- Previous version availability
- Migration path documentation
- User notification process
- Support procedures

## 14. Future Enhancements

### Advanced Features
- Automatic update mechanism
- Beta channel support
- Feature flag integration
- Telemetry collection

### Distribution Expansion
- More package managers
- Enterprise distribution
- Cloud marketplace presence
- Integration partnerships

## Implementation Priority

1. **High Priority**: Version command, CHANGELOG, basic GoReleaser
2. **Medium Priority**: GitHub Actions, Homebrew tap, installation scripts
3. **Low Priority**: Additional package managers, advanced automation

## Current Status Summary

**Ready for Release Implementation**: The CLI is functionally complete with 17,151 LOC, 47 Go files, and comprehensive test coverage. The release automation system is the primary remaining component for Phase 1.4 completion.

---

*Last Updated: 2025-08-27 | Planning complete, implementation pending*