# Forward Email CLI Makefile

# Build information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-X github.com/ginsys/forward-email/internal/version.Version=$(VERSION) \
                     -X github.com/ginsys/forward-email/internal/version.Commit=$(COMMIT) \
                     -X github.com/ginsys/forward-email/internal/version.Date=$(DATE)"

# Build targets
BINARY_NAME := forward-email
BUILD_DIR := bin
MAIN_PACKAGE := ./cmd/forward-email

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test

# Test environment (avoid interactive keyring prompts)
# Disable colors for stable CI logs
TEST_ENV ?= FORWARDEMAIL_KEYRING_BACKEND=none FORWARDEMAIL_NO_COLOR=1
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
GOFMT := $(GOCMD) fmt

# Installation directory fallback when GOBIN is unset
BIN_DIR := $(shell go env GOBIN)
ifeq ($(BIN_DIR),)
BIN_DIR := $(shell go env GOPATH)/bin
endif

.PHONY: all build clean test test-ci test-quick test-unit test-race test-pkg test-verbose test-bench coverage deps lint lint-ci lint-fast fmt fmt-check pre-commit pre-commit-full install-hooks uninstall-hooks check check-all install uninstall dev-setup help help-test

# Default target
all: clean deps test build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)

# Build for multiple platforms
build-all: clean
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PACKAGE)
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PACKAGE)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PACKAGE)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PACKAGE)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PACKAGE)

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html auth_coverage* cmd_coverage.out

# Testing commands - aligned with CI workflow
test:
	@echo "Running tests with race detector (default)..."
	$(TEST_ENV) $(GOTEST) -v -race ./...

test-ci:
	@echo "Running tests exactly as CI does..."
	$(TEST_ENV) $(GOTEST) -v -race -covermode=atomic -coverpkg=./... -coverprofile coverage.out ./...

test-quick:
	@echo "Running quick tests without race detector..."
	$(TEST_ENV) $(GOTEST) -short ./...

test-unit:
	@echo "Running unit tests only..."
	$(TEST_ENV) $(GOTEST) -short -tags=unit ./...

test-race:
	@echo "Running tests with race detector..."
	$(TEST_ENV) $(GOTEST) -v -race ./...

test-pkg:
	@echo "Running tests for package $(PKG)..."
	@if [ -d "./pkg/$(PKG)" ]; then \
		$(TEST_ENV) $(GOTEST) -v ./pkg/$(PKG)/...; \
	fi
	@if [ -d "./internal/$(PKG)" ]; then \
		$(TEST_ENV) $(GOTEST) -v ./internal/$(PKG)/...; \
	fi
	@if [ ! -d "./pkg/$(PKG)" ] && [ ! -d "./internal/$(PKG)" ]; then \
		echo "Package $(PKG) not found in ./pkg/ or ./internal/"; \
		exit 1; \
	fi

test-verbose:
	@echo "Running tests with verbose output..."
	$(TEST_ENV) $(GOTEST) -v ./...

test-bench:
	@echo "Running benchmarks..."
	$(TEST_ENV) $(GOTEST) -bench=. -benchmem ./...

# Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	$(TEST_ENV) $(GOTEST) -v -race -covermode=atomic -coverpkg=./... -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Linting commands
lint:
	@echo "Running linters..."
	@if [ -n "$$CI" ]; then \
		golangci-lint run; \
	else \
		if command -v golangci-lint >/dev/null 2>&1; then \
				golangci-lint run || { \
					echo "golangci-lint failed; running basic checks..."; \
					gofmt -l . | grep -E '\.go$$' && exit 1 || true; \
					$(GOCMD) vet ./...; \
				}; \
		else \
				echo "golangci-lint not installed, running basic checks..."; \
				gofmt -l . | grep -E '\.go$$' && exit 1 || true; \
				$(GOCMD) vet ./...; \
		fi; \
	fi

lint-ci:
	@echo "Running linters exactly as CI does..."
	golangci-lint run --timeout=5m

lint-fast:
	@echo "Running fast linters for pre-commit..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout=1m 2>/dev/null || { \
			echo "golangci-lint failed; running basic checks..."; \
			gofmt -l . | grep -E '\.go$$' && exit 1 || true; \
			$(GOCMD) vet ./...; \
		}; \
	else \
		echo "golangci-lint not installed, running basic checks..."; \
		gofmt -l . | grep -E '\.go$$' && exit 1 || true; \
		$(GOCMD) vet ./...; \
	fi

# Formatting commands
fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...

fmt-check:
	@echo "Checking code formatting..."
	@test -z "$$(gofmt -l .)" || (echo "Files need formatting. Run 'make fmt'" && exit 1)

# Install the binary to GOBIN
install: build
	@echo "Installing $(BINARY_NAME)..."
	cp $(BUILD_DIR)/$(BINARY_NAME) $(BIN_DIR)/$(BINARY_NAME)

# Uninstall the binary
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	rm -f $(BIN_DIR)/$(BINARY_NAME)

# Pre-commit commands
pre-commit: fmt-check lint-fast test-quick
	@echo "✅ Pre-commit checks passed!"

pre-commit-full: fmt-check lint test-ci
	@echo "✅ Full pre-commit checks passed!"

# Git hook management
install-hooks:
	@echo "Installing git hooks..."
	@echo '#!/bin/sh' > .git/hooks/pre-commit
	@echo '# Auto-generated by make install-hooks' >> .git/hooks/pre-commit
	@echo 'echo "Running pre-commit checks..."' >> .git/hooks/pre-commit
	@echo 'make pre-commit' >> .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "✅ Pre-commit hook installed"

uninstall-hooks:
	@echo "Removing git hooks..."
	@rm -f .git/hooks/pre-commit
	@echo "✅ Pre-commit hook removed"

# Developer workflow commands
check: fmt-check lint-fast test-quick
	@echo "✅ Basic checks passed"

check-all: fmt-check lint test-ci
	@echo "✅ All checks passed"

# Development setup
dev-setup: deps install-hooks
	@echo "Setting up development environment..."
	@if command -v mise >/dev/null 2>&1; then \
		echo "Installing tools via mise (including golangci-lint v2.6.2)..."; \
		mise install || echo "⚠️  Some mise tools failed to install (non-critical)"; \
	else \
		echo "⚠️  mise not found. Install from https://mise.jdx.dev"; \
		echo "   Tools defined in mise.toml will not be installed."; \
	fi
	@echo "✅ Development environment ready"

# Run the application
run:
	@echo "Running $(BINARY_NAME)..."
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)
	$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)

# Generate shell completions
completions: build
	@echo "Generating shell completions..."
	@mkdir -p completions
	$(BUILD_DIR)/$(BINARY_NAME) completion bash > completions/$(BINARY_NAME).bash
	$(BUILD_DIR)/$(BINARY_NAME) completion zsh > completions/$(BINARY_NAME).zsh
	$(BUILD_DIR)/$(BINARY_NAME) completion fish > completions/$(BINARY_NAME).fish

# Show help
help:
	@echo "Forward Email CLI - Build & Development Commands"
	@echo ""
	@echo "🚀 Quick Start:"
	@echo "  dev-setup     - Setup development environment and git hooks"
	@echo "  check         - Run basic checks (fmt, lint-fast, test-quick)"
	@echo "  check-all     - Run all checks (fmt, lint, test-ci)"
	@echo ""
	@echo "🏗️  Build Commands:"
	@echo "  build         - Build the binary"
	@echo "  build-all     - Build for multiple platforms (Linux, macOS, Windows)"
	@echo "  clean         - Clean build artifacts"
	@echo "  install       - Install binary to GOBIN"
	@echo "  uninstall     - Remove binary from GOBIN"
	@echo ""
	@echo "🧪 Testing Commands:"
	@echo "  test          - Run tests with race detector (default)"
	@echo "  test-ci       - Run tests exactly as CI does (with coverage)"
	@echo "  test-quick    - Fast tests without race detector"
	@echo "  test-unit     - Unit tests only"
	@echo "  test-race     - Tests with race detector"
	@echo "  test-pkg PKG= - Test specific package (e.g., make test-pkg PKG=api)"
	@echo "  test-verbose  - Tests with verbose output"
	@echo "  test-bench    - Run benchmarks"
	@echo "  coverage      - Generate HTML coverage report"
	@echo ""
	@echo "🔍 Quality Commands:"
	@echo "  fmt           - Format code"
	@echo "  fmt-check     - Check if formatting needed (fails if unformatted)"
	@echo "  lint          - Run linters (smart: CI mode in CI, fallback otherwise)"
	@echo "  lint-ci       - Run linters exactly as CI does"
	@echo "  lint-fast     - Quick lint check for pre-commit"
	@echo ""
	@echo "🎯 Pre-commit Commands:"
	@echo "  pre-commit    - Run quick pre-commit checks (fmt-check, lint-fast, test-quick)"
	@echo "  pre-commit-full - Run full pre-commit checks (fmt-check, lint, test-ci)"
	@echo "  install-hooks - Install git pre-commit hook"
	@echo "  uninstall-hooks - Remove git pre-commit hook"
	@echo ""
	@echo "⚙️  Development Commands:"
	@echo "  all           - Clean, download deps, test, and build"
	@echo "  deps          - Download and tidy dependencies"
	@echo "  run ARGS=     - Build and run (e.g., make run ARGS='auth status')"
	@echo "  completions   - Generate shell completions (bash, zsh, fish)"
	@echo ""
	@echo "📚 Help Commands:"
	@echo "  help          - Show this help"
	@echo "  help-test     - Show detailed testing help"
	@echo ""
	@echo "🔗 CI/CD Alignment:"
	@echo "  The following commands match CI exactly:"
	@echo "  • make test-ci    = CI test job"
	@echo "  • make lint-ci    = CI lint job"
	@echo "  • make build-all  = CI build job"
	@echo ""
	@echo "🏷️  Release Commands:"
	@echo "  make release <type>        - Dry-run: show exact steps (no changes)"
	@echo "  make release-do <type>     - Execute steps shown by dry-run"
	@echo "    <type>: alpha | beta | stable | bump {minor|major}"
	@echo "      stable: tag base as stable, then bump to next patch alpha.0"

help-test:
	@echo "Testing Commands - Detailed Guide"
	@echo ""
	@echo "🧪 Basic Testing:"
	@echo "  make test         - Standard test run with race detector"
	@echo "  make test-quick   - Fast feedback loop (no race detector, -short flag)"
	@echo "  make test-verbose - Detailed test output for debugging"
	@echo ""
	@echo "🎯 Targeted Testing:"
	@echo "  make test-pkg PKG=api       - Test only pkg/api package"
	@echo "  make test-pkg PKG=cmd       - Test only internal/cmd package"
	@echo "  make test-unit              - Unit tests only (with -tags=unit)"
	@echo ""
	@echo "⚡ Performance Testing:"
	@echo "  make test-bench             - Run all benchmarks"
	@echo "  make test-race              - Race condition detection"
	@echo ""
	@echo "📊 Coverage & CI:"
	@echo "  make test-ci                - Exact CI test execution (with coverage)"
	@echo "  make coverage               - Generate HTML coverage report"
	@echo ""
	@echo "🔍 Pre-commit Testing:"
	@echo "  make pre-commit             - Quick checks before commit"
	@echo "  make pre-commit-full        - Comprehensive pre-commit checks"
	@echo ""
	@echo "Examples:"
	@echo "  make test-pkg PKG=auth      # Test authentication package"
	@echo "  make test-quick && make lint-fast  # Fast development cycle"
	@echo "  make test-ci                # Run exactly what CI runs"
# Version management (local only; does not push)
.PHONY: release release-do

release:
	@k="$(filter alpha beta stable bump,$(MAKECMDGOALS))"; \
	if [ "$$k" = "bump" ]; then \
		m="$(filter minor major,$(MAKECMDGOALS))"; \
		if [ -z "$$m" ]; then echo "usage: make release bump {minor|major}"; exit 1; fi; \
		DRY_RUN=1 sh scripts/semver.sh release bump "$$m"; \
	elif [ -n "$$k" ]; then \
		DRY_RUN=1 sh scripts/semver.sh release "$$k"; \
	else \
		kind=$$kind; \
		if [ -z "$$kind" ]; then echo "usage: make release {alpha|beta|stable} or 'make release bump {minor|major}'"; exit 1; fi; \
		DRY_RUN=1 sh scripts/semver.sh release "$$kind"; \
	fi

release-do:
	@k="$(filter alpha beta stable bump,$(MAKECMDGOALS))"; \
	if [ "$$k" = "bump" ]; then \
		m="$(filter minor major,$(MAKECMDGOALS))"; \
		if [ -z "$$m" ]; then echo "usage: make release-do bump {minor|major}"; exit 1; fi; \
		sh scripts/semver.sh release bump "$$m"; \
	elif [ -n "$$k" ]; then \
		sh scripts/semver.sh release "$$k"; \
	else \
		kind=$$kind; \
		if [ -z "$$kind" ]; then echo "usage: make release-do {alpha|beta|stable} or 'make release-do bump {minor|major}'"; exit 1; fi; \
		sh scripts/semver.sh release "$$kind"; \
	fi

# Positional-style usage examples:
#   make release alpha
#   make release beta
#   make release stable
#   make release bump minor
#   make release bump major

alpha beta stable bump minor major:
	@:
