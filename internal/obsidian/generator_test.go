package obsidian

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/nyasuto/mory/internal/memory"
)

// MockMemoryStore is a mock implementation of memory.MemoryStore for testing
type MockMemoryStore struct {
	memories map[string]*memory.Memory
	nextID   int
}

func NewMockMemoryStore() *MockMemoryStore {
	return &MockMemoryStore{
		memories: make(map[string]*memory.Memory),
		nextID:   1,
	}
}

func (m *MockMemoryStore) Save(mem *memory.Memory) (string, error) {
	id := fmt.Sprintf("memory_test_%d", m.nextID)
	m.nextID++
	mem.ID = id
	mem.CreatedAt = time.Now()
	mem.UpdatedAt = time.Now()
	m.memories[id] = mem
	return id, nil
}

func (m *MockMemoryStore) Get(key string) (*memory.Memory, error) {
	// Search by ID first
	if mem, exists := m.memories[key]; exists {
		return mem, nil
	}

	// Search by key
	for _, mem := range m.memories {
		if mem.Key == key {
			return mem, nil
		}
	}

	return nil, &memory.MemoryNotFoundError{Key: key}
}

func (m *MockMemoryStore) List(category string) ([]*memory.Memory, error) {
	result := make([]*memory.Memory, 0, len(m.memories))
	for _, mem := range m.memories {
		if category == "" || mem.Category == category {
			result = append(result, mem)
		}
	}
	return result, nil
}

func (m *MockMemoryStore) Update(mem *memory.Memory) error {
	if _, exists := m.memories[mem.ID]; !exists {
		return &memory.MemoryNotFoundError{Key: mem.ID}
	}
	mem.UpdatedAt = time.Now()
	m.memories[mem.ID] = mem
	return nil
}

func (m *MockMemoryStore) Delete(id string) error {
	if _, exists := m.memories[id]; !exists {
		return &memory.MemoryNotFoundError{Key: id}
	}
	delete(m.memories, id)
	return nil
}

func (m *MockMemoryStore) GetByID(id string) (*memory.Memory, error) {
	if mem, exists := m.memories[id]; exists {
		return mem, nil
	}
	return nil, &memory.MemoryNotFoundError{Key: id}
}

func (m *MockMemoryStore) Search(query memory.SearchQuery) ([]*memory.SearchResult, error) {
	// Simple implementation for testing
	return nil, nil
}

func (m *MockMemoryStore) DeleteByID(id string) error {
	if _, exists := m.memories[id]; !exists {
		return &memory.MemoryNotFoundError{Key: id}
	}
	delete(m.memories, id)
	return nil
}

func (m *MockMemoryStore) LogOperation(log *memory.OperationLog) error {
	// Mock implementation - just return nil for testing
	return nil
}

// Semantic search methods for MemoryStore interface
func (m *MockMemoryStore) SetSemanticEngine(engine memory.SemanticSearchEngine) {
	// Mock implementation - do nothing
}

func (m *MockMemoryStore) GenerateEmbeddings() error {
	// Mock implementation - do nothing
	return nil
}

func (m *MockMemoryStore) GetSemanticStats() map[string]interface{} {
	// Mock implementation - return basic stats
	return map[string]interface{}{
		"semantic_engine_available": false,
		"total_memories":            len(m.memories),
		"memories_with_embeddings":  0,
		"embedding_coverage":        0.0,
	}
}

// Helper function to create test memories
func createTestMemory(category, key, value string, tags []string) *memory.Memory {
	return &memory.Memory{
		Category:  category,
		Key:       key,
		Value:     value,
		Tags:      tags,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func TestNewNoteGenerator(t *testing.T) {
	store := NewMockMemoryStore()
	generator := NewNoteGenerator(store)

	if generator == nil {
		t.Fatal("Expected non-nil generator")
	}

	if generator.store != store {
		t.Fatal("Expected store to be set correctly")
	}

	if generator.templates == nil {
		t.Fatal("Expected templates to be initialized")
	}

	// Check default templates are loaded
	expectedTemplates := []string{"daily", "summary", "report"}
	for _, templateName := range expectedTemplates {
		if _, exists := generator.templates[templateName]; !exists {
			t.Errorf("Expected template '%s' to be loaded", templateName)
		}
	}
}

func TestGenerateNote_BasicDaily(t *testing.T) {
	store := NewMockMemoryStore()
	generator := NewNoteGenerator(store)

	// Add test memories
	mem1 := createTestMemory("work", "project-update", "Updated the API documentation", []string{"documentation", "api"})
	mem2 := createTestMemory("work", "meeting-notes", "Discussed quarterly goals", []string{"meeting", "goals"})

	_, _ = store.Save(mem1)
	_, _ = store.Save(mem2)

	req := GenerateRequest{
		Template:       "daily",
		Category:       "work",
		Title:          "Work Update",
		IncludeRelated: false,
	}

	note, err := generator.GenerateNote(req)
	if err != nil {
		t.Fatalf("Failed to generate note: %v", err)
	}

	// Verify note properties
	if note.Title != "Work Update" {
		t.Errorf("Expected title 'Work Update', got '%s'", note.Title)
	}

	if note.TemplateUsed != "daily" {
		t.Errorf("Expected template 'daily', got '%s'", note.TemplateUsed)
	}

	if note.MemoryCount != 2 {
		t.Errorf("Expected memory count 2, got %d", note.MemoryCount)
	}

	if note.RelatedCount != 0 {
		t.Errorf("Expected related count 0, got %d", note.RelatedCount)
	}

	// Verify content contains expected elements
	if !strings.Contains(note.Content, "Work Update") {
		t.Error("Expected content to contain title")
	}

	if !strings.Contains(note.Content, "project-update") {
		t.Error("Expected content to contain memory key")
	}

	if !strings.Contains(note.Content, "Updated the API documentation") {
		t.Error("Expected content to contain memory value")
	}
}

func TestGenerateNote_SummaryTemplate(t *testing.T) {
	store := NewMockMemoryStore()
	generator := NewNoteGenerator(store)

	// Add test memories
	mem1 := createTestMemory("learning", "golang-basics", "Learned about goroutines", []string{"golang", "concurrency"})
	mem2 := createTestMemory("learning", "design-patterns", "Studied observer pattern", []string{"patterns", "design"})

	_, _ = store.Save(mem1)
	_, _ = store.Save(mem2)

	req := GenerateRequest{
		Template:       "summary",
		Category:       "learning",
		Title:          "Learning Summary",
		IncludeRelated: false,
	}

	note, err := generator.GenerateNote(req)
	if err != nil {
		t.Fatalf("Failed to generate note: %v", err)
	}

	// Verify template-specific content
	if !strings.Contains(note.Content, "learningカテゴリの記憶をまとめたものです") {
		t.Error("Expected summary template specific content")
	}

	if !strings.Contains(note.Content, "| 総記憶数 | 2 |") {
		t.Error("Expected stats table in summary template")
	}
}

func TestGenerateNote_ReportTemplate(t *testing.T) {
	store := NewMockMemoryStore()
	generator := NewNoteGenerator(store)

	// Add test memories
	mem1 := createTestMemory("project", "milestone-1", "Completed user authentication", []string{"milestone", "auth"})
	mem2 := createTestMemory("project", "milestone-2", "Implemented data persistence", []string{"milestone", "database"})

	_, _ = store.Save(mem1)
	_, _ = store.Save(mem2)

	req := GenerateRequest{
		Template:       "report",
		Category:       "project",
		Title:          "Project Report",
		IncludeRelated: false,
	}

	note, err := generator.GenerateNote(req)
	if err != nil {
		t.Fatalf("Failed to generate note: %v", err)
	}

	// Verify report template content
	if !strings.Contains(note.Content, "エグゼクティブサマリー") {
		t.Error("Expected executive summary section")
	}

	if !strings.Contains(note.Content, "推奨事項") {
		t.Error("Expected recommendations section")
	}

	if !strings.Contains(note.Content, "次のステップ") {
		t.Error("Expected next steps section")
	}
}

func TestGenerateNote_WithRelatedMemories(t *testing.T) {
	store := NewMockMemoryStore()
	generator := NewNoteGenerator(store)

	// Add test memories with similar tags
	mem1 := createTestMemory("tech", "golang-study", "Learning Go programming", []string{"golang", "programming"})
	mem2 := createTestMemory("tech", "rust-study", "Learning Rust programming", []string{"rust", "programming"})
	mem3 := createTestMemory("books", "programming-book", "Read Clean Code", []string{"programming", "books"})

	_, _ = store.Save(mem1)
	_, _ = store.Save(mem2)
	_, _ = store.Save(mem3)

	req := GenerateRequest{
		Template:       "daily",
		Category:       "tech",
		Title:          "Tech Learning",
		IncludeRelated: true,
	}

	note, err := generator.GenerateNote(req)
	if err != nil {
		t.Fatalf("Failed to generate note: %v", err)
	}

	// Should find related memories based on shared tags
	if note.RelatedCount == 0 {
		t.Error("Expected to find related memories")
	}

	// Related memories should be included in content
	if !strings.Contains(note.Content, "関連する記憶") {
		t.Error("Expected related memories section in content")
	}
}

func TestGenerateNote_EmptyCategory(t *testing.T) {
	store := NewMockMemoryStore()
	generator := NewNoteGenerator(store)

	// Add some memories but not in the requested category
	mem1 := createTestMemory("work", "task1", "Complete project", []string{"work"})
	_, _ = store.Save(mem1)

	req := GenerateRequest{
		Template:       "daily",
		Category:       "nonexistent",
		Title:          "Empty Category Test",
		IncludeRelated: false,
	}

	note, err := generator.GenerateNote(req)
	if err != nil {
		t.Fatalf("Failed to generate note: %v", err)
	}

	// Should still generate note even with no memories
	if note.MemoryCount != 0 {
		t.Errorf("Expected memory count 0, got %d", note.MemoryCount)
	}

	if !strings.Contains(note.Content, "記憶がありません") {
		t.Error("Expected 'no memories' message in content")
	}
}

func TestGenerateNote_AllCategories(t *testing.T) {
	store := NewMockMemoryStore()
	generator := NewNoteGenerator(store)

	// Add memories in different categories
	mem1 := createTestMemory("work", "task1", "Work task", []string{"work"})
	mem2 := createTestMemory("personal", "task2", "Personal task", []string{"personal"})

	_, _ = store.Save(mem1)
	_, _ = store.Save(mem2)

	req := GenerateRequest{
		Template:       "summary",
		Category:       "", // Empty category means all memories
		Title:          "All Memories",
		IncludeRelated: false,
	}

	note, err := generator.GenerateNote(req)
	if err != nil {
		t.Fatalf("Failed to generate note: %v", err)
	}

	// Should include all memories when no category filter
	if note.MemoryCount != 2 {
		t.Errorf("Expected memory count 2, got %d", note.MemoryCount)
	}
}

func TestGenerateNote_InvalidTemplate(t *testing.T) {
	store := NewMockMemoryStore()
	generator := NewNoteGenerator(store)

	req := GenerateRequest{
		Template: "invalid-template",
		Title:    "Test",
	}

	_, err := generator.GenerateNote(req)
	if err == nil {
		t.Fatal("Expected error for invalid template")
	}

	if !strings.Contains(err.Error(), "template 'invalid-template' not found") {
		t.Errorf("Expected template not found error, got: %v", err)
	}
}

func TestGenerateNote_MissingTitle(t *testing.T) {
	store := NewMockMemoryStore()
	generator := NewNoteGenerator(store)

	req := GenerateRequest{
		Template: "daily",
		Title:    "", // Missing title
	}

	_, err := generator.GenerateNote(req)
	if err == nil {
		t.Fatal("Expected error for missing title")
	}

	if !strings.Contains(err.Error(), "title is required") {
		t.Errorf("Expected title required error, got: %v", err)
	}
}

func TestGenerateNote_MissingTemplate(t *testing.T) {
	store := NewMockMemoryStore()
	generator := NewNoteGenerator(store)

	req := GenerateRequest{
		Template: "", // Missing template
		Title:    "Test",
	}

	_, err := generator.GenerateNote(req)
	if err == nil {
		t.Fatal("Expected error for missing template")
	}

	if !strings.Contains(err.Error(), "template is required") {
		t.Errorf("Expected template required error, got: %v", err)
	}
}

func TestGenerateNote_OutputPath(t *testing.T) {
	store := NewMockMemoryStore()
	generator := NewNoteGenerator(store)

	mem1 := createTestMemory("test", "key1", "value1", []string{"tag1"})
	_, _ = store.Save(mem1)

	expectedPath := "notes/test-note.md"
	req := GenerateRequest{
		Template:   "daily",
		Title:      "Test Note",
		OutputPath: expectedPath,
	}

	note, err := generator.GenerateNote(req)
	if err != nil {
		t.Fatalf("Failed to generate note: %v", err)
	}

	if note.OutputPath != expectedPath {
		t.Errorf("Expected output path '%s', got '%s'", expectedPath, note.OutputPath)
	}
}

func TestGenerateNote_DefaultOutputPath(t *testing.T) {
	store := NewMockMemoryStore()
	generator := NewNoteGenerator(store)

	mem1 := createTestMemory("test", "key1", "value1", []string{"tag1"})
	_, _ = store.Save(mem1)

	req := GenerateRequest{
		Template: "daily",
		Title:    "Test Note",
		// No OutputPath specified
	}

	note, err := generator.GenerateNote(req)
	if err != nil {
		t.Fatalf("Failed to generate note: %v", err)
	}

	expectedPattern := "daily-Test-Note.md"
	if note.OutputPath != expectedPattern {
		t.Errorf("Expected output path '%s', got '%s'", expectedPattern, note.OutputPath)
	}
}

func TestFindRelatedMemories(t *testing.T) {
	store := NewMockMemoryStore()
	generator := NewNoteGenerator(store)

	// Add memories with shared tags and categories
	mem1 := createTestMemory("tech", "golang", "Learning Go", []string{"programming", "language"})
	mem2 := createTestMemory("tech", "rust", "Learning Rust", []string{"programming", "language"})
	mem3 := createTestMemory("books", "go-book", "Read Go programming book", []string{"programming", "books"})
	mem4 := createTestMemory("unrelated", "cooking", "Made pasta", []string{"food", "cooking"})

	_, _ = store.Save(mem1)
	_, _ = store.Save(mem2)
	_, _ = store.Save(mem3)
	_, _ = store.Save(mem4)

	// Test finding related memories
	inputMemories := []*memory.Memory{mem1}
	related := generator.findRelatedMemories(inputMemories, 5)

	// Should find memories with shared category or tags
	foundRust := false
	foundBook := false
	foundCooking := false

	for _, mem := range related {
		switch mem.Key {
		case "rust":
			foundRust = true
		case "go-book":
			foundBook = true
		case "cooking":
			foundCooking = true
		}
	}

	if !foundRust {
		t.Error("Expected to find 'rust' memory (same category)")
	}

	if !foundBook {
		t.Error("Expected to find 'go-book' memory (shared tags)")
	}

	if foundCooking {
		t.Error("Should not find 'cooking' memory (no relation)")
	}
}

func TestTemplateVariables(t *testing.T) {
	templates := getDefaultTemplates()

	for name, template := range templates {
		// Verify template has required fields
		if template.Name == "" {
			t.Errorf("Template '%s' has empty name", name)
		}

		if template.Description == "" {
			t.Errorf("Template '%s' has empty description", name)
		}

		if template.Content == "" {
			t.Errorf("Template '%s' has empty content", name)
		}

		// Verify title variable exists and is required
		titleVarFound := false
		for _, variable := range template.Variables {
			if variable.Name == "title" {
				titleVarFound = true
				if !variable.Required {
					t.Errorf("Template '%s' title variable should be required", name)
				}
			}
		}

		if !titleVarFound {
			t.Errorf("Template '%s' should have title variable", name)
		}
	}
}
