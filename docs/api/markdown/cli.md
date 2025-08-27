# cli

Package: `github.com/sehwan505/langextract-go/cmd/langextract/internal/cli`

```go
package cli // import "github.com/sehwan505/langextract-go/cmd/langextract/internal/cli"


FUNCTIONS

func NewBatchCommand(cfg *config.GlobalConfig, log *logger.Logger) *cobra.Command
    NewBatchCommand creates the batch command

func NewConfigCommand(cfg *config.GlobalConfig, log *logger.Logger) *cobra.Command
    NewConfigCommand creates the config command with subcommands

func NewConfigEditCommand(cfg *config.GlobalConfig, log *logger.Logger, opts *ConfigOptions) *cobra.Command
    NewConfigEditCommand creates the config edit subcommand

func NewConfigGetCommand(cfg *config.GlobalConfig, log *logger.Logger, opts *ConfigOptions) *cobra.Command
    NewConfigGetCommand creates the config get subcommand

func NewConfigInitCommand(cfg *config.GlobalConfig, log *logger.Logger, opts *ConfigOptions) *cobra.Command
    NewConfigInitCommand creates the config init subcommand

func NewConfigListCommand(cfg *config.GlobalConfig, log *logger.Logger, opts *ConfigOptions) *cobra.Command
    NewConfigListCommand creates the config list subcommand

func NewConfigSetCommand(cfg *config.GlobalConfig, log *logger.Logger, opts *ConfigOptions) *cobra.Command
    NewConfigSetCommand creates the config set subcommand

func NewConfigUnsetCommand(cfg *config.GlobalConfig, log *logger.Logger, opts *ConfigOptions) *cobra.Command
    NewConfigUnsetCommand creates the config unset subcommand

func NewConfigValidateCommand(cfg *config.GlobalConfig, log *logger.Logger, opts *ConfigOptions) *cobra.Command
    NewConfigValidateCommand creates the config validate subcommand

func NewExtractCommand(cfg *config.GlobalConfig, log *logger.Logger) *cobra.Command
    NewExtractCommand creates the extract command

func NewProvidersCommand(cfg *config.GlobalConfig, log *logger.Logger) *cobra.Command
    NewProvidersCommand creates the providers command with subcommands

func NewProvidersConfigCommand(cfg *config.GlobalConfig, log *logger.Logger, opts *ProvidersOptions) *cobra.Command
    NewProvidersConfigCommand creates the providers config subcommand

func NewProvidersInstallCommand(cfg *config.GlobalConfig, log *logger.Logger, opts *ProvidersOptions) *cobra.Command
    NewProvidersInstallCommand creates the providers install subcommand

func NewProvidersListCommand(cfg *config.GlobalConfig, log *logger.Logger, opts *ProvidersOptions) *cobra.Command
    NewProvidersListCommand creates the providers list subcommand

func NewProvidersTestCommand(cfg *config.GlobalConfig, log *logger.Logger, opts *ProvidersOptions) *cobra.Command
    NewProvidersTestCommand creates the providers test subcommand

func NewRootCommand(cfg *config.GlobalConfig, log *logger.Logger, version *VersionInfo) *cobra.Command
    NewRootCommand creates the root langextract command with all subcommands

func NewValidateCommand(cfg *config.GlobalConfig, log *logger.Logger) *cobra.Command
    NewValidateCommand creates the validate command

func NewVisualizeCommand(cfg *config.GlobalConfig, log *logger.Logger) *cobra.Command
    NewVisualizeCommand creates the visualize command


TYPES

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
    BatchOptions contains all options for the batch command

type BatchResult struct {
	InputFile    string
	OutputFile   string
	Success      bool
	Error        error
	Duration     time.Duration
	NumExtracted int
}
    BatchResult represents the result of processing a single file

type ConfigOptions struct {
	// Operations
	Operation string // get, set, unset, list, init, edit, validate

	// Target configuration
	Key   string // Configuration key to operate on
	Value string // Value to set

	// Scope
	Global bool // Operate on global configuration
	Local  bool // Operate on local (project) configuration
	User   bool // Operate on user configuration

	// File operations
	ConfigFile   string // Specific configuration file
	CreateConfig bool   // Create configuration file if it doesn't exist
	BackupConfig bool   // Create backup before modifying

	// Output options
	Output     string // Output file
	Format     string // Output format: yaml, json, table
	ShowAll    bool   // Show all configuration values
	ShowOrigin bool   // Show where each value is defined

	// Validation
	Validate bool // Validate configuration after changes
	DryRun   bool // Show what would be changed without applying
}
    ConfigOptions contains all options for the config command

type ExtractOptions struct {
	// Input options
	Input     string   // Input file or URL
	InputType string   // Input type: auto, file, url, text
	Schema    string   // Schema file path
	Examples  []string // Example files

	// Model options
	ModelID     string  // Model ID to use
	Temperature float32 // Temperature for generation
	MaxTokens   int     // Maximum tokens to generate

	// Processing options
	ContextWindow int // Context window size for chunking
	ChunkSize     int // Chunk size for large documents
	ChunkOverlap  int // Overlap between chunks

	// Output options
	Output     string // Output file (default: stdout)
	Format     string // Output format
	Pretty     bool   // Pretty print
	ShowSource bool   // Include source text in output

	// Validation options
	ValidateSchema   bool // Validate against schema
	ValidateExamples bool // Validate examples
}
    ExtractOptions contains all options for the extract command

type ProviderInfo struct {
	Name       string                 `json:"name"`
	Available  bool                   `json:"available"`
	Models     []string               `json:"models"`
	Aliases    map[string]string      `json:"aliases"`
	Config     map[string]interface{} `json:"config"`
	Status     string                 `json:"status"`
	LastTested *time.Time             `json:"last_tested,omitempty"`
	TestResult *TestResult            `json:"test_result,omitempty"`
}
    ProviderInfo represents information about a language model provider

type ProvidersOptions struct {
	// Operation
	Operation string // list, test, config, install, uninstall

	// Provider selection
	Provider  string   // Specific provider to operate on
	Providers []string // Multiple providers to operate on
	All       bool     // Apply operation to all providers

	// Test options
	TestModel   string // Model to test
	TestPrompt  string // Custom test prompt
	TestTimeout int    // Test timeout in seconds
	QuickTest   bool   // Run quick test instead of full test

	// Output options
	Output     string // Output file
	Format     string // Output format: table, json, yaml
	ShowConfig bool   // Show provider configuration
	ShowModels bool   // Show available models
	Detailed   bool   // Show detailed information

	// Configuration options
	SetConfig   map[string]string // Configuration values to set
	UnsetConfig []string          // Configuration keys to unset
	Interactive bool              // Interactive configuration

	// Installation options
	InstallDir string // Installation directory for providers
	Force      bool   // Force installation/uninstallation
}
    ProvidersOptions contains all options for the providers command

type TestResult struct {
	Success   bool                   `json:"success"`
	Duration  time.Duration          `json:"duration"`
	Error     string                 `json:"error,omitempty"`
	ModelInfo map[string]interface{} `json:"model_info,omitempty"`
	Metrics   map[string]interface{} `json:"metrics,omitempty"`
}
    TestResult represents the result of testing a provider

type ValidateOptions struct {
	// Schema validation
	Schema       string // Schema file path
	SchemaFormat string // Schema format: yaml, json, auto
	SchemaDir    string // Directory containing schemas

	// Example validation
	Examples      []string // Example files or directories
	ExampleFormat string   // Example format: yaml, json, text, auto

	// Test validation
	TestFiles []string // Test files with expected results
	TestDir   string   // Directory containing test files

	// Validation options
	Strict       bool // Enable strict validation
	AllowPartial bool // Allow partial matches in examples
	ShowDetails  bool // Show detailed validation results
	ShowWarnings bool // Show validation warnings

	// Output options
	Output  string // Output file for validation report
	Format  string // Report format: text, json, yaml, html
	Quiet   bool   // Suppress non-error output
	Verbose bool   // Enable verbose output

	// Provider options for testing
	Provider     string // Provider to test examples against
	ModelID      string // Model ID for testing
	TestExamples bool   // Test examples against actual LLM

	// Performance options
	MaxConcurrency int // Maximum concurrent validations
	Timeout        int // Timeout per validation in seconds
}
    ValidateOptions contains all options for the validate command

type ValidationResult struct {
	Type     string                 `json:"type"`              // schema, example, test
	Target   string                 `json:"target"`            // file or item being validated
	Valid    bool                   `json:"valid"`             // validation result
	Errors   []string               `json:"errors"`            // validation errors
	Warnings []string               `json:"warnings"`          // validation warnings
	Duration time.Duration          `json:"duration"`          // validation duration
	Details  map[string]interface{} `json:"details,omitempty"` // additional details
}
    ValidationResult represents the result of a single validation

type ValidationSummary struct {
	TotalValidations int                `json:"total_validations"`
	SuccessCount     int                `json:"success_count"`
	ErrorCount       int                `json:"error_count"`
	WarningCount     int                `json:"warning_count"`
	Duration         time.Duration      `json:"duration"`
	Results          []ValidationResult `json:"results"`
}
    ValidationSummary represents the overall validation results

type VersionInfo struct {
	Version string
	Commit  string
	Date    string
}
    VersionInfo contains build-time version information

type VisualizeOptions struct {
	// Input options
	Input      string   // Input file or directory containing extraction results
	InputFiles []string // Specific input files to process
	Merge      bool     // Merge multiple results into single visualization

	// Output options
	Output   string // Output file or directory
	Format   string // Output format: html, json, yaml, csv, markdown, text
	Pretty   bool   // Pretty print output
	Template string // Custom template file
	Theme    string // Theme for HTML output

	// Content options
	ShowSource    bool // Include source text in output
	ShowMetadata  bool // Include metadata in output
	ShowTimestamp bool // Include processing timestamp
	ContextWindow int  // Context window around extractions

	// Formatting options
	HighlightExtractions bool     // Highlight extractions in source text
	GroupByType          bool     // Group extractions by type
	SortBy               string   // Sort extractions by: position, type, confidence
	FilterTypes          []string // Include only specified extraction types
	ExcludeTypes         []string // Exclude specified extraction types

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
    VisualizeOptions contains all options for the visualize command

```
