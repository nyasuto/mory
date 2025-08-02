package memory

import (
	"sort"
	"strings"
)

// SearchResult represents a search result with relevance score
type SearchResult struct {
	Memory *Memory `json:"memory"`
	Score  float64 `json:"score"` // Relevance score (0.0 - 1.0)
}

// SearchQuery represents search parameters
type SearchQuery struct {
	Query    string `json:"query"`              // Search query string
	Category string `json:"category,omitempty"` // Optional category filter
}

// SearchEngine provides search functionality for memories
type SearchEngine struct {
	store MemoryStore
}

// NewSearchEngine creates a new search engine instance
func NewSearchEngine(store MemoryStore) *SearchEngine {
	return &SearchEngine{store: store}
}

// Search performs a search across memories based on the given query
func (se *SearchEngine) Search(query SearchQuery) ([]*SearchResult, error) {
	// Get all memories for the specified category (or all if no category)
	memories, err := se.store.List(query.Category)
	if err != nil {
		return nil, err
	}

	// If no query string provided, return all memories with equal score
	if strings.TrimSpace(query.Query) == "" {
		results := make([]*SearchResult, len(memories))
		for i, memory := range memories {
			results[i] = &SearchResult{
				Memory: memory,
				Score:  1.0,
			}
		}
		return results, nil
	}

	// Search and score memories
	var results []*SearchResult
	queryLower := strings.ToLower(strings.TrimSpace(query.Query))

	for _, memory := range memories {
		score := se.calculateRelevanceScore(memory, queryLower)
		if score > 0 {
			results = append(results, &SearchResult{
				Memory: memory,
				Score:  score,
			})
		}
	}

	// Sort by relevance score (highest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results, nil
}

// calculateRelevanceScore calculates relevance score for a memory against a query
func (se *SearchEngine) calculateRelevanceScore(memory *Memory, queryLower string) float64 {
	var score float64

	// Exact matches get highest scores
	keyLower := strings.ToLower(memory.Key)
	valueLower := strings.ToLower(memory.Value)
	categoryLower := strings.ToLower(memory.Category)

	// Exact key match (highest priority)
	if keyLower == queryLower {
		score += 1.0
	} else if strings.Contains(keyLower, queryLower) {
		// Partial key match
		score += 0.8
	}

	// Exact value match
	if valueLower == queryLower {
		score += 0.9
	} else if strings.Contains(valueLower, queryLower) {
		// Partial value match
		score += 0.6
	}

	// Category match
	if categoryLower == queryLower {
		score += 0.7
	} else if strings.Contains(categoryLower, queryLower) {
		score += 0.5
	}

	// Tags match
	for _, tag := range memory.Tags {
		tagLower := strings.ToLower(tag)
		if tagLower == queryLower {
			score += 0.6
		} else if strings.Contains(tagLower, queryLower) {
			score += 0.4
		}
	}

	// Word boundary matching (bonus for matching word boundaries)
	words := strings.Fields(queryLower)
	for _, word := range words {
		if se.containsWord(keyLower, word) {
			score += 0.3
		}
		if se.containsWord(valueLower, word) {
			score += 0.2
		}
	}

	// Normalize score to 0-1 range
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// containsWord checks if text contains a word at word boundaries
func (se *SearchEngine) containsWord(text, word string) bool {
	words := strings.Fields(text)
	for _, w := range words {
		if strings.Contains(w, word) {
			return true
		}
	}
	return false
}

// SearchMemories is a convenience function that creates a search engine and performs search
func SearchMemories(store MemoryStore, query SearchQuery) ([]*SearchResult, error) {
	engine := NewSearchEngine(store)
	return engine.Search(query)
}
