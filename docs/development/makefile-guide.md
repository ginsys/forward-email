# Makefile Guide

This guide explains how to use the Forward Email CLI Makefile for development, testing, and CI/CD alignment.

## Quick Start

### First-Time Setup
```bash
# Clone the repository
git clone https://github.com/ginsys/forward-email.git
cd forward-email

# Setup development environment
make dev-setup      # Downloads dependencies, installs git hooks
```

### Daily Development Workflow
```bash
# Quick feedback loop
make check          # fmt-check + lint-fast + test-quick
make test-quick     # Fast tests for development

# Before committing
make pre-commit     # Quick pre-commit validation
# or
make pre-commit-full  # Full validation (same as CI)

# Full validation
make check-all      # fmt-check + lint + test-ci
```

## Command Categories

### 🚀 Quick Start Commands

| Command | Description | Use Case |
|---------|-------------|----------|
| `make dev-setup` | Setup development environment | First-time setup |
| `make check` | Basic checks (fmt, lint-fast, test-quick) | Quick validation |
| `make check-all` | Full checks (fmt, lint, test-ci) | Pre-commit validation |

### 🧪 Testing Commands

| Command | Description | Equivalent CI/Manual |
|---------|-------------|---------------------|
| `make test` | Standard tests with race detector | Default for development |
| `make test-ci` | Exact CI test execution | GitHub Actions test job |
| `make test-quick` | Fast tests, no race detector | Quick feedback loop |
| `make test-unit` | Unit tests only | Targeted unit testing |
| `make test-verbose` | Tests with verbose output | Debugging test failures |
| `make test-bench` | Run benchmarks | Performance testing |
| `make test-pkg PKG=<name>` | Test specific package | Targeted testing |

#### Testing Examples
```bash
# Test specific packages
make test-pkg PKG=api       # Test pkg/api only
make test-pkg PKG=auth      # Test pkg/auth only
make test-pkg PKG=cmd       # Test internal/cmd only

# Development workflows
make test-quick             # Fast feedback
make test-ci                # CI simulation
make test-bench             # Performance testing
```

### 🔍 Quality Commands

| Command | Description | CI Equivalent |
|---------|-------------|---------------|
| `make fmt` | Format code | Manual formatting |
| `make fmt-check` | Check formatting (fails if needed) | CI formatting check |
| `make lint` | Smart linting (CI-aware) | Context-dependent |
| `make lint-ci` | Exact CI linting | GitHub Actions lint job |
| `make lint-fast` | Quick lint for pre-commit | Pre-commit hook |

### 🎯 Pre-commit Commands

| Command | Description | When to Use |
|---------|-------------|-------------|
| `make pre-commit` | Quick pre-commit checks | Before every commit |
| `make pre-commit-full` | Full pre-commit validation | Important commits |
| `make install-hooks` | Install git pre-commit hook | One-time setup |
| `make uninstall-hooks` | Remove git hooks | If needed |

### 🏗️ Build Commands

| Command | Description | Output |
|---------|-------------|--------|
| `make build` | Build single binary | `bin/forward-email` |
| `make build-all` | Multi-platform builds | All platform binaries |
| `make clean` | Clean build artifacts | Removes `bin/`, coverage files |

## CI/CD Alignment

The Makefile is designed to provide exact CI/local parity:

### GitHub Actions ↔ Makefile Mapping

| GitHub Actions Step | Makefile Command | Purpose |
|-------------------|------------------|---------|
| Download dependencies | `make deps` | Dependency management |
| Check formatting | `make fmt-check` | Code formatting validation |
| Run tests | `make test-ci` | Testing with coverage |
| Run linting | `make lint-ci` | Code quality validation |
| Build artifacts | `make build-all` | Multi-platform builds |

### Local CI Simulation
```bash
# Run exactly what CI runs
make deps
make fmt-check
make test-ci
make lint-ci
make build-all

# Or use convenience command
make check-all  # Runs fmt-check, lint, test-ci
```

## Pre-commit Hook Integration

### Automatic Installation
```bash
make dev-setup  # Includes hook installation
# or manually
make install-hooks
```

### What the Hook Does
The installed pre-commit hook runs:
```bash
make pre-commit  # fmt-check + lint-fast + test-quick
```

### Hook Behavior
- **Fails**: If formatting, linting, or tests fail
- **Success**: All checks pass, commit proceeds
- **Performance**: Optimized for speed (~10-30 seconds)

### Manual Pre-commit Validation
```bash
# Quick validation (what hook runs)
make pre-commit

# Full validation (same as CI)
make pre-commit-full
```

## Package-Specific Testing

### Testing Individual Packages
```bash
# Core packages
make test-pkg PKG=api       # pkg/api (HTTP client, services)
make test-pkg PKG=auth      # pkg/auth (authentication)
make test-pkg PKG=config    # pkg/config (configuration)
make test-pkg PKG=errors    # pkg/errors (error handling)
make test-pkg PKG=output    # pkg/output (formatting)

# Internal packages
make test-pkg PKG=cmd       # internal/cmd (CLI commands)
make test-pkg PKG=client    # internal/client (client wrapper)
make test-pkg PKG=keyring   # internal/keyring (OS keyring)
```

### Common Development Patterns
```bash
# Working on API package
make test-pkg PKG=api       # Test changes
make lint-fast              # Quick lint check
# Repeat cycle

# Before committing API changes
make test-pkg PKG=api       # Final package test
make pre-commit             # Full pre-commit check
```

## Performance and Optimization

### Command Performance Characteristics

| Command | Speed | Use Case |
|---------|-------|----------|
| `make test-quick` | ~5-15s | Development loop |
| `make lint-fast` | ~3-10s | Pre-commit |
| `make pre-commit` | ~10-30s | Pre-commit validation |
| `make test-ci` | ~30-60s | Full validation |
| `make check-all` | ~45-90s | Complete check |

### Optimization Tips
```bash
# Fast development cycle
make test-quick && make lint-fast  # ~8-25s total

# Targeted testing during development
make test-pkg PKG=api              # Test only what you're changing

# Pre-commit optimization
make pre-commit                    # Quick checks only
```

## Troubleshooting

### Common Issues

#### Formatting Failures
```bash
# Check what needs formatting
make fmt-check
# Fix formatting
make fmt
```

#### Lint Failures
```bash
# Run full linting with details
make lint-ci
# Quick lint for common issues
make lint-fast
```

#### Test Failures
```bash
# Verbose test output
make test-verbose
# Test specific package
make test-pkg PKG=<failing-package>
```

#### Pre-commit Hook Issues
```bash
# Reinstall hooks
make uninstall-hooks && make install-hooks
# Manual pre-commit check
make pre-commit
```

### CI Failures Debugging

If CI fails but local tests pass:

1. **Check CI alignment**:
   ```bash
   make test-ci    # Exact CI test command
   make lint-ci    # Exact CI lint command
   ```

2. **Verify dependencies**:
   ```bash
   make deps       # Update dependencies
   ```

3. **Check formatting**:
   ```bash
   make fmt-check  # CI formatting check
   ```

## Advanced Usage

### Custom Test Execution
```bash
# Environment variables
PKG=api make test-pkg              # Test pkg/api
ARGS="auth status" make run        # Run with arguments
```

### Development Workflows

#### Feature Development
```bash
# Start feature
make test-quick                    # Baseline check

# Development loop
make test-pkg PKG=<your-package>   # Test your changes
make lint-fast                     # Quick lint

# Pre-commit
make pre-commit                    # Final validation
```

#### Bug Fixing
```bash
# Reproduce issue
make test-verbose                  # Detailed test output
make test-pkg PKG=<affected>       # Focus on affected package

# Fix and validate
make test-pkg PKG=<affected>       # Test fix
make test-ci                       # Full validation
```

#### Release Preparation
```bash
# Full validation
make check-all                     # Complete validation
make build-all                     # Multi-platform builds
make coverage                      # Coverage report
```

## Integration with IDEs

### VS Code
Add tasks to `.vscode/tasks.json`:
```json
{
    "version": "2.0.0",
    "tasks": [
        {
            "label": "test-quick",
            "type": "shell",
            "command": "make test-quick",
            "group": "test"
        },
        {
            "label": "pre-commit",
            "type": "shell", 
            "command": "make pre-commit",
            "group": "test"
        }
    ]
}
```

### GoLand/IntelliJ
Configure run configurations for common make targets.

## References

- [Testing Strategy](testing.md) - Comprehensive testing guide
- [Contributing Guide](contributing.md) - Contribution workflow
- [Architecture Overview](architecture.md) - System architecture
- [GitHub Actions CI](.github/workflows/ci.yml) - CI configuration