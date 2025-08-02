package obsidian

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewParser(t *testing.T) {
	vaultPath := "/test/vault"
	parser := NewParser(vaultPath)

	if parser.vaultPath != vaultPath {
		t.Errorf("Expected vault path %s, got %s", vaultPath, parser.vaultPath)
	}
}

func TestParseFrontmatter(t *testing.T) {
	parser := NewParser("/test/vault")

	tests := []struct {
		name          string
		content       string
		expectedNote  Note
		expectedError bool
	}{
		{
			name: "Valid frontmatter",
			content: `---
category: test
tags: ["tag1", "tag2"]
---
# Test Content

This is a test note.`,
			expectedNote: Note{
				Content: `# Test Content

This is a test note.`,
				Frontmatter: map[string]interface{}{
					"category": "test",
					"tags":     []string{"tag1", "tag2"},
				},
			},
			expectedError: false,
		},
		{
			name: "No frontmatter",
			content: `# Test Content

This is a test note without frontmatter.`,
			expectedNote: Note{
				Content: `# Test Content

This is a test note without frontmatter.`,
				Frontmatter: nil,
			},
			expectedError: false,
		},
		{
			name: "Simple key-value frontmatter",
			content: `---
title: Test Note
category: personal
author: John Doe
---

Content goes here.`,
			expectedNote: Note{
				Content: `
Content goes here.`,
				Frontmatter: map[string]interface{}{
					"title":    "Test Note",
					"category": "personal",
					"author":   "John Doe",
				},
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			note := &Note{}
			err := parser.parseFrontmatter(tt.content, note)

			if tt.expectedError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if note.Content != tt.expectedNote.Content {
				t.Errorf("Expected content %q, got %q", tt.expectedNote.Content, note.Content)
			}

			if tt.expectedNote.Frontmatter != nil {
				if note.Frontmatter == nil {
					t.Error("Expected frontmatter but got nil")
				} else {
					for key, expectedValue := range tt.expectedNote.Frontmatter {
						actualValue, exists := note.Frontmatter[key]
						if !exists {
							t.Errorf("Expected key %s in frontmatter", key)
							continue
						}

						// Handle string comparison
						switch ev := expectedValue.(type) {
						case string:
							if av, ok := actualValue.(string); !ok || av != ev {
								t.Errorf("Expected frontmatter[%s] = %v, got %v", key, ev, actualValue)
							}
						case []string:
							av, ok := actualValue.([]string)
							if !ok {
								t.Errorf("Expected frontmatter[%s] to be []string, got %T", key, actualValue)
								continue
							}
							if len(av) != len(ev) {
								t.Errorf("Expected frontmatter[%s] length %d, got %d", key, len(ev), len(av))
								continue
							}
							for i, expectedTag := range ev {
								if i >= len(av) || av[i] != expectedTag {
									t.Errorf("Expected frontmatter[%s][%d] = %s, got %s", key, i, expectedTag, av[i])
								}
							}
						}
					}
				}
			}
		})
	}
}

func TestExtractTags(t *testing.T) {
	parser := NewParser("/test/vault")

	tests := []struct {
		name         string
		note         Note
		expectedTags []string
	}{
		{
			name: "Tags from frontmatter and content",
			note: Note{
				Content: "This is a #test note with #hashtags.",
				Frontmatter: map[string]interface{}{
					"tags": []string{"frontmatter-tag"},
				},
			},
			expectedTags: []string{"frontmatter-tag", "test", "hashtags"},
		},
		{
			name: "Tags from content only",
			note: Note{
				Content: "This note has #programming and #golang tags.",
			},
			expectedTags: []string{"programming", "golang"},
		},
		{
			name: "No tags",
			note: Note{
				Content: "This note has no tags.",
			},
			expectedTags: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser.extractTags(&tt.note)

			if len(tt.note.Tags) != len(tt.expectedTags) {
				t.Errorf("Expected %d tags, got %d", len(tt.expectedTags), len(tt.note.Tags))
			}

			// Create a map for easy lookup since order doesn't matter
			tagMap := make(map[string]bool)
			for _, tag := range tt.note.Tags {
				tagMap[tag] = true
			}

			for _, expectedTag := range tt.expectedTags {
				if !tagMap[expectedTag] {
					t.Errorf("Expected tag %s not found", expectedTag)
				}
			}
		})
	}
}

func TestDetermineCategory(t *testing.T) {
	vaultPath := "/test/vault"
	parser := NewParser(vaultPath)

	tests := []struct {
		name             string
		filePath         string
		frontmatter      map[string]interface{}
		expectedCategory string
	}{
		{
			name:             "Category from frontmatter",
			filePath:         "/test/vault/folder/note.md",
			frontmatter:      map[string]interface{}{"category": "work"},
			expectedCategory: "work",
		},
		{
			name:             "Category from folder structure",
			filePath:         "/test/vault/personal/daily/note.md",
			frontmatter:      nil,
			expectedCategory: "personal",
		},
		{
			name:             "Root level note",
			filePath:         "/test/vault/note.md",
			frontmatter:      nil,
			expectedCategory: "general",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			note := &Note{
				FilePath:    tt.filePath,
				Frontmatter: tt.frontmatter,
			}

			parser.determineCategory(note)

			if note.Category != tt.expectedCategory {
				t.Errorf("Expected category %s, got %s", tt.expectedCategory, note.Category)
			}
		})
	}
}

func TestNoteToMemory(t *testing.T) {
	note := &Note{
		Title:    "Test Note",
		Content:  "This is test content.",
		Category: "test",
		Tags:     []string{"tag1", "tag2"},
	}

	memory := note.ToMemory()

	if memory.Key != note.Title {
		t.Errorf("Expected memory key %s, got %s", note.Title, memory.Key)
	}
	if memory.Value != note.Content {
		t.Errorf("Expected memory value %s, got %s", note.Content, memory.Value)
	}
	if memory.Category != note.Category {
		t.Errorf("Expected memory category %s, got %s", note.Category, memory.Category)
	}
	if len(memory.Tags) != len(note.Tags) {
		t.Errorf("Expected %d tags, got %d", len(note.Tags), len(memory.Tags))
	}
	if memory.ID == "" {
		t.Error("Expected memory ID to be set")
	}
	if memory.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}
	if memory.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt to be set")
	}
}

func TestGenerateMemoryID(t *testing.T) {
	id1 := generateMemoryID()
	time.Sleep(time.Millisecond) // Ensure different timestamps
	id2 := generateMemoryID()

	if id1 == id2 {
		t.Error("Expected different memory IDs")
	}
	if id1 == "" || id2 == "" {
		t.Error("Expected non-empty memory IDs")
	}
	if len(id1) < 10 || len(id2) < 10 {
		t.Error("Expected memory IDs to have reasonable length")
	}
}

// Integration test with actual file creation
func TestParseFileIntegration(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "obsidian_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Create test markdown file
	testContent := `---
category: test-category
tags: ["integration", "test"]
author: test-author
---

# Test Note

This is a test note content.

It has multiple paragraphs and #inline-tags.`

	testFile := filepath.Join(tempDir, "test-note.md")
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Parse the file
	parser := NewParser(tempDir)
	note, err := parser.ParseFile(testFile)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	// Verify results
	if note.Title != "test-note" {
		t.Errorf("Expected title 'test-note', got '%s'", note.Title)
	}
	if note.Category != "test-category" {
		t.Errorf("Expected category 'test-category', got '%s'", note.Category)
	}
	if note.FilePath != testFile {
		t.Errorf("Expected file path '%s', got '%s'", testFile, note.FilePath)
	}

	// Check that content excludes frontmatter
	expectedContent := `
# Test Note

This is a test note content.

It has multiple paragraphs and #inline-tags.`
	if note.Content != expectedContent {
		t.Errorf("Expected content %q, got %q", expectedContent, note.Content)
	}

	// Check tags (should include both frontmatter tags and inline tags)
	expectedTags := map[string]bool{
		"integration": true,
		"test":        true,
		"inline-tags": true,
	}

	if len(note.Tags) != len(expectedTags) {
		t.Errorf("Expected %d tags, got %d: %v", len(expectedTags), len(note.Tags), note.Tags)
	}

	for _, tag := range note.Tags {
		if !expectedTags[tag] {
			t.Errorf("Unexpected tag: %s", tag)
		}
	}
}
