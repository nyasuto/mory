package semantic

import (
	"math"
	"testing"
)

func TestNewLocalVectorStore(t *testing.T) {
	store := NewLocalVectorStore()

	if store == nil {
		t.Fatal("NewLocalVectorStore returned nil")
	}

	if store.vectors == nil {
		t.Error("vectors map should be initialized")
	}

	if store.Size() != 0 {
		t.Errorf("Expected size 0, got %d", store.Size())
	}
}

func TestLocalVectorStore_Store(t *testing.T) {
	store := NewLocalVectorStore()

	id := "test-id"
	vector := []float32{0.1, 0.2, 0.3, 0.4}

	err := store.Store(id, vector)
	if err != nil {
		t.Errorf("Store failed: %v", err)
	}

	if store.Size() != 1 {
		t.Errorf("Expected size 1 after storing, got %d", store.Size())
	}

	// Test storing same ID overwrites
	newVector := []float32{0.5, 0.6, 0.7, 0.8}
	err = store.Store(id, newVector)
	if err != nil {
		t.Errorf("Store failed on overwrite: %v", err)
	}

	if store.Size() != 1 {
		t.Errorf("Expected size 1 after overwrite, got %d", store.Size())
	}
}

func TestLocalVectorStore_Search(t *testing.T) {
	store := NewLocalVectorStore()

	// Store test vectors
	vectors := map[string][]float32{
		"vec1": {1.0, 0.0, 0.0},
		"vec2": {0.0, 1.0, 0.0},
		"vec3": {0.0, 0.0, 1.0},
		"vec4": {0.707, 0.707, 0.0}, // Similar to vec1 and vec2
	}

	for id, vec := range vectors {
		err := store.Store(id, vec)
		if err != nil {
			t.Fatalf("Failed to store vector %s: %v", id, err)
		}
	}

	// Search for vec1
	query := []float32{1.0, 0.0, 0.0}
	results, err := store.Search(query, 3)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("Expected search results, got empty")
	}

	// First result should be exact match
	if results[0].ID != "vec1" {
		t.Errorf("Expected first result to be vec1, got %s", results[0].ID)
	}

	if math.Abs(results[0].Score-1.0) > 1e-6 {
		t.Errorf("Expected score 1.0 for exact match, got %f", results[0].Score)
	}
}

func TestLocalVectorStore_SearchWithThreshold(t *testing.T) {
	store := NewLocalVectorStore()

	// Store test vectors
	err := store.Store("high_sim", []float32{1.0, 0.0})
	if err != nil {
		t.Fatalf("Failed to store high_sim: %v", err)
	}
	err = store.Store("low_sim", []float32{0.0, 1.0})
	if err != nil {
		t.Fatalf("Failed to store low_sim: %v", err)
	}

	query := []float32{1.0, 0.0}

	// Search without threshold filtering (implementation dependent)
	results, err := store.Search(query, 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	// Should return results sorted by similarity
	if len(results) == 0 {
		t.Error("Expected some results")
	}

	if len(results) > 0 && results[0].ID != "high_sim" {
		t.Errorf("Expected high_sim result first, got %s", results[0].ID)
	}
}

func TestLocalVectorStore_Delete(t *testing.T) {
	store := NewLocalVectorStore()

	id := "test-id"
	vector := []float32{0.1, 0.2, 0.3}

	// Store vector
	err := store.Store(id, vector)
	if err != nil {
		t.Fatalf("Failed to store vector: %v", err)
	}
	if store.Size() != 1 {
		t.Errorf("Expected size 1 after storing, got %d", store.Size())
	}

	// Delete vector
	err = store.Delete(id)
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}

	if store.Size() != 0 {
		t.Errorf("Expected size 0 after delete, got %d", store.Size())
	}

	// Delete non-existent vector should not error
	err = store.Delete("non-existent")
	if err != nil {
		t.Errorf("Delete of non-existent vector should not error: %v", err)
	}
}

func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		a, b     []float32
		expected float64
	}{
		{"identical_vectors", []float32{1, 0, 0}, []float32{1, 0, 0}, 1.0},
		{"orthogonal_vectors", []float32{1, 0, 0}, []float32{0, 1, 0}, 0.0},
		{"opposite_vectors", []float32{1, 0, 0}, []float32{-1, 0, 0}, 0.0}, // Clamped to 0
		{"similar_vectors", []float32{1, 1, 0}, []float32{1, 0, 0}, math.Sqrt(2) / 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CosineSimilarity(tt.a, tt.b)
			if err != nil {
				t.Errorf("CosineSimilarity failed: %v", err)
				return
			}
			if math.Abs(result-tt.expected) > 1e-6 {
				t.Errorf("CosineSimilarity(%v, %v) = %f, want %f", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestCosineSimilarity_EdgeCases(t *testing.T) {
	// Zero vectors
	result, err := CosineSimilarity([]float32{0, 0, 0}, []float32{1, 0, 0})
	if err != nil {
		t.Logf("CosineSimilarity with zero vector returned error (expected): %v", err)
	} else if result != 0.0 {
		t.Errorf("CosineSimilarity with zero vector should return 0, got %f", result)
	}

	// Different lengths should return error
	_, err = CosineSimilarity([]float32{1, 0}, []float32{1, 0, 0})
	if err == nil {
		t.Error("CosineSimilarity with different length vectors should return error")
	}

	// Empty vectors
	_, err = CosineSimilarity([]float32{}, []float32{})
	if err == nil {
		t.Error("CosineSimilarity with empty vectors should return error")
	}
}

func TestEuclideanDistance(t *testing.T) {
	tests := []struct {
		name     string
		a, b     []float32
		expected float64
	}{
		{"identical_vectors", []float32{1, 0, 0}, []float32{1, 0, 0}, 0.0},
		{"unit_distance", []float32{0, 0, 0}, []float32{1, 0, 0}, 1.0},
		{"diagonal_distance", []float32{0, 0, 0}, []float32{1, 1, 0}, math.Sqrt(2)},
		{"three_d_distance", []float32{0, 0, 0}, []float32{1, 1, 1}, math.Sqrt(3)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := EuclideanDistance(tt.a, tt.b)
			if err != nil {
				t.Errorf("EuclideanDistance failed: %v", err)
				return
			}
			if math.Abs(result-tt.expected) > 1e-6 {
				t.Errorf("EuclideanDistance(%v, %v) = %f, want %f", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestNormalizeVector(t *testing.T) {
	tests := []struct {
		name     string
		input    []float32
		expected []float32
	}{
		{"unit_vector", []float32{1, 0, 0}, []float32{1, 0, 0}},
		{"normalize_vector", []float32{3, 4, 0}, []float32{0.6, 0.8, 0}},
		{"negative_values", []float32{-1, -1, 0}, []float32{-float32(math.Sqrt(2)) / 2, -float32(math.Sqrt(2)) / 2, 0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeVector(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected length %d, got %d", len(tt.expected), len(result))
				return
			}

			for i, v := range tt.expected {
				if math.Abs(float64(result[i])-float64(v)) > 1e-6 {
					t.Errorf("NormalizeVector(%v)[%d] = %f, want %f", tt.input, i, result[i], v)
				}
			}
		})
	}
}

func TestNormalizeVector_ZeroVector(t *testing.T) {
	// Zero vector normalization should handle gracefully
	result := NormalizeVector([]float32{0, 0, 0})
	for i, v := range result {
		if !math.IsNaN(float64(v)) && v != 0 {
			t.Errorf("NormalizeVector of zero vector[%d] = %f, expected NaN or 0", i, v)
		}
	}
}
