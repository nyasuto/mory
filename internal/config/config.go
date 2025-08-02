package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

// LoadConfig loads configuration from file, environment variables, and .env file
func LoadConfig(configPath string) (*Config, error) {
	// Load .env file if it exists
	loadDotEnv()

	// Start with default config
	config := DefaultConfig()

	// Load from JSON file if it exists
	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(data, config); err != nil {
			return nil, err
		}
	}

	// Override with environment variables
	applyEnvironmentVariables(config)

	return config, nil
}

// loadDotEnv loads environment variables from .env file
func loadDotEnv() {
	envFile := ".env"
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		return
	}

	data, err := os.ReadFile(envFile)
	if err != nil {
		return
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			// Remove surrounding quotes
			value = strings.Trim(value, `"'`)
			_ = os.Setenv(key, value) // Ignore error - not critical for .env loading
		}
	}
}

// applyEnvironmentVariables applies environment variables to config
func applyEnvironmentVariables(config *Config) {
	// General config
	if val := os.Getenv("MORY_DATA_PATH"); val != "" {
		config.DataPath = val
	}
	if val := os.Getenv("MORY_SERVER_PORT"); val != "" {
		if port, err := strconv.Atoi(val); err == nil {
			config.ServerPort = port
		}
	}
	if val := os.Getenv("MORY_LOG_LEVEL"); val != "" {
		config.LogLevel = val
	}

	// Semantic config
	if config.Semantic == nil {
		config.Semantic = &SemanticConfig{}
	}

	if val := os.Getenv("MORY_OPENAI_API_KEY"); val != "" {
		config.Semantic.OpenAIAPIKey = val
	}
	if val := os.Getenv("MORY_EMBEDDING_MODEL"); val != "" {
		config.Semantic.EmbeddingModel = val
	}
	if val := os.Getenv("MORY_MAX_BATCH_SIZE"); val != "" {
		if size, err := strconv.Atoi(val); err == nil {
			config.Semantic.MaxBatchSize = size
		}
	}
	if val := os.Getenv("MORY_CACHE_ENABLED"); val != "" {
		if enabled, err := strconv.ParseBool(val); err == nil {
			config.Semantic.CacheEnabled = enabled
		}
	}
	if val := os.Getenv("MORY_HYBRID_WEIGHT"); val != "" {
		if weight, err := strconv.ParseFloat(val, 64); err == nil {
			config.Semantic.HybridWeight = weight
		}
	}
	if val := os.Getenv("MORY_SIMILARITY_THRESHOLD"); val != "" {
		if threshold, err := strconv.ParseFloat(val, 64); err == nil {
			config.Semantic.SimilarityThreshold = threshold
		}
	}
	if val := os.Getenv("MORY_SEMANTIC_ENABLED"); val != "" {
		if enabled, err := strconv.ParseBool(val); err == nil {
			config.Semantic.Enabled = enabled
		}
	}

	// Auto-enable semantic search if API key is provided
	if config.Semantic.OpenAIAPIKey != "" && !config.Semantic.Enabled {
		config.Semantic.Enabled = true
	}

	// Obsidian config
	if val := os.Getenv("MORY_OBSIDIAN_VAULT_PATH"); val != "" {
		if config.Obsidian == nil {
			config.Obsidian = &ObsidianConfig{}
		}
		config.Obsidian.VaultPath = val
	}
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
