package memory

import (
	"strings"
	"testing"
	"time"
)

func TestGenerateID(t *testing.T) {
	id1 := GenerateID()
	if id1 == "" {
		t.Error("GenerateID returned empty string")
	}

	// Check format (nanosecond timestamp should be longer)
	if len(id1) < 15 || !strings.HasPrefix(id1, "memory_") {
		t.Errorf("GenerateID returned unexpected format: %s", id1)
	}

	if id1[:7] != "memory_" {
		t.Errorf("GenerateID should start with 'memory_': %s", id1)
	}

	// Generate another ID to ensure uniqueness (with small delay)
	time.Sleep(time.Microsecond)
	id2 := GenerateID()
	if id1 == id2 {
		t.Error("GenerateID should generate unique IDs")
	}
}

func TestGenerateOperationID(t *testing.T) {
	opID1 := GenerateOperationID()
	if opID1 == "" {
		t.Error("GenerateOperationID returned empty string")
	}

	if opID1[:3] != "op_" {
		t.Errorf("GenerateOperationID should start with 'op_': %s", opID1)
	}

	// Generate another ID to ensure uniqueness
	time.Sleep(time.Microsecond)
	opID2 := GenerateOperationID()
	if opID1 == opID2 {
		t.Error("GenerateOperationID should generate unique IDs")
	}
}

func TestMemoryStruct(t *testing.T) {
	now := time.Now()
	memory := &Memory{
		ID:        GenerateID(),
		Category:  "personal",
		Key:       "birthday",
		Value:     "1990年5月15日",
		Tags:      []string{"personal", "important"},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if memory.ID == "" {
		t.Error("Memory ID should not be empty")
	}

	if len(memory.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(memory.Tags))
	}

	if memory.Tags[0] != "personal" || memory.Tags[1] != "important" {
		t.Errorf("Tags not set correctly: %v", memory.Tags)
	}
}

func TestOperationLogStruct(t *testing.T) {
	now := time.Now()
	memory := &Memory{
		ID:        GenerateID(),
		Category:  "test",
		Key:       "test_key",
		Value:     "test_value",
		Tags:      []string{},
		CreatedAt: now,
		UpdatedAt: now,
	}

	log := &OperationLog{
		Timestamp:   now,
		OperationID: GenerateOperationID(),
		Operation:   "save",
		Key:         "test_key",
		Before:      nil,
		After:       memory,
		Success:     true,
		Error:       "",
	}

	if log.OperationID == "" {
		t.Error("OperationLog ID should not be empty")
	}

	if log.Operation != "save" {
		t.Errorf("Expected operation 'save', got '%s'", log.Operation)
	}

	if log.After == nil {
		t.Error("OperationLog After should not be nil")
	}

	if !log.Success {
		t.Error("OperationLog Success should be true")
	}
}
