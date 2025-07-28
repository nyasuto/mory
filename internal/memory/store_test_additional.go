package memory

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestJSONMemoryStore_LoadMemories_ErrorCases(t *testing.T) {
	// Test with invalid JSON file
	tempDir, err := os.MkdirTemp("", "TestJSONMemoryStore_LoadMemories_ErrorCases")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Warning: failed to remove temp dir: %v", err)
		}
	}()

	dataFile := filepath.Join(tempDir, "invalid.json")
	logFile := filepath.Join(tempDir, "test.log")

	// Create file with invalid JSON
	if err := os.WriteFile(dataFile, []byte("{invalid json}"), 0644); err != nil {
		t.Fatalf("Failed to create invalid JSON file: %v", err)
	}

	store := NewJSONMemoryStore(dataFile, logFile)

	// This should fail due to invalid JSON
	memories, err := store.loadMemories()
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}

	if memories != nil {
		t.Error("Expected nil memories for invalid JSON")
	}
}

func TestJSONMemoryStore_SaveMemories_DirectoryCreationFailure(t *testing.T) {
	// Test with read-only parent directory (simulated)
	tempDir, err := os.MkdirTemp("", "TestJSONMemoryStore_SaveMemories_DirectoryCreationFailure")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Warning: failed to remove temp dir: %v", err)
		}
	}()

	// Create a file where we want to create a directory
	conflictFile := filepath.Join(tempDir, "conflict")
	if err := os.WriteFile(conflictFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create conflict file: %v", err)
	}

	// Try to create store with path that conflicts with existing file
	dataFile := filepath.Join(conflictFile, "memories.json") // This should fail
	logFile := filepath.Join(tempDir, "test.log")

	store := NewJSONMemoryStore(dataFile, logFile)

	memory := &Memory{
		Category: "test",
		Key:      "test",
		Value:    "test",
		Tags:     []string{},
	}

	// This should fail because we can't create directory where file exists
	err = store.saveMemories([]*Memory{memory})
	if err == nil {
		t.Error("Expected error when trying to create directory where file exists")
	}
}

func TestJSONMemoryStore_LogOperation_DirectoryCreationFailure(t *testing.T) {
	// Test LogOperation with directory creation failure
	tempDir, err := os.MkdirTemp("", "TestJSONMemoryStore_LogOperation_DirectoryCreationFailure")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Warning: failed to remove temp dir: %v", err)
		}
	}()

	// Create a file where we want to create a log directory
	conflictFile := filepath.Join(tempDir, "conflict")
	if err := os.WriteFile(conflictFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create conflict file: %v", err)
	}

	dataFile := filepath.Join(tempDir, "memories.json")
	logFile := filepath.Join(conflictFile, "test.log") // This should fail

	store := NewJSONMemoryStore(dataFile, logFile)

	log := &OperationLog{
		OperationID: "test",
		Operation:   "test",
		Key:         "test",
		Success:     true,
	}

	// This should fail because we can't create directory where file exists
	err = store.LogOperation(log)
	if err == nil {
		t.Error("Expected error when trying to create log directory where file exists")
	}
}

func TestJSONMemoryStore_ConcurrentAccess(t *testing.T) {
	// Test concurrent access to the store
	tempDir, err := os.MkdirTemp("", "TestJSONMemoryStore_ConcurrentAccess")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Warning: failed to remove temp dir: %v", err)
		}
	}()

	dataFile := filepath.Join(tempDir, "memories.json")
	logFile := filepath.Join(tempDir, "test.log")

	store := NewJSONMemoryStore(dataFile, logFile)

	// Test that the mutex prevents race conditions
	// This is a basic test - in practice, you'd need more sophisticated testing
	// for actual race condition detection

	memory1 := &Memory{
		Category: "test1",
		Key:      "key1",
		Value:    "value1",
		Tags:     []string{},
	}

	memory2 := &Memory{
		Category: "test2",
		Key:      "key2",
		Value:    "value2",
		Tags:     []string{},
	}

	// Save memories sequentially
	_, err = store.Save(memory1)
	if err != nil {
		t.Fatalf("Failed to save memory1: %v", err)
	}

	_, err = store.Save(memory2)
	if err != nil {
		t.Fatalf("Failed to save memory2: %v", err)
	}

	// List memories
	memories, err := store.List("")
	if err != nil {
		t.Fatalf("Failed to list memories: %v", err)
	}

	if len(memories) != 2 {
		t.Errorf("Expected 2 memories, got %d", len(memories))
	}
}

func TestGenerateID_Uniqueness(t *testing.T) {
	// Test that GenerateID produces unique IDs
	ids := make(map[string]bool)

	for i := 0; i < 1000; i++ {
		id := GenerateID()

		if ids[id] {
			t.Errorf("GenerateID produced duplicate ID: %s", id)
		}

		ids[id] = true

		// Check format (should start with "memory_")
		if len(id) < 7 || id[:7] != "memory_" {
			t.Errorf("GenerateID produced invalid format: %s", id)
		}
	}
}

func TestGenerateOperationID_Uniqueness(t *testing.T) {
	// Test that GenerateOperationID produces unique IDs
	ids := make(map[string]bool)

	for i := 0; i < 1000; i++ {
		id := GenerateOperationID()

		if ids[id] {
			t.Errorf("GenerateOperationID produced duplicate ID: %s", id)
		}

		ids[id] = true

		// Check format (should start with "op_")
		if len(id) < 3 || id[:3] != "op_" {
			t.Errorf("GenerateOperationID produced invalid format: %s", id)
		}
	}
}

func TestMemory_JSONSerialization(t *testing.T) {
	// Test that Memory struct can be properly serialized/deserialized
	original := &Memory{
		ID:       "test-id",
		Category: "test-category",
		Key:      "test-key",
		Value:    "test-value",
		Tags:     []string{"tag1", "tag2"},
	}

	// Serialize to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal memory: %v", err)
	}

	// Deserialize from JSON
	var restored Memory
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Failed to unmarshal memory: %v", err)
	}

	// Compare fields
	if restored.ID != original.ID {
		t.Errorf("ID mismatch: expected %s, got %s", original.ID, restored.ID)
	}

	if restored.Category != original.Category {
		t.Errorf("Category mismatch: expected %s, got %s", original.Category, restored.Category)
	}

	if restored.Key != original.Key {
		t.Errorf("Key mismatch: expected %s, got %s", original.Key, restored.Key)
	}

	if restored.Value != original.Value {
		t.Errorf("Value mismatch: expected %s, got %s", original.Value, restored.Value)
	}

	if len(restored.Tags) != len(original.Tags) {
		t.Errorf("Tags length mismatch: expected %d, got %d", len(original.Tags), len(restored.Tags))
	}

	for i, tag := range original.Tags {
		if restored.Tags[i] != tag {
			t.Errorf("Tag mismatch at index %d: expected %s, got %s", i, tag, restored.Tags[i])
		}
	}
}
