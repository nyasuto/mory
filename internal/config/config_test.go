package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.DataPath != "data/memories.json" {
		t.Errorf("Expected DataPath 'data/memories.json', got '%s'", config.DataPath)
	}

	if config.ServerPort != 8080 {
		t.Errorf("Expected ServerPort 8080, got %d", config.ServerPort)
	}

	if config.LogLevel != "info" {
		t.Errorf("Expected LogLevel 'info', got '%s'", config.LogLevel)
	}
}

func TestLoadConfig_NonExistentFile(t *testing.T) {
	// Test loading non-existent file returns default config
	config, err := LoadConfig("nonexistent.json")
	if err != nil {
		t.Fatalf("Expected no error for non-existent file, got %v", err)
	}

	// Should return default config
	defaultConfig := DefaultConfig()
	if config.DataPath != defaultConfig.DataPath {
		t.Errorf("Expected default DataPath, got '%s'", config.DataPath)
	}

	if config.ServerPort != defaultConfig.ServerPort {
		t.Errorf("Expected default ServerPort, got %d", config.ServerPort)
	}

	if config.LogLevel != defaultConfig.LogLevel {
		t.Errorf("Expected default LogLevel, got '%s'", config.LogLevel)
	}
}

func TestLoadConfig_ValidFile(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	// Create test config
	testConfig := &Config{
		DataPath:   "custom/path.json",
		ServerPort: 9090,
		LogLevel:   "debug",
	}

	// Write test config to file
	data, err := json.MarshalIndent(testConfig, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal test config: %v", err)
	}

	err = os.WriteFile(configPath, data, 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Load config
	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify loaded values
	if config.DataPath != "custom/path.json" {
		t.Errorf("Expected DataPath 'custom/path.json', got '%s'", config.DataPath)
	}

	if config.ServerPort != 9090 {
		t.Errorf("Expected ServerPort 9090, got %d", config.ServerPort)
	}

	if config.LogLevel != "debug" {
		t.Errorf("Expected LogLevel 'debug', got '%s'", config.LogLevel)
	}
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "invalid.json")

	// Write invalid JSON
	err := os.WriteFile(configPath, []byte("invalid json"), 0644)
	if err != nil {
		t.Fatalf("Failed to write invalid config file: %v", err)
	}

	// Should return error
	_, err = LoadConfig(configPath)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestLoadConfig_EmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "empty.json")

	// Write empty file
	err := os.WriteFile(configPath, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to write empty config file: %v", err)
	}

	// Should return error
	_, err = LoadConfig(configPath)
	if err == nil {
		t.Error("Expected error for empty file")
	}
}

func TestLoadConfig_PartialConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "partial.json")

	// Write partial config (only some fields)
	partialJSON := `{"data_path": "custom/path.json"}`
	err := os.WriteFile(configPath, []byte(partialJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to write partial config file: %v", err)
	}

	// Load config
	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load partial config: %v", err)
	}

	// Verify loaded and default values
	if config.DataPath != "custom/path.json" {
		t.Errorf("Expected DataPath 'custom/path.json', got '%s'", config.DataPath)
	}

	// These should be zero values since not specified in JSON
	if config.ServerPort != 0 {
		t.Errorf("Expected ServerPort 0 (zero value), got %d", config.ServerPort)
	}

	if config.LogLevel != "" {
		t.Errorf("Expected LogLevel '' (zero value), got '%s'", config.LogLevel)
	}
}

func TestConfig_Save(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "subdir", "config.json")

	config := &Config{
		DataPath:   "test/path.json",
		ServerPort: 3000,
		LogLevel:   "warn",
	}

	// Save config
	err := config.Save(configPath)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Verify directory was created
	if _, err := os.Stat(filepath.Dir(configPath)); os.IsNotExist(err) {
		t.Error("Config directory was not created")
	}

	// Load and verify content
	loadedConfig, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if loadedConfig.DataPath != config.DataPath {
		t.Errorf("Expected DataPath '%s', got '%s'", config.DataPath, loadedConfig.DataPath)
	}

	if loadedConfig.ServerPort != config.ServerPort {
		t.Errorf("Expected ServerPort %d, got %d", config.ServerPort, loadedConfig.ServerPort)
	}

	if loadedConfig.LogLevel != config.LogLevel {
		t.Errorf("Expected LogLevel '%s', got '%s'", config.LogLevel, loadedConfig.LogLevel)
	}
}

func TestConfig_SaveInvalidPath(t *testing.T) {
	config := DefaultConfig()

	// Try to save to invalid path (e.g., root directory that we can't write to)
	invalidPath := "/root/readonly/config.json"

	err := config.Save(invalidPath)
	if err == nil {
		t.Error("Expected error when saving to invalid path")
	}
}

func TestConfig_SaveFormatting(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "formatted.json")

	config := &Config{
		DataPath:   "test/path.json",
		ServerPort: 3000,
		LogLevel:   "warn",
	}

	// Save config
	err := config.Save(configPath)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Read raw file content
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read saved config: %v", err)
	}

	content := string(data)

	// Verify JSON is properly formatted (indented)
	if !containsString(content, "{\n") {
		t.Error("Expected formatted JSON with newlines")
	}

	// Verify it contains expected fields
	if !containsString(content, `"data_path"`) {
		t.Error("Expected data_path field in JSON")
	}

	if !containsString(content, `"server_port"`) {
		t.Error("Expected server_port field in JSON")
	}

	if !containsString(content, `"log_level"`) {
		t.Error("Expected log_level field in JSON")
	}
}

// Helper function for string containment check
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && func() bool {
		for i := 0; i <= len(s)-len(substr); i++ {
			if s[i:i+len(substr)] == substr {
				return true
			}
		}
		return false
	}()
}
