#!/usr/bin/env python3
"""Script to generate missing vector embeddings for existing memories"""

import asyncio
import sys
from pathlib import Path

# Add parent directory to path to import app modules
sys.path.append(str(Path(__file__).parent.parent))

from app.core.database import SessionLocal
from app.models.memory import Memory
from app.services.embedding import embedding_service


async def generate_missing_embeddings():
    """Generate embeddings for memories that don't have them"""
    db = SessionLocal()

    try:
        # Get all memories without embeddings
        memories_without_embeddings = db.query(Memory).filter(Memory.embedding.is_(None)).all()

        total_count = len(memories_without_embeddings)
        print(f"Found {total_count} memories without embeddings")

        if total_count == 0:
            print("✅ All memories already have embeddings!")
            return

        if not embedding_service.enabled:
            print("❌ Embedding service is not enabled (OpenAI API key not configured)")
            return

        generated_count = 0
        failed_count = 0

        for i, memory in enumerate(memories_without_embeddings, 1):
            print(f"Processing memory {i}/{total_count}: {memory.id}")

            try:
                # Generate embedding using the text from summary if available, otherwise value
                text_for_embedding = memory.summary or memory.value
                print(f"  Text preview: {text_for_embedding[:100]}...")

                embedding_generated = await embedding_service.generate_embedding_for_memory(memory)
                if embedding_generated:
                    print("  ✅ Successfully generated embedding")
                    generated_count += 1
                else:
                    print("  ❌ Failed to generate embedding")
                    failed_count += 1

            except Exception as e:
                print(f"  ❌ Error generating embedding: {e}")
                failed_count += 1

        # Commit all changes at once
        if generated_count > 0:
            db.commit()
            print(f"\n🎉 Successfully generated embeddings for {generated_count} memories")

        if failed_count > 0:
            print(f"⚠️  Failed to generate embeddings for {failed_count} memories")

        print("\n📊 Summary:")
        print(f"  Total processed: {total_count}")
        print(f"  Successfully generated: {generated_count}")
        print(f"  Failed: {failed_count}")

    except Exception as e:
        print(f"❌ Error: {e}")
        db.rollback()
    finally:
        db.close()


if __name__ == "__main__":
    print("🚀 Starting missing embeddings generation...")
    asyncio.run(generate_missing_embeddings())
    print("🎉 Complete!")
