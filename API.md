# 🔧 Mory Python版 APIリファレンス・技術文書

この文書では、Mory Python実装MCPサーバーの詳細な技術仕様を提供します。

## 📋 目次

- [データモデル](#データモデル)
- [MCPツールリファレンス](#mcpツールリファレンス)
- [Python実装アーキテクチャ](#python実装アーキテクチャ)
- [ストレージアーキテクチャ](#ストレージアーキテクチャ)
- [使用例](#使用例)
- [エラーハンドリング](#エラーハンドリング)

## データモデル

### コアタイプ（Python実装）

```python
from datetime import datetime, UTC
from pydantic import BaseModel, Field

class Memory(BaseModel):
    """Memory represents a stored memory item."""
    
    id: str = Field(
        default_factory=lambda: f"memory_{int(datetime.now().timestamp() * 1_000_000)}"
    )
    category: str = Field(..., description="Category of the memory")
    key: str = Field(default="", description="Optional user-friendly alias")
    value: str = Field(..., description="The actual memory content")
    tags: list[str] = Field(
        default_factory=list, description="Related tags for future search"
    )
    created_at: datetime = Field(default_factory=lambda: datetime.now(UTC))
    updated_at: datetime = Field(default_factory=lambda: datetime.now(UTC))
    
    # Semantic search fields (Phase 3予定)
    embedding: list[float] | None = Field(
        default=None, description="Semantic embedding vector"
    )

class OperationLog(BaseModel):
    """Log entry for memory operations."""
    
    timestamp: datetime = Field(default_factory=lambda: datetime.now(UTC))
    operation_id: str = Field(default_factory=lambda: f"op_{uuid4().hex[:8]}")
    operation: str = Field(..., description="Operation type (save, get, delete, etc.)")
    key: str | None = Field(default=None, description="Memory key if applicable")
    before: Memory | None = Field(
        default=None, description="Memory state before operation"
    )
    after: Memory | None = Field(
        default=None, description="Memory state after operation"
    )
    success: bool = Field(default=True, description="Whether operation succeeded")
    error: str | None = Field(
        default=None, description="Error message if operation failed"
    )
```

### 検索タイプ

```python
class SearchResult(BaseModel):
    """Search result with relevance score."""
    
    memory: Memory
    score: float = Field(..., ge=0.0, le=1.0, description="Relevance score (0.0 - 1.0)")

class SearchQuery(BaseModel):
    """Search query parameters."""
    
    query: str = Field(..., description="Search query string")
    category: str | None = Field(default=None, description="Optional category filter")
    limit: int = Field(
        default=20, ge=1, le=100, description="Maximum number of results"
    )
    min_score: float = Field(
        default=0.0, ge=0.0, le=1.0, description="Minimum relevance score"
    )
```

### エラーハンドリング

```python
class MemoryNotFoundError(Exception):
    """Raised when a memory is not found."""
    
    def __init__(self, key: str) -> None:
        self.key = key
        super().__init__(f"Memory not found: {key}")
```

## MCPツールリファレンス（Python実装）

### コアメモリツール

#### 1. save_memory

カテゴリ、キー、値を指定して情報を保存します。

**パラメータ:**
- `category` (string, 必須): メモリのカテゴリ
- `value` (string, 必須): 保存する値
- `key` (string, オプション): メモリのユーザーフレンドリーな別名
- `tags` (array[string], オプション): 検索用タグのリスト

**戻り値:**
- `memory_id` (string): 自動生成されたメモリID

**例:**
```json
{
  "category": "personal",
  "value": "誕生日は1990年5月15日",
  "key": "birthday"
}
```

#### 2. get_memory

キーまたはIDでメモリを取得します。

**パラメータ:**
- `key` (string, 必須): 取得するメモリキーまたはID

**例:**
```json
{
  "key": "birthday"
}
```

#### 3. list_memories

すべてのメモリを一覧表示、またはカテゴリで絞り込み（時系列ソート）。

**パラメータ:**
- `category` (string, オプション): カテゴリフィルタ

**例:**
```json
{
  "category": "personal"
}
```

### 高度な検索ツール（Phase 2）

#### 4. search_memories

関連度スコアリング付きの高度なテキスト検索。

**パラメータ:**
- `query` (string, 必須): 検索クエリ文字列
- `category` (string, オプション): オプションのカテゴリフィルタ

**機能:**
- メモリコンテンツ全体にわたる全文検索
- 曖昧マッチング付き関連度スコアリング
- 大文字小文字を区別しない検索
- カテゴリフィルタリングサポート

**例:**
```json
{
  "query": "プログラミング Go言語",
  "category": "learning"
}
```

#### 5. delete_memory

指定されたキーまたはIDでメモリを削除します。

**パラメータ:**
- `key` (string, 必須): 削除するメモリキーまたはID

**例:**
```json
{
  "key": "old-note-id"
}
```

## Python実装アーキテクチャ

### 非同期処理

Python実装では`async/await`パターンを使用：

```python
async def save(self, memory: Memory) -> str:
    """Save a memory and return its ID."""
    self._memories[memory.id] = memory
    await self.log_operation(log)
    await self._save_memories()
    return memory.id
```

### 型安全性

Pydantic v2を使用したデータ検証：

```python
# 自動的にバリデーションとシリアライゼーション
memory = Memory(
    category="learning",
    value="Python async programming",
    tags=["python", "async"]
)
```

### 将来の機能拡張

- 🚧 **Obsidian連携**: ノートインポート/エクスポート機能（開発予定）
- 🚧 **セマンティック検索**: sentence-transformersによる意味検索
- 🚧 **AI自動分類**: 自動カテゴリ化・タグ付け機能

## ストレージアーキテクチャ

### JSONベースストレージ

Python実装では、JSONファイルベースの永続化を使用：

```
data/
├── memories.json      # メモリデータ
└── operations.json    # 操作ログ
```

### データ互換性

Go実装との完全な互換性を維持：

- 同一のJSONスキーマ
- 同一のID生成アルゴリズム
- 同一のタイムスタンプ形式（ISO 8601 UTC）

## 設定

### 環境変数

- `MORY_DATA_DIR`: カスタムデータディレクトリパス（デフォルト: `./data`）

### Claude Desktop連携

Claude Desktop設定に追加:

**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
**Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "mory": {
      "command": "python",
      "args": ["/full/path/to/mory/main.py"],
      "env": {
        "PYTHONPATH": "/full/path/to/mory/src"
      }
    }
  }
}
```

## 使用例

### 基本メモリ操作
```json
// メモリの保存
{
  "tool": "save_memory",
  "parameters": {
    "category": "learning",
    "key": "python-async",
    "value": "Pythonのasync/awaitパターンについて学習",
    "tags": ["python", "async", "programming"]
  }
}

// メモリの取得
{
  "tool": "get_memory",
  "parameters": {
    "key": "python-async"
  }
}

// カテゴリ別メモリ一覧
{
  "tool": "list_memories",
  "parameters": {
    "category": "learning"
  }
}
```

### 高度な検索
```json
// 全文検索
{
  "tool": "search_memories",
  "parameters": {
    "query": "async await Python",
    "category": "learning"
  }
}

// 全カテゴリ横断検索
{
  "tool": "search_memories",
  "parameters": {
    "query": "API設計パターン"
  }
}
```

## エラーハンドリング

### 一般的なエラー

- `MemoryNotFoundError`: 指定されたキーやIDのメモリが見つからない
- `ValidationError`: 無効なパラメータや型エラー
- `FileNotFoundError`: データファイルにアクセスできない

### エラーレスポンス例

```json
{
  "error": {
    "type": "MemoryNotFoundError",
    "message": "Memory not found: invalid-key",
    "details": {
      "key": "invalid-key"
    }
  }
}

