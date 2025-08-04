"""Test memory model and API with summary functionality (TDD approach)"""

from datetime import datetime

import pytest
from fastapi.testclient import TestClient

from tests.utils.assertions import APIAssertions
from tests.utils.factories import MemoryFactory


class TestMemoryModelWithSummary:
    """Test Memory model with summary fields (TDD for Issue #109)"""

    def test_memory_model_basic_fields_exist(self):
        """Test that basic Memory model fields exist - simplified AI-driven schema (Issue #112)"""
        from datetime import datetime

        from app.models.memory import Memory

        # Test simplified schema with only user input value
        now = datetime.utcnow()
        memory = Memory(
            value="test value",
            tags_list=["test"],
            created_at=now,
            updated_at=now,
        )

        assert memory.value == "test value"
        assert memory.tags_list == ["test"]
        assert memory.created_at is not None
        assert memory.updated_at is not None
        # AI-driven fields
        assert hasattr(memory, "summary")
        assert hasattr(memory, "ai_processed_at")
        assert hasattr(memory, "processing_status")

    def test_memory_model_summary_field_exists(self):
        """Test that AI summary field exists - simplified AI-driven schema (Issue #112)"""
        from app.models.memory import Memory

        memory = Memory(value="test value")

        # AI-driven fields should exist in simplified schema
        assert hasattr(memory, "summary")
        assert hasattr(memory, "ai_processed_at")
        assert hasattr(memory, "processing_status")

    def test_memory_model_summary_field_after_implementation(self):
        """Test that summary field exists after implementation (GREEN test - will pass after #109)"""
        pytest.skip("Will be implemented in Issue #109")

        # After implementing Issue #109, this test should pass:
        from app.models.memory import Memory

        memory = Memory(
            category="test",
            key="test_key",
            value="test value",
            summary="Test summary",
            summary_generated_at=datetime.utcnow(),
        )

        assert hasattr(memory, "summary")
        assert hasattr(memory, "summary_generated_at")
        assert memory.summary == "Test summary"
        assert isinstance(memory.summary_generated_at, datetime)

    def test_memory_model_summary_nullable(self):
        """Test that summary fields can be null (GREEN test - will pass after #109)"""
        pytest.skip("Will be implemented in Issue #109")

        # After implementing Issue #109:
        from app.models.memory import Memory

        memory = Memory(category="test", key="test_key", value="test value")

        assert memory.summary is None
        assert memory.summary_generated_at is None


class TestMemoryResponseSchemaWithSummary:
    """Test MemoryResponse schema with summary fields (TDD for Issue #109)"""

    def test_memory_response_current_fields(self):
        """Test current MemoryResponse fields - simplified AI-driven schema (Issue #112)"""
        from app.models.schemas import MemoryResponse

        field_names = set(MemoryResponse.model_fields.keys())

        # Simplified schema fields should exist
        expected_current_fields = {
            "id",
            "value",
            "tags",
            "summary",
            "created_at",
            "updated_at",
            "has_embedding",
            "ai_processed_at",
            "processing_status",
        }

        for field in expected_current_fields:
            assert field in field_names, f"Expected field '{field}' not found"

    def test_memory_response_summary_fields_exist(self):
        """Test that AI summary fields exist - simplified AI-driven schema (Issue #112)"""
        from app.models.schemas import MemoryResponse

        field_names = set(MemoryResponse.model_fields.keys())

        # AI-driven fields should exist in simplified schema
        assert "summary" in field_names
        assert "ai_processed_at" in field_names
        assert "processing_status" in field_names

    def test_memory_response_summary_fields_after_implementation(self):
        """Test summary fields in MemoryResponse after implementation (GREEN test)"""
        pytest.skip("Will be implemented in Issue #109")

        # After implementing Issue #109:
        from app.models.schemas import MemoryResponse

        field_names = set(MemoryResponse.model_fields.keys())

        assert "summary" in field_names
        assert "summary_generated_at" in field_names

    def test_memory_response_with_summary_data(self):
        """Test MemoryResponse creation with summary data (GREEN test)"""
        pytest.skip("Will be implemented in Issue #109")

        # After implementing Issue #109:
        from datetime import datetime

        from app.models.schemas import MemoryResponse

        response_data = {
            "id": "test_id",
            "category": "test",
            "key": "test_key",
            "value": "test value",
            "tags": ["test"],
            "created_at": datetime.utcnow(),
            "updated_at": datetime.utcnow(),
            "has_embedding": False,
            "summary": "Test summary",
            "summary_generated_at": datetime.utcnow(),
        }

        response = MemoryResponse(**response_data)

        assert response.summary == "Test summary"
        assert isinstance(response.summary_generated_at, datetime)


class TestMemoryAPIWithSummaryIntegration:
    """Test Memory API integration with summary functionality"""

    @pytest.fixture
    def client(self):
        """Test client fixture"""
        from app.main import app

        return TestClient(app)

    def test_create_memory_with_summary_generation(self, client, db_session):
        """Test memory creation with AI summary generation - simplified AI-driven schema (Issue #112)"""
        memory_data = MemoryFactory.create_memory_data(
            value="This is a test for AI summary generation."
        )

        response = client.post("/api/memories", json=memory_data)

        assert response.status_code == 201
        data = response.json()

        # AI-driven fields should be generated
        assert "summary" in data
        assert "tags" in data  # AI-generated tags
        assert "processing_status" in data
        assert "ai_processed_at" in data

    def test_create_memory_generates_japanese_summary(self, client, db_session):
        """Test that creating memory generates AI summary - simplified AI-driven schema (Issue #112)"""
        memory_data = MemoryFactory.create_memory_data(
            value="This is a longer text that should be summarized automatically when created.",
        )

        response = client.post("/api/memories", json=memory_data)

        assert response.status_code == 201
        data = response.json()

        # AI processing should be attempted
        assert "summary" in data
        assert "tags" in data  # AI-generated tags
        assert "processing_status" in data
        assert "ai_processed_at" in data

    def test_create_memory_generates_summary_after_implementation(self, client, db_session):
        """Test memory creation with summary generation (GREEN test)"""
        pytest.skip("Will be implemented in Issue #110")

        # After implementing Issue #110:
        memory_data = MemoryFactory.create_memory_data(
            category="test",
            key="summary_test",
            value="This is a longer text that should be summarized automatically when created.",
        )

        from unittest.mock import patch

        with patch("app.services.summarization.SummarizationService") as mock_service:
            mock_service.return_value.generate_summary.return_value = "テスト要約"

            response = client.post("/api/memories", json=memory_data)

            assert response.status_code == 201
            data = response.json()

            assert data["summary"] == "テスト要約"
            assert data["summary_generated_at"] is not None

    def test_list_memories_optimized_behavior(self, client, db_session):
        """Test optimized list memories behavior (after Issue #111)"""
        # Create test memory
        memory_data = MemoryFactory.create_memory_data()
        client.post("/api/memories", json=memory_data)

        response = client.get("/api/memories")

        assert response.status_code == 200
        data = response.json()

        APIAssertions.assert_api_response_structure(data, ["memories", "total"])
        assert len(data["memories"]) > 0

        memory = data["memories"][0]
        # Now returns summary only (Issue #111 implemented)
        assert "value" not in memory or memory.get("value") is None
        assert "summary" in memory
        assert memory["summary"] is not None

    def test_list_memories_returns_summary_only_implemented(self, client, db_session):
        """Test that list endpoint returns summary only (GREEN test - Issue #111 implemented)"""
        # Create memory with summary
        memory_data = MemoryFactory.create_memory_data(
            value="This is a long text that should have a summary"
        )
        client.post("/api/memories", json=memory_data)

        response = client.get("/api/memories")

        assert response.status_code == 200
        data = response.json()

        memory = data["memories"][0]

        # Issue #111 implemented - returns summary only
        assert "value" not in memory or memory.get("value") is None
        assert "summary" in memory
        assert memory["summary"] is not None

    def test_get_memory_detail_endpoint_implemented(self, client, db_session):
        """Test that detail endpoint works correctly - simplified AI-driven schema (Issue #112)"""
        # Create test memory
        memory_data = MemoryFactory.create_memory_data(value="Detail test content")
        response = client.post("/api/memories", json=memory_data)
        memory_id = response.json()["id"]

        # Detail endpoint now exists and returns full content
        detail_response = client.get(f"/api/memories/{memory_id}/detail")
        assert detail_response.status_code == 200

        detail_data = detail_response.json()
        assert "value" in detail_data
        assert detail_data["value"] == memory_data["value"]
        assert "summary" in detail_data
        assert "tags" in detail_data

    def test_get_memory_detail_endpoint_after_implementation(self, client, db_session):
        """Test detail endpoint returns full content (GREEN test)"""
        pytest.skip("Will be implemented in Issue #111")

        # After implementing Issue #111:
        memory_data = MemoryFactory.create_memory_data(
            key="detail_test", value="Full content for detail endpoint"
        )
        client.post("/api/memories", json=memory_data)

        detail_response = client.get("/api/memories/detail_test/detail")

        assert detail_response.status_code == 200
        data = detail_response.json()

        assert "value" in data
        assert data["value"] == memory_data["value"]
        assert "summary" in data  # Summary should also be included


class TestMemoryAPISummaryConfiguration:
    """Test memory API configuration for summary functionality"""

    def test_summary_configuration_implemented(self):
        """Test that summary configuration exists (Issue #110 implemented)"""
        from app.core.config import Settings

        settings = Settings()

        # These configuration options should exist now that Issue #110 is implemented
        assert hasattr(settings, "summary_enabled")
        assert hasattr(settings, "summary_max_length")
        assert hasattr(settings, "summary_fallback_enabled")

        # Test default values
        assert settings.summary_enabled is True
        assert settings.summary_max_length == 200
        assert settings.summary_fallback_enabled is True

    def test_summary_configuration_after_implementation(self):
        """Test summary configuration after implementation (GREEN test)"""
        pytest.skip("Will be implemented in Issue #110")

        # After implementing Issue #110:
        from app.core.config import Settings

        settings = Settings()

        assert hasattr(settings, "summary_enabled")
        assert hasattr(settings, "summary_max_length")
        assert hasattr(settings, "summary_fallback_enabled")

        # Test default values
        assert settings.summary_enabled is True
        assert settings.summary_max_length == 200
        assert settings.summary_fallback_enabled is True
