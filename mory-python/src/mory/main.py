"""Main entry point for Mory MCP server."""

import asyncio
import sys

from .server import MoryServer


def main() -> None:
    """Main entry point."""
    try:
        server = MoryServer()
        asyncio.run(server.run())
    except KeyboardInterrupt:
        print("Server stopped by user", file=sys.stderr)
    except Exception as e:
        print(f"Server error: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
