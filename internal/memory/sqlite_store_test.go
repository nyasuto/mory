package memory

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"
)

func TestSQLiteMemoryStore_BasicOperations(t *testing.T) {
	// Create temporary database
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	store, err := NewSQLiteMemoryStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create SQLiteMemoryStore: %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			t.Logf("Warning: failed to close store: %v", err)
		}
	}()

	// Test Save
	memory := &Memory{
		Category: "test",
		Key:      "test_key",
		Value:    "test_value",
		Tags:     []string{"tag1", "tag2"},
	}

	id, err := store.Save(memory)
	if err != nil {
		t.Fatalf("Failed to save memory: %v", err)
	}

	if id == "" {
		t.Fatal("Expected non-empty ID")
	}

	// Test Get by key
	retrieved, err := store.Get("test_key")
	if err != nil {
		t.Fatalf("Failed to get memory by key: %v", err)
	}

	if retrieved.Key != "test_key" {
		t.Errorf("Expected key 'test_key', got '%s'", retrieved.Key)
	}
	if retrieved.Value != "test_value" {
		t.Errorf("Expected value 'test_value', got '%s'", retrieved.Value)
	}
	if len(retrieved.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(retrieved.Tags))
	}

	// Test GetByID
	retrievedByID, err := store.GetByID(id)
	if err != nil {
		t.Fatalf("Failed to get memory by ID: %v", err)
	}

	if retrievedByID.ID != id {
		t.Errorf("Expected ID '%s', got '%s'", id, retrievedByID.ID)
	}

	// Test List
	memories, err := store.List("")
	if err != nil {
		t.Fatalf("Failed to list memories: %v", err)
	}

	if len(memories) != 1 {
		t.Errorf("Expected 1 memory, got %d", len(memories))
	}

	// Test List by category
	testMemories, err := store.List("test")
	if err != nil {
		t.Fatalf("Failed to list memories by category: %v", err)
	}

	if len(testMemories) != 1 {
		t.Errorf("Expected 1 memory in 'test' category, got %d", len(testMemories))
	}

	// Test Delete
	err = store.Delete("test_key")
	if err != nil {
		t.Fatalf("Failed to delete memory: %v", err)
	}

	// Verify deletion
	_, err = store.Get("test_key")
	if err == nil {
		t.Error("Expected error when getting deleted memory")
	}

	memories, err = store.List("")
	if err != nil {
		t.Fatalf("Failed to list memories after deletion: %v", err)
	}

	if len(memories) != 0 {
		t.Errorf("Expected 0 memories after deletion, got %d", len(memories))
	}
}

func TestSQLiteMemoryStore_Search(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_search.db")

	store, err := NewSQLiteMemoryStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create SQLiteMemoryStore: %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			t.Logf("Warning: failed to close store: %v", err)
		}
	}()

	// Add test memories
	memories := []*Memory{
		{Category: "programming", Key: "go", Value: "Go is a programming language", Tags: []string{"language", "google"}},
		{Category: "programming", Key: "python", Value: "Python is versatile", Tags: []string{"language", "scripting"}},
		{Category: "food", Key: "sushi", Value: "Japanese food", Tags: []string{"japanese", "fish"}},
	}

	for _, memory := range memories {
		_, err := store.Save(memory)
		if err != nil {
			t.Fatalf("Failed to save test memory: %v", err)
		}
	}

	// Test search
	query := SearchQuery{Query: "programming"}
	results, err := store.Search(query)
	if err != nil {
		t.Fatalf("Failed to search memories: %v", err)
	}

	if len(results) < 1 {
		t.Errorf("Expected at least 1 search result, got %d", len(results))
	}

	// Test category filtering
	query = SearchQuery{Query: "language", Category: "programming"}
	results, err = store.Search(query)
	if err != nil {
		t.Fatalf("Failed to search memories with category filter: %v", err)
	}

	if len(results) < 1 {
		t.Errorf("Expected at least 1 filtered search result, got %d", len(results))
	}
}

func TestSQLiteMemoryStore_Update(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_update.db")

	store, err := NewSQLiteMemoryStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create SQLiteMemoryStore: %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			t.Logf("Warning: failed to close store: %v", err)
		}
	}()

	// Save initial memory
	memory := &Memory{
		Category: "test",
		Key:      "update_test",
		Value:    "initial value",
		Tags:     []string{"initial"},
	}

	id, err := store.Save(memory)
	if err != nil {
		t.Fatalf("Failed to save initial memory: %v", err)
	}

	// Wait a bit to ensure different timestamp
	time.Sleep(time.Millisecond * 10)

	// Update memory
	memory.Value = "updated value"
	memory.Tags = []string{"updated", "test"}

	updatedID, err := store.Save(memory)
	if err != nil {
		t.Fatalf("Failed to update memory: %v", err)
	}

	if updatedID != id {
		t.Errorf("Expected same ID after update, got different: %s vs %s", id, updatedID)
	}

	// Verify update
	retrieved, err := store.Get("update_test")
	if err != nil {
		t.Fatalf("Failed to get updated memory: %v", err)
	}

	if retrieved.Value != "updated value" {
		t.Errorf("Expected updated value 'updated value', got '%s'", retrieved.Value)
	}

	if len(retrieved.Tags) != 2 || retrieved.Tags[0] != "updated" {
		t.Errorf("Expected updated tags, got %v", retrieved.Tags)
	}

	// Verify timestamps
	if retrieved.CreatedAt.IsZero() || retrieved.UpdatedAt.IsZero() {
		t.Error("Expected non-zero timestamps")
	}

	if !retrieved.UpdatedAt.After(retrieved.CreatedAt) {
		t.Error("Expected updated_at to be after created_at")
	}
}

func TestSQLiteMemoryStore_EmbeddingOperations(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_embedding.db")

	store, err := NewSQLiteMemoryStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create SQLiteMemoryStore: %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			t.Logf("Warning: failed to close store: %v", err)
		}
	}()

	// Test memory with embedding
	embedding := []float32{0.1, 0.2, 0.3, 0.4, 0.5}
	memory := &Memory{
		Category:      "test",
		Key:           "embedding_test",
		Value:         "test with embedding",
		Embedding:     embedding,
		EmbeddingHash: "test_hash",
	}

	id, err := store.Save(memory)
	if err != nil {
		t.Fatalf("Failed to save memory with embedding: %v", err)
	}

	// Retrieve and verify embedding
	retrieved, err := store.GetByID(id)
	if err != nil {
		t.Fatalf("Failed to get memory with embedding: %v", err)
	}

	if len(retrieved.Embedding) != len(embedding) {
		t.Errorf("Expected embedding length %d, got %d", len(embedding), len(retrieved.Embedding))
	}

	for i, val := range embedding {
		if retrieved.Embedding[i] != val {
			t.Errorf("Expected embedding[%d] = %f, got %f", i, val, retrieved.Embedding[i])
		}
	}

	if retrieved.EmbeddingHash != "test_hash" {
		t.Errorf("Expected embedding hash 'test_hash', got '%s'", retrieved.EmbeddingHash)
	}
}

func TestSQLiteMemoryStore_SemanticStats(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_stats.db")

	store, err := NewSQLiteMemoryStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create SQLiteMemoryStore: %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			t.Logf("Warning: failed to close store: %v", err)
		}
	}()

	// Add memories with and without embeddings
	memoryWithEmbedding := &Memory{
		Category:  "test",
		Key:       "with_embedding",
		Value:     "has embedding",
		Embedding: []float32{0.1, 0.2, 0.3},
	}

	memoryWithoutEmbedding := &Memory{
		Category: "test",
		Key:      "without_embedding",
		Value:    "no embedding",
	}

	_, err = store.Save(memoryWithEmbedding)
	if err != nil {
		t.Fatalf("Failed to save memory with embedding: %v", err)
	}

	_, err = store.Save(memoryWithoutEmbedding)
	if err != nil {
		t.Fatalf("Failed to save memory without embedding: %v", err)
	}

	// Get stats
	stats := store.GetSemanticStats()

	if stats["storage_type"] != "sqlite" {
		t.Errorf("Expected storage_type 'sqlite', got '%v'", stats["storage_type"])
	}

	if stats["total_memories"] != 2 {
		t.Errorf("Expected total_memories 2, got %v", stats["total_memories"])
	}

	if stats["memories_with_embeddings"] != 1 {
		t.Errorf("Expected memories_with_embeddings 1, got %v", stats["memories_with_embeddings"])
	}

	if stats["embedding_coverage"] != 0.5 {
		t.Errorf("Expected embedding_coverage 0.5, got %v", stats["embedding_coverage"])
	}
}

func TestSQLiteMemoryStore_ConcurrentAccess(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_concurrent.db")

	store, err := NewSQLiteMemoryStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create SQLiteMemoryStore: %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			t.Logf("Warning: failed to close store: %v", err)
		}
	}()

	// Test concurrent writes
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(index int) {
			memory := &Memory{
				Category: "concurrent",
				Key:      fmt.Sprintf("key_%d", index),
				Value:    fmt.Sprintf("value_%d", index),
			}

			_, err := store.Save(memory)
			if err != nil {
				t.Errorf("Failed to save memory in goroutine %d: %v", index, err)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all memories were saved
	memories, err := store.List("concurrent")
	if err != nil {
		t.Fatalf("Failed to list memories after concurrent access: %v", err)
	}

	if len(memories) != 10 {
		t.Errorf("Expected 10 memories after concurrent writes, got %d", len(memories))
	}
}

// TestSQLiteMemoryStore_FTS5Support tests FTS5 functionality specifically
func TestSQLiteMemoryStore_FTS5Support(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_fts5.db")

	store, err := NewSQLiteMemoryStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create SQLite store: %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			t.Logf("Warning: failed to close store: %v", err)
		}
	}()

	// Add test memories with Japanese content
	memories := []*Memory{
		{
			ID:       "test-dog-1",
			Category: "pet",
			Key:      "my_dog",
			Value:    "私の犬はとてもかわいいです。毎日散歩しています。",
			Tags:     []string{"動物", "ペット"},
		},
		{
			ID:       "test-cat-1",
			Category: "pet",
			Key:      "my_cat",
			Value:    "猫も飼っています。とても優雅な動物です。",
			Tags:     []string{"動物", "ペット"},
		},
		{
			ID:       "test-bird-1",
			Category: "nature",
			Key:      "birds",
			Value:    "公園で鳥を観察するのが趣味です。",
			Tags:     []string{"動物", "自然"},
		},
	}

	// Save memories
	for _, mem := range memories {
		_, err := store.Save(mem)
		if err != nil {
			t.Fatalf("Failed to save memory %s: %v", mem.ID, err)
		}
	}

	// Test Japanese FTS5 search
	testCases := []struct {
		name          string
		query         string
		expectedCount int
		shouldContain []string
	}{
		{
			name:          "Search for 動物 (animals)",
			query:         "動物",
			expectedCount: 3,
			shouldContain: []string{"test-dog-1", "test-cat-1", "test-bird-1"},
		},
		{
			name:          "Search for 犬 (dog)",
			query:         "犬",
			expectedCount: 1,
			shouldContain: []string{"test-dog-1"},
		},
		{
			name:          "Search for ペット (pet)",
			query:         "ペット",
			expectedCount: 2,
			shouldContain: []string{"test-dog-1", "test-cat-1"},
		},
		{
			name:          "Search for 散歩 (walk)",
			query:         "散歩",
			expectedCount: 1,
			shouldContain: []string{"test-dog-1"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			searchQuery := SearchQuery{
				Query:    tc.query,
				Category: "",
			}

			results, err := store.Search(searchQuery)
			if err != nil {
				t.Fatalf("Search failed: %v", err)
			}

			if len(results) != tc.expectedCount {
				t.Errorf("Expected %d results, got %d", tc.expectedCount, len(results))
			}

			// Check that expected memories are included
			resultIDs := make(map[string]bool)
			for _, result := range results {
				resultIDs[result.Memory.ID] = true
			}

			for _, expectedID := range tc.shouldContain {
				if !resultIDs[expectedID] {
					t.Errorf("Expected result %s not found in search results", expectedID)
				}
			}
		})
	}
}

// TestSQLiteMemoryStore_FTS5Performance tests search performance with larger dataset
func TestSQLiteMemoryStore_FTS5Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_fts5_perf.db")

	store, err := NewSQLiteMemoryStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create SQLite store: %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			t.Logf("Warning: failed to close store: %v", err)
		}
	}()

	// Create a larger dataset
	numMemories := 1000
	for i := 0; i < numMemories; i++ {
		memory := &Memory{
			ID:       fmt.Sprintf("perf-test-%d", i),
			Category: "test",
			Key:      fmt.Sprintf("test_key_%d", i),
			Value:    fmt.Sprintf("テストデータ %d: これは性能テスト用のデータです。検索機能をテストしています。", i),
			Tags:     []string{"テスト", "性能"},
		}

		// Add some variety
		if i%10 == 0 {
			memory.Value += " 特別なキーワード"
			memory.Tags = append(memory.Tags, "特別")
		}

		_, err := store.Save(memory)
		if err != nil {
			t.Fatalf("Failed to save memory %d: %v", i, err)
		}
	}

	// Measure search performance
	start := time.Now()

	searchQuery := SearchQuery{
		Query:    "特別",
		Category: "",
	}

	results, err := store.Search(searchQuery)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	duration := time.Since(start)

	// Verify results (FTS5 has LIMIT 50, so expect min(100, 50) = 50)
	expectedResults := numMemories / 10 // Every 10th item has "特別" = 100
	maxResults := 50                    // FTS5 query has LIMIT 50
	if expectedResults > maxResults {
		expectedResults = maxResults
	}
	if len(results) != expectedResults {
		t.Errorf("Expected %d results, got %d", expectedResults, len(results))
	}

	// Performance assertion (should be fast with FTS5)
	if duration > time.Second {
		t.Errorf("Search took too long: %v (expected < 1s)", duration)
	}

	t.Logf("FTS5 search of %d records took %v", numMemories, duration)
}
