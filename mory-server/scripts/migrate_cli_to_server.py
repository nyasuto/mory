#!/usr/bin/env python3
"""
Data migration script for Mory: CLI version to FastAPI server version

This script migrates data from the existing CLI-based Mory database
to the new FastAPI server database format.
"""

import argparse
import json
import sqlite3
import sys
from datetime import UTC, datetime
from pathlib import Path
from typing import Any

from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker

# Add app to path
sys.path.append(str(Path(__file__).parent.parent))

from app.core.database import Base, check_fts5_support, create_fts5_table
from app.models.memory import Memory


class DataMigrator:
    """Handles migration from CLI database to server database"""

    def __init__(self, old_db_path: str, new_db_path: str, dry_run: bool = False):
        """Initialize migrator with database paths"""
        self.old_db_path = Path(old_db_path)
        self.new_db_path = Path(new_db_path)
        self.dry_run = dry_run

        # Validate old database exists
        if not self.old_db_path.exists():
            raise FileNotFoundError(f"Old database not found: {old_db_path}")

        # Initialize connections
        self.old_db = sqlite3.connect(str(self.old_db_path))
        self.old_db.row_factory = sqlite3.Row  # Enable dict-like access

        # Create new database
        self.new_db_path.parent.mkdir(parents=True, exist_ok=True)
        self.new_engine = create_engine(f"sqlite:///{self.new_db_path}")

        # Create tables in new database
        Base.metadata.create_all(self.new_engine)

        # Initialize FTS5 if available
        if check_fts5_support():
            create_fts5_table()
            print("✅ FTS5 search tables created")

        self.SessionLocal = sessionmaker(bind=self.new_engine)

        # Migration statistics
        self.stats = {
            "memories_processed": 0,
            "memories_migrated": 0,
            "embeddings_processed": 0,
            "embeddings_migrated": 0,
            "errors": [],
        }

    def analyze_old_database(self) -> dict[str, Any]:
        """Analyze the structure and content of the old database"""
        print("🔍 Analyzing old database structure...")

        cursor = self.old_db.cursor()

        # Get table names
        cursor.execute("SELECT name FROM sqlite_master WHERE type='table'")
        tables = [row[0] for row in cursor.fetchall()]

        analysis = {"tables": {}}

        for table in tables:
            # Get table schema
            cursor.execute(f"PRAGMA table_info({table})")
            columns = cursor.fetchall()

            # Get row count
            cursor.execute(f"SELECT COUNT(*) FROM {table}")
            row_count = cursor.fetchone()[0]

            analysis["tables"][table] = {
                "columns": [
                    {"name": col[1], "type": col[2], "nullable": not col[3]} for col in columns
                ],
                "row_count": row_count,
            }

            print(f"  📊 Table '{table}': {row_count} rows")
            for col in columns:
                print(f"    - {col[1]} ({col[2]})")

        return analysis

    def migrate_memories(self) -> bool:
        """Migrate memory records from old to new database"""
        print("\n📝 Migrating memory records...")

        cursor = self.old_db.cursor()

        # Check if memories table exists
        cursor.execute("SELECT name FROM sqlite_master WHERE type='table' AND name='memories'")
        if not cursor.fetchone():
            print("⚠️  No 'memories' table found in old database")
            return True

        # Get all memories from old database
        cursor.execute("SELECT * FROM memories ORDER BY created_at")
        old_memories = cursor.fetchall()

        print(f"Found {len(old_memories)} memories to migrate")

        if self.dry_run:
            print("🔄 DRY RUN: Would migrate the following memories:")
            for memory in old_memories[:5]:  # Show first 5
                key = memory["key"] if memory["key"] else "no-key"
                print(f"  - {memory['category']}/{key}: {memory['value'][:50]}...")
            if len(old_memories) > 5:
                print(f"  ... and {len(old_memories) - 5} more")
            return True

        session = self.SessionLocal()

        try:
            for old_memory in old_memories:
                self.stats["memories_processed"] += 1

                try:
                    # Parse dates (CLIデータベースではUNIXタイムスタンプ)
                    created_at = self._parse_datetime(old_memory["created_at"])
                    updated_at = self._parse_datetime(old_memory["updated_at"])

                    # Parse tags
                    tags = self._parse_tags(old_memory["tags"] or "[]")

                    # Create new memory record
                    new_memory = Memory(
                        id=old_memory["id"] or f"mem_{self._generate_id()}",
                        category=old_memory["category"],
                        key=old_memory["key"],
                        value=old_memory["value"],
                        tags=json.dumps(tags),
                        created_at=created_at,
                        updated_at=updated_at,
                        embedding=old_memory["embedding"],  # Binary data
                        embedding_hash=old_memory["embedding_hash"],
                    )

                    session.add(new_memory)
                    self.stats["memories_migrated"] += 1

                    if self.stats["memories_processed"] % 100 == 0:
                        print(f"  📝 Processed {self.stats['memories_processed']} memories...")

                except Exception as e:
                    error_msg = f"Error migrating memory {old_memory['id'] if old_memory['id'] else 'unknown'}: {e}"
                    self.stats["errors"].append(error_msg)
                    print(f"  ❌ {error_msg}")

            session.commit()
            print(f"✅ Successfully migrated {self.stats['memories_migrated']} memories")
            return True

        except Exception as e:
            session.rollback()
            print(f"❌ Failed to migrate memories: {e}")
            return False
        finally:
            session.close()

    def migrate_embeddings(self) -> bool:
        """Migrate embedding data and regenerate if necessary"""
        print("\n🧠 Processing embeddings...")

        session = self.SessionLocal()

        try:
            # Count memories with embeddings
            memories_with_embeddings = (
                session.query(Memory).filter(Memory.embedding.isnot(None)).count()
            )

            total_memories = session.query(Memory).count()

            print(f"Found {memories_with_embeddings}/{total_memories} memories with embeddings")

            if memories_with_embeddings == 0:
                print("⚠️  No embeddings found. You may want to regenerate them using OpenAI API.")
                print(
                    "   Use the server's semantic search functionality to generate embeddings for new content."
                )

            self.stats["embeddings_processed"] = total_memories
            self.stats["embeddings_migrated"] = memories_with_embeddings

            return True

        except Exception as e:
            print(f"❌ Failed to process embeddings: {e}")
            return False
        finally:
            session.close()

    def verify_migration(self) -> bool:
        """Verify the migration was successful"""
        print("\n✅ Verifying migration...")

        session = self.SessionLocal()

        try:
            # Count records
            new_memory_count = session.query(Memory).count()

            cursor = self.old_db.cursor()
            cursor.execute("SELECT COUNT(*) FROM memories")
            old_memory_count = cursor.fetchone()[0]

            print(f"Old database: {old_memory_count} memories")
            print(f"New database: {new_memory_count} memories")

            if new_memory_count != old_memory_count:
                print(
                    f"⚠️  Memory count mismatch! Expected {old_memory_count}, got {new_memory_count}"
                )
                return False

            # Verify some random samples
            sample_memories = session.query(Memory).limit(5).all()
            print("\n📋 Sample migrated records:")
            for memory in sample_memories:
                print(f"  - {memory.category}/{memory.key or 'no-key'}: {len(memory.value)} chars")
                print(f"    Tags: {len(memory.tags_list)} | Created: {memory.created_at}")

            print("✅ Migration verification passed!")
            return True

        except Exception as e:
            print(f"❌ Verification failed: {e}")
            return False
        finally:
            session.close()

    def print_migration_summary(self):
        """Print detailed migration summary"""
        print("\n" + "=" * 60)
        print("📊 MIGRATION SUMMARY")
        print("=" * 60)
        print(f"Memories processed: {self.stats['memories_processed']}")
        print(f"Memories migrated:  {self.stats['memories_migrated']}")
        print(f"Embeddings found:   {self.stats['embeddings_migrated']}")
        print(f"Errors encountered: {len(self.stats['errors'])}")

        if self.stats["errors"]:
            print("\n❌ Errors:")
            for error in self.stats["errors"][:10]:  # Show first 10 errors
                print(f"  - {error}")
            if len(self.stats["errors"]) > 10:
                print(f"  ... and {len(self.stats['errors']) - 10} more errors")

        if self.dry_run:
            print("\n🔄 This was a DRY RUN - no actual migration performed")
        else:
            print(f"\n📂 New database created: {self.new_db_path}")
        print("=" * 60)

    def _parse_datetime(self, date_str: str | int | None) -> datetime:
        """Parse datetime string from various formats"""
        if not date_str:
            return datetime.now(UTC)

        # Handle Unix timestamp (integer)
        if isinstance(date_str, int):
            try:
                return datetime.fromtimestamp(date_str / 1000)  # Assuming milliseconds
            except (ValueError, OSError):
                return datetime.now(UTC)

        # Handle string formats
        try:
            # Try ISO format first
            return datetime.fromisoformat(str(date_str).replace("Z", "+00:00"))
        except (ValueError, AttributeError):
            try:
                # Try alternative formats
                return datetime.strptime(str(date_str), "%Y-%m-%d %H:%M:%S")
            except ValueError:
                # Fallback to current time
                return datetime.now(UTC)

    def _parse_tags(self, tags_str: str) -> list[str]:
        """Parse tags from JSON string"""
        if not tags_str:
            return []

        try:
            parsed = json.loads(tags_str)
            return parsed if isinstance(parsed, list) else []
        except (json.JSONDecodeError, TypeError):
            # Try to parse as comma-separated string
            if isinstance(tags_str, str) and tags_str.strip():
                return [tag.strip() for tag in tags_str.split(",") if tag.strip()]
            return []

    def _generate_id(self) -> str:
        """Generate a unique ID for memories without one"""
        import uuid

        return uuid.uuid4().hex[:8]

    def close(self):
        """Close database connections"""
        if hasattr(self, "old_db"):
            self.old_db.close()


def main():
    """Main migration function"""
    parser = argparse.ArgumentParser(description="Migrate Mory CLI database to server database")
    parser.add_argument("old_db", help="Path to old CLI database file")
    parser.add_argument("new_db", help="Path for new server database file")
    parser.add_argument("--dry-run", action="store_true", help="Analyze only, don't migrate")
    parser.add_argument("--backup", action="store_true", help="Create backup of old database")
    parser.add_argument("--force", action="store_true", help="Skip confirmation prompt")

    args = parser.parse_args()

    print("🚀 Mory CLI to Server Migration Tool")
    print("=" * 50)

    try:
        # Create backup if requested
        if args.backup:
            backup_path = f"{args.old_db}.backup.{datetime.now().strftime('%Y%m%d_%H%M%S')}"
            import shutil

            shutil.copy2(args.old_db, backup_path)
            print(f"📁 Backup created: {backup_path}")

        # Initialize migrator
        migrator = DataMigrator(args.old_db, args.new_db, dry_run=args.dry_run)

        # Analyze old database
        migrator.analyze_old_database()

        if not args.dry_run and not args.force:
            response = input("\n❓ Proceed with migration? (y/N): ")
            if response.lower() != "y":
                print("❌ Migration cancelled")
                return 1

        # Perform migration
        success = True
        success &= migrator.migrate_memories()
        success &= migrator.migrate_embeddings()

        if not args.dry_run and success:
            success &= migrator.verify_migration()

        # Print summary
        migrator.print_migration_summary()

        if success:
            print("\n🎉 Migration completed successfully!")
            if not args.dry_run:
                print(f"👉 New database ready at: {args.new_db}")
            return 0
        else:
            print("\n❌ Migration completed with errors")
            return 1

    except Exception as e:
        print(f"\n💥 Migration failed: {e}")
        return 1
    finally:
        if "migrator" in locals():
            migrator.close()


if __name__ == "__main__":
    sys.exit(main())
