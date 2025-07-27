# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Mory (モリー) is an MCP (Model Context Protocol) server that adds personal memory functionality to Claude. It enables persistent memory across conversations, similar to ChatGPT's memory feature, allowing for more personalized interactions. The name comes from a hedgehog metaphor - small but able to hold many memories safely.

## Development Commands

This is a Go project currently in MVP Phase 1. The codebase structure is planned but not yet implemented.

### Planned Build Commands (from README)
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
```

## Architecture & Structure

### Planned Project Structure
```
mory/
├── cmd/mory/main.go           # Entry point
├── internal/
│   ├── memory/
│   │   ├── store.go           # Memory storage logic
│   │   ├── search.go          # Search implementation
│   │   └── types.go           # Type definitions
│   ├── mcp/
│   │   ├── server.go          # MCP server implementation
│   │   └── handlers.go        # Tool handlers
│   └── config/config.go       # Configuration management
├── data/memories.json         # Local storage (git-ignored)
└── test/
```

### Core Data Model
```go
type Memory struct {
    ID        string    `json:"id"`
    Category  string    `json:"category"`
    Key       string    `json:"key"`
    Value     string    `json:"value"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

### MCP Tools (MVP Phase 1)
1. **save_memory**: Save information with category, key, value
2. **get_memory**: Retrieve information by exact key match
3. **list_memories**: List all or category-filtered memories

### Key Design Principles
- **Privacy First**: All data stored locally in JSON files
- **Explicit Control**: Phase 1 only saves when explicitly instructed ("覚えて", "記憶して", "メモして")
- **Simple Storage**: JSON file with file locking for concurrent access
- **Go Standards**: Uses Go 1.21+ with standard project layout

### Development Phases
- **Phase 1 (Current)**: Basic key-value storage with explicit save commands
- **Phase 2**: Search functionality + confirmation-based suggestions
- **Phase 3**: Semantic search + automatic categorization
- **Phase 4**: Management UI + bulk operations

### Integration
Claude Desktop configuration:
```json
{
  "mcpServers": {
    "mory": {
      "command": "/path/to/mory/bin/mory"
    }
  }
}
```

## Development Notes

- Project uses Japanese documentation and naming conventions
- No actual Go code exists yet - only planning documentation
- Focus on MVP simplicity over complex features
- Reliability prioritized over feature richness
- All design decisions should be documented