"""Custom assertions for testing"""

import asyncio
import time
from functools import wraps
from typing import Any

from app.models.memory import Memory


class MemoryAssertions:
    """Custom assertions for Memory objects"""

    @staticmethod
    def assert_memory_fields(memory: Memory, expected_data: dict[str, Any]):
        """Assert memory has expected field values"""
        assert memory.category == expected_data["category"]
        assert memory.key == expected_data["key"]
        assert memory.value == expected_data["value"]
        assert memory.tags_list == expected_data["tags"]

        if "summary" in expected_data:
            assert memory.summary == expected_data["summary"]

        if "summary_generated_at" in expected_data:
            assert memory.summary_generated_at == expected_data["summary_generated_at"]

    @staticmethod
    def assert_memory_response(response, expected_data: dict[str, Any]):
        """Assert MemoryResponse has expected structure and values"""
        # Handle both dict and MemoryResponse objects
        if isinstance(response, dict):
            data = response
        else:
            data = response.dict() if hasattr(response, "dict") else response

        assert data["id"] is not None
        assert data["category"] == expected_data["category"]
        assert data["key"] == expected_data["key"]
        assert data["value"] == expected_data["value"]
        assert data["tags"] == expected_data["tags"]
        assert "created_at" in data
        assert "updated_at" in data

        if "summary" in expected_data:
            assert response.summary == expected_data["summary"]

        if "has_embedding" in expected_data:
            assert response.has_embedding == expected_data["has_embedding"]

    @staticmethod
    def assert_summary_response(response: dict[str, Any], expected_summary: str | None = None):
        """Assert response contains summary but not full content"""
        assert "summary" in response
        assert "value" not in response or response["value"] is None

        if expected_summary:
            assert response["summary"] == expected_summary

    @staticmethod
    def assert_detail_response(response: dict[str, Any], expected_value: str):
        """Assert response contains full content"""
        assert "value" in response
        assert response["value"] == expected_value
        assert "summary" in response  # Summary should also be included


class APIAssertions:
    """Custom assertions for API responses"""

    @staticmethod
    def assert_api_response_structure(response: dict[str, Any], expected_fields: list[str]):
        """Assert API response has expected structure"""
        for field in expected_fields:
            assert field in response, f"Expected field '{field}' not found in response"

    @staticmethod
    def assert_pagination_response(
        response: dict[str, Any],
        expected_total: int,
        expected_items_count: int,
        has_pagination: bool = True,
    ):
        """Assert pagination response structure"""
        assert "total" in response
        assert response["total"] == expected_total

        if "memories" in response:
            assert len(response["memories"]) == expected_items_count
        elif "results" in response:
            assert len(response["results"]) == expected_items_count

        if has_pagination:
            # Check that pagination fields exist in original request context
            pass  # Will be expanded based on actual pagination implementation

    @staticmethod
    def assert_error_response(
        response: dict[str, Any], expected_status: int, expected_message: str = None
    ):
        """Assert error response structure"""
        assert "detail" in response or "error" in response

        if expected_message:
            error_msg = response.get("detail") or response.get("error", {}).get("message", "")
            assert expected_message in str(error_msg)

    @staticmethod
    def assert_performance_response(response: dict[str, Any], max_time_ms: float):
        """Assert response time performance"""
        if "execution_time_ms" in response:
            assert response["execution_time_ms"] <= max_time_ms, (
                f"Response took {response['execution_time_ms']}ms, expected <={max_time_ms}ms"
            )


class SummaryAssertions:
    """Custom assertions for summary functionality"""

    @staticmethod
    def assert_summary_quality(summary: str, original_text: str, max_length: int = 200):
        """Assert summary meets quality criteria"""
        assert summary is not None
        assert len(summary) > 0
        assert len(summary) <= max_length, f"Summary too long: {len(summary)} > {max_length}"

        # Check that summary is not just truncated text
        if len(original_text) > max_length:
            assert summary != original_text[:max_length], "Summary appears to be simple truncation"

        # Summary should start with common prefixes for Japanese
        if any(char in original_text for char in "あいうえおかきくけこ"):
            assert any(prefix in summary for prefix in ["要約:", "概要:", "まとめ:"]), (
                "Japanese summary should have appropriate prefix"
            )

    @staticmethod
    def assert_summary_generation_time(generation_time_ms: float, max_time_ms: float = 5000):
        """Assert summary generation time is reasonable"""
        assert generation_time_ms <= max_time_ms, (
            f"Summary generation took {generation_time_ms}ms, expected <={max_time_ms}ms"
        )

    @staticmethod
    def assert_batch_summary_results(results: dict[str, str], input_count: int):
        """Assert batch summary results"""
        assert len(results) == input_count

        for _key, summary in results.items():
            assert summary is not None
            assert len(summary) > 0
            # Allow error results in batch processing
            if not summary.startswith("Error:"):
                SummaryAssertions.assert_summary_quality(summary, "dummy text")


class TestHelpers:
    """Helper functions for testing"""

    @staticmethod
    def extract_response_data(response, field: str = "data"):
        """Extract data from API response"""
        if hasattr(response, "json"):
            data = response.json()
        else:
            data = response

        if field and field in data:
            return data[field]
        return data

    @staticmethod
    def create_test_memory_in_db(db_session, memory_data: dict[str, Any]) -> Memory:
        """Create and persist memory in test database"""
        memory = Memory(**memory_data)
        db_session.add(memory)
        db_session.commit()
        db_session.refresh(memory)
        return memory

    @staticmethod
    def measure_execution_time(func):
        """Decorator to measure function execution time"""

        @wraps(func)
        async def async_wrapper(*args, **kwargs):
            start_time = time.time()
            result = await func(*args, **kwargs)
            end_time = time.time()
            execution_time = (end_time - start_time) * 1000  # Convert to milliseconds
            return result, execution_time

        @wraps(func)
        def sync_wrapper(*args, **kwargs):
            start_time = time.time()
            result = func(*args, **kwargs)
            end_time = time.time()
            execution_time = (end_time - start_time) * 1000  # Convert to milliseconds
            return result, execution_time

        if asyncio.iscoroutinefunction(func):
            return async_wrapper
        else:
            return sync_wrapper
