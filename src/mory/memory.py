"""Memory data models and storage interface."""

from __future__ import annotations

from datetime import datetime
from typing import Protocol
from uuid import uuid4

from pydantic import BaseModel, Field


class Memory(BaseModel):
    """Memory represents a stored memory item."""

    id: str = Field(
        default_factory=lambda: f"memory_{int(datetime.now().timestamp() * 1_000_000)}"
    )
    category: str = Field(..., description="Category of the memory")
    key: str = Field(default="", description="Optional user-friendly alias")
    value: str = Field(..., description="The actual memory content")
    tags: list[str] = Field(
        default_factory=list, description="Related tags for future search"
    )
    created_at: datetime = Field(default_factory=datetime.now)
    updated_at: datetime = Field(default_factory=datetime.now)

    # Semantic search fields (will be added in Phase 2)
    embedding: list[float] | None = Field(
        default=None, description="Semantic embedding vector"
    )

    class Config:
        """Pydantic configuration."""

        json_encoders = {datetime: lambda v: v.isoformat()}

    def update_timestamp(self) -> None:
        """Update the updated_at timestamp."""
        self.updated_at = datetime.now()


class SearchResult(BaseModel):
    """Search result with relevance score."""

    memory: Memory
    score: float = Field(..., ge=0.0, le=1.0, description="Relevance score (0.0 - 1.0)")

    class Config:
        """Pydantic configuration."""

        json_encoders = {datetime: lambda v: v.isoformat()}


class SearchQuery(BaseModel):
    """Search query parameters."""

    query: str = Field(..., description="Search query string")
    category: str | None = Field(
        default=None, description="Optional category filter"
    )
    limit: int = Field(
        default=20, ge=1, le=100, description="Maximum number of results"
    )
    min_score: float = Field(
        default=0.0, ge=0.0, le=1.0, description="Minimum relevance score"
    )


class OperationLog(BaseModel):
    """Log entry for memory operations."""

    timestamp: datetime = Field(default_factory=datetime.now)
    operation_id: str = Field(default_factory=lambda: f"op_{uuid4().hex[:8]}")
    operation: str = Field(..., description="Operation type (save, get, delete, etc.)")
    key: str | None = Field(default=None, description="Memory key if applicable")
    before: Memory | None = Field(
        default=None, description="Memory state before operation"
    )
    after: Memory | None = Field(
        default=None, description="Memory state after operation"
    )
    success: bool = Field(default=True, description="Whether operation succeeded")
    error: str | None = Field(
        default=None, description="Error message if operation failed"
    )

    class Config:
        """Pydantic configuration."""

        json_encoders = {datetime: lambda v: v.isoformat()}


class MemoryNotFoundError(Exception):
    """Raised when a memory is not found."""

    def __init__(self, key: str) -> None:
        self.key = key
        super().__init__(f"Memory not found: {key}")


class MemoryStore(Protocol):
    """Interface for memory storage operations."""

    async def save(self, memory: Memory) -> str:
        """Save a memory and return its ID."""
        ...

    async def get(self, key: str) -> Memory:
        """Get a memory by key or ID."""
        ...

    async def get_by_id(self, id: str) -> Memory:
        """Get a memory by ID."""
        ...

    async def list(self, category: str | None = None) -> list[Memory]:
        """List memories, optionally filtered by category."""
        ...

    async def search(self, query: SearchQuery) -> list[SearchResult]:
        """Search memories."""
        ...

    async def delete(self, key: str) -> None:
        """Delete a memory by key."""
        ...

    async def delete_by_id(self, id: str) -> None:
        """Delete a memory by ID."""
        ...

    async def log_operation(self, log: OperationLog) -> None:
        """Log an operation."""
        ...
