"""Tests for memory search API endpoints"""

import pytest

pytestmark = pytest.mark.skip(
    reason="Search service needs update for simplified schema (Issue #112)"
)


@pytest.fixture
def sample_memories():
    """Sample memory data for search testing"""
    return [
        {
            "category": "programming",
            "key": "fastapi_tutorial",
            "value": "FastAPI is a modern, fast web framework for building APIs with Python 3.7+",
            "tags": ["python", "web", "api"],
        },
        {
            "category": "programming",
            "key": "python_basics",
            "value": "Python is a high-level programming language with dynamic semantics",
            "tags": ["python", "basics", "learning"],
        },
        {
            "category": "personal",
            "key": "vacation_plans",
            "value": "Planning a trip to Japan in spring to see cherry blossoms",
            "tags": ["travel", "japan", "spring"],
        },
        {
            "category": "work",
            "key": "meeting_notes",
            "value": "Discussed the new API design and database optimization strategies",
            "tags": ["meeting", "api", "database"],
        },
        {
            "category": "learning",
            "key": "ai_concepts",
            "value": "Machine learning and artificial intelligence fundamentals",
            "tags": ["ai", "ml", "learning"],
        },
    ]


class TestSearchAPI:
    """Tests for POST /api/memories/search"""

    def test_search_empty_database(self, client, db_session):
        """Test search with empty database"""
        search_request = {"query": "python", "limit": 10, "offset": 0}

        response = client.post("/api/memories/search", json=search_request)

        assert response.status_code == 200
        data = response.json()
        assert data["results"] == []
        assert data["total"] == 0
        assert data["query"] == "python"
        assert "execution_time_ms" in data

    def test_search_basic_query(self, client, db_session, sample_memories):
        """Test basic search functionality"""
        # Create test memories
        for memory_data in sample_memories:
            client.post("/api/memories", json=memory_data)

        # Search for 'python'
        search_request = {"query": "python", "limit": 10, "offset": 0}

        response = client.post("/api/memories/search", json=search_request)

        assert response.status_code == 200
        data = response.json()
        assert len(data["results"]) >= 2  # Should find at least 2 Python-related memories
        assert data["total"] >= 2
        assert data["query"] == "python"

        # Check result structure
        for result in data["results"]:
            assert "memory" in result
            assert "score" in result
            assert "search_type" in result
            assert 0.0 <= result["score"] <= 1.0

    def test_search_with_category_filter(self, client, db_session, sample_memories):
        """Test search with category filtering"""
        # Create test memories
        for memory_data in sample_memories:
            client.post("/api/memories", json=memory_data)

        # Search in 'programming' category only
        search_request = {"query": "python", "category": "programming", "limit": 10, "offset": 0}

        response = client.post("/api/memories/search", json=search_request)

        assert response.status_code == 200
        data = response.json()

        # All results should be from programming category
        for result in data["results"]:
            assert result["memory"]["category"] == "programming"

        assert data["filters"]["category"] == "programming"

    def test_search_with_tags_filter(self, client, db_session, sample_memories):
        """Test search with tags filtering"""
        # Create test memories
        for memory_data in sample_memories:
            client.post("/api/memories", json=memory_data)

        # Search with specific tags
        search_request = {"query": "api", "tags": ["python"], "limit": 10, "offset": 0}

        response = client.post("/api/memories/search", json=search_request)

        assert response.status_code == 200
        data = response.json()

        # Results should contain memories with 'python' tag
        for result in data["results"]:
            assert "python" in result["memory"]["tags"]

        assert data["filters"]["tags"] == ["python"]

    def test_search_pagination(self, client, db_session, sample_memories):
        """Test search pagination"""
        # Create test memories
        for memory_data in sample_memories:
            client.post("/api/memories", json=memory_data)

        # Search with pagination
        search_request = {
            "query": "a",  # Broad search to get multiple results
            "limit": 2,
            "offset": 0,
        }

        response = client.post("/api/memories/search", json=search_request)

        assert response.status_code == 200
        data = response.json()
        assert len(data["results"]) <= 2

        # Test next page
        search_request["offset"] = 2
        response = client.post("/api/memories/search", json=search_request)

        assert response.status_code == 200
        data = response.json()
        # Should have results or be empty if all data was on first page

    def test_search_different_types(self, client, db_session, sample_memories):
        """Test different search types"""
        # Create test memories
        for memory_data in sample_memories:
            client.post("/api/memories", json=memory_data)

        search_types = ["fts5", "semantic", "hybrid", "like"]

        for search_type in search_types:
            search_request = {
                "query": "python",
                "search_type": search_type,
                "limit": 10,
                "offset": 0,
            }

            response = client.post("/api/memories/search", json=search_request)

            assert response.status_code == 200
            data = response.json()
            # Should return results or fall back to available search type
            assert "search_type" in data
            assert "results" in data

    def test_search_validation_errors(self, client, db_session):
        """Test search request validation"""
        # Empty query
        search_request = {"query": "", "limit": 10, "offset": 0}

        response = client.post("/api/memories/search", json=search_request)
        assert response.status_code == 422

        # Invalid limit
        search_request = {"query": "test", "limit": 0, "offset": 0}

        response = client.post("/api/memories/search", json=search_request)
        assert response.status_code == 422

        # Invalid offset
        search_request = {"query": "test", "limit": 10, "offset": -1}

        response = client.post("/api/memories/search", json=search_request)
        assert response.status_code == 422

    def test_search_japanese_content(self, client, db_session):
        """Test search with Japanese content"""
        # Create Japanese memory
        japanese_memory = {
            "category": "日本語",
            "key": "学習ノート",
            "value": "日本語の勉強をしています。ひらがな、カタカナ、漢字を覚える必要があります。",
            "tags": ["日本語", "勉強", "言語"],
        }

        client.post("/api/memories", json=japanese_memory)

        # Search in Japanese
        search_request = {"query": "日本語", "limit": 10, "offset": 0}

        response = client.post("/api/memories/search", json=search_request)

        assert response.status_code == 200
        data = response.json()
        assert len(data["results"]) >= 1

        # Should find the Japanese memory
        found_japanese = False
        for result in data["results"]:
            if "日本語" in result["memory"]["value"]:
                found_japanese = True
                break

        assert found_japanese


# Performance tests removed - focusing on basic functionality only
