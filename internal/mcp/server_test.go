package mcp

import (
	"context"
	"strings"
	"testing"

	"github.com/nyasuto/mory/internal/config"
	"github.com/nyasuto/mory/internal/memory"
)

// MockMemoryStore implements memory.MemoryStore for testing
type MockMemoryStore struct {
	memories map[string]*memory.Memory
	idToKey  map[string]string
	lastID   string
}

func NewMockMemoryStore() *MockMemoryStore {
	return &MockMemoryStore{
		memories: make(map[string]*memory.Memory),
		idToKey:  make(map[string]string),
	}
}

func (m *MockMemoryStore) Save(mem *memory.Memory) (string, error) {
	if mem.ID == "" {
		mem.ID = memory.GenerateID()
	}
	m.lastID = mem.ID
	
	// Store by key if provided
	if mem.Key != "" {
		m.memories[mem.Key] = mem
		m.idToKey[mem.ID] = mem.Key
	} else {
		m.memories[mem.ID] = mem
	}
	
	return mem.ID, nil
}

func (m *MockMemoryStore) Get(key string) (*memory.Memory, error) {
	if mem, exists := m.memories[key]; exists {
		return mem, nil
	}
	return nil, &memory.MemoryNotFoundError{Key: key}
}

func (m *MockMemoryStore) GetByID(id string) (*memory.Memory, error) {
	// Check if ID maps to a key
	if key, exists := m.idToKey[id]; exists {
		return m.memories[key], nil
	}
	// Otherwise check if stored directly by ID
	if mem, exists := m.memories[id]; exists {
		return mem, nil
	}
	return nil, &memory.MemoryNotFoundError{Key: id}
}

func (m *MockMemoryStore) List(category string) ([]*memory.Memory, error) {
	var result []*memory.Memory
	for _, mem := range m.memories {
		if category == "" || mem.Category == category {
			result = append(result, mem)
		}
	}
	return result, nil
}

func (m *MockMemoryStore) Delete(key string) error {
	if _, exists := m.memories[key]; exists {
		delete(m.memories, key)
		return nil
	}
	return &memory.MemoryNotFoundError{Key: key}
}

func (m *MockMemoryStore) DeleteByID(id string) error {
	if key, exists := m.idToKey[id]; exists {
		delete(m.memories, key)
		delete(m.idToKey, id)
		return nil
	}
	if _, exists := m.memories[id]; exists {
		delete(m.memories, id)
		return nil
	}
	return &memory.MemoryNotFoundError{Key: id}
}

func (m *MockMemoryStore) LogOperation(log *memory.OperationLog) error {
	// Mock implementation - do nothing
	return nil
}

// Additional test cases for better coverage

func TestNewServer(t *testing.T) {
	cfg := config.DefaultConfig()
	store := NewMockMemoryStore()
	
	server := NewServer(cfg, store)
	
	if server.config != cfg {
		t.Error("Expected config to be set")
	}
	
	if server.store != store {
		t.Error("Expected store to be set")
	}
	
	if server.server != nil {
		t.Error("Expected server to be nil initially")
	}
}

func TestHandleSaveMemory_Success(t *testing.T) {
	cfg := config.DefaultConfig()
	store := NewMockMemoryStore()
	server := NewServer(cfg, store)
	
	arguments := map[string]interface{}{
		"category": "personal",
		"value":    "test value",
		"key":      "test_key",
	}
	
	result, err := server.handleSaveMemory(context.Background(), arguments)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if result.IsError {
		t.Error("Expected success result")
	}
	
	if len(result.Content) != 1 {
		t.Errorf("Expected 1 content item, got %d", len(result.Content))
	}
	
	content, ok := result.Content[0].(map[string]interface{})
	if !ok {
		t.Error("Expected content to be map[string]interface{}")
	}
	
	text, ok := content["text"].(string)
	if !ok {
		t.Error("Expected text field in content")
	}
	
	if !strings.Contains(text, "Memory saved successfully") {
		t.Errorf("Expected success message, got: %s", text)
	}
	
	if !strings.Contains(text, "personal") {
		t.Errorf("Expected category in response, got: %s", text)
	}
	
	if !strings.Contains(text, "test value") {
		t.Errorf("Expected value in response, got: %s", text)
	}
	
	if !strings.Contains(text, "test_key") {
		t.Errorf("Expected key in response, got: %s", text)
	}
}

func TestHandleSaveMemory_MissingCategory(t *testing.T) {
	cfg := config.DefaultConfig()
	store := NewMockMemoryStore()
	server := NewServer(cfg, store)
	
	arguments := map[string]interface{}{
		"value": "test value",
	}
	
	result, err := server.handleSaveMemory(context.Background(), arguments)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if !result.IsError {
		t.Error("Expected error result")
	}
	
	content := result.Content[0].(map[string]interface{})
	text := content["text"].(string)
	
	if !strings.Contains(text, "category parameter is required") {
		t.Errorf("Expected category error, got: %s", text)
	}
}

func TestHandleSaveMemory_MissingValue(t *testing.T) {
	cfg := config.DefaultConfig()
	store := NewMockMemoryStore()
	server := NewServer(cfg, store)
	
	arguments := map[string]interface{}{
		"category": "personal",
	}
	
	result, err := server.handleSaveMemory(context.Background(), arguments)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if !result.IsError {
		t.Error("Expected error result")
	}
	
	content := result.Content[0].(map[string]interface{})
	text := content["text"].(string)
	
	if !strings.Contains(text, "value parameter is required") {
		t.Errorf("Expected value error, got: %s", text)
	}
}

func TestHandleSaveMemory_NilStore(t *testing.T) {
	cfg := config.DefaultConfig()
	server := NewServer(cfg, nil)
	
	arguments := map[string]interface{}{
		"category": "personal",
		"value":    "test value",
	}
	
	result, err := server.handleSaveMemory(context.Background(), arguments)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if !result.IsError {
		t.Error("Expected error result")
	}
	
	content := result.Content[0].(map[string]interface{})
	text := content["text"].(string)
	
	if !strings.Contains(text, "memory store not initialized") {
		t.Errorf("Expected store error, got: %s", text)
	}
}

func TestHandleGetMemory_Success(t *testing.T) {
	cfg := config.DefaultConfig()
	store := NewMockMemoryStore()
	server := NewServer(cfg, store)
	
	// First save a memory
	testMemory := &memory.Memory{
		Category: "personal",
		Key:      "test_key",
		Value:    "test value",
		Tags:     []string{"tag1", "tag2"},
	}
	_, err := store.Save(testMemory)
	if err != nil {
		t.Fatalf("Failed to save test memory: %v", err)
	}
	
	arguments := map[string]interface{}{
		"key": "test_key",
	}
	
	result, err := server.handleGetMemory(context.Background(), arguments)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if result.IsError {
		t.Error("Expected success result")
	}
	
	content := result.Content[0].(map[string]interface{})
	text := content["text"].(string)
	
	if !strings.Contains(text, "Memory retrieved successfully") {
		t.Errorf("Expected success message, got: %s", text)
	}
	
	if !strings.Contains(text, "personal") {
		t.Errorf("Expected category in response, got: %s", text)
	}
	
	if !strings.Contains(text, "test value") {
		t.Errorf("Expected value in response, got: %s", text)
	}
	
	if !strings.Contains(text, "test_key") {
		t.Errorf("Expected key in response, got: %s", text)
	}
	
	if !strings.Contains(text, "tag1") {
		t.Errorf("Expected tags in response, got: %s", text)
	}
}

func TestHandleGetMemory_NotFound(t *testing.T) {
	cfg := config.DefaultConfig()
	store := NewMockMemoryStore()
	server := NewServer(cfg, store)
	
	arguments := map[string]interface{}{
		"key": "nonexistent",
	}
	
	result, err := server.handleGetMemory(context.Background(), arguments)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if !result.IsError {
		t.Error("Expected error result")
	}
	
	content := result.Content[0].(map[string]interface{})
	text := content["text"].(string)
	
	if !strings.Contains(text, "Memory not found") {
		t.Errorf("Expected not found error, got: %s", text)
	}
}

func TestHandleListMemories_Empty(t *testing.T) {
	cfg := config.DefaultConfig()
	store := NewMockMemoryStore()
	server := NewServer(cfg, store)
	
	arguments := map[string]interface{}{}
	
	result, err := server.handleListMemories(context.Background(), arguments)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if result.IsError {
		t.Error("Expected success result")
	}
	
	content := result.Content[0].(map[string]interface{})
	text := content["text"].(string)
	
	if !strings.Contains(text, "No memories stored yet") {
		t.Errorf("Expected empty message, got: %s", text)
	}
}

func TestHandleListMemories_WithData(t *testing.T) {
	cfg := config.DefaultConfig()
	store := NewMockMemoryStore()
	server := NewServer(cfg, store)
	
	// Add test memories
	memory1 := &memory.Memory{Category: "personal", Key: "key1", Value: "value1"}
	memory2 := &memory.Memory{Category: "work", Key: "key2", Value: "value2"}
	
	_, err := store.Save(memory1)
	if err != nil {
		t.Fatalf("Failed to save memory1: %v", err)
	}
	
	_, err = store.Save(memory2)
	if err != nil {
		t.Fatalf("Failed to save memory2: %v", err)
	}
	
	arguments := map[string]interface{}{}
	
	result, err := server.handleListMemories(context.Background(), arguments)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if result.IsError {
		t.Error("Expected success result")
	}
	
	content := result.Content[0].(map[string]interface{})
	text := content["text"].(string)
	
	if !strings.Contains(text, "All stored memories") {
		t.Errorf("Expected list header, got: %s", text)
	}
	
	if !strings.Contains(text, "total: 2") {
		t.Errorf("Expected count in response, got: %s", text)
	}
	
	if !strings.Contains(text, "value1") {
		t.Errorf("Expected value1 in response, got: %s", text)
	}
	
	if !strings.Contains(text, "value2") {
		t.Errorf("Expected value2 in response, got: %s", text)
	}
}

func TestHandleListMemories_WithCategoryFilter(t *testing.T) {
	cfg := config.DefaultConfig()
	store := NewMockMemoryStore()
	server := NewServer(cfg, store)
	
	// Add test memories
	memory1 := &memory.Memory{Category: "personal", Key: "key1", Value: "value1"}
	memory2 := &memory.Memory{Category: "work", Key: "key2", Value: "value2"}
	
	_, err := store.Save(memory1)
	if err != nil {
		t.Fatalf("Failed to save memory1: %v", err)
	}
	
	_, err = store.Save(memory2)
	if err != nil {
		t.Fatalf("Failed to save memory2: %v", err)
	}
	
	arguments := map[string]interface{}{
		"category": "personal",
	}
	
	result, err := server.handleListMemories(context.Background(), arguments)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if result.IsError {
		t.Error("Expected success result")
	}
	
	content := result.Content[0].(map[string]interface{})
	text := content["text"].(string)
	
	if !strings.Contains(text, "Memories in category 'personal'") {
		t.Errorf("Expected category filter header, got: %s", text)
	}
	
	if !strings.Contains(text, "total: 1") {
		t.Errorf("Expected filtered count in response, got: %s", text)
	}
	
	if !strings.Contains(text, "value1") {
		t.Errorf("Expected value1 in response, got: %s", text)
	}
	
	if strings.Contains(text, "value2") {
		t.Errorf("Expected value2 NOT in response, got: %s", text)
	}
}

func TestHandleGetMemory_MissingKey(t *testing.T) {
	cfg := config.DefaultConfig()
	store := NewMockMemoryStore()
	server := NewServer(cfg, store)

	arguments := map[string]interface{}{}

	result, err := server.handleGetMemory(context.Background(), arguments)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !result.IsError {
		t.Error("Expected error result")
	}

	content := result.Content[0].(map[string]interface{})
	text := content["text"].(string)

	if !strings.Contains(text, "key parameter is required") {
		t.Errorf("Expected key error, got: %s", text)
	}
}

func TestHandleGetMemory_EmptyKey(t *testing.T) {
	cfg := config.DefaultConfig()
	store := NewMockMemoryStore()
	server := NewServer(cfg, store)

	arguments := map[string]interface{}{
		"key": "",
	}

	result, err := server.handleGetMemory(context.Background(), arguments)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !result.IsError {
		t.Error("Expected error result")
	}

	content := result.Content[0].(map[string]interface{})
	text := content["text"].(string)

	if !strings.Contains(text, "key parameter is required") {
		t.Errorf("Expected key error, got: %s", text)
	}
}

func TestHandleGetMemory_NilStore(t *testing.T) {
	cfg := config.DefaultConfig()
	server := NewServer(cfg, nil)

	arguments := map[string]interface{}{
		"key": "test",
	}

	result, err := server.handleGetMemory(context.Background(), arguments)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !result.IsError {
		t.Error("Expected error result")
	}

	content := result.Content[0].(map[string]interface{})
	text := content["text"].(string)

	if !strings.Contains(text, "memory store not initialized") {
		t.Errorf("Expected store error, got: %s", text)
	}
}

func TestHandleGetMemory_ByID(t *testing.T) {
	cfg := config.DefaultConfig()
	store := NewMockMemoryStore()
	server := NewServer(cfg, store)

	// Save a memory without key (will be stored by ID)
	testMemory := &memory.Memory{
		Category: "personal",
		Value:    "test value",
		Tags:     []string{"tag1"},
	}
	id, err := store.Save(testMemory)
	if err != nil {
		t.Fatalf("Failed to save test memory: %v", err)
	}

	arguments := map[string]interface{}{
		"key": id, // Using ID as key parameter
	}

	result, err := server.handleGetMemory(context.Background(), arguments)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.IsError {
		t.Error("Expected success result")
	}

	content := result.Content[0].(map[string]interface{})
	text := content["text"].(string)

	if !strings.Contains(text, "Memory retrieved successfully") {
		t.Errorf("Expected success message, got: %s", text)
	}

	if !strings.Contains(text, "test value") {
		t.Errorf("Expected value in response, got: %s", text)
	}
}

func TestHandleListMemories_NilStore(t *testing.T) {
	cfg := config.DefaultConfig()
	server := NewServer(cfg, nil)

	arguments := map[string]interface{}{}

	result, err := server.handleListMemories(context.Background(), arguments)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !result.IsError {
		t.Error("Expected error result")
	}

	content := result.Content[0].(map[string]interface{})
	text := content["text"].(string)

	if !strings.Contains(text, "memory store not initialized") {
		t.Errorf("Expected store error, got: %s", text)
	}
}

func TestHandleListMemories_EmptyCategoryFilter(t *testing.T) {
	cfg := config.DefaultConfig()
	store := NewMockMemoryStore()
	server := NewServer(cfg, store)

	// Add test memories
	memory1 := &memory.Memory{Category: "personal", Key: "key1", Value: "value1"}
	memory2 := &memory.Memory{Category: "work", Key: "key2", Value: "value2"}

	_, err := store.Save(memory1)
	if err != nil {
		t.Fatalf("Failed to save memory1: %v", err)
	}

	_, err = store.Save(memory2)
	if err != nil {
		t.Fatalf("Failed to save memory2: %v", err)
	}

	arguments := map[string]interface{}{
		"category": "nonexistent",
	}

	result, err := server.handleListMemories(context.Background(), arguments)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.IsError {
		t.Error("Expected success result")
	}

	content := result.Content[0].(map[string]interface{})
	text := content["text"].(string)

	if !strings.Contains(text, "No memories found in category 'nonexistent'") {
		t.Errorf("Expected empty category message, got: %s", text)
	}
}

func TestHandleSaveMemory_WithTags(t *testing.T) {
	cfg := config.DefaultConfig()
	store := NewMockMemoryStore()
	server := NewServer(cfg, store)

	arguments := map[string]interface{}{
		"category": "personal",
		"value":    "test value with tags",
		"key":      "tagged_memory",
	}

	result, err := server.handleSaveMemory(context.Background(), arguments)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.IsError {
		t.Error("Expected success result")
	}

	// Verify the memory was saved with empty tags (as initialized in handler)
	savedMemory, err := store.Get("tagged_memory")
	if err != nil {
		t.Fatalf("Failed to retrieve saved memory: %v", err)
	}

	if savedMemory.Tags == nil {
		t.Error("Expected Tags to be initialized")
	}

	if len(savedMemory.Tags) != 0 {
		t.Errorf("Expected empty tags, got %v", savedMemory.Tags)
	}
}

func TestHandleSaveMemory_InvalidArguments(t *testing.T) {
	cfg := config.DefaultConfig()
	store := NewMockMemoryStore()
	server := NewServer(cfg, store)

	// Test with non-string category
	arguments := map[string]interface{}{
		"category": 123,
		"value":    "test value",
	}

	result, err := server.handleSaveMemory(context.Background(), arguments)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !result.IsError {
		t.Error("Expected error result")
	}

	// Test with non-string value
	arguments2 := map[string]interface{}{
		"category": "personal",
		"value":    123,
	}

	result2, err := server.handleSaveMemory(context.Background(), arguments2)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !result2.IsError {
		t.Error("Expected error result")
	}
}