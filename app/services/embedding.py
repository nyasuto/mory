"""Embedding service for generating and managing vector embeddings"""

import numpy as np
import openai
from sqlalchemy.orm import Session

from ..core.config import settings
from ..models.memory import Memory


class EmbeddingService:
    """Service for generating vector embeddings"""

    def __init__(self) -> None:
        """Initialize embedding service"""
        self.enabled = settings.is_semantic_available
        if self.enabled:
            openai.api_key = settings.openai_api_key

    async def generate_embedding(self, text: str) -> np.ndarray | None:
        """Generate embedding vector for given text

        Args:
            text: Text to generate embedding for

        Returns:
            numpy array of embedding vector, or None if service disabled

        """
        if not self.enabled or not text.strip():
            return None

        try:
            response = openai.embeddings.create(model=settings.openai_model, input=text)
            embedding_vector = response.data[0].embedding
            return np.array(embedding_vector, dtype=np.float32)
        except Exception as e:
            print(f"Embedding generation failed: {e}")
            return None

    async def generate_embedding_for_memory(self, memory: Memory) -> bool:
        """Generate and store embedding for a memory

        Args:
            memory: Memory object to generate embedding for

        Returns:
            True if embedding was generated and stored, False otherwise

        """
        if not self.enabled:
            return False

        # Use summary if available, otherwise use original value
        text_for_embedding = memory.summary or memory.value

        embedding = await self.generate_embedding(text_for_embedding)
        if embedding is not None:
            memory.embedding = embedding.tobytes()
            memory.embedding_model = settings.openai_model
            return True

        return False

    async def generate_embeddings_batch(self, memories: list[Memory], db: Session) -> int:
        """Generate embeddings for multiple memories

        Args:
            memories: List of Memory objects
            db: Database session

        Returns:
            Number of embeddings successfully generated

        """
        if not self.enabled:
            return 0

        generated_count = 0

        for memory in memories:
            if await self.generate_embedding_for_memory(memory):
                generated_count += 1

        if generated_count > 0:
            db.commit()

        return generated_count


# Global embedding service instance
embedding_service = EmbeddingService()
