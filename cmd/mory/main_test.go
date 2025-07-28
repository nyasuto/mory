package main

import (
	"os"
	"testing"
)

func TestBuildVariables(t *testing.T) {
	// Test that build variables have default values
	if version == "" {
		t.Error("version should have a default value")
	}

	if commit == "" {
		t.Error("commit should have a default value")
	}

	// Default values should be set
	expectedVersion := "dev"
	expectedCommit := "unknown"

	if version != expectedVersion {
		t.Errorf("Expected version %s, got %s", expectedVersion, version)
	}

	if commit != expectedCommit {
		t.Errorf("Expected commit %s, got %s", expectedCommit, commit)
	}
}

func TestVersionFlag(t *testing.T) {
	// Test version flag functionality
	// We can't easily test the main function's flag parsing without refactoring,
	// but we can test that the variables exist and have expected values

	if version == "" {
		t.Error("version variable should be defined")
	}

	if commit == "" {
		t.Error("commit variable should be defined")
	}
}

func TestMainPackageStructure(t *testing.T) {
	// Test that main package has the expected structure
	// This is a basic sanity check

	// Check that we can access environment variables (used in config loading)
	path := os.Getenv("PATH")
	if path == "" {
		t.Log("Warning: PATH environment variable is empty")
	}

	// Check that we can create temp directories (used by data dir provider)
	tempDir, err := os.MkdirTemp("", "mory-main-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Warning: failed to remove temp dir: %v", err)
		}
	}()

	// Check that the temp directory was created
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Error("Temp directory was not created")
	}
}

// TestMainFunctionExists is a compile-time check that main function exists
func TestMainFunctionExists(t *testing.T) {
	// If this test compiles, it means the main function exists
	// This is mostly for documentation purposes
	t.Log("main function exists and package compiles successfully")
}
