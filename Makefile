# Makefile for Mory Python - Modern Python toolchain with uv and ruff

.PHONY: help install install-dev test lint format type-check quality clean run

# Default target
help:
	@echo "Available commands:"
	@echo "  install      - Install dependencies using uv"
	@echo "  install-dev  - Install development dependencies"
	@echo "  test         - Run tests with pytest"
	@echo "  lint         - Run ruff linter"
	@echo "  format       - Format code with ruff"
	@echo "  type-check   - Run mypy type checking"
	@echo "  quality      - Run all quality checks (lint, format-check, type-check, test)"
	@echo "  clean        - Clean build artifacts"
	@echo "  run          - Run the MCP server"

# Installation
install:
	uv sync

install-dev:
	uv sync --extra dev

# Testing
test:
	uv run pytest tests-python/ -v

test-coverage:
	uv run pytest tests-python/ -v --cov=src/mory --cov-report=html --cov-report=term-missing

# Code quality
lint:
	uv run ruff check src tests-python

format:
	uv run ruff format src tests-python

format-check:
	uv run ruff format --check src tests-python

type-check:
	uv run mypy src/mory

# Run all quality checks
quality: lint format-check

# Cleanup
clean:
	rm -rf build/
	rm -rf dist/
	rm -rf *.egg-info/
	rm -rf .pytest_cache/
	rm -rf .mypy_cache/
	rm -rf .ruff_cache/
	rm -rf htmlcov/
	find . -type d -name __pycache__ -exec rm -rf {} +
	find . -type f -name "*.pyc" -delete

# Development
run:
	uv run python main.py

# uv-specific commands (advanced)
uv-sync:
	uv sync

uv-sync-dev:
	uv sync --all-extras