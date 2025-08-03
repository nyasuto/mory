"""Memory model for Mory Server
SQLAlchemy model compatible with existing CLI data structure
"""

import json
from datetime import datetime
from uuid import uuid4

from sqlalchemy import Column, DateTime, Index, LargeBinary, String, Text
from sqlalchemy.orm import validates

from ..core.database import Base


class Memory(Base):
    """Memory model with SQLite storage and FTS5 support"""

    __tablename__ = "memories"

    # Core fields
    id = Column(String, primary_key=True, default=lambda: f"mem_{uuid4().hex[:8]}")
    category = Column(String, nullable=False, index=True)
    key = Column(String, nullable=True, index=True)  # User-friendly alias
    value = Column(Text, nullable=False)
    tags = Column(Text, default="[]")  # JSON serialized list

    # Timestamps
    created_at = Column(DateTime, default=datetime.utcnow, nullable=False)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow, nullable=False)

    # Summary fields (Issue #109)
    summary = Column(Text, nullable=True)  # AI-generated summary
    summary_generated_at = Column(DateTime, nullable=True)  # Summary generation timestamp

    # Semantic search fields
    embedding = Column(LargeBinary, nullable=True)  # Vector embedding
    embedding_hash = Column(String, nullable=True, index=True)  # Content hash for embedding

    # Database indexes
    __table_args__ = (
        Index("idx_category_created", "category", "created_at"),
        Index("idx_updated_at", "updated_at"),
        Index("idx_key_category", "key", "category"),
        Index("idx_summary_generated", "summary_generated_at"),  # Issue #109
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
            return json.loads(self.tags) if self.tags else []  # type: ignore[arg-type]
        except json.JSONDecodeError:
            return []

    @tags_list.setter
    def tags_list(self, value: list[str]):
        """Set tags from Python list"""
        self.tags = json.dumps(value)  # type: ignore[assignment]

    @property
    def has_embedding(self) -> bool:
        """Check if memory has semantic embedding"""
        return self.embedding is not None and len(self.embedding) > 0

    def to_dict(self) -> dict:
        """Convert to dictionary for API responses"""
        return {
            "id": self.id,
            "category": self.category,
            "key": self.key,
            "value": self.value,
            "tags": self.tags_list,  # This already returns a Python list
            "created_at": self.created_at.isoformat() if self.created_at else None,
            "updated_at": self.updated_at.isoformat() if self.updated_at else None,
            "has_embedding": self.has_embedding,
            "summary": self.summary,  # Issue #109
            "summary_generated_at": self.summary_generated_at.isoformat()
            if self.summary_generated_at
            else None,  # Issue #109
        }

    def __repr__(self):
        return f"<Memory(id='{self.id}', category='{self.category}', key='{self.key}')>"
