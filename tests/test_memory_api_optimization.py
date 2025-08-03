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

    def test_list_memories_current_behavior_returns_full_content(
        self, client, db_session, sample_memories_data
    ):
        """Test current list behavior returns full content (RED test)"""
        # Create test memories
        for memory_data in sample_memories_data:
            client.post("/api/memories", json=memory_data)

        # Current behavior: list returns full content
        response = client.get("/api/memories")

        assert response.status_code == 200
        data = response.json()

        APIAssertions.assert_api_response_structure(data, ["memories", "total"])
        assert len(data["memories"]) == len(sample_memories_data)

        # Currently returns full value - this should change in Issue #111
        for memory in data["memories"]:
            assert "value" in memory
            assert memory["value"] is not None
            assert len(memory["value"]) > 0  # Full content is included

    def test_list_memories_should_return_summary_only_not_implemented(
        self, client, db_session, sample_memories_data
    ):
        """Test that list should return summary only (RED test - will fail until #111 implemented)"""
        # Create test memories
        for memory_data in sample_memories_data:
            client.post("/api/memories", json=memory_data)

        response = client.get("/api/memories")

        assert response.status_code == 200
        data = response.json()

        # This will fail until Issue #111 is implemented
        for memory in data["memories"]:
            # After Issue #111: list should return summary, not full value
            assert "value" not in memory or memory["value"] is None
            assert "summary" in memory
            assert memory["summary"] is not None

    def test_memory_list_response_schema_missing_summary_fields(self, client, db_session):
        """Test that list response doesn't use summary schema yet (RED test)"""
        # Create a memory
        memory_data = MemoryFactory.create_memory_data()
        client.post("/api/memories", json=memory_data)

        response = client.get("/api/memories")
        data = response.json()

        # Current response structure - will change in Issue #111
        assert "memories" in data

        # These assertions will fail until we implement MemorySummaryResponse
        try:
            # Try to import the new schema - should fail until implemented
            from app.models.schemas import MemorySummaryResponse  # noqa: F401

            pytest.fail("MemorySummaryResponse should not exist yet")
        except ImportError:
            pass  # Expected until Issue #111 implementation

    def test_backward_compatibility_include_full_text_parameter_not_implemented(
        self, client, db_session
    ):
        """Test backward compatibility parameter doesn't exist yet (RED test)"""
        memory_data = MemoryFactory.create_memory_data()
        client.post("/api/memories", json=memory_data)

        # This parameter should be added in Issue #111 for backward compatibility
        response = client.get("/api/memories?include_full_text=true")

        # Currently this parameter is ignored - should be implemented in Issue #111
        assert response.status_code == 200
        data = response.json()

        # The parameter is currently ignored - this test documents desired behavior
        for memory in data["memories"]:
            # After Issue #111: include_full_text=true should return full content
            assert "value" in memory  # Currently always included


class TestMemoryDetailEndpoint:
    """Test new memory detail endpoint (TDD for Issue #111)"""

    @pytest.fixture
    def client(self):
        """Test client fixture"""
        from app.main import app

        return TestClient(app)

    def test_memory_detail_endpoint_not_implemented(self, client, db_session):
        """Test that detail endpoint doesn't exist yet (RED test)"""
        # Create test memory
        memory_data = MemoryFactory.create_memory_data(key="detail_test")
        response = client.post("/api/memories", json=memory_data)
        memory_key = response.json()["key"]

        # Detail endpoint should not exist yet
        detail_response = client.get(f"/api/memories/{memory_key}/detail")
        assert detail_response.status_code == 404

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

    def test_response_size_not_optimized_yet(self, client, db_session):
        """Test that response size is not optimized yet (RED test)"""
        # Create memories with large content
        large_memories = []
        for i in range(5):
            memory_data = MemoryFactory.create_memory_data(
                key=f"large_memory_{i}",
                value="This is a very large memory content that will increase response size significantly. "
                * 50,
                tags=["performance", "large"],
            )
            large_memories.append(memory_data)
            client.post("/api/memories", json=memory_data)

        # Get current response size
        response = client.get("/api/memories")
        assert response.status_code == 200

        # Measure response size (this is the baseline before optimization)
        import json

        response_size = len(json.dumps(response.json()).encode("utf-8"))

        # Store baseline for comparison (this test documents current state)
        # After Issue #111: response size should be 80-90% smaller
        print(f"Current response size: {response_size} bytes")

        # This assertion will fail after Issue #111 implementation
        # when response size is optimized (which is the goal)
        assert response_size > 10000  # Large response due to full content

    def test_performance_not_optimized_yet(self, client, db_session):
        """Test that list performance is not optimized yet (RED test)"""
        # Create many memories to test performance
        for i in range(20):
            memory_data = MemoryFactory.create_memory_data(
                key=f"perf_memory_{i}",
                value="Performance test content. " * 100,
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

    def test_search_response_not_optimized_yet(self, client, db_session):
        """Test that search response is not optimized yet (RED test)"""
        # Create searchable memories
        for i in range(3):
            memory_data = MemoryFactory.create_memory_data(
                key=f"search_memory_{i}",
                value=f"Searchable content number {i}. " * 30,
                tags=["searchable"],
            )
            client.post("/api/memories", json=memory_data)

        # Search request
        search_request = {"query": "searchable", "search_type": "keyword", "limit": 10}

        response = client.post("/api/memories/search", json=search_request)
        assert response.status_code == 200
        data = response.json()

        # Currently returns full content in search results
        for result in data["results"]:
            assert "value" in result
            assert result["value"] is not None
            assert len(result["value"]) > 100  # Full content

        # After Issue #111: search should have include_full_text parameter
        # and return summaries by default


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


@pytest.mark.performance
class TestAPIOptimizationPerformance:
    """Performance tests for API optimization (TDD for Issue #111)"""

    @pytest.fixture
    def client(self):
        """Test client fixture"""
        from app.main import app

        return TestClient(app)

    def test_baseline_performance_measurement(self, client, db_session):
        """Measure baseline performance before optimization (RED test)"""
        # Create dataset for performance testing
        for i in range(50):
            memory_data = MemoryFactory.create_memory_data(
                key=f"perf_test_{i}",
                value="Performance testing content. " * 200,  # Large content
                tags=["performance", f"batch_{i // 10}"],
            )
            client.post("/api/memories", json=memory_data)

        # Measure various performance metrics
        import json
        import time

        # List performance
        start_time = time.time()
        response = client.get("/api/memories?limit=50")
        list_time = (time.time() - start_time) * 1000

        # Response size
        response_size = len(json.dumps(response.json()).encode("utf-8"))

        # Search performance
        search_request = {"query": "performance", "search_type": "keyword", "limit": 20}
        start_time = time.time()
        search_response = client.post("/api/memories/search", json=search_request)
        search_time = (time.time() - start_time) * 1000

        # Document baseline metrics
        print("Baseline metrics:")
        print(f"  List time: {list_time:.2f}ms")
        print(f"  Response size: {response_size} bytes")
        print(f"  Search time: {search_time:.2f}ms")

        # These will be the targets for optimization in Issue #111
        # After implementation:
        # - List time should decrease by 30-50%
        # - Response size should decrease by 80-90%
        # - Search time should decrease by 30-50%

        assert response.status_code == 200
        assert search_response.status_code == 200
