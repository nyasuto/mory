#!/usr/bin/env python3
"""
Keyword Search Test - Test keyword search functionality without semantic search
"""

import json
import subprocess
import asyncio

class KeywordSearchTester:
    def __init__(self):
        self.server_process = None
    
    async def start_server(self):
        """Start the MCP server process"""
        self.server_process = await asyncio.create_subprocess_exec(
            "go", "run", "./cmd/mory",
            stdin=asyncio.subprocess.PIPE,
            stdout=asyncio.subprocess.PIPE,
            stderr=asyncio.subprocess.PIPE,
            cwd="/Users/yast/git/mory"
        )
        
        await asyncio.sleep(1)
        return self.server_process.returncode is None
    
    async def send_request(self, request):
        """Send a JSON-RPC request to the MCP server"""
        request_json = json.dumps(request) + "\n"
        
        try:
            self.server_process.stdin.write(request_json.encode())
            await self.server_process.stdin.drain()
            
            response_line = await asyncio.wait_for(
                self.server_process.stdout.readline(),
                timeout=5.0
            )
            
            if response_line:
                response_str = response_line.decode().strip()
                if response_str:
                    return json.loads(response_str)
        except Exception:
            pass
        
        return None
    
    async def initialize_server(self):
        """Initialize MCP server"""
        init_request = {
            "jsonrpc": "2.0",
            "id": 1,
            "method": "initialize",
            "params": {
                "protocolVersion": "2024-11-05",
                "capabilities": {"tools": {}},
                "clientInfo": {"name": "keyword-test-client", "version": "1.0.0"}
            }
        }
        
        response = await self.send_request(init_request)
        if response and "result" in response:
            await self.send_request({
                "jsonrpc": "2.0",
                "method": "notifications/initialized"
            })
            return True
        return False
    
    async def test_keyword_searches(self):
        """Test keyword search functionality"""
        print("\n🔍 Testing keyword search functionality...")
        
        # Test searches that should find content
        test_cases = [
            {
                "query": "犬",
                "expected": "Should find the philosophy content mentioning dogs",
                "should_find": True
            },
            {
                "query": "人生",
                "expected": "Should find life philosophy content",
                "should_find": True
            },
            {
                "query": "技術",
                "expected": "Should find technology-related memories",
                "should_find": True
            },
            {
                "query": "哲学",
                "expected": "Should find philosophy content",
                "should_find": True
            },
            {
                "query": "nonexistent_term_xyz",
                "expected": "Should not find anything",
                "should_find": False
            }
        ]
        
        for test in test_cases:
            print(f"\n   🔎 Testing '{test['query']}'")
            print(f"      Expected: {test['expected']}")
            
            search_request = {
                "jsonrpc": "2.0",
                "id": 2,
                "method": "tools/call",
                "params": {
                    "name": "search_memories",
                    "arguments": {"query": test["query"]}
                }
            }
            
            response = await self.send_request(search_request)
            if response and "result" in response:
                content = response["result"]["content"][0]["text"]
                
                # Parse results
                if "found: 0" in content or "No memories found" in content:
                    found_count = 0
                else:
                    lines = content.split('\n')
                    for line in lines:
                        if "found:" in line:
                            try:
                                count_part = line.split("found: ")[1].split(",")[0]
                                found_count = int(count_part)
                                break
                            except:
                                found_count = 0
                
                if test["should_find"] and found_count > 0:
                    print(f"      ✅ Found {found_count} results (as expected)")
                elif not test["should_find"] and found_count == 0:
                    print(f"      ✅ No results found (as expected)")
                elif test["should_find"] and found_count == 0:
                    print(f"      ❌ Expected results but found none")
                else:
                    print(f"      ⚠️ Found {found_count} results (unexpected)")
                    
                # Show first result snippet if found
                if found_count > 0:
                    lines = content.split('\n')
                    in_results = False
                    for line in lines:
                        if line.strip() and line[0].isdigit() and ":" in line:
                            result_text = line.split(":", 1)[1].strip()
                            print(f"      📝 Sample: {result_text[:60]}...")
                            break
    
    async def stop_server(self):
        """Stop the MCP server"""
        if self.server_process:
            try:
                self.server_process.terminate()
                await asyncio.wait_for(self.server_process.wait(), timeout=5.0)
            except asyncio.TimeoutError:
                self.server_process.kill()
                await self.server_process.wait()
    
    async def run_test(self):
        """Run keyword search test"""
        print("🔤 Keyword Search Test (No Semantic Search)")
        print("=" * 45)
        
        try:
            if not await self.start_server():
                print("❌ Failed to start server")
                return
            
            if not await self.initialize_server():
                print("❌ Failed to initialize server")
                return
            
            await self.test_keyword_searches()
            
            print(f"\n🏁 Keyword search test completed")
            
        finally:
            await self.stop_server()

async def main():
    tester = KeywordSearchTester()
    await tester.run_test()

if __name__ == "__main__":
    asyncio.run(main())