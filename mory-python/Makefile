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
	uv pip install -r requirements.txt

install-dev:
	uv pip install -r requirements-dev.txt

# Testing
test:
	pytest tests/ -v

test-coverage:
	pytest tests/ -v --cov=mory --cov-report=html --cov-report=term-missing

# Code quality
lint:
	ruff check src tests

format:
	ruff format src tests

format-check:
	ruff format --check src tests

type-check:
	mypy src/mory

# Run all quality checks
quality: lint format-check type-check test

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
	python src/mory/main.py

# uv-specific commands
uv-install:
	uv pip install -e .

uv-sync:
	uv lock
	uv pip sync requirements.txt

uv-sync-dev:
	uv lock
	uv pip sync requirements-dev.txt