"""
MCP (Model Context Protocol) server implementation for Mory
Provides memory management tools for Claude Desktop integration via HTTP API
"""

import json
import logging
import os
from typing import Any

import httpx
from mcp import types
from mcp.server import Server

# Initialize MCP server
mcp_server = Server("mory")
logger = logging.getLogger(__name__)

# API base URL from environment
API_BASE_URL = os.getenv("MORY_API_URL", "http://localhost:8080")


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
    """Execute MCP tool calls via HTTP API"""
    try:
        async with httpx.AsyncClient() as client:
            if name == "save_memory":
                return await _save_memory(arguments, client)
            elif name == "get_memory":
                return await _get_memory(arguments, client)
            elif name == "list_memories":
                return await _list_memories(arguments, client)
            elif name == "search_memories":
                return await _search_memories(arguments, client)
            else:
                raise ValueError(f"Unknown tool: {name}")

    except Exception as e:
        logger.error(f"Tool {name} failed: {str(e)}")
        return [types.TextContent(type="text", text=f"Error: {str(e)}")]


async def _save_memory(
    arguments: dict[str, Any], client: httpx.AsyncClient
) -> list[types.TextContent]:
    """Save or update a memory via HTTP API"""
    try:
        # Prepare request data
        memory_data = {
            "category": arguments["category"],
            "key": arguments.get("key"),
            "value": arguments["value"],
            "tags": arguments.get("tags", []),
        }

        # Make HTTP request to FastAPI server
        response = await client.post(
            f"{API_BASE_URL}/api/memories",
            json=memory_data,
            headers={"Content-Type": "application/json"},
        )
        response.raise_for_status()

        result = response.json()
        return [types.TextContent(type="text", text=json.dumps(result, indent=2))]

    except httpx.HTTPStatusError as e:
        error_detail = e.response.text if e.response else str(e)
        raise ValueError(f"HTTP {e.response.status_code}: {error_detail}") from e
    except Exception as e:
        raise ValueError(f"Failed to save memory: {str(e)}") from e


async def _get_memory(
    arguments: dict[str, Any], client: httpx.AsyncClient
) -> list[types.TextContent]:
    """Retrieve a specific memory by key via HTTP API"""
    try:
        key = arguments["key"]
        category = arguments.get("category")

        # Build query parameters
        params = {}
        if category:
            params["category"] = category

        # Make HTTP request
        response = await client.get(f"{API_BASE_URL}/api/memories/{key}", params=params)
        response.raise_for_status()

        result = response.json()
        return [types.TextContent(type="text", text=json.dumps(result, indent=2))]

    except httpx.HTTPStatusError as e:
        if e.response.status_code == 404:
            error_msg = f"Memory with key '{key}'"
            if arguments.get("category"):
                error_msg += f" in category '{arguments['category']}'"
            error_msg += " not found"
            raise ValueError(error_msg) from e
        else:
            error_detail = e.response.text if e.response else str(e)
            raise ValueError(f"HTTP {e.response.status_code}: {error_detail}") from e
    except Exception as e:
        raise ValueError(f"Failed to get memory: {str(e)}") from e


async def _list_memories(
    arguments: dict[str, Any], client: httpx.AsyncClient
) -> list[types.TextContent]:
    """List memories with optional filtering via HTTP API"""
    try:
        # Build query parameters
        params = {}
        if arguments.get("category"):
            params["category"] = arguments["category"]
        if arguments.get("limit"):
            params["limit"] = arguments["limit"]
        if arguments.get("offset"):
            params["offset"] = arguments["offset"]

        # Make HTTP request
        response = await client.get(f"{API_BASE_URL}/api/memories", params=params)
        response.raise_for_status()

        result = response.json()
        return [types.TextContent(type="text", text=json.dumps(result, indent=2))]

    except httpx.HTTPStatusError as e:
        error_detail = e.response.text if e.response else str(e)
        raise ValueError(f"HTTP {e.response.status_code}: {error_detail}") from e
    except Exception as e:
        raise ValueError(f"Failed to list memories: {str(e)}") from e


async def _search_memories(
    arguments: dict[str, Any], client: httpx.AsyncClient
) -> list[types.TextContent]:
    """Search memories using full-text search via HTTP API"""
    try:
        # Prepare search request data
        search_data = {
            "query": arguments["query"],
            "category": arguments.get("category"),
            "tags": arguments.get("tags", []),
            "limit": arguments.get("limit", 10),
        }

        # Make HTTP request
        response = await client.post(
            f"{API_BASE_URL}/api/memories/search",
            json=search_data,
            headers={"Content-Type": "application/json"},
        )
        response.raise_for_status()

        result = response.json()
        return [types.TextContent(type="text", text=json.dumps(result, indent=2))]

    except httpx.HTTPStatusError as e:
        error_detail = e.response.text if e.response else str(e)
        raise ValueError(f"HTTP {e.response.status_code}: {error_detail}") from e
    except Exception as e:
        raise ValueError(f"Failed to search memories: {str(e)}") from e


# Server configuration
async def start_mcp_server():
    """Start the MCP server"""
    logger.info("Starting Mory MCP Server...")
    logger.info(f"API Base URL: {API_BASE_URL}")

    # The server will be started by the MCP runtime
    # This function is here for any initialization we might need
    pass


# Export the server instance
__all__ = ["mcp_server", "start_mcp_server"]
