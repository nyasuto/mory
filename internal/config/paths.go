package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

// DataDirProvider provides platform-specific data directory paths
type DataDirProvider struct{}

// GetDataDir returns the appropriate data directory for Mory
func (p *DataDirProvider) GetDataDir() (string, error) {
	log.Printf("[DataDir] Starting data directory resolution process")
	log.Printf("[DataDir] Operating System: %s", runtime.GOOS)
	log.Printf("[DataDir] Architecture: %s", runtime.GOARCH)

	// Check for environment variable override
	if dataDir := os.Getenv("MORY_DATA_DIR"); dataDir != "" {
		log.Printf("[DataDir] Using environment variable override: MORY_DATA_DIR=%s", dataDir)
		return dataDir, nil
	}
	log.Printf("[DataDir] No MORY_DATA_DIR environment variable found")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("[DataDir] ERROR: Failed to get home directory: %v", err)
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	log.Printf("[DataDir] Home directory: %s", homeDir)

	var dataDir string
	switch runtime.GOOS {
	case "darwin": // macOS
		dataDir = filepath.Join(homeDir, "Library", "Application Support", "Mory")
		log.Printf("[DataDir] macOS detected, using: %s", dataDir)
	case "linux":
		// Follow XDG Base Directory Specification
		if xdgData := os.Getenv("XDG_DATA_HOME"); xdgData != "" {
			dataDir = filepath.Join(xdgData, "mory")
			log.Printf("[DataDir] Linux with XDG_DATA_HOME=%s, using: %s", xdgData, dataDir)
		} else {
			dataDir = filepath.Join(homeDir, ".local", "share", "mory")
			log.Printf("[DataDir] Linux with default XDG, using: %s", dataDir)
		}
	case "windows":
		if appData := os.Getenv("LOCALAPPDATA"); appData != "" {
			dataDir = filepath.Join(appData, "Mory")
			log.Printf("[DataDir] Windows with LOCALAPPDATA=%s, using: %s", appData, dataDir)
		} else {
			dataDir = filepath.Join(homeDir, "AppData", "Local", "Mory")
			log.Printf("[DataDir] Windows with default path, using: %s", dataDir)
		}
	default:
		// Fallback for unknown platforms
		dataDir = filepath.Join(homeDir, "mory-data")
		log.Printf("[DataDir] Unknown platform, using fallback: %s", dataDir)
	}

	log.Printf("[DataDir] Final resolved data directory: %s", dataDir)
	return dataDir, nil
}

// EnsureDataDir creates the data directory if it doesn't exist
func (p *DataDirProvider) EnsureDataDir() (string, error) {
	log.Printf("[EnsureDataDir] Starting directory creation process")

	dataDir, err := p.GetDataDir()
	if err != nil {
		log.Printf("[EnsureDataDir] ERROR: Failed to get data directory: %v", err)
		return "", fmt.Errorf("failed to get data directory: %w", err)
	}

	log.Printf("[EnsureDataDir] Target directory: %s", dataDir)

	// Check if directory already exists
	if info, err := os.Stat(dataDir); err == nil {
		if info.IsDir() {
			log.Printf("[EnsureDataDir] Directory already exists: %s", dataDir)
			log.Printf("[EnsureDataDir] Directory permissions: %s", info.Mode())
			return dataDir, nil
		} else {
			log.Printf("[EnsureDataDir] ERROR: Path exists but is not a directory: %s", dataDir)
			return "", fmt.Errorf("path exists but is not a directory: %s", dataDir)
		}
	} else if !os.IsNotExist(err) {
		log.Printf("[EnsureDataDir] ERROR: Failed to stat directory: %v", err)
		return "", fmt.Errorf("failed to stat directory: %w", err)
	}

	log.Printf("[EnsureDataDir] Directory does not exist, attempting to create: %s", dataDir)

	// Get working directory for debugging
	if wd, err := os.Getwd(); err == nil {
		log.Printf("[EnsureDataDir] Current working directory: %s", wd)
	} else {
		log.Printf("[EnsureDataDir] WARNING: Could not get working directory: %v", err)
	}

	// Check parent directory permissions
	parentDir := filepath.Dir(dataDir)
	log.Printf("[EnsureDataDir] Parent directory: %s", parentDir)

	if parentInfo, err := os.Stat(parentDir); err == nil {
		log.Printf("[EnsureDataDir] Parent directory exists, permissions: %s", parentInfo.Mode())
		if !parentInfo.IsDir() {
			log.Printf("[EnsureDataDir] ERROR: Parent path is not a directory: %s", parentDir)
			return "", fmt.Errorf("parent path is not a directory: %s", parentDir)
		}
	} else {
		log.Printf("[EnsureDataDir] Parent directory stat error: %v", err)
		log.Printf("[EnsureDataDir] Attempting to create parent directories recursively")
	}

	// Attempt to create directory with detailed error logging
	log.Printf("[EnsureDataDir] Creating directory with mode 0755: %s", dataDir)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Printf("[EnsureDataDir] ERROR: Failed to create directory: %v", err)
		log.Printf("[EnsureDataDir] Error type: %T", err)

		// Additional debugging for permission errors
		if os.IsPermission(err) {
			log.Printf("[EnsureDataDir] Permission denied - checking filesystem status")
			p.debugFilesystemPermissions(dataDir)
		}

		return "", fmt.Errorf("failed to create data directory '%s': %w", dataDir, err)
	}

	log.Printf("[EnsureDataDir] Successfully created directory: %s", dataDir)

	// Verify the directory was created properly
	if info, err := os.Stat(dataDir); err == nil {
		log.Printf("[EnsureDataDir] Verification successful - Directory permissions: %s", info.Mode())
	} else {
		log.Printf("[EnsureDataDir] WARNING: Could not verify created directory: %v", err)
	}

	return dataDir, nil
}

// debugFilesystemPermissions provides detailed filesystem debugging information
func (p *DataDirProvider) debugFilesystemPermissions(targetPath string) {
	log.Printf("[Debug] Filesystem debugging for path: %s", targetPath)

	// Check if we're trying to write to a read-only filesystem
	testFile := filepath.Join(filepath.Dir(targetPath), ".mory-write-test")
	log.Printf("[Debug] Testing write permissions with file: %s", testFile)

	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		log.Printf("[Debug] Write test failed: %v", err)
		if os.IsPermission(err) {
			log.Printf("[Debug] Confirmed: Permission denied for write operations")
		}
	} else {
		log.Printf("[Debug] Write test successful, cleaning up test file")
		if removeErr := os.Remove(testFile); removeErr != nil {
			log.Printf("[Debug] Warning: failed to cleanup test file: %v", removeErr)
		}
	}

	// Check mount points and filesystem info
	if wd, err := os.Getwd(); err == nil {
		log.Printf("[Debug] Current working directory: %s", wd)
		if testErr := os.WriteFile(filepath.Join(wd, ".mory-wd-test"), []byte("test"), 0644); testErr != nil {
			log.Printf("[Debug] Working directory write test failed: %v", testErr)
		} else {
			log.Printf("[Debug] Working directory is writable")
			wdTestFile := filepath.Join(wd, ".mory-wd-test")
			if removeErr := os.Remove(wdTestFile); removeErr != nil {
				log.Printf("[Debug] Warning: failed to cleanup working directory test file: %v", removeErr)
			}
		}
	}
}
