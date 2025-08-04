"""Dashboard API for memory management"""

from fastapi import APIRouter, Depends, HTTPException, Request
from fastapi.responses import HTMLResponse
from fastapi.templating import Jinja2Templates
from sqlalchemy.orm import Session

from ..core.database import get_db
from ..models.memory import Memory

router = APIRouter()
templates = Jinja2Templates(directory="app/templates")


@router.get("/dashboard", response_class=HTMLResponse)
async def dashboard(request: Request, db: Session = Depends(get_db)):
    """Memory management dashboard"""
    # Get all memories with basic stats
    memories = db.query(Memory).order_by(Memory.updated_at.desc()).all()

    # Calculate stats
    total_memories = len(memories)
    memories_with_embeddings = sum(1 for m in memories if m.has_embedding)
    ai_processed = sum(1 for m in memories if m.is_ai_processed)

    stats = {
        "total_memories": total_memories,
        "memories_with_embeddings": memories_with_embeddings,
        "ai_processed": ai_processed,
        "pending_processing": total_memories - ai_processed,
    }

    return templates.TemplateResponse(
        "dashboard.html",
        {
            "request": request,
            "memories": memories,
            "stats": stats,
        },
    )


@router.delete("/dashboard/memories/{memory_id}")
async def delete_memory_api(memory_id: str, db: Session = Depends(get_db)):
    """Delete a memory via dashboard"""
    memory = db.query(Memory).filter(Memory.id == memory_id).first()
    if not memory:
        raise HTTPException(status_code=404, detail="Memory not found")

    db.delete(memory)
    db.commit()

    return {"success": True, "message": f"Memory {memory_id} deleted successfully"}


@router.get("/dashboard/api/memories")
async def get_memories_api(db: Session = Depends(get_db)):
    """Get all memories for dashboard API"""
    memories = db.query(Memory).order_by(Memory.updated_at.desc()).all()

    return {
        "memories": [
            {
                **memory.to_dict(),
                "value_preview": memory.value[:100] + "..."
                if len(memory.value) > 100
                else memory.value,
                "created_at_formatted": memory.created_at.strftime("%Y-%m-%d %H:%M")
                if memory.created_at
                else None,
                "updated_at_formatted": memory.updated_at.strftime("%Y-%m-%d %H:%M")
                if memory.updated_at
                else None,
            }
            for memory in memories
        ],
        "total": len(memories),
    }
