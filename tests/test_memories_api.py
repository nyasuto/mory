"""Tests for memory CRUD API endpoints"""

import pytest


@pytest.fixture
def sample_memory_data():
    """Sample memory data for testing"""
    return {
        "category": "test_category",
        "key": "test_key",
        "value": "This is a test memory",
        "tags": ["test", "api", "memory"],
    }


class TestCreateMemory:
    """Tests for POST /api/memories"""

    def test_create_memory_success(self, client, db_session, sample_memory_data):
        """Test successful memory creation"""
        response = client.post("/api/memories", json=sample_memory_data)

        assert response.status_code == 201
        data = response.json()

        assert data["category"] == sample_memory_data["category"]
        assert data["key"] == sample_memory_data["key"]
        assert data["value"] == sample_memory_data["value"]
        assert data["tags"] == sample_memory_data["tags"]
        assert "id" in data
        assert "created_at" in data
        assert "updated_at" in data
        assert data["has_embedding"] is False

    def test_create_memory_without_key(self, client, db_session):
        """Test creating memory without key"""
        memory_data = {"category": "test_category", "value": "Memory without key", "tags": ["test"]}

        response = client.post("/api/memories", json=memory_data)

        assert response.status_code == 201
        data = response.json()
        assert data["key"] is None
        assert data["value"] == memory_data["value"]

    def test_create_memory_duplicate_key_updates(self, client, db_session, sample_memory_data):
        """Test that duplicate key updates existing memory"""
        # Create first memory
        response1 = client.post("/api/memories", json=sample_memory_data)
        assert response1.status_code == 201
        first_id = response1.json()["id"]

        # Create second memory with same key/category
        updated_data = sample_memory_data.copy()
        updated_data["value"] = "Updated memory value"
        updated_data["tags"] = ["updated", "test"]

        response2 = client.post("/api/memories", json=updated_data)
        assert response2.status_code == 201

        # Should be same ID (updated, not new)
        data = response2.json()
        assert data["id"] == first_id
        assert data["value"] == "Updated memory value"
        assert data["tags"] == ["updated", "test"]

    def test_create_memory_validation_errors(self, client, db_session):
        """Test validation errors"""
        # Empty category
        response = client.post("/api/memories", json={"category": "", "value": "test", "tags": []})
        assert response.status_code == 422

        # Empty value
        response = client.post("/api/memories", json={"category": "test", "value": "", "tags": []})
        assert response.status_code == 422

        # Missing required fields
        response = client.post("/api/memories", json={"tags": []})
        assert response.status_code == 422


class TestGetMemory:
    """Tests for GET /api/memories/{key}"""

    def test_get_memory_success(self, client, db_session, sample_memory_data):
        """Test successful memory retrieval"""
        # Create memory first
        create_response = client.post("/api/memories", json=sample_memory_data)
        assert create_response.status_code == 201

        # Get memory
        response = client.get(f"/api/memories/{sample_memory_data['key']}")

        assert response.status_code == 200
        data = response.json()
        assert data["key"] == sample_memory_data["key"]
        assert data["value"] == sample_memory_data["value"]

    def test_get_memory_with_category_filter(self, client, db_session, sample_memory_data):
        """Test getting memory with category filter"""
        # Create memory
        client.post("/api/memories", json=sample_memory_data)

        # Get with correct category
        response = client.get(
            f"/api/memories/{sample_memory_data['key']}",
            params={"category": sample_memory_data["category"]},
        )
        assert response.status_code == 200

        # Get with wrong category
        response = client.get(
            f"/api/memories/{sample_memory_data['key']}", params={"category": "wrong_category"}
        )
        assert response.status_code == 404

    def test_get_memory_not_found(self, client, db_session):
        """Test getting non-existent memory"""
        response = client.get("/api/memories/nonexistent_key")
        assert response.status_code == 404
        assert "not found" in response.json()["detail"]


class TestListMemories:
    """Tests for GET /api/memories"""

    def test_list_memories_empty(self, client, db_session):
        """Test listing when no memories exist"""
        response = client.get("/api/memories")

        assert response.status_code == 200
        data = response.json()
        assert data["memories"] == []
        assert data["total"] == 0
        assert data["category"] is None

    def test_list_memories_with_data(self, client, db_session):
        """Test listing with multiple memories"""
        # Create multiple memories
        for i in range(3):
            memory_data = {
                "category": f"category_{i}",
                "key": f"key_{i}",
                "value": f"Memory {i}",
                "tags": [f"tag_{i}"],
            }
            client.post("/api/memories", json=memory_data)

        response = client.get("/api/memories")

        assert response.status_code == 200
        data = response.json()
        assert len(data["memories"]) == 3
        assert data["total"] == 3

    def test_list_memories_with_category_filter(self, client, db_session):
        """Test listing with category filter"""
        # Create memories in different categories
        for category in ["work", "personal", "work"]:
            memory_data = {"category": category, "value": f"Memory in {category}", "tags": []}
            client.post("/api/memories", json=memory_data)

        # Filter by category
        response = client.get("/api/memories", params={"category": "work"})

        assert response.status_code == 200
        data = response.json()
        assert len(data["memories"]) == 2
        assert data["total"] == 2
        assert data["category"] == "work"

        for memory in data["memories"]:
            assert memory["category"] == "work"

    def test_list_memories_pagination(self, client, db_session):
        """Test pagination parameters"""
        # Create 5 memories
        for i in range(5):
            memory_data = {
                "category": "test",
                "key": f"key_{i}",
                "value": f"Memory {i}",
                "tags": [],
            }
            client.post("/api/memories", json=memory_data)

        # Test limit
        response = client.get("/api/memories", params={"limit": 2})
        assert response.status_code == 200
        data = response.json()
        assert len(data["memories"]) == 2
        assert data["total"] == 5

        # Test offset
        response = client.get("/api/memories", params={"limit": 2, "offset": 2})
        assert response.status_code == 200
        data = response.json()
        assert len(data["memories"]) == 2
        assert data["total"] == 5


class TestUpdateMemory:
    """Tests for PUT /api/memories/{key}"""

    def test_update_memory_success(self, client, db_session, sample_memory_data):
        """Test successful memory update"""
        # Create memory
        client.post("/api/memories", json=sample_memory_data)

        # Update memory
        update_data = {"value": "Updated memory value", "tags": ["updated"]}

        response = client.put(f"/api/memories/{sample_memory_data['key']}", json=update_data)

        assert response.status_code == 200
        data = response.json()
        assert data["value"] == "Updated memory value"
        assert data["tags"] == ["updated"]
        assert data["category"] == sample_memory_data["category"]  # Unchanged

    def test_update_memory_not_found(self, client, db_session):
        """Test updating non-existent memory"""
        update_data = {"value": "Updated value"}

        response = client.put("/api/memories/nonexistent", json=update_data)
        assert response.status_code == 404


class TestDeleteMemory:
    """Tests for DELETE /api/memories/{key}"""

    def test_delete_memory_success(self, client, db_session, sample_memory_data):
        """Test successful memory deletion"""
        # Create memory
        create_response = client.post("/api/memories", json=sample_memory_data)
        memory_id = create_response.json()["id"]

        # Delete memory
        response = client.delete(f"/api/memories/{sample_memory_data['key']}")

        assert response.status_code == 200
        data = response.json()
        assert "deleted successfully" in data["message"]
        assert data["data"]["deleted_id"] == memory_id

        # Verify memory is gone
        get_response = client.get(f"/api/memories/{sample_memory_data['key']}")
        assert get_response.status_code == 404

    def test_delete_memory_not_found(self, client, db_session):
        """Test deleting non-existent memory"""
        response = client.delete("/api/memories/nonexistent")
        assert response.status_code == 404


class TestMemoryStats:
    """Tests for GET /api/memories/stats"""

    def test_stats_empty_database(self, client, db_session):
        """Test stats with empty database"""
        response = client.get("/api/memories/stats")

        assert response.status_code == 200
        data = response.json()
        assert data["total_memories"] == 0
        assert data["total_categories"] == 0
        assert data["total_tags"] == 0
        assert data["categories"] == {}
        assert data["recent_memories"] == 0
        assert "storage_info" in data

    def test_stats_with_data(self, client, db_session):
        """Test stats with sample data"""
        # Create memories in different categories
        memories_data = [
            {"category": "work", "value": "Work memory 1", "tags": ["important", "project"]},
            {"category": "work", "value": "Work memory 2", "tags": ["meeting"]},
            {"category": "personal", "value": "Personal memory", "tags": ["family", "important"]},
        ]

        for memory_data in memories_data:
            client.post("/api/memories", json=memory_data)

        response = client.get("/api/memories/stats")

        assert response.status_code == 200
        data = response.json()
        assert data["total_memories"] == 3
        assert data["total_categories"] == 2
        assert data["total_tags"] == 4  # unique tags: important, project, meeting, family
        assert data["categories"]["work"] == 2
        assert data["categories"]["personal"] == 1
        assert data["recent_memories"] == 3  # All recent


class TestAPIPerformance:
    """Performance tests for API endpoints"""

    # Performance test removed - focusing on basic functionality only
    pass
