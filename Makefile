# Makefile for Mory - Personal Memory MCP Server

# Variables
BINARY_NAME=mory
BINARY_PATH=bin/$(BINARY_NAME)
CMD_PATH=./cmd/mory
MAIN_GO=./cmd/mory/main.go

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOFMT=$(GOCMD) fmt
GOMOD=$(GOCMD) mod

# Build flags
LDFLAGS=-ldflags "-w -s"
BUILD_FLAGS=-v
FTS5_TAGS=-tags "sqlite_fts5"
STANDARD_TAGS=

# Default target
.PHONY: all
all: clean fmt test build

# Build the binary (with FTS5 support by default)
.PHONY: build
build:
	@echo "Building $(BINARY_NAME) with FTS5 support..."
	@mkdir -p bin
	$(GOBUILD) $(BUILD_FLAGS) $(FTS5_TAGS) $(LDFLAGS) -o $(BINARY_PATH) $(CMD_PATH)
	@echo "Binary built: $(BINARY_PATH)"

# Build without FTS5 (fallback for compatibility)
.PHONY: build-standard
build-standard:
	@echo "Building $(BINARY_NAME) without FTS5 (fallback mode)..."
	@mkdir -p bin
	$(GOBUILD) $(BUILD_FLAGS) $(STANDARD_TAGS) $(LDFLAGS) -o $(BINARY_PATH) $(CMD_PATH)
	@echo "Binary built: $(BINARY_PATH)"

# Run in development mode (with FTS5 by default)
.PHONY: run
run:
	@echo "Running $(BINARY_NAME) in development mode with FTS5..."
	$(GOCMD) run $(FTS5_TAGS) $(CMD_PATH)

# Run in development mode without FTS5
.PHONY: run-standard
run-standard:
	@echo "Running $(BINARY_NAME) in development mode without FTS5..."
	$(GOCMD) run $(STANDARD_TAGS) $(CMD_PATH)

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting Go code..."
	$(GOFMT) ./...

# Run linter (if golangci-lint is available)
.PHONY: lint
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found, running go vet instead..."; \
		$(GOCMD) vet ./...; \
	fi

# Run tests (with FTS5 by default)
.PHONY: test
test:
	@echo "Running tests with FTS5 support..."
	$(GOTEST) $(FTS5_TAGS) -v ./...

# Run tests without FTS5
.PHONY: test-standard
test-standard:
	@echo "Running tests without FTS5..."
	$(GOTEST) $(STANDARD_TAGS) -v ./...

# Run tests with coverage (FTS5 enabled)
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage (FTS5 enabled)..."
	$(GOTEST) $(FTS5_TAGS) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run tests with coverage (standard mode)
.PHONY: test-coverage-standard
test-coverage-standard:
	@echo "Running tests with coverage (standard mode)..."
	$(GOTEST) $(STANDARD_TAGS) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -f $(BINARY_PATH)
	@rm -f coverage.out coverage.html
	@echo "Clean completed"

# Tidy go modules
.PHONY: tidy
tidy:
	@echo "Tidying Go modules..."
	$(GOMOD) tidy

# Download dependencies
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download

# Install the binary to GOPATH/bin
.PHONY: install
install:
	@echo "Installing $(BINARY_NAME)..."
	$(GOCMD) install $(CMD_PATH)

# Migration tools for semantic search
.PHONY: migrate-dry-run
migrate-dry-run:
	@echo "Running migration dry-run (preview only)..."
	$(GOCMD) run migrate_embeddings.go --dry-run

.PHONY: migrate-embeddings
migrate-embeddings:
	@echo "Migrating existing memories to add embeddings..."
	@echo "⚠️  This will call OpenAI API and may incur costs"
	@read -p "Continue? [y/N] " confirm && [ "$$confirm" = "y" ] || exit 1
	$(GOCMD) run migrate_embeddings.go

# Development setup
.PHONY: dev-setup
dev-setup: deps tidy
	@echo "Development setup completed"
	@echo "Available commands:"
	@echo "  make build             - Build the binary"
	@echo "  make run               - Run in development mode"
	@echo "  make test              - Run tests"
	@echo "  make fmt               - Format code"
	@echo "  make lint              - Run linter"
	@echo "  make clean             - Clean build artifacts"
	@echo "  make migrate-dry-run   - Preview embedding migration"
	@echo "  make migrate-embeddings - Migrate existing memories (costs apply)"

# Quality checks
.PHONY: quality
quality: fmt lint test
	@echo "Quality checks completed"

# Release build (optimized with FTS5)
.PHONY: release
release: clean fmt test
	@echo "Building release version with FTS5..."
	@mkdir -p bin
	CGO_ENABLED=1 $(GOBUILD) $(FTS5_TAGS) $(LDFLAGS) -a -o $(BINARY_PATH) $(CMD_PATH)
	@echo "Release binary built: $(BINARY_PATH)"

# Release build without FTS5 (for environments without CGO)
.PHONY: release-standard
release-standard: clean fmt test-standard
	@echo "Building release version without FTS5 (CGO disabled)..."
	@mkdir -p bin
	CGO_ENABLED=0 $(GOBUILD) $(STANDARD_TAGS) $(LDFLAGS) -a -installsuffix cgo -o $(BINARY_PATH) $(CMD_PATH)
	@echo "Release binary built: $(BINARY_PATH)"

# Install git hooks
.PHONY: install-hooks
install-hooks:
	@echo "Installing git hooks..."
	@if [ ! -f .git/hooks/pre-commit ]; then \
		echo "Error: Git hooks not found. Please run this from the repository root."; \
		exit 1; \
	fi
	@echo "✓ Pre-commit hook installed"
	@echo ""
	@echo "Git hooks are now active and will run automatically on commit."
	@echo "The hooks will:"
	@echo "  - Format code (go fmt)"
	@echo "  - Run linter (golangci-lint)"
	@echo "  - Run tests"
	@echo "  - Tidy modules"
	@echo ""
	@echo "To bypass hooks (not recommended): git commit --no-verify"

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo ""
	@echo "FTS5-enabled builds (default):"
	@echo "  build          - Build the binary with FTS5 support"
	@echo "  run            - Run in development mode with FTS5"
	@echo "  test           - Run tests with FTS5 support"
	@echo "  test-coverage  - Run tests with coverage report (FTS5)"
	@echo "  release        - Build optimized release binary (FTS5)"
	@echo ""
	@echo "Standard builds (fallback without FTS5):"
	@echo "  build-standard     - Build without FTS5 support"
	@echo "  run-standard       - Run in development mode without FTS5"
	@echo "  test-standard      - Run tests without FTS5"
	@echo "  test-coverage-standard - Run tests with coverage (no FTS5)"
	@echo "  release-standard   - Build release binary without FTS5"
	@echo ""
	@echo "Development:"
	@echo "  fmt            - Format Go code"
	@echo "  lint           - Run linter"
	@echo "  clean          - Clean build artifacts"
	@echo "  tidy           - Tidy Go modules"
	@echo "  deps           - Download dependencies"
	@echo "  install        - Install binary to GOPATH/bin"
	@echo "  dev-setup      - Setup development environment"
	@echo "  quality        - Run all quality checks (fmt, lint, test)"
	@echo "  install-hooks  - Install git hooks for quality checks"
	@echo "  help           - Show this help"