"""Pydantic schemas for request/response models"""

import json
from datetime import datetime
from typing import Any

from pydantic import BaseModel, Field, field_validator


class MemoryBase(BaseModel):
    """Base memory model with common fields"""

    category: str = Field(..., description="Memory category for organization")
    key: str | None = Field(None, description="User-friendly key for the memory")
    value: str = Field(..., description="Memory content")
    tags: list[str] = Field(default_factory=list, description="Tags for categorization")


class MemoryCreate(MemoryBase):
    """Request model for creating memories"""

    @field_validator("category")
    @classmethod
    def validate_category(cls, v):
        if not v or not v.strip():
            raise ValueError("Category cannot be empty")
        return v.strip()

    @field_validator("value")
    @classmethod
    def validate_value(cls, v):
        if not v or not v.strip():
            raise ValueError("Value cannot be empty")
        return v.strip()

    @field_validator("key")
    @classmethod
    def validate_key(cls, v):
        if v is not None:
            v = v.strip()
            if not v:
                return None
        return v


class MemoryUpdate(BaseModel):
    """Request model for updating memories"""

    category: str | None = Field(None, description="Updated category")
    key: str | None = Field(None, description="Updated key")
    value: str | None = Field(None, description="Updated content")
    tags: list[str] | None = Field(None, description="Updated tags")

    @field_validator("category")
    @classmethod
    def validate_category(cls, v):
        if v is not None and (not v or not v.strip()):
            raise ValueError("Category cannot be empty")
        return v.strip() if v else v

    @field_validator("value")
    @classmethod
    def validate_value(cls, v):
        if v is not None and (not v or not v.strip()):
            raise ValueError("Value cannot be empty")
        return v.strip() if v else v


class MemoryResponse(MemoryBase):
    """Response model for memory data"""

    id: str = Field(..., description="Unique memory identifier")
    created_at: datetime = Field(..., description="Creation timestamp")
    updated_at: datetime = Field(..., description="Last update timestamp")
    has_embedding: bool = Field(False, description="Whether memory has semantic embedding")

    @field_validator("tags", mode="before")
    @classmethod
    def parse_tags(cls, v):
        """Parse tags from JSON string if needed"""
        if isinstance(v, str):
            try:
                return json.loads(v)
            except json.JSONDecodeError:
                return []
        elif isinstance(v, list):
            return v
        return []

    model_config = {"from_attributes": True}


class MemoryListResponse(BaseModel):
    """Response model for memory lists"""

    memories: list[MemoryResponse] = Field(..., description="List of memories")
    total: int = Field(..., description="Total number of memories")
    category: str | None = Field(None, description="Filtered category")


class MemoryStatsResponse(BaseModel):
    """Response model for memory statistics"""

    total_memories: int = Field(..., description="Total number of memories")
    total_categories: int = Field(..., description="Number of unique categories")
    total_tags: int = Field(..., description="Number of unique tags")
    categories: dict[str, int] = Field(..., description="Memory count per category")
    recent_memories: int = Field(..., description="Memories created in last 24 hours")
    storage_info: dict[str, Any] = Field(..., description="Storage backend information")


class ErrorResponse(BaseModel):
    """Standard error response model"""

    error: str = Field(..., description="Error type")
    message: str = Field(..., description="Human-readable error message")
    details: dict[str, Any] | None = Field(None, description="Additional error details")


class MessageResponse(BaseModel):
    """Standard success message response"""

    message: str = Field(..., description="Success message")
    data: dict[str, Any] | None = Field(None, description="Additional response data")
