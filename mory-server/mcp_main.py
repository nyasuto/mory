#!/usr/bin/env python3
"""
Standalone MCP server entry point for Mory
This script runs the MCP server for Claude Desktop integration
"""

import asyncio
import logging
import sys
from pathlib import Path

# Add app to Python path
sys.path.insert(0, str(Path(__file__).parent))

from mcp.server.stdio import stdio_server

from app.mcp_server import mcp_server

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
    handlers=[
        logging.FileHandler("mcp_server.log"),
        logging.StreamHandler(sys.stderr),
    ],
)

logger = logging.getLogger(__name__)


async def main():
    """Main entry point for MCP server"""
    logger.info("Starting Mory MCP Server...")

    try:
        # Run the server with stdio transport (required for Claude Desktop)
        async with stdio_server() as (read_stream, write_stream):
            await mcp_server.run(
                read_stream,
                write_stream,
                mcp_server.create_initialization_options(),
            )
    except Exception as e:
        logger.error(f"MCP Server failed: {e}")
        raise


if __name__ == "__main__":
    asyncio.run(main())
