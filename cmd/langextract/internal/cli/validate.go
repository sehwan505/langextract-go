package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/sehwan505/langextract-go/cmd/langextract/internal/config"
	"github.com/sehwan505/langextract-go/cmd/langextract/internal/logger"
)

// ValidateOptions contains all options for the validate command
type ValidateOptions struct {
	// Schema validation
	Schema        string   // Schema file path
	SchemaFormat  string   // Schema format: yaml, json, auto
	SchemaDir     string   // Directory containing schemas
	
	// Example validation
	Examples      []string // Example files or directories
	ExampleFormat string   // Example format: yaml, json, text, auto
	
	// Test validation
	TestFiles     []string // Test files with expected results
	TestDir       string   // Directory containing test files
	
	// Validation options
	Strict        bool     // Enable strict validation
	AllowPartial  bool     // Allow partial matches in examples
	ShowDetails   bool     // Show detailed validation results
	ShowWarnings  bool     // Show validation warnings
	
	// Output options
	Output        string   // Output file for validation report
	Format        string   // Report format: text, json, yaml, html
	Quiet         bool     // Suppress non-error output
	Verbose       bool     // Enable verbose output
	
	// Provider options for testing
	Provider      string   // Provider to test examples against
	ModelID       string   // Model ID for testing
	TestExamples  bool     // Test examples against actual LLM
	
	// Performance options
	MaxConcurrency int     // Maximum concurrent validations
	Timeout        int     // Timeout per validation in seconds
}

// ValidationResult represents the result of a single validation
type ValidationResult struct {
	Type        string    `json:"type"`        // schema, example, test
	Target      string    `json:"target"`      // file or item being validated
	Valid       bool      `json:"valid"`       // validation result
	Errors      []string  `json:"errors"`      // validation errors
	Warnings    []string  `json:"warnings"`    // validation warnings
	Duration    time.Duration `json:"duration"` // validation duration
	Details     map[string]interface{} `json:"details,omitempty"` // additional details
}

// ValidationSummary represents the overall validation results
type ValidationSummary struct {
	TotalValidations int                 `json:"total_validations"`
	SuccessCount     int                 `json:"success_count"`
	ErrorCount       int                 `json:"error_count"`
	WarningCount     int                 `json:"warning_count"`
	Duration         time.Duration       `json:"duration"`
	Results          []ValidationResult  `json:"results"`
}

// NewValidateCommand creates the validate command
func NewValidateCommand(cfg *config.GlobalConfig, log *logger.Logger) *cobra.Command {
	opts := &ValidateOptions{
		SchemaFormat:   "auto",
		ExampleFormat:  "auto",
		Strict:         cfg.StrictValidation,
		AllowPartial:   false,
		ShowDetails:    true,
		ShowWarnings:   true,
		Format:         "text",
		Quiet:          cfg.Quiet,
		Verbose:        cfg.Verbose,
		Provider:       cfg.DefaultProvider,
		TestExamples:   false,
		MaxConcurrency: cfg.Concurrency,
		Timeout:        30,
	}

	cmd := &cobra.Command{
		Use:   "validate [flags]",
		Short: "Validate schemas, examples, and test cases",
		Long: `Validate extraction schemas, examples, and test cases to ensure they are
properly formatted and compatible with the extraction system. This command can
validate schema syntax, example format compliance, and test expected results.

Examples:
  # Validate a single schema file
  langextract validate --schema person.yaml
  
  # Validate all schemas in a directory
  langextract validate --schema-dir schemas/
  
  # Validate examples against a schema
  langextract validate --schema person.yaml --examples examples/
  
  # Validate test cases with expected results
  langextract validate --schema schema.yaml --test-files tests/*.yaml
  
  # Test examples against actual LLM (requires API key)
  langextract validate --schema schema.yaml --examples examples/ --test-examples --provider openai
  
  # Generate detailed validation report
  langextract validate --schema-dir schemas/ --examples examples/ --output report.html --format html
  
  # Strict validation with all checks
  langextract validate --schema schema.yaml --examples examples/ --strict --show-warnings`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runValidate(cmd.Context(), opts, cfg, log)
		},
	}

	// Schema flags
	cmd.Flags().StringVar(&opts.Schema, "schema", "", "Schema file path")
	cmd.Flags().StringVar(&opts.SchemaFormat, "schema-format", opts.SchemaFormat, "Schema format (yaml, json, auto)")
	cmd.Flags().StringVar(&opts.SchemaDir, "schema-dir", "", "Directory containing schemas")

	// Example flags
	cmd.Flags().StringArrayVar(&opts.Examples, "examples", nil, "Example files or directories")
	cmd.Flags().StringVar(&opts.ExampleFormat, "example-format", opts.ExampleFormat, "Example format (yaml, json, text, auto)")

	// Test flags
	cmd.Flags().StringArrayVar(&opts.TestFiles, "test-files", nil, "Test files with expected results")
	cmd.Flags().StringVar(&opts.TestDir, "test-dir", "", "Directory containing test files")

	// Validation flags
	cmd.Flags().BoolVar(&opts.Strict, "strict", opts.Strict, "Enable strict validation")
	cmd.Flags().BoolVar(&opts.AllowPartial, "allow-partial", opts.AllowPartial, "Allow partial matches in examples")
	cmd.Flags().BoolVar(&opts.ShowDetails, "show-details", opts.ShowDetails, "Show detailed validation results")
	cmd.Flags().BoolVar(&opts.ShowWarnings, "show-warnings", opts.ShowWarnings, "Show validation warnings")

	// Output flags
	cmd.Flags().StringVarP(&opts.Output, "output", "o", "", "Output file for validation report")
	cmd.Flags().StringVar(&opts.Format, "format", opts.Format, "Report format (text, json, yaml, html)")
	cmd.Flags().BoolVar(&opts.Quiet, "quiet", opts.Quiet, "Suppress non-error output")
	cmd.Flags().BoolVar(&opts.Verbose, "verbose", opts.Verbose, "Enable verbose output")

	// Provider flags
	cmd.Flags().StringVar(&opts.Provider, "provider", opts.Provider, "Provider to test examples against")
	cmd.Flags().StringVar(&opts.ModelID, "model", "", "Model ID for testing")
	cmd.Flags().BoolVar(&opts.TestExamples, "test-examples", opts.TestExamples, "Test examples against actual LLM")

	// Performance flags
	cmd.Flags().IntVar(&opts.MaxConcurrency, "max-concurrency", opts.MaxConcurrency, "Maximum concurrent validations")
	cmd.Flags().IntVar(&opts.Timeout, "timeout", opts.Timeout, "Timeout per validation in seconds")

	return cmd
}

// runValidate executes the validate command
func runValidate(ctx context.Context, opts *ValidateOptions, cfg *config.GlobalConfig, log *logger.Logger) error {
	log.WithOperation("validate").Info("Starting validation")

	// Validate options
	if err := validateValidateOptions(opts); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}

	startTime := time.Now()
	summary := &ValidationSummary{
		Results: []ValidationResult{},
	}

	// Validate schemas
	if opts.Schema != "" || opts.SchemaDir != "" {
		results, err := validateSchemas(ctx, opts, log)
		if err != nil {
			return fmt.Errorf("schema validation failed: %w", err)
		}
		summary.Results = append(summary.Results, results...)
	}

	// Validate examples
	if len(opts.Examples) > 0 {
		results, err := validateExamples(ctx, opts, log)
		if err != nil {
			return fmt.Errorf("example validation failed: %w", err)
		}
		summary.Results = append(summary.Results, results...)
	}

	// Validate test files
	if len(opts.TestFiles) > 0 || opts.TestDir != "" {
		results, err := validateTestFiles(ctx, opts, log)
		if err != nil {
			return fmt.Errorf("test validation failed: %w", err)
		}
		summary.Results = append(summary.Results, results...)
	}

	// Test examples against LLM if requested
	if opts.TestExamples && len(opts.Examples) > 0 {
		results, err := testExamplesWithLLM(ctx, opts, cfg, log)
		if err != nil {
			return fmt.Errorf("LLM testing failed: %w", err)
		}
		summary.Results = append(summary.Results, results...)
	}

	// Calculate summary statistics
	summary.Duration = time.Since(startTime)
	summary.TotalValidations = len(summary.Results)
	
	for _, result := range summary.Results {
		if result.Valid {
			summary.SuccessCount++
		} else {
			summary.ErrorCount++
		}
		summary.WarningCount += len(result.Warnings)
	}

	// Generate and output report
	if err := generateValidationReport(summary, opts); err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}

	// Show summary
	if !opts.Quiet {
		showValidationSummary(summary, log)
	}

	// Return error if any validations failed
	if summary.ErrorCount > 0 {
		return fmt.Errorf("validation failed: %d errors found", summary.ErrorCount)
	}

	log.Success("All validations passed")
	return nil
}

// validateValidateOptions validates the validate command options
func validateValidateOptions(opts *ValidateOptions) error {
	if opts.Schema == "" && opts.SchemaDir == "" && len(opts.Examples) == 0 && 
		len(opts.TestFiles) == 0 && opts.TestDir == "" {
		return fmt.Errorf("at least one validation target must be specified")
	}

	validFormats := []string{"text", "json", "yaml", "html"}
	if !contains(validFormats, opts.Format) {
		return fmt.Errorf("invalid format '%s', must be one of: %v", opts.Format, validFormats)
	}

	if opts.MaxConcurrency < 1 || opts.MaxConcurrency > 100 {
		return fmt.Errorf("max-concurrency must be between 1 and 100")
	}

	if opts.Timeout < 1 || opts.Timeout > 300 {
		return fmt.Errorf("timeout must be between 1 and 300 seconds")
	}

	return nil
}

// validateSchemas validates schema files
func validateSchemas(ctx context.Context, opts *ValidateOptions, log *logger.Logger) ([]ValidationResult, error) {
	var results []ValidationResult
	var schemaFiles []string

	// Collect schema files
	if opts.Schema != "" {
		schemaFiles = append(schemaFiles, opts.Schema)
	}

	if opts.SchemaDir != "" {
		files, err := findFilesInDirectory(opts.SchemaDir, []string{"*.yaml", "*.yml", "*.json"})
		if err != nil {
			return nil, fmt.Errorf("failed to find schemas in directory: %w", err)
		}
		schemaFiles = append(schemaFiles, files...)
	}

	log.WithCount(len(schemaFiles)).Info("Validating schemas")

	// Validate each schema file
	for _, schemaFile := range schemaFiles {
		result := validateSingleSchema(schemaFile, opts)
		results = append(results, result)

		if opts.Verbose {
			status := "PASS"
			if !result.Valid {
				status = "FAIL"
			}
			log.WithFile(schemaFile).Infof("Schema validation: %s", status)
		}
	}

	return results, nil
}

// validateSingleSchema validates a single schema file
func validateSingleSchema(schemaFile string, opts *ValidateOptions) ValidationResult {
	startTime := time.Now()
	result := ValidationResult{
		Type:     "schema",
		Target:   schemaFile,
		Valid:    true,
		Errors:   []string{},
		Warnings: []string{},
		Details:  make(map[string]interface{}),
	}

	// Check if file exists
	if _, err := os.Stat(schemaFile); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Schema file not found: %v", err))
		result.Duration = time.Since(startTime)
		return result
	}

	// Read schema file
	content, err := os.ReadFile(schemaFile)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to read schema: %v", err))
		result.Duration = time.Since(startTime)
		return result
	}

	// Determine format
	format := opts.SchemaFormat
	if format == "auto" {
		ext := strings.ToLower(filepath.Ext(schemaFile))
		switch ext {
		case ".yaml", ".yml":
			format = "yaml"
		case ".json":
			format = "json"
		default:
			result.Warnings = append(result.Warnings, "Unknown schema format, assuming YAML")
			format = "yaml"
		}
	}

	// Validate schema syntax
	if err := validateSchemaContent(content, format, opts.Strict); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Schema validation error: %v", err))
	}

	result.Details["format"] = format
	result.Details["size"] = len(content)
	result.Duration = time.Since(startTime)

	return result
}

// validateSchemaContent validates the content of a schema
func validateSchemaContent(content []byte, format string, strict bool) error {
	// Placeholder validation logic
	// In a full implementation, this would parse and validate the schema structure
	if len(content) == 0 {
		return fmt.Errorf("schema is empty")
	}

	// Basic format checks
	switch format {
	case "yaml":
		// Check for basic YAML syntax
		if !strings.Contains(string(content), ":") {
			return fmt.Errorf("invalid YAML format: no key-value pairs found")
		}
	case "json":
		// Check for basic JSON syntax
		if !strings.HasPrefix(strings.TrimSpace(string(content)), "{") {
			return fmt.Errorf("invalid JSON format: does not start with '{'")
		}
	}

	return nil
}

// validateExamples validates example files
func validateExamples(ctx context.Context, opts *ValidateOptions, log *logger.Logger) ([]ValidationResult, error) {
	var results []ValidationResult
	var exampleFiles []string

	// Collect example files
	for _, example := range opts.Examples {
		if info, err := os.Stat(example); err == nil {
			if info.IsDir() {
				files, err := findFilesInDirectory(example, []string{"*.txt", "*.md", "*.yaml", "*.json"})
				if err != nil {
					return nil, fmt.Errorf("failed to find examples in directory: %w", err)
				}
				exampleFiles = append(exampleFiles, files...)
			} else {
				exampleFiles = append(exampleFiles, example)
			}
		}
	}

	log.WithCount(len(exampleFiles)).Info("Validating examples")

	// Validate each example file
	for _, exampleFile := range exampleFiles {
		result := validateSingleExample(exampleFile, opts)
		results = append(results, result)

		if opts.Verbose {
			status := "PASS"
			if !result.Valid {
				status = "FAIL"
			}
			log.WithFile(exampleFile).Infof("Example validation: %s", status)
		}
	}

	return results, nil
}

// validateSingleExample validates a single example file
func validateSingleExample(exampleFile string, opts *ValidateOptions) ValidationResult {
	startTime := time.Now()
	result := ValidationResult{
		Type:     "example",
		Target:   exampleFile,
		Valid:    true,
		Errors:   []string{},
		Warnings: []string{},
		Details:  make(map[string]interface{}),
	}

	// Read example file
	content, err := os.ReadFile(exampleFile)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to read example: %v", err))
		result.Duration = time.Since(startTime)
		return result
	}

	// Basic validation
	if len(content) == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, "Example file is empty")
	}

	// Check for minimum content requirements
	if len(strings.TrimSpace(string(content))) < 10 {
		result.Warnings = append(result.Warnings, "Example content is very short")
	}

	result.Details["size"] = len(content)
	result.Details["lines"] = strings.Count(string(content), "\n") + 1
	result.Duration = time.Since(startTime)

	return result
}

// validateTestFiles validates test files with expected results
func validateTestFiles(ctx context.Context, opts *ValidateOptions, log *logger.Logger) ([]ValidationResult, error) {
	var results []ValidationResult
	
	// Placeholder implementation
	log.WithOperation("validate").Info("Test file validation not yet implemented")
	
	return results, nil
}

// testExamplesWithLLM tests examples against actual LLM
func testExamplesWithLLM(ctx context.Context, opts *ValidateOptions, cfg *config.GlobalConfig, log *logger.Logger) ([]ValidationResult, error) {
	var results []ValidationResult
	
	// Placeholder implementation
	log.WithOperation("validate").Info("LLM testing not yet implemented")
	
	return results, nil
}

// generateValidationReport generates a validation report in the specified format
func generateValidationReport(summary *ValidationSummary, opts *ValidateOptions) error {
	if opts.Output == "" || opts.Output == "-" {
		// Output to stdout is handled in showValidationSummary
		return nil
	}

	var content string
	var err error

	switch opts.Format {
	case "json":
		content, err = generateJSONReport(summary)
	case "yaml":
		content, err = generateYAMLReport(summary)
	case "html":
		content, err = generateHTMLReport(summary)
	default:
		content, err = generateTextReport(summary)
	}

	if err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}

	// Write report to file
	if err := os.WriteFile(opts.Output, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}

	return nil
}

// generateTextReport generates a text format validation report
func generateTextReport(summary *ValidationSummary) (string, error) {
	var report strings.Builder

	report.WriteString("Validation Report\n")
	report.WriteString("================\n\n")
	report.WriteString(fmt.Sprintf("Total Validations: %d\n", summary.TotalValidations))
	report.WriteString(fmt.Sprintf("Successful: %d\n", summary.SuccessCount))
	report.WriteString(fmt.Sprintf("Failed: %d\n", summary.ErrorCount))
	report.WriteString(fmt.Sprintf("Warnings: %d\n", summary.WarningCount))
	report.WriteString(fmt.Sprintf("Duration: %v\n\n", summary.Duration))

	// Details for each validation
	for _, result := range summary.Results {
		report.WriteString(fmt.Sprintf("%s: %s\n", strings.ToUpper(result.Type), result.Target))
		status := "PASS"
		if !result.Valid {
			status = "FAIL"
		}
		report.WriteString(fmt.Sprintf("  Status: %s\n", status))
		report.WriteString(fmt.Sprintf("  Duration: %v\n", result.Duration))

		if len(result.Errors) > 0 {
			report.WriteString("  Errors:\n")
			for _, err := range result.Errors {
				report.WriteString(fmt.Sprintf("    - %s\n", err))
			}
		}

		if len(result.Warnings) > 0 {
			report.WriteString("  Warnings:\n")
			for _, warning := range result.Warnings {
				report.WriteString(fmt.Sprintf("    - %s\n", warning))
			}
		}

		report.WriteString("\n")
	}

	return report.String(), nil
}

// generateJSONReport generates a JSON format validation report
func generateJSONReport(summary *ValidationSummary) (string, error) {
	// Placeholder JSON serialization
	return fmt.Sprintf(`{"total_validations": %d, "success_count": %d, "error_count": %d}`,
		summary.TotalValidations, summary.SuccessCount, summary.ErrorCount), nil
}

// generateYAMLReport generates a YAML format validation report
func generateYAMLReport(summary *ValidationSummary) (string, error) {
	// Placeholder YAML serialization
	return fmt.Sprintf("total_validations: %d\nsuccess_count: %d\nerror_count: %d\n",
		summary.TotalValidations, summary.SuccessCount, summary.ErrorCount), nil
}

// generateHTMLReport generates an HTML format validation report
func generateHTMLReport(summary *ValidationSummary) (string, error) {
	// Placeholder HTML generation
	return fmt.Sprintf("<html><body><h1>Validation Report</h1><p>Total: %d, Success: %d, Errors: %d</p></body></html>",
		summary.TotalValidations, summary.SuccessCount, summary.ErrorCount), nil
}

// showValidationSummary shows validation summary in the console
func showValidationSummary(summary *ValidationSummary, log *logger.Logger) {
	log.WithFields(map[string]interface{}{
		"total":    summary.TotalValidations,
		"success":  summary.SuccessCount,
		"errors":   summary.ErrorCount,
		"warnings": summary.WarningCount,
		"duration": summary.Duration.String(),
	}).Info("Validation summary")
}

// findFilesInDirectory finds files in a directory matching patterns
func findFilesInDirectory(dir string, patterns []string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Check if file matches any pattern
		for _, pattern := range patterns {
			if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
				files = append(files, path)
				break
			}
		}

		return nil
	})

	return files, err
}