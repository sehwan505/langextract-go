package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/sehwan505/langextract-go/cmd/langextract/internal/config"
	"github.com/sehwan505/langextract-go/cmd/langextract/internal/logger"
)

// VersionInfo contains build-time version information
type VersionInfo struct {
	Version string
	Commit  string
	Date    string
}

// NewRootCommand creates the root langextract command with all subcommands
func NewRootCommand(cfg *config.GlobalConfig, log *logger.Logger, version *VersionInfo) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "langextract",
		Short: "Extract structured information from text using LLMs",
		Long: `LangExtract is a powerful tool for extracting structured information from unstructured text
using Large Language Models (LLMs). It supports multiple providers, batch processing,
visualization, and schema validation.

Examples:
  # Extract entities from a text file
  langextract extract --schema person.yaml input.txt

  # Process multiple files in parallel
  langextract batch --schema schema.yaml --output results/ *.txt

  # Generate interactive visualization
  langextract visualize --format html results.json

  # Validate schema and examples
  langextract validate --schema schema.yaml --examples examples/

  # Manage and test providers
  langextract providers list`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Apply global flags to configuration and logger
			if err := applyGlobalFlags(cmd, cfg, log); err != nil {
				return fmt.Errorf("failed to apply global flags: %w", err)
			}
			return nil
		},
	}

	// Add version flag and command
	rootCmd.Version = version.Version
	rootCmd.SetVersionTemplate(fmt.Sprintf("langextract version %s (commit: %s, built: %s)\n", 
		version.Version, version.Commit, version.Date))

	// Add global flags
	addGlobalFlags(rootCmd, cfg)

	// Add subcommands
	rootCmd.AddCommand(NewExtractCommand(cfg, log))
	rootCmd.AddCommand(NewBatchCommand(cfg, log))
	rootCmd.AddCommand(NewVisualizeCommand(cfg, log))
	rootCmd.AddCommand(NewValidateCommand(cfg, log))
	rootCmd.AddCommand(NewProvidersCommand(cfg, log))
	rootCmd.AddCommand(NewConfigCommand(cfg, log))

	return rootCmd
}

// addGlobalFlags adds persistent flags that apply to all commands
func addGlobalFlags(cmd *cobra.Command, cfg *config.GlobalConfig) {
	flags := cmd.PersistentFlags()

	// Logging flags
	flags.String("log-level", cfg.LogLevel, "Set log level (debug, info, warn, error)")
	flags.String("log-format", cfg.LogFormat, "Set log format (text, json)")

	// Provider flags  
	flags.String("provider", cfg.DefaultProvider, "Set default provider (openai, gemini, ollama)")
	flags.String("openai-api-key", "", "OpenAI API key (or LANGEXTRACT_OPENAI_API_KEY)")
	flags.String("gemini-api-key", "", "Gemini API key (or LANGEXTRACT_GEMINI_API_KEY)")
	flags.String("ollama-endpoint", cfg.OllamaEndpoint, "Ollama endpoint URL")

	// Request configuration flags
	flags.Int("timeout", cfg.RequestTimeout, "Request timeout in seconds")
	flags.Int("max-retries", cfg.MaxRetries, "Maximum number of retries")
	flags.Int("retry-delay", cfg.RetryDelay, "Retry delay in seconds")

	// Output flags
	flags.String("format", cfg.DefaultFormat, "Output format (json, yaml, csv, html, markdown, text)")
	flags.Bool("pretty", cfg.PrettyPrint, "Pretty print output")
	flags.Bool("color", cfg.ColorOutput, "Enable colored output")

	// Performance flags
	flags.Int("concurrency", cfg.Concurrency, "Number of concurrent requests")
	flags.Bool("cache", cfg.CacheEnabled, "Enable response caching")
	flags.Int("cache-size", cfg.CacheSize, "Cache size limit")

	// Progress and verbosity flags
	flags.Bool("progress", cfg.ShowProgress, "Show progress indicators")
	flags.Bool("quiet", cfg.Quiet, "Suppress non-error output")
	flags.Bool("verbose", cfg.Verbose, "Enable verbose output")

	// Validation flags
	flags.Bool("strict", cfg.StrictValidation, "Enable strict validation")

	// Directory flags
	flags.String("config-dir", cfg.ConfigDir, "Configuration directory")
	flags.String("cache-dir", cfg.CacheDir, "Cache directory")  
	flags.String("output-dir", cfg.OutputDir, "Output directory")
}

// applyGlobalFlags applies command-line flags to configuration and logger
func applyGlobalFlags(cmd *cobra.Command, cfg *config.GlobalConfig, log *logger.Logger) error {
	flags := cmd.PersistentFlags()

	// Apply logging flags
	if flags.Changed("log-level") {
		if level, err := flags.GetString("log-level"); err == nil {
			cfg.LogLevel = level
		}
	}
	if flags.Changed("log-format") {
		if format, err := flags.GetString("log-format"); err == nil {
			cfg.LogFormat = format
		}
	}

	// Apply provider flags
	if flags.Changed("provider") {
		if provider, err := flags.GetString("provider"); err == nil {
			cfg.DefaultProvider = provider
		}
	}
	if flags.Changed("openai-api-key") {
		if key, err := flags.GetString("openai-api-key"); err == nil && key != "" {
			cfg.OpenAIAPIKey = key
		}
	}
	if flags.Changed("gemini-api-key") {
		if key, err := flags.GetString("gemini-api-key"); err == nil && key != "" {
			cfg.GeminiAPIKey = key
		}
	}
	if flags.Changed("ollama-endpoint") {
		if endpoint, err := flags.GetString("ollama-endpoint"); err == nil {
			cfg.OllamaEndpoint = endpoint
		}
	}

	// Apply request configuration flags
	if flags.Changed("timeout") {
		if timeout, err := flags.GetInt("timeout"); err == nil {
			cfg.RequestTimeout = timeout
		}
	}
	if flags.Changed("max-retries") {
		if retries, err := flags.GetInt("max-retries"); err == nil {
			cfg.MaxRetries = retries
		}
	}
	if flags.Changed("retry-delay") {
		if delay, err := flags.GetInt("retry-delay"); err == nil {
			cfg.RetryDelay = delay
		}
	}

	// Apply output flags
	if flags.Changed("format") {
		if format, err := flags.GetString("format"); err == nil {
			cfg.DefaultFormat = format
		}
	}
	if flags.Changed("pretty") {
		if pretty, err := flags.GetBool("pretty"); err == nil {
			cfg.PrettyPrint = pretty
		}
	}
	if flags.Changed("color") {
		if color, err := flags.GetBool("color"); err == nil {
			cfg.ColorOutput = color
		}
	}

	// Apply performance flags
	if flags.Changed("concurrency") {
		if concurrency, err := flags.GetInt("concurrency"); err == nil {
			cfg.Concurrency = concurrency
		}
	}
	if flags.Changed("cache") {
		if cache, err := flags.GetBool("cache"); err == nil {
			cfg.CacheEnabled = cache
		}
	}
	if flags.Changed("cache-size") {
		if size, err := flags.GetInt("cache-size"); err == nil {
			cfg.CacheSize = size
		}
	}

	// Apply progress and verbosity flags
	if flags.Changed("progress") {
		if progress, err := flags.GetBool("progress"); err == nil {
			cfg.ShowProgress = progress
		}
	}
	if flags.Changed("quiet") {
		if quiet, err := flags.GetBool("quiet"); err == nil {
			cfg.Quiet = quiet
			log.SetQuiet(quiet)
		}
	}
	if flags.Changed("verbose") {
		if verbose, err := flags.GetBool("verbose"); err == nil {
			cfg.Verbose = verbose
			log.SetVerbose(verbose)
		}
	}

	// Apply validation flags
	if flags.Changed("strict") {
		if strict, err := flags.GetBool("strict"); err == nil {
			cfg.StrictValidation = strict
		}
	}

	// Apply directory flags
	if flags.Changed("config-dir") {
		if dir, err := flags.GetString("config-dir"); err == nil {
			cfg.ConfigDir = dir
		}
	}
	if flags.Changed("cache-dir") {
		if dir, err := flags.GetString("cache-dir"); err == nil {
			cfg.CacheDir = dir
		}
	}
	if flags.Changed("output-dir") {
		if dir, err := flags.GetString("output-dir"); err == nil {
			cfg.OutputDir = dir
		}
	}

	// Validate updated configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	return nil
}

// setupDirectories ensures required directories exist
func setupDirectories(cfg *config.GlobalConfig) error {
	dirs := []string{cfg.ConfigDir, cfg.CacheDir, cfg.OutputDir}
	
	for _, dir := range dirs {
		if dir == "" {
			continue
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	
	return nil
}