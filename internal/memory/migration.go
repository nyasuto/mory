package memory

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// MigrationOptions configures how the migration should be performed
type MigrationOptions struct {
	// Source configuration
	JSONFile string // Path to source JSON file
	LogFile  string // Path to source log file (optional)

	// Target configuration
	SQLitePath string // Path to target SQLite database

	// Migration options
	SkipExisting       bool // Skip memories that already exist in target
	ValidateAfter      bool // Validate migration by comparing counts
	BackupJSON         bool // Create backup of JSON file before migration
	PreserveTimestamps bool // Preserve original timestamps
	BatchSize          int  // Number of memories to process in each batch
}

// DefaultMigrationOptions returns sensible defaults for migration
func DefaultMigrationOptions() *MigrationOptions {
	return &MigrationOptions{
		SkipExisting:       true,
		ValidateAfter:      true,
		BackupJSON:         true,
		PreserveTimestamps: true,
		BatchSize:          1000,
	}
}

// MigrationResult contains statistics about the migration process
type MigrationResult struct {
	TotalMemories     int           `json:"total_memories"`
	MigratedMemories  int           `json:"migrated_memories"`
	SkippedMemories   int           `json:"skipped_memories"`
	FailedMemories    int           `json:"failed_memories"`
	MigrationDuration time.Duration `json:"migration_duration"`
	ValidationPassed  bool          `json:"validation_passed"`
	BackupCreated     bool          `json:"backup_created"`
	BackupPath        string        `json:"backup_path,omitempty"`
	Errors            []string      `json:"errors,omitempty"`
}

// Migrator handles migration from JSON to SQLite storage
type Migrator struct {
	options *MigrationOptions
	result  *MigrationResult
}

// NewMigrator creates a new migration instance
func NewMigrator(options *MigrationOptions) *Migrator {
	if options == nil {
		options = DefaultMigrationOptions()
	}

	return &Migrator{
		options: options,
		result: &MigrationResult{
			Errors: make([]string, 0),
		},
	}
}

// Migrate performs the migration from JSON to SQLite
func (m *Migrator) Migrate() (*MigrationResult, error) {
	startTime := time.Now()
	log.Printf("[Migration] Starting JSON to SQLite migration")
	log.Printf("[Migration] Source: %s", m.options.JSONFile)
	log.Printf("[Migration] Target: %s", m.options.SQLitePath)

	// Validate source file exists
	if _, err := os.Stat(m.options.JSONFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("source JSON file does not exist: %s", m.options.JSONFile)
	}

	// Create backup if requested
	if m.options.BackupJSON {
		if err := m.createBackup(); err != nil {
			log.Printf("[Migration] Warning: failed to create backup: %v", err)
			m.result.Errors = append(m.result.Errors, fmt.Sprintf("backup failed: %v", err))
		}
	}

	// Load source memories
	sourceMemories, err := m.loadSourceMemories()
	if err != nil {
		return nil, fmt.Errorf("failed to load source memories: %w", err)
	}

	m.result.TotalMemories = len(sourceMemories)
	log.Printf("[Migration] Loaded %d memories from JSON", len(sourceMemories))

	// Initialize target SQLite store
	targetStore, err := NewSQLiteMemoryStore(m.options.SQLitePath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize SQLite store: %w", err)
	}
	defer func() {
		if err := targetStore.Close(); err != nil {
			log.Printf("Warning: failed to close target store: %v", err)
		}
	}()

	// Perform migration in batches
	if err := m.migrateMemories(sourceMemories, targetStore); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	// Validate migration if requested
	if m.options.ValidateAfter {
		m.result.ValidationPassed = m.validateMigration(sourceMemories, targetStore)
	}

	m.result.MigrationDuration = time.Since(startTime)
	log.Printf("[Migration] Completed in %v", m.result.MigrationDuration)
	log.Printf("[Migration] Total: %d, Migrated: %d, Skipped: %d, Failed: %d",
		m.result.TotalMemories, m.result.MigratedMemories,
		m.result.SkippedMemories, m.result.FailedMemories)

	return m.result, nil
}

// createBackup creates a backup of the source JSON file
func (m *Migrator) createBackup() error {
	timestamp := time.Now().Format("20060102_150405")
	backupPath := fmt.Sprintf("%s.backup_%s", m.options.JSONFile, timestamp)

	sourceData, err := os.ReadFile(m.options.JSONFile)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	if err := os.WriteFile(backupPath, sourceData, 0644); err != nil {
		return fmt.Errorf("failed to write backup file: %w", err)
	}

	m.result.BackupCreated = true
	m.result.BackupPath = backupPath
	log.Printf("[Migration] Created backup: %s", backupPath)

	return nil
}

// loadSourceMemories loads memories from the JSON file
func (m *Migrator) loadSourceMemories() ([]*Memory, error) {
	data, err := os.ReadFile(m.options.JSONFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read JSON file: %w", err)
	}

	if len(data) == 0 {
		return []*Memory{}, nil
	}

	var memories []*Memory
	if err := json.Unmarshal(data, &memories); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return memories, nil
}

// migrateMemories migrates memories to SQLite store in batches
func (m *Migrator) migrateMemories(sourceMemories []*Memory, targetStore *SQLiteMemoryStore) error {
	batchSize := m.options.BatchSize
	if batchSize <= 0 {
		batchSize = 1000
	}

	for i := 0; i < len(sourceMemories); i += batchSize {
		end := i + batchSize
		if end > len(sourceMemories) {
			end = len(sourceMemories)
		}

		batch := sourceMemories[i:end]
		log.Printf("[Migration] Processing batch %d-%d (%d memories)",
			i+1, end, len(batch))

		if err := m.migrateBatch(batch, targetStore); err != nil {
			log.Printf("[Migration] Batch %d-%d failed: %v", i+1, end, err)
			m.result.Errors = append(m.result.Errors,
				fmt.Sprintf("batch %d-%d failed: %v", i+1, end, err))
			// Continue with next batch
		}
	}

	return nil
}

// migrateBatch migrates a batch of memories
func (m *Migrator) migrateBatch(batch []*Memory, targetStore *SQLiteMemoryStore) error {
	for _, memory := range batch {
		if err := m.migrateMemory(memory, targetStore); err != nil {
			log.Printf("[Migration] Failed to migrate memory %s: %v", memory.ID, err)
			m.result.FailedMemories++
			m.result.Errors = append(m.result.Errors,
				fmt.Sprintf("memory %s failed: %v", memory.ID, err))
			continue
		}
	}

	return nil
}

// migrateMemory migrates a single memory
func (m *Migrator) migrateMemory(memory *Memory, targetStore *SQLiteMemoryStore) error {
	// Check if memory already exists in target
	if m.options.SkipExisting {
		if memory.Key != "" {
			if _, err := targetStore.Get(memory.Key); err == nil {
				log.Printf("[Migration] Skipping existing memory (key): %s", memory.Key)
				m.result.SkippedMemories++
				return nil
			}
		}

		if _, err := targetStore.GetByID(memory.ID); err == nil {
			log.Printf("[Migration] Skipping existing memory (ID): %s", memory.ID)
			m.result.SkippedMemories++
			return nil
		}
	}

	// Preserve timestamps if requested
	if !m.options.PreserveTimestamps {
		now := time.Now()
		memory.CreatedAt = now
		memory.UpdatedAt = now
	}

	// Ensure required fields are set
	if memory.ID == "" {
		memory.ID = GenerateID()
	}
	if memory.Tags == nil {
		memory.Tags = []string{}
	}

	// Save to target store
	_, err := targetStore.Save(memory)
	if err != nil {
		return fmt.Errorf("failed to save memory: %w", err)
	}

	m.result.MigratedMemories++
	return nil
}

// validateMigration validates that the migration was successful
func (m *Migrator) validateMigration(sourceMemories []*Memory, targetStore *SQLiteMemoryStore) bool {
	log.Printf("[Migration] Validating migration...")

	// Check total count
	targetMemories, err := targetStore.List("")
	if err != nil {
		log.Printf("[Migration] Validation failed: could not list target memories: %v", err)
		m.result.Errors = append(m.result.Errors, fmt.Sprintf("validation list failed: %v", err))
		return false
	}

	expectedCount := len(sourceMemories)
	actualCount := len(targetMemories)

	// Account for skipped memories
	expectedMinimum := m.result.MigratedMemories

	if actualCount < expectedMinimum {
		log.Printf("[Migration] Validation failed: expected at least %d memories, found %d",
			expectedMinimum, actualCount)
		m.result.Errors = append(m.result.Errors,
			fmt.Sprintf("count mismatch: expected >= %d, got %d", expectedMinimum, actualCount))
		return false
	}

	// Sample validation: check a few random memories
	sampleSize := 10
	if expectedCount < sampleSize {
		sampleSize = expectedCount
	}

	for i := 0; i < sampleSize && i < len(sourceMemories); i++ {
		sourceMemory := sourceMemories[i]

		var targetMemory *Memory
		var err error

		// Try to find by key first, then by ID
		if sourceMemory.Key != "" {
			targetMemory, err = targetStore.Get(sourceMemory.Key)
		} else {
			targetMemory, err = targetStore.GetByID(sourceMemory.ID)
		}

		if err != nil {
			log.Printf("[Migration] Validation warning: sample memory not found: %s", sourceMemory.ID)
			continue
		}

		// Compare core fields
		if targetMemory.Category != sourceMemory.Category ||
			targetMemory.Value != sourceMemory.Value {
			log.Printf("[Migration] Validation failed: memory content mismatch for %s", sourceMemory.ID)
			m.result.Errors = append(m.result.Errors,
				fmt.Sprintf("content mismatch for memory %s", sourceMemory.ID))
			return false
		}
	}

	log.Printf("[Migration] Validation passed: %d memories validated", actualCount)
	return true
}

// GetResult returns the current migration result
func (m *Migrator) GetResult() *MigrationResult {
	return m.result
}

// MigrateFromJSON is a convenience function for simple migrations
func MigrateFromJSON(jsonFile, sqlitePath string) (*MigrationResult, error) {
	options := DefaultMigrationOptions()
	options.JSONFile = jsonFile
	options.SQLitePath = sqlitePath

	migrator := NewMigrator(options)
	return migrator.Migrate()
}

// MigrateBulk migrates multiple JSON files to a single SQLite database
func MigrateBulk(jsonFiles []string, sqlitePath string) (*MigrationResult, error) {
	if len(jsonFiles) == 0 {
		return nil, fmt.Errorf("no JSON files provided")
	}

	// Initialize target store first
	targetStore, err := NewSQLiteMemoryStore(sqlitePath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize SQLite store: %w", err)
	}
	defer func() {
		if err := targetStore.Close(); err != nil {
			log.Printf("Warning: failed to close target store: %v", err)
		}
	}()

	totalResult := &MigrationResult{
		Errors: make([]string, 0),
	}
	startTime := time.Now()

	log.Printf("[BulkMigration] Starting bulk migration of %d files", len(jsonFiles))

	for i, jsonFile := range jsonFiles {
		log.Printf("[BulkMigration] Processing file %d/%d: %s", i+1, len(jsonFiles), jsonFile)

		options := DefaultMigrationOptions()
		options.JSONFile = jsonFile
		options.SQLitePath = sqlitePath
		options.BackupJSON = false    // Don't backup in bulk operations
		options.ValidateAfter = false // Validate only at the end

		migrator := NewMigrator(options)

		// Load and migrate this file
		sourceMemories, err := migrator.loadSourceMemories()
		if err != nil {
			log.Printf("[BulkMigration] Failed to load %s: %v", jsonFile, err)
			totalResult.Errors = append(totalResult.Errors,
				fmt.Sprintf("file %s: %v", filepath.Base(jsonFile), err))
			continue
		}

		if err := migrator.migrateMemories(sourceMemories, targetStore); err != nil {
			log.Printf("[BulkMigration] Failed to migrate %s: %v", jsonFile, err)
			totalResult.Errors = append(totalResult.Errors,
				fmt.Sprintf("file %s: %v", filepath.Base(jsonFile), err))
		}

		// Accumulate results
		result := migrator.GetResult()
		totalResult.TotalMemories += len(sourceMemories) // Use actual source count
		totalResult.MigratedMemories += result.MigratedMemories
		totalResult.SkippedMemories += result.SkippedMemories
		totalResult.FailedMemories += result.FailedMemories
		totalResult.Errors = append(totalResult.Errors, result.Errors...)
	}

	totalResult.MigrationDuration = time.Since(startTime)

	// Final validation
	targetMemories, err := targetStore.List("")
	if err == nil {
		actualCount := len(targetMemories)
		expectedMinimum := totalResult.MigratedMemories
		totalResult.ValidationPassed = actualCount >= expectedMinimum

		log.Printf("[BulkMigration] Final validation: %d memories in database (expected >= %d)",
			actualCount, expectedMinimum)
	}

	log.Printf("[BulkMigration] Completed in %v", totalResult.MigrationDuration)
	log.Printf("[BulkMigration] Total: %d, Migrated: %d, Skipped: %d, Failed: %d",
		totalResult.TotalMemories, totalResult.MigratedMemories,
		totalResult.SkippedMemories, totalResult.FailedMemories)

	return totalResult, nil
}
