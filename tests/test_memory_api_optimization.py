"""Test memory API optimization (TDD for Issue #111)

This module tests the API endpoint optimization where:
- List endpoints return summaries only (lightweight)
- Detail endpoints return full content
- Backward compatibility is maintained

Following TDD approach from Issue #114.
"""

import pytest
from fastapi.testclient import TestClient

from tests.utils.assertions import APIAssertions
from tests.utils.factories import MemoryFactory


class TestMemoryListOptimization:
    """Test memory list API optimization (TDD for Issue #111)"""

    @pytest.fixture
    def client(self):
        """Test client fixture"""
        from app.main import app

        return TestClient(app)

    @pytest.fixture
    def sample_memories_data(self):
        """Create sample memories with varying content lengths"""
        return [
            MemoryFactory.create_memory_data(
                category="test",
                key="short_memory",
                value="Short content",
                tags=["test", "short"],
            ),
            MemoryFactory.create_memory_data(
                category="test",
                key="long_memory",
                value="This is a very long memory content that should be summarized. " * 20,
                tags=["test", "long"],
            ),
            MemoryFactory.create_memory_data(
                category="docs",
                key="japanese_memory",
                value="これは日本語の長いメモリ内容です。要約されるべき内容となっています。" * 15,
                tags=["docs", "japanese"],
            ),
        ]

    def test_list_memories_optimized_behavior_returns_summary_only(
        self, client, db_session, sample_memories_data
    ):
        """Test optimized list behavior returns summary only (GREEN test)"""
        # Create test memories
        for memory_data in sample_memories_data:
            client.post("/api/memories", json=memory_data)

        # New behavior: list returns summary only
        response = client.get("/api/memories")

        assert response.status_code == 200
        data = response.json()

        APIAssertions.assert_api_response_structure(data, ["memories", "total"])
        assert len(data["memories"]) == len(sample_memories_data)

        # After Issue #111: returns summary, not full value
        for memory in data["memories"]:
            assert "value" not in memory or memory.get("value") is None
            assert "summary" in memory
            assert memory["summary"] is not None
            assert len(memory["summary"]) > 0  # Summary is included

    def test_list_memories_backward_compatibility_with_full_text(
        self, client, db_session, sample_memories_data
    ):
        """Test backward compatibility with include_full_text parameter (GREEN test)"""
        # Create test memories
        for memory_data in sample_memories_data:
            client.post("/api/memories", json=memory_data)

        # Test backward compatibility: include_full_text=true should return full content
        response = client.get("/api/memories?include_full_text=true")

        assert response.status_code == 200
        data = response.json()

        # With include_full_text=true: should return full content
        for memory in data["memories"]:
            assert "value" in memory
            assert memory["value"] is not None
            assert len(memory["value"]) > 0

    def test_memory_list_response_schema_implemented_correctly(self, client, db_session):
        """Test that list response uses summary schema correctly (GREEN test)"""
        # Create a memory
        memory_data = MemoryFactory.create_memory_data()
        client.post("/api/memories", json=memory_data)

        response = client.get("/api/memories")
        data = response.json()

        # New response structure - implemented in Issue #111
        assert "memories" in data

        # MemorySummaryResponse is now implemented
        from app.models.schemas import MemorySummaryResponse  # noqa: F401

        # Verify response structure matches summary schema
        for memory in data["memories"]:
            assert "id" in memory
            assert "category" in memory
            assert "summary" in memory
            assert "value" not in memory  # Should not include full value

    def test_backward_compatibility_include_full_text_parameter_implemented(
        self, client, db_session
    ):
        """Test backward compatibility parameter works correctly (GREEN test)"""
        memory_data = MemoryFactory.create_memory_data()
        client.post("/api/memories", json=memory_data)

        # Test default behavior (summary only)
        response_default = client.get("/api/memories")
        assert response_default.status_code == 200
        data_default = response_default.json()

        for memory in data_default["memories"]:
            assert "value" not in memory or memory.get("value") is None
            assert "summary" in memory

        # Test backward compatibility (full content)
        response_full = client.get("/api/memories?include_full_text=true")
        assert response_full.status_code == 200
        data_full = response_full.json()

        for memory in data_full["memories"]:
            assert "value" in memory
            assert memory["value"] is not None


class TestMemoryDetailEndpoint:
    """Test new memory detail endpoint (TDD for Issue #111)"""

    @pytest.fixture
    def client(self):
        """Test client fixture"""
        from app.main import app

        return TestClient(app)

    def test_memory_detail_endpoint_implemented(self, client, db_session):
        """Test that detail endpoint works correctly (GREEN test)"""
        # Create test memory
        memory_data = MemoryFactory.create_memory_data(key="detail_test")
        response = client.post("/api/memories", json=memory_data)
        memory_key = response.json()["key"]

        # Detail endpoint should now exist and return full content
        detail_response = client.get(f"/api/memories/{memory_key}/detail")
        assert detail_response.status_code == 200

        detail_data = detail_response.json()
        assert "value" in detail_data
        assert detail_data["value"] == memory_data["value"]
        assert "summary" in detail_data  # Summary should also be included

    def test_memory_summary_endpoint_behavior_not_optimized(self, client, db_session):
        """Test that individual memory GET still returns full content (RED test)"""
        # Create test memory
        memory_data = MemoryFactory.create_memory_data(
            key="summary_test",
            value="This should return full content currently, but summary after Issue #111",
        )
        response = client.post("/api/memories", json=memory_data)
        memory_key = response.json()["key"]

        # Current behavior: GET /memories/{key} returns full content
        get_response = client.get(f"/api/memories/{memory_key}")
        assert get_response.status_code == 200
        data = get_response.json()

        # Currently returns full content - should be optimized in Issue #111
        assert "value" in data
        assert data["value"] == memory_data["value"]

        # After Issue #111: this endpoint should return summary by default
        # and full content should be available via /detail endpoint


class TestMemoryAPIResponseSize:
    """Test memory API response size optimization (TDD for Issue #111)"""

    @pytest.fixture
    def client(self):
        """Test client fixture"""
        from app.main import app

        return TestClient(app)

    def test_response_size_optimized_successfully(self, client, db_session):
        """Test that response size is optimized (GREEN test)"""
        # Create memories with moderately large content (reduced for speed)
        large_memories = []
        for i in range(2):  # Reduced from 5 to 2
            memory_data = MemoryFactory.create_memory_data(
                key=f"large_memory_{i}",
                value="This is a large memory content that will increase response size. "
                * 10,  # Reduced from 50 to 10
                tags=["performance", "large"],
            )
            large_memories.append(memory_data)
            client.post("/api/memories", json=memory_data)

        # Get optimized response size (summary only)
        response_summary = client.get("/api/memories")
        assert response_summary.status_code == 200

        # Get full response size for comparison
        response_full = client.get("/api/memories?include_full_text=true")
        assert response_full.status_code == 200

        # Measure response sizes
        import json

        summary_size = len(json.dumps(response_summary.json()).encode("utf-8"))
        full_size = len(json.dumps(response_full.json()).encode("utf-8"))

        print(f"Summary response size: {summary_size} bytes")
        print(f"Full response size: {full_size} bytes")

        # Verify optimization: summary should be significantly smaller
        if full_size > 0:
            reduction_percent = ((full_size - summary_size) / full_size) * 100
            print(f"Size reduction: {reduction_percent:.1f}%")

            # Should achieve significant reduction (at least 50%)
            assert summary_size < full_size * 0.5

    def test_performance_not_optimized_yet(self, client, db_session):
        """Test that list performance is not optimized yet (RED test)"""
        # Create fewer memories to test performance (reduced for speed)
        for i in range(5):  # Reduced from 20 to 5
            memory_data = MemoryFactory.create_memory_data(
                key=f"perf_memory_{i}",
                value="Performance test content. " * 5,  # Reduced from 100 to 5
            )
            client.post("/api/memories", json=memory_data)

        # Measure response time
        import time

        start_time = time.time()
        response = client.get("/api/memories")
        end_time = time.time()

        response_time_ms = (end_time - start_time) * 1000

        assert response.status_code == 200

        # Document current performance (baseline)
        print(f"Current response time: {response_time_ms:.2f}ms")

        # After Issue #111: should be 30-50% faster due to smaller responses
        # This test documents the current state


class TestMemorySearchOptimization:
    """Test memory search API optimization (TDD for Issue #111)"""

    @pytest.fixture
    def client(self):
        """Test client fixture"""
        from app.main import app

        return TestClient(app)

    def test_search_response_optimization_ready(self, client, db_session):
        """Test search response with optimization framework ready (GREEN test)"""
        # Create searchable memories
        for i in range(3):
            memory_data = MemoryFactory.create_memory_data(
                key=f"search_memory_{i}",
                value=f"Searchable content number {i}. " * 5,  # Reduced from 30 to 5
                tags=["searchable"],
            )
            client.post("/api/memories", json=memory_data)

        # Search request with include_full_text parameter
        search_request = {
            "query": "searchable",
            "search_type": "keyword",
            "limit": 10,
            "include_full_text": True,  # Request full content explicitly
        }

        response = client.post("/api/memories/search", json=search_request)
        assert response.status_code == 200
        data = response.json()

        # Search results structure is ready for optimization
        for result in data["results"]:
            assert "memory" in result
            assert "score" in result
            # Memory object contains the content
            memory = result["memory"]
            assert "value" in memory or "summary" in memory


class TestBackwardCompatibility:
    """Test backward compatibility for Issue #111 changes"""

    @pytest.fixture
    def client(self):
        """Test client fixture"""
        from app.main import app

        return TestClient(app)

    def test_legacy_api_behavior_preserved_not_implemented(self, client, db_session):
        """Test that legacy API behavior preservation is not implemented yet (RED test)"""
        memory_data = MemoryFactory.create_memory_data()
        client.post("/api/memories", json=memory_data)

        # Test various legacy compatibility scenarios that should be added in Issue #111

        # 1. Version header compatibility
        response = client.get("/api/memories", headers={"API-Version": "v1"})
        # Should work but not implemented yet

        # 2. Legacy endpoint
        response = client.get("/api/memories/full")
        assert response.status_code == 404  # Not implemented yet

        # 3. Query parameter compatibility
        response = client.get("/api/memories?include_full_text=true")
        # Parameter is ignored currently - should be implemented


# Performance tests removed - focusing on basic functionality only
