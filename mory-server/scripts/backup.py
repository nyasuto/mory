#!/usr/bin/env python3
"""
Backup utility for Mory Server
Provides comprehensive backup and restore functionality for database and configuration
"""

import argparse
import json
import shutil
import sqlite3
import sys
import tarfile
from datetime import datetime
from pathlib import Path


class MoryBackup:
    """Handles backup and restore operations for Mory Server"""

    def __init__(self, data_dir: str, backup_dir: str):
        """Initialize backup manager"""
        self.data_dir = Path(data_dir)
        self.backup_dir = Path(backup_dir)
        self.backup_dir.mkdir(parents=True, exist_ok=True)

    def create_backup(self, name: str | None = None) -> Path:
        """Create a complete backup of Mory data"""
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        backup_name = name or f"mory_backup_{timestamp}"
        backup_file = self.backup_dir / f"{backup_name}.tar.gz"

        print(f"🔄 Creating backup: {backup_file.name}")

        # Verify database integrity before backup
        db_path = self.data_dir / "mory.db"
        if db_path.exists():
            if not self._verify_database_integrity(db_path):
                print("⚠️  Database integrity check failed - proceeding with backup anyway")

        # Create backup metadata
        metadata = self._create_metadata()

        # Create temporary directory for backup preparation
        temp_dir = self.backup_dir / f"temp_{timestamp}"
        temp_dir.mkdir(exist_ok=True)

        try:
            # Copy data files
            if self.data_dir.exists():
                shutil.copytree(self.data_dir, temp_dir / "data", dirs_exist_ok=True)

            # Save metadata
            with open(temp_dir / "backup_metadata.json", "w") as f:
                json.dump(metadata, f, indent=2, default=str)

            # Create compressed archive
            with tarfile.open(backup_file, "w:gz") as tar:
                tar.add(temp_dir, arcname=".", recursive=True)

            print(f"✅ Backup created successfully: {backup_file}")
            print(f"   Size: {self._format_size(backup_file.stat().st_size)}")
            return backup_file

        finally:
            # Clean up temporary directory
            if temp_dir.exists():
                shutil.rmtree(temp_dir)

    def restore_backup(self, backup_file: str, force: bool = False) -> bool:
        """Restore from backup file"""
        backup_path = Path(backup_file)

        if not backup_path.exists():
            print(f"❌ Backup file not found: {backup_file}")
            return False

        print(f"🔄 Restoring from backup: {backup_path.name}")

        # Check if data directory exists and warn user
        if self.data_dir.exists() and any(self.data_dir.iterdir()) and not force:
            print("⚠️  Data directory is not empty!")
            response = input("This will overwrite existing data. Continue? (y/N): ")
            if response.lower() != "y":
                print("❌ Restore cancelled")
                return False

        # Create temporary extraction directory
        temp_dir = self.backup_dir / f"restore_temp_{datetime.now().strftime('%Y%m%d_%H%M%S')}"
        temp_dir.mkdir(exist_ok=True)

        try:
            # Extract backup
            with tarfile.open(backup_path, "r:gz") as tar:
                tar.extractall(temp_dir)

            # Verify backup metadata
            metadata_file = temp_dir / "backup_metadata.json"
            if metadata_file.exists():
                with open(metadata_file) as f:
                    metadata = json.load(f)
                print(f"📋 Backup created: {metadata.get('timestamp')}")
                print(f"   Version: {metadata.get('version', 'unknown')}")
                print(f"   Records: {metadata.get('record_count', 'unknown')}")

            # Backup existing data if it exists
            if self.data_dir.exists():
                backup_existing = self.data_dir.with_suffix(
                    f".backup_{datetime.now().strftime('%Y%m%d_%H%M%S')}"
                )
                shutil.move(self.data_dir, backup_existing)
                print(f"📁 Existing data backed up to: {backup_existing}")

            # Restore data
            restore_data_dir = temp_dir / "data"
            if restore_data_dir.exists():
                shutil.move(restore_data_dir, self.data_dir)
                print("✅ Data restored successfully")
            else:
                print("⚠️  No data directory found in backup")

            # Verify restored database
            db_path = self.data_dir / "mory.db"
            if db_path.exists():
                if self._verify_database_integrity(db_path):
                    print("✅ Database integrity verified")
                else:
                    print("⚠️  Database integrity check failed after restore")

            return True

        except Exception as e:
            print(f"❌ Restore failed: {e}")
            return False

        finally:
            # Clean up temporary directory
            if temp_dir.exists():
                shutil.rmtree(temp_dir)

    def list_backups(self) -> list[dict]:
        """List available backups with metadata"""
        backups = []

        for backup_file in self.backup_dir.glob("*.tar.gz"):
            try:
                # Extract metadata from backup
                with tarfile.open(backup_file, "r:gz") as tar:
                    metadata_member = None
                    for member in tar.getmembers():
                        if member.name.endswith("backup_metadata.json"):
                            metadata_member = member
                            break

                    if metadata_member:
                        metadata_file = tar.extractfile(metadata_member)
                        if metadata_file:
                            metadata = json.load(metadata_file)
                        else:
                            metadata = {}
                    else:
                        metadata = {}

                backup_info = {
                    "file": backup_file.name,
                    "path": str(backup_file),
                    "size": backup_file.stat().st_size,
                    "size_formatted": self._format_size(backup_file.stat().st_size),
                    "created": datetime.fromtimestamp(backup_file.stat().st_mtime),
                    "metadata": metadata,
                }
                backups.append(backup_info)

            except Exception as e:
                print(f"⚠️  Could not read backup {backup_file.name}: {e}")

        # Sort by creation time (newest first)
        backups.sort(key=lambda x: x["created"], reverse=True)
        return backups

    def cleanup_old_backups(self, keep_days: int = 30) -> int:
        """Remove backups older than specified days"""
        cutoff_time = datetime.now().timestamp() - (keep_days * 24 * 60 * 60)
        removed_count = 0

        for backup_file in self.backup_dir.glob("*.tar.gz"):
            if backup_file.stat().st_mtime < cutoff_time:
                try:
                    backup_file.unlink()
                    print(f"🗑️  Removed old backup: {backup_file.name}")
                    removed_count += 1
                except Exception as e:
                    print(f"⚠️  Could not remove {backup_file.name}: {e}")

        return removed_count

    def _create_metadata(self) -> dict:
        """Create backup metadata"""
        metadata = {
            "timestamp": datetime.now().isoformat(),
            "version": "1.0.0-alpha",  # Should match pyproject.toml version
            "backup_type": "full",
        }

        # Add database statistics
        db_path = self.data_dir / "mory.db"
        if db_path.exists():
            try:
                conn = sqlite3.connect(db_path)
                cursor = conn.cursor()

                # Get memory count
                cursor.execute("SELECT COUNT(*) FROM memories")
                metadata["record_count"] = cursor.fetchone()[0]

                # Get database file size
                metadata["database_size"] = db_path.stat().st_size

                # Get table info
                cursor.execute("SELECT name FROM sqlite_master WHERE type='table'")
                metadata["tables"] = [row[0] for row in cursor.fetchall()]

                conn.close()
            except Exception as e:
                metadata["database_error"] = str(e)

        return metadata

    def _verify_database_integrity(self, db_path: Path) -> bool:
        """Verify SQLite database integrity"""
        try:
            conn = sqlite3.connect(db_path)
            cursor = conn.cursor()
            cursor.execute("PRAGMA integrity_check")
            result = cursor.fetchone()[0]
            conn.close()
            return result == "ok"
        except Exception:
            return False

    def _format_size(self, size: int) -> str:
        """Format file size in human readable format"""
        for unit in ["B", "KB", "MB", "GB"]:
            if size < 1024.0:
                return f"{size:.1f} {unit}"
            size /= 1024.0
        return f"{size:.1f} TB"


def main():
    """Main backup utility function"""
    parser = argparse.ArgumentParser(description="Mory Server Backup Utility")
    parser.add_argument(
        "--data-dir", default="/opt/mory-server/data", help="Path to Mory data directory"
    )
    parser.add_argument(
        "--backup-dir", default="/opt/mory-server/backups", help="Path to backup directory"
    )

    subparsers = parser.add_subparsers(dest="command", help="Available commands")

    # Create backup
    create_parser = subparsers.add_parser("create", help="Create backup")
    create_parser.add_argument("--name", help="Custom backup name")

    # Restore backup
    restore_parser = subparsers.add_parser("restore", help="Restore from backup")
    restore_parser.add_argument("backup_file", help="Path to backup file")
    restore_parser.add_argument(
        "--force", action="store_true", help="Force restore without confirmation"
    )

    # List backups
    subparsers.add_parser("list", help="List available backups")

    # Cleanup old backups
    cleanup_parser = subparsers.add_parser("cleanup", help="Remove old backups")
    cleanup_parser.add_argument(
        "--keep-days", type=int, default=30, help="Keep backups newer than N days (default: 30)"
    )

    args = parser.parse_args()

    if not args.command:
        parser.print_help()
        return 1

    # Initialize backup manager
    backup_manager = MoryBackup(args.data_dir, args.backup_dir)

    try:
        if args.command == "create":
            backup_file = backup_manager.create_backup(args.name)
            print(f"\n🎉 Backup completed: {backup_file}")

        elif args.command == "restore":
            success = backup_manager.restore_backup(args.backup_file, args.force)
            if success:
                print("\n🎉 Restore completed successfully")
                return 0
            else:
                return 1

        elif args.command == "list":
            backups = backup_manager.list_backups()
            if not backups:
                print("📭 No backups found")
                return 0

            print(f"\n📋 Found {len(backups)} backup(s):")
            print("-" * 80)
            for backup in backups:
                print(f"📁 {backup['file']}")
                print(f"   Created: {backup['created'].strftime('%Y-%m-%d %H:%M:%S')}")
                print(f"   Size: {backup['size_formatted']}")
                if "record_count" in backup["metadata"]:
                    print(f"   Records: {backup['metadata']['record_count']}")
                print()

        elif args.command == "cleanup":
            removed = backup_manager.cleanup_old_backups(args.keep_days)
            print(f"\n🧹 Removed {removed} old backup(s)")

        return 0

    except Exception as e:
        print(f"\n❌ Operation failed: {e}")
        return 1


if __name__ == "__main__":
    sys.exit(main())
