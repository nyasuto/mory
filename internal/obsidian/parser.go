package obsidian

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/nyasuto/mory/internal/memory"
)

// Note represents a parsed Obsidian note
type Note struct {
	Title       string
	Content     string
	Frontmatter map[string]interface{}
	Tags        []string
	Category    string
	FilePath    string
}

// Parser handles parsing of Obsidian markdown files
type Parser struct {
	vaultPath string
}

// NewParser creates a new Obsidian parser
func NewParser(vaultPath string) *Parser {
	return &Parser{
		vaultPath: vaultPath,
	}
}

// ParseFile parses a single markdown file and returns a Note
func (p *Parser) ParseFile(filePath string) (*Note, error) {
	content, err := readFileContent(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	note := &Note{
		FilePath: filePath,
		Title:    strings.TrimSuffix(filepath.Base(filePath), ".md"),
	}

	// Parse frontmatter and content
	if err := p.parseFrontmatter(content, note); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	// Extract tags from content and frontmatter
	p.extractTags(note)

	// Determine category from folder structure
	p.determineCategory(note)

	return note, nil
}

// parseFrontmatter extracts YAML frontmatter from the content
func (p *Parser) parseFrontmatter(content string, note *Note) error {
	lines := strings.Split(content, "\n")

	// Check if file starts with frontmatter
	if len(lines) < 3 || lines[0] != "---" {
		note.Content = content
		return nil
	}

	// Find end of frontmatter
	endIndex := -1
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			endIndex = i
			break
		}
	}

	if endIndex == -1 {
		note.Content = content
		return nil
	}

	// Parse frontmatter (simple key-value pairs)
	note.Frontmatter = make(map[string]interface{})
	for i := 1; i < endIndex; i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			// Handle arrays (simple comma-separated values)
			if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
				value = strings.Trim(value, "[]")
				items := strings.Split(value, ",")
				var cleanItems []string
				for _, item := range items {
					cleanItems = append(cleanItems, strings.Trim(strings.TrimSpace(item), "\""))
				}
				note.Frontmatter[key] = cleanItems
			} else {
				// Remove quotes if present
				value = strings.Trim(value, "\"")
				note.Frontmatter[key] = value
			}
		}
	}

	// Content is everything after frontmatter
	if endIndex+1 < len(lines) {
		note.Content = strings.Join(lines[endIndex+1:], "\n")
	}

	return nil
}

// extractTags extracts tags from both frontmatter and content
func (p *Parser) extractTags(note *Note) {
	var tags []string

	// Extract from frontmatter
	if tagsValue, exists := note.Frontmatter["tags"]; exists {
		switch v := tagsValue.(type) {
		case string:
			tags = append(tags, v)
		case []string:
			tags = append(tags, v...)
		case []interface{}:
			for _, tag := range v {
				if s, ok := tag.(string); ok {
					tags = append(tags, s)
				}
			}
		}
	}

	// Extract hashtags from content
	hashtagRegex := regexp.MustCompile(`#([a-zA-Z0-9_-]+)`)
	matches := hashtagRegex.FindAllStringSubmatch(note.Content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			tags = append(tags, match[1])
		}
	}

	// Remove duplicates
	tagMap := make(map[string]bool)
	for _, tag := range tags {
		tagMap[tag] = true
	}

	note.Tags = make([]string, 0, len(tagMap))
	for tag := range tagMap {
		note.Tags = append(note.Tags, tag)
	}
}

// determineCategory determines the category based on folder structure
func (p *Parser) determineCategory(note *Note) {
	// Check frontmatter first
	if category, exists := note.Frontmatter["category"]; exists {
		if s, ok := category.(string); ok {
			note.Category = s
			return
		}
	}

	// Use folder structure as category
	relPath, err := filepath.Rel(p.vaultPath, note.FilePath)
	if err != nil {
		note.Category = "general"
		return
	}

	dir := filepath.Dir(relPath)
	if dir == "." {
		note.Category = "general"
	} else {
		// Use first folder as category
		parts := strings.Split(dir, string(filepath.Separator))
		note.Category = parts[0]
	}
}

// ToMemory converts a Note to a Memory object
func (n *Note) ToMemory() *memory.Memory {
	return &memory.Memory{
		ID:        generateMemoryID(),
		Category:  n.Category,
		Key:       n.Title,
		Value:     n.Content,
		Tags:      n.Tags,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// generateMemoryID generates a unique memory ID
func generateMemoryID() string {
	return fmt.Sprintf("memory_%d", time.Now().UnixNano())
}

// readFileContent reads the content of a file
func readFileContent(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
