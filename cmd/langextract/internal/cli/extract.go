package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/sehwan505/langextract-go/cmd/langextract/internal/config"
	"github.com/sehwan505/langextract-go/cmd/langextract/internal/logger"
	"github.com/sehwan505/langextract-go/pkg/langextract"
)

// ExtractOptions contains all options for the extract command
type ExtractOptions struct {
	// Input options
	Input      string   // Input file or URL
	InputType  string   // Input type: auto, file, url, text
	Schema     string   // Schema file path
	Examples   []string // Example files
	
	// Model options
	ModelID     string // Model ID to use
	Temperature float32 // Temperature for generation
	MaxTokens   int    // Maximum tokens to generate
	
	// Processing options
	ContextWindow int    // Context window size for chunking
	ChunkSize     int    // Chunk size for large documents
	ChunkOverlap  int    // Overlap between chunks
	
	// Output options
	Output     string // Output file (default: stdout)
	Format     string // Output format
	Pretty     bool   // Pretty print
	ShowSource bool   // Include source text in output
	
	// Validation options
	ValidateSchema   bool // Validate against schema
	ValidateExamples bool // Validate examples
}

// NewExtractCommand creates the extract command
func NewExtractCommand(cfg *config.GlobalConfig, log *logger.Logger) *cobra.Command {
	opts := &ExtractOptions{
		Temperature:   0.1,
		MaxTokens:     4096,
		ContextWindow: 1000,
		ChunkSize:     2048,
		ChunkOverlap:  256,
		Format:        cfg.DefaultFormat,
		Pretty:        cfg.PrettyPrint,
	}

	cmd := &cobra.Command{
		Use:   "extract [flags] [input]",
		Short: "Extract structured information from a single document",
		Long: `Extract structured information from a single document using a specified schema.
The input can be a file path, URL, or text content. The schema defines what information
to extract and how to structure it.

Examples:
  # Extract from a text file using a schema
  langextract extract --schema person.yaml input.txt
  
  # Extract from a URL with custom model
  langextract extract --schema article.yaml --model gpt-4 https://example.com/article
  
  # Extract with examples for better accuracy
  langextract extract --schema event.yaml --examples events/ input.txt
  
  # Extract with specific output format
  langextract extract --schema data.yaml --format csv --output results.csv input.txt
  
  # Extract from stdin
  echo "John Doe is 30 years old" | langextract extract --schema person.yaml`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Determine input source
			if len(args) > 0 {
				opts.Input = args[0]
			} else if opts.Input == "" {
				opts.Input = "-" // stdin
			}

			return runExtract(cmd.Context(), opts, cfg, log)
		},
	}

	// Input flags
	cmd.Flags().StringVar(&opts.Input, "input", "", "Input file, URL, or '-' for stdin")
	cmd.Flags().StringVar(&opts.InputType, "input-type", "auto", "Input type: auto, file, url, text")
	cmd.Flags().StringVar(&opts.Schema, "schema", "", "Schema file path (required)")
	cmd.Flags().StringArrayVar(&opts.Examples, "examples", nil, "Example files or directories")

	// Model flags
	cmd.Flags().StringVar(&opts.ModelID, "model", "", "Model ID to use (overrides provider default)")
	cmd.Flags().Float32Var(&opts.Temperature, "temperature", opts.Temperature, "Temperature for generation (0.0-2.0)")
	cmd.Flags().IntVar(&opts.MaxTokens, "max-tokens", opts.MaxTokens, "Maximum tokens to generate")

	// Processing flags
	cmd.Flags().IntVar(&opts.ContextWindow, "context-window", opts.ContextWindow, "Context window size for chunking")
	cmd.Flags().IntVar(&opts.ChunkSize, "chunk-size", opts.ChunkSize, "Chunk size for large documents")
	cmd.Flags().IntVar(&opts.ChunkOverlap, "chunk-overlap", opts.ChunkOverlap, "Overlap between chunks")

	// Output flags
	cmd.Flags().StringVarP(&opts.Output, "output", "o", "", "Output file (default: stdout)")
	cmd.Flags().StringVar(&opts.Format, "format", opts.Format, "Output format (json, yaml, csv)")
	cmd.Flags().BoolVar(&opts.Pretty, "pretty", opts.Pretty, "Pretty print output")
	cmd.Flags().BoolVar(&opts.ShowSource, "show-source", false, "Include source text in output")

	// Validation flags
	cmd.Flags().BoolVar(&opts.ValidateSchema, "validate-schema", true, "Validate against schema")
	cmd.Flags().BoolVar(&opts.ValidateExamples, "validate-examples", false, "Validate examples")

	// Mark required flags
	cmd.MarkFlagRequired("schema")

	return cmd
}

// runExtract executes the extract command
func runExtract(ctx context.Context, opts *ExtractOptions, cfg *config.GlobalConfig, log *logger.Logger) error {
	log.WithOperation("extract").Info("Starting extraction")

	// Validate options
	if err := validateExtractOptions(opts); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}

	// Read input
	input, err := readInput(opts.Input, opts.InputType)
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	// Load schema
	schema, err := loadSchema(opts.Schema)
	if err != nil {
		return fmt.Errorf("failed to load schema: %w", err)
	}

	// Load examples if provided
	var examples []string
	if len(opts.Examples) > 0 {
		examples, err = loadExamples(opts.Examples)
		if err != nil {
			return fmt.Errorf("failed to load examples: %w", err)
		}
	}

	// Create extraction options
	extractOpts := langextract.NewExtractOptions().
		WithSchema(schema).
		WithTimeout(time.Duration(cfg.RequestTimeout) * time.Second).
		WithRetryCount(cfg.MaxRetries)

	// Set model if specified
	if opts.ModelID != "" {
		extractOpts = extractOpts.WithModelID(opts.ModelID)
	}

	// Set model parameters
	if opts.Temperature > 0 {
		extractOpts = extractOpts.WithTemperature(opts.Temperature)
	}
	if opts.MaxTokens > 0 {
		extractOpts = extractOpts.WithMaxTokens(opts.MaxTokens)
	}

	// Add examples
	if len(examples) > 0 {
		// Convert string examples to ExampleData - this would need proper implementation
		// For now, just skip examples until the ExampleData conversion is implemented
		log.WithOperation("extract").Warning("Examples not yet implemented in CLI")
	}

	// Perform extraction
	log.WithOperation("extract").WithFile(opts.Input).Info("Performing extraction")
	
	result, err := langextract.Extract(input, extractOpts)
	if err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}

	// Log extraction summary
	log.WithOperation("extract").
		WithCount(len(result.Extractions)).
		Info("Extraction completed")

	// Create visualization options for output
	vizOpts := langextract.NewVisualizeOptions().
		WithFormat(opts.Format).
		WithPretty(opts.Pretty).
		WithIncludeText(opts.ShowSource)

	// Generate output
	output, err := langextract.Visualize(result, vizOpts)
	if err != nil {
		return fmt.Errorf("failed to generate output: %w", err)
	}

	// Write output
	if err := writeOutput(output, opts.Output); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	log.Success("Extraction completed successfully")
	return nil
}

// validateExtractOptions validates extract command options
func validateExtractOptions(opts *ExtractOptions) error {
	if opts.Schema == "" {
		return fmt.Errorf("schema is required")
	}

	if opts.Temperature < 0 || opts.Temperature > 2.0 {
		return fmt.Errorf("temperature must be between 0.0 and 2.0")
	}

	if opts.MaxTokens < 1 {
		return fmt.Errorf("max-tokens must be positive")
	}

	if opts.ChunkSize < 100 {
		return fmt.Errorf("chunk-size must be at least 100")
	}

	if opts.ChunkOverlap >= opts.ChunkSize {
		return fmt.Errorf("chunk-overlap must be less than chunk-size")
	}

	return nil
}

// readInput reads input from file, URL, or stdin
func readInput(input, inputType string) (string, error) {
	switch input {
	case "", "-":
		// Read from stdin
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("failed to read from stdin: %w", err)
		}
		return string(data), nil
		
	default:
		// Auto-detect input type if not specified
		if inputType == "auto" {
			if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
				inputType = "url"
			} else if _, err := os.Stat(input); err == nil {
				inputType = "file"
			} else {
				inputType = "text"
			}
		}

		switch inputType {
		case "file":
			data, err := os.ReadFile(input)
			if err != nil {
				return "", fmt.Errorf("failed to read file %s: %w", input, err)
			}
			return string(data), nil

		case "url":
			// For now, return an error for URL input - this would need HTTP client implementation
			return "", fmt.Errorf("URL input not yet implemented")

		case "text":
			return input, nil

		default:
			return "", fmt.Errorf("unsupported input type: %s", inputType)
		}
	}
}

// loadSchema loads extraction schema from file
func loadSchema(schemaPath string) (interface{}, error) {
	data, err := os.ReadFile(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema file %s: %w", schemaPath, err)
	}

	// For now, return the raw schema data
	// In a full implementation, this would parse YAML/JSON schema
	return string(data), nil
}

// loadExamples loads example files from paths
func loadExamples(examplePaths []string) ([]string, error) {
	var examples []string

	for _, path := range examplePaths {
		info, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("failed to stat example path %s: %w", path, err)
		}

		if info.IsDir() {
			// Load all files from directory
			dirExamples, err := loadExamplesFromDir(path)
			if err != nil {
				return nil, err
			}
			examples = append(examples, dirExamples...)
		} else {
			// Load single file
			data, err := os.ReadFile(path)
			if err != nil {
				return nil, fmt.Errorf("failed to read example file %s: %w", path, err)
			}
			examples = append(examples, string(data))
		}
	}

	return examples, nil
}

// loadExamplesFromDir loads all example files from a directory
func loadExamplesFromDir(dirPath string) ([]string, error) {
	var examples []string

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && (strings.HasSuffix(path, ".txt") || 
			strings.HasSuffix(path, ".md") || strings.HasSuffix(path, ".yaml") || 
			strings.HasSuffix(path, ".json")) {
			
			data, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read example file %s: %w", path, err)
			}
			examples = append(examples, string(data))
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory %s: %w", dirPath, err)
	}

	return examples, nil
}

// writeOutput writes output to file or stdout
func writeOutput(output, outputPath string) error {
	if outputPath == "" || outputPath == "-" {
		// Write to stdout
		fmt.Print(output)
		return nil
	}

	// Write to file
	if err := os.WriteFile(outputPath, []byte(output), 0644); err != nil {
		return fmt.Errorf("failed to write to file %s: %w", outputPath, err)
	}

	return nil
}