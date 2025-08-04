"""Memory model for Mory Server
SQLAlchemy model compatible with existing CLI data structure
"""

import json
from datetime import datetime
from uuid import uuid4

from sqlalchemy import DateTime, Index, LargeBinary, String, Text
from sqlalchemy.orm import Mapped, mapped_column, validates

from ..core.database import Base


class Memory(Base):
    """Simplified AI-driven memory model (Issue #112)"""

    __tablename__ = "memories"

    # 🎯 User input (single field)
    id: Mapped[str] = mapped_column(
        String, primary_key=True, default=lambda: f"mem_{uuid4().hex[:8]}"
    )
    value: Mapped[str] = mapped_column(Text)  # Only user input required

    # 🤖 AI-generated fields (all automatic)
    summary: Mapped[str | None] = mapped_column(Text)  # AI-generated summary
    tags: Mapped[str] = mapped_column(Text, default="[]")  # AI-generated comprehensive tags

    # ⏰ System timestamps
    created_at: Mapped[datetime] = mapped_column(DateTime, default=datetime.utcnow)
    updated_at: Mapped[datetime] = mapped_column(
        DateTime, default=datetime.utcnow, onupdate=datetime.utcnow
    )
    ai_processed_at: Mapped[datetime | None] = mapped_column(DateTime)  # AI processing completion

    # 🔍 Search optimization (single embedding from summary)
    embedding: Mapped[bytes | None] = mapped_column(LargeBinary)  # Summary-based vector
    embedding_hash: Mapped[str | None] = mapped_column(String, index=True)
    embedding_model: Mapped[str | None] = mapped_column(String)  # Model used for embedding

    # Simplified indexes
    __table_args__ = (
        Index("idx_updated_at", "updated_at"),
        Index("idx_ai_processed", "ai_processed_at"),
        Index("idx_tags_search", "tags"),
        Index("idx_embedding_hash", "embedding_hash"),
    )

    @validates("tags")
    def validate_tags(self, key, value):
        """Ensure tags is always valid JSON"""
        if isinstance(value, list):
            return json.dumps(value)
        elif isinstance(value, str):
            try:
                # Validate it's valid JSON
                json.loads(value)
                return value
            except json.JSONDecodeError:
                return "[]"
        return "[]"

    @property
    def tags_list(self) -> list[str]:
        """Get tags as Python list"""
        try:
            return json.loads(self.tags) if self.tags else []
        except json.JSONDecodeError:
            return []

    @tags_list.setter
    def tags_list(self, value: list[str]):
        """Set tags from Python list"""
        self.tags = json.dumps(value)

    @property
    def has_embedding(self) -> bool:
        """Check if memory has semantic embedding"""
        return self.embedding is not None and len(self.embedding) > 0

    @property
    def is_ai_processed(self) -> bool:
        """Check if AI processing is complete"""
        return self.ai_processed_at is not None

    @property
    def processing_status(self) -> str:
        """Get processing status"""
        if not self.is_ai_processed:
            return "pending"
        elif self.summary and self.tags_list and self.has_embedding:
            return "complete"
        else:
            return "partial"

    def to_dict(self) -> dict:
        """Convert to dictionary for API responses"""
        return {
            "id": self.id,
            "value": self.value,
            "tags": self.tags_list,  # AI-generated comprehensive tags
            "created_at": self.created_at.isoformat() if self.created_at else None,
            "updated_at": self.updated_at.isoformat() if self.updated_at else None,
            "has_embedding": self.has_embedding,
            "summary": self.summary,
            "ai_processed_at": self.ai_processed_at.isoformat() if self.ai_processed_at else None,
            "processing_status": self.processing_status,
        }

    def __repr__(self):
        tags_preview = self.tags_list[:2] if self.tags_list else []
        return f"<Memory(id='{self.id}', tags={tags_preview}, status='{self.processing_status}')>"
