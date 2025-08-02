# 🔧 Mory APIリファレンス・技術文書

この文書では、Mory MCPサーバーの詳細な技術仕様を提供します。

## 📋 目次

- [データモデル](#データモデル)
- [MCPツールリファレンス](#mcpツールリファレンス)
- [設定](#設定)
- [ストレージアーキテクチャ](#ストレージアーキテクチャ)
- [使用例](#使用例)
- [エラーハンドリング](#エラーハンドリング)

## データモデル

### コアタイプ

```go
type Memory struct {
    ID        string    `json:"id"`         // Auto-generated: memory_20250127123456
    Category  string    `json:"category"`   // User-defined category
    Key       string    `json:"key"`        // Optional user-friendly alias
    Value     string    `json:"value"`      // Stored content
    Tags      []string  `json:"tags"`       // Related tags for search
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

type OperationLog struct {
    Timestamp   time.Time `json:"timestamp"`
    OperationID string    `json:"operation_id"`
    Operation   string    `json:"operation"`    // save, update, delete
    Key         string    `json:"key,omitempty"`
    Before      *Memory   `json:"before,omitempty"`
    After       *Memory   `json:"after,omitempty"`
    Success     bool      `json:"success"`
    Error       string    `json:"error,omitempty"`
}
```

### 検索タイプ（Phase 2）

```go
type SearchResult struct {
    Memory *Memory `json:"memory"`
    Score  float64 `json:"score"` // Relevance score (0.0 - 1.0)
}

type SearchQuery struct {
    Query    string `json:"query"`              // Search query string
    Category string `json:"category,omitempty"` // Optional category filter
}
```

### Obsidian連携タイプ（Phase 2）

```go
type ImportResult struct {
    TotalFiles       int      `json:"total_files"`
    ImportedFiles    int      `json:"imported_files"`
    SkippedFiles     int      `json:"skipped_files"`
    DuplicateFiles   int      `json:"duplicate_files"`
    Errors           []string `json:"errors"`
    ImportedMemories []*Memory `json:"imported_memories"`
    DryRun           bool     `json:"dry_run"`
}

type GeneratedNote struct {
    Title         string        `json:"title"`
    Content       string        `json:"content"`
    OutputPath    string        `json:"output_path"`
    MemoryCount   int          `json:"memory_count"`
    RelatedCount  int          `json:"related_count"`
    UsedMemories  []*Memory    `json:"used_memories"`
    RelatedLinks  []string     `json:"related_links"`
    GeneratedAt   time.Time    `json:"generated_at"`
    TemplateUsed  string       `json:"template_used"`
}
```

## MCPツールリファレンス

### コアメモリツール

#### 1. save_memory

カテゴリ、キー、値を指定して情報を保存します。

**パラメータ:**
- `category` (string, 必須): メモリのカテゴリ
- `value` (string, 必須): 保存する値
- `key` (string, オプション): メモリのユーザーフレンドリーな別名

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

### Obsidian連携ツール（Phase 2）

#### 5. obsidian_import

Obsidianボルトのノートをメモリストレージにインポートします。

**パラメータ:**
- `import_type` (string, 必須): インポート種類: 'vault', 'category', または 'file'
- `path` (string, オプション): 特定ファイルパス（'file'タイプ用）またはカテゴリ名（'category'タイプ用）
- `dry_run` (boolean, オプション): 保存せずにインポートをプレビュー（デフォルト: false）
- `category_mapping` (object, オプション): フォルダ名をカスタムカテゴリにマッピング
- `skip_duplicates` (boolean, オプション): 重複インポートをスキップ（デフォルト: true）

**例:**
```json
{
  "import_type": "vault",
  "dry_run": false,
  "skip_duplicates": true,
  "category_mapping": {
    "Daily Notes": "journal",
    "Work": "professional"
  }
}
```

#### 6. generate_obsidian_note

テンプレートを使用してメモリからObsidianノートを生成します。

**パラメータ:**
- `template` (string, 必須): テンプレートタイプ: 'daily', 'summary', または 'report'
- `title` (string, 必須): 生成されるノートのタイトル
- `category` (string, オプション): メモリのカテゴリフィルタ
- `output_path` (string, オプション): カスタム出力ファイルパス
- `include_related` (boolean, オプション): 関連メモリを含める（デフォルト: true）

**利用可能なテンプレート:**
- **daily**: 日次ノート用の時系列メモリ整理
- **summary**: サマリー用のカテゴリベースメモリ編集
- **report**: 統計付きの構造化プロジェクト文書

**例:**
```json
{
  "template": "daily",
  "title": "今日の学習まとめ",
  "category": "learning",
  "include_related": true
}
```

## 設定

### 環境変数

- `MORY_DATA_DIR`: カスタムデータディレクトリパス
- `MORY_OBSIDIAN_VAULT_PATH`: 連携用Obsidianボルトへのパス

### 設定ファイル

`~/.mory/config.json` を作成するか、環境変数で指定:

```json
{
  "data_dir": "/custom/path/to/data",
  "obsidian": {
    "vault_path": "/path/to/obsidian/vault"
  }
}
```

### Claude Desktop連携

Claude Desktop設定に追加:

**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
**Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "mory": {
      "command": "/full/path/to/mory/bin/mory"
    }
  }
}
```

## ストレージアーキテクチャ

### ファイル構造
```
~/.mory/                           # デフォルトデータディレクトリ
├── memories.json                  # メモリストレージ（JSON）
├── operations.log                 # 操作監査ログ（JSONL）
└── config.json                    # 設定ファイル
```

### ストレージ機能
- **JSONベース**: シンプルで人間が読める形式
- **ファイルロック**: 同時アクセス保護
- **操作ログ**: 完全な監査証跡
- **バックアップフレンドリー**: プレーンテキストファイル、バックアップ・復元が簡単

### キー生成戦略
```go
// 自動生成IDフォーマット: memory_20250127123456
id := fmt.Sprintf("memory_%s", time.Now().Format("20060102150405"))
```

## 使用例

### 基本メモリ操作
```go
// メモリの保存
save_memory: category="learning", key="go-basics", value="Goのチャンネルとゴルーチンについて学習"

// メモリの取得
get_memory: key="go-basics"

// カテゴリ別メモリ一覧
list_memories: category="learning"
```

### 高度な検索
```go
// 全文検索
search_memories: query="ゴルーチン チャンネル", category="learning"

// 全カテゴリ横断検索
search_memories: query="API設計パターン"
```

### Obsidian連携
```go
// ボルト全体のインポート
obsidian_import: import_type="vault", skip_duplicates=true

// マッピング付き特定カテゴリのインポート
obsidian_import: import_type="category", path="Projects", 
                category_mapping={"Projects": "work"}

// 日次ノートの生成
generate_obsidian_note: template="daily", title="今日の進捗", 
                       category="work", include_related=true

// サマリーレポートの生成
generate_obsidian_note: template="report", title="週間まとめ", 
                       output_path="reports/week-summary.md"
```

## エラーハンドリング

### 一般的なエラータイプ

1. **バリデーションエラー**: 無効なパラメータや必須フィールドの不足
2. **ストレージエラー**: ファイルシステムの問題、権限、ディスク容量
3. **設定エラー**: 無効な設定、Obsidianボルトパスの不足
4. **インポートエラー**: Markdownパース問題、ファイルアクセス問題

### エラーレスポンス形式

```json
{
  "error": "メモリが見つかりません",
  "details": "キー 'nonexistent' のメモリが見つかりません",
  "code": "MEMORY_NOT_FOUND"
}
```

### ベストプラクティス

1. **処理前に常に入力を検証**
2. **実行可能な情報と共に明確なエラーメッセージを提供**
3. **デバッグと監査目的で操作をログ記録**
4. **ファイルロックで同時アクセスを適切に処理**
5. **インポート操作前にObsidian設定を検証**

## パフォーマンス考慮事項

### 検索パフォーマンス
- **インデックス化**: 高速テキスト検索のためのインメモリインデックス
- **キャッシュ**: 繰り返しクエリのための検索結果キャッシュ
- **ページ化**: 大きな結果セットを効率的に処理

### メモリ使用量
- **遅延読み込み**: オンデマンドでのメモリ読み込み
- **メモリ制限**: 大きな結果セットの自動クリーンアップ
- **同時安全性**: 適切なロックによるスレッドセーフ操作

### スケーラビリティ
- **ファイルベースストレージ**: 数千のメモリに適している
- **検索最適化**: 関連度スコアリングのための効率的アルゴリズム
- **将来の最適化**: データベースバックエンド移行への準備完了

---

実装詳細については、`internal/` ディレクトリのソースコードを参照してください。
セットアップ手順については、[QUICKSTART.md](./QUICKSTART.md)を参照してください。
貢献ガイドラインについては、[CONTRIBUTING.md](./CONTRIBUTING.md)を参照してください。