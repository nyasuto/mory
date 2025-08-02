"""JSON-based storage implementation for memories."""

from __future__ import annotations

import json
from pathlib import Path

# No additional typing imports needed for Python 3.11+
import aiofiles

from .memory import (
    Memory,
    MemoryNotFoundError,
    OperationLog,
    SearchQuery,
    SearchResult,
)


class JSONMemoryStore:
    """JSON-based implementation of MemoryStore."""

    def __init__(self, data_dir: str | None = None) -> None:
        """Initialize the JSON memory store.

        Args:
            data_dir: Directory to store data files
        """
        self.data_dir = Path(data_dir or "data")
        self.memories_file = self.data_dir / "memories.json"
        self.operations_file = self.data_dir / "operations.json"
        self._memories: dict[str, Memory] = {}
        self._operations: list[OperationLog] = []
        self._ensure_data_dir()

    def _ensure_data_dir(self) -> None:
        """Ensure data directory exists."""
        self.data_dir.mkdir(exist_ok=True)

    async def load(self) -> None:
        """Load memories and operations from JSON files."""
        await self._load_memories()
        await self._load_operations()

    async def _load_memories(self) -> None:
        """Load memories from JSON file."""
        if not self.memories_file.exists():
            self._memories = {}
            return

        try:
            async with aiofiles.open(self.memories_file, encoding="utf-8") as f:
                content = await f.read()
                if content.strip():
                    data = json.loads(content)
                    self._memories = {
                        mem_id: Memory.model_validate(mem_data)
                        for mem_id, mem_data in data.items()
                    }
                else:
                    self._memories = {}
        except (json.JSONDecodeError, Exception):
            self._memories = {}

    async def _load_operations(self) -> None:
        """Load operations from JSON file."""
        if not self.operations_file.exists():
            self._operations = []
            return

        try:
            async with aiofiles.open(self.operations_file, encoding="utf-8") as f:
                content = await f.read()
                if content.strip():
                    data = json.loads(content)
                    self._operations = [
                        OperationLog.model_validate(op_data) for op_data in data
                    ]
                else:
                    self._operations = []
        except (json.JSONDecodeError, Exception):
            self._operations = []

    async def save_to_disk(self) -> None:
        """Save memories and operations to JSON files."""
        await self._save_memories()
        await self._save_operations()

    async def _save_memories(self) -> None:
        """Save memories to JSON file."""
        data = {
            mem_id: memory.model_dump() for mem_id, memory in self._memories.items()
        }

        async with aiofiles.open(self.memories_file, "w", encoding="utf-8") as f:
            await f.write(json.dumps(data, indent=2, ensure_ascii=False, default=str))

    async def _save_operations(self) -> None:
        """Save operations to JSON file."""
        data = [op.model_dump() for op in self._operations]

        async with aiofiles.open(self.operations_file, "w", encoding="utf-8") as f:
            await f.write(json.dumps(data, indent=2, ensure_ascii=False, default=str))

    async def save(self, memory: Memory) -> str:
        """Save a memory and return its ID."""
        if memory.id in self._memories:
            # Update existing memory
            memory.update_timestamp()
            old_memory = self._memories[memory.id]
            log = OperationLog(
                operation="update",
                key=memory.key or memory.id,
                before=old_memory,
                after=memory,
            )
        else:
            # Create new memory
            log = OperationLog(
                operation="save", key=memory.key or memory.id, after=memory
            )

        self._memories[memory.id] = memory
        await self.log_operation(log)
        await self._save_memories()

        return memory.id

    async def get(self, key: str) -> Memory:
        """Get a memory by key or ID."""
        # Try by ID first
        if key in self._memories:
            return self._memories[key]

        # Try by key
        for memory in self._memories.values():
            if memory.key == key:
                return memory

        raise MemoryNotFoundError(key)

    async def get_by_id(self, id: str) -> Memory:
        """Get a memory by ID."""
        if id not in self._memories:
            raise MemoryNotFoundError(id)
        return self._memories[id]

    async def list_memories(self, category: str | None = None) -> list[Memory]:
        """List memories, optionally filtered by category."""
        memories = list(self._memories.values())

        if category:
            memories = [m for m in memories if m.category == category]

        # Sort by created_at descending
        memories.sort(key=lambda m: m.created_at, reverse=True)
        return memories

    async def search(self, query: SearchQuery) -> list[SearchResult]:
        """Search memories using string-based matching."""
        if not query.query.strip():
            # Return all memories for empty query
            memories = await self.list_memories(query.category)
            return [
                SearchResult(memory=memory, score=1.0)
                for memory in memories[: query.limit]
            ]

        # Get candidate memories
        memories = await self.list_memories(query.category)
        results = []
        query_lower = query.query.lower().strip()

        for memory in memories:
            score = self._calculate_relevance_score(memory, query_lower)
            if score >= query.min_score:
                results.append(SearchResult(memory=memory, score=score))

        # Sort by score descending
        results.sort(key=lambda r: r.score, reverse=True)
        return results[: query.limit]

    def _calculate_relevance_score(self, memory: Memory, query_lower: str) -> float:
        """Calculate relevance score for a memory against a query."""
        score = 0.0

        # Text fields to search
        key_lower = memory.key.lower()
        value_lower = memory.value.lower()
        category_lower = memory.category.lower()

        # Exact matches get highest scores
        if key_lower == query_lower:
            score += 1.0
        elif query_lower in key_lower:
            score += 0.8

        if value_lower == query_lower:
            score += 0.9
        elif query_lower in value_lower:
            score += 0.6

        if category_lower == query_lower:
            score += 0.7
        elif query_lower in category_lower:
            score += 0.5

        # Tags match
        for tag in memory.tags:
            tag_lower = tag.lower()
            if tag_lower == query_lower:
                score += 0.6
            elif query_lower in tag_lower:
                score += 0.4

        # Word boundary matching
        words = query_lower.split()
        for word in words:
            if self._contains_word(key_lower, word):
                score += 0.3
            if self._contains_word(value_lower, word):
                score += 0.2

        # Normalize score to 0-1 range
        return min(score, 1.0)

    def _contains_word(self, text: str, word: str) -> bool:
        """Check if text contains a word at word boundaries."""
        words = text.split()
        return any(word in w for w in words)

    async def delete(self, key: str) -> None:
        """Delete a memory by key."""
        memory = await self.get(key)  # This will raise MemoryNotFoundError if not found
        await self.delete_by_id(memory.id)

    async def delete_by_id(self, id: str) -> None:
        """Delete a memory by ID."""
        if id not in self._memories:
            raise MemoryNotFoundError(id)

        memory = self._memories[id]
        log = OperationLog(
            operation="delete", key=memory.key or memory.id, before=memory
        )

        del self._memories[id]
        await self.log_operation(log)
        await self._save_memories()

    async def log_operation(self, log: OperationLog) -> None:
        """Log an operation."""
        self._operations.append(log)
        # Keep only last 1000 operations to prevent file growth
        if len(self._operations) > 1000:
            self._operations = self._operations[-1000:]
        await self._save_operations()
