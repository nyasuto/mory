package semantic

import (
	"fmt"
	"math"
	"sort"
	"sync"
)

// VectorStore defines the interface for vector storage and similarity search
type VectorStore interface {
	Store(id string, embedding []float32) error
	Search(queryEmbedding []float32, topK int) ([]VectorResult, error)
	Delete(id string) error
	Size() int
}

// VectorResult represents a similarity search result
type VectorResult struct {
	ID    string  `json:"id"`
	Score float64 `json:"score"` // Cosine similarity score (0.0 - 1.0)
}

// LocalVectorStore implements VectorStore using in-memory storage
type LocalVectorStore struct {
	vectors map[string][]float32
	mutex   sync.RWMutex
}

// NewLocalVectorStore creates a new local vector store
func NewLocalVectorStore() *LocalVectorStore {
	return &LocalVectorStore{
		vectors: make(map[string][]float32),
	}
}

// Store stores a vector with the given ID
func (vs *LocalVectorStore) Store(id string, embedding []float32) error {
	if len(embedding) == 0 {
		return fmt.Errorf("embedding cannot be empty")
	}

	vs.mutex.Lock()
	defer vs.mutex.Unlock()

	// Copy embedding to avoid external modifications
	embeddingCopy := make([]float32, len(embedding))
	copy(embeddingCopy, embedding)

	vs.vectors[id] = embeddingCopy
	return nil
}

// Search performs similarity search and returns top K results
func (vs *LocalVectorStore) Search(queryEmbedding []float32, topK int) ([]VectorResult, error) {
	if len(queryEmbedding) == 0 {
		return nil, fmt.Errorf("query embedding cannot be empty")
	}

	vs.mutex.RLock()
	defer vs.mutex.RUnlock()

	if len(vs.vectors) == 0 {
		return []VectorResult{}, nil
	}

	// Calculate similarities for all vectors
	var results []VectorResult
	for id, vector := range vs.vectors {
		similarity, err := CosineSimilarity(queryEmbedding, vector)
		if err != nil {
			continue // Skip invalid vectors
		}

		results = append(results, VectorResult{
			ID:    id,
			Score: similarity,
		})
	}

	// Sort by similarity score (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Return top K results
	if topK > 0 && topK < len(results) {
		results = results[:topK]
	}

	return results, nil
}

// Delete removes a vector from the store
func (vs *LocalVectorStore) Delete(id string) error {
	vs.mutex.Lock()
	defer vs.mutex.Unlock()

	delete(vs.vectors, id)
	return nil
}

// Size returns the number of vectors in the store
func (vs *LocalVectorStore) Size() int {
	vs.mutex.RLock()
	defer vs.mutex.RUnlock()

	return len(vs.vectors)
}

// CosineSimilarity calculates cosine similarity between two vectors
func CosineSimilarity(a, b []float32) (float64, error) {
	if len(a) != len(b) {
		return 0, fmt.Errorf("vector dimensions don't match: %d vs %d", len(a), len(b))
	}

	if len(a) == 0 {
		return 0, fmt.Errorf("vectors cannot be empty")
	}

	var dotProduct, normA, normB float64

	for i := 0; i < len(a); i++ {
		dotProduct += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}

	normA = math.Sqrt(normA)
	normB = math.Sqrt(normB)

	if normA == 0 || normB == 0 {
		return 0, nil
	}

	similarity := dotProduct / (normA * normB)

	// Ensure similarity is in [0, 1] range (cosine similarity can be [-1, 1])
	// For semantic search, we typically want [0, 1] where 1 is most similar
	if similarity < 0 {
		similarity = 0
	}

	return similarity, nil
}

// EuclideanDistance calculates Euclidean distance between two vectors
func EuclideanDistance(a, b []float32) (float64, error) {
	if len(a) != len(b) {
		return 0, fmt.Errorf("vector dimensions don't match: %d vs %d", len(a), len(b))
	}

	var sum float64
	for i := 0; i < len(a); i++ {
		diff := float64(a[i]) - float64(b[i])
		sum += diff * diff
	}

	return math.Sqrt(sum), nil
}

// NormalizeVector normalizes a vector to unit length
func NormalizeVector(vector []float32) []float32 {
	var norm float64
	for _, v := range vector {
		norm += float64(v) * float64(v)
	}
	norm = math.Sqrt(norm)

	if norm == 0 {
		return vector
	}

	normalized := make([]float32, len(vector))
	for i, v := range vector {
		normalized[i] = float32(float64(v) / norm)
	}

	return normalized
}
