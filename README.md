# 🦔 Mory - パーソナルメモリMCPサーバー

> **🐍 Python版PoC開始！** Python実装によるモダンなMCPサーバーをClaude Desktopで利用可能です。

MoryはModel Context Protocol (MCP) サーバーで、Claude DesktopやMCP対応クライアントにパーソナルメモリ機能を提供します。会話を跨いで情報を記憶し、簡単に取得できます。

*ChatGPTのメモリ機能のように、Claudeに永続的なメモリ機能を追加して、よりパーソナライズされた対話を実現するPython製MCPサーバーです。*

## 🎯 主要機能

### コアメモリシステム（Python実装）
- ✅ **永続的メモリ**: 会話を跨いで個人情報を記憶・取得
- ✅ **プライバシー重視**: すべてのデータをローカル保存、クラウド依存なし
- ✅ **カテゴリ管理**: 効率的な情報整理
- ✅ **操作ログ**: すべてのメモリ操作を監査証跡として記録
- ✅ **非同期処理**: Python async/awaitによる高速処理
- ✅ **型安全性**: Pydantic v2によるデータ検証

### 高度な検索機能
- ✅ **全文検索**: 関連度スコアリング付きの高度なテキスト検索
- ✅ **スマートフィルタリング**: カテゴリベースの絞り込みと曖昧検索
- ✅ **関連度ランキング**: スコアベースの検索結果順位付け
- 🚧 **セマンティック検索**: sentence-transformersによる意味検索（Phase 3）

### 将来の機能拡張
- 🚧 **Obsidian連携**: ノートインポート/エクスポート機能
- 🚧 **AI要約**: メモリ内容の自動要約・カテゴリ化
- 🚧 **推奨システム**: 関連メモリの自動提案

## 🚀 クイックスタート

### 1. Python環境のセットアップ
```bash
git clone https://github.com/nyasuto/mory.git
cd mory

# 仮想環境の作成
python3 -m venv venv
source venv/bin/activate  # Linux/macOS
# venv\Scripts\activate    # Windows

# 依存関係のインストール
pip install -e .
```

### 2. Claude Desktop設定
Claude Desktop設定ファイルに追加：

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

### 3. 基本的な使用方法
```
私の誕生日は1990年5月15日です。記憶してください。
→ save_memory ツールが実行されます

私について何か知ってる？
→ list_memories ツールが記憶した情報を表示します

プログラミングに関する記憶を検索して
→ search_memories ツールが関連する記憶を検索します
```

**📖 詳細なセットアップガイド**: [QUICKSTART.md](./docs/QUICKSTART.md) で詳しい手順と例を確認してください。

**🔧 技術仕様**: [API.md](./docs/API.md) で詳細なAPIリファレンスと技術仕様を確認してください。

## 🛠️ 利用可能なMCPツール（Python実装）

1. **save_memory** - カテゴリとタグ付きで情報を保存
2. **get_memory** - キーやIDで特定のメモリを取得  
3. **list_memories** - オプションのカテゴリフィルタ付きでメモリを一覧表示
4. **search_memories** - 関連度スコアリング付きの高度な全文検索
5. **delete_memory** - 指定したメモリを削除（新機能）

> **注意**: Python版ではObsidian連携機能は開発中です。コアメモリ機能のPoCが完了後に実装予定です。

## 📋 開発状況

### 🐍 Python版PoC（現在）
- ✅ **コアメモリ管理**: save/get/list/delete/search
- ✅ **関連度スコアリング付き検索**: 高度なテキスト検索
- ✅ **非同期処理**: Python async/awaitによる高速化
- ✅ **型安全性**: Pydantic v2によるデータ検証
- ✅ **テストカバレッジ**: 95%+ Go実装互換性テスト
- ✅ **Claude Desktop対応**: 完全なMCP v1.0互換

### 📋 今後の予定
- 🚧 **Obsidian連携**: ノートインポート/エクスポート機能
- 🚧 **セマンティック検索**: sentence-transformersによる意味検索
- 🚧 **AI自動カテゴリ化**: 自動分類・タグ付け機能
- 🚧 **推奨システム**: 関連メモリの自動提案

## 🔒 プライバシー・セキュリティ

- **ローカル専用ストレージ**: すべてのデータはあなたのマシンに保存
- **外部依存なし**: 完全にオフラインで動作
- **ユーザーコントロール**: 何をいつ保存するかを完全制御
- **監査証跡**: 透明性のための完全な操作ログ

## 🤝 コントリビューション

Moryは活発に開発中です。IssueやPull Requestを歓迎します！

### 開発コマンド
```bash
make build     # プロジェクトのビルド
make test      # テストの実行
make quality   # すべての品質チェック（フォーマット、リント、テスト）
```

詳細なガイドラインは [CONTRIBUTING.md](./CONTRIBUTING.md) を参照してください。

## 🦔 なぜ「Mory」？

ハリネズミが小さな体に多くの針（記憶）を安全に収めるように、Moryはあなたの大切な記憶を安全でアクセスしやすい形で保管します。

## 📄 ライセンス

MIT License - このプロジェクトを自由に使用・改変してください。

---

**現在のステータス**: ✅ Phase 2 完了 - 検索・Obsidian連携機能がプロダクション対応

**クイックリンク**:
- 📖 [完全セットアップガイド](./docs/QUICKSTART.md)
- 🔧 [技術文書](./docs/API.md)
- 🚀 [コントリビューションガイド](./CONTRIBUTING.md)

*Mory - あなたのClaudeのためのパーソナルメモリハリネズミ 🦔*