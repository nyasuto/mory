package semantic

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/nyasuto/mory/internal/memory"
)

// SemanticSearchEngine provides hybrid search functionality
type SemanticSearchEngine struct {
	keywordEngine    *memory.SearchEngine
	embeddingService EmbeddingService
	vectorStore      VectorStore
	hybridWeight     float64 // Weight for semantic search (0.0 - 1.0)
	threshold        float64 // Minimum similarity threshold
	enabled          bool    // Whether semantic search is enabled
}

// HybridSearchResult represents a search result with both keyword and semantic scores
type HybridSearchResult struct {
	Memory        *memory.Memory `json:"memory"`
	KeywordScore  float64        `json:"keyword_score"`
	SemanticScore float64        `json:"semantic_score"`
	FinalScore    float64        `json:"final_score"`
}

// NewSemanticSearchEngine creates a new semantic search engine
func NewSemanticSearchEngine(
	keywordEngine *memory.SearchEngine,
	embeddingService EmbeddingService,
	vectorStore VectorStore,
	hybridWeight float64,
	threshold float64,
	enabled bool,
) *SemanticSearchEngine {
	return &SemanticSearchEngine{
		keywordEngine:    keywordEngine,
		embeddingService: embeddingService,
		vectorStore:      vectorStore,
		hybridWeight:     hybridWeight,
		threshold:        threshold,
		enabled:          enabled,
	}
}

// Search performs hybrid search combining keyword and semantic search
func (se *SemanticSearchEngine) Search(query memory.SearchQuery) ([]*memory.SearchResult, error) {
	// Always perform keyword search as fallback
	keywordResults, err := se.keywordEngine.Search(query)
	if err != nil {
		return nil, fmt.Errorf("keyword search failed: %w", err)
	}

	// If semantic search is disabled or query is empty, return keyword results
	if !se.enabled || strings.TrimSpace(query.Query) == "" {
		return keywordResults, nil
	}

	// Perform semantic search
	semanticResults, err := se.performSemanticSearch(query.Query)
	if err != nil {
		log.Printf("[SemanticSearch] Semantic search failed, falling back to keyword: %v", err)
		return keywordResults, nil
	}

	// Combine and rank results
	hybridResults := se.combineResults(keywordResults, semanticResults)

	// Convert back to memory.SearchResult format
	return se.convertToMemoryResults(hybridResults), nil
}

// performSemanticSearch performs semantic similarity search
func (se *SemanticSearchEngine) performSemanticSearch(query string) ([]VectorResult, error) {
	// Generate embedding for query
	queryEmbedding, err := se.embeddingService.GetEmbedding(query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Search vector store
	results, err := se.vectorStore.Search(queryEmbedding, 50) // Get top 50 semantic matches
	if err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}

	// Filter by threshold
	var filteredResults []VectorResult
	for _, result := range results {
		if result.Score >= se.threshold {
			filteredResults = append(filteredResults, result)
		}
	}

	return filteredResults, nil
}

// combineResults combines keyword and semantic search results
func (se *SemanticSearchEngine) combineResults(
	keywordResults []*memory.SearchResult,
	semanticResults []VectorResult,
) []HybridSearchResult {
	// Create maps for quick lookup
	keywordMap := make(map[string]float64)
	semanticMap := make(map[string]float64)
	memoryMap := make(map[string]*memory.Memory)

	// Index keyword results
	for _, result := range keywordResults {
		id := result.Memory.ID
		keywordMap[id] = result.Score
		memoryMap[id] = result.Memory
	}

	// Index semantic results
	for _, result := range semanticResults {
		semanticMap[result.ID] = result.Score
	}

	// Combine all unique memory IDs
	allIDs := make(map[string]bool)
	for id := range keywordMap {
		allIDs[id] = true
	}
	for id := range semanticMap {
		allIDs[id] = true
	}

	// Create hybrid results
	var hybridResults []HybridSearchResult
	for id := range allIDs {
		memory := memoryMap[id]
		if memory == nil {
			continue // Skip if memory not found in keyword results
		}

		keywordScore := keywordMap[id]
		semanticScore := semanticMap[id]

		// Calculate hybrid score: weighted combination
		// hybridWeight is for semantic, (1-hybridWeight) is for keyword
		finalScore := (1-se.hybridWeight)*keywordScore + se.hybridWeight*semanticScore

		hybridResults = append(hybridResults, HybridSearchResult{
			Memory:        memory,
			KeywordScore:  keywordScore,
			SemanticScore: semanticScore,
			FinalScore:    finalScore,
		})
	}

	// Sort by final score (descending)
	sort.Slice(hybridResults, func(i, j int) bool {
		return hybridResults[i].FinalScore > hybridResults[j].FinalScore
	})

	return hybridResults
}

// convertToMemoryResults converts hybrid results to memory.SearchResult format
func (se *SemanticSearchEngine) convertToMemoryResults(hybridResults []HybridSearchResult) []*memory.SearchResult {
	results := make([]*memory.SearchResult, len(hybridResults))
	for i, hybrid := range hybridResults {
		results[i] = &memory.SearchResult{
			Memory: hybrid.Memory,
			Score:  hybrid.FinalScore,
		}
	}
	return results
}

// GenerateEmbedding generates and stores embedding for a memory
func (se *SemanticSearchEngine) GenerateEmbedding(mem *memory.Memory) error {
	if !se.enabled {
		return nil // Skip if semantic search is disabled
	}

	// Combine relevant fields for embedding
	text := se.buildEmbeddingText(mem)

	// Generate content hash
	textHash := GenerateTextHash(text)

	// Check if embedding is up to date
	if mem.EmbeddingHash == textHash && len(mem.Embedding) > 0 {
		return nil // Embedding is already up to date
	}

	// Generate new embedding
	embedding, err := se.embeddingService.GetEmbedding(text)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Update memory with embedding
	mem.Embedding = embedding
	mem.EmbeddingHash = textHash

	// Store in vector store
	if err := se.vectorStore.Store(mem.ID, embedding); err != nil {
		return fmt.Errorf("failed to store embedding: %w", err)
	}

	return nil
}

// buildEmbeddingText combines memory fields for embedding generation
func (se *SemanticSearchEngine) buildEmbeddingText(mem *memory.Memory) string {
	var parts []string

	if mem.Key != "" {
		parts = append(parts, mem.Key)
	}

	if mem.Value != "" {
		parts = append(parts, mem.Value)
	}

	if len(mem.Tags) > 0 {
		parts = append(parts, strings.Join(mem.Tags, " "))
	}

	return strings.Join(parts, " ")
}

// RemoveEmbedding removes embedding for a memory
func (se *SemanticSearchEngine) RemoveEmbedding(memoryID string) error {
	if !se.enabled {
		return nil
	}

	return se.vectorStore.Delete(memoryID)
}

// GetStats returns search engine statistics
func (se *SemanticSearchEngine) GetStats() map[string]interface{} {
	stats := map[string]interface{}{
		"enabled":              se.enabled,
		"hybrid_weight":        se.hybridWeight,
		"similarity_threshold": se.threshold,
	}

	if se.enabled {
		stats["vector_count"] = se.vectorStore.Size()
	}

	return stats
}
