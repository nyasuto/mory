"""Memory CRUD API endpoints"""

from datetime import datetime, timedelta

from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy import func
from sqlalchemy.orm import Session

from ..core.database import get_db
from ..models.memory import Memory
from ..models.schemas import (
    MemoryCreate,
    MemoryListResponse,
    MemoryResponse,
    MemoryStatsResponse,
    MemoryUpdate,
    MessageResponse,
    SearchRequest,
    SearchResponse,
)

router = APIRouter()


@router.post("/memories", response_model=MemoryResponse, status_code=201)
async def save_memory(memory_data: MemoryCreate, db: Session = Depends(get_db)) -> MemoryResponse:
    """Save a new memory or update existing one by key"""
    # Check if memory with this key already exists
    existing_memory = None
    if memory_data.key:
        existing_memory = (
            db.query(Memory)
            .filter(Memory.key == memory_data.key, Memory.category == memory_data.category)
            .first()
        )

    if existing_memory:
        # Update existing memory
        existing_memory.value = memory_data.value
        existing_memory.tags_list = memory_data.tags
        existing_memory.updated_at = datetime.utcnow()
        db.commit()
        db.refresh(existing_memory)
        return MemoryResponse.model_validate(existing_memory)

    # Create new memory
    new_memory = Memory(
        category=memory_data.category,
        key=memory_data.key,
        value=memory_data.value,
        tags_list=memory_data.tags,
    )

    db.add(new_memory)
    db.commit()
    db.refresh(new_memory)

    return MemoryResponse.model_validate(new_memory)


@router.get("/memories/stats", response_model=MemoryStatsResponse)
async def get_memory_stats(db: Session = Depends(get_db)) -> MemoryStatsResponse:
    """Get memory statistics"""
    # Basic counts
    total_memories = db.query(Memory).count()
    total_categories = db.query(func.count(func.distinct(Memory.category))).scalar()

    # Category breakdown
    category_counts = dict(
        db.query(Memory.category, func.count(Memory.id)).group_by(Memory.category).all()
    )

    # Recent memories (last 24 hours)
    yesterday = datetime.utcnow() - timedelta(days=1)
    recent_memories = db.query(Memory).filter(Memory.created_at >= yesterday).count()

    # Unique tags count (approximate)
    all_tags = []
    memories_with_tags = db.query(Memory.tags).filter(Memory.tags != "[]").all()
    for (tags_json,) in memories_with_tags:
        try:
            import json

            tags = json.loads(tags_json)
            all_tags.extend(tags)
        except json.JSONDecodeError:
            continue

    total_tags = len(set(all_tags))

    return MemoryStatsResponse(
        total_memories=total_memories,
        total_categories=total_categories,
        total_tags=total_tags,
        categories=category_counts,
        recent_memories=recent_memories,
        storage_info={
            "backend": "sqlite",
            "database_file": "memories.db",
            "supports_fts": True,
            "supports_semantic": False,  # Will be updated when semantic search is implemented
        },
    )


@router.get("/memories/{memory_key}", response_model=MemoryResponse)
async def get_memory(
    memory_key: str,
    category: str | None = Query(None, description="Filter by category"),
    db: Session = Depends(get_db),
) -> MemoryResponse:
    """Get memory by key"""
    query = db.query(Memory).filter(Memory.key == memory_key)

    if category:
        query = query.filter(Memory.category == category)

    memory = query.first()

    if not memory:
        raise HTTPException(
            status_code=404,
            detail=f"Memory with key '{memory_key}'"
            + (f" in category '{category}'" if category else "")
            + " not found",
        )

    return MemoryResponse.model_validate(memory)


@router.get("/memories", response_model=MemoryListResponse)
async def list_memories(
    category: str | None = Query(None, description="Filter by category"),
    limit: int = Query(100, ge=1, le=1000, description="Maximum number of memories to return"),
    offset: int = Query(0, ge=0, description="Number of memories to skip"),
    db: Session = Depends(get_db),
) -> MemoryListResponse:
    """List memories with optional filtering"""
    query = db.query(Memory)

    if category:
        query = query.filter(Memory.category == category)

    # Get total count
    total = query.count()

    # Apply pagination and ordering
    memories = query.order_by(Memory.updated_at.desc()).offset(offset).limit(limit).all()

    return MemoryListResponse(
        memories=[MemoryResponse.model_validate(memory) for memory in memories],
        total=total,
        category=category,
    )


@router.delete("/memories/{memory_key}", response_model=MessageResponse)
async def delete_memory(
    memory_key: str,
    category: str | None = Query(None, description="Filter by category"),
    db: Session = Depends(get_db),
) -> MessageResponse:
    """Delete memory by key"""
    query = db.query(Memory).filter(Memory.key == memory_key)

    if category:
        query = query.filter(Memory.category == category)

    memory = query.first()

    if not memory:
        raise HTTPException(
            status_code=404,
            detail=f"Memory with key '{memory_key}'"
            + (f" in category '{category}'" if category else "")
            + " not found",
        )

    db.delete(memory)
    db.commit()

    return MessageResponse(
        message=f"Memory '{memory_key}' deleted successfully", data={"deleted_id": memory.id}
    )


@router.put("/memories/{memory_key}", response_model=MemoryResponse)
async def update_memory(
    memory_key: str,
    memory_update: MemoryUpdate,
    category: str | None = Query(None, description="Filter by category"),
    db: Session = Depends(get_db),
) -> MemoryResponse:
    """Update memory by key"""
    query = db.query(Memory).filter(Memory.key == memory_key)

    if category:
        query = query.filter(Memory.category == category)

    memory = query.first()

    if not memory:
        raise HTTPException(
            status_code=404,
            detail=f"Memory with key '{memory_key}'"
            + (f" in category '{category}'" if category else "")
            + " not found",
        )

    # Update fields
    update_data = memory_update.model_dump(exclude_unset=True)
    for field, value in update_data.items():
        if field == "tags":
            memory.tags_list = value
        else:
            setattr(memory, field, value)

    memory.updated_at = datetime.utcnow()
    db.commit()
    db.refresh(memory)

    return MemoryResponse.model_validate(memory)


@router.post("/memories/search", response_model=SearchResponse)
async def search_memories(
    search_request: SearchRequest,
    db: Session = Depends(get_db),
) -> SearchResponse:
    """Advanced memory search with FTS5 and semantic search support"""
    from ..services.search import search_service

    try:
        return await search_service.search_memories(search_request, db)
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Search failed: {str(e)}") from e
