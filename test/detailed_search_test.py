#!/usr/bin/env python3
"""
Detailed MCP Search Test - examining what memories exist and why searches aren't working
"""

import json
import subprocess
import asyncio
import os

class DetailedMCPTester:
    def __init__(self):
        self.server_process = None
    
    async def start_server(self):
        """Start the MCP server process"""
        print("🚀 Starting MCP server...")
        
        self.server_process = await asyncio.create_subprocess_exec(
            "go", "run", "./cmd/mory",
            stdin=asyncio.subprocess.PIPE,
            stdout=asyncio.subprocess.PIPE,
            stderr=asyncio.subprocess.PIPE,
            cwd="/Users/yast/git/mory"
        )
        
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
                "clientInfo": {"name": "detailed-test-client", "version": "1.0.0"}
            }
        }
        
        response = await self.send_request(init_request)
        if response and "result" in response:
            # Send initialized notification
            await self.send_request({
                "jsonrpc": "2.0",
                "method": "notifications/initialized"
            })
            return True
        return False
    
    async def list_all_memories(self):
        """List all memories to see what data exists"""
        print("\n📚 Listing all memories...")
        
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
            print("📝 Memory summary:")
            
            lines = content.split('\n')
            count = 0
            for line in lines:
                if line.strip() and (line.startswith('   ') or any(char.isdigit() for char in line[:5])):
                    if count < 10:  # Show first 10
                        print(f"   {line.strip()[:80]}...")
                    count += 1
            
            if count > 10:
                print(f"   ... and {count - 10} more memories")
            
            print(f"\n📊 Total memories found: {count}")
    
    async def test_specific_searches(self):
        """Test specific searches with detailed output"""
        print("\n🔍 Testing specific searches with detailed output...")
        
        searches = [
            "動物",  # animals
            "犬",    # dog
            "ペット", # pet
            "人生",  # life
            "哲学",  # philosophy
        ]
        
        for query in searches:
            print(f"\n   🔎 Search for '{query}':")
            
            search_request = {
                "jsonrpc": "2.0",
                "id": 3,
                "method": "tools/call",
                "params": {
                    "name": "search_memories",
                    "arguments": {"query": query}
                }
            }
            
            response = await self.send_request(search_request)
            if response and "result" in response:
                content = response["result"]["content"][0]["text"]
                
                # Extract result summary
                lines = content.split('\n')
                for line in lines:
                    if "found:" in line and "type:" in line:
                        print(f"      📈 {line.strip()}")
                        break
                
                # Check for semantic search info
                in_semantic_info = False
                for line in lines:
                    if "📊 Semantic Search Info:" in line:
                        in_semantic_info = True
                        continue
                    if in_semantic_info and line.strip().startswith("•"):
                        print(f"      🧠 {line.strip()}")
                    elif in_semantic_info and not line.strip():
                        break
            else:
                print(f"      ❌ Search failed")
    
    async def test_generate_embeddings(self):
        """Test embedding generation"""
        print("\n🧠 Testing embedding generation...")
        
        embedding_request = {
            "jsonrpc": "2.0",
            "id": 4,
            "method": "tools/call",
            "params": {
                "name": "generate_embeddings",
                "arguments": {}
            }
        }
        
        response = await self.send_request(embedding_request)
        if response and "result" in response:
            content = response["result"]["content"][0]["text"]
            print("📊 Embedding generation result:")
            
            lines = content.split('\n')
            for line in lines:
                if "Total memories:" in line or "embeddings:" in line or "Coverage:" in line:
                    print(f"   {line.strip()}")
        else:
            print("❌ Embedding generation failed")
    
    async def stop_server(self):
        """Stop the MCP server"""
        if self.server_process:
            try:
                self.server_process.terminate()
                await asyncio.wait_for(self.server_process.wait(), timeout=5.0)
            except asyncio.TimeoutError:
                self.server_process.kill()
                await self.server_process.wait()
    
    async def run_detailed_test(self):
        """Run detailed test"""
        print("🔬 Detailed MCP Search Analysis")
        print("=" * 40)
        
        try:
            if not await self.start_server():
                return
            
            if not await self.initialize_server():
                print("❌ Failed to initialize server")
                return
            
            await self.list_all_memories()
            await self.test_generate_embeddings()
            await self.test_specific_searches()
            
            print(f"\n🏁 Detailed analysis completed")
            
        finally:
            await self.stop_server()

async def main():
    tester = DetailedMCPTester()
    await tester.run_detailed_test()

if __name__ == "__main__":
    asyncio.run(main())