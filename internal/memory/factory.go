package memory

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/nyasuto/mory/internal/config"
)

// StorageFactory creates memory storage instances based on configuration
type StorageFactory struct {
	config *config.Config
}

// NewStorageFactory creates a new storage factory
func NewStorageFactory(cfg *config.Config) *StorageFactory {
	return &StorageFactory{config: cfg}
}

// CreateMemoryStore creates a memory store based on the configuration
func (f *StorageFactory) CreateMemoryStore() (MemoryStore, error) {
	if f.config.Storage == nil {
		log.Printf("[StorageFactory] No storage config found, using default JSON storage")
		return f.createJSONStore()
	}

	storageType := strings.ToLower(f.config.Storage.Type)

	switch storageType {
	case "json":
		return f.createJSONStore()
	case "sqlite":
		return f.createSQLiteStore()
	default:
		log.Printf("[StorageFactory] Unknown storage type '%s', falling back to JSON", storageType)
		return f.createJSONStore()
	}
}

// createJSONStore creates a JSON-based memory store
func (f *StorageFactory) createJSONStore() (MemoryStore, error) {
	var jsonPath, logPath string

	if f.config.Storage != nil && f.config.Storage.JSONPath != "" {
		jsonPath = f.config.Storage.JSONPath
		logPath = f.config.Storage.LogPath
	} else {
		// Fallback to legacy DataPath
		jsonPath = f.config.DataPath
		if jsonPath == "" {
			jsonPath = "data/memories.json"
		}

		// Generate log path based on JSON path
		dir := filepath.Dir(jsonPath)
		logPath = filepath.Join(dir, "operations.log")
	}

	// Ensure directories exist
	if err := f.ensureDirectoryExists(filepath.Dir(jsonPath)); err != nil {
		return nil, fmt.Errorf("failed to create JSON data directory: %w", err)
	}

	if err := f.ensureDirectoryExists(filepath.Dir(logPath)); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	log.Printf("[StorageFactory] Creating JSON memory store: %s", jsonPath)
	store := NewJSONMemoryStore(jsonPath, logPath)

	return store, nil
}

// createSQLiteStore creates a SQLite-based memory store
func (f *StorageFactory) createSQLiteStore() (MemoryStore, error) {
	sqlitePath := f.config.Storage.SQLitePath
	if sqlitePath == "" {
		sqlitePath = "data/memories.db"
	}

	// Ensure directory exists
	if err := f.ensureDirectoryExists(filepath.Dir(sqlitePath)); err != nil {
		return nil, fmt.Errorf("failed to create SQLite data directory: %w", err)
	}

	log.Printf("[StorageFactory] Creating SQLite memory store: %s", sqlitePath)
	store, err := NewSQLiteMemoryStore(sqlitePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create SQLite memory store: %w", err)
	}

	return store, nil
}

// ensureDirectoryExists creates the directory if it doesn't exist
func (f *StorageFactory) ensureDirectoryExists(dir string) error {
	if dir == "" || dir == "." {
		return nil
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
		log.Printf("[StorageFactory] Created directory: %s", dir)
	}

	return nil
}

// GetStorageType returns the configured storage type
func (f *StorageFactory) GetStorageType() string {
	if f.config.Storage == nil || f.config.Storage.Type == "" {
		return "json"
	}
	return strings.ToLower(f.config.Storage.Type)
}

// GetStoragePaths returns the paths for different storage types
func (f *StorageFactory) GetStoragePaths() map[string]string {
	paths := make(map[string]string)

	if f.config.Storage != nil {
		if f.config.Storage.JSONPath != "" {
			paths["json"] = f.config.Storage.JSONPath
		}
		if f.config.Storage.SQLitePath != "" {
			paths["sqlite"] = f.config.Storage.SQLitePath
		}
		if f.config.Storage.LogPath != "" {
			paths["log"] = f.config.Storage.LogPath
		}
	}

	// Add defaults if not specified
	if _, exists := paths["json"]; !exists {
		if f.config.DataPath != "" {
			paths["json"] = f.config.DataPath
		} else {
			paths["json"] = "data/memories.json"
		}
	}

	if _, exists := paths["sqlite"]; !exists {
		paths["sqlite"] = "data/memories.db"
	}

	if _, exists := paths["log"]; !exists {
		dir := filepath.Dir(paths["json"])
		paths["log"] = filepath.Join(dir, "operations.log")
	}

	return paths
}

// MigrateToSQLite migrates from JSON to SQLite storage
func (f *StorageFactory) MigrateToSQLite() (*MigrationResult, error) {
	paths := f.GetStoragePaths()

	jsonPath, jsonExists := paths["json"]
	sqlitePath, sqliteExists := paths["sqlite"]

	if !jsonExists || !sqliteExists {
		return nil, fmt.Errorf("missing required paths for migration")
	}

	// Check if JSON file exists
	if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("source JSON file does not exist: %s", jsonPath)
	}

	log.Printf("[StorageFactory] Starting migration from JSON to SQLite")
	log.Printf("[StorageFactory] Source: %s", jsonPath)
	log.Printf("[StorageFactory] Target: %s", sqlitePath)

	// Create migration options
	options := DefaultMigrationOptions()
	options.JSONFile = jsonPath
	options.SQLitePath = sqlitePath
	options.BackupJSON = true
	options.ValidateAfter = true

	// Perform migration
	migrator := NewMigrator(options)
	result, err := migrator.Migrate()
	if err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	log.Printf("[StorageFactory] Migration completed successfully")
	log.Printf("[StorageFactory] Migrated: %d/%d memories",
		result.MigratedMemories, result.TotalMemories)

	return result, nil
}

// UpdateConfigForSQLite updates the configuration to use SQLite storage
func (f *StorageFactory) UpdateConfigForSQLite() {
	if f.config.Storage == nil {
		f.config.Storage = &config.StorageConfig{}
	}

	f.config.Storage.Type = "sqlite"

	// Set default paths if not already set
	if f.config.Storage.SQLitePath == "" {
		f.config.Storage.SQLitePath = "data/memories.db"
	}
	if f.config.Storage.JSONPath == "" {
		if f.config.DataPath != "" {
			f.config.Storage.JSONPath = f.config.DataPath
		} else {
			f.config.Storage.JSONPath = "data/memories.json"
		}
	}
	if f.config.Storage.LogPath == "" {
		f.config.Storage.LogPath = "data/operations.log"
	}

	log.Printf("[StorageFactory] Updated configuration to use SQLite storage")
}

// CanMigrate checks if migration from JSON to SQLite is possible
func (f *StorageFactory) CanMigrate() (bool, string) {
	paths := f.GetStoragePaths()

	jsonPath, exists := paths["json"]
	if !exists {
		return false, "JSON path not configured"
	}

	if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
		return false, fmt.Sprintf("JSON file does not exist: %s", jsonPath)
	}

	sqlitePath := paths["sqlite"]
	sqliteDir := filepath.Dir(sqlitePath)
	if err := os.MkdirAll(sqliteDir, 0755); err != nil {
		return false, fmt.Sprintf("Cannot create SQLite directory: %v", err)
	}

	return true, ""
}

// GetStorageStats returns statistics about the current storage
func (f *StorageFactory) GetStorageStats() (map[string]interface{}, error) {
	store, err := f.CreateMemoryStore()
	if err != nil {
		return nil, fmt.Errorf("failed to create memory store: %w", err)
	}

	// Close store if it's SQLite to release resources
	if sqliteStore, ok := store.(*SQLiteMemoryStore); ok {
		defer func() {
			if err := sqliteStore.Close(); err != nil {
				log.Printf("Warning: failed to close SQLite store: %v", err)
			}
		}()
	}

	stats := map[string]interface{}{
		"storage_type":  f.GetStorageType(),
		"storage_paths": f.GetStoragePaths(),
	}

	// Get basic counts
	memories, err := store.List("")
	if err != nil {
		log.Printf("[StorageFactory] Warning: failed to list memories for stats: %v", err)
		stats["error"] = err.Error()
	} else {
		stats["total_memories"] = len(memories)

		// Count memories by category
		categories := make(map[string]int)
		for _, memory := range memories {
			categories[memory.Category]++
		}
		stats["categories"] = categories
	}

	// Get semantic stats if available
	semanticStats := store.GetSemanticStats()
	for k, v := range semanticStats {
		stats[k] = v
	}

	return stats, nil
}
