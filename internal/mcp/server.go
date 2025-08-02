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

	// generate_embeddings tool (only if semantic search is enabled)
	if s.config.Semantic != nil && s.config.Semantic.Enabled && s.config.Semantic.OpenAIAPIKey != "" {
		generateEmbeddingsTool := mcp.Tool{
			Name:        "generate_embeddings",
			Description: "Generate embeddings for all memories that don't have them (requires OpenAI API)",
			InputSchema: mcp.ToolInputSchema{
				Type:       "object",
				Properties: map[string]any{},
			},
		}
		mcpServer.AddTool(generateEmbeddingsTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return s.handleGenerateEmbeddings(ctx, request.GetArguments())
		})
	}

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

	// Get semantic search stats for informational purposes
	semanticStats := s.store.GetSemanticStats()
	isSemanticEnabled := false
	if enabled, ok := semanticStats["semantic_engine_available"].(bool); ok && enabled {
		isSemanticEnabled = true
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
		// Add search type indicator to header
		searchType := "keyword"
		if isSemanticEnabled {
			searchType = "hybrid (semantic + keyword)"
		}

		if category != "" {
			responseText = fmt.Sprintf("🔍 Search results for '%s' in category '%s' (found: %d, type: %s):\n\n",
				query, category, len(results), searchType)
		} else {
			responseText = fmt.Sprintf("🔍 Search results for '%s' (found: %d, type: %s):\n\n",
				query, len(results), searchType)
		}

		for i, result := range results {
			mem := result.Memory
			var displayName string
			if mem.Key != "" {
				displayName = mem.Key
			} else {
				displayName = mem.ID
			}

			// Enhanced score display for semantic search
			scoreDisplay := fmt.Sprintf("%.2f", result.Score)
			if isSemanticEnabled {
				scoreDisplay += " 🧠" // Brain emoji to indicate semantic scoring
			}

			responseText += fmt.Sprintf("%d. %s: %s (score: %s)\n",
				i+1, displayName, mem.Value, scoreDisplay)

			responseText += fmt.Sprintf("   📝 Category: %s\n", mem.Category)
			responseText += fmt.Sprintf("   🆔 ID: %s\n", mem.ID)
			responseText += fmt.Sprintf("   📅 Created: %s\n", mem.CreatedAt.Format("2006-01-02 15:04:05"))

			if len(mem.Tags) > 0 {
				responseText += fmt.Sprintf("   🏷️ Tags: %v\n", mem.Tags)
			}

			// Show embedding status if semantic search is available
			if isSemanticEnabled {
				hasEmbedding := len(mem.Embedding) > 0
				embeddingStatus := "❌"
				if hasEmbedding {
					embeddingStatus = "✅"
				}
				responseText += fmt.Sprintf("   🧠 Embedding: %s\n", embeddingStatus)
			}

			responseText += "\n"
		}

		// Add semantic search statistics at the end
		if isSemanticEnabled && len(semanticStats) > 0 {
			responseText += "📊 Semantic Search Info:\n"
			if totalMems, ok := semanticStats["total_memories"].(int); ok {
				responseText += fmt.Sprintf("   • Total memories: %d\n", totalMems)
			}
			if embeddedMems, ok := semanticStats["memories_with_embeddings"].(int); ok {
				responseText += fmt.Sprintf("   • With embeddings: %d\n", embeddedMems)
			}
			if coverage, ok := semanticStats["embedding_coverage"].(float64); ok {
				responseText += fmt.Sprintf("   • Coverage: %.1f%%\n", coverage*100)
			}
			if hybridWeight, ok := semanticStats["hybrid_weight"].(float64); ok {
				responseText += fmt.Sprintf("   • Semantic weight: %.1f%%\n", hybridWeight*100)
			}
		}
	}

	log.Printf("[SearchMemoriesTool] Search completed: query='%s', category='%s', results=%d", query, category, len(results))
	return mcp.NewToolResultText(responseText), nil
}

// handleGenerateEmbeddings handles the generate_embeddings tool
func (s *Server) handleGenerateEmbeddings(ctx context.Context, arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	// Check if store is initialized
	if s.store == nil {
		return mcp.NewToolResultError("memory store not initialized"), nil
	}

	// Check if semantic search is available
	semanticStats := s.store.GetSemanticStats()
	isSemanticEnabled := false
	if enabled, ok := semanticStats["semantic_engine_available"].(bool); ok && enabled {
		isSemanticEnabled = true
	}

	if !isSemanticEnabled {
		return mcp.NewToolResultError("🚫 Semantic search is not enabled. Please configure OpenAI API key and enable semantic search."), nil
	}

	log.Printf("[GenerateEmbeddings] Starting embedding generation process...")

	// Get current stats before generation
	beforeStats := s.store.GetSemanticStats()
	beforeEmbedded, _ := beforeStats["memories_with_embeddings"].(int)

	// Generate embeddings
	err := s.store.GenerateEmbeddings()
	if err != nil {
		log.Printf("[GenerateEmbeddings] ERROR: Failed to generate embeddings: %v", err)
		return mcp.NewToolResultError(fmt.Sprintf("❌ Failed to generate embeddings: %v", err)), nil
	}

	// Get stats after generation
	afterStats := s.store.GetSemanticStats()
	afterTotal, _ := afterStats["total_memories"].(int)
	afterEmbedded, _ := afterStats["memories_with_embeddings"].(int)
	coverage, _ := afterStats["embedding_coverage"].(float64)

	// Calculate newly generated embeddings
	newlyGenerated := afterEmbedded - beforeEmbedded

	// Create response
	responseText := "✅ Embedding generation completed!\n\n"
	responseText += "📊 Results:\n"
	responseText += fmt.Sprintf("   • Total memories: %d\n", afterTotal)
	responseText += fmt.Sprintf("   • Memories with embeddings: %d\n", afterEmbedded)
	responseText += fmt.Sprintf("   • Newly generated: %d\n", newlyGenerated)
	responseText += fmt.Sprintf("   • Coverage: %.1f%%\n", coverage*100)

	if newlyGenerated > 0 {
		responseText += "\n🧠 Semantic search is now more effective with the updated embeddings!"
	} else {
		responseText += "\n📝 All memories already had embeddings - no new embeddings were needed."
	}

	log.Printf("[GenerateEmbeddings] Completed: total=%d, embedded=%d, new=%d",
		afterTotal, afterEmbedded, newlyGenerated)

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
				"dry_run": map[string]any{
					"type":        "boolean",
					"description": "Preview import without saving (optional, default: false)",
				},
				"category_mapping": map[string]any{
					"type":        "object",
					"description": "Folder to category mapping (optional)",
					"additionalProperties": map[string]any{
						"type": "string",
					},
				},
				"skip_duplicates": map[string]any{
					"type":        "boolean",
					"description": "Skip files that already exist as memories (optional, default: true)",
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
		Description: "Generate Obsidian note from Mory memories with advanced template support",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"template": map[string]any{
					"type":        "string",
					"description": "Template type (daily, summary, report)",
					"enum":        []string{"daily", "summary", "report"},
				},
				"category": map[string]any{
					"type":        "string",
					"description": "Category filter for memories (optional)",
				},
				"title": map[string]any{
					"type":        "string",
					"description": "Generated note title",
				},
				"output_path": map[string]any{
					"type":        "string",
					"description": "Output file path (optional)",
				},
				"include_related": map[string]any{
					"type":        "boolean",
					"description": "Include related memories (default: true)",
				},
			},
			Required: []string{"template", "title"},
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

	// Parse additional options
	opts := &obsidian.ImportOptions{
		SkipDuplicates: true, // Default to true
	}

	if dryRun, ok := arguments["dry_run"].(bool); ok {
		opts.DryRun = dryRun
	}

	if skipDuplicates, ok := arguments["skip_duplicates"].(bool); ok {
		opts.SkipDuplicates = skipDuplicates
	}

	if categoryMapping, ok := arguments["category_mapping"].(map[string]interface{}); ok {
		opts.CategoryMapping = make(map[string]string)
		for k, v := range categoryMapping {
			if strVal, ok := v.(string); ok {
				opts.CategoryMapping[k] = strVal
			}
		}
	}

	// Use the store directly as it implements the MemoryStore interface
	importer := obsidian.NewImporter(s.config.Obsidian.VaultPath, s.store)

	var result *obsidian.ImportResult
	var err error

	switch importType {
	case "vault":
		result, err = importer.ImportVaultWithOptions(opts)
	case "category":
		category, ok := arguments["path"].(string)
		if !ok || category == "" {
			return mcp.NewToolResultError("path parameter is required for category import"), nil
		}
		result, err = importer.ImportByCategoryWithOptions(category, opts)
	case "file":
		filePath, ok := arguments["path"].(string)
		if !ok || filePath == "" {
			return mcp.NewToolResultError("path parameter is required for file import"), nil
		}
		mem, err := importer.ImportFileWithOptions(filePath, opts)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to import file: %v", err)), nil
		}

		var responseText string
		if opts.DryRun {
			responseText = "🔍 File import preview (dry run):\n"
		} else {
			responseText = "✅ Successfully imported file:\n"
		}

		responseText += fmt.Sprintf("📁 File: %s\n", filePath)
		responseText += fmt.Sprintf("📝 Memory ID: %s\n", mem.ID)
		responseText += fmt.Sprintf("📂 Category: %s\n", mem.Category)
		responseText += fmt.Sprintf("🔑 Key: %s\n", mem.Key)

		if len(mem.Tags) > 0 {
			responseText += fmt.Sprintf("🏷️ Tags: %v\n", mem.Tags)
		}

		return mcp.NewToolResultText(responseText), nil
	default:
		return mcp.NewToolResultError("invalid import_type: must be 'vault', 'category', or 'file'"), nil
	}

	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("import failed: %v", err)), nil
	}

	// Format result
	var responseText string
	if result.DryRun {
		responseText = "🔍 Import preview (dry run):\n\n"
	} else {
		responseText = "✅ Import completed successfully!\n\n"
	}

	// Show import options used
	if opts.DryRun || len(opts.CategoryMapping) > 0 || !opts.SkipDuplicates {
		responseText += "⚙️ Import options:\n"
		if opts.DryRun {
			responseText += "   - Dry run: Preview mode enabled\n"
		}
		if len(opts.CategoryMapping) > 0 {
			responseText += "   - Category mapping:\n"
			for from, to := range opts.CategoryMapping {
				responseText += fmt.Sprintf("     • %s → %s\n", from, to)
			}
		}
		if !opts.SkipDuplicates {
			responseText += "   - Duplicate handling: Allow overwrites\n"
		} else {
			responseText += "   - Duplicate handling: Skip duplicates\n"
		}
		responseText += "\n"
	}

	responseText += "📊 Statistics:\n"
	responseText += fmt.Sprintf("   - Total files found: %d\n", result.TotalFiles)
	responseText += fmt.Sprintf("   - Successfully imported: %d\n", result.ImportedFiles)
	responseText += fmt.Sprintf("   - Skipped files: %d\n", result.SkippedFiles)

	if result.DuplicateFiles > 0 {
		responseText += fmt.Sprintf("   - Duplicate files skipped: %d\n", result.DuplicateFiles)
	}

	if len(result.Errors) > 0 {
		responseText += "\n⚠️ Errors encountered:\n"
		for _, errMsg := range result.Errors {
			responseText += fmt.Sprintf("   - %s\n", errMsg)
		}
	}

	if len(result.ImportedMemories) > 0 {
		if result.DryRun {
			responseText += "\n📚 Would import:\n"
		} else {
			responseText += "\n📚 Imported memories:\n"
		}
		for i, mem := range result.ImportedMemories {
			if i >= 10 { // Limit display to first 10
				responseText += fmt.Sprintf("   ... and %d more\n", len(result.ImportedMemories)-10)
				break
			}
			responseText += fmt.Sprintf("   - %s (%s)", mem.Key, mem.Category)
			if len(mem.Tags) > 0 {
				responseText += fmt.Sprintf(" [%v]", mem.Tags)
			}
			responseText += "\n"
		}
	}

	log.Printf("[ObsidianImport] Import completed: type=%s, imported=%d, errors=%d", importType, result.ImportedFiles, len(result.Errors))
	return mcp.NewToolResultText(responseText), nil
}

// handleGenerateObsidianNote handles the generate_obsidian_note tool
func (s *Server) handleGenerateObsidianNote(ctx context.Context, arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	// Validate required parameters
	template, ok := arguments["template"].(string)
	if !ok || template == "" {
		return mcp.NewToolResultError("template parameter is required"), nil
	}

	title, ok := arguments["title"].(string)
	if !ok || title == "" {
		return mcp.NewToolResultError("title parameter is required"), nil
	}

	// Optional parameters
	category := ""
	if categoryArg, ok := arguments["category"].(string); ok {
		category = categoryArg
	}

	outputPath := ""
	if outputPathArg, ok := arguments["output_path"].(string); ok {
		outputPath = outputPathArg
	}

	includeRelated := true // default
	if includeRelatedArg, ok := arguments["include_related"].(bool); ok {
		includeRelated = includeRelatedArg
	}

	// Create note generator
	generator := obsidian.NewNoteGenerator(s.store)

	// Create generation request
	req := obsidian.GenerateRequest{
		Template:       template,
		Category:       category,
		Title:          title,
		OutputPath:     outputPath,
		IncludeRelated: includeRelated,
		CustomData:     make(map[string]string),
	}

	// Generate note
	note, err := generator.GenerateNote(req)
	if err != nil {
		log.Printf("[GenerateObsidianNote] ERROR: Failed to generate note: %v", err)
		return mcp.NewToolResultError(fmt.Sprintf("failed to generate note: %v", err)), nil
	}

	log.Printf("[GenerateObsidianNote] Note generated successfully: title=%s, template=%s, memories=%d, related=%d",
		title, template, note.MemoryCount, note.RelatedCount)

	// Create response
	responseText := "✅ Obsidianノートが生成されました！\n\n"
	responseText += fmt.Sprintf("📝 **タイトル**: %s\n", note.Title)
	responseText += fmt.Sprintf("🎨 **テンプレート**: %s\n", note.TemplateUsed)
	responseText += fmt.Sprintf("📊 **記憶数**: %d\n", note.MemoryCount)
	responseText += fmt.Sprintf("🔗 **関連記憶数**: %d\n", note.RelatedCount)
	responseText += fmt.Sprintf("📅 **生成日時**: %s\n", note.GeneratedAt.Format("2006-01-02 15:04:05"))

	if note.OutputPath != "" {
		responseText += fmt.Sprintf("📁 **出力パス**: %s\n", note.OutputPath)
	}

	responseText += "\n---\n\n"
	responseText += "**生成されたノート内容:**\n\n"
	responseText += "```markdown\n"
	responseText += note.Content
	responseText += "\n```"

	return mcp.NewToolResultText(responseText), nil
}
