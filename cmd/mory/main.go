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

// RunOptions contains configuration for running the application
type RunOptions struct {
	Args       []string
	ConfigPath string
}

// Run executes the main application logic with the given options
func Run(opts RunOptions) error {
	// Set up custom flag set for testing
	flagSet := flag.NewFlagSet("mory", flag.ContinueOnError)
	showVersion := flagSet.Bool("version", false, "Show version information")

	// Parse provided arguments
	if err := flagSet.Parse(opts.Args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	// Handle version flag
	if *showVersion {
		fmt.Printf("Mory %s (commit: %s)\n", version, commit)
		return nil
	}

	// Load configuration
	configPath := opts.ConfigPath
	if configPath == "" {
		configPath = "config.json"
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize memory store with platform-appropriate data directory
	log.Printf("[Main] Initializing memory store...")
	pathProvider := &config.DataDirProvider{}
	dataDir, err := pathProvider.EnsureDataDir()
	if err != nil {
		log.Printf("[Main] FATAL: Failed to initialize data directory: %v", err)
		return fmt.Errorf("failed to initialize data directory: %w", err)
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
		return fmt.Errorf("server failed: %w", err)
	}

	return nil
}

func main() {
	// Parse command line flags
	var showVersion = flag.Bool("version", false, "Show version information")
	flag.Parse()

	// Handle version flag
	if *showVersion {
		fmt.Printf("Mory %s (commit: %s)\n", version, commit)
		os.Exit(0)
	}

	// Run with default options
	opts := RunOptions{
		Args:       os.Args[1:],
		ConfigPath: "config.json",
	}

	if err := Run(opts); err != nil {
		log.Fatalf("Application failed: %v", err)
	}
}
