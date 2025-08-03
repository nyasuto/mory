# Docker による Mory Server 起動ガイド

Mory ServerをDockerで起動する方法について説明します。

## 📋 概要

Mory ServerはDockerとdocker-composeを使用して簡単にデプロイできます。Docker版では以下の機能が利用可能です：

- ✅ FTS5検索
- ✅ セマンティック検索（OpenAI API Key設定時）
- ✅ REST API
- ✅ MCP Server機能
- ✅ Obsidian統合（ボリュームマウント時）

## 🚀 クイックスタート

### 1. 環境設定ファイルの準備

```bash
# 環境設定テンプレートをコピー
cp .env.example .env

# OpenAI API Keyを設定（セマンティック検索を使用する場合）
nano .env
```

### 2. Docker Compose での起動

```bash
# バックグラウンドで起動
docker-compose up -d

# ログを確認
docker-compose logs -f mory-server
```

### 3. 動作確認

```bash
# ヘルスチェック
curl http://localhost:8080/api/health

# API ドキュメント
open http://localhost:8080/docs
```

## ⚙️ 環境設定

### 必須設定

`.env`ファイルで以下の設定を行ってください：

```bash
# OpenAI API Key（セマンティック検索を使用する場合は必須）
OPENAI_API_KEY=your_openai_api_key_here
```

### オプション設定

```bash
# セマンティック検索の無効化
MORY_SEMANTIC_SEARCH_ENABLED=false

# 使用するOpenAIモデルの変更
MORY_OPENAI_MODEL=text-embedding-3-small

# デバッグモードの有効化
MORY_DEBUG=true

# Obsidian統合（ボリュームマウント設定も必要）
MORY_OBSIDIAN_VAULT_PATH=/obsidian
```

## 📁 データ永続化

### データベース

```bash
# データは自動的にボリュームに保存されます
docker volume ls | grep mory

# データベースの場所
docker-compose exec mory-server ls -la /app/data/
```

### バックアップ

```bash
# データベースをホストにコピー
docker-compose exec mory-server cp /app/data/memories.db /tmp/
docker cp $(docker-compose ps -q mory-server):/tmp/memories.db ./backup-$(date +%Y%m%d).db
```

## 🔗 Obsidian統合

Obsidian Vaultを統合する場合：

```yaml
# docker-compose.yml に追加
services:
  mory-server:
    volumes:
      - ./path/to/your/obsidian/vault:/obsidian:ro
    environment:
      - MORY_OBSIDIAN_VAULT_PATH=/obsidian
```

## 🔍 トラブルシューティング

### よくある問題

#### 1. API Keyが読み込まれない

```bash
# 環境変数の確認
docker-compose exec mory-server env | grep OPENAI

# 設定の確認
docker-compose exec mory-server python -c "from app.core.config import settings; print(f'API Key: {settings.openai_api_key[:10] if settings.openai_api_key else None}...')"
```

#### 2. セマンティック検索が無効

```bash
# セマンティック検索の状態確認
curl http://localhost:8080/api/health/detailed | jq '.semantic_search'

# ログでエラーを確認
docker-compose logs mory-server | grep -i openai
```

#### 3. データが保持されない

```bash
# ボリュームの確認
docker volume inspect mory_mory_data

# データディレクトリの確認
docker-compose exec mory-server ls -la /app/data/
```

### ログの確認

```bash
# リアルタイムログ
docker-compose logs -f mory-server

# エラーログのみ
docker-compose logs mory-server 2>&1 | grep -i error

# 起動ログの確認
docker-compose logs mory-server | head -20
```

## 🛠️ 開発環境

開発用設定を使用する場合：

```bash
# 開発用docker-compose使用
docker-compose -f docker-compose.dev.yml up -d

# ライブリロード付きで起動
docker-compose -f docker-compose.dev.yml up
```

## 🔒 セキュリティ

### プロダクション環境での注意点

```bash
# セキュアな.envファイル権限
chmod 600 .env

# ファイアウォール設定
sudo ufw allow 8080/tcp

# リバースプロキシの設定（推奨）
# nginx等でSSL終端を行う
```

## 📊 監視

### ヘルスチェック

```bash
# 基本ヘルスチェック
curl http://localhost:8080/api/health

# 詳細ヘルスチェック（DB接続、FTS5、セマンティック検索状態）
curl http://localhost:8080/api/health/detailed
```

### リソース監視

```bash
# Docker統計情報
docker stats $(docker-compose ps -q mory-server)

# メモリ使用量
docker-compose exec mory-server cat /proc/meminfo | grep MemAvailable
```

## 🔄 更新手順

### アプリケーションの更新

```bash
# 1. サービス停止
docker-compose down

# 2. 新しいイメージをプル/ビルド
docker-compose pull  # または
docker-compose build

# 3. サービス再開
docker-compose up -d

# 4. 動作確認
curl http://localhost:8080/api/health
```

### 設定変更の反映

```bash
# .envファイル変更後
docker-compose restart mory-server

# docker-compose.yml変更後
docker-compose down && docker-compose up -d
```

## 💡 便利なコマンド

```bash
# サービス状態確認
docker-compose ps

# コンテナ内でシェル起動
docker-compose exec mory-server bash

# データベースアクセス
docker-compose exec mory-server sqlite3 /app/data/memories.db

# 設定確認
docker-compose exec mory-server python -c "from app.core.config import settings; print(vars(settings))"

# すべて削除（データ含む）
docker-compose down -v
docker volume rm mory_mory_data
```

## 📚 参考情報

- [API ドキュメント](http://localhost:8080/docs)
- [基本的な使い方](./docs/QUICKSTART.md)
- [本格的なデプロイメント](./DEPLOYMENT.md)
- [MCP Server設定](./README.md#mcp-server)