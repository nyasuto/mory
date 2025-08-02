#!/usr/bin/env python3
"""
MCP サーバの検索機能テスト
セマンティック検索がMCPプロトコル経由で正常に動作するかテスト
"""

import json
import subprocess
import sys
import time

def test_mcp_search():
    print("🧪 MCP サーバ検索機能テスト")
    print("=" * 30)
    
    # MCPクライアントテスト用のJSONRPCリクエスト
    test_requests = [
        {
            "name": "動物で検索 (セマンティック検索テスト)",
            "request": {
                "jsonrpc": "2.0",
                "id": 1,
                "method": "tools/call",
                "params": {
                    "name": "search_memories",
                    "arguments": {
                        "query": "動物"
                    }
                }
            }
        },
        {
            "name": "ペットで検索 (概念関連性テスト)",
            "request": {
                "jsonrpc": "2.0",
                "id": 2,
                "method": "tools/call",
                "params": {
                    "name": "search_memories",
                    "arguments": {
                        "query": "ペット"
                    }
                }
            }
        },
        {
            "name": "犬で検索 (直接マッチテスト)",
            "request": {
                "jsonrpc": "2.0",
                "id": 3,
                "method": "tools/call",
                "params": {
                    "name": "search_memories",
                    "arguments": {
                        "query": "犬"
                    }
                }
            }
        }
    ]
    
    for test in test_requests:
        print(f"\n🔍 {test['name']}")
        print("-" * 40)
        
        try:
            # MCPサーバにJSON-RPCリクエストを送信
            process = subprocess.Popen(
                ["go", "run", "./cmd/mory"],
                stdin=subprocess.PIPE,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                text=True,
                cwd="/Users/yast/git/mory"
            )
            
            # リクエストを送信
            request_json = json.dumps(test['request'])
            stdout, stderr = process.communicate(input=request_json, timeout=10)
            
            if stdout:
                try:
                    response = json.loads(stdout.strip())
                    if 'result' in response:
                        result = response['result']
                        if 'content' in result:
                            content = result['content']
                            if isinstance(content, list) and len(content) > 0:
                                # 結果解析
                                text_content = content[0].get('text', '')
                                if 'memories found' in text_content or '件の記憶' in text_content:
                                    print(f"✅ 検索成功: {text_content[:100]}...")
                                else:
                                    print(f"⚠️ 結果: {text_content[:100]}...")
                            else:
                                print(f"❌ 予期しない結果形式: {result}")
                        else:
                            print(f"❌ contentフィールドなし: {result}")
                    else:
                        print(f"❌ エラーレスポンス: {response}")
                except json.JSONDecodeError:
                    print(f"❌ JSON解析エラー: {stdout[:100]}...")
            else:
                print(f"❌ レスポンスなし。stderr: {stderr[:100]}...")
                
        except subprocess.TimeoutExpired:
            print("❌ タイムアウト")
            if process:
                process.kill()
        except Exception as e:
            print(f"❌ エラー: {e}")
    
    print(f"\n🏁 MCPサーバ検索テスト完了")

if __name__ == "__main__":
    test_mcp_search()