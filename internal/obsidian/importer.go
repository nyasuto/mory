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

// ImportResult represents the result of an import operation
type ImportResult struct {
	TotalFiles       int
	ImportedFiles    int
	SkippedFiles     int
	Errors           []string
	ImportedMemories []*memory.Memory
}

// ImportVault imports all markdown files from the Obsidian vault
func (i *Importer) ImportVault() (*ImportResult, error) {
	result := &ImportResult{
		Errors:           make([]string, 0),
		ImportedMemories: make([]*memory.Memory, 0),
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

		// Convert to memory and save
		mem := note.ToMemory()
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

	// Convert to memory and save
	mem := note.ToMemory()
	id, err := i.store.Save(mem)
	if err != nil {
		return nil, fmt.Errorf("failed to save memory: %w", err)
	}
	mem.ID = id

	return mem, nil
}

// ImportByCategory imports all files from a specific category (folder)
func (i *Importer) ImportByCategory(category string) (*ImportResult, error) {
	result := &ImportResult{
		Errors:           make([]string, 0),
		ImportedMemories: make([]*memory.Memory, 0),
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

		// Parse and import the file
		note, err := i.parser.ParseFile(path)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to parse %s: %v", path, err))
			result.SkippedFiles++
			return nil
		}

		mem := note.ToMemory()
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
