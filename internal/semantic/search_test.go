package semantic

import (
	"testing"

	"github.com/nyasuto/mory/internal/memory"
)

// Mock implementations for testing
type MockEmbeddingService struct {
	embeddings map[string][]float32
}

func NewMockEmbeddingService() *MockEmbeddingService {
	return &MockEmbeddingService{
		embeddings: make(map[string][]float32),
	}
}

func (m *MockEmbeddingService) GetEmbedding(text string) ([]float32, error) {
	// Return predefined embeddings for testing
	switch text {
	case "hello world":
		return []float32{1.0, 0.0, 0.0}, nil
	case "goodbye world":
		return []float32{0.0, 1.0, 0.0}, nil
	case "test query":
		return []float32{0.7, 0.7, 0.0}, nil
	default:
		return []float32{0.5, 0.5, 0.5}, nil
	}
}

func (m *MockEmbeddingService) GetBatchEmbeddings(texts []string) ([][]float32, error) {
	results := make([][]float32, len(texts))
	for i, text := range texts {
		embedding, err := m.GetEmbedding(text)
		if err != nil {
			return nil, err
		}
		results[i] = embedding
	}
	return results, nil
}

type MockSearchEngine struct {
	memories []*memory.Memory
}

func NewMockSearchEngine() *MockSearchEngine {
	return &MockSearchEngine{
		memories: []*memory.Memory{
			{
				ID:       "1",
				Key:      "greeting",
				Value:    "hello world",
				Category: "test",
			},
			{
				ID:       "2",
				Key:      "farewell",
				Value:    "goodbye world",
				Category: "test",
			},
		},
	}
}

func (m *MockSearchEngine) Search(query memory.SearchQuery) ([]*memory.SearchResult, error) {
	var results []*memory.SearchResult
	for _, mem := range m.memories {
		if query.Query == "" || mem.Value == query.Query {
			results = append(results, &memory.SearchResult{
				Memory: mem,
				Score:  0.8,
			})
		}
	}
	return results, nil
}

func (m *MockSearchEngine) GetStore() memory.MemoryStore {
	return nil // Not needed for these tests
}

func TestNewSemanticSearchEngine(t *testing.T) {
	// Create a proper memory.SearchEngine instead of mock
	memStore := memory.NewJSONMemoryStore("test-data.json", "test-log.json")
	keywordEngine := memory.NewSearchEngine(memStore)
	embeddingService := NewMockEmbeddingService()
	vectorStore := NewLocalVectorStore()

	hybridWeight := 0.5
	threshold := 0.7
	enabled := true

	engine := NewSemanticSearchEngine(keywordEngine, embeddingService, vectorStore, hybridWeight, threshold, enabled)

	if engine == nil {
		t.Fatal("NewSemanticSearchEngine returned nil")
	}

	if engine.embeddingService != embeddingService {
		t.Error("embeddingService not set correctly")
	}

	if engine.vectorStore != vectorStore {
		t.Error("vectorStore not set correctly")
	}

	// Check parameter values
	if engine.hybridWeight != hybridWeight {
		t.Errorf("Expected hybridWeight %f, got %f", hybridWeight, engine.hybridWeight)
	}

	if engine.threshold != threshold {
		t.Errorf("Expected threshold %f, got %f", threshold, engine.threshold)
	}

	if engine.enabled != enabled {
		t.Errorf("Expected enabled %t, got %t", enabled, engine.enabled)
	}
}

func TestSemanticSearchEngine_GenerateEmbedding(t *testing.T) {
	memStore := memory.NewJSONMemoryStore("test-data.json", "test-log.json")
	keywordEngine := memory.NewSearchEngine(memStore)
	embeddingService := NewMockEmbeddingService()
	vectorStore := NewLocalVectorStore()

	engine := NewSemanticSearchEngine(keywordEngine, embeddingService, vectorStore, 0.5, 0.7, true)

	mem := &memory.Memory{
		ID:    "test-id",
		Key:   "test-key",
		Value: "hello world",
	}

	err := engine.GenerateEmbedding(mem)
	if err != nil {
		t.Errorf("GenerateEmbedding failed: %v", err)
	}

	// Check if embedding was stored
	if vectorStore.Size() != 1 {
		t.Errorf("Expected 1 embedding in vector store, got %d", vectorStore.Size())
	}
}

func TestSemanticSearchEngine_buildEmbeddingText(t *testing.T) {
	memStore := memory.NewJSONMemoryStore("test-data.json", "test-log.json")
	keywordEngine := memory.NewSearchEngine(memStore)
	embeddingService := NewMockEmbeddingService()
	vectorStore := NewLocalVectorStore()

	engine := NewSemanticSearchEngine(keywordEngine, embeddingService, vectorStore, 0.5, 0.7, true)

	tests := []struct {
		name     string
		memory   *memory.Memory
		expected string
	}{
		{
			"basic_memory",
			&memory.Memory{
				Key:      "test-key",
				Value:    "test value",
				Category: "test-category",
				Tags:     []string{"tag1", "tag2"},
			},
			"test-key test value tag1 tag2",
		},
		{
			"memory_with_empty_fields",
			&memory.Memory{
				Key:   "key-only",
				Value: "value only",
			},
			"key-only value only",
		},
		{
			"memory_with_no_tags",
			&memory.Memory{
				Key:      "test",
				Value:    "value",
				Category: "cat",
				Tags:     nil,
			},
			"test value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.buildEmbeddingText(tt.memory)
			if result != tt.expected {
				t.Errorf("buildEmbeddingText() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSemanticSearchEngine_RemoveEmbedding(t *testing.T) {
	memStore := memory.NewJSONMemoryStore("test-data.json", "test-log.json")
	keywordEngine := memory.NewSearchEngine(memStore)
	embeddingService := NewMockEmbeddingService()
	vectorStore := NewLocalVectorStore()

	engine := NewSemanticSearchEngine(keywordEngine, embeddingService, vectorStore, 0.5, 0.7, true)

	// First, add an embedding
	mem := &memory.Memory{
		ID:    "test-id",
		Value: "hello world",
	}

	err := engine.GenerateEmbedding(mem)
	if err != nil {
		t.Fatalf("GenerateEmbedding failed: %v", err)
	}

	if vectorStore.Size() != 1 {
		t.Errorf("Expected 1 embedding, got %d", vectorStore.Size())
	}

	// Now remove it
	err = engine.RemoveEmbedding("test-id")
	if err != nil {
		t.Errorf("RemoveEmbedding failed: %v", err)
	}

	if vectorStore.Size() != 0 {
		t.Errorf("Expected 0 embeddings after removal, got %d", vectorStore.Size())
	}
}

func TestSemanticSearchEngine_GetStats(t *testing.T) {
	memStore := memory.NewJSONMemoryStore("test-data.json", "test-log.json")
	keywordEngine := memory.NewSearchEngine(memStore)
	embeddingService := NewMockEmbeddingService()
	vectorStore := NewLocalVectorStore()

	engine := NewSemanticSearchEngine(keywordEngine, embeddingService, vectorStore, 0.5, 0.7, true)

	// Add some embeddings
	memories := []*memory.Memory{
		{ID: "1", Value: "hello"},
		{ID: "2", Value: "world"},
		{ID: "3", Value: "test"},
	}

	for _, mem := range memories {
		err := engine.GenerateEmbedding(mem)
		if err != nil {
			t.Logf("Failed to generate embedding for %s: %v", mem.ID, err)
		}
	}

	stats := engine.GetStats()

	if stats == nil {
		t.Fatal("GetStats returned nil")
	}

	// Check if stats contain expected fields
	if _, ok := stats["enabled"]; !ok {
		t.Error("Stats should contain 'enabled' field")
	}

	if _, ok := stats["vector_count"]; !ok {
		t.Error("Stats should contain 'vector_count' field")
	}

	if _, ok := stats["similarity_threshold"]; !ok {
		t.Error("Stats should contain 'similarity_threshold' field")
	}
}

func TestSemanticSearchEngine_Search_DisabledMode(t *testing.T) {
	memStore := memory.NewJSONMemoryStore("test-data.json", "test-log.json")
	keywordEngine := memory.NewSearchEngine(memStore)
	embeddingService := NewMockEmbeddingService()
	vectorStore := NewLocalVectorStore()

	engine := NewSemanticSearchEngine(keywordEngine, embeddingService, vectorStore, 0.5, 0.7, false)

	query := memory.SearchQuery{
		Query: "test", // Use a query that would match stored memories
	}

	results, err := engine.Search(query)
	if err != nil {
		t.Errorf("Search failed: %v", err)
	}

	// When semantic search is disabled, it should not panic and should handle gracefully
	// Results may be empty since no data was stored in the test memory store
	t.Logf("Search completed with %d results when semantic search is disabled", len(results))
}

// Test structure validation without actual API calls
func TestSemanticSearchEngine_Structure(t *testing.T) {
	memStore := memory.NewJSONMemoryStore("test-data.json", "test-log.json")
	keywordEngine := memory.NewSearchEngine(memStore)
	embeddingService := NewMockEmbeddingService()
	vectorStore := NewLocalVectorStore()

	engine := NewSemanticSearchEngine(keywordEngine, embeddingService, vectorStore, 0.5, 0.7, true)

	// Test that all methods exist and are accessible
	query := memory.SearchQuery{Query: "test"}

	_, err := engine.Search(query)
	if err != nil {
		t.Logf("Search method accessible: %v", err)
	}

	_ = engine.GetStats()
	t.Log("GetStats method accessible")

	err = engine.RemoveEmbedding("test-id")
	if err != nil {
		t.Logf("RemoveEmbedding method accessible: %v", err)
	}
}
