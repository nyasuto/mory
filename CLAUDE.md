# CLAUDE.md

Mory - FastAPI-based MCP server for Claude Desktop personal memory functionality.

## 🔧 Development Tools

**必須: Serena MCPツールの使用**

このプロジェクトではSerenaのMCPツールを優先的に使用してください:
- **シンボル検索・参照**: `find_symbol`, `find_referencing_symbols`
- **コード編集・置換**: `replace_symbol_body`, `insert_after_symbol`
- **プロジェクト把握**: `get_symbols_overview`, `list_dir`

## Quick Commands

```bash
# Development
make run              # Start server
make quality          # All checks (lint, format, test)
make setup-hooks      # Install pre-commit hooks

# Docker
docker-compose up -d  # Production server
curl http://localhost:8080/api/health  # Health check
```

## Project Structure

```
app/
├── main.py          # FastAPI app entry
├── api/             # Route handlers
├── core/            # Config, database
├── models/          # SQLAlchemy models  
├── services/        # Business logic
└── templates/       # HTML templates
```

## Key Features

**MCP Tools**: save_memory, get_memory, list_memories, search_memories, obsidian_import, generate_obsidian_note

**Web Dashboard**: http://localhost:8080/dashboard

## Development Rules

- **Branch Strategy**: feat/, fix/, docs/ - no direct main commits
- **Quality Checks**: Auto-run via pre-commit hooks
- **Testing**: 95%+ coverage required
- **Documentation**: Japanese localization maintained

## Technical Stack

- Python 3.11+ / FastAPI / SQLAlchemy / SQLite + FTS5
- Jinja2 templates / Docker ready
- Phase 2 Complete - Production ready

Emergency bypass: `git commit --no-verify`