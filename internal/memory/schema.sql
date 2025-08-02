-- Mory SQLite Schema
-- 記憶データ管理とセマンティック検索用のSQLiteスキーマ定義

-- メイン記憶テーブル
CREATE TABLE IF NOT EXISTS memories (
    id TEXT PRIMARY KEY,
    category TEXT NOT NULL,
    key TEXT,                    -- NULL許可（オプション）
    value TEXT NOT NULL,
    tags TEXT,                   -- JSON配列として格納 ["tag1", "tag2"]
    created_at INTEGER NOT NULL, -- Unix timestamp (nanoseconds)
    updated_at INTEGER NOT NULL, -- Unix timestamp (nanoseconds)
    
    -- セマンティック検索用フィールド
    embedding BLOB,              -- バイナリエンコードされたembedding
    embedding_hash TEXT          -- キャッシュ無効化用ハッシュ
);

-- パフォーマンス最適化用インデックス
CREATE INDEX IF NOT EXISTS idx_memories_category ON memories(category);
CREATE INDEX IF NOT EXISTS idx_memories_key ON memories(key) WHERE key IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_memories_created_at ON memories(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_memories_updated_at ON memories(updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_memories_embedding_hash ON memories(embedding_hash) WHERE embedding_hash IS NOT NULL;

-- 全文検索用仮想テーブル（FTS5）
CREATE VIRTUAL TABLE IF NOT EXISTS memories_fts USING fts5(
    id UNINDEXED,
    category,
    key,
    value,
    tags,
    content='memories',
    content_rowid='rowid',
    tokenize='unicode61 remove_diacritics 2'
);

-- FTS5自動更新用トリガー
CREATE TRIGGER IF NOT EXISTS memories_fts_insert AFTER INSERT ON memories BEGIN
    INSERT INTO memories_fts(id, category, key, value, tags)
    VALUES (new.id, new.category, new.key, new.value, new.tags);
END;

CREATE TRIGGER IF NOT EXISTS memories_fts_update AFTER UPDATE ON memories BEGIN
    UPDATE memories_fts SET
        category = new.category,
        key = new.key,
        value = new.value,
        tags = new.tags
    WHERE id = new.id;
END;

CREATE TRIGGER IF NOT EXISTS memories_fts_delete AFTER DELETE ON memories BEGIN
    DELETE FROM memories_fts WHERE id = old.id;
END;

-- 操作ログテーブル
CREATE TABLE IF NOT EXISTS operation_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp INTEGER NOT NULL,  -- Unix timestamp (nanoseconds)
    operation_id TEXT NOT NULL,
    operation TEXT NOT NULL,     -- 'save', 'update', 'delete'
    memory_key TEXT,             -- 操作対象のkey（あれば）
    memory_id TEXT,              -- 操作対象のID
    before_data TEXT,            -- JSON形式の変更前データ
    after_data TEXT,             -- JSON形式の変更後データ
    success BOOLEAN NOT NULL,
    error TEXT                   -- エラーメッセージ（失敗時）
);

-- 操作ログのインデックス
CREATE INDEX IF NOT EXISTS idx_operation_logs_timestamp ON operation_logs(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_operation_logs_operation ON operation_logs(operation);
CREATE INDEX IF NOT EXISTS idx_operation_logs_memory_id ON operation_logs(memory_id);

-- データベース設定とパフォーマンス最適化
PRAGMA journal_mode = WAL;       -- Write-Ahead Logging for better concurrency
PRAGMA synchronous = NORMAL;     -- Balance between safety and performance
PRAGMA cache_size = 10000;       -- 10MB cache
PRAGMA temp_store = memory;      -- Use memory for temporary storage
PRAGMA mmap_size = 268435456;    -- 256MB memory-mapped I/O