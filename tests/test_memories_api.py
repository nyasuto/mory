"""Tests for memory CRUD API endpoints"""

import pytest


@pytest.fixture
def sample_memory_data():
    """Sample memory data for testing - simplified AI-driven schema (Issue #112)"""
    return {
        "value": "This is a test memory",
    }


class TestCreateMemory:
    """Tests for POST /api/memories"""

    def test_create_memory_success(self, client, db_session, sample_memory_data):
        """Test successful memory creation - simplified AI-driven schema (Issue #112)"""
        response = client.post("/api/memories", json=sample_memory_data)

        assert response.status_code == 201
        data = response.json()

        assert data["value"] == sample_memory_data["value"]
        assert "id" in data
        assert "created_at" in data
        assert "updated_at" in data
        assert data["has_embedding"] is False
        # AI-generated fields should be present but may be None initially
        assert "tags" in data  # AI-generated comprehensive tags
        assert "summary" in data  # AI-generated summary
        assert "processing_status" in data  # AI processing status

    def test_create_memory_minimal_input(self, client, db_session):
        """Test creating memory with minimal input - simplified AI-driven schema (Issue #112)"""
        memory_data = {"value": "Memory with minimal input"}

        response = client.post("/api/memories", json=memory_data)

        assert response.status_code == 201
        data = response.json()
        assert data["value"] == memory_data["value"]
        assert "id" in data
        assert "tags" in data  # AI will generate tags
        assert "summary" in data  # AI will generate summary

    def test_create_memory_creates_new_each_time(self, client, db_session, sample_memory_data):
        """Test that each memory creation creates a new memory - simplified AI-driven schema (Issue #112)"""
        # Create first memory
        response1 = client.post("/api/memories", json=sample_memory_data)
        assert response1.status_code == 201
        first_id = response1.json()["id"]

        # Create second memory with different content
        updated_data = {"value": "Updated memory value"}

        response2 = client.post("/api/memories", json=updated_data)
        assert response2.status_code == 201

        # Should be different ID (new memory)
        data = response2.json()
        assert data["id"] != first_id
        assert data["value"] == "Updated memory value"

    def test_create_memory_validation_errors(self, client, db_session):
        """Test validation errors - simplified AI-driven schema (Issue #112)"""
        # Empty value
        response = client.post("/api/memories", json={"value": ""})
        assert response.status_code == 422

        # Missing required fields
        response = client.post("/api/memories", json={})
        assert response.status_code == 422

        # Only whitespace value
        response = client.post("/api/memories", json={"value": "   "})
        assert response.status_code == 422


class TestGetMemory:
    """Tests for GET /api/memories/{id} - simplified AI-driven schema (Issue #112)"""

    def test_get_memory_success(self, client, db_session, sample_memory_data):
        """Test successful memory retrieval - simplified AI-driven schema (Issue #112)"""
        # Create memory first
        create_response = client.post("/api/memories", json=sample_memory_data)
        assert create_response.status_code == 201
        memory_id = create_response.json()["id"]

        # Get memory by ID
        response = client.get(f"/api/memories/{memory_id}")

        assert response.status_code == 200
        data = response.json()
        assert data["id"] == memory_id
        assert data["value"] == sample_memory_data["value"]

    def test_get_memory_shows_ai_processing_status(self, client, db_session, sample_memory_data):
        """Test getting memory shows AI processing status - simplified AI-driven schema (Issue #112)"""
        # Create memory
        create_response = client.post("/api/memories", json=sample_memory_data)
        memory_id = create_response.json()["id"]

        # Get memory
        response = client.get(f"/api/memories/{memory_id}")
        assert response.status_code == 200

        data = response.json()
        assert "processing_status" in data
        assert data["processing_status"] in ["pending", "partial", "complete"]
        assert "tags" in data  # AI-generated tags
        assert "summary" in data  # AI-generated summary

    def test_get_memory_not_found(self, client, db_session):
        """Test getting non-existent memory - simplified AI-driven schema (Issue #112)"""
        response = client.get("/api/memories/nonexistent_id")
        assert response.status_code == 404
        assert "not found" in response.json()["detail"]


class TestListMemories:
    """Tests for GET /api/memories"""

    def test_list_memories_empty(self, client, db_session):
        """Test listing when no memories exist - simplified AI-driven schema (Issue #112)"""
        response = client.get("/api/memories")

        assert response.status_code == 200
        data = response.json()
        assert data["memories"] == []
        assert data["total"] == 0

    def test_list_memories_with_data(self, client, db_session):
        """Test listing with multiple memories - simplified AI-driven schema (Issue #112)"""
        # Create multiple memories
        for i in range(3):
            memory_data = {"value": f"Memory {i}"}
            client.post("/api/memories", json=memory_data)

        response = client.get("/api/memories")

        assert response.status_code == 200
        data = response.json()
        assert len(data["memories"]) == 3
        assert data["total"] == 3

        # Check AI-driven fields are present (optimized response)
        for memory in data["memories"]:
            assert "id" in memory
            assert "value" not in memory  # Optimized response doesn't include full value
            assert "tags" in memory  # AI-generated
            assert "summary" in memory  # AI-generated summary instead of full value
            assert "processing_status" in memory

    def test_list_memories_shows_ai_processing(self, client, db_session):
        """Test listing shows AI processing status - simplified AI-driven schema (Issue #112)"""
        # Create memories with different content
        memory_contents = ["Work related memory", "Personal thoughts", "Project notes"]
        for content in memory_contents:
            memory_data = {"value": content}
            client.post("/api/memories", json=memory_data)

        response = client.get("/api/memories")

        assert response.status_code == 200
        data = response.json()
        assert len(data["memories"]) == 3
        assert data["total"] == 3

        for memory in data["memories"]:
            assert "processing_status" in memory
            assert memory["processing_status"] in ["pending", "partial", "complete"]
            assert "tags" in memory  # AI-generated tags
            assert "summary" in memory  # AI-generated summary

    def test_list_memories_pagination(self, client, db_session):
        """Test pagination parameters - simplified AI-driven schema (Issue #112)"""
        # Create 5 memories
        for i in range(5):
            memory_data = {"value": f"Memory {i}"}
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
    """Tests for PUT /api/memories/{id} - simplified AI-driven schema (Issue #112)"""

    def test_update_memory_success(self, client, db_session, sample_memory_data):
        """Test successful memory update - simplified AI-driven schema (Issue #112)"""
        # Create memory
        create_response = client.post("/api/memories", json=sample_memory_data)
        memory_id = create_response.json()["id"]

        # Update memory - AI will re-process tags and summary when value changes
        update_data = {"value": "Updated memory value"}

        response = client.put(f"/api/memories/{memory_id}", json=update_data)

        assert response.status_code == 200
        data = response.json()
        assert data["value"] == "Updated memory value"
        assert data["id"] == memory_id
        # AI will re-process the content, so processing_status may change
        assert "processing_status" in data
        assert "tags" in data  # AI will regenerate tags
        assert "summary" in data  # AI will regenerate summary

    def test_update_memory_not_found(self, client, db_session):
        """Test updating non-existent memory - simplified AI-driven schema (Issue #112)"""
        update_data = {"value": "Updated value"}

        response = client.put("/api/memories/nonexistent_id", json=update_data)
        assert response.status_code == 404


class TestDeleteMemory:
    """Tests for DELETE /api/memories/{id} - simplified AI-driven schema (Issue #112)"""

    def test_delete_memory_success(self, client, db_session, sample_memory_data):
        """Test successful memory deletion - simplified AI-driven schema (Issue #112)"""
        # Create memory
        create_response = client.post("/api/memories", json=sample_memory_data)
        memory_id = create_response.json()["id"]

        # Delete memory by ID
        response = client.delete(f"/api/memories/{memory_id}")

        assert response.status_code == 200
        data = response.json()
        assert "deleted successfully" in data["message"]
        assert data["data"]["deleted_id"] == memory_id

        # Verify memory is gone
        get_response = client.get(f"/api/memories/{memory_id}")
        assert get_response.status_code == 404

    def test_delete_memory_not_found(self, client, db_session):
        """Test deleting non-existent memory - simplified AI-driven schema (Issue #112)"""
        response = client.delete("/api/memories/nonexistent_id")
        assert response.status_code == 404


class TestMemoryStats:
    """Tests for GET /api/memories/stats"""

    def test_stats_empty_database(self, client, db_session):
        """Test stats with empty database - simplified AI-driven schema (Issue #112)"""
        response = client.get("/api/memories/stats")

        assert response.status_code == 200
        data = response.json()
        assert data["total_memories"] == 0
        assert data["total_tags"] == 0
        assert data["recent_memories"] == 0
        assert "storage_info" in data

    def test_stats_with_data(self, client, db_session):
        """Test stats with sample data - simplified AI-driven schema (Issue #112)"""
        # Create memories with different content (AI will generate tags)
        memories_data = [
            {"value": "Work memory about important project"},
            {"value": "Work memory about meeting notes"},
            {"value": "Personal memory about family"},
        ]

        for memory_data in memories_data:
            client.post("/api/memories", json=memory_data)

        response = client.get("/api/memories/stats")

        assert response.status_code == 200
        data = response.json()
        assert data["total_memories"] == 3
        assert data["recent_memories"] == 3  # All recent
        # Note: total_tags will depend on AI-generated tags
        assert "total_tags" in data
        assert "storage_info" in data


class TestAPIPerformance:
    """Performance tests for API endpoints"""

    # Performance test removed - focusing on basic functionality only
    pass
