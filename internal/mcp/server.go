package mcp

import (
	"context"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nyasuto/mory/internal/config"
	"github.com/nyasuto/mory/internal/memory"
)

// Server represents the MCP server for Mory
type Server struct {
	config *config.Config
	store  memory.MemoryStore
	server *server.MCPServer
}

// NewServer creates a new MCP server instance
func NewServer(cfg *config.Config, store memory.MemoryStore) *Server {
	return &Server{
		config: cfg,
		store:  store,
	}
}

// Start starts the MCP server
func (s *Server) Start(ctx context.Context) error {
	// Create MCP server with tool capabilities
	mcpServer := server.NewMCPServer(
		"mory",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	// Register tools
	s.registerTools(mcpServer)

	s.server = mcpServer

	log.Printf("Starting Mory MCP server...")
	return server.ServeStdio(mcpServer)
}

// registerTools registers all available MCP tools
func (s *Server) registerTools(mcpServer *server.MCPServer) {
	// save_memory tool
	saveMemoryTool := mcp.Tool{
		Name:        "save_memory",
		Description: "Save a memory with category, value, and optional key",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"category": map[string]any{
					"type":        "string",
					"description": "Category for the memory",
				},
				"value": map[string]any{
					"type":        "string",
					"description": "Value to store",
				},
				"key": map[string]any{
					"type":        "string",
					"description": "Optional user-friendly alias for the memory",
				},
			},
		},
	}
	mcpServer.AddTool(saveMemoryTool, func(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
		return s.handleSaveMemory(context.Background(), arguments)
	})

	// get_memory tool
	getMemoryTool := mcp.Tool{
		Name:        "get_memory",
		Description: "Retrieve a memory by key or ID",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"key": map[string]any{
					"type":        "string",
					"description": "Memory key or ID to retrieve",
				},
			},
		},
	}
	mcpServer.AddTool(getMemoryTool, func(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
		return s.handleGetMemory(context.Background(), arguments)
	})

	// list_memories tool
	listMemoriesTool := mcp.Tool{
		Name:        "list_memories",
		Description: "List all memories or filter by category (chronologically sorted)",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"category": map[string]any{
					"type":        "string",
					"description": "Optional category filter",
				},
			},
		},
	}
	mcpServer.AddTool(listMemoriesTool, func(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
		return s.handleListMemories(context.Background(), arguments)
	})
}

// handleSaveMemory handles the save_memory tool
func (s *Server) handleSaveMemory(ctx context.Context, arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	category, ok := arguments["category"].(string)
	if !ok || category == "" {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []interface{}{
				map[string]interface{}{
					"type": "text",
					"text": "Error: category parameter is required and must be a non-empty string",
				},
			},
		}, nil
	}

	value, ok := arguments["value"].(string)
	if !ok || value == "" {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []interface{}{
				map[string]interface{}{
					"type": "text",
					"text": "Error: value parameter is required and must be a non-empty string",
				},
			},
		}, nil
	}

	// Key is optional
	key := ""
	if keyArg, ok := arguments["key"].(string); ok {
		key = keyArg
	}

	// Create memory object
	mem := &memory.Memory{
		Category: category,
		Key:      key,
		Value:    value,
		Tags:     []string{}, // Initialize empty tags
	}

	// Save memory using the store
	if s.store == nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []interface{}{
				map[string]interface{}{
					"type": "text",
					"text": "Error: memory store not initialized",
				},
			},
		}, nil
	}

	id, err := s.store.Save(mem)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []interface{}{
				map[string]interface{}{
					"type": "text",
					"text": fmt.Sprintf("Error: failed to save memory: %v", err),
				},
			},
		}, nil
	}

	// Success response
	var responseText string
	if key != "" {
		responseText = fmt.Sprintf("✅ Memory saved successfully!\n📝 Category: %s\n🔑 Key: %s\n💾 Value: %s\n🆔 ID: %s", 
			category, key, value, id)
	} else {
		responseText = fmt.Sprintf("✅ Memory saved successfully!\n📝 Category: %s\n💾 Value: %s\n🆔 ID: %s", 
			category, value, id)
	}

	return &mcp.CallToolResult{
		Content: []interface{}{
			map[string]interface{}{
				"type": "text",
				"text": responseText,
			},
		},
	}, nil
}

// handleGetMemory handles the get_memory tool
func (s *Server) handleGetMemory(ctx context.Context, arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	key, ok := arguments["key"].(string)
	if !ok || key == "" {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []interface{}{
				map[string]interface{}{
					"type": "text",
					"text": "Error: key parameter is required and must be a non-empty string",
				},
			},
		}, nil
	}

	// Check if store is initialized
	if s.store == nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []interface{}{
				map[string]interface{}{
					"type": "text",
					"text": "Error: memory store not initialized",
				},
			},
		}, nil
	}

	// Try to get memory by key first
	memory, err := s.store.Get(key)
	if err != nil {
		// If not found by key, try by ID
		memory, err = s.store.GetByID(key)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []interface{}{
					map[string]interface{}{
						"type": "text",
						"text": fmt.Sprintf("❌ Memory not found with key or ID: %s", key),
					},
				},
			}, nil
		}
	}

	// Success response with memory details
	var responseText string
	if memory.Key != "" {
		responseText = fmt.Sprintf("✅ Memory retrieved successfully!\n📝 Category: %s\n🔑 Key: %s\n💾 Value: %s\n🆔 ID: %s\n📅 Created: %s\n🔄 Updated: %s",
			memory.Category, memory.Key, memory.Value, memory.ID,
			memory.CreatedAt.Format("2006-01-02 15:04:05"),
			memory.UpdatedAt.Format("2006-01-02 15:04:05"))
	} else {
		responseText = fmt.Sprintf("✅ Memory retrieved successfully!\n📝 Category: %s\n💾 Value: %s\n🆔 ID: %s\n📅 Created: %s\n🔄 Updated: %s",
			memory.Category, memory.Value, memory.ID,
			memory.CreatedAt.Format("2006-01-02 15:04:05"),
			memory.UpdatedAt.Format("2006-01-02 15:04:05"))
	}

	// Add tags if present
	if len(memory.Tags) > 0 {
		responseText += fmt.Sprintf("\n🏷️ Tags: %v", memory.Tags)
	}

	return &mcp.CallToolResult{
		Content: []interface{}{
			map[string]interface{}{
				"type": "text",
				"text": responseText,
			},
		},
	}, nil
}

// handleListMemories handles the list_memories tool
func (s *Server) handleListMemories(ctx context.Context, arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	// Category is optional
	category := ""
	if categoryArg, ok := arguments["category"].(string); ok {
		category = categoryArg
	}

	// Check if store is initialized
	if s.store == nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []interface{}{
				map[string]interface{}{
					"type": "text",
					"text": "Error: memory store not initialized",
				},
			},
		}, nil
	}

	// Get memories from store
	memories, err := s.store.List(category)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []interface{}{
				map[string]interface{}{
					"type": "text",
					"text": fmt.Sprintf("Error: failed to list memories: %v", err),
				},
			},
		}, nil
	}

	// Format response
	var responseText string
	if len(memories) == 0 {
		if category != "" {
			responseText = fmt.Sprintf("📋 No memories found in category '%s'", category)
		} else {
			responseText = "📋 No memories stored yet"
		}
	} else {
		if category != "" {
			responseText = fmt.Sprintf("📋 Memories in category '%s' (total: %d):\n\n", category, len(memories))
		} else {
			responseText = fmt.Sprintf("📋 All stored memories (total: %d):\n\n", len(memories))
		}

		for i, mem := range memories {
			var displayName string
			if mem.Key != "" {
				displayName = mem.Key
			} else {
				displayName = mem.ID
			}
			
			responseText += fmt.Sprintf("%d. %s: %s (%s)\n", 
				i+1, displayName, mem.Value, 
				mem.CreatedAt.Format("2006-01-02 15:04:05"))
			
			if mem.Category != "" {
				responseText += fmt.Sprintf("   📝 Category: %s\n", mem.Category)
			}
			
			if len(mem.Tags) > 0 {
				responseText += fmt.Sprintf("   🏷️ Tags: %v\n", mem.Tags)
			}
			
			responseText += "\n"
		}
	}

	return &mcp.CallToolResult{
		Content: []interface{}{
			map[string]interface{}{
				"type": "text",
				"text": responseText,
			},
		},
	}, nil
}