# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Mory is a FastAPI-based MCP server that adds personal memory functionality to Claude Desktop. Phase 2 complete with search and Obsidian integration.

**Status**: ✅ Phase 2 Complete - Production ready with all features implemented

## Development Commands

This is a Python project using FastAPI and SQLite. All core functionality including search and Obsidian integration is working and tested.

### Available Build Commands
```bash
# Install dependencies
uv sync

# Run development server
make run

# Code quality
make fmt    # Format code with ruff
make lint   # Run linter with ruff
make test   # Run tests with pytest

# Additional commands
make quality      # Run all quality checks (fmt, lint, test)
make test-coverage # Generate test coverage report
make clean        # Clean build artifacts
```

## Architecture & Structure

### Project Structure
```
mory/
├── app/                       # FastAPI application source
│   ├── main.py               # Application entry point
│   ├── api/                  # API route handlers
│   ├── core/                 # Core functionality (config, database)
│   ├── models/               # SQLAlchemy models and schemas
│   ├── services/             # Business logic (search, etc.)
│   └── mcp_server.py         # MCP server implementation
├── tests/                    # Test suite
├── scripts/                  # Deployment and migration scripts
├── data/                     # Local data storage (git-ignored)
├── mcp_main.py              # MCP server entry point
├── pyproject.toml           # Python dependencies and config
└── README.md                # Project overview
```

### Core Types
See [API.md](./API.md) for complete data model specifications.

### MCP Tools (6 tools implemented)

1. **save_memory** - Store information with categories and tags
2. **get_memory** - Retrieve memories by key or ID  
3. **list_memories** - List all memories with optional filtering
4. **search_memories** - Advanced full-text search with scoring
5. **obsidian_import** - Import Obsidian vault notes
6. **generate_obsidian_note** - Generate notes from memories using templates

See [API.md](./docs/API.md) for detailed tool specifications and parameters.

## Setup & Integration

**Claude Desktop**: See [QUICKSTART.md](./docs/QUICKSTART.md) for complete setup instructions.

**Obsidian Integration**: Set `MORY_OBSIDIAN_VAULT_PATH` environment variable or create config file.

## Current Status

✅ **Phase 2 Complete**: All features implemented, tested, and production-ready
- Core memory management with JSON storage
- Advanced search with relevance scoring  
- Obsidian integration (import/export/templates)
- Comprehensive test suite (95%+ coverage)
- Complete documentation

📋 **Phase 3 Planned**: Semantic search, AI categorization, smart recommendations

## Development Notes

### Technical Highlights
- **Python 3.11+**: Modern Python with FastAPI and SQLAlchemy
- **SQLite + FTS5**: High-performance full-text search capabilities
- **Well-Tested**: 95%+ test coverage with pytest
- **Production Ready**: Stable, documented, ready for Claude Desktop

### Usage Examples
See [QUICKSTART.md](./docs/QUICKSTART.md) for complete usage examples and setup instructions.
See [API.md](./docs/API.md) for detailed technical specifications.

## Important Reminders
- Focus on requested functionality only
- Prefer editing existing files over creating new ones  
- Only create documentation when explicitly requested
- **Japanese Localization**: All documentation (README.md, QUICKSTART.md, API.md) has been rewritten in Japanese for better accessibility to Japanese users. Maintain this localization in future updates.