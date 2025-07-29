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

func TestImportVaultWithOptions(t *testing.T) {
	// Create temporary directory structure for test
	tempDir, err := os.MkdirTemp("", "obsidian_options_test")
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
		filepath.Join(personalDir, "personal-note.md"): `---
category: personal
tags: ["diary"]
---
# Personal Note
This is a personal note.`,

		filepath.Join(workDir, "work-note.md"): `# Work Note
This is a work note.`,
	}

	for filePath, content := range files {
		err := os.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to write test file %s: %v", filePath, err)
		}
	}

	t.Run("DryRun", func(t *testing.T) {
		store := NewMockStore()
		importer := NewImporter(tempDir, store)

		opts := &ImportOptions{
			DryRun:         true,
			SkipDuplicates: true,
		}

		result, err := importer.ImportVaultWithOptions(opts)
		if err != nil {
			t.Fatalf("Failed to import vault with dry run: %v", err)
		}

		// Should find files but not save them
		if result.TotalFiles != 2 {
			t.Errorf("Expected 2 total files, got %d", result.TotalFiles)
		}
		if result.ImportedFiles != 2 {
			t.Errorf("Expected 2 imported files in dry run, got %d", result.ImportedFiles)
		}
		if !result.DryRun {
			t.Error("Expected DryRun flag to be true")
		}

		// Verify no memories were actually saved
		memories, err := store.List("")
		if err != nil {
			t.Fatalf("Failed to list memories: %v", err)
		}
		if len(memories) != 0 {
			t.Errorf("Expected 0 memories saved in dry run, got %d", len(memories))
		}
	})

	t.Run("CategoryMapping", func(t *testing.T) {
		store := NewMockStore()
		importer := NewImporter(tempDir, store)

		opts := &ImportOptions{
			SkipDuplicates: true,
			CategoryMapping: map[string]string{
				"personal": "diary",
				"work":     "projects",
			},
		}

		_, err := importer.ImportVaultWithOptions(opts)
		if err != nil {
			t.Fatalf("Failed to import vault with category mapping: %v", err)
		}

		// Verify category mapping was applied
		memories, err := store.List("")
		if err != nil {
			t.Fatalf("Failed to list memories: %v", err)
		}

		categoryCount := make(map[string]int)
		for _, mem := range memories {
			categoryCount[mem.Category]++
		}

		expectedCategories := map[string]int{
			"diary":    1, // personal -> diary
			"projects": 1, // work -> projects
		}

		for category, expectedCount := range expectedCategories {
			if count, exists := categoryCount[category]; !exists || count != expectedCount {
				t.Errorf("Expected %d memories in category '%s', got %d", expectedCount, category, count)
			}
		}
	})

	t.Run("SkipDuplicates", func(t *testing.T) {
		store := NewMockStore()
		importer := NewImporter(tempDir, store)

		// First import
		opts1 := &ImportOptions{SkipDuplicates: true}
		result1, err := importer.ImportVaultWithOptions(opts1)
		if err != nil {
			t.Fatalf("Failed to import vault first time: %v", err)
		}

		if result1.ImportedFiles != 2 {
			t.Errorf("Expected 2 imported files first time, got %d", result1.ImportedFiles)
		}

		// Second import with skip duplicates
		result2, err := importer.ImportVaultWithOptions(opts1)
		if err != nil {
			t.Fatalf("Failed to import vault second time: %v", err)
		}

		if result2.ImportedFiles != 0 {
			t.Errorf("Expected 0 imported files second time (duplicates), got %d", result2.ImportedFiles)
		}
		if result2.DuplicateFiles != 2 {
			t.Errorf("Expected 2 duplicate files, got %d", result2.DuplicateFiles)
		}

		// Third import without skip duplicates
		opts3 := &ImportOptions{SkipDuplicates: false}
		result3, err := importer.ImportVaultWithOptions(opts3)
		if err != nil {
			t.Fatalf("Failed to import vault third time: %v", err)
		}

		if result3.ImportedFiles != 2 {
			t.Errorf("Expected 2 imported files third time (allow duplicates), got %d", result3.ImportedFiles)
		}
		if result3.DuplicateFiles != 0 {
			t.Errorf("Expected 0 duplicate files when not skipping, got %d", result3.DuplicateFiles)
		}
	})
}

func TestImportFileWithOptions(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "obsidian_file_options_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create test markdown file
	testContent := `---
category: test-category
tags: ["test", "import"]
---

# Test File
This file is for testing file import with options.`

	testFile := filepath.Join(tempDir, "test-file.md")
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	t.Run("DryRun", func(t *testing.T) {
		store := NewMockStore()
		importer := NewImporter(tempDir, store)

		opts := &ImportOptions{
			DryRun:         true,
			SkipDuplicates: true,
		}

		mem, err := importer.ImportFileWithOptions(testFile, opts)
		if err != nil {
			t.Fatalf("Failed to import file with dry run: %v", err)
		}

		if mem.Key != "test-file" {
			t.Errorf("Expected memory key 'test-file', got '%s'", mem.Key)
		}

		// Verify no memory was actually saved
		memories, err := store.List("")
		if err != nil {
			t.Fatalf("Failed to list memories: %v", err)
		}
		if len(memories) != 0 {
			t.Errorf("Expected 0 memories saved in dry run, got %d", len(memories))
		}
	})

	t.Run("CategoryMapping", func(t *testing.T) {
		store := NewMockStore()
		importer := NewImporter(tempDir, store)

		opts := &ImportOptions{
			SkipDuplicates: true,
			CategoryMapping: map[string]string{
				"test-category": "mapped-category",
			},
		}

		mem, err := importer.ImportFileWithOptions(testFile, opts)
		if err != nil {
			t.Fatalf("Failed to import file with category mapping: %v", err)
		}

		if mem.Category != "mapped-category" {
			t.Errorf("Expected category 'mapped-category', got '%s'", mem.Category)
		}
	})

	t.Run("SkipDuplicates", func(t *testing.T) {
		store := NewMockStore()
		importer := NewImporter(tempDir, store)

		// First import
		opts1 := &ImportOptions{SkipDuplicates: true}
		_, err := importer.ImportFileWithOptions(testFile, opts1)
		if err != nil {
			t.Fatalf("Failed to import file first time: %v", err)
		}

		// Second import with skip duplicates - should fail
		_, err = importer.ImportFileWithOptions(testFile, opts1)
		if err == nil {
			t.Error("Expected error when importing duplicate file with SkipDuplicates=true")
		}

		// Third import without skip duplicates - should succeed
		opts2 := &ImportOptions{SkipDuplicates: false}
		_, err = importer.ImportFileWithOptions(testFile, opts2)
		if err != nil {
			t.Errorf("Expected no error when importing duplicate file with SkipDuplicates=false: %v", err)
		}
	})
}

func TestImportByCategoryWithOptions(t *testing.T) {
	// Create temporary directory structure for test
	tempDir, err := os.MkdirTemp("", "obsidian_category_options_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create category directory and files
	categoryDir := filepath.Join(tempDir, "projects")
	_ = os.MkdirAll(categoryDir, 0755)

	files := map[string]string{
		filepath.Join(categoryDir, "project1.md"): `---
category: projects
---
# Project 1
Description of project 1.`,

		filepath.Join(categoryDir, "project2.md"): `---
tags: ["project", "work"]
---
# Project 2
Description of project 2.`,
	}

	for filePath, content := range files {
		err := os.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to write test file %s: %v", filePath, err)
		}
	}

	t.Run("DryRun", func(t *testing.T) {
		store := NewMockStore()
		importer := NewImporter(tempDir, store)

		opts := &ImportOptions{
			DryRun:         true,
			SkipDuplicates: true,
		}

		result, err := importer.ImportByCategoryWithOptions("projects", opts)
		if err != nil {
			t.Fatalf("Failed to import category with dry run: %v", err)
		}

		if result.TotalFiles != 2 {
			t.Errorf("Expected 2 total files, got %d", result.TotalFiles)
		}
		if result.ImportedFiles != 2 {
			t.Errorf("Expected 2 imported files in dry run, got %d", result.ImportedFiles)
		}
		if !result.DryRun {
			t.Error("Expected DryRun flag to be true")
		}

		// Verify no memories were actually saved
		memories, err := store.List("")
		if err != nil {
			t.Fatalf("Failed to list memories: %v", err)
		}
		if len(memories) != 0 {
			t.Errorf("Expected 0 memories saved in dry run, got %d", len(memories))
		}
	})

	t.Run("CategoryMapping", func(t *testing.T) {
		store := NewMockStore()
		importer := NewImporter(tempDir, store)

		opts := &ImportOptions{
			SkipDuplicates: true,
			CategoryMapping: map[string]string{
				"projects": "work-projects",
			},
		}

		_, err := importer.ImportByCategoryWithOptions("projects", opts)
		if err != nil {
			t.Fatalf("Failed to import category with mapping: %v", err)
		}

		// Verify all memories have the mapped category
		memories, err := store.List("")
		if err != nil {
			t.Fatalf("Failed to list memories: %v", err)
		}

		for _, mem := range memories {
			if mem.Category != "work-projects" {
				t.Errorf("Expected category 'work-projects', got '%s' for memory %s", mem.Category, mem.Key)
			}
		}
	})
}
