package memory

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// JSONMemoryStore implements MemoryStore interface using JSON files
type JSONMemoryStore struct {
	dataFile string
	logFile  string
	mutex    sync.RWMutex
}

// NewJSONMemoryStore creates a new JSON-based memory store
func NewJSONMemoryStore(dataFile, logFile string) *JSONMemoryStore {
	return &JSONMemoryStore{
		dataFile: dataFile,
		logFile:  logFile,
	}
}

// Save stores a memory item and returns the generated ID
func (s *JSONMemoryStore) Save(memory *Memory) (string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Generate ID if not provided
	if memory.ID == "" {
		memory.ID = GenerateID()
	}

	// Set timestamps
	now := time.Now()
	if memory.CreatedAt.IsZero() {
		memory.CreatedAt = now
	}
	memory.UpdatedAt = now

	// Initialize Tags if nil
	if memory.Tags == nil {
		memory.Tags = []string{}
	}

	// Load existing memories
	memories, err := s.loadMemories()
	if err != nil {
		return "", fmt.Errorf("failed to load existing memories: %w", err)
	}

	// Check for existing key and update if found
	var existingMemory *Memory
	for i, m := range memories {
		if (memory.Key != "" && m.Key == memory.Key) || m.ID == memory.ID {
			existingMemory = &Memory{
				ID:        m.ID,
				Category:  m.Category,
				Key:       m.Key,
				Value:     m.Value,
				Tags:      append([]string{}, m.Tags...),
				CreatedAt: m.CreatedAt,
				UpdatedAt: m.UpdatedAt,
			}
			memories[i] = memory
			break
		}
	}

	// If not updating existing, add new memory
	if existingMemory == nil {
		memories = append(memories, memory)
	}

	// Save memories to file
	if err := s.saveMemories(memories); err != nil {
		return "", fmt.Errorf("failed to save memories: %w", err)
	}

	// Log the operation
	operation := "save"
	if existingMemory != nil {
		operation = "update"
	}

	log := &OperationLog{
		Timestamp:   now,
		OperationID: GenerateOperationID(),
		Operation:   operation,
		Key:         memory.Key,
		Before:      existingMemory,
		After:       memory,
		Success:     true,
	}

	if err := s.LogOperation(log); err != nil {
		// Log error but don't fail the save operation
		fmt.Printf("Warning: failed to log operation: %v\n", err)
	}

	return memory.ID, nil
}

// Get retrieves a memory by key
func (s *JSONMemoryStore) Get(key string) (*Memory, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	memories, err := s.loadMemories()
	if err != nil {
		return nil, fmt.Errorf("failed to load memories: %w", err)
	}

	for _, memory := range memories {
		if memory.Key == key {
			return memory, nil
		}
	}

	return nil, fmt.Errorf("memory with key '%s' not found", key)
}

// GetByID retrieves a memory by ID
func (s *JSONMemoryStore) GetByID(id string) (*Memory, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	memories, err := s.loadMemories()
	if err != nil {
		return nil, fmt.Errorf("failed to load memories: %w", err)
	}

	for _, memory := range memories {
		if memory.ID == id {
			return memory, nil
		}
	}

	return nil, fmt.Errorf("memory with ID '%s' not found", id)
}

// List retrieves all memories, optionally filtered by category
func (s *JSONMemoryStore) List(category string) ([]*Memory, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	memories, err := s.loadMemories()
	if err != nil {
		return nil, fmt.Errorf("failed to load memories: %w", err)
	}

	var result []*Memory
	for _, memory := range memories {
		if category == "" || memory.Category == category {
			result = append(result, memory)
		}
	}

	// Sort by creation time (newest first)
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})

	return result, nil
}

// Delete removes a memory by key
func (s *JSONMemoryStore) Delete(key string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	memories, err := s.loadMemories()
	if err != nil {
		return fmt.Errorf("failed to load memories: %w", err)
	}

	var deletedMemory *Memory
	var filteredMemories []*Memory

	for _, memory := range memories {
		if memory.Key == key {
			deletedMemory = memory
		} else {
			filteredMemories = append(filteredMemories, memory)
		}
	}

	if deletedMemory == nil {
		return fmt.Errorf("memory with key '%s' not found", key)
	}

	// Save updated memories
	if err := s.saveMemories(filteredMemories); err != nil {
		return fmt.Errorf("failed to save memories after deletion: %w", err)
	}

	// Log the operation
	log := &OperationLog{
		Timestamp:   time.Now(),
		OperationID: GenerateOperationID(),
		Operation:   "delete",
		Key:         key,
		Before:      deletedMemory,
		After:       nil,
		Success:     true,
	}

	if err := s.LogOperation(log); err != nil {
		fmt.Printf("Warning: failed to log deletion operation: %v\n", err)
	}

	return nil
}

// DeleteByID removes a memory by ID
func (s *JSONMemoryStore) DeleteByID(id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	memories, err := s.loadMemories()
	if err != nil {
		return fmt.Errorf("failed to load memories: %w", err)
	}

	var deletedMemory *Memory
	var filteredMemories []*Memory

	for _, memory := range memories {
		if memory.ID == id {
			deletedMemory = memory
		} else {
			filteredMemories = append(filteredMemories, memory)
		}
	}

	if deletedMemory == nil {
		return fmt.Errorf("memory with ID '%s' not found", id)
	}

	// Save updated memories
	if err := s.saveMemories(filteredMemories); err != nil {
		return fmt.Errorf("failed to save memories after deletion: %w", err)
	}

	// Log the operation
	log := &OperationLog{
		Timestamp:   time.Now(),
		OperationID: GenerateOperationID(),
		Operation:   "delete",
		Key:         deletedMemory.Key,
		Before:      deletedMemory,
		After:       nil,
		Success:     true,
	}

	if err := s.LogOperation(log); err != nil {
		fmt.Printf("Warning: failed to log deletion operation: %v\n", err)
	}

	return nil
}

// LogOperation records an operation log
func (s *JSONMemoryStore) LogOperation(log *OperationLog) error {
	// Create log directory if it doesn't exist
	logDir := filepath.Dir(s.logFile)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open log file for append
	file, err := os.OpenFile(s.logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close log file: %v\n", closeErr)
		}
	}()

	// Convert log to JSON and write
	logData, err := json.Marshal(log)
	if err != nil {
		return fmt.Errorf("failed to marshal log: %w", err)
	}

	_, err = file.Write(append(logData, '\n'))
	if err != nil {
		return fmt.Errorf("failed to write log: %w", err)
	}

	return nil
}

// loadMemories loads memories from the JSON file
func (s *JSONMemoryStore) loadMemories() ([]*Memory, error) {
	if _, err := os.Stat(s.dataFile); os.IsNotExist(err) {
		return []*Memory{}, nil
	}

	data, err := os.ReadFile(s.dataFile)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return []*Memory{}, nil
	}

	var memories []*Memory
	if err := json.Unmarshal(data, &memories); err != nil {
		return nil, err
	}

	return memories, nil
}

// saveMemories saves memories to the JSON file
func (s *JSONMemoryStore) saveMemories(memories []*Memory) error {
	// Create data directory if it doesn't exist
	dataDir := filepath.Dir(s.dataFile)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	data, err := json.MarshalIndent(memories, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.dataFile, data, 0644)
}
