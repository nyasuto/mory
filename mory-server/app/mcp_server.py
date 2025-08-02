"""
MCP (Model Context Protocol) server implementation for Mory
Provides memory management tools for Claude Desktop integration
"""

import json
import logging
from typing import Any

from mcp import types
from mcp.server import Server
from sqlalchemy.orm import Session

from .core.database import SessionLocal
from .models.memory import Memory
from .models.schemas import MemoryCreate, SearchRequest
from .services.search import search_service

# Initialize MCP server
mcp_server = Server("mory")
logger = logging.getLogger(__name__)


@mcp_server.list_tools()
async def handle_list_tools() -> list[types.Tool]:
    """List available MCP tools for memory management"""
    return [
        types.Tool(
            name="save_memory",
            description="Save or update a memory with optional categorization and tags",
            inputSchema={
                "type": "object",
                "properties": {
                    "category": {
                        "type": "string",
                        "description": "Memory category for organization",
                    },
                    "key": {
                        "type": "string",
                        "description": "Unique key for the memory (optional)",
                    },
                    "value": {
                        "type": "string",
                        "description": "The memory content/value to store",
                    },
                    "tags": {
                        "type": "array",
                        "items": {"type": "string"},
                        "description": "Tags for categorization and search",
                        "default": [],
                    },
                },
                "required": ["category", "value"],
            },
        ),
        types.Tool(
            name="get_memory",
            description="Retrieve a specific memory by key",
            inputSchema={
                "type": "object",
                "properties": {
                    "key": {
                        "type": "string",
                        "description": "The memory key to retrieve",
                    },
                    "category": {
                        "type": "string",
                        "description": "Filter by category (optional)",
                    },
                },
                "required": ["key"],
            },
        ),
        types.Tool(
            name="list_memories",
            description="List memories with optional filtering and pagination",
            inputSchema={
                "type": "object",
                "properties": {
                    "category": {
                        "type": "string",
                        "description": "Filter by category (optional)",
                    },
                    "limit": {
                        "type": "integer",
                        "description": "Maximum number of memories to return",
                        "default": 20,
                        "minimum": 1,
                        "maximum": 100,
                    },
                    "offset": {
                        "type": "integer",
                        "description": "Number of memories to skip",
                        "default": 0,
                        "minimum": 0,
                    },
                },
            },
        ),
        types.Tool(
            name="search_memories",
            description="Search memories using full-text search with optional semantic search",
            inputSchema={
                "type": "object",
                "properties": {
                    "query": {
                        "type": "string",
                        "description": "Search query text",
                    },
                    "category": {
                        "type": "string",
                        "description": "Filter results by category (optional)",
                    },
                    "tags": {
                        "type": "array",
                        "items": {"type": "string"},
                        "description": "Filter by tags (optional)",
                    },
                    "limit": {
                        "type": "integer",
                        "description": "Maximum number of results",
                        "default": 10,
                        "minimum": 1,
                        "maximum": 50,
                    },
                },
                "required": ["query"],
            },
        ),
    ]


@mcp_server.call_tool()
async def handle_call_tool(name: str, arguments: dict[str, Any]) -> list[types.TextContent]:
    """Execute MCP tool calls"""
    try:
        with SessionLocal() as db:
            if name == "save_memory":
                return await _save_memory(arguments, db)
            elif name == "get_memory":
                return await _get_memory(arguments, db)
            elif name == "list_memories":
                return await _list_memories(arguments, db)
            elif name == "search_memories":
                return await _search_memories(arguments, db)
            else:
                raise ValueError(f"Unknown tool: {name}")

    except Exception as e:
        logger.error(f"Tool {name} failed: {str(e)}")
        return [types.TextContent(type="text", text=f"Error: {str(e)}")]


async def _save_memory(arguments: dict[str, Any], db: Session) -> list[types.TextContent]:
    """Save or update a memory"""
    try:
        # Create memory data object
        memory_data = MemoryCreate(
            category=arguments["category"],
            key=arguments.get("key"),
            value=arguments["value"],
            tags=arguments.get("tags", []),
        )

        # Check if memory with this key already exists
        existing_memory = None
        if memory_data.key:
            existing_memory = (
                db.query(Memory)
                .filter(Memory.key == memory_data.key, Memory.category == memory_data.category)
                .first()
            )

        if existing_memory:
            # Update existing memory
            existing_memory.value = memory_data.value
            existing_memory.tags_list = memory_data.tags
            from datetime import datetime

            existing_memory.updated_at = datetime.utcnow()
            db.commit()
            db.refresh(existing_memory)

            result = {
                "action": "updated",
                "id": existing_memory.id,
                "category": existing_memory.category,
                "key": existing_memory.key,
                "value": existing_memory.value[:100] + "..."
                if len(existing_memory.value) > 100
                else existing_memory.value,
                "tags": existing_memory.tags_list,
                "updated_at": existing_memory.updated_at.isoformat(),
            }
        else:
            # Create new memory
            new_memory = Memory(
                category=memory_data.category,
                key=memory_data.key,
                value=memory_data.value,
                tags_list=memory_data.tags,
            )

            db.add(new_memory)
            db.commit()
            db.refresh(new_memory)

            result = {
                "action": "created",
                "id": new_memory.id,
                "category": new_memory.category,
                "key": new_memory.key,
                "value": new_memory.value[:100] + "..."
                if len(new_memory.value) > 100
                else new_memory.value,
                "tags": new_memory.tags_list,
                "created_at": new_memory.created_at.isoformat(),
            }

        return [types.TextContent(type="text", text=json.dumps(result, indent=2))]

    except Exception as e:
        raise ValueError(f"Failed to save memory: {str(e)}") from e


async def _get_memory(arguments: dict[str, Any], db: Session) -> list[types.TextContent]:
    """Retrieve a specific memory by key"""
    try:
        key = arguments["key"]
        category = arguments.get("category")

        query = db.query(Memory).filter(Memory.key == key)
        if category:
            query = query.filter(Memory.category == category)

        memory = query.first()

        if not memory:
            error_msg = f"Memory with key '{key}'"
            if category:
                error_msg += f" in category '{category}'"
            error_msg += " not found"
            raise ValueError(error_msg)

        result = {
            "id": memory.id,
            "category": memory.category,
            "key": memory.key,
            "value": memory.value,
            "tags": memory.tags_list,
            "created_at": memory.created_at.isoformat(),
            "updated_at": memory.updated_at.isoformat(),
        }

        return [types.TextContent(type="text", text=json.dumps(result, indent=2))]

    except Exception as e:
        raise ValueError(f"Failed to get memory: {str(e)}") from e


async def _list_memories(arguments: dict[str, Any], db: Session) -> list[types.TextContent]:
    """List memories with optional filtering"""
    try:
        category = arguments.get("category")
        limit = arguments.get("limit", 20)
        offset = arguments.get("offset", 0)

        query = db.query(Memory)

        if category:
            query = query.filter(Memory.category == category)

        # Get total count
        total = query.count()

        # Apply pagination and ordering
        memories = query.order_by(Memory.updated_at.desc()).offset(offset).limit(limit).all()

        result = {
            "memories": [
                {
                    "id": memory.id,
                    "category": memory.category,
                    "key": memory.key,
                    "value": memory.value[:100] + "..."
                    if len(memory.value) > 100
                    else memory.value,
                    "tags": memory.tags_list,
                    "created_at": memory.created_at.isoformat(),
                    "updated_at": memory.updated_at.isoformat(),
                }
                for memory in memories
            ],
            "total": total,
            "category": category,
            "limit": limit,
            "offset": offset,
        }

        return [types.TextContent(type="text", text=json.dumps(result, indent=2))]

    except Exception as e:
        raise ValueError(f"Failed to list memories: {str(e)}") from e


async def _search_memories(arguments: dict[str, Any], db: Session) -> list[types.TextContent]:
    """Search memories using full-text search"""
    try:
        search_request = SearchRequest(
            query=arguments["query"],
            category=arguments.get("category"),
            tags=arguments.get("tags", []),
            limit=arguments.get("limit", 10),
        )

        # Use the existing search service
        search_response = await search_service.search_memories(search_request, db)

        # Convert to JSON-serializable format
        result = {
            "query": search_response.query,
            "results": [
                {
                    "memory": {
                        "id": search_result.memory.id,
                        "category": search_result.memory.category,
                        "key": search_result.memory.key,
                        "value": search_result.memory.value,
                        "tags": search_result.memory.tags,
                        "created_at": search_result.memory.created_at.isoformat()
                        if search_result.memory.created_at
                        else None,
                        "updated_at": search_result.memory.updated_at.isoformat()
                        if search_result.memory.updated_at
                        else None,
                    },
                    "score": search_result.score,
                    "search_type": search_result.search_type,
                }
                for search_result in search_response.results
            ],
            "total": search_response.total,
            "search_type": search_response.search_type,
            "execution_time_ms": search_response.execution_time_ms,
        }

        return [types.TextContent(type="text", text=json.dumps(result, indent=2))]

    except Exception as e:
        raise ValueError(f"Failed to search memories: {str(e)}") from e


# Server configuration
async def start_mcp_server():
    """Start the MCP server"""
    logger.info("Starting Mory MCP Server...")

    # The server will be started by the MCP runtime
    # This function is here for any initialization we might need
    pass


# Export the server instance
__all__ = ["mcp_server", "start_mcp_server"]
