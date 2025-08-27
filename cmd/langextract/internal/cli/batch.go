package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/sehwan505/langextract-go/cmd/langextract/internal/config"
	"github.com/sehwan505/langextract-go/cmd/langextract/internal/logger"
	"github.com/sehwan505/langextract-go/pkg/langextract"
)

// BatchOptions contains all options for the batch command
type BatchOptions struct {
	// Input options
	Inputs    []string // Input files, directories, or patterns
	Schema    string   // Schema file path
	Examples  []string // Example files
	Recursive bool     // Process directories recursively
	Include   []string // File patterns to include
	Exclude   []string // File patterns to exclude

	// Model options
	ModelID     string  // Model ID to use
	Temperature float32 // Temperature for generation
	MaxTokens   int     // Maximum tokens to generate

	// Processing options
	Concurrency   int // Number of concurrent extractions
	ContextWindow int // Context window size for chunking
	ChunkSize     int // Chunk size for large documents
	ChunkOverlap  int // Overlap between chunks

	// Output options
	OutputDir    string // Output directory
	OutputFormat string // Output format
	Pretty       bool   // Pretty print
	ShowSource   bool   // Include source text in output
	Merge        bool   // Merge results into single file

	// Progress options
	ShowProgress bool // Show progress bar
	ShowSummary  bool // Show processing summary

	// Error handling
	ContinueOnError bool // Continue processing on errors
	MaxErrors       int  // Maximum errors before stopping

	// Validation options
	ValidateSchema   bool // Validate against schema
	ValidateExamples bool // Validate examples
}

// BatchResult represents the result of processing a single file
type BatchResult struct {
	InputFile  string
	OutputFile string
	Success    bool
	Error      error
	Duration   time.Duration
	NumExtracted int
}

// NewBatchCommand creates the batch command
func NewBatchCommand(cfg *config.GlobalConfig, log *logger.Logger) *cobra.Command {
	opts := &BatchOptions{
		Temperature:     0.1,
		MaxTokens:       4096,
		Concurrency:     cfg.Concurrency,
		ContextWindow:   1000,
		ChunkSize:       2048,
		ChunkOverlap:    256,
		OutputFormat:    cfg.DefaultFormat,
		Pretty:          cfg.PrettyPrint,
		ShowProgress:    cfg.ShowProgress,
		ShowSummary:     true,
		ContinueOnError: true,
		MaxErrors:       10,
	}

	cmd := &cobra.Command{
		Use:   "batch [flags] [inputs...]",
		Short: "Process multiple documents in parallel",
		Long: `Process multiple documents in parallel using a specified schema.
Input can be individual files, directories, or glob patterns. Results are saved
to an output directory with configurable naming and format.

Examples:
  # Process all text files in a directory
  langextract batch --schema person.yaml --output results/ docs/*.txt
  
  # Process directory recursively with custom concurrency
  langextract batch --schema article.yaml --recursive --concurrency 8 --output results/ docs/
  
  # Process with file filtering
  langextract batch --schema data.yaml --include "*.md,*.txt" --exclude "*.tmp" --output results/ inputs/
  
  # Merge all results into single file
  langextract batch --schema events.yaml --merge --output results.json docs/
  
  # Continue processing despite errors
  langextract batch --schema schema.yaml --continue-on-error --max-errors 5 --output results/ files/`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Inputs = args
			return runBatch(cmd.Context(), opts, cfg, log)
		},
	}

	// Input flags
	cmd.Flags().StringVar(&opts.Schema, "schema", "", "Schema file path (required)")
	cmd.Flags().StringArrayVar(&opts.Examples, "examples", nil, "Example files or directories")
	cmd.Flags().BoolVarP(&opts.Recursive, "recursive", "r", false, "Process directories recursively")
	cmd.Flags().StringArrayVar(&opts.Include, "include", nil, "File patterns to include (comma-separated)")
	cmd.Flags().StringArrayVar(&opts.Exclude, "exclude", nil, "File patterns to exclude (comma-separated)")

	// Model flags
	cmd.Flags().StringVar(&opts.ModelID, "model", "", "Model ID to use (overrides provider default)")
	cmd.Flags().Float32Var(&opts.Temperature, "temperature", opts.Temperature, "Temperature for generation (0.0-2.0)")
	cmd.Flags().IntVar(&opts.MaxTokens, "max-tokens", opts.MaxTokens, "Maximum tokens to generate")

	// Processing flags
	cmd.Flags().IntVar(&opts.Concurrency, "concurrency", opts.Concurrency, "Number of concurrent extractions")
	cmd.Flags().IntVar(&opts.ContextWindow, "context-window", opts.ContextWindow, "Context window size for chunking")
	cmd.Flags().IntVar(&opts.ChunkSize, "chunk-size", opts.ChunkSize, "Chunk size for large documents")
	cmd.Flags().IntVar(&opts.ChunkOverlap, "chunk-overlap", opts.ChunkOverlap, "Overlap between chunks")

	// Output flags
	cmd.Flags().StringVarP(&opts.OutputDir, "output", "o", "", "Output directory (required)")
	cmd.Flags().StringVar(&opts.OutputFormat, "format", opts.OutputFormat, "Output format (json, yaml, csv)")
	cmd.Flags().BoolVar(&opts.Pretty, "pretty", opts.Pretty, "Pretty print output")
	cmd.Flags().BoolVar(&opts.ShowSource, "show-source", false, "Include source text in output")
	cmd.Flags().BoolVar(&opts.Merge, "merge", false, "Merge results into single file")

	// Progress flags
	cmd.Flags().BoolVar(&opts.ShowProgress, "progress", opts.ShowProgress, "Show progress bar")
	cmd.Flags().BoolVar(&opts.ShowSummary, "summary", opts.ShowSummary, "Show processing summary")

	// Error handling flags
	cmd.Flags().BoolVar(&opts.ContinueOnError, "continue-on-error", opts.ContinueOnError, "Continue processing on errors")
	cmd.Flags().IntVar(&opts.MaxErrors, "max-errors", opts.MaxErrors, "Maximum errors before stopping")

	// Validation flags
	cmd.Flags().BoolVar(&opts.ValidateSchema, "validate-schema", true, "Validate against schema")
	cmd.Flags().BoolVar(&opts.ValidateExamples, "validate-examples", false, "Validate examples")

	// Mark required flags
	cmd.MarkFlagRequired("schema")
	cmd.MarkFlagRequired("output")

	return cmd
}

// runBatch executes the batch command
func runBatch(ctx context.Context, opts *BatchOptions, cfg *config.GlobalConfig, log *logger.Logger) error {
	log.WithOperation("batch").Info("Starting batch processing")

	// Validate options
	if err := validateBatchOptions(opts); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}

	// Create output directory
	if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Find input files
	inputFiles, err := findInputFiles(opts.Inputs, opts.Recursive, opts.Include, opts.Exclude)
	if err != nil {
		return fmt.Errorf("failed to find input files: %w", err)
	}

	if len(inputFiles) == 0 {
		return fmt.Errorf("no input files found")
	}

	log.WithOperation("batch").WithCount(len(inputFiles)).Info("Found input files")

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
		extractOpts = extractOpts.WithTemperature(float64(opts.Temperature))
	}
	// Note: MaxTokens should be configured via ModelConfig, not directly on options

	// Add examples
	if len(examples) > 0 {
		// Convert string examples to ExampleData - this would need proper implementation
		// For now, just skip examples until the ExampleData conversion is implemented
		log.WithOperation("batch").Warning("Examples not yet implemented in CLI")
	}

	// Process files in parallel
	results := make(chan BatchResult, len(inputFiles))
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, opts.Concurrency)
	
	startTime := time.Now()
	errorCount := 0

	for _, inputFile := range inputFiles {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		wg.Add(1)
		go func(file string) {
			defer wg.Done()
			
			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			result := processSingleFile(ctx, file, opts, extractOpts, log)
			results <- result

			if !result.Success {
				errorCount++
				if errorCount >= opts.MaxErrors && !opts.ContinueOnError {
					log.WithError(fmt.Errorf("max errors reached")).Error("Stopping batch processing")
					return
				}
			}
		}(inputFile)
	}

	// Close results channel when all goroutines complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect and process results
	var allResults []BatchResult
	var mergedResults []interface{}
	successCount := 0

	for result := range results {
		allResults = append(allResults, result)
		
		if result.Success {
			successCount++
			if opts.ShowProgress {
				log.Progress("Processing files", successCount, len(inputFiles))
			}
			
			// If merging, collect results
			if opts.Merge {
				// In a full implementation, we'd read and parse the result file
				// For now, just track that we have a result
				mergedResults = append(mergedResults, map[string]interface{}{
					"input_file": result.InputFile,
					"output_file": result.OutputFile,
				})
			}
		} else {
			log.WithError(result.Error).Errorf("Failed to process file: %s", result.InputFile)
		}
	}

	duration := time.Since(startTime)

	// Write merged results if requested
	if opts.Merge && len(mergedResults) > 0 {
		if err := writeMergedResults(mergedResults, opts); err != nil {
			log.WithError(err).Error("Failed to write merged results")
		}
	}

	// Show summary
	if opts.ShowSummary {
		showBatchSummary(allResults, duration, log)
	}

	log.WithOperation("batch").
		WithCount(successCount).
		WithDuration(duration).
		Success("Batch processing completed")

	if errorCount > 0 && !opts.ContinueOnError {
		return fmt.Errorf("batch processing failed with %d errors", errorCount)
	}

	return nil
}

// validateBatchOptions validates batch command options
func validateBatchOptions(opts *BatchOptions) error {
	if opts.Schema == "" {
		return fmt.Errorf("schema is required")
	}

	if opts.OutputDir == "" {
		return fmt.Errorf("output directory is required")
	}

	if opts.Concurrency < 1 || opts.Concurrency > 100 {
		return fmt.Errorf("concurrency must be between 1 and 100")
	}

	if opts.Temperature < 0 || opts.Temperature > 2.0 {
		return fmt.Errorf("temperature must be between 0.0 and 2.0")
	}

	if opts.MaxTokens < 1 {
		return fmt.Errorf("max-tokens must be positive")
	}

	if opts.MaxErrors < 1 {
		return fmt.Errorf("max-errors must be positive")
	}

	return nil
}

// findInputFiles finds all input files based on patterns and options
func findInputFiles(inputs []string, recursive bool, include, exclude []string) ([]string, error) {
	var files []string

	for _, input := range inputs {
		info, err := os.Stat(input)
		if err != nil {
			// Try as glob pattern
			matches, globErr := filepath.Glob(input)
			if globErr != nil {
				return nil, fmt.Errorf("failed to stat %s: %w", input, err)
			}
			files = append(files, matches...)
			continue
		}

		if info.IsDir() {
			dirFiles, err := findFilesInDir(input, recursive, include, exclude)
			if err != nil {
				return nil, err
			}
			files = append(files, dirFiles...)
		} else {
			if matchesPatterns(input, include, exclude) {
				files = append(files, input)
			}
		}
	}

	return files, nil
}

// findFilesInDir finds files in a directory based on patterns
func findFilesInDir(dir string, recursive bool, include, exclude []string) ([]string, error) {
	var files []string

	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories unless recursive
		if info.IsDir() {
			if !recursive && path != dir {
				return filepath.SkipDir
			}
			return nil
		}

		if matchesPatterns(path, include, exclude) {
			files = append(files, path)
		}

		return nil
	}

	if err := filepath.Walk(dir, walkFunc); err != nil {
		return nil, fmt.Errorf("failed to walk directory %s: %w", dir, err)
	}

	return files, nil
}

// matchesPatterns checks if a file matches include/exclude patterns
func matchesPatterns(file string, include, exclude []string) bool {
	// Check exclude patterns first
	for _, pattern := range exclude {
		if matched, _ := filepath.Match(pattern, filepath.Base(file)); matched {
			return false
		}
	}

	// If no include patterns, include all
	if len(include) == 0 {
		return true
	}

	// Check include patterns
	for _, pattern := range include {
		if matched, _ := filepath.Match(pattern, filepath.Base(file)); matched {
			return true
		}
	}

	return false
}

// processSingleFile processes a single file and returns the result
func processSingleFile(ctx context.Context, inputFile string, opts *BatchOptions, 
	extractOpts *langextract.ExtractOptions, log *logger.Logger) BatchResult {
	
	startTime := time.Now()
	
	result := BatchResult{
		InputFile: inputFile,
		Success:   false,
	}

	// Read input file
	input, err := os.ReadFile(inputFile)
	if err != nil {
		result.Error = fmt.Errorf("failed to read file: %w", err)
		result.Duration = time.Since(startTime)
		return result
	}

	// Perform extraction
	extracted, err := langextract.Extract(string(input), extractOpts)
	if err != nil {
		result.Error = fmt.Errorf("extraction failed: %w", err)
		result.Duration = time.Since(startTime)
		return result
	}

	result.NumExtracted = len(extracted.Extractions)

	// Create visualization options
	vizOpts := langextract.NewVisualizeOptions().
		WithFormat(opts.OutputFormat).
		WithContext(opts.ShowSource)

	// Generate output
	output, err := langextract.Visualize(extracted, vizOpts)
	if err != nil {
		result.Error = fmt.Errorf("failed to generate output: %w", err)
		result.Duration = time.Since(startTime)
		return result
	}

	// Determine output file path
	outputFile := generateOutputPath(inputFile, opts.OutputDir, opts.OutputFormat)
	result.OutputFile = outputFile

	// Write output
	if err := os.WriteFile(outputFile, []byte(output), 0644); err != nil {
		result.Error = fmt.Errorf("failed to write output: %w", err)
		result.Duration = time.Since(startTime)
		return result
	}

	result.Success = true
	result.Duration = time.Since(startTime)
	return result
}

// generateOutputPath generates an output file path based on input file and options
func generateOutputPath(inputFile, outputDir, format string) string {
	baseName := filepath.Base(inputFile)
	ext := filepath.Ext(baseName)
	nameWithoutExt := strings.TrimSuffix(baseName, ext)
	
	var outputExt string
	switch format {
	case "json":
		outputExt = ".json"
	case "yaml":
		outputExt = ".yaml"
	case "csv":
		outputExt = ".csv"
	case "html":
		outputExt = ".html"
	default:
		outputExt = ".json"
	}

	return filepath.Join(outputDir, nameWithoutExt+outputExt)
}

// writeMergedResults writes merged results to a single file
func writeMergedResults(results []interface{}, opts *BatchOptions) error {
	outputFile := filepath.Join(opts.OutputDir, "merged."+opts.OutputFormat)
	
	// In a full implementation, this would serialize the results properly
	// For now, just create a placeholder
	content := fmt.Sprintf("Merged results: %d files processed\n", len(results))
	
	return os.WriteFile(outputFile, []byte(content), 0644)
}

// showBatchSummary shows a summary of batch processing results
func showBatchSummary(results []BatchResult, duration time.Duration, log *logger.Logger) {
	successCount := 0
	errorCount := 0
	totalExtracted := 0

	for _, result := range results {
		if result.Success {
			successCount++
			totalExtracted += result.NumExtracted
		} else {
			errorCount++
		}
	}

	log.WithFields(map[string]interface{}{
		"total_files": len(results),
		"successful":  successCount,
		"failed":      errorCount,
		"extracted":   totalExtracted,
		"duration":    duration.String(),
	}).Info("Batch processing summary")
}

