"""Memory CRUD API endpoints"""

from datetime import datetime, timedelta

from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session

from ..core.database import get_db
from ..models.memory import Memory
from ..models.schemas import (
    MemoryCreate,
    MemoryListResponse,
    MemoryListSummaryResponse,
    MemoryResponse,
    MemoryStatsResponse,
    MemorySummaryResponse,
    MemoryUpdate,
    MessageResponse,
    SearchRequest,
    SearchResponse,
)
from ..services.embedding import embedding_service
from ..services.summarization import summarization_service

router = APIRouter()


@router.post("/memories", response_model=MemoryResponse, status_code=201)
async def save_memory(memory_data: MemoryCreate, db: Session = Depends(get_db)) -> MemoryResponse:
    """Save a new memory - simplified AI-driven schema (Issue #112)"""
    import traceback
    import uuid

    request_id = str(uuid.uuid4())[:8]
    errors = []  # Track non-fatal errors

    try:
        # Create new memory (each save creates a new memory in simplified schema)
        new_memory = Memory(
            value=memory_data.value,
        )

        # Generate AI summary and tags if enabled (Issue #112)
        if summarization_service.enabled:
            try:
                # Generate AI summary
                summary = await summarization_service.generate_summary(memory_data.value)
                new_memory.summary = summary

                # Generate comprehensive AI tags based on content
                # TODO: Implement AI tag generation service
                # For now, use improved keyword extraction supporting Japanese
                import re

                # Extract meaningful words (both English and Japanese)
                text = memory_data.value.lower()
                # Remove common markup and symbols
                text = re.sub(r'[#\*`\-_=+(){}\\[\]|<>"\';:.?,!]', " ", text)

                words = text.split()
                important_words = []

                for word in words:
                    # Include words with 2+ characters (for Japanese) or 3+ English letters
                    if len(word) >= 2 and (
                        word.isalpha()
                        or any(
                            "\u3040" <= c <= "\u309f"
                            or "\u30a0" <= c <= "\u30ff"
                            or "\u4e00" <= c <= "\u9faf"
                            for c in word
                        )
                    ):
                        important_words.append(word)

                ai_tags = list(set(important_words[:8]))  # Take up to 8 unique words as tags
                new_memory.tags_list = ai_tags

                new_memory.ai_processed_at = datetime.utcnow()
            except Exception as e:
                # If AI processing fails, continue without AI enhancements
                error_msg = f"AI processing failed: {str(e)} (request_id: {request_id})"
                print(error_msg)
                errors.append(
                    {
                        "stage": "ai_processing",
                        "error": str(e),
                        "error_type": type(e).__name__,
                        "recoverable": True,
                    }
                )
                new_memory.tags_list = []  # Empty tags if AI processing fails

        # Database save operation
        try:
            db.add(new_memory)
            db.commit()
            db.refresh(new_memory)
        except Exception as e:
            db.rollback()
            raise HTTPException(
                status_code=500,
                detail={
                    "error": "Database save failed",
                    "message": f"Failed to save memory to database: {str(e)}",
                    "error_type": type(e).__name__,
                    "stage": "database_save",
                    "request_id": request_id,
                    "recoverable": False,
                },
            ) from e

        # Generate vector embedding automatically (Issue #112 enhancement)
        if embedding_service.enabled:
            try:
                embedding_generated = await embedding_service.generate_embedding_for_memory(
                    new_memory
                )
                if embedding_generated:
                    db.commit()
                    db.refresh(new_memory)
            except Exception as e:
                error_msg = f"Embedding generation failed: {str(e)} (request_id: {request_id})"
                print(error_msg)
                errors.append(
                    {
                        "stage": "embedding_generation",
                        "error": str(e),
                        "error_type": type(e).__name__,
                        "recoverable": True,
                    }
                )

        # Add warnings to response if there were non-fatal errors
        response = MemoryResponse.model_validate(new_memory)
        if errors:
            # Add warning header for partial failures
            print(f"Memory saved with warnings (request_id: {request_id}): {errors}")

        return response

    except HTTPException:
        # Re-raise HTTP exceptions as-is
        raise
    except Exception as e:
        # Catch any unexpected errors
        db.rollback()
        error_trace = traceback.format_exc()
        print(f"Unexpected error saving memory (request_id: {request_id}): {error_trace}")

        raise HTTPException(
            status_code=500,
            detail={
                "error": "Unexpected error occurred",
                "message": f"An unexpected error occurred while saving memory: {str(e)}",
                "error_type": type(e).__name__,
                "stage": "unknown",
                "request_id": request_id,
                "recoverable": False,
                "suggestion": "Please try again. If the problem persists, contact support with the request_id.",
            },
        ) from e


@router.get("/memories/stats", response_model=MemoryStatsResponse)
async def get_memory_stats(db: Session = Depends(get_db)) -> MemoryStatsResponse:
    """Get memory statistics - simplified AI-driven schema (Issue #112)"""
    # Basic counts
    total_memories = db.query(Memory).count()

    # Recent memories (last 24 hours)
    yesterday = datetime.utcnow() - timedelta(days=1)
    recent_memories = db.query(Memory).filter(Memory.created_at >= yesterday).count()

    # AI-generated tags count
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
        total_categories=0,  # No categories in simplified schema
        total_tags=total_tags,
        categories={},  # No categories in simplified schema
        recent_memories=recent_memories,
        storage_info={
            "backend": "sqlite",
            "database_file": "memories.db",
            "supports_fts": True,
            "supports_semantic": False,  # Will be updated when semantic search is implemented
            "ai_driven": True,  # New: Indicates AI-driven tag and summary generation
        },
    )


@router.get("/memories/{memory_id}", response_model=MemoryResponse)
async def get_memory(
    memory_id: str,
    db: Session = Depends(get_db),
) -> MemoryResponse:
    """Get memory by ID - simplified AI-driven schema (Issue #112)"""
    memory = db.query(Memory).filter(Memory.id == memory_id).first()

    if not memory:
        raise HTTPException(
            status_code=404,
            detail=f"Memory with ID '{memory_id}' not found",
        )

    return MemoryResponse.model_validate(memory)


# Issue #111: Detail endpoint for full content access - simplified schema (Issue #112)
@router.get("/memories/{memory_id}/detail", response_model=MemoryResponse)
async def get_memory_detail(
    memory_id: str,
    db: Session = Depends(get_db),
) -> MemoryResponse:
    """Get full memory details by ID - simplified AI-driven schema (Issue #112)"""
    memory = db.query(Memory).filter(Memory.id == memory_id).first()

    if not memory:
        raise HTTPException(
            status_code=404,
            detail=f"Memory with ID '{memory_id}' not found",
        )

    return MemoryResponse.model_validate(memory)


# Issue #111: Optimized list endpoint - simplified AI-driven schema (Issue #112)
@router.get("/memories")
async def list_memories(
    limit: int = Query(100, ge=1, le=300, description="Maximum number of memories to return"),
    offset: int = Query(0, ge=0, description="Number of memories to skip"),
    include_full_text: bool = Query(
        False, description="Include full content (backward compatibility)"
    ),
    db: Session = Depends(get_db),
):
    """List memories with optimized responses - simplified AI-driven schema (Issue #112)"""
    query = db.query(Memory)

    # Get total count
    total = query.count()

    # Apply pagination and ordering
    memories = query.order_by(Memory.updated_at.desc()).offset(offset).limit(limit).all()

    # Return different response based on include_full_text parameter
    if include_full_text:
        # Backward compatibility: return full content
        return MemoryListResponse(
            memories=[MemoryResponse.model_validate(memory) for memory in memories],
            total=total,
        )
    else:
        # Optimized response: summary only
        summary_memories = []
        for memory in memories:
            # Create summary response with AI-generated summary or fallback
            summary = memory.summary
            if not summary:
                # Create very short fallback summary to prevent context overflow
                summary = (memory.value[:50] + "...") if len(memory.value) > 50 else memory.value

            summary_memory = MemorySummaryResponse(
                id=str(memory.id),
                tags=memory.tags_list or [],
                summary=str(summary) if summary else None,
                created_at=memory.created_at,
                updated_at=memory.updated_at,
                has_embedding=memory.has_embedding,
                processing_status=memory.processing_status,
            )
            summary_memories.append(summary_memory)

        return MemoryListSummaryResponse(
            memories=summary_memories,
            total=total,
        )


@router.delete("/memories/{memory_id}", response_model=MessageResponse)
async def delete_memory(
    memory_id: str,
    db: Session = Depends(get_db),
) -> MessageResponse:
    """Delete memory by ID - simplified AI-driven schema (Issue #112)"""
    memory = db.query(Memory).filter(Memory.id == memory_id).first()

    if not memory:
        raise HTTPException(
            status_code=404,
            detail=f"Memory with ID '{memory_id}' not found",
        )

    db.delete(memory)
    db.commit()

    return MessageResponse(
        message=f"Memory '{memory_id}' deleted successfully", data={"deleted_id": memory.id}
    )


@router.put("/memories/{memory_id}", response_model=MemoryResponse)
async def update_memory(
    memory_id: str,
    memory_update: MemoryUpdate,
    db: Session = Depends(get_db),
) -> MemoryResponse:
    """Update memory by ID - simplified AI-driven schema (Issue #112)"""
    import traceback
    import uuid

    request_id = str(uuid.uuid4())[:8]
    errors = []  # Track non-fatal errors

    try:
        memory = db.query(Memory).filter(Memory.id == memory_id).first()

        if not memory:
            raise HTTPException(
                status_code=404,
                detail={
                    "error": "Memory not found",
                    "message": f"Memory with ID '{memory_id}' not found",
                    "memory_id": memory_id,
                    "request_id": request_id,
                    "suggestion": "Please check the memory ID and ensure it exists",
                },
            )

        # Update value (only field that can be updated in simplified schema)
        update_data = memory_update.model_dump(exclude_unset=True)
        if "value" in update_data:
            memory.value = update_data["value"]

            # Re-process with AI when value changes
            if summarization_service.enabled:
                try:
                    # Regenerate AI summary
                    summary = await summarization_service.generate_summary(memory.value)
                    memory.summary = summary

                    # Regenerate comprehensive AI tags with improved Japanese support
                    import re

                    text = memory.value.lower()
                    text = re.sub(r'[#\*`\-_=+(){}\\[\]|<>"\';:.?,!]', " ", text)

                    words = text.split()
                    important_words = []

                    for word in words:
                        if len(word) >= 2 and (
                            word.isalpha()
                            or any(
                                "\u3040" <= c <= "\u309f"
                                or "\u30a0" <= c <= "\u30ff"
                                or "\u4e00" <= c <= "\u9faf"
                                for c in word
                            )
                        ):
                            important_words.append(word)

                    ai_tags = list(set(important_words[:8]))
                    memory.tags_list = ai_tags

                    memory.ai_processed_at = datetime.utcnow()
                except Exception as e:
                    error_msg = f"AI re-processing failed: {str(e)} (request_id: {request_id})"
                    print(error_msg)
                    errors.append(
                        {
                            "stage": "ai_reprocessing",
                            "error": str(e),
                            "error_type": type(e).__name__,
                            "recoverable": True,
                        }
                    )

            # Regenerate vector embedding when content changes
            if embedding_service.enabled:
                try:
                    await embedding_service.generate_embedding_for_memory(memory)
                except Exception as e:
                    error_msg = (
                        f"Embedding regeneration failed: {str(e)} (request_id: {request_id})"
                    )
                    print(error_msg)
                    errors.append(
                        {
                            "stage": "embedding_regeneration",
                            "error": str(e),
                            "error_type": type(e).__name__,
                            "recoverable": True,
                        }
                    )

            # Database update operation
            try:
                memory.updated_at = datetime.utcnow()
                db.commit()
                db.refresh(memory)
            except Exception as e:
                db.rollback()
                raise HTTPException(
                    status_code=500,
                    detail={
                        "error": "Database update failed",
                        "message": f"Failed to update memory in database: {str(e)}",
                        "error_type": type(e).__name__,
                        "stage": "database_update",
                        "memory_id": memory_id,
                        "request_id": request_id,
                        "recoverable": False,
                    },
                ) from e

        # Add warnings to response if there were non-fatal errors
        response = MemoryResponse.model_validate(memory)
        if errors:
            print(f"Memory updated with warnings (request_id: {request_id}): {errors}")

        return response

    except HTTPException:
        # Re-raise HTTP exceptions as-is
        raise
    except Exception as e:
        # Catch any unexpected errors
        db.rollback()
        error_trace = traceback.format_exc()
        print(f"Unexpected error updating memory (request_id: {request_id}): {error_trace}")

        raise HTTPException(
            status_code=500,
            detail={
                "error": "Unexpected error occurred",
                "message": f"An unexpected error occurred while updating memory: {str(e)}",
                "error_type": type(e).__name__,
                "stage": "unknown",
                "memory_id": memory_id,
                "request_id": request_id,
                "recoverable": False,
                "suggestion": "Please try again. If the problem persists, contact support with the request_id.",
            },
        ) from e


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
