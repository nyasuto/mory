package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDataDirProvider_GetDataDir(t *testing.T) {
	provider := &DataDirProvider{}

	// Test environment variable override
	t.Run("environment variable override", func(t *testing.T) {
		expected := "/custom/mory/path"
		if err := os.Setenv("MORY_DATA_DIR", expected); err != nil {
			t.Fatalf("Failed to set environment variable: %v", err)
		}
		defer func() {
			if err := os.Unsetenv("MORY_DATA_DIR"); err != nil {
				t.Logf("Warning: failed to unset environment variable: %v", err)
			}
		}()

		dataDir, err := provider.GetDataDir()
		if err != nil {
			t.Fatalf("GetDataDir failed: %v", err)
		}

		if dataDir != expected {
			t.Errorf("Expected %s, got %s", expected, dataDir)
		}
	})

	// Test platform-specific defaults
	t.Run("platform defaults", func(t *testing.T) {
		if err := os.Unsetenv("MORY_DATA_DIR"); err != nil {
			t.Logf("Warning: failed to unset MORY_DATA_DIR: %v", err)
		}
		if err := os.Unsetenv("XDG_DATA_HOME"); err != nil {
			t.Logf("Warning: failed to unset XDG_DATA_HOME: %v", err)
		}
		if err := os.Unsetenv("LOCALAPPDATA"); err != nil {
			t.Logf("Warning: failed to unset LOCALAPPDATA: %v", err)
		}

		dataDir, err := provider.GetDataDir()
		if err != nil {
			t.Fatalf("GetDataDir failed: %v", err)
		}

		homeDir, _ := os.UserHomeDir()
		var expected string

		switch runtime.GOOS {
		case "darwin":
			expected = filepath.Join(homeDir, "Library", "Application Support", "Mory")
		case "linux":
			expected = filepath.Join(homeDir, ".local", "share", "mory")
		case "windows":
			expected = filepath.Join(homeDir, "AppData", "Local", "Mory")
		default:
			expected = filepath.Join(homeDir, "mory-data")
		}

		if dataDir != expected {
			t.Errorf("Expected %s, got %s", expected, dataDir)
		}
	})
}

func TestDataDirProvider_EnsureDataDir(t *testing.T) {
	provider := &DataDirProvider{}

	// Use temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mory-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Warning: failed to remove temp dir: %v", err)
		}
	}()

	testDataDir := filepath.Join(tempDir, "test-mory-data")
	if err := os.Setenv("MORY_DATA_DIR", testDataDir); err != nil {
		t.Fatalf("Failed to set environment variable: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("MORY_DATA_DIR"); err != nil {
			t.Logf("Warning: failed to unset environment variable: %v", err)
		}
	}()

	dataDir, err := provider.EnsureDataDir()
	if err != nil {
		t.Fatalf("EnsureDataDir failed: %v", err)
	}

	if dataDir != testDataDir {
		t.Errorf("Expected %s, got %s", testDataDir, dataDir)
	}

	// Verify directory was created
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		t.Errorf("Data directory was not created: %s", dataDir)
	}
}
