package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/nyasuto/mory/internal/memory"
)

func main() {
	// Create temporary directory for demo
	tempDir, err := os.MkdirTemp("", "mory-sqlite-demo")
	if err != nil {
		log.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			log.Printf("Warning: failed to clean up temp dir: %v", err)
		}
	}()

	fmt.Printf("🔬 SQLite Memory Store Demo\n")
	fmt.Printf("Demo directory: %s\n\n", tempDir)

	// Initialize SQLite store
	dbPath := filepath.Join(tempDir, "demo.db")
	store, err := memory.NewSQLiteMemoryStore(dbPath)
	if err != nil {
		log.Fatalf("Failed to create SQLite store: %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			log.Printf("Warning: failed to close store: %v", err)
		}
	}()

	fmt.Println("✅ SQLite store initialized successfully")

	// Add sample memories
	memories := []*memory.Memory{
		{
			Category: "programming",
			Key:      "go_language",
			Value:    "Go is a statically typed, compiled programming language designed at Google",
			Tags:     []string{"golang", "google", "backend"},
		},
		{
			Category: "database",
			Key:      "sqlite_benefits",
			Value:    "SQLite is serverless, self-contained, and requires zero configuration",
			Tags:     []string{"sql", "embedded", "database"},
		},
		{
			Category: "ai",
			Key:      "semantic_search",
			Value:    "Semantic search uses vector embeddings to find conceptually similar content",
			Tags:     []string{"nlp", "embeddings", "search"},
		},
	}

	fmt.Printf("\n📝 Adding %d sample memories...\n", len(memories))
	for i, mem := range memories {
		id, err := store.Save(mem)
		if err != nil {
			log.Printf("Failed to save memory %d: %v", i+1, err)
			continue
		}
		fmt.Printf("  ✓ Saved memory: %s (ID: %s)\n", mem.Key, id)
	}

	// Test basic operations
	fmt.Println("\n🔍 Testing basic operations...")

	// List all memories
	allMemories, err := store.List("")
	if err != nil {
		log.Printf("Failed to list memories: %v", err)
	} else {
		fmt.Printf("  ✓ Listed %d memories\n", len(allMemories))
	}

	// List by category
	programmingMemories, err := store.List("programming")
	if err != nil {
		log.Printf("Failed to list programming memories: %v", err)
	} else {
		fmt.Printf("  ✓ Found %d programming memories\n", len(programmingMemories))
	}

	// Get by key
	goMemory, err := store.Get("go_language")
	if err != nil {
		log.Printf("Failed to get go_language: %v", err)
	} else {
		fmt.Printf("  ✓ Retrieved memory: %s\n", goMemory.Key)
	}

	// Search functionality
	fmt.Println("\n🔎 Testing search functionality...")

	searchQuery := memory.SearchQuery{Query: "programming"}
	results, err := store.Search(searchQuery)
	if err != nil {
		log.Printf("Failed to search: %v", err)
	} else {
		fmt.Printf("  ✓ Search for 'programming' found %d results\n", len(results))
		for _, result := range results {
			fmt.Printf("    - %s (score: %.2f)\n", result.Memory.Key, result.Score)
		}
	}

	// Get statistics
	fmt.Println("\n📊 Storage statistics:")
	stats := store.GetSemanticStats()
	for key, value := range stats {
		fmt.Printf("  %s: %v\n", key, value)
	}

	fmt.Println("\n🎉 Demo completed successfully!")
	fmt.Printf("SQLite database created at: %s\n", dbPath)
}
