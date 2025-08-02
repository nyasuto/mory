# Mory Python - セマンティック検索対応パーソナルメモリMCPサーバー

MoryのPython実装版です。セマンティック検索機能を搭載し、既存のGoバージョンとの完全な互換性を提供します。

## 🚀 機能

### Phase 1: 基盤機能（実装済み）
- ✅ **メモリ管理**: 記憶の保存、取得、一覧表示、削除
- ✅ **カテゴリ分類**: 記憶をカテゴリで整理
- ✅ **タグ付け**: 柔軟な分類とフィルタリング
- ✅ **文字列検索**: キーワードベースの基本検索
- ✅ **Go互換性**: 既存のGoデータとの完全互換

### Phase 2: セマンティック検索（計画中）
- 🔄 **ベクトル検索**: Sentence Transformersによる意味的検索
- 🔄 **ハイブリッド検索**: キーワード + セマンティック検索
- 🔄 **多言語対応**: 日本語・英語の意味理解
- 🔄 **関連性スコア**: 高精度な検索結果ランキング

## 📦 インストール

```bash
# uvのインストール (未インストールの場合)
curl -LsSf https://astral.sh/uv/install.sh | sh

# プロジェクトのクローン
cd mory-python

# 依存関係のインストール
uv pip install -e .

# 開発環境のセットアップ
uv pip install -e ".[dev]"
```

## 🔧 設定

### 環境変数
```bash
export MORY_DATA_DIR="/path/to/your/data"  # デフォルト: "data"
```

### Claude Desktop設定
```json
{
  "mcpServers": {
    "mory-python": {
      "command": "python",
      "args": ["/path/to/mory-python/src/mory/main.py"],
      "env": {
        "MORY_DATA_DIR": "/Users/your-name/.mory"
      }
    }
  }
}
```

## 🛠️ 開発

### 開発コマンド

```bash
# すべての品質チェック（推奨）
make quality

# 個別コマンド
make test                # テスト実行
make lint                # ruffリンター
make format              # ruffフォーマッター
make type-check          # mypy型チェック
make run                 # サーバー起動

# uvを使用した開発環境同期
make uv-sync-dev         # 高速同期
```

### モダンな開発体験

このプロジェクトは最新のPythonエコシステムを採用：

- **Python 3.11+**: モダンな型アノテーション (`X | Y`) と高性能
- **uv**: 高速なPythonパッケージマネージャー
- **ruff**: 超高速なリンター・フォーマッター（Black + isort + flake8の代替）
- **pytest**: 最新のテストフレームワーク
- **mypy**: 静的型チェック

## 📊 Go版との互換性

このPython実装は既存のGo版と100%互換性があります：

- ✅ **データ形式**: 同じJSON形式でデータ保存
- ✅ **ID生成**: 同じタイムスタンプベースのID
- ✅ **検索API**: 同一のインターフェース
- ✅ **MCP仕様**: 同じツール定義と動作

## 🔄 Go版からの移行

1. **データ保持**: 既存のデータディレクトリをそのまま使用
2. **設定変更**: Claude Desktop設定でコマンドをPython版に変更
3. **即座に利用可能**: データ変換や移行作業は不要

## 🎯 ロードマップ

### Phase 2: セマンティック検索（開発中）
- Sentence Transformersによる埋め込み生成
- FAISSによる高速ベクトル検索
- ハイブリッド検索アルゴリズム実装

### Phase 3: 高度な機能（計画中）
- 自動カテゴリ分類
- 関連記憶の推奨
- 検索結果の要約生成

## 📝 API

### 利用可能なMCPツール

1. **save_memory** - 記憶を保存
2. **get_memory** - 記憶を取得
3. **list_memories** - 記憶一覧の表示
4. **search_memories** - 記憶の検索
5. **delete_memory** - 記憶の削除

詳細は[Go版のAPI.md](../API.md)を参照してください（互換性保証）。

## 🤝 貢献

1. このリポジトリをフォーク
2. フィーチャーブランチを作成 (`git checkout -b feature/amazing-feature`)
3. 変更をコミット (`git commit -m 'Add amazing feature'`)
4. ブランチにプッシュ (`git push origin feature/amazing-feature`)
5. プルリクエストを作成

## 📄 ライセンス

このプロジェクトはMITライセンスの下で公開されています。