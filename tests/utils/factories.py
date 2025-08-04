"""Test data factories for creating consistent test data"""

from datetime import datetime
from uuid import uuid4

from app.models.memory import Memory
from app.models.schemas import MemoryCreate, SearchRequest


class MemoryFactory:
    """Factory for creating Memory test data"""

    @staticmethod
    def create_memory_data(
        category: str = "test",
        key: str | None = None,
        value: str = "Test memory content",
        tags: list[str] | None = None,
        **kwargs,
    ) -> dict:
        """Create memory data dict"""
        if tags is None:
            tags = ["test", "factory"]

        return {
            "category": category,
            "key": key or f"test_key_{uuid4().hex[:8]}",
            "value": value,
            "tags": tags,
            **kwargs,
        }

    @staticmethod
    def create_memory_create(
        category: str = "test",
        key: str | None = None,
        value: str = "Test memory content",
        tags: list[str] | None = None,
    ) -> MemoryCreate:
        """Create MemoryCreate instance"""
        if tags is None:
            tags = ["test", "factory"]

        return MemoryCreate(
            category=category, key=key or f"test_key_{uuid4().hex[:8]}", value=value, tags=tags
        )

    @staticmethod
    def create_memory_model(
        category: str = "test",
        key: str | None = None,
        value: str = "Test memory content",
        tags: list[str] | None = None,
        summary: str | None = None,
        **kwargs,
    ) -> Memory:
        """Create Memory model instance"""
        if tags is None:
            tags = ["test", "factory"]

        memory = Memory(
            id=kwargs.get("id", f"mem_{uuid4().hex[:8]}"),
            category=category,
            key=key or f"test_key_{uuid4().hex[:8]}",
            value=value,
            tags_list=tags,
            created_at=kwargs.get("created_at", datetime.utcnow()),
            updated_at=kwargs.get("updated_at", datetime.utcnow()),
            summary=summary,
            summary_generated_at=kwargs.get("summary_generated_at"),
        )
        return memory

    @staticmethod
    def create_memory_with_summary(
        category: str = "test",
        key: str | None = None,
        value: str = "This is a longer test memory content that should be summarized",
        summary: str = "Test memory summary",
        tags: list[str] | None = None,
        **kwargs,
    ) -> Memory:
        """Create Memory with summary"""
        if tags is None:
            tags = ["test", "summarized"]

        return MemoryFactory.create_memory_model(
            category=category,
            key=key,
            value=value,
            tags=tags,
            summary=summary,
            summary_generated_at=datetime.utcnow(),
            **kwargs,
        )

    @staticmethod
    def create_japanese_memory(
        category: str = "日本語",
        key: str | None = None,
        value: str = "これは日本語のテストメモリです。長い文章で要約のテストに使用します。",
        summary: str = "日本語テストメモリ",
        tags: list[str] | None = None,
    ) -> Memory:
        """Create Japanese memory for testing"""
        if tags is None:
            tags = ["日本語", "テスト"]

        return MemoryFactory.create_memory_with_summary(
            category=category, key=key, value=value, summary=summary, tags=tags
        )

    @staticmethod
    def create_large_memory(
        category: str = "large", key: str | None = None, size: int = 1000, **kwargs
    ) -> Memory:
        """Create memory with large content"""
        large_content = " ".join(
            [f"This is sentence {i} in a large memory content." for i in range(size)]
        )
        summary = f"Large memory with {size} sentences"

        return MemoryFactory.create_memory_with_summary(
            category=category,
            key=key,
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
        category: str | None = None,
        tags: list[str] | None = None,
        limit: int = 20,
        offset: int = 0,
        search_type: str = "hybrid",
    ) -> SearchRequest:
        """Create SearchRequest instance"""
        return SearchRequest(
            query=query,
            category=category,
            tags=tags,
            limit=limit,
            offset=offset,
            search_type=search_type,
        )

    @staticmethod
    def create_japanese_search_request(
        query: str = "日本語検索", category: str | None = None, **kwargs
    ) -> SearchRequest:
        """Create Japanese search request"""
        return SearchFactory.create_search_request(
            query=query,
            category=category,
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
                    category=f"category_{i % 3}",
                    value=f"Memory content {i}",
                    tags=[f"tag_{i}", "regular"],
                )
            )

        # Memories with summaries
        for i in range(count // 3):
            memories.append(
                MemoryFactory.create_memory_with_summary(
                    category=f"category_{i % 3}",
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
                    category="日本語",
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
            memory = MemoryFactory.create_large_memory(
                category=f"perf_cat_{i % 5}", key=f"perf_key_{i}", size=content_size
            )
            memories.append(memory)

        return memories
