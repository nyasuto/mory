"""Mock services for testing"""

import asyncio
from unittest.mock import AsyncMock, MagicMock


class MockOpenAIService:
    """Mock OpenAI service for testing without API calls"""

    def __init__(self):
        """Initialize mock OpenAI service"""
        self.call_count = 0
        self.should_fail = False
        self.fail_count = 0

    async def generate_summary(self, text: str, max_length: int = 200) -> str:
        """Generate deterministic summary for testing"""
        self.call_count += 1

        if self.should_fail:
            self.fail_count += 1
            raise Exception("Mock OpenAI API failure")

        # Simulate processing delay
        await asyncio.sleep(0.001)

        # Generate deterministic summary based on text
        if len(text) <= max_length:
            return f"要約: {text}"

        # Create a meaningful summary for testing
        summary = f"要約: {text[: max_length - 20]}..."
        return summary

    async def generate_embedding(self, text: str) -> list[float]:
        """Generate deterministic embedding for testing"""
        self.call_count += 1

        if self.should_fail:
            self.fail_count += 1
            raise Exception("Mock OpenAI API failure")

        # Simulate processing delay
        await asyncio.sleep(0.001)

        # Generate deterministic embedding based on text hash
        text_hash = hash(text) % 1000
        base_vector = [float(text_hash % 100) / 100.0] * 1536  # OpenAI embedding dimension

        # Add some variation based on text length
        length_factor = len(text) % 10 / 10.0
        return [v + length_factor * 0.1 for v in base_vector]

    async def batch_generate_summaries(self, texts: list[str]) -> dict[str, str]:
        """Generate multiple summaries in batch"""
        results = {}
        for i, text in enumerate(texts):
            try:
                summary = await self.generate_summary(text)
                results[f"text_{i}"] = summary
            except Exception as e:
                results[f"text_{i}"] = f"Error: {str(e)}"
        return results

    def reset(self):
        """Reset mock state"""
        self.call_count = 0
        self.should_fail = False
        self.fail_count = 0

    def set_failure_mode(self, should_fail: bool = True):
        """Set the service to fail for testing error handling"""
        self.should_fail = should_fail


class MockSearchService:
    """Mock search service for testing"""

    def __init__(self):
        """Initialize mock search service"""
        self.search_results = []
        self.search_time_ms = 10.0

    async def search_memories(self, request, db):
        """Mock search that returns predefined results"""
        await asyncio.sleep(self.search_time_ms / 1000.0)  # Simulate search time

        from app.models.schemas import SearchResponse

        return SearchResponse(
            results=self.search_results,
            total=len(self.search_results),
            query=request.query,
            search_type="mock",
            execution_time_ms=self.search_time_ms,
            filters={},
        )

    def set_results(self, results):
        """Set predefined search results"""
        self.search_results = results

    def set_search_time(self, time_ms: float):
        """Set simulated search time"""
        self.search_time_ms = time_ms


def create_mock_openai_client():
    """Create a mock OpenAI client for testing"""
    mock_client = MagicMock()

    # Mock embeddings
    mock_embeddings_response = MagicMock()
    mock_embeddings_response.data = [MagicMock()]
    mock_embeddings_response.data[0].embedding = [0.1] * 1536

    mock_client.embeddings.create = AsyncMock(return_value=mock_embeddings_response)

    # Mock chat completions for summarization
    mock_completion_response = MagicMock()
    mock_completion_response.choices = [MagicMock()]
    mock_completion_response.choices[0].message.content = "テスト要約"

    mock_client.chat.completions.create = AsyncMock(return_value=mock_completion_response)

    return mock_client


def create_test_config():
    """Create test configuration"""
    return {
        "summary_enabled": True,
        "summary_model": "gpt-4-turbo",
        "summary_max_length": 200,
        "summary_fallback_enabled": True,
        "openai_api_key": "test-key-123",
        "openai_model": "text-embedding-3-large",
    }
