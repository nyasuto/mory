"""Database configuration and session management
SQLite with SQLAlchemy for Mory Server
"""

from sqlalchemy import create_engine, event
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
def set_sqlite_pragma(dbapi_connection, connection_record):
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


def create_tables():
    """Create all database tables"""
    Base.metadata.create_all(bind=engine)


def check_fts5_support() -> bool:
    """Check if SQLite FTS5 extension is available"""
    try:
        with engine.connect() as conn:
            conn.execute("CREATE VIRTUAL TABLE IF NOT EXISTS fts_test USING fts5(content)")
            conn.execute("DROP TABLE fts_test")
            return True
    except Exception:
        return False
