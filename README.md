# 🦔 Mory - Personal Memory MCP Server

> **✅ Ready for Use!** Mory MVP Phase 1 is complete and ready for Claude Desktop integration.

Mory is a Model Context Protocol (MCP) server that provides personal memory functionality to Claude Desktop and other MCP-compatible clients. Store and retrieve information across conversations with ease.

*ClaudeにパーソナルメモリーサービスWebAPIを追加するMCPサーバーです。ChatGPTのメモリ機能のように、会話を跨いで情報を記憶し、よりパーソナライズされた対話を実現します。*

## 🎯 プロジェクトビジョン

Mory（モリー）は、ハリネズミのように小さくて賢い、あなただけのメモリアシスタントです。Claudeに永続的なメモリ機能を追加し、ユーザーの情報を記憶して文脈に応じた返答ができるようにします。すべてのデータはローカルに保存され、ユーザーが完全にコントロールできます。

## 🚀 実装済み機能（MVP Phase 1）

- ✅ **永続的メモリ**: 会話を跨いで個人情報を保存・取得
- ✅ **プライバシー重視**: すべてのデータをローカル保存、クラウド依存なし
- ✅ **カテゴリ管理**: 効率的な情報整理と検索
- ✅ **操作ログ**: 情報の更新履歴を追跡
- ✅ **エラーハンドリング**: 適切なエラーメッセージと安定動作
- ✅ **MCPツール**: save_memory, get_memory, list_memories

## 📋 開発ロードマップ

### ✅ MVP Phase 1: 基本的なメモリ操作（完了）
- シンプルなキー・バリューストレージ
- 基本的な保存・取得操作
- ローカルJSONファイルストレージ
- **明示的な指示による保存のみ**（「覚えて」「記憶して」「メモして」）

### Phase 2: 検索機能の強化 + スマート提案
- 部分一致検索
- 複数の検索戦略
- 検索結果のランキング
- **確認付き自動提案**（重要な個人情報を検出時に保存を提案）

### Phase 3: 高度な機能 + 自動化
- セマンティック検索
- 自動カテゴリ分類
- 競合解決メカニズム
- データ暗号化
- **カスタマイズ可能な自動保存**（ユーザー設定に基づく）

### Phase 4: ユーザー体験
- メモリ管理UI
- 一括操作
- 分析・インサイト機能
- 保存ルールのカスタマイズUI

## 🛠️ 技術スタック

- **言語**: Go
- **プロトコル**: MCP (Model Context Protocol)
- **ストレージ**: JSONファイル (MVP)、SQLite (将来)
- **ランタイム**: Go 1.21+

## 📁 プロジェクト構造

```
mory/
├── cmd/
│   └── mory/
│       └── main.go         # エントリーポイント
├── internal/
│   ├── memory/
│   │   ├── store.go        # メモリストレージロジック
│   │   ├── search.go       # 検索実装
│   │   └── types.go        # 型定義
│   ├── mcp/
│   │   ├── server.go       # MCPサーバー実装
│   │   └── handlers.go     # ツールハンドラー
│   └── config/
│       └── config.go       # 設定管理
├── data/
│   └── memories.json       # ローカルストレージ (git-ignored)
├── test/
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## 🚀 クイックスタート

> **📖 詳細なセットアップガイド**: [QUICKSTART.md](./QUICKSTART.md) をご覧ください

### 1. インストール

```bash
# リポジトリのクローン
git clone https://github.com/nyasuto/mory.git
cd mory

# 依存関係のインストールとビルド
make build

# バイナリが生成されることを確認
./bin/mory --version
```

### 2. Claude Desktop設定

Claude Desktopの設定ファイルに以下を追加：

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

### 3. テスト

Claude Desktopを再起動後、以下を試してください：

```
私の誕生日は1990年5月15日です。記憶してください。
↓
save_memoryツールが実行され、保存成功メッセージが表示される

私の誕生日はいつですか？
↓
get_memoryツールで「1990年5月15日」が返される
```

## 📝 MVP Phase 1 仕様

### コア機能

1. **メモリの保存** 🦔
   - ツール: `save_memory`
   - パラメータ: category, value, key (オプション)
   - タイムスタンプ付きで情報を保存
   - プロンプト: 会話中の重要な個人情報を自動的に判断して保存

2. **メモリの取得** 
   - ツール: `get_memory`
   - パラメータ: key またはid
   - 完全一致で取得

3. **メモリの一覧表示**
   - ツール: `list_memories`
   - パラメータ: category (オプション)
   - 保存されたすべての情報を時系列で表示
   - **重要**: ユーザーに関する質問があった場合、最初に必ず呼ぶべきツール

4. **メモリの削除**
   - ツール: `delete_memory`
   - パラメータ: key またはid
   - 指定したメモリを削除

### 保存ルール（Phase別）

#### 🦔 Phase 1: 明示的な指示のみ（現在）
Claudeは以下の場合のみメモリを保存します：
- 「覚えて」「記憶して」「メモして」などの明示的な指示がある場合
- 例: 「私の誕生日は5月15日です。覚えておいて」

#### Phase 2: 確認付き自動提案
Claudeが重要な個人情報を検出した場合：
- 保存前に必ず確認を取る
- 例: 「この情報を記憶しておきましょうか？」
- 対象: 名前、誕生日、家族情報、好み、重要な予定など

#### Phase 3: カスタマイズ可能な自動保存
ユーザーの設定に基づいて：
- 自動保存レベルの調整（なし/確認/自動）
- カテゴリ別の保存設定
- NGワードリストの管理

### データモデル

```go
type Memory struct {
    ID        string    `json:"id"`         // 自動生成: memory_20250127123456
    Category  string    `json:"category"`
    Key       string    `json:"key"`        // オプション: ユーザーフレンドリーなエイリアス
    Value     string    `json:"value"`
    Tags      []string  `json:"tags"`       // 関連タグ（将来的な検索用）
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

type OperationLog struct {
    Timestamp   time.Time `json:"timestamp"`
    OperationID string    `json:"operation_id"`
    Operation   string    `json:"operation"`
    Key         string    `json:"key,omitempty"`
    Before      *Memory   `json:"before,omitempty"`
    After       *Memory   `json:"after,omitempty"`
    Success     bool      `json:"success"`
    Error       string    `json:"error,omitempty"`
}
```

### ストレージ

- シンプルなJSONファイル: `data/memories.json`
- 操作ログ: `data/operations.log` (JSONL形式)
- ファイルロックによる同期的な読み書き
- 単一ユーザー想定（同時アクセス制御は最小限）

### キー生成戦略

IDは自動生成されるタイムスタンプベース：
```go
// 自動生成されるID例: memory_20250127123456
id := fmt.Sprintf("memory_%s", time.Now().Format("20060102150405"))
```

これにより：
- ユーザーがキーを考える必要なし
- 時系列での管理が容易
- 重複の心配なし

### 使用例

```go
// メモリの保存（IDは自動生成）
memory := &Memory{
    Category: "personal",
    Key:      "birthday",  // オプション：わかりやすいエイリアス
    Value:    "1990年5月15日",
}
id, err := store.Save(memory)  // "memory_20250127123456"が返される

// メモリの取得（IDまたはキーで検索）
result, err := store.Get("birthday")        // キーで検索
result, err := store.GetByID("memory_20250127123456")  // IDで検索

// カテゴリ別一覧（時系列でソート）
memories, err := store.List("personal")
```

### Claudeとの対話例

**Phase 1（明示的な指示）:**
```
ユーザー: 私の誕生日は5月15日です。覚えておいて
Claude: [save_memory実行] はい、誕生日を5月15日として記憶しました。
        保存ID: memory_20250127123456

ユーザー: 私は大阪に住んでいます
Claude: 大阪にお住まいなんですね。（保存しない）

ユーザー: メモ：好きな色は青
Claude: [save_memory実行] 好きな色が青であることを記憶しました。

ユーザー: 私について何か知ってる？
Claude: [list_memories実行] はい、以下の情報を記憶しています：
        1. 誕生日: 5月15日 (2025-01-27 12:34:56)
        2. 好きな色: 青 (2025-01-27 12:35:20)
```

### プロンプト設計

save_memoryツールには以下のような指示を含めます：

```
会話中に見つけた重要なユーザー情報を保存してください。
明示的な保存指示がなくても、以下の情報は保存対象です：

保存する情報の例：
- 好み: 食べ物、音楽、趣味、ブランド
- 興味: 学習中のトピック、関心事
- 個人情報: 仕事、専門分野、居住地、家族
- 現在の状況: プロジェクト、目標、最近の出来事
- 性格・価値観: 思考スタイル、優先事項
- 習慣・ライフスタイル: 日常のルーティン

フォーマット: "User is [具体的な情報]"
例: "User likes strawberry", "User is learning Go"
```

## 💡 設計思想

Moryは「シンプルさ」を重視しています。複雑な要約・抽出ロジックは実装せず、LLMの賢さを信頼する設計です。

### なぜシンプルで十分なのか

実際の使用経験から、以下のことがわかっています：
- LLMは適切なツールがあれば、賢く使い分けてくれる
- 複雑な要約・統合ロジックは不要
- 「重複を整理して」と言えばLLMが自動で処理
- 適切なプロンプトがあれば、想定以上の使い方をしてくれる

### 操作ログによる改善

すべての操作を記録することで：
- 使用パターンの分析
- デバッグの容易さ
- 将来的なUndo機能の基盤
- ユーザーとの協働的な改善

## 🔒 セキュリティとプライバシー

- すべてのデータはローカルに保存
- 外部サービスへの通信なし
- 将来的に暗号化オプションを追加予定

### 保存されない情報

Moryは以下の情報を自動的に保存しません：
- 一時的な状態（今日の体調、現在の気分）
- 明示的な指示のない会話内容
- パスワードや機密情報

### ユーザーコントロール

- いつでもメモリの削除が可能
- カテゴリ単位での一括管理
- エクスポート/インポート機能（将来実装）

## 🤝 コントリビューション

Moryは開発初期段階です。イシューやプルリクエストを歓迎します！

### 開発ガイドライン

1. MVPはシンプルに保つ
2. 機能より信頼性を優先
3. すべての設計判断をドキュメント化
4. コア機能のテストを作成

### コーディング規約

```bash
# フォーマット
make fmt

# リント
make lint

# テスト
make test
```

## 🦔 なぜ「Mory」？

ハリネズミは小さな体に多くの針（記憶）を持っています。Moryも同じように、コンパクトながら多くの大切な記憶を安全に保管します。

## 🎨 将来的な拡張

### Obsidian連携（計画中）
- `[[概念]]`記法でのリンク生成
- 知識グラフの自動構築
- Markdownエクスポート機能

### 高度な検索（Phase 2以降）
- 類義語での検索
- 時間範囲での絞り込み
- 関連メモリの提案

## 📄 ライセンス

MIT License - このプロジェクトを自由に使用・改変してください。

## 🙏 謝辞

- MCPプロトコルを作成したAnthropic
- インスピレーションとフィードバックをくれたClaudeコミュニティ

---

**現在のステータス**: ✅ MVP Phase 1 完了 - Claude Desktop対応済み

**次のステップ**: 
1. ✅ 基本的なストレージメカニズムの実装
2. ✅ MCPツールハンドラーの作成
3. 🚧 Claude Desktopでのテスト ([QUICKSTART.md](./QUICKSTART.md) 参照)
4. 📋 フィードバックを元にPhase 2機能検討

## 💬 お問い合わせ

質問や提案がある場合は、Issueを作成してください。

---

*Mory - Your personal memory hedgehog for Claude 🦔*