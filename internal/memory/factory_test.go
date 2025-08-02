package memory

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nyasuto/mory/internal/config"
)

func TestNewStorageFactory(t *testing.T) {
	cfg := &config.Config{
		Storage: &config.StorageConfig{
			Type:     "json",
			JSONPath: "/tmp/test/memories.json",
		},
	}

	factory := NewStorageFactory(cfg)
	if factory == nil {
		t.Fatal("NewStorageFactory returned nil")
	}

	if factory.config != cfg {
		t.Error("config not set correctly")
	}
}

func TestStorageFactory_CreateMemoryStore_JSON(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.Config{
		Storage: &config.StorageConfig{
			Type:     "json",
			JSONPath: filepath.Join(tempDir, "memories.json"),
			LogPath:  filepath.Join(tempDir, "operations.log"),
		},
	}

	factory := NewStorageFactory(cfg)
	store, err := factory.CreateMemoryStore()
	if err != nil {
		t.Fatalf("CreateMemoryStore failed: %v", err)
	}

	if store == nil {
		t.Fatal("Expected store, got nil")
	}

	// Check if it's the correct type
	_, ok := store.(*JSONMemoryStore)
	if !ok {
		t.Error("Expected JSONMemoryStore")
	}
}

func TestStorageFactory_CreateMemoryStore_SQLite(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.Config{
		Storage: &config.StorageConfig{
			Type:       "sqlite",
			SQLitePath: filepath.Join(tempDir, "memories.db"),
			LogPath:    filepath.Join(tempDir, "operations.log"),
		},
	}

	factory := NewStorageFactory(cfg)
	store, err := factory.CreateMemoryStore()
	if err != nil {
		t.Fatalf("CreateMemoryStore failed: %v", err)
	}

	if store == nil {
		t.Fatal("Expected store, got nil")
	}

	// Check if it's the correct type
	sqliteStore, ok := store.(*SQLiteMemoryStore)
	if !ok {
		t.Error("Expected SQLiteMemoryStore")
	}

	// Clean up
	if sqliteStore != nil {
		err := sqliteStore.Close()
		if err != nil {
			t.Logf("Failed to close SQLite store: %v", err)
		}
	}
}

func TestStorageFactory_CreateMemoryStore_InvalidType(t *testing.T) {
	cfg := &config.Config{
		Storage: &config.StorageConfig{
			Type: "invalid",
		},
	}

	factory := NewStorageFactory(cfg)
	store, err := factory.CreateMemoryStore()

	// Factory should fallback to JSON for invalid types, not return an error
	if err != nil {
		t.Errorf("Factory should fallback to JSON, got error: %v", err)
	}

	if store == nil {
		t.Error("Expected store to be created (fallback to JSON)")
	}

	// Should create a JSON store as fallback
	_, ok := store.(*JSONMemoryStore)
	if !ok {
		t.Error("Expected JSONMemoryStore as fallback for invalid type")
	}
}

func TestStorageFactory_GetStorageType(t *testing.T) {
	cfg := &config.Config{
		Storage: &config.StorageConfig{
			Type: "sqlite",
		},
	}

	factory := NewStorageFactory(cfg)
	storageType := factory.GetStorageType()
	if storageType != "sqlite" {
		t.Errorf("Expected 'sqlite', got '%s'", storageType)
	}
}

func TestStorageFactory_GetStoragePaths(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.Config{
		Storage: &config.StorageConfig{
			Type:     "json",
			JSONPath: filepath.Join(tempDir, "memories.json"),
			LogPath:  filepath.Join(tempDir, "operations.log"),
		},
	}

	factory := NewStorageFactory(cfg)
	paths := factory.GetStoragePaths()

	if paths == nil {
		t.Fatal("Expected paths, got nil")
	}

	// Check for expected keys
	expectedKeys := []string{"json", "log"}
	for _, key := range expectedKeys {
		if _, ok := paths[key]; !ok {
			t.Errorf("Expected key '%s' in paths", key)
		}
	}
}

func TestStorageFactory_EnsureDirectoryExists(t *testing.T) {
	tempDir := t.TempDir()
	testDir := filepath.Join(tempDir, "test", "nested", "dir")

	cfg := &config.Config{}
	factory := NewStorageFactory(cfg)

	err := factory.ensureDirectoryExists(testDir)
	if err != nil {
		t.Fatalf("ensureDirectoryExists failed: %v", err)
	}

	// Check if directory was created
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		t.Error("Directory was not created")
	}
}

func TestStorageFactory_CanMigrate(t *testing.T) {
	tempDir := t.TempDir()

	// Create a JSON data file
	dataFile := filepath.Join(tempDir, "memories.json")
	err := os.WriteFile(dataFile, []byte(`{"test": {}}`), 0644)
	if err != nil {
		t.Fatalf("Failed to create test data file: %v", err)
	}

	cfg := &config.Config{
		Storage: &config.StorageConfig{
			Type:     "json",
			JSONPath: filepath.Join(tempDir, "memories.json"),
			LogPath:  filepath.Join(tempDir, "operations.log"),
		},
	}

	factory := NewStorageFactory(cfg)
	canMigrate, reason := factory.CanMigrate()
	if reason != "" && !canMigrate {
		t.Logf("Cannot migrate: %s", reason)
	}

	if !canMigrate {
		t.Error("Expected to be able to migrate")
	}
}

func TestStorageFactory_CanMigrate_NoData(t *testing.T) {
	tempDir := t.TempDir() // Empty directory

	cfg := &config.Config{
		Storage: &config.StorageConfig{
			Type:     "json",
			JSONPath: filepath.Join(tempDir, "memories.json"),
			LogPath:  filepath.Join(tempDir, "operations.log"),
		},
	}

	factory := NewStorageFactory(cfg)
	canMigrate, reason := factory.CanMigrate()
	if reason != "" {
		t.Logf("Cannot migrate: %s", reason)
	}

	if canMigrate {
		t.Error("Expected NOT to be able to migrate with no data")
	}
}

func TestStorageFactory_CanMigrate_SQLiteType(t *testing.T) {
	cfg := &config.Config{
		Storage: &config.StorageConfig{
			Type: "sqlite", // Already SQLite
		},
	}

	factory := NewStorageFactory(cfg)
	canMigrate, reason := factory.CanMigrate()
	if reason != "" {
		t.Logf("Cannot migrate: %s", reason)
	}

	if canMigrate {
		t.Error("Expected NOT to be able to migrate when already using SQLite")
	}
}

func TestStorageFactory_GetStorageStats(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.Config{
		Storage: &config.StorageConfig{
			Type:     "json",
			JSONPath: filepath.Join(tempDir, "memories.json"),
			LogPath:  filepath.Join(tempDir, "operations.log"),
		},
	}

	factory := NewStorageFactory(cfg)
	stats, err := factory.GetStorageStats()
	if err != nil {
		t.Fatalf("GetStorageStats failed: %v", err)
	}

	if stats == nil {
		t.Fatal("Expected stats, got nil")
	}

	// Check for expected keys
	expectedKeys := []string{"storage_type", "storage_paths"}
	for _, key := range expectedKeys {
		if _, ok := stats[key]; !ok {
			t.Errorf("Expected key '%s' in stats", key)
		}
	}
}

func TestStorageFactory_UpdateConfigForSQLite(t *testing.T) {
	cfg := &config.Config{
		Storage: &config.StorageConfig{
			Type:     "json",
			JSONPath: "/tmp/test/memories.json",
		},
	}

	factory := NewStorageFactory(cfg)
	factory.UpdateConfigForSQLite()

	if cfg.Storage.Type != "sqlite" {
		t.Errorf("Expected storage type to be 'sqlite', got '%s'", cfg.Storage.Type)
	}
}

// Mock test for MigrateToSQLite - this would require more complex setup
func TestStorageFactory_MigrateToSQLite_NoSource(t *testing.T) {
	tempDir := t.TempDir() // Empty directory

	cfg := &config.Config{
		Storage: &config.StorageConfig{
			Type:       "json",
			JSONPath:   filepath.Join(tempDir, "memories.json"),
			SQLitePath: filepath.Join(tempDir, "memories.db"),
			LogPath:    filepath.Join(tempDir, "operations.log"),
		},
	}

	factory := NewStorageFactory(cfg)
	result, err := factory.MigrateToSQLite()

	// Should handle case where no source data exists
	if err != nil {
		t.Logf("MigrateToSQLite failed as expected: %v", err)
	}
	if result != nil {
		t.Log("MigrateToSQLite returned result")
	}
}
