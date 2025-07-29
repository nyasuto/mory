package obsidian

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/nyasuto/mory/internal/memory"
)

// MockStore implements memory.MemoryStore for testing
type MockStore struct {
	memories map[string]*memory.Memory
	saveErr  error
}

func NewMockStore() *MockStore {
	return &MockStore{
		memories: make(map[string]*memory.Memory),
	}
}

func (m *MockStore) Save(mem *memory.Memory) (string, error) {
	if m.saveErr != nil {
		return "", m.saveErr
	}
	if mem.ID == "" {
		mem.ID = memory.GenerateID()
	}
	m.memories[mem.ID] = mem
	return mem.ID, nil
}

func (m *MockStore) Get(key string) (*memory.Memory, error) {
	for _, mem := range m.memories {
		if mem.Key == key {
			return mem, nil
		}
	}
	return nil, &memory.MemoryNotFoundError{Key: key}
}

func (m *MockStore) GetByID(id string) (*memory.Memory, error) {
	if mem, exists := m.memories[id]; exists {
		return mem, nil
	}
	return nil, &memory.MemoryNotFoundError{Key: id}
}

func (m *MockStore) List(category string) ([]*memory.Memory, error) {
	var result []*memory.Memory
	for _, mem := range m.memories {
		if category == "" || mem.Category == category {
			result = append(result, mem)
		}
	}
	return result, nil
}

func (m *MockStore) Delete(key string) error {
	for id, mem := range m.memories {
		if mem.Key == key {
			delete(m.memories, id)
			return nil
		}
	}
	return &memory.MemoryNotFoundError{Key: key}
}

func (m *MockStore) DeleteByID(id string) error {
	if _, exists := m.memories[id]; exists {
		delete(m.memories, id)
		return nil
	}
	return &memory.MemoryNotFoundError{Key: id}
}

func (m *MockStore) Search(query memory.SearchQuery) ([]*memory.SearchResult, error) {
	return nil, nil
}

func (m *MockStore) LogOperation(log *memory.OperationLog) error {
	return nil
}

func TestNewImporter(t *testing.T) {
	vaultPath := "/test/vault"
	store := NewMockStore()
	importer := NewImporter(vaultPath, store)

	if importer.parser.vaultPath != vaultPath {
		t.Errorf("Expected vault path %s, got %s", vaultPath, importer.parser.vaultPath)
	}
	if importer.store != store {
		t.Error("Expected store to be set correctly")
	}
}

func TestImportFile(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "obsidian_import_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create test markdown file
	testContent := `---
category: test-import
tags: ["test", "import"]
---

# Import Test

This file is for testing import functionality.`

	testFile := filepath.Join(tempDir, "import-test.md")
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test successful import
	store := NewMockStore()
	importer := NewImporter(tempDir, store)

	mem, err := importer.ImportFile(testFile)
	if err != nil {
		t.Fatalf("Failed to import file: %v", err)
	}

	if mem.Key != "import-test" {
		t.Errorf("Expected memory key 'import-test', got '%s'", mem.Key)
	}
	if mem.Category != "test-import" {
		t.Errorf("Expected category 'test-import', got '%s'", mem.Category)
	}

	// Verify memory was saved to store
	savedMem, err := store.GetByID(mem.ID)
	if err != nil {
		t.Errorf("Failed to get saved memory: %v", err)
	}
	if savedMem.Key != mem.Key {
		t.Errorf("Saved memory key mismatch: expected %s, got %s", mem.Key, savedMem.Key)
	}

	// Test non-markdown file
	nonMdFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(nonMdFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to write non-md file: %v", err)
	}

	_, err = importer.ImportFile(nonMdFile)
	if err == nil {
		t.Error("Expected error for non-markdown file")
	}

	// Test non-existent file
	_, err = importer.ImportFile(filepath.Join(tempDir, "nonexistent.md"))
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestImportVault(t *testing.T) {
	// Create temporary directory structure for test
	tempDir, err := os.MkdirTemp("", "obsidian_vault_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create directory structure
	personalDir := filepath.Join(tempDir, "personal")
	workDir := filepath.Join(tempDir, "work")
	_ = os.MkdirAll(personalDir, 0755)
	_ = os.MkdirAll(workDir, 0755)

	// Create test files
	files := map[string]string{
		filepath.Join(tempDir, "root-note.md"): `---
category: general
---
# Root Note
This is a root level note.`,

		filepath.Join(personalDir, "personal-note.md"): `---
tags: ["personal", "diary"]
---
# Personal Note
This is a personal note.`,

		filepath.Join(workDir, "work-note.md"): `# Work Note
This is a work note without frontmatter.`,

		filepath.Join(tempDir, "non-markdown.txt"): "This should be ignored",
	}

	for filePath, content := range files {
		err := os.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to write test file %s: %v", filePath, err)
		}
	}

	// Test vault import
	store := NewMockStore()
	importer := NewImporter(tempDir, store)

	result, err := importer.ImportVault()
	if err != nil {
		t.Fatalf("Failed to import vault: %v", err)
	}

	// Should have found 3 markdown files and imported all of them
	expectedFiles := 3
	if result.TotalFiles != expectedFiles {
		t.Errorf("Expected %d total files, got %d", expectedFiles, result.TotalFiles)
	}
	if result.ImportedFiles != expectedFiles {
		t.Errorf("Expected %d imported files, got %d", expectedFiles, result.ImportedFiles)
	}
	if result.SkippedFiles != 0 {
		t.Errorf("Expected 0 skipped files, got %d", result.SkippedFiles)
	}
	if len(result.Errors) != 0 {
		t.Errorf("Expected 0 errors, got %d: %v", len(result.Errors), result.Errors)
	}

	// Verify memories were created with correct categories
	memories, err := store.List("")
	if err != nil {
		t.Fatalf("Failed to list memories: %v", err)
	}
	if len(memories) != expectedFiles {
		t.Errorf("Expected %d memories in store, got %d", expectedFiles, len(memories))
	}

	// Check categories are set correctly
	categoryCount := make(map[string]int)
	for _, mem := range memories {
		categoryCount[mem.Category]++
	}

	expectedCategories := map[string]int{
		"general":  1, // root-note.md
		"personal": 1, // personal-note.md
		"work":     1, // work-note.md
	}

	for category, expectedCount := range expectedCategories {
		if count, exists := categoryCount[category]; !exists || count != expectedCount {
			t.Errorf("Expected %d memories in category '%s', got %d", expectedCount, category, count)
		}
	}
}

func TestImportByCategory(t *testing.T) {
	// Create temporary directory structure for test
	tempDir, err := os.MkdirTemp("", "obsidian_category_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create category directory and files
	categoryDir := filepath.Join(tempDir, "projects")
	_ = os.MkdirAll(categoryDir, 0755)

	files := map[string]string{
		filepath.Join(categoryDir, "project1.md"): `# Project 1
Description of project 1.`,

		filepath.Join(categoryDir, "project2.md"): `---
tags: ["project", "work"]
---
# Project 2
Description of project 2.`,

		filepath.Join(tempDir, "outside-category.md"): `# Outside
This file is outside the category.`,
	}

	for filePath, content := range files {
		err := os.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to write test file %s: %v", filePath, err)
		}
	}

	// Test category import
	store := NewMockStore()
	importer := NewImporter(tempDir, store)

	result, err := importer.ImportByCategory("projects")
	if err != nil {
		t.Fatalf("Failed to import category: %v", err)
	}

	// Should have found and imported 2 files from the projects category
	expectedFiles := 2
	if result.TotalFiles != expectedFiles {
		t.Errorf("Expected %d total files, got %d", expectedFiles, result.TotalFiles)
	}
	if result.ImportedFiles != expectedFiles {
		t.Errorf("Expected %d imported files, got %d", expectedFiles, result.ImportedFiles)
	}

	// Verify all imported memories have the correct category
	memories, err := store.List("")
	if err != nil {
		t.Fatalf("Failed to list memories: %v", err)
	}

	for _, mem := range memories {
		if mem.Category != "projects" {
			t.Errorf("Expected category 'projects', got '%s' for memory %s", mem.Category, mem.Key)
		}
	}

	// Test non-existent category
	_, err = importer.ImportByCategory("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent category")
	}
}

func TestImportWithErrors(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "obsidian_error_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create a valid markdown file
	validFile := filepath.Join(tempDir, "valid.md")
	err = os.WriteFile(validFile, []byte("# Valid Note\nContent"), 0644)
	if err != nil {
		t.Fatalf("Failed to write valid file: %v", err)
	}

	// Test with store that returns errors
	store := NewMockStore()
	store.saveErr = errors.New("simulated save error")

	importer := NewImporter(tempDir, store)

	result, err := importer.ImportVault()
	if err != nil {
		t.Fatalf("ImportVault should not return error even if individual saves fail: %v", err)
	}

	// Should have 1 file found but 0 imported due to save error
	if result.TotalFiles != 1 {
		t.Errorf("Expected 1 total file, got %d", result.TotalFiles)
	}
	if result.ImportedFiles != 0 {
		t.Errorf("Expected 0 imported files due to save error, got %d", result.ImportedFiles)
	}
	if result.SkippedFiles != 1 {
		t.Errorf("Expected 1 skipped file due to save error, got %d", result.SkippedFiles)
	}
	if len(result.Errors) == 0 {
		t.Error("Expected errors to be recorded")
	}
}
