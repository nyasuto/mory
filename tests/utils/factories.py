"""Test data factories for creating consistent test data"""

from datetime import datetime
from uuid import uuid4

from app.models.memory import Memory
from app.models.schemas import MemoryCreate, SearchRequest


class MemoryFactory:
    """Factory for creating Memory test data"""

    @staticmethod
    def create_memory_data(
        value: str = "Test memory content",
        **kwargs,
    ) -> dict:
        """Create memory data dict - simplified AI-driven schema (Issue #112)"""
        return {
            "value": value,
            **kwargs,
        }

    @staticmethod
    def create_memory_create(
        value: str = "Test memory content",
    ) -> MemoryCreate:
        """Create MemoryCreate instance - simplified AI-driven schema (Issue #112)"""
        return MemoryCreate(value=value)

    @staticmethod
    def create_memory_model(
        value: str = "Test memory content",
        tags: list[str] | None = None,
        summary: str | None = None,
        **kwargs,
    ) -> Memory:
        """Create Memory model instance - simplified AI-driven schema (Issue #112)"""
        if tags is None:
            tags = ["test", "factory"]

        memory = Memory(
            id=kwargs.get("id", f"mem_{uuid4().hex[:8]}"),
            value=value,
            tags_list=tags,
            created_at=kwargs.get("created_at", datetime.utcnow()),
            updated_at=kwargs.get("updated_at", datetime.utcnow()),
            summary=summary,
            ai_processed_at=kwargs.get("ai_processed_at"),
        )
        return memory

    @staticmethod
    def create_memory_with_summary(
        value: str = "This is a longer test memory content that should be summarized",
        summary: str = "Test memory summary",
        tags: list[str] | None = None,
        **kwargs,
    ) -> Memory:
        """Create Memory with summary - simplified AI-driven schema (Issue #112)"""
        if tags is None:
            tags = ["test", "summarized"]

        return MemoryFactory.create_memory_model(
            value=value,
            tags=tags,
            summary=summary,
            ai_processed_at=datetime.utcnow(),
            **kwargs,
        )

    @staticmethod
    def create_japanese_memory(
        value: str = "これは日本語のテストメモリです。長い文章で要約のテストに使用します。",
        summary: str = "日本語テストメモリ",
        tags: list[str] | None = None,
    ) -> Memory:
        """Create Japanese memory for testing - simplified AI-driven schema (Issue #112)"""
        if tags is None:
            tags = ["日本語", "テスト"]

        return MemoryFactory.create_memory_with_summary(value=value, summary=summary, tags=tags)

    @staticmethod
    def create_large_memory(size: int = 1000, **kwargs) -> Memory:
        """Create memory with large content - simplified AI-driven schema (Issue #112)"""
        large_content = " ".join(
            [f"This is sentence {i} in a large memory content." for i in range(size)]
        )
        summary = f"Large memory with {size} sentences"

        return MemoryFactory.create_memory_with_summary(
            value=large_content,
            summary=summary,
            tags=["large", "performance"],
            **kwargs,
        )


class SearchFactory:
    """Factory for creating search-related test data"""

    @staticmethod
    def create_search_request(
        query: str = "test query",
        tags: list[str] | None = None,
        limit: int = 20,
        offset: int = 0,
        search_type: str = "hybrid",
    ) -> SearchRequest:
        """Create SearchRequest instance - simplified AI-driven schema (Issue #112)"""
        return SearchRequest(
            query=query,
            tags=tags,
            limit=limit,
            offset=offset,
            search_type=search_type,
        )

    @staticmethod
    def create_japanese_search_request(query: str = "日本語検索", **kwargs) -> SearchRequest:
        """Create Japanese search request - simplified AI-driven schema (Issue #112)"""
        return SearchFactory.create_search_request(
            query=query,
            tags=["日本語"] if not kwargs.get("tags") else kwargs["tags"],
            **{k: v for k, v in kwargs.items() if k != "tags"},
        )


class TestDataSets:
    """Pre-defined test data sets"""

    @staticmethod
    def create_mixed_memories(count: int = 10) -> list[Memory]:
        """Create mixed set of memories for testing"""
        memories = []

        # Regular memories
        for i in range(count // 3):
            memories.append(
                MemoryFactory.create_memory_model(
                    value=f"Memory content {i}",
                    tags=[f"tag_{i}", "regular"],
                )
            )

        # Memories with summaries
        for i in range(count // 3):
            memories.append(
                MemoryFactory.create_memory_with_summary(
                    value=f"Longer memory content {i} that has been summarized",
                    summary=f"Content {i}",
                    tags=[f"tag_{i}", "summarized"],
                )
            )

        # Japanese memories
        remaining = count - len(memories)
        for i in range(remaining):
            memories.append(
                MemoryFactory.create_japanese_memory(
                    value=f"日本語メモリ {i} の内容です。",
                    summary=f"日本語メモリ {i}",
                    tags=["日本語", f"タグ_{i}"],
                )
            )

        return memories

    @staticmethod
    def create_performance_dataset(size: int = 100) -> list[Memory]:
        """Create large dataset for performance testing"""
        memories = []

        for i in range(size):
            content_size = 50 + (i % 10) * 20  # Vary content size
            memory = MemoryFactory.create_large_memory(size=content_size)
            memories.append(memory)

        return memories
