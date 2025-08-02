# Mory Server - FastAPI Implementation

Personal Memory Server with advanced search capabilities built with FastAPI and SQLite.

## 🚀 Quick Start

### Prerequisites

- Python 3.11+
- [uv](https://github.com/astral-sh/uv) for dependency management
- Docker (optional, for containerized development)

### Development Setup

1. **Clone and navigate to the server directory:**
   ```bash
   cd mory-server
   ```

2. **Install dependencies with uv:**
   ```bash
   # Install dependencies and create virtual environment
   uv sync

   # Or create virtual environment manually
   uv venv
   source .venv/bin/activate  # On Windows: .venv\Scripts\activate
   uv pip install -e .
   ```

3. **Set up environment variables:**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. **Run the development server:**
   ```bash
   # With uv
   uv run uvicorn app.main:app --reload --host 0.0.0.0 --port 8080

   # Or with activated venv
   uvicorn app.main:app --reload --host 0.0.0.0 --port 8080
   ```

5. **Access the application:**
   - API: http://localhost:8080
   - Documentation: http://localhost:8080/docs
   - Health Check: http://localhost:8080/api/health

### Docker Development

```bash
# Build and run with Docker Compose
docker-compose -f docker-compose.dev.yml up --build

# With database browser (optional)
docker-compose -f docker-compose.dev.yml --profile debug up
```

## 🛠️ Development Tools

### Code Quality

```bash
# Format code with ruff
uv run ruff format .

# Lint code
uv run ruff check .

# Type checking
uv run mypy app/

# Run all quality checks
uv run ruff check . && uv run ruff format --check . && uv run mypy app/
```

### Testing

```bash
# Run tests
uv run pytest

# Run tests with coverage
uv run pytest --cov=app --cov-report=html

# Run specific test file
uv run pytest tests/test_health.py -v
```

## 📁 Project Structure

```
mory-server/
├── app/                     # Application source code
│   ├── __init__.py
│   ├── main.py             # FastAPI application
│   ├── core/               # Core functionality
│   │   ├── config.py       # Configuration management
│   │   └── database.py     # Database setup
│   ├── models/             # SQLAlchemy models
│   │   └── memory.py       # Memory model
│   └── api/                # API routes
│       └── health.py       # Health check endpoints
├── tests/                  # Test files
├── data/                   # Database and data files (created at runtime)
# ├── requirements.txt removed (using pyproject.toml with uv)
├── pyproject.toml         # Project configuration
├── Dockerfile             # Container configuration
├── docker-compose.dev.yml # Development environment
└── README.md              # This file
```

## ⚙️ Configuration

Configuration is managed through environment variables and `.env` files:

### Core Settings
- `MORY_HOST`: Server host (default: 0.0.0.0)
- `MORY_PORT`: Server port (default: 8080)
- `MORY_DEBUG`: Debug mode (default: false)
- `MORY_DATA_DIR`: Data directory path (default: data)

### Database
- `MORY_DATABASE_URL`: Full database URL (optional, auto-generated if not set)

### Semantic Search
- `OPENAI_API_KEY`: OpenAI API key for embeddings
- `MORY_SEMANTIC_SEARCH_ENABLED`: Enable semantic search (default: true)
- `MORY_OPENAI_MODEL`: OpenAI model (default: text-embedding-3-large)
- `MORY_HYBRID_SEARCH_WEIGHT`: Semantic vs keyword search weight (default: 0.7)

### Obsidian Integration
- `MORY_OBSIDIAN_VAULT_PATH`: Path to Obsidian vault

## 🔍 API Endpoints

### Health & Status
- `GET /`: Basic service information
- `GET /api/health`: Basic health check
- `GET /api/health/detailed`: Detailed system status

### Documentation
- `GET /docs`: Interactive API documentation (Swagger UI)
- `GET /redoc`: Alternative API documentation (ReDoc)

## 🚧 Development Status

This is **Phase 1-1** of the Mory server migration from CLI to web API architecture.

### ✅ Completed
- FastAPI application setup
- SQLAlchemy database configuration
- Basic health check endpoints
- Docker development environment
- Modern tooling (uv, ruff)
- Project structure and documentation

### 🔄 Next Steps (Phase 1-2)
- Memory CRUD API implementation
- Request/response models
- Basic testing suite
- Data validation

### 🔮 Future Phases
- Phase 1-3: Search functionality with FTS5
- Phase 1-4: Data migration from CLI version
- Phase 2: Obsidian integration
- Phase 3: Web UI

## 🤝 Contributing

1. Use `uv` for dependency management
2. Format code with `ruff format`
3. Lint with `ruff check`
4. Add tests for new functionality
5. Update documentation as needed

## 📄 License

MIT License - see the main project LICENSE file.