package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/nyasuto/mory/internal/memory"
)

func main() {
	var (
		jsonFile   = flag.String("json", "data/memories.json", "Path to source JSON file")
		sqlitePath = flag.String("sqlite", "data/memories.db", "Path to target SQLite database")
		backup     = flag.Bool("backup", true, "Create backup of JSON file")
		validate   = flag.Bool("validate", true, "Validate migration after completion")
		batchSize  = flag.Int("batch", 1000, "Number of memories to process in each batch")
	)
	flag.Parse()

	fmt.Printf("🔄 Mory JSON to SQLite Migration Tool\n")
	fmt.Printf("Source: %s\n", *jsonFile)
	fmt.Printf("Target: %s\n", *sqlitePath)
	fmt.Printf("Backup: %v\n", *backup)
	fmt.Printf("Validate: %v\n", *validate)
	fmt.Printf("Batch Size: %d\n\n", *batchSize)

	// Check if source file exists
	if _, err := os.Stat(*jsonFile); os.IsNotExist(err) {
		log.Fatalf("❌ Source JSON file does not exist: %s", *jsonFile)
	}

	// Create migration options
	options := memory.DefaultMigrationOptions()
	options.JSONFile = *jsonFile
	options.SQLitePath = *sqlitePath
	options.BackupJSON = *backup
	options.ValidateAfter = *validate
	options.BatchSize = *batchSize

	// Perform migration
	migrator := memory.NewMigrator(options)
	result, err := migrator.Migrate()
	if err != nil {
		log.Fatalf("❌ Migration failed: %v", err)
	}

	// Display results
	fmt.Printf("✅ Migration completed successfully!\n\n")
	fmt.Printf("📊 Migration Statistics:\n")
	fmt.Printf("  Total memories: %d\n", result.TotalMemories)
	fmt.Printf("  Migrated: %d\n", result.MigratedMemories)
	fmt.Printf("  Skipped: %d\n", result.SkippedMemories)
	fmt.Printf("  Failed: %d\n", result.FailedMemories)
	fmt.Printf("  Duration: %v\n", result.MigrationDuration)
	fmt.Printf("  Validation: %v\n", result.ValidationPassed)

	if result.BackupCreated {
		fmt.Printf("  Backup: %s\n", result.BackupPath)
	}

	if len(result.Errors) > 0 {
		fmt.Printf("\n⚠️  Errors encountered:\n")
		for _, err := range result.Errors {
			fmt.Printf("  - %s\n", err)
		}
	}

	fmt.Printf("\n🎉 SQLite database is ready at: %s\n", *sqlitePath)
}
