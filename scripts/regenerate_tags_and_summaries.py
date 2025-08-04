#!/usr/bin/env python3
"""Script to regenerate tags and summaries for existing memories with improved Japanese support"""

import asyncio
import re
import sys
from pathlib import Path

# Add parent directory to path to import app modules
sys.path.append(str(Path(__file__).parent.parent))

from app.core.database import SessionLocal
from app.models.memory import Memory
from app.services.summarization import summarization_service


def extract_improved_tags(text: str) -> list[str]:
    """Extract tags with improved Japanese support"""
    # Remove common markup and symbols
    clean_text = re.sub(r'[#\*`\-_=+(){}\\[\]|<>"\';:.?,!]', " ", text.lower())

    words = clean_text.split()
    important_words = []

    for word in words:
        # Include words with 2+ characters (for Japanese) or 3+ English letters
        if len(word) >= 2 and (
            word.isalpha()
            or any(
                "\u3040" <= c <= "\u309f" or "\u30a0" <= c <= "\u30ff" or "\u4e00" <= c <= "\u9faf"
                for c in word
            )
        ):
            important_words.append(word)

    return list(set(important_words[:8]))  # Take up to 8 unique words as tags


async def regenerate_tags_and_summaries():
    """Regenerate tags and summaries for all memories"""
    db = SessionLocal()

    try:
        # Get all memories
        memories = db.query(Memory).all()
        print(f"Found {len(memories)} memories to process")

        updated_count = 0

        for i, memory in enumerate(memories, 1):
            print(f"Processing memory {i}/{len(memories)}: {memory.id}")

            updated = False

            # Regenerate summary if enabled
            if summarization_service.enabled:
                try:
                    summary = await summarization_service.generate_summary(memory.value)
                    if summary != memory.summary:
                        memory.summary = summary
                        updated = True
                        print("  Updated summary")
                except Exception as e:
                    print(f"  Summary generation failed: {e}")

            # Regenerate tags with improved logic
            try:
                new_tags = extract_improved_tags(memory.value)
                if new_tags != memory.tags_list:
                    memory.tags_list = new_tags
                    updated = True
                    print(f"  Updated tags: {new_tags}")
            except Exception as e:
                print(f"  Tag generation failed: {e}")

            if updated:
                memory.ai_processed_at = memory.updated_at
                updated_count += 1

        if updated_count > 0:
            db.commit()
            print(f"\n✅ Successfully updated {updated_count} memories")
        else:
            print("\n✅ No updates needed - all memories already have current tags and summaries")

    except Exception as e:
        print(f"❌ Error: {e}")
        db.rollback()
    finally:
        db.close()


if __name__ == "__main__":
    print("🚀 Starting tags and summaries regeneration...")
    asyncio.run(regenerate_tags_and_summaries())
    print("🎉 Complete!")
