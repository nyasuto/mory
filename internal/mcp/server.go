package mcp

import (
	"context"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/nyasuto/mory/internal/config"
	"github.com/nyasuto/mory/internal/memory"
	"github.com/nyasuto/mory/internal/obsidian"
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
	mcpServer.AddTool(saveMemoryTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return s.handleSaveMemory(ctx, request.GetArguments())
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
	mcpServer.AddTool(getMemoryTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return s.handleGetMemory(ctx, request.GetArguments())
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
	mcpServer.AddTool(listMemoriesTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return s.handleListMemories(ctx, request.GetArguments())
	})

	// search_memories tool
	searchMemoriesTool := mcp.Tool{
		Name:        "search_memories",
		Description: "Search memories with text queries and optional category filter",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"query": map[string]any{
					"type":        "string",
					"description": "Search query string",
				},
				"category": map[string]any{
					"type":        "string",
					"description": "Optional category filter",
				},
			},
			Required: []string{"query"},
		},
	}
	mcpServer.AddTool(searchMemoriesTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return s.handleSearchMemories(ctx, request.GetArguments())
	})

	// Obsidianツールの登録（Obsidian設定が有効な場合のみ）
	if s.config.Obsidian != nil && s.config.Obsidian.VaultPath != "" {
		s.registerObsidianTools(mcpServer)
	}
}

// handleSaveMemory handles the save_memory tool
func (s *Server) handleSaveMemory(ctx context.Context, arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	category, ok := arguments["category"].(string)
	if !ok || category == "" {
		return mcp.NewToolResultError("category parameter is required and must be a non-empty string"), nil
	}

	value, ok := arguments["value"].(string)
	if !ok || value == "" {
		return mcp.NewToolResultError("value parameter is required and must be a non-empty string"), nil
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
		log.Printf("[SaveMemoryTool] ERROR: Memory store not initialized")
		return mcp.NewToolResultError("memory store not initialized"), nil
	}

	log.Printf("[SaveMemoryTool] Attempting to save memory: category=%s, key=%s, value_length=%d",
		category, key, len(value))

	id, err := s.store.Save(mem)
	if err != nil {
		log.Printf("[SaveMemoryTool] ERROR: Failed to save memory: %v", err)
		return mcp.NewToolResultErrorFromErr("failed to save memory", err), nil
	}

	log.Printf("[SaveMemoryTool] Memory saved successfully with ID: %s", id)

	// Success response
	var responseText string
	if key != "" {
		responseText = fmt.Sprintf("✅ Memory saved successfully!\n📝 Category: %s\n🔑 Key: %s\n💾 Value: %s\n🆔 ID: %s",
			category, key, value, id)
	} else {
		responseText = fmt.Sprintf("✅ Memory saved successfully!\n📝 Category: %s\n💾 Value: %s\n🆔 ID: %s",
			category, value, id)
	}

	return mcp.NewToolResultText(responseText), nil
}

// handleGetMemory handles the get_memory tool
func (s *Server) handleGetMemory(ctx context.Context, arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	key, ok := arguments["key"].(string)
	if !ok || key == "" {
		return mcp.NewToolResultError("key parameter is required and must be a non-empty string"), nil
	}

	// Check if store is initialized
	if s.store == nil {
		return mcp.NewToolResultError("memory store not initialized"), nil
	}

	// Try to get memory by key first
	memory, err := s.store.Get(key)
	if err != nil {
		// If not found by key, try by ID
		memory, err = s.store.GetByID(key)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("❌ Memory not found with key or ID: %s", key)), nil
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

	return mcp.NewToolResultText(responseText), nil
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
		return mcp.NewToolResultError("memory store not initialized"), nil
	}

	// Get memories from store
	memories, err := s.store.List(category)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to list memories", err), nil
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

	return mcp.NewToolResultText(responseText), nil
}

// handleSearchMemories handles the search_memories tool
func (s *Server) handleSearchMemories(ctx context.Context, arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	// Query is required
	query, ok := arguments["query"].(string)
	if !ok || query == "" {
		return mcp.NewToolResultError("query parameter is required and must be a non-empty string"), nil
	}

	// Category is optional
	category := ""
	if categoryArg, ok := arguments["category"].(string); ok {
		category = categoryArg
	}

	// Check if store is initialized
	if s.store == nil {
		return mcp.NewToolResultError("memory store not initialized"), nil
	}

	// Create search query
	searchQuery := memory.SearchQuery{
		Query:    query,
		Category: category,
	}

	// Perform search
	results, err := s.store.Search(searchQuery)
	if err != nil {
		log.Printf("[SearchMemoriesTool] ERROR: Failed to search memories: %v", err)
		return mcp.NewToolResultErrorFromErr("failed to search memories", err), nil
	}

	// Format response
	var responseText string
	if len(results) == 0 {
		if category != "" {
			responseText = fmt.Sprintf("🔍 No memories found for query '%s' in category '%s'", query, category)
		} else {
			responseText = fmt.Sprintf("🔍 No memories found for query '%s'", query)
		}
	} else {
		if category != "" {
			responseText = fmt.Sprintf("🔍 Search results for '%s' in category '%s' (found: %d):\n\n", query, category, len(results))
		} else {
			responseText = fmt.Sprintf("🔍 Search results for '%s' (found: %d):\n\n", query, len(results))
		}

		for i, result := range results {
			mem := result.Memory
			var displayName string
			if mem.Key != "" {
				displayName = mem.Key
			} else {
				displayName = mem.ID
			}

			responseText += fmt.Sprintf("%d. %s: %s (score: %.2f)\n",
				i+1, displayName, mem.Value, result.Score)

			responseText += fmt.Sprintf("   📝 Category: %s\n", mem.Category)
			responseText += fmt.Sprintf("   🆔 ID: %s\n", mem.ID)
			responseText += fmt.Sprintf("   📅 Created: %s\n", mem.CreatedAt.Format("2006-01-02 15:04:05"))

			if len(mem.Tags) > 0 {
				responseText += fmt.Sprintf("   🏷️ Tags: %v\n", mem.Tags)
			}

			responseText += "\n"
		}
	}

	log.Printf("[SearchMemoriesTool] Search completed: query='%s', category='%s', results=%d", query, category, len(results))
	return mcp.NewToolResultText(responseText), nil
}

// registerObsidianTools registers Obsidian-related MCP tools
func (s *Server) registerObsidianTools(mcpServer *server.MCPServer) {
	// obsidian_import tool
	importTool := mcp.Tool{
		Name:        "obsidian_import",
		Description: "Import Obsidian notes into Mory memory storage",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"import_type": map[string]any{
					"type":        "string",
					"description": "Type of import: 'vault' (all files), 'category' (specific folder), or 'file' (single file)",
					"enum":        []string{"vault", "category", "file"},
				},
				"path": map[string]any{
					"type":        "string",
					"description": "Specific file path (for 'file' type) or category name (for 'category' type)",
				},
			},
			Required: []string{"import_type"},
		},
	}
	mcpServer.AddTool(importTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return s.handleObsidianImport(ctx, request.GetArguments())
	})

	// generate_obsidian_note tool
	generateTool := mcp.Tool{
		Name:        "generate_obsidian_note",
		Description: "Generate an Obsidian note from memories in a specific category",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"category": map[string]any{
					"type":        "string",
					"description": "Category of memories to include in the note",
				},
				"title": map[string]any{
					"type":        "string",
					"description": "Title for the generated note",
				},
				"template": map[string]any{
					"type":        "string",
					"description": "Note template: 'summary', 'diary', or 'list'",
					"enum":        []string{"summary", "diary", "list"},
				},
			},
			Required: []string{"category", "title"},
		},
	}
	mcpServer.AddTool(generateTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return s.handleGenerateObsidianNote(ctx, request.GetArguments())
	})
}

// handleObsidianImport handles the obsidian_import tool
func (s *Server) handleObsidianImport(ctx context.Context, arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	if s.config.Obsidian == nil || s.config.Obsidian.VaultPath == "" {
		return mcp.NewToolResultError("Obsidian vault path is not configured"), nil
	}

	importType, ok := arguments["import_type"].(string)
	if !ok {
		return mcp.NewToolResultError("import_type parameter is required"), nil
	}

	// Use the store directly as it implements the MemoryStore interface
	importer := obsidian.NewImporter(s.config.Obsidian.VaultPath, s.store)

	var result *obsidian.ImportResult
	var err error

	switch importType {
	case "vault":
		result, err = importer.ImportVault()
	case "category":
		category, ok := arguments["path"].(string)
		if !ok || category == "" {
			return mcp.NewToolResultError("path parameter is required for category import"), nil
		}
		result, err = importer.ImportByCategory(category)
	case "file":
		filePath, ok := arguments["path"].(string)
		if !ok || filePath == "" {
			return mcp.NewToolResultError("path parameter is required for file import"), nil
		}
		mem, err := importer.ImportFile(filePath)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to import file: %v", err)), nil
		}

		responseText := fmt.Sprintf("✅ Successfully imported file: %s\n", filePath)
		responseText += fmt.Sprintf("📝 Memory ID: %s\n", mem.ID)
		responseText += fmt.Sprintf("📂 Category: %s\n", mem.Category)
		responseText += fmt.Sprintf("🔑 Key: %s\n", mem.Key)

		return mcp.NewToolResultText(responseText), nil
	default:
		return mcp.NewToolResultError("invalid import_type: must be 'vault', 'category', or 'file'"), nil
	}

	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("import failed: %v", err)), nil
	}

	// Format result
	responseText := "✅ Import completed successfully!\n\n"
	responseText += "📊 Statistics:\n"
	responseText += fmt.Sprintf("   - Total files found: %d\n", result.TotalFiles)
	responseText += fmt.Sprintf("   - Successfully imported: %d\n", result.ImportedFiles)
	responseText += fmt.Sprintf("   - Skipped files: %d\n", result.SkippedFiles)

	if len(result.Errors) > 0 {
		responseText += "\n⚠️ Errors encountered:\n"
		for _, errMsg := range result.Errors {
			responseText += fmt.Sprintf("   - %s\n", errMsg)
		}
	}

	if len(result.ImportedMemories) > 0 {
		responseText += "\n📚 Imported memories:\n"
		for i, mem := range result.ImportedMemories {
			if i >= 10 { // Limit display to first 10
				responseText += fmt.Sprintf("   ... and %d more\n", len(result.ImportedMemories)-10)
				break
			}
			responseText += fmt.Sprintf("   - %s (%s)\n", mem.Key, mem.Category)
		}
	}

	log.Printf("[ObsidianImport] Import completed: type=%s, imported=%d, errors=%d", importType, result.ImportedFiles, len(result.Errors))
	return mcp.NewToolResultText(responseText), nil
}

// handleGenerateObsidianNote handles the generate_obsidian_note tool
func (s *Server) handleGenerateObsidianNote(ctx context.Context, arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	category, ok := arguments["category"].(string)
	if !ok || category == "" {
		return mcp.NewToolResultError("category parameter is required"), nil
	}

	title, ok := arguments["title"].(string)
	if !ok || title == "" {
		return mcp.NewToolResultError("title parameter is required"), nil
	}

	template := "list" // default template
	if templateArg, ok := arguments["template"].(string); ok {
		template = templateArg
	}

	// Get memories from the specified category
	memories, err := s.store.List(category)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get memories: %v", err)), nil
	}

	if len(memories) == 0 {
		return mcp.NewToolResultError(fmt.Sprintf("no memories found in category '%s'", category)), nil
	}

	// Generate note content based on template
	var noteContent string
	switch template {
	case "summary":
		noteContent = s.generateSummaryNote(title, category, memories)
	case "diary":
		noteContent = s.generateDiaryNote(title, category, memories)
	case "list":
		noteContent = s.generateListNote(title, category, memories)
	default:
		return mcp.NewToolResultError("invalid template: must be 'summary', 'diary', or 'list'"), nil
	}

	log.Printf("[GenerateObsidianNote] Generated note: title=%s, category=%s, template=%s, memories=%d", title, category, template, len(memories))
	return mcp.NewToolResultText(noteContent), nil
}

// generateSummaryNote generates a summary-style note
func (s *Server) generateSummaryNote(title, category string, memories []*memory.Memory) string {
	content := fmt.Sprintf("# %s\n\n", title)
	content += "---\n"
	content += fmt.Sprintf("category: %s\n", category)
	content += fmt.Sprintf("generated: %s\n", "{{date}}")
	content += fmt.Sprintf("tags: [summary, %s]\n", category)
	content += "---\n\n"
	content += "## Summary\n\n"
	content += fmt.Sprintf("This note contains a summary of %d memories from the '%s' category.\n\n", len(memories), category)

	for _, mem := range memories {
		content += fmt.Sprintf("### %s\n\n", mem.Key)
		content += fmt.Sprintf("%s\n\n", mem.Value)
		if len(mem.Tags) > 0 {
			content += fmt.Sprintf("*Tags: %v*\n\n", mem.Tags)
		}
	}

	return content
}

// generateDiaryNote generates a diary-style note
func (s *Server) generateDiaryNote(title, category string, memories []*memory.Memory) string {
	content := fmt.Sprintf("# %s\n\n", title)
	content += "---\n"
	content += fmt.Sprintf("category: %s\n", category)
	content += fmt.Sprintf("generated: %s\n", "{{date}}")
	content += fmt.Sprintf("tags: [diary, %s]\n", category)
	content += "---\n\n"

	for _, mem := range memories {
		content += fmt.Sprintf("## %s\n\n", mem.CreatedAt.Format("2006-01-02"))
		content += fmt.Sprintf("**%s**\n\n", mem.Key)
		content += fmt.Sprintf("%s\n\n", mem.Value)
	}

	return content
}

// generateListNote generates a list-style note
func (s *Server) generateListNote(title, category string, memories []*memory.Memory) string {
	content := fmt.Sprintf("# %s\n\n", title)
	content += "---\n"
	content += fmt.Sprintf("category: %s\n", category)
	content += fmt.Sprintf("generated: %s\n", "{{date}}")
	content += fmt.Sprintf("tags: [list, %s]\n", category)
	content += "---\n\n"
	content += fmt.Sprintf("## Memories in %s\n\n", category)

	for i, mem := range memories {
		content += fmt.Sprintf("%d. **%s** - %s\n", i+1, mem.Key, mem.Value)
		if len(mem.Tags) > 0 {
			content += fmt.Sprintf("   - Tags: %v\n", mem.Tags)
		}
		content += fmt.Sprintf("   - Created: %s\n", mem.CreatedAt.Format("2006-01-02 15:04:05"))
		content += "\n"
	}

	return content
}
