"""Database configuration and session management
SQLite with SQLAlchemy for Mory Server
"""

from typing import Any

from sqlalchemy import create_engine, event, text
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker
from sqlalchemy.pool import StaticPool

from .config import settings

# SQLAlchemy setup
engine = create_engine(
    settings.sqlite_url,
    poolclass=StaticPool,
    connect_args={"check_same_thread": False, "timeout": 20},
    echo=settings.debug,
)


# Enable SQLite optimizations and FTS5
@event.listens_for(engine, "connect")
def set_sqlite_pragma(dbapi_connection: Any, connection_record: Any) -> None:
    """Set SQLite optimizations and enable FTS5"""
    cursor = dbapi_connection.cursor()

    # Performance optimizations
    cursor.execute("PRAGMA journal_mode=WAL")
    cursor.execute("PRAGMA synchronous=NORMAL")
    cursor.execute("PRAGMA cache_size=10000")
    cursor.execute("PRAGMA temp_store=memory")
    cursor.execute("PRAGMA mmap_size=268435456")

    # Enable foreign key constraints
    cursor.execute("PRAGMA foreign_keys=ON")

    cursor.close()


# Session factory
SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)

# Base class for all models
Base = declarative_base()


def get_db():
    """Database dependency for FastAPI"""
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()


def create_tables(engine_override: Any = None) -> None:
    """Create all database tables and FTS5 search tables"""
    db_engine = engine_override if engine_override else engine
    Base.metadata.create_all(bind=db_engine)

    # Initialize FTS5 search functionality if available
    if check_fts5_support(db_engine):
        create_fts5_table(db_engine)
        print("✅ FTS5 search enabled")
    else:
        print("⚠️  FTS5 not available, falling back to LIKE search")


def check_fts5_support(engine_override: Any = None) -> bool:
    """Check if SQLite FTS5 extension is available"""
    db_engine = engine_override if engine_override else engine
    try:
        with db_engine.connect() as conn:
            conn.execute(text("CREATE VIRTUAL TABLE IF NOT EXISTS fts_test USING fts5(content)"))
            conn.execute(text("DROP TABLE fts_test"))
            return True
    except Exception:
        return False


def create_fts5_table(engine_override: Any = None) -> bool:
    """Create FTS5 virtual table for full-text search"""
    db_engine = engine_override if engine_override else engine
    try:
        with db_engine.connect() as conn:
            # Create FTS5 virtual table with Japanese tokenizer support
            conn.execute(
                text("""
                CREATE VIRTUAL TABLE IF NOT EXISTS memories_fts USING fts5(
                    id UNINDEXED,
                    category,
                    key,
                    value,
                    tags,
                    content='memories',
                    tokenize='unicode61 remove_diacritics 2'
                )
            """)
            )

            # Create triggers for automatic synchronization
            conn.execute(
                text("""
                CREATE TRIGGER IF NOT EXISTS memories_fts_insert
                AFTER INSERT ON memories
                BEGIN
                    INSERT INTO memories_fts(id, category, key, value, tags)
                    VALUES (new.id, new.category, new.key, new.value, new.tags);
                END
            """)
            )

            conn.execute(
                text("""
                CREATE TRIGGER IF NOT EXISTS memories_fts_update
                AFTER UPDATE ON memories
                BEGIN
                    UPDATE memories_fts
                    SET category = new.category,
                        key = new.key,
                        value = new.value,
                        tags = new.tags
                    WHERE id = new.id;
                END
            """)
            )

            conn.execute(
                text("""
                CREATE TRIGGER IF NOT EXISTS memories_fts_delete
                AFTER DELETE ON memories
                BEGIN
                    DELETE FROM memories_fts WHERE id = old.id;
                END
            """)
            )

            conn.commit()
            return True
    except Exception as e:
        print(f"Failed to create FTS5 table: {e}")
        return False


def rebuild_fts5_index(engine_override: Any = None) -> bool:
    """Rebuild FTS5 index with all existing memories"""
    db_engine = engine_override if engine_override else engine
    try:
        with db_engine.connect() as conn:
            # Clear existing FTS5 data
            conn.execute(text("DELETE FROM memories_fts"))

            # Populate FTS5 table with existing data
            conn.execute(
                text("""
                INSERT INTO memories_fts(id, category, key, value, tags)
                SELECT id, category, key, value, tags FROM memories
            """)
            )

            conn.commit()
            return True
    except Exception as e:
        print(f"Failed to rebuild FTS5 index: {e}")
        return False
