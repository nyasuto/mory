package semantic

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
)

// EmbeddingService defines the interface for generating embeddings
type EmbeddingService interface {
	GetEmbedding(text string) ([]float32, error)
	GetBatchEmbeddings(texts []string) ([][]float32, error)
}

// OpenAIEmbeddingService implements EmbeddingService using OpenAI API
type OpenAIEmbeddingService struct {
	client       *openai.Client
	model        string
	cache        *EmbeddingCache
	maxBatchSize int
}

// EmbeddingCache provides caching for embeddings to reduce API calls
type EmbeddingCache struct {
	cache map[string]CachedEmbedding
	ttl   time.Duration
}

// CachedEmbedding represents a cached embedding with timestamp
type CachedEmbedding struct {
	Embedding []float32
	CreatedAt time.Time
}

// NewOpenAIEmbeddingService creates a new OpenAI embedding service
func NewOpenAIEmbeddingService(apiKey, model string) *OpenAIEmbeddingService {
	client := openai.NewClient(apiKey)
	cache := &EmbeddingCache{
		cache: make(map[string]CachedEmbedding),
		ttl:   24 * time.Hour, // Cache embeddings for 24 hours
	}

	return &OpenAIEmbeddingService{
		client:       client,
		model:        model,
		cache:        cache,
		maxBatchSize: 100,
	}
}

// GetEmbedding generates an embedding for a single text
func (s *OpenAIEmbeddingService) GetEmbedding(text string) ([]float32, error) {
	// Generate cache key
	cacheKey := s.generateCacheKey(text)

	// Check cache first
	if cached, exists := s.cache.get(cacheKey); exists {
		return cached.Embedding, nil
	}

	// Call OpenAI API
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := s.client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Input: []string{text},
		Model: openai.EmbeddingModel(s.model),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no embedding data returned")
	}

	embedding := resp.Data[0].Embedding

	// Cache the result
	s.cache.set(cacheKey, embedding)

	return embedding, nil
}

// GetBatchEmbeddings generates embeddings for multiple texts
func (s *OpenAIEmbeddingService) GetBatchEmbeddings(texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return [][]float32{}, nil
	}

	var results [][]float32

	// Process in batches
	for i := 0; i < len(texts); i += s.maxBatchSize {
		end := i + s.maxBatchSize
		if end > len(texts) {
			end = len(texts)
		}

		batch := texts[i:end]
		batchResults, err := s.processBatch(batch)
		if err != nil {
			return nil, err
		}

		results = append(results, batchResults...)
	}

	return results, nil
}

// processBatch processes a batch of texts
func (s *OpenAIEmbeddingService) processBatch(texts []string) ([][]float32, error) {
	// Check cache for each text
	var uncachedTexts []string
	var uncachedIndices []int
	results := make([][]float32, len(texts))

	for i, text := range texts {
		cacheKey := s.generateCacheKey(text)
		if cached, exists := s.cache.get(cacheKey); exists {
			results[i] = cached.Embedding
		} else {
			uncachedTexts = append(uncachedTexts, text)
			uncachedIndices = append(uncachedIndices, i)
		}
	}

	// If all texts are cached, return results
	if len(uncachedTexts) == 0 {
		return results, nil
	}

	// Call OpenAI API for uncached texts
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	resp, err := s.client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Input: uncachedTexts,
		Model: openai.EmbeddingModel(s.model),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create batch embeddings: %w", err)
	}

	if len(resp.Data) != len(uncachedTexts) {
		return nil, fmt.Errorf("embedding count mismatch: expected %d, got %d",
			len(uncachedTexts), len(resp.Data))
	}

	// Store results and cache them
	for i, embeddingData := range resp.Data {
		originalIndex := uncachedIndices[i]
		embedding := embeddingData.Embedding
		results[originalIndex] = embedding

		// Cache the result
		cacheKey := s.generateCacheKey(uncachedTexts[i])
		s.cache.set(cacheKey, embedding)
	}

	return results, nil
}

// generateCacheKey generates a cache key for text
func (s *OpenAIEmbeddingService) generateCacheKey(text string) string {
	normalized := strings.TrimSpace(strings.ToLower(text))
	hash := sha256.Sum256([]byte(s.model + ":" + normalized))
	return fmt.Sprintf("%x", hash)
}

// get retrieves a cached embedding if it exists and is not expired
func (c *EmbeddingCache) get(key string) (CachedEmbedding, bool) {
	cached, exists := c.cache[key]
	if !exists {
		return CachedEmbedding{}, false
	}

	// Check if expired
	if time.Since(cached.CreatedAt) > c.ttl {
		delete(c.cache, key)
		return CachedEmbedding{}, false
	}

	return cached, true
}

// set stores an embedding in the cache
func (c *EmbeddingCache) set(key string, embedding []float32) {
	c.cache[key] = CachedEmbedding{
		Embedding: embedding,
		CreatedAt: time.Now(),
	}
}

// GenerateTextHash generates a hash for content change detection
func GenerateTextHash(text string) string {
	normalized := strings.TrimSpace(text)
	hash := sha256.Sum256([]byte(normalized))
	return fmt.Sprintf("%x", hash)
}
