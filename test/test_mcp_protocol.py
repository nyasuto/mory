#!/usr/bin/env python3
"""
MCP Server Search Functionality Test
Tests semantic search functionality through proper MCP JSON-RPC protocol communication
"""

import json
import subprocess
import asyncio
import os
import sys

class MCPTester:
    def __init__(self):
        self.server_process = None
    
    async def start_server(self):
        """Start the MCP server process"""
        print("🚀 Starting MCP server...")
        
        # Start the Go MCP server
        self.server_process = await asyncio.create_subprocess_exec(
            "go", "run", "./cmd/mory",
            stdin=asyncio.subprocess.PIPE,
            stdout=asyncio.subprocess.PIPE,
            stderr=asyncio.subprocess.PIPE,
            cwd="/Users/yast/git/mory"
        )
        
        # Wait a moment for server to initialize
        await asyncio.sleep(1)
        
        if self.server_process.returncode is not None:
            stderr = await self.server_process.stderr.read()
            print(f"❌ Server failed to start: {stderr.decode()}")
            return False
        
        print("✅ MCP server started successfully")
        return True
    
    async def send_request(self, request):
        """Send a JSON-RPC request to the MCP server"""
        if not self.server_process:
            return None
        
        request_json = json.dumps(request) + "\n"
        
        try:
            # Send request
            self.server_process.stdin.write(request_json.encode())
            await self.server_process.stdin.drain()
            
            # Read response (with timeout)
            response_line = await asyncio.wait_for(
                self.server_process.stdout.readline(),
                timeout=10.0
            )
            
            if response_line:
                response_str = response_line.decode().strip()
                if response_str:
                    return json.loads(response_str)
            
        except asyncio.TimeoutError:
            print("❌ Request timeout")
        except json.JSONDecodeError as e:
            print(f"❌ JSON decode error: {e}")
        except Exception as e:
            print(f"❌ Request error: {e}")
        
        return None
    
    async def test_initialization(self):
        """Test MCP server initialization"""
        print("\n🔧 Testing MCP initialization...")
        
        # Send initialize request
        init_request = {
            "jsonrpc": "2.0",
            "id": 1,
            "method": "initialize",
            "params": {
                "protocolVersion": "2024-11-05",
                "capabilities": {
                    "tools": {}
                },
                "clientInfo": {
                    "name": "mcp-test-client",
                    "version": "1.0.0"
                }
            }
        }
        
        response = await self.send_request(init_request)
        if response and "result" in response:
            print("✅ Server initialization successful")
            
            # Send initialized notification
            initialized_request = {
                "jsonrpc": "2.0",
                "method": "notifications/initialized"
            }
            await self.send_request(initialized_request)
            return True
        else:
            print(f"❌ Initialization failed: {response}")
            return False
    
    async def test_list_tools(self):
        """Test tools/list to see available tools"""
        print("\n🛠️ Testing tools/list...")
        
        list_tools_request = {
            "jsonrpc": "2.0",
            "id": 2,
            "method": "tools/list"
        }
        
        response = await self.send_request(list_tools_request)
        if response and "result" in response and "tools" in response["result"]:
            tools = response["result"]["tools"]
            print(f"✅ Found {len(tools)} tools:")
            for tool in tools:
                print(f"   - {tool['name']}: {tool['description']}")
            return True
        else:
            print(f"❌ Failed to list tools: {response}")
            return False
    
    async def test_search_queries(self):
        """Test various search queries"""
        print("\n🔍 Testing search queries...")
        
        search_tests = [
            {
                "name": "動物で検索 (Semantic search for animals)",
                "query": "動物"
            },
            {
                "name": "ペットで検索 (Semantic search for pets)",
                "query": "ペット"
            },
            {
                "name": "犬で検索 (Direct keyword search for dogs)",
                "query": "犬"
            },
            {
                "name": "技術で検索 (Search for technology)",
                "query": "技術"
            },
            {
                "name": "人生で検索 (Search for life)",
                "query": "人生"
            }
        ]
        
        for test in search_tests:
            print(f"\n   🔎 {test['name']}")
            
            search_request = {
                "jsonrpc": "2.0",
                "id": 3,
                "method": "tools/call",
                "params": {
                    "name": "search_memories",
                    "arguments": {
                        "query": test["query"]
                    }
                }
            }
            
            response = await self.send_request(search_request)
            if response and "result" in response:
                result = response["result"]
                if "content" in result:
                    content = result["content"]
                    if isinstance(content, list) and len(content) > 0:
                        text_content = content[0].get("text", "")
                        
                        # Parse the results
                        if "found:" in text_content:
                            lines = text_content.split('\n')
                            for line in lines:
                                if "found:" in line:
                                    print(f"      ✅ {line.strip()}")
                                    break
                        elif "No memories found" in text_content or "記憶が見つかりません" in text_content:
                            print(f"      ⚠️ No results found")
                        else:
                            print(f"      📝 Result preview: {text_content[:100]}...")
                    else:
                        print(f"      ❌ Unexpected result format")
                else:
                    print(f"      ❌ No content in response: {result}")
            else:
                print(f"      ❌ Search failed: {response}")
    
    async def stop_server(self):
        """Stop the MCP server"""
        if self.server_process:
            try:
                self.server_process.terminate()
                await asyncio.wait_for(self.server_process.wait(), timeout=5.0)
            except asyncio.TimeoutError:
                self.server_process.kill()
                await self.server_process.wait()
            
            print("🛑 MCP server stopped")
    
    async def run_tests(self):
        """Run all tests"""
        print("🧪 MCP Server Search Functionality Test")
        print("=" * 50)
        
        try:
            # Start server
            if not await self.start_server():
                return
            
            # Test initialization
            if not await self.test_initialization():
                return
            
            # Test tools list
            if not await self.test_list_tools():
                return
            
            # Test search functionality
            await self.test_search_queries()
            
            print(f"\n🏁 MCP server test completed")
            
        finally:
            await self.stop_server()

async def main():
    """Main test function"""
    tester = MCPTester()
    await tester.run_tests()

if __name__ == "__main__":
    # Check if we're in the right directory
    if not os.path.exists("cmd/mory/main.go"):
        print("❌ Please run this script from the mory project root directory")
        sys.exit(1)
    
    # Run the async test
    asyncio.run(main())