package obsidian

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nyasuto/mory/internal/memory"
)

// NoteGenerator handles generating Obsidian notes from Mory memories
type NoteGenerator struct {
	store     memory.MemoryStore
	templates map[string]*Template
}

// NewNoteGenerator creates a new note generator
func NewNoteGenerator(store memory.MemoryStore) *NoteGenerator {
	return &NoteGenerator{
		store:     store,
		templates: getDefaultTemplates(),
	}
}

// GenerateRequest represents a request to generate a note
type GenerateRequest struct {
	Template       string            `json:"template"`        // Template type (daily, summary, report)
	Category       string            `json:"category"`        // Category filter for memories
	Title          string            `json:"title"`           // Generated note title
	OutputPath     string            `json:"output_path"`     // Output file path (optional)
	IncludeRelated bool              `json:"include_related"` // Include related memories
	CustomData     map[string]string `json:"custom_data"`     // Custom template data
}

// GeneratedNote represents a generated note
type GeneratedNote struct {
	Title        string           `json:"title"`
	Content      string           `json:"content"`
	OutputPath   string           `json:"output_path"`
	MemoryCount  int              `json:"memory_count"`
	RelatedCount int              `json:"related_count"`
	UsedMemories []*memory.Memory `json:"used_memories"`
	RelatedLinks []string         `json:"related_links"`
	GeneratedAt  time.Time        `json:"generated_at"`
	TemplateUsed string           `json:"template_used"`
}

// TemplateData represents data passed to templates
type TemplateData struct {
	Title           string            `json:"title"`
	Date            string            `json:"date"`
	Memories        []*memory.Memory  `json:"memories"`
	RelatedMemories []*memory.Memory  `json:"related_memories"`
	Category        string            `json:"category"`
	CustomData      map[string]string `json:"custom_data"`
	Stats           map[string]int    `json:"stats"`
}

// GenerateNote generates an Obsidian note based on the request
func (g *NoteGenerator) GenerateNote(req GenerateRequest) (*GeneratedNote, error) {
	// Validate request
	if req.Template == "" {
		return nil, fmt.Errorf("template is required")
	}
	if req.Title == "" {
		return nil, fmt.Errorf("title is required")
	}

	// Get template
	template, exists := g.templates[req.Template]
	if !exists {
		return nil, fmt.Errorf("template '%s' not found", req.Template)
	}

	// Get memories
	var memories []*memory.Memory
	var err error

	// List method takes category parameter, empty string means all memories
	memories, err = g.store.List(req.Category)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve memories: %w", err)
	}

	// Find related memories if requested
	var relatedMemories []*memory.Memory
	if req.IncludeRelated && len(memories) > 0 {
		relatedMemories = g.findRelatedMemories(memories, 10)
	}

	// Prepare template data
	data := TemplateData{
		Title:           req.Title,
		Date:            time.Now().Format("2006-01-02"),
		Memories:        memories,
		RelatedMemories: relatedMemories,
		Category:        req.Category,
		CustomData:      req.CustomData,
		Stats: map[string]int{
			"total_memories":   len(memories),
			"related_memories": len(relatedMemories),
		},
	}

	// Generate content
	content, err := g.applyTemplate(template, data)
	if err != nil {
		return nil, fmt.Errorf("failed to apply template: %w", err)
	}

	// Determine output path
	outputPath := req.OutputPath
	if outputPath == "" {
		// Generate default path based on template and title
		filename := fmt.Sprintf("%s-%s.md", req.Template, strings.ReplaceAll(req.Title, " ", "-"))
		outputPath = filename
	}

	// Create related links
	relatedLinks := make([]string, 0, len(relatedMemories))
	for _, mem := range relatedMemories {
		if mem.Key != "" {
			relatedLinks = append(relatedLinks, fmt.Sprintf("[[%s]]", mem.Key))
		}
	}

	return &GeneratedNote{
		Title:        req.Title,
		Content:      content,
		OutputPath:   outputPath,
		MemoryCount:  len(memories),
		RelatedCount: len(relatedMemories),
		UsedMemories: memories,
		RelatedLinks: relatedLinks,
		GeneratedAt:  time.Now(),
		TemplateUsed: req.Template,
	}, nil
}

// findRelatedMemories finds memories related to the given memories
func (g *NoteGenerator) findRelatedMemories(memories []*memory.Memory, limit int) []*memory.Memory {
	// Simple implementation: find memories with similar tags or categories
	tagMap := make(map[string]int)
	categoryMap := make(map[string]int)

	// Count tags and categories from input memories
	for _, mem := range memories {
		categoryMap[mem.Category]++
		for _, tag := range mem.Tags {
			tagMap[tag]++
		}
	}

	// Get all memories to search through
	allMemories, err := g.store.List("") // Empty category means all memories
	if err != nil {
		return nil
	}

	// Score memories based on similarity
	type scoredMemory struct {
		memory *memory.Memory
		score  int
	}

	scored := make([]scoredMemory, 0)
	memoryIDs := make(map[string]bool)

	// Create a map of existing memory IDs to avoid duplicates
	for _, mem := range memories {
		memoryIDs[mem.ID] = true
	}

	for _, mem := range allMemories {
		// Skip if already in the input list
		if memoryIDs[mem.ID] {
			continue
		}

		score := 0
		// Score based on category similarity
		if categoryMap[mem.Category] > 0 {
			score += 3
		}

		// Score based on tag similarity
		for _, tag := range mem.Tags {
			if tagMap[tag] > 0 {
				score += 2
			}
		}

		if score > 0 {
			scored = append(scored, scoredMemory{mem, score})
		}
	}

	// Sort by score (simple selection for top items)
	result := make([]*memory.Memory, 0, limit)
	for i := 0; i < limit && len(scored) > 0; i++ {
		maxIdx := 0
		maxScore := scored[0].score

		for j, item := range scored {
			if item.score > maxScore {
				maxScore = item.score
				maxIdx = j
			}
		}

		result = append(result, scored[maxIdx].memory)
		// Remove the selected item
		scored = append(scored[:maxIdx], scored[maxIdx+1:]...)
	}

	return result
}

// applyTemplate applies a template to the given data
func (g *NoteGenerator) applyTemplate(template *Template, data TemplateData) (string, error) {
	content := template.Content

	// Simple template replacement (basic implementation)
	content = strings.ReplaceAll(content, "{{title}}", data.Title)
	content = strings.ReplaceAll(content, "{{date}}", data.Date)
	content = strings.ReplaceAll(content, "{{category}}", data.Category)

	// Replace custom data
	for key, value := range data.CustomData {
		placeholder := fmt.Sprintf("{{%s}}", key)
		content = strings.ReplaceAll(content, placeholder, value)
	}

	// Replace stats
	for key, value := range data.Stats {
		placeholder := fmt.Sprintf("{{stats.%s}}", key)
		content = strings.ReplaceAll(content, placeholder, fmt.Sprintf("%d", value))
	}

	// Handle memories section
	if strings.Contains(content, "{{#memories}}") {
		memoriesSection := g.buildMemoriesSection(data.Memories)
		content = g.replaceSection(content, "memories", memoriesSection)
	}

	// Handle related memories section
	if strings.Contains(content, "{{#related_memories}}") {
		relatedSection := g.buildRelatedMemoriesSection(data.RelatedMemories)
		content = g.replaceSection(content, "related_memories", relatedSection)
	}

	return content, nil
}

// buildMemoriesSection builds the memories section content
func (g *NoteGenerator) buildMemoriesSection(memories []*memory.Memory) string {
	if len(memories) == 0 {
		return "記憶がありません。"
	}

	var builder strings.Builder
	for _, mem := range memories {
		builder.WriteString(fmt.Sprintf("- **%s**: %s\n", mem.Category, mem.Value))
		if mem.Key != "" {
			builder.WriteString(fmt.Sprintf("  - キー: %s\n", mem.Key))
		}
		if len(mem.Tags) > 0 {
			builder.WriteString(fmt.Sprintf("  - タグ: %s\n", strings.Join(mem.Tags, ", ")))
		}
		builder.WriteString("\n")
	}
	return builder.String()
}

// buildRelatedMemoriesSection builds the related memories section content
func (g *NoteGenerator) buildRelatedMemoriesSection(memories []*memory.Memory) string {
	if len(memories) == 0 {
		return "関連する記憶がありません。"
	}

	var builder strings.Builder
	for _, mem := range memories {
		// Truncate value if too long
		value := mem.Value
		if len(value) > 100 {
			value = value[:97] + "..."
		}

		if mem.Key != "" {
			builder.WriteString(fmt.Sprintf("- [[%s]]: %s\n", mem.Key, value))
		} else {
			builder.WriteString(fmt.Sprintf("- **%s**: %s\n", mem.Category, value))
		}
	}
	return builder.String()
}

// replaceSection replaces a template section with content
func (g *NoteGenerator) replaceSection(template, sectionName, content string) string {
	startTag := fmt.Sprintf("{{#%s}}", sectionName)
	endTag := fmt.Sprintf("{{/%s}}", sectionName)

	startIdx := strings.Index(template, startTag)
	endIdx := strings.Index(template, endTag)

	if startIdx == -1 || endIdx == -1 {
		return template
	}

	endIdx += len(endTag)

	before := template[:startIdx]
	after := template[endIdx:]

	return before + content + after
}

// WriteNote writes a generated note to a file
func (g *NoteGenerator) WriteNote(note *GeneratedNote, outputDir string) error {
	if outputDir == "" {
		return fmt.Errorf("output directory is required")
	}

	fullPath := filepath.Join(outputDir, note.OutputPath)

	// Create directory if it doesn't exist
	dir := filepath.Dir(fullPath)
	if err := ensureDir(dir); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write content to file
	if err := writeFile(fullPath, note.Content); err != nil {
		return fmt.Errorf("failed to write note file: %w", err)
	}

	return nil
}

// ensureDir creates a directory if it doesn't exist
func ensureDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}

// writeFile writes content to a file
func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}
