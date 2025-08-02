package memory

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// setupTestStore creates a temporary test store
func setupTestStore(t *testing.T) *JSONMemoryStore {
	t.Helper()

	tempDir := t.TempDir()
	dataFile := filepath.Join(tempDir, "test_memories.json")
	logFile := filepath.Join(tempDir, "test_operations.log")

	return NewJSONMemoryStore(dataFile, logFile)
}

func TestNewJSONMemoryStore(t *testing.T) {
	store := NewJSONMemoryStore("data.json", "log.json")
	if store.dataFile != "data.json" {
		t.Errorf("Expected dataFile to be 'data.json', got '%s'", store.dataFile)
	}
	if store.logFile != "log.json" {
		t.Errorf("Expected logFile to be 'log.json', got '%s'", store.logFile)
	}
}

func TestJSONMemoryStore_Save(t *testing.T) {
	store := setupTestStore(t)

	memory := &Memory{
		Category: "personal",
		Key:      "birthday",
		Value:    "1990-05-15",
		Tags:     []string{"personal", "important"},
	}

	// Test saving new memory
	id, err := store.Save(memory)
	if err != nil {
		t.Fatalf("Failed to save memory: %v", err)
	}

	if id == "" {
		t.Error("Expected non-empty ID")
	}

	if memory.ID != id {
		t.Errorf("Expected memory.ID to be '%s', got '%s'", id, memory.ID)
	}

	if memory.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}

	if memory.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt to be set")
	}
}

func TestJSONMemoryStore_SaveUpdate(t *testing.T) {
	store := setupTestStore(t)

	// Save initial memory
	memory := &Memory{
		Category: "personal",
		Key:      "birthday",
		Value:    "1990-05-15",
	}

	id1, err := store.Save(memory)
	if err != nil {
		t.Fatalf("Failed to save memory: %v", err)
	}

	// Update the same memory by key
	updatedMemory := &Memory{
		Category: "personal",
		Key:      "birthday",
		Value:    "1990-05-16", // Updated value
	}

	id2, err := store.Save(updatedMemory)
	if err != nil {
		t.Fatalf("Failed to update memory: %v", err)
	}

	if id1 != id2 {
		t.Errorf("Expected same ID after update, got '%s' vs '%s'", id1, id2)
	}

	// Verify the value was updated
	retrieved, err := store.Get("birthday")
	if err != nil {
		t.Fatalf("Failed to get memory: %v", err)
	}

	if retrieved.Value != "1990-05-16" {
		t.Errorf("Expected updated value '1990-05-16', got '%s'", retrieved.Value)
	}
}

func TestJSONMemoryStore_Get(t *testing.T) {
	store := setupTestStore(t)

	// Test getting non-existent memory
	_, err := store.Get("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent key")
	}

	// Save and retrieve memory
	memory := &Memory{
		Category: "test",
		Key:      "testkey",
		Value:    "testvalue",
	}

	_, err = store.Save(memory)
	if err != nil {
		t.Fatalf("Failed to save memory: %v", err)
	}

	retrieved, err := store.Get("testkey")
	if err != nil {
		t.Fatalf("Failed to get memory: %v", err)
	}

	if retrieved.Key != "testkey" {
		t.Errorf("Expected key 'testkey', got '%s'", retrieved.Key)
	}

	if retrieved.Value != "testvalue" {
		t.Errorf("Expected value 'testvalue', got '%s'", retrieved.Value)
	}
}

func TestJSONMemoryStore_GetByID(t *testing.T) {
	store := setupTestStore(t)

	// Test getting non-existent memory
	_, err := store.GetByID("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent ID")
	}

	// Save and retrieve by ID
	memory := &Memory{
		Category: "test",
		Value:    "testvalue",
	}

	id, err := store.Save(memory)
	if err != nil {
		t.Fatalf("Failed to save memory: %v", err)
	}

	retrieved, err := store.GetByID(id)
	if err != nil {
		t.Fatalf("Failed to get memory by ID: %v", err)
	}

	if retrieved.ID != id {
		t.Errorf("Expected ID '%s', got '%s'", id, retrieved.ID)
	}
}

func TestJSONMemoryStore_List(t *testing.T) {
	store := setupTestStore(t)

	// Test empty store
	memories, err := store.List("")
	if err != nil {
		t.Fatalf("Failed to list memories: %v", err)
	}

	if len(memories) != 0 {
		t.Errorf("Expected 0 memories, got %d", len(memories))
	}

	// Add some memories
	memory1 := &Memory{Category: "personal", Value: "value1"}
	memory2 := &Memory{Category: "work", Value: "value2"}
	memory3 := &Memory{Category: "personal", Value: "value3"}

	_, err = store.Save(memory1)
	if err != nil {
		t.Fatalf("Failed to save memory1: %v", err)
	}

	// Add small delay to ensure different timestamps
	time.Sleep(time.Millisecond)
	_, err = store.Save(memory2)
	if err != nil {
		t.Fatalf("Failed to save memory2: %v", err)
	}

	time.Sleep(time.Millisecond)
	_, err = store.Save(memory3)
	if err != nil {
		t.Fatalf("Failed to save memory3: %v", err)
	}

	// Test listing all memories
	allMemories, err := store.List("")
	if err != nil {
		t.Fatalf("Failed to list all memories: %v", err)
	}

	if len(allMemories) != 3 {
		t.Errorf("Expected 3 memories, got %d", len(allMemories))
	}

	// Test that memories are sorted by creation time (newest first)
	if allMemories[0].Value != "value3" {
		t.Errorf("Expected newest memory first, got '%s'", allMemories[0].Value)
	}

	// Test filtering by category
	personalMemories, err := store.List("personal")
	if err != nil {
		t.Fatalf("Failed to list personal memories: %v", err)
	}

	if len(personalMemories) != 2 {
		t.Errorf("Expected 2 personal memories, got %d", len(personalMemories))
	}

	workMemories, err := store.List("work")
	if err != nil {
		t.Fatalf("Failed to list work memories: %v", err)
	}

	if len(workMemories) != 1 {
		t.Errorf("Expected 1 work memory, got %d", len(workMemories))
	}
}

func TestJSONMemoryStore_Delete(t *testing.T) {
	store := setupTestStore(t)

	// Test deleting non-existent memory
	err := store.Delete("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent key")
	}

	// Save and delete memory
	memory := &Memory{
		Category: "test",
		Key:      "testkey",
		Value:    "testvalue",
	}

	_, err = store.Save(memory)
	if err != nil {
		t.Fatalf("Failed to save memory: %v", err)
	}

	// Verify it exists
	_, err = store.Get("testkey")
	if err != nil {
		t.Fatalf("Memory should exist before deletion: %v", err)
	}

	// Delete it
	err = store.Delete("testkey")
	if err != nil {
		t.Fatalf("Failed to delete memory: %v", err)
	}

	// Verify it's gone
	_, err = store.Get("testkey")
	if err == nil {
		t.Error("Memory should not exist after deletion")
	}
}

func TestJSONMemoryStore_DeleteByID(t *testing.T) {
	store := setupTestStore(t)

	// Test deleting non-existent memory
	err := store.DeleteByID("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent ID")
	}

	// Save and delete memory
	memory := &Memory{
		Category: "test",
		Value:    "testvalue",
	}

	id, err := store.Save(memory)
	if err != nil {
		t.Fatalf("Failed to save memory: %v", err)
	}

	// Verify it exists
	_, err = store.GetByID(id)
	if err != nil {
		t.Fatalf("Memory should exist before deletion: %v", err)
	}

	// Delete it
	err = store.DeleteByID(id)
	if err != nil {
		t.Fatalf("Failed to delete memory by ID: %v", err)
	}

	// Verify it's gone
	_, err = store.GetByID(id)
	if err == nil {
		t.Error("Memory should not exist after deletion")
	}
}

func TestJSONMemoryStore_LogOperation(t *testing.T) {
	store := setupTestStore(t)

	log := &OperationLog{
		Timestamp:   time.Now(),
		OperationID: GenerateOperationID(),
		Operation:   "test",
		Key:         "testkey",
		Success:     true,
	}

	err := store.LogOperation(log)
	if err != nil {
		t.Fatalf("Failed to log operation: %v", err)
	}

	// Verify log file was created and contains the log
	logData, err := os.ReadFile(store.logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	if len(logData) == 0 {
		t.Error("Expected log data to be written")
	}

	// Verify it's valid JSON
	var logEntry OperationLog
	lines := strings.Split(strings.TrimSpace(string(logData)), "\n")
	if len(lines) != 1 {
		t.Errorf("Expected 1 log line, got %d", len(lines))
	}

	err = json.Unmarshal([]byte(lines[0]), &logEntry)
	if err != nil {
		t.Fatalf("Failed to unmarshal log entry: %v", err)
	}

	if logEntry.Operation != "test" {
		t.Errorf("Expected operation 'test', got '%s'", logEntry.Operation)
	}
}

func TestJSONMemoryStore_Persistence(t *testing.T) {
	tempDir := t.TempDir()
	dataFile := filepath.Join(tempDir, "memories.json")
	logFile := filepath.Join(tempDir, "operations.log")

	// Create first store instance
	store1 := NewJSONMemoryStore(dataFile, logFile)

	memory := &Memory{
		Category: "test",
		Key:      "persistent",
		Value:    "value",
	}

	id, err := store1.Save(memory)
	if err != nil {
		t.Fatalf("Failed to save memory: %v", err)
	}

	// Create second store instance with same files
	store2 := NewJSONMemoryStore(dataFile, logFile)

	// Verify data persisted
	retrieved, err := store2.GetByID(id)
	if err != nil {
		t.Fatalf("Failed to retrieve memory from second store: %v", err)
	}

	if retrieved.Value != "value" {
		t.Errorf("Expected value 'value', got '%s'", retrieved.Value)
	}
}

func TestJSONMemoryStore_EmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	dataFile := filepath.Join(tempDir, "empty.json")
	logFile := filepath.Join(tempDir, "operations.log")

	// Create empty file
	err := os.WriteFile(dataFile, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	store := NewJSONMemoryStore(dataFile, logFile)

	// Should handle empty file gracefully
	memories, err := store.List("")
	if err != nil {
		t.Fatalf("Failed to list memories from empty file: %v", err)
	}

	if len(memories) != 0 {
		t.Errorf("Expected 0 memories from empty file, got %d", len(memories))
	}
}

func TestJSONMemoryStore_TagsInitialization(t *testing.T) {
	store := setupTestStore(t)

	// Test with nil tags
	memory := &Memory{
		Category: "test",
		Value:    "value",
		Tags:     nil,
	}

	_, err := store.Save(memory)
	if err != nil {
		t.Fatalf("Failed to save memory: %v", err)
	}

	if memory.Tags == nil {
		t.Error("Expected Tags to be initialized to empty slice")
	}

	if len(memory.Tags) != 0 {
		t.Errorf("Expected empty Tags slice, got %v", memory.Tags)
	}
}
