package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
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

	// Initialize memory store
	store := memory.NewJSONMemoryStore("data/memories.json", "data/operations.log")

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
