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
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
GOFMT := $(GOCMD) fmt

.PHONY: all build clean test coverage deps lint install uninstall help

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
	rm -f coverage.out coverage.html

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v -race ./...

# Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Run linters
lint:
	@echo "Running linters..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, running basic checks..."; \
		$(GOFMT) -l . | grep -E '\.go$$' && exit 1 || true; \
		$(GOCMD) vet ./...; \
	fi

# Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...

# Install the binary to GOBIN
install: build
	@echo "Installing $(BINARY_NAME)..."
	cp $(BUILD_DIR)/$(BINARY_NAME) $(shell go env GOBIN)/$(BINARY_NAME)

# Uninstall the binary
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	rm -f $(shell go env GOBIN)/$(BINARY_NAME)

# Development setup
dev-setup:
	@echo "Setting up development environment..."
	$(GOGET) -u github.com/golangci/golangci-lint/cmd/golangci-lint
	$(GOMOD) download

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
	@echo "Available targets:"
	@echo "  all         - Clean, download deps, test, and build"
	@echo "  build       - Build the binary"
	@echo "  build-all   - Build for multiple platforms"
	@echo "  clean       - Clean build artifacts"
	@echo "  test        - Run tests"
	@echo "  coverage    - Run tests with coverage"
	@echo "  deps        - Download dependencies"
	@echo "  lint        - Run linters"
	@echo "  fmt         - Format code"
	@echo "  install     - Install binary to GOBIN"
	@echo "  uninstall   - Remove binary from GOBIN"
	@echo "  dev-setup   - Setup development environment"
	@echo "  run         - Build and run (use ARGS=... for arguments)"
	@echo "  completions - Generate shell completions"
	@echo "  help        - Show this help"
