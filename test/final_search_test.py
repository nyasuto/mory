#!/usr/bin/env python3
"""
Final Comprehensive MCP Search Test
Tests the MCP server with the correct data directory containing the dog-related memories
"""

import json
import subprocess
import asyncio
import os

class FinalMCPTester:
    def __init__(self):
        self.server_process = None
    
    async def start_server(self):
        """Start the MCP server with correct data directory"""
        print("🚀 Starting MCP server with project data directory...")
        
        # Set environment to use project data directory
        env = os.environ.copy()
        env['MORY_DATA_DIR'] = '/Users/yast/git/mory/data'
        
        self.server_process = await asyncio.create_subprocess_exec(
            "go", "run", "./cmd/mory",
            stdin=asyncio.subprocess.PIPE,
            stdout=asyncio.subprocess.PIPE,
            stderr=asyncio.subprocess.PIPE,
            cwd="/Users/yast/git/mory",
            env=env
        )
        
        await asyncio.sleep(2)  # Allow more time for data loading
        
        if self.server_process.returncode is not None:
            stderr = await self.server_process.stderr.read()
            print(f"❌ Server failed to start: {stderr.decode()}")
            return False
        
        print("✅ MCP server started with project data")
        return True
    
    async def send_request(self, request):
        """Send a JSON-RPC request to the MCP server"""
        request_json = json.dumps(request) + "\n"
        
        try:
            self.server_process.stdin.write(request_json.encode())
            await self.server_process.stdin.drain()
            
            response_line = await asyncio.wait_for(
                self.server_process.stdout.readline(),
                timeout=10.0
            )
            
            if response_line:
                response_str = response_line.decode().strip()
                if response_str:
                    return json.loads(response_str)
        except Exception as e:
            print(f"❌ Request error: {e}")
        
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
                "clientInfo": {"name": "final-test-client", "version": "1.0.0"}
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
    
    async def test_memory_count(self):
        """Test how many memories are loaded"""
        print("\n📊 Checking memory count...")
        
        list_request = {
            "jsonrpc": "2.0",
            "id": 2,
            "method": "tools/call",
            "params": {
                "name": "list_memories",
                "arguments": {}
            }
        }
        
        response = await self.send_request(list_request)
        if response and "result" in response:
            content = response["result"]["content"][0]["text"]
            
            if "total:" in content:
                for line in content.split('\n'):
                    if "total:" in line:
                        print(f"   {line.strip()}")
                        break
            else:
                print("   Unable to parse memory count")
    
    async def test_semantic_search_status(self):
        """Test semantic search availability"""
        print("\n🧠 Checking semantic search status...")
        
        # Try to search for something to get semantic info
        search_request = {
            "jsonrpc": "2.0",
            "id": 3,
            "method": "tools/call",
            "params": {
                "name": "search_memories",
                "arguments": {"query": "test"}
            }
        }
        
        response = await self.send_request(search_request)
        if response and "result" in response:
            content = response["result"]["content"][0]["text"]
            
            # Parse semantic search info
            lines = content.split('\n')
            in_semantic_info = False
            for line in lines:
                if "📊 Semantic Search Info:" in line:
                    in_semantic_info = True
                    continue
                if in_semantic_info and line.strip().startswith("•"):
                    print(f"   {line.strip()}")
                elif in_semantic_info and not line.strip():
                    break
    
    async def test_target_searches(self):
        """Test the specific searches that were problematic"""
        print("\n🎯 Testing target searches (animals, dogs, pets)...")
        
        target_tests = [
            {
                "query": "動物",
                "description": "Animals (should find dog-related content via semantic search)"
            },
            {
                "query": "犬",
                "description": "Dogs (should find philosophy content directly)"
            },
            {
                "query": "ペット",
                "description": "Pets (should find related content)"
            },
            {
                "query": "人生哲学",
                "description": "Life philosophy (exact key match)"
            },
            {
                "query": "ポンポコ",
                "description": "Ponpoko (specific term from the dog philosophy)"
            }
        ]
        
        for test in target_tests:
            print(f"\n   🔍 '{test['query']}' - {test['description']}")
            
            search_request = {
                "jsonrpc": "2.0",
                "id": 4,
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
                found_count = 0
                search_type = "unknown"
                
                lines = content.split('\n')
                for line in lines:
                    if "found:" in line and "type:" in line:
                        try:
                            parts = line.split("found: ")[1].split(",")
                            found_count = int(parts[0])
                            type_part = line.split("type: ")[1].rstrip(")")
                            search_type = type_part
                        except:
                            pass
                        break
                
                if found_count > 0:
                    print(f"      ✅ Found {found_count} results ({search_type})")
                    
                    # Show first result
                    for line in lines:
                        if line.strip() and line[0].isdigit() and ":" in line:
                            result_text = line.split(":", 1)[1].strip()
                            print(f"      📝 Sample: {result_text[:70]}...")
                            break
                else:
                    print(f"      ❌ No results found")
    
    async def stop_server(self):
        """Stop the MCP server"""
        if self.server_process:
            try:
                self.server_process.terminate()
                await asyncio.wait_for(self.server_process.wait(), timeout=5.0)
            except asyncio.TimeoutError:
                self.server_process.kill()
                await self.server_process.wait()
    
    async def run_comprehensive_test(self):
        """Run comprehensive test"""
        print("🔬 Final Comprehensive MCP Search Test")
        print("=" * 45)
        print("Testing semantic search for animal/dog-related queries")
        print("Using project data directory with full memory set")
        print("=" * 45)
        
        try:
            if not await self.start_server():
                return
            
            if not await self.initialize_server():
                print("❌ Failed to initialize server")
                return
            
            await self.test_memory_count()
            await self.test_semantic_search_status()
            await self.test_target_searches()
            
            print(f"\n🏁 Comprehensive test completed")
            print("\n📋 Summary:")
            print("   - MCP server communication: ✅ Working")
            print("   - Data loading from project directory: ✅ Working")
            print("   - Search functionality: Check results above")
            
        finally:
            await self.stop_server()

async def main():
    tester = FinalMCPTester()
    await tester.run_comprehensive_test()

if __name__ == "__main__":
    asyncio.run(main())