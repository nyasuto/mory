package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/nyasuto/mory/internal/config"
	"github.com/nyasuto/mory/internal/mcp"
	"github.com/nyasuto/mory/internal/memory"
)

// Build-time variables (set by ldflags)
var (
	version = "dev"
	commit  = "unknown"
)

func main() {
	// Parse command line flags
	var showVersion = flag.Bool("version", false, "Show version information")
	flag.Parse()

	// Handle version flag
	if *showVersion {
		fmt.Printf("Mory %s (commit: %s)\n", version, commit)
		os.Exit(0)
	}
	// Load configuration
	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize memory store with platform-appropriate data directory
	log.Printf("[Main] Initializing memory store...")
	pathProvider := &config.DataDirProvider{}
	dataDir, err := pathProvider.EnsureDataDir()
	if err != nil {
		log.Printf("[Main] FATAL: Failed to initialize data directory: %v", err)
		log.Fatalf("Failed to initialize data directory: %v", err)
	}

	memoriesFile := filepath.Join(dataDir, "memories.json")
	logFile := filepath.Join(dataDir, "operations.log")

	log.Printf("[Main] Memory store configuration:")
	log.Printf("[Main]   Data directory: %s", dataDir)
	log.Printf("[Main]   Memories file: %s", memoriesFile)
	log.Printf("[Main]   Log file: %s", logFile)

	store := memory.NewJSONMemoryStore(memoriesFile, logFile)
	log.Printf("[Main] Memory store initialized successfully")

	// Create MCP server
	server := mcp.NewServer(cfg, store)

	// Create context that cancels on interrupt
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down gracefully...")
		cancel()
	}()

	// Start server
	if err := server.Start(ctx); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
