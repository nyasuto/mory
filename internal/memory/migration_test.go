package memory

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMigrator_BasicMigration(t *testing.T) {
	tempDir := t.TempDir()

	// Create test JSON file
	jsonPath := filepath.Join(tempDir, "test.json")
	sqlitePath := filepath.Join(tempDir, "test.db")

	testMemories := []*Memory{
		{
			ID:        "memory_1",
			Category:  "test",
			Key:       "key1",
			Value:     "value1",
			Tags:      []string{"tag1"},
			CreatedAt: time.Now().Add(-time.Hour),
			UpdatedAt: time.Now().Add(-time.Hour),
		},
		{
			ID:        "memory_2",
			Category:  "test",
			Key:       "key2",
			Value:     "value2",
			Tags:      []string{"tag2", "tag3"},
			CreatedAt: time.Now().Add(-time.Minute),
			UpdatedAt: time.Now().Add(-time.Minute),
		},
	}

	// Write test data to JSON file
	data, err := json.MarshalIndent(testMemories, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	if err := os.WriteFile(jsonPath, data, 0644); err != nil {
		t.Fatalf("Failed to write test JSON file: %v", err)
	}

	// Create migration options
	options := DefaultMigrationOptions()
	options.JSONFile = jsonPath
	options.SQLitePath = sqlitePath
	options.BackupJSON = false // Skip backup for test
	options.ValidateAfter = true

	// Perform migration
	migrator := NewMigrator(options)
	result, err := migrator.Migrate()
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Verify migration result
	if result.TotalMemories != 2 {
		t.Errorf("Expected 2 total memories, got %d", result.TotalMemories)
	}

	if result.MigratedMemories != 2 {
		t.Errorf("Expected 2 migrated memories, got %d", result.MigratedMemories)
	}

	if result.FailedMemories != 0 {
		t.Errorf("Expected 0 failed memories, got %d", result.FailedMemories)
	}

	if !result.ValidationPassed {
		t.Error("Expected validation to pass")
	}

	// Verify SQLite database content
	store, err := NewSQLiteMemoryStore(sqlitePath)
	if err != nil {
		t.Fatalf("Failed to open SQLite store: %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			t.Logf("Warning: failed to close store: %v", err)
		}
	}()

	memories, err := store.List("")
	if err != nil {
		t.Fatalf("Failed to list memories from SQLite: %v", err)
	}

	if len(memories) != 2 {
		t.Errorf("Expected 2 memories in SQLite, got %d", len(memories))
	}

	// Verify data integrity
	for _, original := range testMemories {
		retrieved, err := store.GetByID(original.ID)
		if err != nil {
			t.Errorf("Failed to get memory %s: %v", original.ID, err)
			continue
		}

		if retrieved.Category != original.Category {
			t.Errorf("Category mismatch for %s: expected %s, got %s",
				original.ID, original.Category, retrieved.Category)
		}

		if retrieved.Value != original.Value {
			t.Errorf("Value mismatch for %s: expected %s, got %s",
				original.ID, original.Value, retrieved.Value)
		}

		if len(retrieved.Tags) != len(original.Tags) {
			t.Errorf("Tags length mismatch for %s: expected %d, got %d",
				original.ID, len(original.Tags), len(retrieved.Tags))
		}
	}
}

func TestMigrator_SkipExisting(t *testing.T) {
	tempDir := t.TempDir()

	jsonPath := filepath.Join(tempDir, "test.json")
	sqlitePath := filepath.Join(tempDir, "test.db")

	// Create test data
	testMemories := []*Memory{
		{
			ID:       "memory_1",
			Category: "test",
			Key:      "existing_key",
			Value:    "json_value",
		},
		{
			ID:       "memory_2",
			Category: "test",
			Key:      "new_key",
			Value:    "new_value",
		},
	}

	data, err := json.MarshalIndent(testMemories, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	if err := os.WriteFile(jsonPath, data, 0644); err != nil {
		t.Fatalf("Failed to write test JSON file: %v", err)
	}

	// Pre-populate SQLite with one memory
	store, err := NewSQLiteMemoryStore(sqlitePath)
	if err != nil {
		t.Fatalf("Failed to create SQLite store: %v", err)
	}

	existingMemory := &Memory{
		Category: "test",
		Key:      "existing_key",
		Value:    "sqlite_value", // Different value
	}

	_, err = store.Save(existingMemory)
	if err != nil {
		t.Fatalf("Failed to save existing memory: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Logf("Warning: failed to close store: %v", err)
	}

	// Perform migration with skip existing
	options := DefaultMigrationOptions()
	options.JSONFile = jsonPath
	options.SQLitePath = sqlitePath
	options.SkipExisting = true
	options.BackupJSON = false
	options.ValidateAfter = false // Skip validation since we're testing skip behavior

	migrator := NewMigrator(options)
	result, err := migrator.Migrate()
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Should skip 1 and migrate 1
	if result.SkippedMemories != 1 {
		t.Errorf("Expected 1 skipped memory, got %d", result.SkippedMemories)
	}

	if result.MigratedMemories != 1 {
		t.Errorf("Expected 1 migrated memory, got %d", result.MigratedMemories)
	}

	// Verify that existing memory was not overwritten
	store, err = NewSQLiteMemoryStore(sqlitePath)
	if err != nil {
		t.Fatalf("Failed to reopen SQLite store: %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			t.Logf("Warning: failed to close store: %v", err)
		}
	}()

	retrieved, err := store.Get("existing_key")
	if err != nil {
		t.Fatalf("Failed to get existing memory: %v", err)
	}

	if retrieved.Value != "sqlite_value" {
		t.Errorf("Expected existing value to be preserved: 'sqlite_value', got '%s'", retrieved.Value)
	}
}

func TestMigrator_WithEmbeddings(t *testing.T) {
	tempDir := t.TempDir()

	jsonPath := filepath.Join(tempDir, "test_embeddings.json")
	sqlitePath := filepath.Join(tempDir, "test_embeddings.db")

	// Create test data with embeddings
	testMemories := []*Memory{
		{
			ID:            "memory_1",
			Category:      "test",
			Key:           "embedding_test",
			Value:         "test with embedding",
			Embedding:     []float32{0.1, 0.2, 0.3, 0.4},
			EmbeddingHash: "test_hash_123",
		},
		{
			ID:       "memory_2",
			Category: "test",
			Key:      "no_embedding",
			Value:    "test without embedding",
		},
	}

	data, err := json.MarshalIndent(testMemories, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	if err := os.WriteFile(jsonPath, data, 0644); err != nil {
		t.Fatalf("Failed to write test JSON file: %v", err)
	}

	// Perform migration
	result, err := MigrateFromJSON(jsonPath, sqlitePath)
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	if result.MigratedMemories != 2 {
		t.Errorf("Expected 2 migrated memories, got %d", result.MigratedMemories)
	}

	// Verify embeddings were preserved
	store, err := NewSQLiteMemoryStore(sqlitePath)
	if err != nil {
		t.Fatalf("Failed to open SQLite store: %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			t.Logf("Warning: failed to close store: %v", err)
		}
	}()

	// Check memory with embedding
	withEmbedding, err := store.Get("embedding_test")
	if err != nil {
		t.Fatalf("Failed to get memory with embedding: %v", err)
	}

	if len(withEmbedding.Embedding) != 4 {
		t.Errorf("Expected 4 embedding values, got %d", len(withEmbedding.Embedding))
	}

	if withEmbedding.EmbeddingHash != "test_hash_123" {
		t.Errorf("Expected embedding hash 'test_hash_123', got '%s'", withEmbedding.EmbeddingHash)
	}

	// Check memory without embedding
	withoutEmbedding, err := store.Get("no_embedding")
	if err != nil {
		t.Fatalf("Failed to get memory without embedding: %v", err)
	}

	if len(withoutEmbedding.Embedding) != 0 {
		t.Errorf("Expected no embedding, got %d values", len(withoutEmbedding.Embedding))
	}
}

func TestMigrator_BackupCreation(t *testing.T) {
	tempDir := t.TempDir()

	jsonPath := filepath.Join(tempDir, "test_backup.json")
	sqlitePath := filepath.Join(tempDir, "test_backup.db")

	testData := []*Memory{
		{Category: "test", Key: "backup_test", Value: "backup value"},
	}

	data, err := json.MarshalIndent(testData, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	if err := os.WriteFile(jsonPath, data, 0644); err != nil {
		t.Fatalf("Failed to write test JSON file: %v", err)
	}

	// Migration with backup enabled
	options := DefaultMigrationOptions()
	options.JSONFile = jsonPath
	options.SQLitePath = sqlitePath
	options.BackupJSON = true
	options.ValidateAfter = false

	migrator := NewMigrator(options)
	result, err := migrator.Migrate()
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	if !result.BackupCreated {
		t.Error("Expected backup to be created")
	}

	if result.BackupPath == "" {
		t.Error("Expected backup path to be set")
	}

	// Verify backup file exists
	if _, err := os.Stat(result.BackupPath); os.IsNotExist(err) {
		t.Errorf("Backup file does not exist: %s", result.BackupPath)
	}

	// Verify backup content
	backupData, err := os.ReadFile(result.BackupPath)
	if err != nil {
		t.Fatalf("Failed to read backup file: %v", err)
	}

	var backupMemories []*Memory
	if err := json.Unmarshal(backupData, &backupMemories); err != nil {
		t.Fatalf("Failed to unmarshal backup data: %v", err)
	}

	if len(backupMemories) != 1 {
		t.Errorf("Expected 1 memory in backup, got %d", len(backupMemories))
	}

	if backupMemories[0].Value != "backup value" {
		t.Errorf("Expected backup value 'backup value', got '%s'", backupMemories[0].Value)
	}
}

func TestMigrator_ErrorHandling(t *testing.T) {
	tempDir := t.TempDir()

	// Test with non-existent JSON file
	options := DefaultMigrationOptions()
	options.JSONFile = filepath.Join(tempDir, "nonexistent.json")
	options.SQLitePath = filepath.Join(tempDir, "test.db")

	migrator := NewMigrator(options)
	_, err := migrator.Migrate()
	if err == nil {
		t.Error("Expected error for non-existent JSON file")
	}

	// Test with invalid JSON
	invalidJSONPath := filepath.Join(tempDir, "invalid.json")
	if err := os.WriteFile(invalidJSONPath, []byte("{invalid json"), 0644); err != nil {
		t.Fatalf("Failed to write invalid JSON file: %v", err)
	}

	options.JSONFile = invalidJSONPath
	migrator = NewMigrator(options)
	_, err = migrator.Migrate()
	if err == nil {
		t.Error("Expected error for invalid JSON file")
	}
}

func TestMigrateBulk(t *testing.T) {
	tempDir := t.TempDir()

	// Create multiple JSON files
	file1 := filepath.Join(tempDir, "file1.json")
	file2 := filepath.Join(tempDir, "file2.json")
	sqlitePath := filepath.Join(tempDir, "bulk.db")

	data1 := []*Memory{{Category: "file1", Key: "key1", Value: "value1"}}
	data2 := []*Memory{{Category: "file2", Key: "key2", Value: "value2"}}

	json1, _ := json.MarshalIndent(data1, "", "  ")
	json2, _ := json.MarshalIndent(data2, "", "  ")

	if err := os.WriteFile(file1, json1, 0644); err != nil {
		t.Fatalf("Failed to write file1: %v", err)
	}

	if err := os.WriteFile(file2, json2, 0644); err != nil {
		t.Fatalf("Failed to write file2: %v", err)
	}

	// Perform bulk migration
	result, err := MigrateBulk([]string{file1, file2}, sqlitePath)
	if err != nil {
		t.Fatalf("Bulk migration failed: %v", err)
	}

	if result.TotalMemories != 2 {
		t.Errorf("Expected 2 total memories, got %d", result.TotalMemories)
	}

	if result.MigratedMemories != 2 {
		t.Errorf("Expected 2 migrated memories, got %d", result.MigratedMemories)
	}

	// Verify data
	store, err := NewSQLiteMemoryStore(sqlitePath)
	if err != nil {
		t.Fatalf("Failed to open SQLite store: %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			t.Logf("Warning: failed to close store: %v", err)
		}
	}()

	memories, err := store.List("")
	if err != nil {
		t.Fatalf("Failed to list memories: %v", err)
	}

	if len(memories) != 2 {
		t.Errorf("Expected 2 memories in database, got %d", len(memories))
	}
}
