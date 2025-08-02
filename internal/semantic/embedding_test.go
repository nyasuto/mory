package semantic

import (
	"crypto/sha256"
	"fmt"
	"testing"
	"time"
)

func TestNewOpenAIEmbeddingService(t *testing.T) {
	apiKey := "test-api-key"
	model := "text-embedding-3-small"

	service := NewOpenAIEmbeddingService(apiKey, model)

	if service == nil {
		t.Fatal("NewOpenAIEmbeddingService returned nil")
	}

	if service.model != model {
		t.Errorf("Expected model %s, got %s", model, service.model)
	}

	if service.maxBatchSize != 100 {
		t.Errorf("Expected maxBatchSize 100, got %d", service.maxBatchSize)
	}

	if service.cache == nil {
		t.Error("Cache should not be nil")
	}

	if service.cache.ttl != 24*time.Hour {
		t.Errorf("Expected cache TTL 24h, got %v", service.cache.ttl)
	}
}

func TestGenerateTextHash(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", fmt.Sprintf("%x", sha256.Sum256([]byte("hello")))},
		{"", fmt.Sprintf("%x", sha256.Sum256([]byte("")))},
		{"test text for hashing", fmt.Sprintf("%x", sha256.Sum256([]byte("test text for hashing")))},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := GenerateTextHash(tt.input)
			if result != tt.expected {
				t.Errorf("GenerateTextHash(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestEmbeddingCache_Operations(t *testing.T) {
	cache := &EmbeddingCache{
		cache: make(map[string]CachedEmbedding),
		ttl:   1 * time.Hour,
	}

	// Test cache miss
	key := "test-key"
	_, found := cache.get(key)
	if found {
		t.Error("Expected cache miss, but got hit")
	}

	// Test cache set and hit
	embedding := []float32{0.1, 0.2, 0.3}
	cache.set(key, embedding)

	cached, found := cache.get(key)
	if !found {
		t.Error("Expected cache hit, but got miss")
	}

	if len(cached.Embedding) != len(embedding) {
		t.Errorf("Expected embedding length %d, got %d", len(embedding), len(cached.Embedding))
	}

	for i, v := range embedding {
		if cached.Embedding[i] != v {
			t.Errorf("Expected embedding[%d] = %f, got %f", i, v, cached.Embedding[i])
		}
	}
}

func TestEmbeddingCache_TTL(t *testing.T) {
	cache := &EmbeddingCache{
		cache: make(map[string]CachedEmbedding),
		ttl:   1 * time.Millisecond, // Very short TTL for testing
	}

	key := "test-key"
	embedding := []float32{0.1, 0.2, 0.3}

	cache.set(key, embedding)

	// Should hit immediately
	_, found := cache.get(key)
	if !found {
		t.Error("Expected cache hit immediately after set")
	}

	// Wait for TTL to expire
	time.Sleep(2 * time.Millisecond)

	// Should miss after TTL
	_, found = cache.get(key)
	if found {
		t.Error("Expected cache miss after TTL expiry")
	}
}

func TestOpenAIEmbeddingService_generateCacheKey(t *testing.T) {
	service := NewOpenAIEmbeddingService("test-key", "test-model")

	tests := []struct {
		text string
	}{
		{"hello"},
		{""},
		{"test text"},
		{"HELLO"},     // Test case normalization
		{"  hello  "}, // Test trimming
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			result := service.generateCacheKey(tt.text)
			// Just verify that the result is a valid hex string of expected length (64 chars for SHA256)
			if len(result) != 64 {
				t.Errorf("generateCacheKey(%q) returned hash of length %d, expected 64", tt.text, len(result))
			}
			// Verify it's valid hex
			for _, c := range result {
				if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
					t.Errorf("generateCacheKey(%q) returned non-hex character: %c", tt.text, c)
				}
			}
		})
	}

	// Test consistency - same input should produce same output
	key1 := service.generateCacheKey("hello")
	key2 := service.generateCacheKey("hello")
	if key1 != key2 {
		t.Errorf("generateCacheKey should be consistent: %q != %q", key1, key2)
	}

	// Test normalization
	key3 := service.generateCacheKey("HELLO")
	key4 := service.generateCacheKey("hello")
	if key3 != key4 {
		t.Errorf("generateCacheKey should normalize case: %q != %q", key3, key4)
	}
}

// Mock tests for API-dependent functions would require HTTP mocking
// These are basic structure tests to ensure the functions exist and can be called
func TestOpenAIEmbeddingService_Structure(t *testing.T) {
	service := NewOpenAIEmbeddingService("test-key", "test-model")

	// Test GetEmbedding method exists (will fail without valid API key)
	_, err := service.GetEmbedding("test")
	if err == nil {
		t.Log("GetEmbedding method is accessible")
	}

	// Test GetBatchEmbeddings method exists
	_, err = service.GetBatchEmbeddings([]string{"test1", "test2"})
	if err == nil {
		t.Log("GetBatchEmbeddings method is accessible")
	}
}
