# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Mory (モリー) is an MCP (Model Context Protocol) server that adds personal memory functionality to Claude. It enables persistent memory across conversations, similar to ChatGPT's memory feature, allowing for more personalized interactions. The name comes from a hedgehog metaphor - small but able to hold many memories safely.

**Status**: ✅ MVP Phase 1 complete and ready for Claude Desktop integration.

## Development Commands

This is a Go project with completed MVP Phase 1 implementation. All core functionality is working and tested.

### Available Build Commands
```bash
# Install dependencies
go mod download

# Build the project
make build

# Run in development mode
make run

# Code quality
make fmt    # Format code
make lint   # Run linter
make test   # Run tests

# Additional commands
make quality      # Run all quality checks (fmt, lint, test)
make test-coverage # Generate test coverage report
make clean        # Clean build artifacts
```

## Architecture & Structure

### Current Project Structure (Implemented)
```
mory/
├── cmd/mory/main.go           # Entry point
├── internal/
│   ├── memory/
│   │   ├── store.go           # Memory storage logic (JSON implementation)
│   │   ├── store_test.go      # Storage tests
│   │   ├── types.go           # Memory and OperationLog type definitions
│   │   └── types_test.go      # Type tests
│   ├── mcp/
│   │   ├── server.go          # MCP server implementation with all tools
│   │   └── server_test.go     # MCP server tests
│   └── config/
│       ├── config.go          # Configuration management
│       └── config_test.go     # Config tests
├── data/                      # Local storage directory (git-ignored)
│   ├── memories.json          # Memory data storage
│   └── operations.log         # Operation log file
├── bin/mory                   # Built binary
├── Makefile                   # Build automation
├── QUICKSTART.md              # POC setup guide
└── coverage.out/.html         # Test coverage reports
```

### Core Data Model
```go
type Memory struct {
    ID        string    `json:"id"`        // Auto-generated: memory_20250127123456
    Category  string    `json:"category"`  // User-defined category
    Key       string    `json:"key"`       // Optional user-friendly alias
    Value     string    `json:"value"`     // Stored content
    Tags      []string  `json:"tags"`      // Related tags (for future search)
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

type OperationLog struct {
    Timestamp   time.Time `json:"timestamp"`
    OperationID string    `json:"operation_id"`
    Operation   string    `json:"operation"`    // save, update, delete
    Key         string    `json:"key,omitempty"`
    Before      *Memory   `json:"before,omitempty"`
    After       *Memory   `json:"after,omitempty"`
    Success     bool      `json:"success"`
    Error       string    `json:"error,omitempty"`
}
```

### MCP Tools (MVP Phase 1 - Implemented)
1. **save_memory**: Save information with category, key, value (✅ Complete)
2. **get_memory**: Retrieve information by exact key/ID match (✅ Complete)
3. **list_memories**: List all or category-filtered memories (✅ Complete)

### Key Design Principles
- **Privacy First**: All data stored locally in JSON files
- **Explicit Control**: Phase 1 only saves when explicitly instructed ("覚えて", "記憶して", "メモして")
- **Simple Storage**: JSON file with file locking for concurrent access
- **Go Standards**: Uses Go 1.21+ with standard project layout

### Development Phases
- **Phase 1 (✅ Complete)**: Basic key-value storage with explicit save commands
- **Phase 2 (Planned)**: Search functionality + confirmation-based suggestions
- **Phase 3 (Planned)**: Semantic search + automatic categorization
- **Phase 4 (Planned)**: Management UI + bulk operations

### Integration
Claude Desktop configuration:
```json
{
  "mcpServers": {
    "mory": {
      "command": "/full/path/to/mory/bin/mory"
    }
  }
}
```

## Implementation Status

### ✅ Completed Features
- Complete MCP server implementation (`internal/mcp/server.go`)
- JSON-based memory storage with file locking (`internal/memory/store.go`)
- All MVP Phase 1 tools (save_memory, get_memory, list_memories)
- Comprehensive test suite (100% of planned tests passing)
- Build automation with Makefile
- Configuration management (`internal/config/config.go`)
- Operation logging for audit trail
- Error handling and input validation

### 🚧 Current Task
- Claude Desktop integration testing (see QUICKSTART.md)

### 📋 Next Steps (Phase 2)
- Enhanced search capabilities
- Automatic categorization suggestions
- Improved user experience based on POC feedback

## Development Notes

- Project uses Japanese documentation and naming conventions
- ✅ Full Go implementation complete for MVP Phase 1
- Focus on MVP simplicity over complex features
- Reliability prioritized over feature richness
- All design decisions documented in code and tests
- Ready for production use with Claude Desktop