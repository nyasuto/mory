# Mory Server デプロイメントガイド

Mory Server の本番環境デプロイメント手順書

## 📋 概要

このガイドでは、Mory Server を Ubuntu/Debian ベースのホームサーバーに本番環境としてデプロイする手順を説明します。

## 🎯 対象環境

- **OS**: Ubuntu 20.04+ / Debian 11+
- **Python**: 3.11+
- **アーキテクチャ**: x86_64 / ARM64
- **RAM**: 最小 512MB、推奨 1GB+
- **ストレージ**: 最小 2GB、推奨 10GB+

## 📦 事前準備

### システム要件の確認

```bash
# OS バージョン確認
lsb_release -a

# Python バージョン確認
python3.11 --version

# 必要パッケージのインストール
sudo apt update
sudo apt install -y python3.11 python3.11-venv curl git sqlite3 systemd
```

### uv パッケージマネージャーのインストール

```bash
# uv をグローバルにインストール
curl -LsSf https://astral.sh/uv/install.sh | sh
source ~/.bashrc
```

## 🚀 自動デプロイメント

### ワンクリック デプロイ

```bash
# プロジェクトルートで実行
sudo ./scripts/deploy.sh
```

このスクリプトは以下を自動実行します：

1. システム依存関係の確認
2. サービスユーザー作成 (`mory`)
3. ディレクトリ構成の作成
4. アプリケーションコードのデプロイ
5. Python 仮想環境の設定
6. systemd サービスの設定
7. 環境設定ファイルの作成
8. ログローテーションの設定
9. バックアップスクリプトの設定
10. サービスの開始

### デプロイ後の設定

```bash
# 環境設定ファイルの編集
sudo nano /opt/mory-server/.env

# OpenAI API キーの設定 (セマンティック検索用)
OPENAI_API_KEY=your_openai_api_key_here

# Obsidian Vault パスの設定 (オプション)
MORY_OBSIDIAN_VAULT_PATH=/path/to/your/obsidian/vault

# サービス再起動で設定反映
sudo systemctl restart mory-server
```

## 🔧 手動デプロイメント

自動デプロイが利用できない場合の手動手順：

### 1. サービスユーザー作成

```bash
sudo useradd --system --create-home --shell /bin/bash mory
```

### 2. ディレクトリ作成

```bash
sudo mkdir -p /opt/mory-server/{data,logs,backups}
sudo chown -R mory:mory /opt/mory-server
sudo chmod 750 /opt/mory-server/{data,logs,backups}
```

### 3. アプリケーションデプロイ

```bash
# ソースコードのコピー
sudo cp -r app/ /opt/mory-server/
sudo cp pyproject.toml /opt/mory-server/
sudo cp -r scripts/ /opt/mory-server/
sudo chown -R mory:mory /opt/mory-server
```

### 4. Python 環境設定

```bash
# mory ユーザーとして実行
sudo -u mory bash -c "
cd /opt/mory-server
uv venv --python python3.11
uv sync
"
```

### 5. systemd サービス設定

```bash
# サービスファイルのコピー
sudo cp scripts/mory-server.service /etc/systemd/system/

# サービス有効化
sudo systemctl daemon-reload
sudo systemctl enable mory-server
sudo systemctl start mory-server
```

## 📁 ディレクトリ構成

```
/opt/mory-server/
├── app/                    # アプリケーションコード
├── scripts/                # デプロイ・管理スクリプト
├── data/                   # データベースファイル
│   └── mory.db            # SQLite データベース
├── logs/                   # ログファイル
├── backups/               # バックアップファイル
├── .env                   # 環境設定
├── .venv/                 # Python 仮想環境
└── pyproject.toml         # プロジェクト設定
```

## 🔍 サービス管理

### 基本操作

```bash
# サービス状態確認
sudo systemctl status mory-server

# サービス開始
sudo systemctl start mory-server

# サービス停止
sudo systemctl stop mory-server

# サービス再起動
sudo systemctl restart mory-server

# ログ確認
sudo journalctl -u mory-server -f

# 設定再読み込み
sudo systemctl reload mory-server
```

### ヘルスチェック

```bash
# API ヘルスチェック
curl http://localhost:8080/api/health

# 詳細ヘルスチェック
sudo -u mory /opt/mory-server/scripts/monitor.py

# JSON 形式での出力
sudo -u mory /opt/mory-server/scripts/monitor.py --json
```

## 💾 データ管理

### データ移行

CLI バージョンからのデータ移行：

```bash
# 移行スクリプトの実行
sudo -u mory python3 /opt/mory-server/scripts/migrate_cli_to_server.py \
    /path/to/old/mory.db \
    /opt/mory-server/data/mory.db \
    --backup

# ドライラン（実際の移行前テスト）
sudo -u mory python3 /opt/mory-server/scripts/migrate_cli_to_server.py \
    /path/to/old/mory.db \
    /opt/mory-server/data/mory.db \
    --dry-run
```

### バックアップ

```bash
# 手動バックアップ作成
sudo -u mory /opt/mory-server/scripts/backup.py create

# カスタム名でバックアップ
sudo -u mory /opt/mory-server/scripts/backup.py create --name "pre-update-backup"

# バックアップ一覧表示
sudo -u mory /opt/mory-server/scripts/backup.py list

# 古いバックアップの削除（30日以上前）
sudo -u mory /opt/mory-server/scripts/backup.py cleanup --keep-days 30
```

### バックアップからの復元

```bash
# バックアップからの復元
sudo systemctl stop mory-server
sudo -u mory /opt/mory-server/scripts/backup.py restore /opt/mory-server/backups/backup_file.tar.gz
sudo systemctl start mory-server
```

## 🔧 トラブルシューティング

### 一般的な問題

#### サービスが開始しない

```bash
# エラーログ確認
sudo journalctl -u mory-server --no-pager

# 設定ファイル確認
sudo -u mory /opt/mory-server/.venv/bin/python -c "from app.core.config import settings; print(settings)"

# ポート使用状況確認
sudo netstat -tlnp | grep 8080
```

#### データベース接続エラー

```bash
# データベースファイル確認
ls -la /opt/mory-server/data/mory.db

# 権限確認
sudo -u mory sqlite3 /opt/mory-server/data/mory.db ".schema"

# データベース整合性チェック
sudo -u mory /opt/mory-server/scripts/monitor.py --check database
```

#### メモリ不足

```bash
# メモリ使用量確認
sudo systemctl show mory-server --property=MemoryCurrent

# プロセス確認
ps aux | grep uvicorn

# ワーカー数を削減（mory-server.service）
ExecStart=/opt/mory-server/.venv/bin/uvicorn app.main:app --host 0.0.0.0 --port 8080 --workers 1
```

### ログファイル

- システムログ: `journalctl -u mory-server`
- アプリケーションログ: `/opt/mory-server/logs/`
- アクセスログ: systemd journal

## 🔒 セキュリティ

### ファイアウォール設定

```bash
# UFW でポート開放
sudo ufw allow 8080/tcp

# 特定のIPからのみアクセス許可
sudo ufw allow from 192.168.1.0/24 to any port 8080
```

### SSL/TLS 設定（オプション）

nginx リバースプロキシ経由でSSL終端：

```nginx
server {
    listen 443 ssl;
    server_name your-domain.com;
    
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## 📊 監視・アラート

### Cron による定期ヘルスチェック

```bash
# mory ユーザーの crontab に追加
sudo -u mory crontab -e

# 5分毎のヘルスチェック
*/5 * * * * /opt/mory-server/scripts/monitor.py --check api > /dev/null 2>&1 || echo "Mory API is down" | mail -s "Mory Alert" admin@example.com
```

### システムメトリクス監視

- CPU使用率: `htop`, `top`
- メモリ使用率: `free -h`
- ディスク使用率: `df -h`
- ネットワーク: `netstat -i`

## 🚀 パフォーマンス最適化

### データベース最適化

```bash
# SQLite 最適化
sudo -u mory sqlite3 /opt/mory-server/data/mory.db "VACUUM; ANALYZE;"
```

### システム設定

```bash
# ファイルディスクリプタ制限の調整
echo "mory soft nofile 65536" | sudo tee -a /etc/security/limits.conf
echo "mory hard nofile 65536" | sudo tee -a /etc/security/limits.conf
```

## 📈 アップグレード手順

### アプリケーション更新

```bash
# 1. サービス停止
sudo systemctl stop mory-server

# 2. バックアップ作成
sudo -u mory /opt/mory-server/scripts/backup.py create --name "pre-upgrade-$(date +%Y%m%d)"

# 3. 新しいコードのデプロイ
sudo cp -r app/ /opt/mory-server/
sudo cp pyproject.toml /opt/mory-server/
sudo chown -R mory:mory /opt/mory-server

# 4. 依存関係更新
sudo -u mory bash -c "cd /opt/mory-server && uv sync"

# 5. データベースマイグレーション（必要に応じて）
# sudo -u mory python /opt/mory-server/scripts/migrate.py

# 6. サービス開始
sudo systemctl start mory-server

# 7. ヘルスチェック
curl http://localhost:8080/api/health
```

## 📞 サポート

問題が発生した場合：

1. ログファイルの確認
2. ヘルスチェックスクリプトの実行
3. システムリソースの確認
4. GitHub Issues での報告

## 📄 参考資料

- [Mory Server README](README.md)
- [API ドキュメント](http://localhost:8080/docs)
- [systemd サービス管理](https://www.freedesktop.org/software/systemd/man/systemctl.html)
- [SQLite 最適化](https://www.sqlite.org/optoverview.html)