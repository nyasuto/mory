"""FastAPI main application for Mory Server
Personal Memory Server with REST API
"""

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from .api.dashboard import router as dashboard_router
from .api.health import router as health_router
from .api.memories import router as memories_router
from .core.config import settings
from .core.database import create_tables

# Create FastAPI application
app = FastAPI(
    title="Mory Server",
    description="Personal Memory Server with Advanced Search",
    version="1.0.0-alpha",
    docs_url="/docs",
    redoc_url="/redoc",
)

# CORS middleware for development
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # In production, specify actual origins
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Include routers
app.include_router(health_router, prefix="/api", tags=["health"])
app.include_router(memories_router, prefix="/api", tags=["memories"])
app.include_router(dashboard_router, tags=["dashboard"])


@app.on_event("startup")
async def startup_event():
    """Initialize application on startup"""
    # Create database tables
    create_tables()

    print(f"🚀 Mory Server starting on {settings.host}:{settings.port}")
    print(f"📊 Database: {settings.sqlite_url}")
    print(f"🔍 Semantic Search: {'Enabled' if settings.is_semantic_available else 'Disabled'}")
    print(f"📝 Obsidian: {'Configured' if settings.obsidian_vault_path else 'Not configured'}")
    print(f"🌐 API Documentation: http://{settings.host}:{settings.port}/docs")


@app.on_event("shutdown")
async def shutdown_event():
    """Cleanup on application shutdown"""
    print("🛑 Mory Server shutting down")


@app.get("/")
async def root():
    """Root endpoint with basic information"""
    return {
        "service": "Mory Server",
        "version": "1.0.0-alpha",
        "description": "Personal Memory Server with Advanced Search",
        "documentation": "/docs",
        "health": "/api/health",
    }


if __name__ == "__main__":
    import uvicorn

    uvicorn.run("app.main:app", host=settings.host, port=settings.port, reload=settings.debug)
