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

# Default target
.PHONY: all
all: clean fmt test build

# Build the binary
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin
	$(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BINARY_PATH) $(CMD_PATH)
	@echo "Binary built: $(BINARY_PATH)"

# Run in development mode
.PHONY: run
run:
	@echo "Running $(BINARY_NAME) in development mode..."
	$(GOCMD) run $(CMD_PATH)

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

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
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

# Development setup
.PHONY: dev-setup
dev-setup: deps tidy
	@echo "Development setup completed"
	@echo "Available commands:"
	@echo "  make build     - Build the binary"
	@echo "  make run       - Run in development mode"
	@echo "  make test      - Run tests"
	@echo "  make fmt       - Format code"
	@echo "  make lint      - Run linter"
	@echo "  make clean     - Clean build artifacts"

# Quality checks
.PHONY: quality
quality: fmt lint test
	@echo "Quality checks completed"

# Release build (optimized)
.PHONY: release
release: clean fmt test
	@echo "Building release version..."
	@mkdir -p bin
	CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -a -installsuffix cgo -o $(BINARY_PATH) $(CMD_PATH)
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
	@echo "  build          - Build the binary"
	@echo "  run            - Run in development mode"
	@echo "  fmt            - Format Go code"
	@echo "  lint           - Run linter"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  clean          - Clean build artifacts"
	@echo "  tidy           - Tidy Go modules"
	@echo "  deps           - Download dependencies"
	@echo "  install        - Install binary to GOPATH/bin"
	@echo "  dev-setup      - Setup development environment"
	@echo "  quality        - Run all quality checks (fmt, lint, test)"
	@echo "  release        - Build optimized release binary"
	@echo "  install-hooks  - Install git hooks for quality checks"
	@echo "  help           - Show this help"