package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/sehwan505/langextract-go/cmd/langextract/internal/config"
	"github.com/sehwan505/langextract-go/cmd/langextract/internal/logger"
	"github.com/sehwan505/langextract-go/pkg/document"
	"github.com/sehwan505/langextract-go/pkg/langextract"
)

// VisualizeOptions contains all options for the visualize command
type VisualizeOptions struct {
	// Input options
	Input      string   // Input file or directory containing extraction results
	InputFiles []string // Specific input files to process
	Merge      bool     // Merge multiple results into single visualization

	// Output options
	Output       string // Output file or directory
	Format       string // Output format: html, json, yaml, csv, markdown, text
	Pretty       bool   // Pretty print output
	Template     string // Custom template file
	Theme        string // Theme for HTML output

	// Content options
	ShowSource    bool // Include source text in output
	ShowMetadata  bool // Include metadata in output
	ShowTimestamp bool // Include processing timestamp
	ContextWindow int  // Context window around extractions

	// Formatting options
	HighlightExtractions bool   // Highlight extractions in source text
	GroupByType         bool   // Group extractions by type
	SortBy              string // Sort extractions by: position, type, confidence
	FilterTypes         []string // Include only specified extraction types
	ExcludeTypes        []string // Exclude specified extraction types

	// HTML-specific options
	Interactive bool   // Generate interactive HTML
	Standalone  bool   // Generate standalone HTML file
	CSSFile     string // External CSS file
	JSFile      string // External JavaScript file

	// CSV-specific options
	CSVDelimiter string   // CSV delimiter character
	CSVHeaders   []string // Custom CSV headers
	FlattenJSON  bool     // Flatten JSON structures in CSV

	// Advanced options
	CustomFields map[string]string // Custom fields to add to output
	Validate     bool              // Validate input before visualization
}

// NewVisualizeCommand creates the visualize command
func NewVisualizeCommand(cfg *config.GlobalConfig, log *logger.Logger) *cobra.Command {
	opts := &VisualizeOptions{
		Format:               cfg.DefaultFormat,
		Pretty:               cfg.PrettyPrint,
		Theme:                "default",
		ShowSource:           true,
		ShowMetadata:         true,
		ShowTimestamp:        true,
		ContextWindow:        100,
		HighlightExtractions: true,
		GroupByType:          false,
		SortBy:               "position",
		Interactive:          true,
		Standalone:           true,
		CSVDelimiter:         ",",
		FlattenJSON:          true,
		Validate:             true,
	}

	cmd := &cobra.Command{
		Use:   "visualize [flags] [input]",
		Short: "Generate interactive visualizations from extraction results",
		Long: `Generate interactive visualizations from extraction results in various formats.
Input can be JSON files containing extraction results, directories with multiple results,
or merged result files from batch processing.

Examples:
  # Generate HTML visualization from JSON results
  langextract visualize --format html --output results.html results.json
  
  # Generate CSV export with custom headers
  langextract visualize --format csv --csv-headers "text,type,confidence" --output data.csv results.json
  
  # Create interactive HTML with custom theme
  langextract visualize --format html --interactive --theme dark --output report.html results/
  
  # Generate markdown report with source context
  langextract visualize --format markdown --show-source --context-window 200 --output report.md results.json
  
  # Export filtered results to YAML
  langextract visualize --format yaml --filter-types "person,organization" --output filtered.yaml results.json
  
  # Merge multiple result files into single visualization
  langextract visualize --merge --format html --output merged.html results/*.json`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.Input = args[0]
			}
			return runVisualize(cmd.Context(), opts, cfg, log)
		},
	}

	// Input flags
	cmd.Flags().StringVar(&opts.Input, "input", "", "Input file or directory with extraction results")
	cmd.Flags().StringArrayVar(&opts.InputFiles, "files", nil, "Specific input files to process")
	cmd.Flags().BoolVar(&opts.Merge, "merge", false, "Merge multiple results into single visualization")

	// Output flags
	cmd.Flags().StringVarP(&opts.Output, "output", "o", "", "Output file or directory")
	cmd.Flags().StringVar(&opts.Format, "format", opts.Format, "Output format (html, json, yaml, csv, markdown, text)")
	cmd.Flags().BoolVar(&opts.Pretty, "pretty", opts.Pretty, "Pretty print output")
	cmd.Flags().StringVar(&opts.Template, "template", "", "Custom template file")
	cmd.Flags().StringVar(&opts.Theme, "theme", opts.Theme, "Theme for HTML output (default, dark, light)")

	// Content flags
	cmd.Flags().BoolVar(&opts.ShowSource, "show-source", opts.ShowSource, "Include source text in output")
	cmd.Flags().BoolVar(&opts.ShowMetadata, "show-metadata", opts.ShowMetadata, "Include metadata in output")
	cmd.Flags().BoolVar(&opts.ShowTimestamp, "show-timestamp", opts.ShowTimestamp, "Include processing timestamp")
	cmd.Flags().IntVar(&opts.ContextWindow, "context-window", opts.ContextWindow, "Context window around extractions")

	// Formatting flags
	cmd.Flags().BoolVar(&opts.HighlightExtractions, "highlight", opts.HighlightExtractions, "Highlight extractions in source text")
	cmd.Flags().BoolVar(&opts.GroupByType, "group-by-type", opts.GroupByType, "Group extractions by type")
	cmd.Flags().StringVar(&opts.SortBy, "sort-by", opts.SortBy, "Sort extractions by: position, type, confidence")
	cmd.Flags().StringArrayVar(&opts.FilterTypes, "filter-types", nil, "Include only specified extraction types")
	cmd.Flags().StringArrayVar(&opts.ExcludeTypes, "exclude-types", nil, "Exclude specified extraction types")

	// HTML-specific flags
	cmd.Flags().BoolVar(&opts.Interactive, "interactive", opts.Interactive, "Generate interactive HTML")
	cmd.Flags().BoolVar(&opts.Standalone, "standalone", opts.Standalone, "Generate standalone HTML file")
	cmd.Flags().StringVar(&opts.CSSFile, "css", "", "External CSS file")
	cmd.Flags().StringVar(&opts.JSFile, "js", "", "External JavaScript file")

	// CSV-specific flags
	cmd.Flags().StringVar(&opts.CSVDelimiter, "csv-delimiter", opts.CSVDelimiter, "CSV delimiter character")
	cmd.Flags().StringArrayVar(&opts.CSVHeaders, "csv-headers", nil, "Custom CSV headers")
	cmd.Flags().BoolVar(&opts.FlattenJSON, "flatten-json", opts.FlattenJSON, "Flatten JSON structures in CSV")

	// Advanced flags
	cmd.Flags().BoolVar(&opts.Validate, "validate", opts.Validate, "Validate input before visualization")

	return cmd
}

// runVisualize executes the visualize command
func runVisualize(ctx context.Context, opts *VisualizeOptions, cfg *config.GlobalConfig, log *logger.Logger) error {
	log.WithOperation("visualize").Info("Starting visualization")

	// Validate options
	if err := validateVisualizeOptions(opts); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}

	// Find input files
	inputFiles, err := findVisualizationInputs(opts)
	if err != nil {
		return fmt.Errorf("failed to find input files: %w", err)
	}

	if len(inputFiles) == 0 {
		return fmt.Errorf("no input files found")
	}

	log.WithOperation("visualize").WithCount(len(inputFiles)).Info("Found input files")

	// Process input files
	var results []*document.AnnotatedDocument
	for _, inputFile := range inputFiles {
		result, err := loadExtractionResults(inputFile)
		if err != nil {
			log.WithError(err).WithFile(inputFile).Warning("Failed to load extraction results")
			continue
		}

		// Validate if requested
		if opts.Validate {
			if err := validateExtractionResults(result); err != nil {
				log.WithError(err).WithFile(inputFile).Warning("Validation failed")
				continue
			}
		}

		results = append(results, result)
	}

	if len(results) == 0 {
		return fmt.Errorf("no valid extraction results found")
	}

	// Create visualization options
	vizOpts := createVisualizationOptions(opts)

	// Generate visualization
	var output string
	if opts.Merge && len(results) > 1 {
		// Merge results and visualize
		merged := mergeExtractionResults(results)
		output, err = langextract.Visualize(merged, vizOpts)
	} else if len(results) == 1 {
		// Single result visualization
		output, err = langextract.Visualize(results[0], vizOpts)
	} else {
		// Multiple separate visualizations
		return generateMultipleVisualizations(results, opts, vizOpts, log)
	}

	if err != nil {
		return fmt.Errorf("failed to generate visualization: %w", err)
	}

	// Write output
	if err := writeVisualizationOutput(output, opts.Output, opts.Format); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	log.Success("Visualization completed successfully")
	return nil
}

// validateVisualizeOptions validates visualize command options
func validateVisualizeOptions(opts *VisualizeOptions) error {
	if opts.Input == "" && len(opts.InputFiles) == 0 {
		return fmt.Errorf("input is required")
	}

	validFormats := []string{"html", "json", "yaml", "csv", "markdown", "text"}
	if !contains(validFormats, opts.Format) {
		return fmt.Errorf("invalid format '%s', must be one of: %v", opts.Format, validFormats)
	}

	if opts.ContextWindow < 0 {
		return fmt.Errorf("context-window must be non-negative")
	}

	validSortBy := []string{"position", "type", "confidence"}
	if !contains(validSortBy, opts.SortBy) {
		return fmt.Errorf("invalid sort-by '%s', must be one of: %v", opts.SortBy, validSortBy)
	}

	validThemes := []string{"default", "dark", "light"}
	if !contains(validThemes, opts.Theme) {
		return fmt.Errorf("invalid theme '%s', must be one of: %v", opts.Theme, validThemes)
	}

	return nil
}

// findVisualizationInputs finds input files for visualization
func findVisualizationInputs(opts *VisualizeOptions) ([]string, error) {
	var files []string

	// Use specific files if provided
	if len(opts.InputFiles) > 0 {
		return opts.InputFiles, nil
	}

	// Use input parameter
	if opts.Input == "" {
		return nil, fmt.Errorf("no input specified")
	}

	info, err := os.Stat(opts.Input)
	if err != nil {
		return nil, fmt.Errorf("failed to stat input %s: %w", opts.Input, err)
	}

	if info.IsDir() {
		// Find JSON files in directory
		err := filepath.Walk(opts.Input, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".json") {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("failed to walk directory: %w", err)
		}
	} else {
		files = append(files, opts.Input)
	}

	return files, nil
}

// loadExtractionResults loads extraction results from a file
func loadExtractionResults(filename string) (*document.AnnotatedDocument, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	
	// Suppress unused variable warning for now
	_ = data

	// For now, create a placeholder AnnotatedDocument
	// In a full implementation, this would parse the JSON and reconstruct the document
	doc := &document.AnnotatedDocument{
		// Document: ...
		// Extractions: ...
	}

	return doc, nil
}

// validateExtractionResults validates extraction results
func validateExtractionResults(result *document.AnnotatedDocument) error {
	// Placeholder validation logic
	if result == nil {
		return fmt.Errorf("result is nil")
	}
	return nil
}

// createVisualizationOptions creates visualization options from command options
func createVisualizationOptions(opts *VisualizeOptions) *langextract.VisualizeOptions {
	vizOpts := langextract.NewVisualizeOptions().
		WithFormat(opts.Format).
		WithPretty(opts.Pretty).
		WithIncludeText(opts.ShowSource)

	// Add context window if specified
	if opts.ContextWindow > 0 {
		vizOpts = vizOpts.WithContextChars(opts.ContextWindow)
	}

	// Add filtering options
	if len(opts.FilterTypes) > 0 {
		vizOpts = vizOpts.WithFilterClasses(opts.FilterTypes)
	}

	// Add CSV-specific options
	if opts.Format == "csv" {
		vizOpts = vizOpts.WithCSVDelimiter(opts.CSVDelimiter)
	}

	return vizOpts
}

// mergeExtractionResults merges multiple extraction results into one
func mergeExtractionResults(results []*document.AnnotatedDocument) *document.AnnotatedDocument {
	// Placeholder merge logic
	// In a full implementation, this would properly merge the documents
	return results[0]
}

// generateMultipleVisualizations generates separate visualizations for multiple results
func generateMultipleVisualizations(results []*document.AnnotatedDocument, 
	opts *VisualizeOptions, vizOpts *langextract.VisualizeOptions, log *logger.Logger) error {
	
	if opts.Output == "" {
		opts.Output = "visualizations"
	}

	// Create output directory
	if err := os.MkdirAll(opts.Output, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	for i, result := range results {
		output, err := langextract.Visualize(result, vizOpts)
		if err != nil {
			log.WithError(err).Warning("Failed to generate visualization")
			continue
		}

		// Generate output filename
		outputFile := filepath.Join(opts.Output, fmt.Sprintf("visualization_%d.%s", i+1, opts.Format))
		
		if err := writeVisualizationOutput(output, outputFile, opts.Format); err != nil {
			log.WithError(err).WithFile(outputFile).Warning("Failed to write visualization")
			continue
		}
	}

	return nil
}

// writeVisualizationOutput writes visualization output to file
func writeVisualizationOutput(output, outputPath, format string) error {
	if outputPath == "" || outputPath == "-" {
		// Write to stdout
		fmt.Print(output)
		return nil
	}

	// Ensure output has correct extension
	if filepath.Ext(outputPath) == "" {
		switch format {
		case "html":
			outputPath += ".html"
		case "json":
			outputPath += ".json"
		case "yaml":
			outputPath += ".yaml"
		case "csv":
			outputPath += ".csv"
		case "markdown":
			outputPath += ".md"
		default:
			outputPath += ".txt"
		}
	}

	// Create output directory if needed
	if dir := filepath.Dir(outputPath); dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	// Write file
	if err := os.WriteFile(outputPath, []byte(output), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}