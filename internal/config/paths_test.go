package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
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

func TestDataDirProvider_debugFilesystemPermissions(t *testing.T) {
	provider := &DataDirProvider{}

	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mory-debug-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Warning: failed to remove temp dir: %v", err)
		}
	}()

	// Test debugFilesystemPermissions with a writable directory
	testPath := filepath.Join(tempDir, "nonexistent-dir")

	// This should not panic and should complete without error
	// We can't easily test the logging output, but we can ensure it doesn't crash
	provider.debugFilesystemPermissions(testPath)

	// The function should have attempted to create test files in the parent directory
	// and current working directory, then cleaned them up
}

func TestDataDirProvider_GetDataDir_ErrorCases(t *testing.T) {
	// Test various edge cases for GetDataDir
	t.Run("home directory error simulation", func(t *testing.T) {
		// We can't easily simulate os.UserHomeDir() failure without mocking,
		// but we can test with invalid environment variables
		provider := &DataDirProvider{}

		// Test with empty MORY_DATA_DIR (should fall back to platform defaults)
		if err := os.Setenv("MORY_DATA_DIR", ""); err != nil {
			t.Fatalf("Failed to set environment variable: %v", err)
		}
		defer func() {
			if err := os.Unsetenv("MORY_DATA_DIR"); err != nil {
				t.Logf("Warning: failed to unset environment variable: %v", err)
			}
		}()

		dataDir, err := provider.GetDataDir()
		if err != nil {
			t.Fatalf("GetDataDir should not fail with empty MORY_DATA_DIR: %v", err)
		}

		if dataDir == "" {
			t.Error("GetDataDir should return non-empty path")
		}
	})
}

func TestDataDirProvider_EnsureDataDir_ErrorCases(t *testing.T) {
	t.Run("file exists where directory should be", func(t *testing.T) {
		provider := &DataDirProvider{}

		// Create temporary file
		tempDir, err := os.MkdirTemp("", "mory-error-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer func() {
			if err := os.RemoveAll(tempDir); err != nil {
				t.Logf("Warning: failed to remove temp dir: %v", err)
			}
		}()

		// Create a file where directory should be
		filePath := filepath.Join(tempDir, "should-be-dir")
		if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Set MORY_DATA_DIR to point to this file
		if err := os.Setenv("MORY_DATA_DIR", filePath); err != nil {
			t.Fatalf("Failed to set environment variable: %v", err)
		}
		defer func() {
			if err := os.Unsetenv("MORY_DATA_DIR"); err != nil {
				t.Logf("Warning: failed to unset environment variable: %v", err)
			}
		}()

		// EnsureDataDir should fail
		_, err = provider.EnsureDataDir()
		if err == nil {
			t.Error("EnsureDataDir should fail when file exists where directory should be")
		}

		if !strings.Contains(err.Error(), "not a directory") {
			t.Errorf("Error should mention 'not a directory', got: %v", err)
		}
	})
}
