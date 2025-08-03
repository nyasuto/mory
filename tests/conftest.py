"""Shared test configuration for pytest"""

import pytest
from fastapi.testclient import TestClient
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker
from sqlalchemy.pool import StaticPool

from app.core.database import Base, get_db
from app.main import app

# Test database setup
SQLALCHEMY_DATABASE_URL = "sqlite:///:memory:"

engine = create_engine(
    SQLALCHEMY_DATABASE_URL,
    poolclass=StaticPool,
    connect_args={"check_same_thread": False},
)
TestingSessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)


def override_get_db():
    """Override database dependency for testing"""
    try:
        db = TestingSessionLocal()
        yield db
    finally:
        db.close()


app.dependency_overrides[get_db] = override_get_db


@pytest.fixture(scope="function")
def db_session():
    """Create a fresh database for each test"""
    from sqlalchemy import text

    from app.core.database import create_tables

    # Clean up any existing FTS5 tables and triggers first
    try:
        with engine.connect() as conn:
            conn.execute(text("DROP TRIGGER IF EXISTS memories_fts_insert"))
            conn.execute(text("DROP TRIGGER IF EXISTS memories_fts_update"))
            conn.execute(text("DROP TRIGGER IF EXISTS memories_fts_delete"))
            conn.execute(text("DROP TABLE IF EXISTS memories_fts"))
            conn.commit()
    except Exception:
        pass

    Base.metadata.create_all(bind=engine)
    # Initialize FTS5 tables for testing
    try:
        create_tables(engine_override=engine)
    except Exception:
        pass  # FTS5 might not be available in test environment
    yield

    # Clean up after test
    try:
        with engine.connect() as conn:
            conn.execute(text("DROP TRIGGER IF EXISTS memories_fts_insert"))
            conn.execute(text("DROP TRIGGER IF EXISTS memories_fts_update"))
            conn.execute(text("DROP TRIGGER IF EXISTS memories_fts_delete"))
            conn.execute(text("DROP TABLE IF EXISTS memories_fts"))
            conn.commit()
    except Exception:
        pass
    Base.metadata.drop_all(bind=engine)


@pytest.fixture
def client():
    """Test client fixture"""
    return TestClient(app)
