// Package main provides the langextract command-line interface.
//
// The langextract CLI provides comprehensive text extraction and visualization
// capabilities through multiple commands:
//
//   - extract: Extract structured information from single documents
//   - batch: Process multiple documents in parallel
//   - visualize: Generate interactive visualizations
//   - validate: Validate schemas and examples
//   - providers: Manage and test language model providers
//
// Each command supports extensive configuration options, progress tracking,
// and multiple output formats.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/sehwan505/langextract-go/cmd/langextract/internal/cli"
	"github.com/sehwan505/langextract-go/cmd/langextract/internal/config"
	"github.com/sehwan505/langextract-go/cmd/langextract/internal/logger"
)

// Version information injected at build time
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	// Set up graceful shutdown
	ctx, cancel := setupGracefulShutdown()
	defer cancel()

	// Initialize global configuration
	cfg, err := config.LoadGlobalConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log, err := logger.NewLogger(cfg.LogLevel, cfg.LogFormat)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	// Create and configure root command
	rootCmd := cli.NewRootCommand(cfg, log, &cli.VersionInfo{
		Version: version,
		Commit:  commit,
		Date:    date,
	})

	// Execute the command with context
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		// Error is already handled by cobra, just exit with non-zero code
		os.Exit(1)
	}
}

// setupGracefulShutdown creates a context that cancels on interrupt signals
func setupGracefulShutdown() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	
	// Handle interrupt signals gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	
	go func() {
		select {
		case sig := <-sigChan:
			fmt.Fprintf(os.Stderr, "\nReceived signal %v, shutting down gracefully...\n", sig)
			cancel()
		case <-ctx.Done():
			// Context was cancelled elsewhere
		}
	}()
	
	return ctx, cancel
}