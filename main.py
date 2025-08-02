#!/usr/bin/env python3
"""Main entry point for Mory MCP server."""

import asyncio
from src.mory.main import main

if __name__ == "__main__":
    asyncio.run(main())