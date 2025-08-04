#!/usr/bin/env python3
"""
Database migration script for Issue #112: Simplified AI-driven Memory Schema

This script migrates from the old complex schema (category/key/tags) to the new
simplified AI-driven schema where only 'value' is user input and AI generates
comprehensive tags and summaries.

Migration Strategy:
1. Add new columns (summary, ai_processed_at, embedding_model)
2. Migrate existing data: combine category/key into comprehensive AI tags
3. Remove old columns (category, key) - keeping tags as AI-generated
4. Update indexes for new schema

Usage:
    python scripts/migrate_to_simplified_schema.py [--dry-run] [--backup]
"""

import argparse
import json
import logging
import shutil
import sqlite3
from datetime import datetime
from pathlib import Path
from typing import Any

# Configure logging
logging.basicConfig(level=logging.INFO, format="%(asctime)s - %(levelname)s - %(message)s")
logger = logging.getLogger(__name__)


class SimplifiedSchemaMigration:
    """Handles migration to simplified AI-driven schema"""

    def __init__(self, db_path: str, backup: bool = True):
        """Initialize migration with database path and backup option."""
        self.db_path = Path(db_path)
        self.backup = backup
        self.backup_path = None

        if not self.db_path.exists():
            raise FileNotFoundError(f"Database not found: {self.db_path}")

    def create_backup(self) -> None:
        """Create database backup before migration"""
        if not self.backup:
            return

        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        self.backup_path = self.db_path.with_suffix(f".backup_{timestamp}.db")

        logger.info(f"Creating backup: {self.backup_path}")
        shutil.copy2(self.db_path, self.backup_path)

    def get_current_schema_info(self, conn: sqlite3.Connection) -> dict[str, Any]:
        """Get information about current schema"""
        cursor = conn.cursor()

        # Check if we have old schema columns
        cursor.execute("PRAGMA table_info(memories)")
        columns = {row[1]: row[2] for row in cursor.fetchall()}

        # Count existing data
        cursor.execute("SELECT COUNT(*) as total FROM memories")
        total_memories = cursor.fetchone()[0]

        # Check for existing categories and keys
        has_category = "category" in columns
        has_key = "key" in columns
        has_summary = "summary" in columns
        has_ai_processed_at = "ai_processed_at" in columns

        return {
            "columns": columns,
            "total_memories": total_memories,
            "has_old_schema": has_category and has_key,
            "has_new_schema": has_summary and has_ai_processed_at,
            "migration_needed": has_category and not has_ai_processed_at,
        }

    def migrate_schema(self, conn: sqlite3.Connection, dry_run: bool = False) -> None:
        """Perform the actual schema migration"""
        cursor = conn.cursor()

        logger.info("🔄 Starting schema migration to simplified AI-driven model...")

        # Step 1: Add new columns for AI-driven model
        new_columns = [
            "ALTER TABLE memories ADD COLUMN summary TEXT",
            "ALTER TABLE memories ADD COLUMN ai_processed_at DATETIME",
            "ALTER TABLE memories ADD COLUMN embedding_model TEXT",
        ]

        for sql in new_columns:
            try:
                if not dry_run:
                    cursor.execute(sql)
                logger.info(f"✅ Added column: {sql.split()[-2]}")
            except sqlite3.OperationalError as e:
                if "duplicate column name" in str(e):
                    logger.info(f"⏭️  Column already exists: {sql.split()[-2]}")
                else:
                    raise

        # Step 2: Migrate existing data - create comprehensive AI tags from category/key
        logger.info("🤖 Migrating existing data to AI-driven format...")

        # Get all existing memories with old schema
        cursor.execute("""
            SELECT id, category, key, value, tags, created_at, updated_at
            FROM memories
            WHERE category IS NOT NULL
        """)

        old_memories = cursor.fetchall()
        migrated_count = 0

        for memory in old_memories:
            id_, category, key, value, old_tags, created_at, updated_at = memory

            # Parse existing tags
            try:
                existing_tags = json.loads(old_tags) if old_tags else []
            except json.JSONDecodeError:
                existing_tags = []

            # Create comprehensive AI-style tags combining category, key, and existing tags
            comprehensive_tags = []

            # Add category as primary tag
            if category:
                comprehensive_tags.append(category.lower())

            # Add key as descriptive tag if different from category
            if key and key.lower() != category.lower():
                comprehensive_tags.append(key.lower().replace(" ", "_"))

            # Add existing tags
            comprehensive_tags.extend([tag.lower() for tag in existing_tags])

            # Add content-based tags (simple extraction for migration)
            content_words = value.lower().split()
            important_words = [
                word for word in content_words[:10] if len(word) > 3 and word.isalpha()
            ]
            comprehensive_tags.extend(important_words[:3])  # Add up to 3 content words

            # Remove duplicates and create final tag list
            final_tags = list(
                dict.fromkeys(comprehensive_tags)
            )  # Preserve order, remove duplicates

            # Create simple summary for migration (first 100 chars + "...")
            simple_summary = value[:100] + "..." if len(value) > 100 else value

            # Update memory with AI-style data
            if not dry_run:
                cursor.execute(
                    """
                    UPDATE memories
                    SET tags = ?, summary = ?, ai_processed_at = ?
                    WHERE id = ?
                """,
                    (json.dumps(final_tags), simple_summary, datetime.utcnow(), id_),
                )

            migrated_count += 1

            if migrated_count <= 5:  # Log first 5 migrations as examples
                logger.info(
                    f"📝 Migrated memory '{id_}': {category}/{key} → {len(final_tags)} AI tags"
                )

        logger.info(f"✅ Migrated {migrated_count} memories to AI-driven format")

        # Step 3: Create new indexes for optimized performance
        new_indexes = [
            "CREATE INDEX IF NOT EXISTS idx_ai_processed ON memories(ai_processed_at)",
            "CREATE INDEX IF NOT EXISTS idx_embedding_hash ON memories(embedding_hash)",
            "CREATE INDEX IF NOT EXISTS idx_tags_search ON memories(tags)",
        ]

        for sql in new_indexes:
            if not dry_run:
                cursor.execute(sql)
            logger.info(f"📊 Created index: {sql.split()[-1].replace('(', ' on ')}")

        # Step 4: Remove old schema columns (category, key)
        # Note: SQLite doesn't support DROP COLUMN directly, so we'll create a new table
        if not dry_run:
            logger.info("🗑️  Removing old schema columns (category, key)...")

            # Create new table with simplified schema
            cursor.execute("""
                CREATE TABLE memories_new (
                    id TEXT PRIMARY KEY,
                    value TEXT NOT NULL,
                    tags TEXT DEFAULT '[]',
                    summary TEXT,
                    ai_processed_at DATETIME,
                    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                    embedding BLOB,
                    embedding_hash TEXT,
                    embedding_model TEXT
                )
            """)

            # Copy data to new table (excluding category, key)
            cursor.execute("""
                INSERT INTO memories_new
                (id, value, tags, summary, ai_processed_at, created_at, updated_at,
                 embedding, embedding_hash, embedding_model)
                SELECT id, value, tags, summary, ai_processed_at, created_at, updated_at,
                       embedding, embedding_hash, embedding_model
                FROM memories
            """)

            # Replace old table with new table
            cursor.execute("DROP TABLE memories")
            cursor.execute("ALTER TABLE memories_new RENAME TO memories")

            # Recreate all indexes on new table
            all_indexes = [
                "CREATE INDEX idx_updated_at ON memories(updated_at)",
                "CREATE INDEX idx_ai_processed ON memories(ai_processed_at)",
                "CREATE INDEX idx_tags_search ON memories(tags)",
                "CREATE INDEX idx_embedding_hash ON memories(embedding_hash)",
            ]

            for sql in all_indexes:
                cursor.execute(sql)

        if not dry_run:
            conn.commit()

        logger.info("🎉 Schema migration completed successfully!")

    def verify_migration(self, conn: sqlite3.Connection) -> None:
        """Verify migration was successful"""
        cursor = conn.cursor()

        # Check new schema
        cursor.execute("PRAGMA table_info(memories)")
        columns = {row[1]: row[2] for row in cursor.fetchall()}

        required_columns = ["id", "value", "tags", "summary", "ai_processed_at"]
        missing_columns = [col for col in required_columns if col not in columns]

        if missing_columns:
            raise RuntimeError(f"Migration verification failed: missing columns {missing_columns}")

        # Check old columns are removed
        old_columns = ["category", "key"]
        remaining_old_columns = [col for col in old_columns if col in columns]

        if remaining_old_columns:
            raise RuntimeError(
                f"Migration verification failed: old columns still exist {remaining_old_columns}"
            )

        # Check data integrity
        cursor.execute("SELECT COUNT(*) FROM memories")
        total_count = cursor.fetchone()[0]

        cursor.execute("SELECT COUNT(*) FROM memories WHERE ai_processed_at IS NOT NULL")
        processed_count = cursor.fetchone()[0]

        logger.info("✅ Verification passed:")
        logger.info(f"   - Total memories: {total_count}")
        logger.info(f"   - AI processed: {processed_count}")
        logger.info("   - Schema simplified: ✓")

    def run(self, dry_run: bool = False) -> None:
        """Run the complete migration process"""
        logger.info(f"🚀 Starting Issue #112 schema migration (dry_run={dry_run})")

        # Create backup
        if not dry_run:
            self.create_backup()

        # Connect to database
        with sqlite3.connect(self.db_path) as conn:
            # Check current state
            schema_info = self.get_current_schema_info(conn)

            logger.info("📊 Current database state:")
            logger.info(f"   - Total memories: {schema_info['total_memories']}")
            logger.info(f"   - Has old schema: {schema_info['has_old_schema']}")
            logger.info(f"   - Has new schema: {schema_info['has_new_schema']}")
            logger.info(f"   - Migration needed: {schema_info['migration_needed']}")

            if not schema_info["migration_needed"]:
                logger.info("✅ No migration needed - schema is already up to date")
                return

            # Perform migration
            self.migrate_schema(conn, dry_run)

            # Verify results
            if not dry_run:
                self.verify_migration(conn)

        if self.backup_path:
            logger.info(f"💾 Backup saved to: {self.backup_path}")

        logger.info(
            "🎯 Migration completed successfully! Memory model is now AI-driven and simplified."
        )


def main():
    """Main CLI entry point"""
    parser = argparse.ArgumentParser(description="Migrate to simplified AI-driven memory schema")
    parser.add_argument(
        "--dry-run", action="store_true", help="Show what would be done without making changes"
    )
    parser.add_argument(
        "--no-backup", action="store_true", help="Skip creating backup (not recommended)"
    )
    parser.add_argument(
        "--db-path",
        default="data/memories.db",
        help="Path to SQLite database (default: data/memories.db)",
    )

    args = parser.parse_args()

    try:
        migration = SimplifiedSchemaMigration(db_path=args.db_path, backup=not args.no_backup)

        migration.run(dry_run=args.dry_run)

    except Exception as e:
        logger.error(f"❌ Migration failed: {e}")
        return 1

    return 0


if __name__ == "__main__":
    exit(main())
