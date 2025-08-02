"""Tests for health check endpoints"""

from fastapi.testclient import TestClient

from app.main import app

client = TestClient(app)


def test_root_endpoint():
    """Test root endpoint returns basic information"""
    response = client.get("/")
    assert response.status_code == 200

    data = response.json()
    assert data["service"] == "Mory Server"
    assert data["version"] == "1.0.0-alpha"
    assert "documentation" in data
    assert "health" in data


def test_health_check():
    """Test basic health check endpoint"""
    response = client.get("/api/health")
    assert response.status_code == 200

    data = response.json()
    assert data["status"] == "healthy"
    assert data["service"] == "mory-server"
    assert data["version"] == "1.0.0-alpha"
    assert "timestamp" in data


def test_detailed_health_check():
    """Test detailed health check endpoint"""
    response = client.get("/api/health/detailed")
    assert response.status_code == 200

    data = response.json()
    assert data["service"] == "mory-server"
    assert "components" in data
    assert "configuration" in data

    # Check components
    components = data["components"]
    assert "database" in components
    assert "semantic_search" in components
    assert "obsidian" in components

    # Check database component
    database = components["database"]
    assert database["type"] == "sqlite"
    assert "status" in database
    assert "fts5_support" in database

    # Check configuration
    config = data["configuration"]
    assert "host" in config
    assert "port" in config
    assert "debug" in config
