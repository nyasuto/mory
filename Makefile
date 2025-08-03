# Makefile for Mory Server - FastAPI Implementation

.PHONY: help install dev test lint format type-check quality clean run docker-build docker-run setup-hooks uninstall-hooks

# Default target
.DEFAULT_GOAL := help

help: ## Show available commands
	@echo "Mory Server - FastAPI Development Commands"
	@echo ""
	@echo "Setup:"
	@echo "  install     - Install dependencies with uv"
	@echo "  dev         - Install development dependencies"
	@echo ""
	@echo "Development:"
	@echo "  run         - Run development server"
	@echo "  test        - Run tests"
	@echo "  lint        - Run ruff linter"
	@echo "  format      - Format code with ruff"
	@echo "  type-check  - Run mypy type checking"
	@echo "  quality     - Run all quality checks"
	@echo ""
	@echo "Docker:"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run with Docker Compose"
	@echo ""
	@echo "Git Hooks:"
	@echo "  setup-hooks     - Install pre-commit hooks"
	@echo "  uninstall-hooks - Remove pre-commit hooks"
	@echo ""
	@echo "Utilities:"
	@echo "  clean       - Clean cache and build files"

install: ## Install dependencies
	@echo "Installing dependencies with uv..."
	uv sync

dev: install ## Install development dependencies
	@echo "Development environment ready"

run: ## Run development server
	@echo "Starting Mory Server..."
	uv run uvicorn app.main:app --reload --host 0.0.0.0 --port 8080

test: ## Run tests
	@echo "Running tests..."
	uv run pytest -v

test-cov: ## Run tests with coverage
	@echo "Running tests with coverage..."
	uv run pytest --cov=app --cov-report=html --cov-report=term

lint: ## Run ruff linter
	@echo "Running ruff linter..."
	uv run ruff check .

format: ## Format code with ruff
	@echo "Formatting code with ruff..."
	uv run ruff format .

format-check: ## Check code formatting
	@echo "Checking code formatting..."
	uv run ruff format --check .

type-check: ## Run mypy type checking
	@echo "Running mypy type checking..."
	uv run mypy app/

quality: lint format-check type-check ## Run all quality checks
	@echo "All quality checks completed"

clean: ## Clean cache and build files
	@echo "Cleaning cache and build files..."
	find . -type d -name "__pycache__" -exec rm -rf {} + 2>/dev/null || true
	find . -type d -name "*.egg-info" -exec rm -rf {} + 2>/dev/null || true
	find . -type d -name ".pytest_cache" -exec rm -rf {} + 2>/dev/null || true
	find . -type d -name ".mypy_cache" -exec rm -rf {} + 2>/dev/null || true
	find . -type d -name ".ruff_cache" -exec rm -rf {} + 2>/dev/null || true
	rm -rf dist/
	rm -rf build/
	rm -rf htmlcov/
	@echo "Clean completed"

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t mory-server .

docker-run: ## Run with Docker Compose
	@echo "Starting with Docker Compose..."
	docker-compose -f docker-compose.dev.yml up --build

docker-stop: ## Stop Docker Compose
	@echo "Stopping Docker Compose..."
	docker-compose -f docker-compose.dev.yml down

# Health check target
health: ## Check if server is running
	@echo "Checking server health..."
	curl -f http://localhost:8080/api/health || echo "Server is not running"

setup-hooks: ## Install pre-commit hooks
	@echo "Setting up Git hooks..."
	./scripts/setup-hooks.sh

uninstall-hooks: ## Remove pre-commit hooks
	@echo "Removing Git hooks..."
	@if [ -f .git/hooks/pre-commit ]; then \
		rm .git/hooks/pre-commit && echo "✅ Pre-commit hook removed"; \
	else \
		echo "ℹ️  No pre-commit hook found"; \
	fi