package memory

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSearchEngine_Search(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "mory_search_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Failed to cleanup temp dir: %v", err)
		}
	}()

	// Create store
	store := NewJSONMemoryStore(
		filepath.Join(tempDir, "memories.json"),
		filepath.Join(tempDir, "operations.log"),
	)

	// Create search engine
	engine := NewSearchEngine(store)

	// Add test memories
	testMemories := []*Memory{
		{
			Category: "programming",
			Key:      "golang_tutorial",
			Value:    "Go is a programming language developed by Google",
			Tags:     []string{"go", "tutorial", "programming"},
		},
		{
			Category: "programming",
			Key:      "python_basics",
			Value:    "Python is an interpreted programming language",
			Tags:     []string{"python", "basics", "programming"},
		},
		{
			Category: "personal",
			Key:      "favorite_food",
			Value:    "I really love sushi and ramen",
			Tags:     []string{"food", "japanese"},
		},
		{
			Category: "work",
			Key:      "meeting_notes",
			Value:    "Weekly team meeting discussed Go microservices",
			Tags:     []string{"meeting", "team", "microservices"},
		},
	}

	// Save all test memories
	for _, mem := range testMemories {
		_, err := store.Save(mem)
		if err != nil {
			t.Fatalf("Failed to save test memory: %v", err)
		}
	}

	tests := []struct {
		name           string
		query          SearchQuery
		expectedCount  int
		expectContains []string // Keys that should be in results
	}{
		{
			name: "Search for programming",
			query: SearchQuery{
				Query: "programming",
			},
			expectedCount:  2,
			expectContains: []string{"golang_tutorial", "python_basics"},
		},
		{
			name: "Search for Go",
			query: SearchQuery{
				Query: "Go",
			},
			expectedCount:  2,
			expectContains: []string{"golang_tutorial", "meeting_notes"},
		},
		{
			name: "Search with category filter",
			query: SearchQuery{
				Query:    "programming",
				Category: "programming",
			},
			expectedCount:  2,
			expectContains: []string{"golang_tutorial", "python_basics"},
		},
		{
			name: "Search for Japanese food",
			query: SearchQuery{
				Query: "japanese",
			},
			expectedCount:  1,
			expectContains: []string{"favorite_food"},
		},
		{
			name: "Empty query returns all",
			query: SearchQuery{
				Query: "",
			},
			expectedCount: 4,
		},
		{
			name: "No matches",
			query: SearchQuery{
				Query: "nonexistent",
			},
			expectedCount: 0,
		},
		{
			name: "Category filter only",
			query: SearchQuery{
				Query:    "",
				Category: "personal",
			},
			expectedCount:  1,
			expectContains: []string{"favorite_food"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := engine.Search(tt.query)
			if err != nil {
				t.Errorf("Search failed: %v", err)
				return
			}

			if len(results) != tt.expectedCount {
				t.Errorf("Expected %d results, got %d", tt.expectedCount, len(results))
				return
			}

			// Check if expected keys are present
			if tt.expectContains != nil {
				resultKeys := make(map[string]bool)
				for _, result := range results {
					resultKeys[result.Memory.Key] = true
				}

				for _, expectedKey := range tt.expectContains {
					if !resultKeys[expectedKey] {
						t.Errorf("Expected result to contain key '%s'", expectedKey)
					}
				}
			}

			// Verify results are sorted by score (highest first)
			for i := 1; i < len(results); i++ {
				if results[i-1].Score < results[i].Score {
					t.Errorf("Results not properly sorted by score: %f < %f at index %d",
						results[i-1].Score, results[i].Score, i)
				}
			}
		})
	}
}

func TestSearchEngine_RelevanceScoring(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "mory_scoring_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Failed to cleanup temp dir: %v", err)
		}
	}()

	// Create store and engine
	store := NewJSONMemoryStore(
		filepath.Join(tempDir, "memories.json"),
		filepath.Join(tempDir, "operations.log"),
	)
	engine := NewSearchEngine(store)

	// Add memories with different relevance levels
	memories := []*Memory{
		{
			Category: "test",
			Key:      "exact_match",
			Value:    "some content",
			Tags:     []string{},
		},
		{
			Category: "test",
			Key:      "partial_match",
			Value:    "exact_match in value",
			Tags:     []string{},
		},
		{
			Category: "exact_match",
			Key:      "category_match",
			Value:    "different content",
			Tags:     []string{},
		},
		{
			Category: "test",
			Key:      "tag_match",
			Value:    "some content",
			Tags:     []string{"exact_match"},
		},
	}

	for _, mem := range memories {
		_, err := store.Save(mem)
		if err != nil {
			t.Fatalf("Failed to save memory: %v", err)
		}
	}

	// Search for "exact_match"
	results, err := engine.Search(SearchQuery{Query: "exact_match"})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("No results found")
	}

	// Verify highest score is the exact key match
	if results[0].Memory.Key != "exact_match" {
		t.Errorf("Expected highest score for exact key match, got: %s", results[0].Memory.Key)
	}

	// Verify all scores are within 0-1 range
	for _, result := range results {
		if result.Score < 0 || result.Score > 1 {
			t.Errorf("Score out of range [0,1]: %f for memory %s", result.Score, result.Memory.Key)
		}
	}
}

func TestSearchMemories_ConvenienceFunction(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "mory_convenience_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Failed to cleanup temp dir: %v", err)
		}
	}()

	// Create store
	store := NewJSONMemoryStore(
		filepath.Join(tempDir, "memories.json"),
		filepath.Join(tempDir, "operations.log"),
	)

	// Add test memory
	memory := &Memory{
		Category: "test",
		Key:      "test_key",
		Value:    "test content for searching",
		Tags:     []string{"test"},
	}

	_, err = store.Save(memory)
	if err != nil {
		t.Fatalf("Failed to save memory: %v", err)
	}

	// Test convenience function
	results, err := SearchMemories(store, SearchQuery{Query: "test"})
	if err != nil {
		t.Fatalf("SearchMemories failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("No results found")
	}

	if results[0].Memory.Key != "test_key" {
		t.Errorf("Expected test_key, got: %s", results[0].Memory.Key)
	}
}

func TestSearchEngine_WordBoundaryMatching(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "mory_boundary_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Failed to cleanup temp dir: %v", err)
		}
	}()

	// Create store and engine
	store := NewJSONMemoryStore(
		filepath.Join(tempDir, "memories.json"),
		filepath.Join(tempDir, "operations.log"),
	)
	engine := NewSearchEngine(store)

	// Add test memory
	memory := &Memory{
		Category: "test",
		Key:      "word_test",
		Value:    "this is a test with multiple words",
		Tags:     []string{},
	}

	_, err = store.Save(memory)
	if err != nil {
		t.Fatalf("Failed to save memory: %v", err)
	}

	// Test word boundary matching
	results, err := engine.Search(SearchQuery{Query: "test multiple"})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("No results found for word boundary test")
	}

	if results[0].Score <= 0 {
		t.Errorf("Expected positive score for word boundary match, got: %f", results[0].Score)
	}
}

func TestSearchEngine_EmptyStore(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "mory_empty_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Failed to cleanup temp dir: %v", err)
		}
	}()

	// Create empty store and engine
	store := NewJSONMemoryStore(
		filepath.Join(tempDir, "memories.json"),
		filepath.Join(tempDir, "operations.log"),
	)
	engine := NewSearchEngine(store)

	// Search empty store
	results, err := engine.Search(SearchQuery{Query: "anything"})
	if err != nil {
		t.Fatalf("Search on empty store failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 results from empty store, got: %d", len(results))
	}
}

func TestJSONMemoryStore_Search(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "mory_store_search_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Failed to cleanup temp dir: %v", err)
		}
	}()

	// Create store
	store := NewJSONMemoryStore(
		filepath.Join(tempDir, "memories.json"),
		filepath.Join(tempDir, "operations.log"),
	)

	// Add test memory
	memory := &Memory{
		Category: "test",
		Key:      "store_test",
		Value:    "testing store search functionality",
		Tags:     []string{"store", "search"},
	}

	_, err = store.Save(memory)
	if err != nil {
		t.Fatalf("Failed to save memory: %v", err)
	}

	// Test store's Search method
	results, err := store.Search(SearchQuery{Query: "store"})
	if err != nil {
		t.Fatalf("Store search failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("Store search returned no results")
	}

	if results[0].Memory.Key != "store_test" {
		t.Errorf("Expected store_test, got: %s", results[0].Memory.Key)
	}
}
