"""Health check endpoints for Mory Server
Basic status and system information
"""

from datetime import datetime
from typing import Any

from fastapi import APIRouter, Depends
from sqlalchemy import text
from sqlalchemy.orm import Session

from ..core.config import settings
from ..core.database import check_fts5_support, get_db

router = APIRouter()


@router.get("/health")
async def health_check() -> dict[str, Any]:
    """Basic health check endpoint"""
    return {
        "status": "healthy",
        "timestamp": datetime.utcnow().isoformat(),
        "version": "1.0.0-alpha",
        "service": "mory-server",
    }


@router.get("/health/detailed")
async def detailed_health_check(db: Session = Depends(get_db)) -> dict[str, Any]:
    """Detailed health check with system information"""
    # Test database connection
    try:
        db.execute(text("SELECT 1"))
        db_status = "connected"
    except Exception as e:
        db_status = f"error: {str(e)}"

    # Check FTS5 support
    fts5_available = check_fts5_support()

    return {
        "status": "healthy" if db_status == "connected" else "degraded",
        "timestamp": datetime.utcnow().isoformat(),
        "version": "1.0.0-alpha",
        "service": "mory-server",
        "components": {
            "database": {
                "status": db_status,
                "type": "sqlite",
                "url": settings.sqlite_url,
                "fts5_support": fts5_available,
            },
            "semantic_search": {
                "enabled": settings.semantic_search_enabled,
                "available": settings.is_semantic_available,
                "model": settings.openai_model if settings.is_semantic_available else None,
            },
            "obsidian": {
                "vault_path": settings.obsidian_vault_path,
                "configured": settings.obsidian_vault_path is not None,
            },
        },
        "configuration": {
            "host": settings.host,
            "port": settings.port,
            "debug": settings.debug,
            "data_dir": settings.data_dir,
        },
    }
