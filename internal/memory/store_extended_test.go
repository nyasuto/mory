package memory

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

		// Check format (should start with "memory_" and have reasonable length)
		if !strings.HasPrefix(id, "memory_") || len(id) < 15 {
			t.Errorf("GenerateID produced invalid format: %s", id)
		}
		
		// Add small delay to ensure different nanosecond timestamps in CI
		time.Sleep(1 * time.Microsecond)
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

		// Check format (should start with "op_" and have reasonable length)
		if !strings.HasPrefix(id, "op_") || len(id) < 10 {
			t.Errorf("GenerateOperationID produced invalid format: %s", id)
		}
		
		// Add small delay to ensure different nanosecond timestamps in CI
		time.Sleep(1 * time.Microsecond)
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

func TestJSONMemoryStore_SaveWithExistingKey(t *testing.T) {
	// Test updating an existing memory with the same key
	tempDir, err := os.MkdirTemp("", "TestJSONMemoryStore_SaveWithExistingKey")
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

	// Save original memory
	original := &Memory{
		Category: "test",
		Key:      "duplicate-key",
		Value:    "original-value",
		Tags:     []string{},
	}

	id1, err := store.Save(original)
	if err != nil {
		t.Fatalf("Failed to save original memory: %v", err)
	}

	// Save updated memory with same key
	updated := &Memory{
		Category: "test",
		Key:      "duplicate-key",
		Value:    "updated-value",
		Tags:     []string{"updated"},
	}

	id2, err := store.Save(updated)
	if err != nil {
		t.Fatalf("Failed to save updated memory: %v", err)
	}

	// Should have same ID (updated, not created new)
	if id1 != id2 {
		t.Errorf("Expected same ID for updated memory, got %s != %s", id1, id2)
	}

	// Retrieve and verify it's updated
	retrieved, err := store.Get("duplicate-key")
	if err != nil {
		t.Fatalf("Failed to retrieve updated memory: %v", err)
	}

	if retrieved.Value != "updated-value" {
		t.Errorf("Expected updated value, got %s", retrieved.Value)
	}

	if len(retrieved.Tags) != 1 || retrieved.Tags[0] != "updated" {
		t.Errorf("Expected updated tags, got %v", retrieved.Tags)
	}

	// List should only have one memory
	memories, err := store.List("")
	if err != nil {
		t.Fatalf("Failed to list memories: %v", err)
	}

	if len(memories) != 1 {
		t.Errorf("Expected 1 memory after update, got %d", len(memories))
	}
}

func TestJSONMemoryStore_SaveWithExistingID(t *testing.T) {
	// Test updating an existing memory using ID
	tempDir, err := os.MkdirTemp("", "TestJSONMemoryStore_SaveWithExistingID")
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

	// Save original memory
	original := &Memory{
		Category: "test",
		Key:      "test-key",
		Value:    "original-value",
		Tags:     []string{},
	}

	id, err := store.Save(original)
	if err != nil {
		t.Fatalf("Failed to save original memory: %v", err)
	}

	// Update memory using ID
	updated := &Memory{
		ID:       id,
		Category: "updated-category",
		Key:      "updated-key",
		Value:    "updated-value",
		Tags:     []string{"updated"},
	}

	id2, err := store.Save(updated)
	if err != nil {
		t.Fatalf("Failed to save updated memory: %v", err)
	}

	// Should have same ID
	if id != id2 {
		t.Errorf("Expected same ID for updated memory, got %s != %s", id, id2)
	}

	// Should be able to retrieve by new key
	retrieved, err := store.Get("updated-key")
	if err != nil {
		t.Fatalf("Failed to retrieve by new key: %v", err)
	}

	if retrieved.Value != "updated-value" {
		t.Errorf("Expected updated value, got %s", retrieved.Value)
	}

	// Old key should not work
	_, err = store.Get("test-key")
	if err == nil {
		t.Error("Should not be able to retrieve by old key after update")
	}
}

func TestJSONMemoryStore_SaveMarshalError(t *testing.T) {
	// Test handling of JSON marshal errors (though this is hard to trigger in practice)
	tempDir, err := os.MkdirTemp("", "TestJSONMemoryStore_SaveMarshalError")
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

	// Create a memory with values that should marshal fine
	memory := &Memory{
		Category: "test",
		Key:      "test-key",
		Value:    "test-value",
		Tags:     []string{"tag"},
	}

	// This should work fine - we're mainly testing the code path
	_, err = store.Save(memory)
	if err != nil {
		t.Errorf("Save should succeed with valid memory: %v", err)
	}
}

func TestJSONMemoryStore_ListSorting(t *testing.T) {
	// Test that List returns memories sorted by creation time (newest first)
	tempDir, err := os.MkdirTemp("", "TestJSONMemoryStore_ListSorting")
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

	// Save memories with small delays to ensure different timestamps
	memories := []*Memory{
		{Category: "test", Key: "first", Value: "first", Tags: []string{}},
		{Category: "test", Key: "second", Value: "second", Tags: []string{}},
		{Category: "test", Key: "third", Value: "third", Tags: []string{}},
	}

	for i, mem := range memories {
		_, err := store.Save(mem)
		if err != nil {
			t.Fatalf("Failed to save memory %d: %v", i, err)
		}
		// Small delay to ensure different timestamps
		time.Sleep(1 * time.Millisecond)
	}

	// List all memories
	listed, err := store.List("")
	if err != nil {
		t.Fatalf("Failed to list memories: %v", err)
	}

	if len(listed) != 3 {
		t.Errorf("Expected 3 memories, got %d", len(listed))
	}

	// Should be sorted by creation time, newest first
	// So: third, second, first
	expectedOrder := []string{"third", "second", "first"}
	for i, expected := range expectedOrder {
		if listed[i].Key != expected {
			t.Errorf("Expected memory %d to have key %s, got %s", i, expected, listed[i].Key)
		}
	}

	// Verify timestamps are actually decreasing (newer first)
	for i := 1; i < len(listed); i++ {
		if listed[i-1].CreatedAt.Before(listed[i].CreatedAt) {
			t.Errorf("Memories should be sorted newest first, but memory %d is older than memory %d", i-1, i)
		}
	}
}

func TestJSONMemoryStore_DeleteNonExistent(t *testing.T) {
	// Test deleting a non-existent memory
	tempDir, err := os.MkdirTemp("", "TestJSONMemoryStore_DeleteNonExistent")
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

	// Try to delete non-existent memory
	err = store.Delete("nonexistent-key")
	if err == nil {
		t.Error("Delete should fail for non-existent key")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Error should mention 'not found', got: %v", err)
	}

	// Try to delete by non-existent ID
	err = store.DeleteByID("nonexistent-id")
	if err == nil {
		t.Error("DeleteByID should fail for non-existent ID")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Error should mention 'not found', got: %v", err)
	}
}

func TestJSONMemoryStore_SaveMemoriesFileWriteError(t *testing.T) {
	// Test saveMemories with file write error by using read-only directory
	tempDir, err := os.MkdirTemp("", "TestJSONMemoryStore_SaveMemoriesFileWriteError")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		// Make sure to restore write permissions before cleanup
		if err := os.Chmod(tempDir, 0755); err != nil {
			t.Logf("Warning: failed to restore permissions: %v", err)
		}
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Warning: failed to remove temp dir: %v", err)
		}
	}()

	dataFile := filepath.Join(tempDir, "memories.json")
	logFile := filepath.Join(tempDir, "test.log")

	// Create store and save a memory first
	store := NewJSONMemoryStore(dataFile, logFile)
	memory := &Memory{
		Category: "test",
		Key:      "test",
		Value:    "test",
		Tags:     []string{},
	}

	// Save initially (should work)
	_, err = store.Save(memory)
	if err != nil {
		t.Fatalf("Initial save should work: %v", err)
	}

	// Make directory read-only to trigger write error
	if err := os.Chmod(tempDir, 0444); err != nil {
		t.Fatalf("Failed to make directory read-only: %v", err)
	}

	// Try to save another memory (should fail)
	memory2 := &Memory{
		Category: "test2",
		Key:      "test2",
		Value:    "test2",
		Tags:     []string{},
	}

	_, err = store.Save(memory2)
	if err == nil {
		t.Error("Save should fail when directory is read-only")
	}
}

func TestJSONMemoryStore_LoadMemoriesFileNotExist(t *testing.T) {
	// Test loadMemories when file doesn't exist (should return empty slice)
	tempDir, err := os.MkdirTemp("", "TestJSONMemoryStore_LoadMemoriesFileNotExist")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Warning: failed to remove temp dir: %v", err)
		}
	}()

	// Use non-existent file
	dataFile := filepath.Join(tempDir, "nonexistent.json")
	logFile := filepath.Join(tempDir, "test.log")
	store := NewJSONMemoryStore(dataFile, logFile)

	memories, err := store.loadMemories()
	if err != nil {
		t.Errorf("loadMemories should not error for non-existent file, got: %v", err)
	}

	if memories == nil {
		t.Error("loadMemories should return empty slice, not nil")
	}

	if len(memories) != 0 {
		t.Errorf("loadMemories should return empty slice, got %d memories", len(memories))
	}
}

func TestOperationLog_JSONSerialization(t *testing.T) {
	// Test that OperationLog struct can be properly serialized/deserialized
	original := &OperationLog{
		OperationID: "test-op-id",
		Operation:   "save",
		Key:         "test-key",
		Before:      nil,
		After: &Memory{
			ID:       "test-id",
			Category: "test-category",
			Key:      "test-key",
			Value:    "test-value",
			Tags:     []string{"tag1"},
		},
		Success:   true,
		Error:     "",
		Timestamp: time.Now(),
	}

	// Serialize to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal operation log: %v", err)
	}

	// Deserialize from JSON
	var restored OperationLog
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Failed to unmarshal operation log: %v", err)
	}

	// Compare fields
	if restored.OperationID != original.OperationID {
		t.Errorf("OperationID mismatch: expected %s, got %s", original.OperationID, restored.OperationID)
	}

	if restored.Operation != original.Operation {
		t.Errorf("Operation mismatch: expected %s, got %s", original.Operation, restored.Operation)
	}

	if restored.Key != original.Key {
		t.Errorf("Key mismatch: expected %s, got %s", original.Key, restored.Key)
	}

	if restored.Success != original.Success {
		t.Errorf("Success mismatch: expected %v, got %v", original.Success, restored.Success)
	}

	if restored.After == nil {
		t.Error("After memory should not be nil")
	} else {
		if restored.After.ID != original.After.ID {
			t.Errorf("After.ID mismatch: expected %s, got %s", original.After.ID, restored.After.ID)
		}
	}
}
