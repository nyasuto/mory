"""Test compatibility between Go and Python implementations."""

import json
import tempfile
from datetime import datetime
from pathlib import Path

import pytest

from mory.memory import Memory
from mory.storage import JSONMemoryStore


class TestGoCompatibility:
    """Test compatibility with Go implementation data format."""

    @pytest.fixture
    def go_memory_data(self):
        """Sample memory data in Go format."""
        return {
            "memory_1643723456000000": {
                "id": "memory_1643723456000000",
                "category": "work",
                "key": "project-update",
                "value": "Updated the API documentation",
                "tags": ["documentation", "api"],
                "created_at": "2025-02-01T10:30:56Z",
                "updated_at": "2025-02-01T10:30:56Z",
            },
            "memory_1643723556000000": {
                "id": "memory_1643723556000000",
                "category": "learning",
                "key": "golang-study",
                "value": "Learned about goroutines and channels",
                "tags": ["golang", "concurrency"],
                "created_at": "2025-02-01T10:32:36Z",
                "updated_at": "2025-02-01T10:32:36Z",
            },
        }

    @pytest.fixture
    async def temp_store(self):
        """Create a temporary store for testing."""
        with tempfile.TemporaryDirectory() as temp_dir:
            store = JSONMemoryStore(temp_dir)
            await store.load()
            yield store

    async def test_load_go_format_data(self, go_memory_data, temp_store):
        """Test loading data in Go format."""
        # Write Go format data to temp file
        memories_file = Path(temp_store.data_dir) / "memories.json"

        with open(memories_file, "w") as f:
            json.dump(go_memory_data, f, indent=2)

        # Load the data
        await temp_store.load()

        # Verify data is loaded correctly
        memories = await temp_store.list_memories()
        assert len(memories) == 2

        # Check first memory
        work_memory = await temp_store.get("project-update")
        assert work_memory.category == "work"
        assert work_memory.key == "project-update"
        assert work_memory.value == "Updated the API documentation"
        assert work_memory.tags == ["documentation", "api"]

        # Check second memory
        learning_memory = await temp_store.get("golang-study")
        assert learning_memory.category == "learning"
        assert learning_memory.key == "golang-study"
        assert learning_memory.value == "Learned about goroutines and channels"
        assert learning_memory.tags == ["golang", "concurrency"]

    async def test_save_compatible_format(self, temp_store):
        """Test saving data in Go-compatible format."""
        # Create a memory
        memory = Memory(
            category="test",
            key="test-key",
            value="Test value",
            tags=["test", "compatibility"],
        )

        # Save the memory
        memory_id = await temp_store.save(memory)

        # Read the raw JSON file
        memories_file = Path(temp_store.data_dir) / "memories.json"
        with open(memories_file) as f:
            data = json.load(f)

        # Verify the format is compatible with Go
        assert memory_id in data
        saved_data = data[memory_id]

        # Check required fields exist and have correct types
        assert isinstance(saved_data["id"], str)
        assert isinstance(saved_data["category"], str)
        assert isinstance(saved_data["key"], str)
        assert isinstance(saved_data["value"], str)
        assert isinstance(saved_data["tags"], list)
        assert isinstance(saved_data["created_at"], str)
        assert isinstance(saved_data["updated_at"], str)

        # Check field values
        assert saved_data["category"] == "test"
        assert saved_data["key"] == "test-key"
        assert saved_data["value"] == "Test value"
        assert saved_data["tags"] == ["test", "compatibility"]

    async def test_id_format_compatibility(self, temp_store):
        """Test that ID format is compatible with Go implementation."""
        memory = Memory(category="test", key="id-test", value="Testing ID format")

        memory_id = await temp_store.save(memory)

        # Check ID format (should start with "memory_" followed by timestamp)
        assert memory_id.startswith("memory_")

        # Extract timestamp part
        timestamp_str = memory_id[7:]  # Remove "memory_" prefix
        timestamp = int(timestamp_str)

        # Should be a reasonable timestamp (after 2020 and before 2030)
        assert timestamp > 1577836800000000  # 2020-01-01 in microseconds
        assert timestamp < 1893456000000000  # 2030-01-01 in microseconds

    async def test_datetime_serialization(self, temp_store):
        """Test that datetime serialization is compatible with Go."""
        memory = Memory(
            category="datetime-test",
            key="dt-test",
            value="Testing datetime serialization",
        )

        await temp_store.save(memory)

        # Read the raw JSON file
        memories_file = Path(temp_store.data_dir) / "memories.json"
        with open(memories_file) as f:
            data = json.load(f)

        # Get the saved memory data
        saved_data = list(data.values())[0]

        # Check datetime format (should be ISO 8601)
        created_at = saved_data["created_at"]
        updated_at = saved_data["updated_at"]

        # Should be able to parse as ISO datetime
        datetime.fromisoformat(created_at.replace("Z", "+00:00"))
        datetime.fromisoformat(updated_at.replace("Z", "+00:00"))

        # Should end with 'Z' for UTC timezone
        assert created_at.endswith("Z") or "+" in created_at
        assert updated_at.endswith("Z") or "+" in updated_at

    async def test_search_compatibility(self, go_memory_data, temp_store):
        """Test that search works with Go format data."""
        # Write Go format data
        memories_file = Path(temp_store.data_dir) / "memories.json"
        with open(memories_file, "w") as f:
            json.dump(go_memory_data, f, indent=2)

        # Load the data
        await temp_store.load()

        # Test search
        from mory.memory import SearchQuery

        # Search for "golang"
        query = SearchQuery(query="golang")
        results = await temp_store.search(query)

        # Should find the golang memory (expects score > 0 for relevant results)
        golang_results = [r for r in results if r.score > 0]
        assert len(golang_results) >= 1
        assert golang_results[0].memory.key == "golang-study"
        assert golang_results[0].score > 0

        # Search for "api"
        query = SearchQuery(query="api")
        results = await temp_store.search(query)

        # Should find the API memory (expects score > 0 for relevant results)
        api_results = [r for r in results if r.score > 0]
        assert len(api_results) >= 1
        assert api_results[0].memory.key == "project-update"
        assert api_results[0].score > 0

    async def test_empty_key_handling(self, temp_store):
        """Test handling of memories with empty keys (Go compatibility)."""
        memory = Memory(
            category="no-key-test",
            key="",  # Empty key
            value="Memory without a key",
        )

        memory_id = await temp_store.save(memory)

        # Should be able to retrieve by ID
        retrieved = await temp_store.get_by_id(memory_id)
        assert retrieved.category == "no-key-test"
        assert retrieved.key == ""
        assert retrieved.value == "Memory without a key"

        # Should also be retrievable by ID using the get method
        retrieved2 = await temp_store.get(memory_id)
        assert retrieved2.id == memory_id
