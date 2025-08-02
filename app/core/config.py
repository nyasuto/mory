"""Configuration management for Mory Server
Supports environment variables and .env files
"""

from pathlib import Path

from pydantic import Field
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    """Application settings with environment variable support"""

    # Server configuration
    host: str = Field(default="0.0.0.0", env="MORY_HOST")
    port: int = Field(default=8080, env="MORY_PORT")
    debug: bool = Field(default=False, env="MORY_DEBUG")

    # Database configuration
    data_dir: str = Field(default="data", env="MORY_DATA_DIR")
    database_url: str = Field(default="", env="MORY_DATABASE_URL")

    # OpenAI configuration (for semantic search)
    openai_api_key: str | None = Field(default=None, env="OPENAI_API_KEY")
    openai_model: str = Field(default="text-embedding-3-large", env="MORY_OPENAI_MODEL")

    # Obsidian integration
    obsidian_vault_path: str | None = Field(default=None, env="MORY_OBSIDIAN_VAULT_PATH")

    # Search configuration
    semantic_search_enabled: bool = Field(default=True, env="MORY_SEMANTIC_SEARCH_ENABLED")
    hybrid_search_weight: float = Field(default=0.7, env="MORY_HYBRID_SEARCH_WEIGHT")

    class Config:
        env_file = ".env"
        env_file_encoding = "utf-8"
        case_sensitive = False
        extra = "ignore"

    @property
    def sqlite_url(self) -> str:
        """Generate SQLite database URL"""
        if self.database_url:
            return self.database_url

        # Ensure data directory exists
        data_path = Path(self.data_dir)
        data_path.mkdir(parents=True, exist_ok=True)

        db_path = data_path / "memories.db"
        return f"sqlite:///{db_path}"

    @property
    def is_semantic_available(self) -> bool:
        """Check if semantic search is available"""
        return self.semantic_search_enabled and self.openai_api_key is not None


# Global settings instance
settings = Settings()
