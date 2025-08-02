# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Mory is an MCP server that adds personal memory functionality to Claude Desktop. Now powered by Python implementation.

**Status**: ✅ Python Implementation Complete - Core features implemented and ready for use

## Development Commands

This is a Python project with core MCP functionality implemented. Memory management and search features are working and tested.

### Available Build Commands
```bash
# Install dependencies
pip install -e .

# Run in development mode
python main.py

# Code quality
make fmt        # Format code with ruff
make lint       # Run linter with ruff
make test       # Run tests with pytest

# Additional commands
make quality    # Run all quality checks (fmt, lint, test)
make install    # Install in development mode
make clean      # Clean build artifacts
```

## Architecture & Structure

### Project Structure
```
mory/
├── src/mory/                  # Main Python package
│   ├── __init__.py
│   ├── main.py                # Application entry point
│   ├── server.py              # MCP server implementation
│   ├── memory.py              # Memory models and types
│   └── storage.py             # JSON storage implementation
├── tests-python/             # Test suite
├── data/                      # Local data storage (git-ignored)
├── main.py                    # Root entry point
├── pyproject.toml             # Python project configuration
├── QUICKSTART.md              # Complete setup guide
├── API.md                     # Technical documentation
└── README.md                  # Project overview
```

### Core Types
See [API.md](./API.md) for complete data model specifications.

### MCP Tools (5 tools implemented)

1. **save_memory** - Store information with categories and tags
2. **get_memory** - Retrieve memories by key or ID  
3. **list_memories** - List all memories with optional filtering
4. **search_memories** - Full-text search with relevance scoring
5. **delete_memory** - Delete memories by key or ID

See [API.md](./API.md) for detailed tool specifications and parameters.

## Setup & Integration

**Claude Desktop**: See [QUICKSTART.md](./QUICKSTART.md) for complete setup instructions.

**Data Directory**: Set `MORY_DATA_DIR` environment variable or use default 'data' directory.

## Current Status

✅ **Python Implementation Complete**: Core features implemented and production-ready
- Core memory management with JSON storage
- Full-text search with relevance scoring  
- Async/await support for MCP protocol
- Type hints and Pydantic models
- Basic test coverage

📋 **Future Enhancements**: Obsidian integration, semantic search, advanced templates

## Development Notes

### Technical Highlights
- **Python 3.11+**: Modern Python with type hints and async support
- **Pydantic Models**: Type-safe data validation and serialization
- **Async/Await**: Non-blocking I/O for MCP server operations
- **JSON Storage**: Human-readable file-based data persistence

### Usage Examples
See [QUICKSTART.md](./QUICKSTART.md) for complete usage examples and setup instructions.
See [API.md](./API.md) for detailed technical specifications.

## Important Reminders
- Focus on requested functionality only
- Prefer editing existing files over creating new ones  
- Only create documentation when explicitly requested
- **Japanese Localization**: All documentation (README.md, QUICKSTART.md, API.md) has been rewritten in Japanese for better accessibility to Japanese users. Maintain this localization in future updates.

# important-instruction-reminders
Do what has been asked; nothing more, nothing less.
NEVER create files unless they're absolutely necessary for achieving your goal.
ALWAYS prefer editing an existing file to creating a new one.
NEVER proactively create documentation files (*.md) or README files. Only create documentation files if explicitly requested by the User.