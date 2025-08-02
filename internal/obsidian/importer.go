package obsidian

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nyasuto/mory/internal/memory"
)

// Importer handles importing Obsidian notes into Mory
type Importer struct {
	parser *Parser
	store  memory.MemoryStore
}

// NewImporter creates a new Obsidian importer
func NewImporter(vaultPath string, store memory.MemoryStore) *Importer {
	return &Importer{
		parser: NewParser(vaultPath),
		store:  store,
	}
}

// ImportOptions represents options for import operation
type ImportOptions struct {
	DryRun          bool              `json:"dry_run"`
	CategoryMapping map[string]string `json:"category_mapping"`
	SkipDuplicates  bool              `json:"skip_duplicates"`
}

// ImportResult represents the result of an import operation
type ImportResult struct {
	TotalFiles       int
	ImportedFiles    int
	SkippedFiles     int
	DuplicateFiles   int
	Errors           []string
	ImportedMemories []*memory.Memory
	DryRun           bool
}

// ImportVault imports all markdown files from the Obsidian vault
func (i *Importer) ImportVault() (*ImportResult, error) {
	return i.ImportVaultWithOptions(&ImportOptions{
		SkipDuplicates: true,
	})
}

// ImportVaultWithOptions imports all markdown files from the Obsidian vault with options
func (i *Importer) ImportVaultWithOptions(opts *ImportOptions) (*ImportResult, error) {
	if opts == nil {
		opts = &ImportOptions{SkipDuplicates: true}
	}

	result := &ImportResult{
		Errors:           make([]string, 0),
		ImportedMemories: make([]*memory.Memory, 0),
		DryRun:           opts.DryRun,
	}

	// Walk through all markdown files in the vault
	err := filepath.Walk(i.parser.vaultPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Error accessing %s: %v", path, err))
			return nil // Continue walking
		}

		// Skip directories and non-markdown files
		if info.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".md") {
			return nil
		}

		result.TotalFiles++

		// Parse the file
		note, err := i.parser.ParseFile(path)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to parse %s: %v", path, err))
			result.SkippedFiles++
			return nil
		}

		// Apply category mapping if provided
		if opts.CategoryMapping != nil {
			if mappedCategory, exists := opts.CategoryMapping[note.Category]; exists {
				note.Category = mappedCategory
			}
		}

		// Convert to memory
		mem := note.ToMemory()

		// Check for duplicates if enabled
		if opts.SkipDuplicates {
			if existing, err := i.store.Get(mem.Key); err == nil && existing != nil {
				result.DuplicateFiles++
				result.SkippedFiles++
				return nil // Skip duplicate
			}
		}

		// For dry run, don't actually save
		if opts.DryRun {
			result.ImportedFiles++
			result.ImportedMemories = append(result.ImportedMemories, mem)
			return nil
		}

		// Save to store
		id, err := i.store.Save(mem)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to save memory for %s: %v", path, err))
			result.SkippedFiles++
			return nil
		}
		mem.ID = id

		result.ImportedFiles++
		result.ImportedMemories = append(result.ImportedMemories, mem)

		return nil
	})

	if err != nil {
		return result, fmt.Errorf("failed to walk vault directory: %w", err)
	}

	return result, nil
}

// ImportFile imports a single markdown file
func (i *Importer) ImportFile(filePath string) (*memory.Memory, error) {
	return i.ImportFileWithOptions(filePath, &ImportOptions{SkipDuplicates: true})
}

// ImportFileWithOptions imports a single markdown file with options
func (i *Importer) ImportFileWithOptions(filePath string, opts *ImportOptions) (*memory.Memory, error) {
	if opts == nil {
		opts = &ImportOptions{SkipDuplicates: true}
	}

	// Check if file exists and is a markdown file
	if !strings.HasSuffix(strings.ToLower(filePath), ".md") {
		return nil, fmt.Errorf("file %s is not a markdown file", filePath)
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file %s does not exist", filePath)
	}

	// Parse the file
	note, err := i.parser.ParseFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	// Apply category mapping if provided
	if opts.CategoryMapping != nil {
		if mappedCategory, exists := opts.CategoryMapping[note.Category]; exists {
			note.Category = mappedCategory
		}
	}

	// Convert to memory
	mem := note.ToMemory()

	// Check for duplicates if enabled
	if opts.SkipDuplicates {
		if existing, err := i.store.Get(mem.Key); err == nil && existing != nil {
			return nil, fmt.Errorf("memory with key '%s' already exists", mem.Key)
		}
	}

	// For dry run, don't actually save
	if opts.DryRun {
		return mem, nil
	}

	// Save to store
	id, err := i.store.Save(mem)
	if err != nil {
		return nil, fmt.Errorf("failed to save memory: %w", err)
	}
	mem.ID = id

	return mem, nil
}

// ImportByCategory imports all files from a specific category (folder)
func (i *Importer) ImportByCategory(category string) (*ImportResult, error) {
	return i.ImportByCategoryWithOptions(category, &ImportOptions{SkipDuplicates: true})
}

// ImportByCategoryWithOptions imports all files from a specific category (folder) with options
func (i *Importer) ImportByCategoryWithOptions(category string, opts *ImportOptions) (*ImportResult, error) {
	if opts == nil {
		opts = &ImportOptions{SkipDuplicates: true}
	}

	result := &ImportResult{
		Errors:           make([]string, 0),
		ImportedMemories: make([]*memory.Memory, 0),
		DryRun:           opts.DryRun,
	}

	categoryPath := filepath.Join(i.parser.vaultPath, category)

	// Check if category directory exists
	if _, err := os.Stat(categoryPath); os.IsNotExist(err) {
		return result, fmt.Errorf("category directory %s does not exist", categoryPath)
	}

	// Walk through the category directory
	err := filepath.Walk(categoryPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Error accessing %s: %v", path, err))
			return nil
		}

		// Skip directories and non-markdown files
		if info.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".md") {
			return nil
		}

		result.TotalFiles++

		// Parse the file
		note, err := i.parser.ParseFile(path)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to parse %s: %v", path, err))
			result.SkippedFiles++
			return nil
		}

		// Apply category mapping if provided
		if opts.CategoryMapping != nil {
			if mappedCategory, exists := opts.CategoryMapping[note.Category]; exists {
				note.Category = mappedCategory
			}
		}

		// Convert to memory
		mem := note.ToMemory()

		// Check for duplicates if enabled
		if opts.SkipDuplicates {
			if existing, err := i.store.Get(mem.Key); err == nil && existing != nil {
				result.DuplicateFiles++
				result.SkippedFiles++
				return nil // Skip duplicate
			}
		}

		// For dry run, don't actually save
		if opts.DryRun {
			result.ImportedFiles++
			result.ImportedMemories = append(result.ImportedMemories, mem)
			return nil
		}

		// Save to store
		id, err := i.store.Save(mem)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to save memory for %s: %v", path, err))
			result.SkippedFiles++
			return nil
		}
		mem.ID = id

		result.ImportedFiles++
		result.ImportedMemories = append(result.ImportedMemories, mem)

		return nil
	})

	if err != nil {
		return result, fmt.Errorf("failed to walk category directory: %w", err)
	}

	return result, nil
}
