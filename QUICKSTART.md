# 🦔 Mory POC クイックスタートガイド

Claude DesktopでMoryメモリ機能を即座にテストするためのガイドです。

## 🚀 セットアップ（2分で完了）

### 1. バイナリの準備
```bash
# プロジェクトルートディレクトリで
cd /path/to/mory
make build

# バイナリが生成されることを確認
ls -la bin/mory
./bin/mory --version
```

### 2. Claude Desktop設定

Claude Desktopの設定ファイルを編集：

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

> **重要**: `/full/path/to/mory/bin/mory` を実際のパスに置き換えてください

### 3. Claude Desktop再起動

設定変更後、Claude Desktopを完全に終了して再起動してください。

## 🧪 テストシナリオ

### シナリオ1: 基本的な記憶と取得

**Step 1: 情報を記憶させる**
```
私の誕生日は1990年5月15日です。記憶してください。
```
期待結果：✅ save_memoryツールが実行され、保存成功メッセージが表示される

**Step 2: 記憶した情報を確認**
```
私の誕生日はいつですか？
```
期待結果：✅ get_memoryツールで「1990年5月15日」が返される

### シナリオ2: カテゴリ別情報管理

**Step 1: 個人情報を保存**
```
名前：田中太郎
住所：東京都渋谷区
職業：エンジニア

これらの情報を記憶してください。
```

**Step 2: 好みを保存**
```
好きな色：青
好きな食べ物：ラーメン
趣味：プログラミング

これらの好みを記憶して。
```

**Step 3: 保存された情報を確認**
```
これまで記憶した情報をすべて見せてください。
```
期待結果：✅ list_memoriesで分類された情報が一覧表示される

### シナリオ3: エラーハンドリング確認

**存在しない情報の検索**
```
私のペットの名前は何ですか？
```
期待結果：✅ 「情報が見つかりません」的なメッセージが表示される

## 🔍 動作確認ポイント

### 正常動作の確認
- [ ] Claude DesktopでMCPツールアイコンが表示される
- [ ] save_memory実行時に成功メッセージが表示される
- [ ] get_memory実行時に正しい情報が返される
- [ ] list_memories実行時に保存済み情報が一覧表示される
- [ ] 存在しない情報検索時に適切なエラーメッセージが表示される

### データ確認
```bash
# 保存されたデータファイルを確認
cat data/memories.json

# 操作ログを確認  
cat data/operations.log
```

## 🐛 トラブルシューティング

### Claude DesktopでMoryが表示されない
1. パスが正しいか確認：`/full/path/to/mory/bin/mory`
2. バイナリに実行権限があるか確認：`chmod +x bin/mory`
3. Claude Desktopを完全再起動
4. 設定ファイルのJSON構文が正しいか確認

### ツール実行時にエラーが発生
```bash
# サーバーログを確認（別ターミナルで）
./bin/mory

# 設定ファイルが適切に読み込まれているか確認
ls -la config.json data/
```

### データが保存されない
```bash
# dataディレクトリの権限確認
ls -la data/
mkdir -p data  # 必要に応じて作成
```

## 📝 テスト結果の記録

### 動作確認チェックリスト
- [ ] バイナリビルド成功
- [ ] Claude Desktop設定完了
- [ ] MCPサーバー接続確認
- [ ] save_memoryツール動作確認
- [ ] get_memoryツール動作確認  
- [ ] list_memoriesツール動作確認
- [ ] エラーハンドリング確認
- [ ] データ永続化確認

### 発見した問題
<!-- テスト中に発見した問題をここに記録 -->

### 改善提案
<!-- テスト結果を基にした改善提案をここに記録 -->

## 🎯 次のステップ

POCテストが完了したら：
1. **Issue #32** に結果を報告
2. 発見した問題の修正
3. 実用化に向けた機能追加検討
4. パフォーマンス最適化

---

**Happy Testing! 🦔**