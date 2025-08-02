package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// ObsidianConfig represents Obsidian integration configuration
type ObsidianConfig struct {
	VaultPath    string `json:"vault_path"`
	AutoImport   bool   `json:"auto_import"`
	SyncInterval string `json:"sync_interval"`
	TemplateDir  string `json:"template_dir"`
}

// SemanticConfig represents semantic search configuration
type SemanticConfig struct {
	OpenAIAPIKey        string  `json:"openai_api_key"`
	EmbeddingModel      string  `json:"embedding_model"`      // "text-embedding-3-small"
	MaxBatchSize        int     `json:"max_batch_size"`       // 100
	CacheEnabled        bool    `json:"cache_enabled"`        // true
	HybridWeight        float64 `json:"hybrid_weight"`        // 0.7 (semantic weight)
	SimilarityThreshold float64 `json:"similarity_threshold"` // 0.3
	Enabled             bool    `json:"enabled"`              // false by default
}

// Config represents the application configuration
type Config struct {
	DataPath   string          `json:"data_path"`
	ServerPort int             `json:"server_port"`
	LogLevel   string          `json:"log_level"`
	Obsidian   *ObsidianConfig `json:"obsidian,omitempty"`
	Semantic   *SemanticConfig `json:"semantic,omitempty"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		DataPath:   "data/memories.json",
		ServerPort: 8080,
		LogLevel:   "info",
		Semantic: &SemanticConfig{
			EmbeddingModel:      "text-embedding-3-small",
			MaxBatchSize:        100,
			CacheEnabled:        true,
			HybridWeight:        0.7,
			SimilarityThreshold: 0.3,
			Enabled:             false, // Disabled by default until API key is set
		},
	}
}

// LoadConfig loads configuration from file or returns default
func LoadConfig(configPath string) (*Config, error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// Save saves the configuration to file
func (c *Config) Save(configPath string) error {
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}
