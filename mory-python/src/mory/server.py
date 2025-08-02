"""MCP server implementation for Mory."""

import os
from typing import Any, Optional

from mcp.server import Server
from mcp.server.models import InitializationOptions
from mcp.server.stdio import stdio_server
from mcp.types import Tool

from .memory import Memory, SearchQuery
from .storage import JSONMemoryStore


class MoryServer:
    """Mory MCP Server implementation."""

    def __init__(self, data_dir: Optional[str] = None) -> None:
        """Initialize the Mory server.

        Args:
            data_dir: Directory to store data files.
                      Defaults to MORY_DATA_DIR env var or 'data'
        """
        data_dir_path = data_dir or os.getenv("MORY_DATA_DIR", "data")
        self.data_dir = data_dir_path
        self.store = JSONMemoryStore(self.data_dir)
        self.server = Server("mory-python")
        self._register_tools()

    def _register_tools(self) -> None:
        """Register MCP tools."""

        @self.server.list_tools()
        async def list_tools() -> list[Tool]:
            """List available tools."""
            return [
                Tool(
                    name="save_memory",
                    description="記憶を保存します",
                    inputSchema={
                        "type": "object",
                        "properties": {
                            "category": {
                                "type": "string",
                                "description": "記憶のカテゴリ",
                            },
                            "key": {
                                "type": "string",
                                "description": "記憶のキー（オプション）",
                            },
                            "value": {"type": "string", "description": "記憶する内容"},
                            "tags": {
                                "type": "array",
                                "items": {"type": "string"},
                                "description": "関連タグのリスト（オプション）",
                            },
                        },
                        "required": ["category", "value"],
                    },
                ),
                Tool(
                    name="get_memory",
                    description="記憶を取得します",
                    inputSchema={
                        "type": "object",
                        "properties": {
                            "key": {
                                "type": "string",
                                "description": "記憶のキーまたはID",
                            }
                        },
                        "required": ["key"],
                    },
                ),
                Tool(
                    name="list_memories",
                    description="記憶の一覧を取得します",
                    inputSchema={
                        "type": "object",
                        "properties": {
                            "category": {
                                "type": "string",
                                "description": "フィルタするカテゴリ（オプション）",
                            }
                        },
                    },
                ),
                Tool(
                    name="search_memories",
                    description="記憶を検索します",
                    inputSchema={
                        "type": "object",
                        "properties": {
                            "query": {"type": "string", "description": "検索クエリ"},
                            "category": {
                                "type": "string",
                                "description": "検索するカテゴリ（オプション）",
                            },
                            "limit": {
                                "type": "integer",
                                "description": "最大結果数（デフォルト: 20）",
                                "minimum": 1,
                                "maximum": 100,
                            },
                            "min_score": {
                                "type": "number",
                                "description": "最小関連度スコア（デフォルト: 0.0）",
                                "minimum": 0.0,
                                "maximum": 1.0,
                            },
                        },
                        "required": ["query"],
                    },
                ),
                Tool(
                    name="delete_memory",
                    description="記憶を削除します",
                    inputSchema={
                        "type": "object",
                        "properties": {
                            "key": {
                                "type": "string",
                                "description": "削除する記憶のキーまたはID",
                            }
                        },
                        "required": ["key"],
                    },
                ),
            ]

        @self.server.call_tool()
        async def call_tool(
            name: str, arguments: dict[str, Any]
        ) -> list[dict[str, Any]]:
            """Handle tool calls."""
            try:
                if name == "save_memory":
                    return await self._save_memory(**arguments)
                elif name == "get_memory":
                    return await self._get_memory(**arguments)
                elif name == "list_memories":
                    return await self._list_memories(**arguments)
                elif name == "search_memories":
                    return await self._search_memories(**arguments)
                elif name == "delete_memory":
                    return await self._delete_memory(**arguments)
                else:
                    return [{"type": "text", "text": f"Unknown tool: {name}"}]
            except Exception as e:
                return [{"type": "text", "text": f"Error: {str(e)}"}]

    async def _save_memory(
        self, category: str, value: str, key: str = "", tags: Optional[list[str]] = None
    ) -> list[dict[str, Any]]:
        """Save a memory."""
        memory = Memory(category=category, key=key, value=value, tags=tags or [])

        memory_id = await self.store.save(memory)

        return [{"type": "text", "text": f"記憶を保存しました。ID: {memory_id}"}]

    async def _get_memory(self, key: str) -> list[dict[str, Any]]:
        """Get a memory by key or ID."""
        memory = await self.store.get(key)

        memory_info = (
            f"カテゴリ: {memory.category}\n"
            f"キー: {memory.key}\n"
            f"内容: {memory.value}\n"
            f"タグ: {', '.join(memory.tags)}\n"
            f"作成日時: {memory.created_at}\n"
            f"更新日時: {memory.updated_at}"
        )

        return [{"type": "text", "text": memory_info}]

    async def _list_memories(
        self, category: Optional[str] = None
    ) -> list[dict[str, Any]]:
        """List memories."""
        memories = await self.store.list(category)

        if not memories:
            return [{"type": "text", "text": "記憶がありません。"}]

        result = f"記憶一覧（{len(memories)}件）:\n\n"
        for memory in memories:
            result += f"ID: {memory.id}\n"
            result += f"カテゴリ: {memory.category}\n"
            if memory.key:
                result += f"キー: {memory.key}\n"
            content_preview = memory.value[:100]
            if len(memory.value) > 100:
                content_preview += "..."
            result += f"内容: {content_preview}\n"
            if memory.tags:
                result += f"タグ: {', '.join(memory.tags)}\n"
            result += f"作成日時: {memory.created_at}\n"
            result += "---\n"

        return [{"type": "text", "text": result}]

    async def _search_memories(
        self,
        query: str,
        category: Optional[str] = None,
        limit: int = 20,
        min_score: float = 0.0,
    ) -> list[dict[str, Any]]:
        """Search memories."""
        search_query = SearchQuery(
            query=query, category=category, limit=limit, min_score=min_score
        )

        results = await self.store.search(search_query)

        if not results:
            return [
                {
                    "type": "text",
                    "text": f"'{query}'に一致する記憶が見つかりませんでした。",
                }
            ]

        result_text = f"検索結果（{len(results)}件）:\n\n"
        for result in results:
            memory = result.memory
            result_text += f"関連度: {result.score:.2f}\n"
            result_text += f"ID: {memory.id}\n"
            result_text += f"カテゴリ: {memory.category}\n"
            if memory.key:
                result_text += f"キー: {memory.key}\n"
            result_text += f"内容: {memory.value}\n"
            if memory.tags:
                result_text += f"タグ: {', '.join(memory.tags)}\n"
            result_text += f"作成日時: {memory.created_at}\n"
            result_text += "---\n"

        return [{"type": "text", "text": result_text}]

    async def _delete_memory(self, key: str) -> list[dict[str, Any]]:
        """Delete a memory."""
        await self.store.delete(key)

        return [{"type": "text", "text": f"記憶を削除しました: {key}"}]

    async def run(self) -> None:
        """Run the MCP server."""
        # Load existing data
        await self.store.load()

        # Start the server
        async with stdio_server() as (read_stream, write_stream):
            await self.server.run(
                read_stream,
                write_stream,
                InitializationOptions(
                    server_name="mory-python",
                    server_version="0.1.0",
                    capabilities=self.server.get_capabilities(
                        notification_options=None, experimental_capabilities={}
                    ),
                ),
            )
