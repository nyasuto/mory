#!/usr/bin/env python3
"""Script to remove unused embedding_hash column and index from database"""

import sys
from pathlib import Path

# Add parent directory to path to import app modules
sys.path.append(str(Path(__file__).parent.parent))

from sqlalchemy import text

from app.core.database import SessionLocal


def remove_embedding_hash():
    """Remove embedding_hash column and its index from the database"""
    db = SessionLocal()

    try:
        print("🔍 Checking current schema...")

        # Check if embedding_hash column exists
        result = db.execute(text("PRAGMA table_info(memories)")).fetchall()
        columns = [row[1] for row in result]

        if "embedding_hash" not in columns:
            print("✅ embedding_hash column doesn't exist, nothing to remove")
            return

        print("📋 Current columns:", columns)

        # Check if index exists
        indexes = db.execute(text("PRAGMA index_list(memories)")).fetchall()
        index_names = [row[1] for row in indexes]
        print("📋 Current indexes:", index_names)

        print("\n🗑️  Removing embedding_hash column and index...")

        # SQLite doesn't support DROP COLUMN directly, so we need to recreate the table
        print("  1. Creating backup table...")
        db.execute(
            text("""
            CREATE TABLE memories_backup AS
            SELECT id, value, summary, tags, created_at, updated_at, ai_processed_at, embedding, embedding_model
            FROM memories
        """)
        )

        print("  2. Dropping original table...")
        db.execute(text("DROP TABLE memories"))

        print("  3. Creating new table without embedding_hash...")
        db.execute(
            text("""
            CREATE TABLE memories (
                id VARCHAR PRIMARY KEY,
                value TEXT NOT NULL,
                summary TEXT,
                tags TEXT DEFAULT '[]',
                created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                ai_processed_at DATETIME,
                embedding BLOB,
                embedding_model VARCHAR
            )
        """)
        )

        print("  4. Restoring data...")
        db.execute(
            text("""
            INSERT INTO memories (id, value, summary, tags, created_at, updated_at, ai_processed_at, embedding, embedding_model)
            SELECT id, value, summary, tags, created_at, updated_at, ai_processed_at, embedding, embedding_model
            FROM memories_backup
        """)
        )

        print("  5. Creating indexes...")
        db.execute(text("CREATE INDEX idx_updated_at ON memories (updated_at)"))
        db.execute(text("CREATE INDEX idx_ai_processed ON memories (ai_processed_at)"))
        db.execute(text("CREATE INDEX idx_tags_search ON memories (tags)"))

        print("  6. Cleaning up backup table...")
        db.execute(text("DROP TABLE memories_backup"))

        db.commit()

        print("\n✅ Successfully removed embedding_hash column and index!")

        # Verify the changes
        result = db.execute(text("PRAGMA table_info(memories)")).fetchall()
        new_columns = [row[1] for row in result]
        print("📋 New columns:", new_columns)

        indexes = db.execute(text("PRAGMA index_list(memories)")).fetchall()
        new_index_names = [row[1] for row in indexes]
        print("📋 New indexes:", new_index_names)

        # Check data integrity
        count = db.execute(text("SELECT COUNT(*) FROM memories")).scalar()
        print(f"📊 Total memories: {count}")

        embedding_count = db.execute(
            text("SELECT COUNT(*) FROM memories WHERE embedding IS NOT NULL")
        ).scalar()
        print(f"📊 Memories with embeddings: {embedding_count}")

    except Exception as e:
        print(f"❌ Error: {e}")
        db.rollback()

        # Try to restore from backup if it exists
        try:
            tables = db.execute(
                text("SELECT name FROM sqlite_master WHERE type='table'")
            ).fetchall()
            table_names = [row[0] for row in tables]
            if "memories_backup" in table_names:
                print("🔄 Attempting to restore from backup...")
                db.execute(text("DROP TABLE IF EXISTS memories"))
                db.execute(text("ALTER TABLE memories_backup RENAME TO memories"))
                db.commit()
                print("✅ Restored from backup")
        except Exception as restore_error:
            print(f"❌ Failed to restore from backup: {restore_error}")
    finally:
        db.close()


if __name__ == "__main__":
    print("🚀 Starting embedding_hash removal...")
    remove_embedding_hash()
    print("🎉 Complete!")
