package langextract

import (
	"context"
	"time"

	"github.com/sehwan505/langextract-go/pkg/extraction"
	"github.com/sehwan505/langextract-go/pkg/providers"
)

// ExtractOptions configures extraction behavior.
// This mirrors the parameters from the Python langextract.extract() function.
type ExtractOptions struct {
	// PromptDescription provides instructions for what to extract
	PromptDescription string

	// Examples provides few-shot learning examples
	Examples []*extraction.ExampleData

	// ModelID specifies which language model to use
	// Default: "gemini-2.5-flash"
	ModelID string

	// ModelConfig provides model-specific configuration
	ModelConfig *providers.ModelConfig

	// ExtractionPasses controls how many sequential extraction attempts to make
	// Default: 1
	ExtractionPasses int

	// ParallelProcessing enables concurrent processing for large documents
	// Default: false
	ParallelProcessing bool

	// MaxTokens limits the maximum tokens for model responses
	// Default: determined by model
	MaxTokens int

	// Temperature controls randomness in model responses (0.0 to 1.0)
	// Default: 0.0 (deterministic)
	Temperature float64

	// Timeout specifies maximum time for extraction operations
	// Default: 60 seconds
	Timeout time.Duration

	// Context for cancellation and deadlines
	Context context.Context

	// Schema defines the structure for extracted data
	Schema extraction.ExtractionSchema

	// ValidateOutput enables validation of extracted data against schema
	// Default: true
	ValidateOutput bool

	// RetryCount specifies number of retries on failure
	// Default: 2
	RetryCount int

	// DebugMode enables detailed logging and debugging
	// Default: false
	DebugMode bool
}

// NewExtractOptions creates ExtractOptions with sensible defaults.
func NewExtractOptions() *ExtractOptions {
	return &ExtractOptions{
		ModelID:            "gemini-2.5-flash",
		ExtractionPasses:   1,
		ParallelProcessing: false,
		Temperature:        0.0,
		Timeout:            60 * time.Second,
		Context:            context.Background(),
		ValidateOutput:     true,
		RetryCount:         2,
		DebugMode:          false,
	}
}

// WithPromptDescription sets the prompt description.
func (opts *ExtractOptions) WithPromptDescription(description string) *ExtractOptions {
	opts.PromptDescription = description
	return opts
}

// WithExamples sets the examples for few-shot learning.
func (opts *ExtractOptions) WithExamples(examples []*extraction.ExampleData) *ExtractOptions {
	opts.Examples = examples
	return opts
}

// WithModelID sets the model identifier.
func (opts *ExtractOptions) WithModelID(modelID string) *ExtractOptions {
	opts.ModelID = modelID
	return opts
}

// WithModelConfig sets the model configuration.
func (opts *ExtractOptions) WithModelConfig(config *providers.ModelConfig) *ExtractOptions {
	opts.ModelConfig = config
	return opts
}

// WithExtractionPasses sets the number of extraction passes.
func (opts *ExtractOptions) WithExtractionPasses(passes int) *ExtractOptions {
	opts.ExtractionPasses = passes
	return opts
}

// WithParallelProcessing enables or disables parallel processing.
func (opts *ExtractOptions) WithParallelProcessing(enabled bool) *ExtractOptions {
	opts.ParallelProcessing = enabled
	return opts
}

// WithTemperature sets the model temperature.
func (opts *ExtractOptions) WithTemperature(temperature float64) *ExtractOptions {
	opts.Temperature = temperature
	return opts
}

// WithTimeout sets the operation timeout.
func (opts *ExtractOptions) WithTimeout(timeout time.Duration) *ExtractOptions {
	opts.Timeout = timeout
	return opts
}

// WithContext sets the context for cancellation.
func (opts *ExtractOptions) WithContext(ctx context.Context) *ExtractOptions {
	opts.Context = ctx
	return opts
}

// WithSchema sets the extraction schema.
func (opts *ExtractOptions) WithSchema(schema extraction.ExtractionSchema) *ExtractOptions {
	opts.Schema = schema
	return opts
}

// WithValidation enables or disables output validation.
func (opts *ExtractOptions) WithValidation(enabled bool) *ExtractOptions {
	opts.ValidateOutput = enabled
	return opts
}

// WithRetryCount sets the number of retries.
func (opts *ExtractOptions) WithRetryCount(count int) *ExtractOptions {
	opts.RetryCount = count
	return opts
}

// WithDebugMode enables or disables debug mode.
func (opts *ExtractOptions) WithDebugMode(enabled bool) *ExtractOptions {
	opts.DebugMode = enabled
	return opts
}

// Validate checks if the options are valid.
func (opts *ExtractOptions) Validate() error {
	if opts.PromptDescription == "" {
		return NewValidationError("PromptDescription", "", "prompt description is required")
	}

	if opts.ModelID == "" {
		return NewValidationError("ModelID", "", "model ID is required")
	}

	if opts.ExtractionPasses < 1 {
		return NewValidationError("ExtractionPasses", string(rune(opts.ExtractionPasses)), "must be at least 1")
	}

	if opts.Temperature < 0.0 || opts.Temperature > 1.0 {
		return NewValidationError("Temperature", string(rune(int(opts.Temperature*100))), "must be between 0.0 and 1.0")
	}

	if opts.Timeout <= 0 {
		return NewValidationError("Timeout", opts.Timeout.String(), "must be positive")
	}

	if opts.RetryCount < 0 {
		return NewValidationError("RetryCount", string(rune(opts.RetryCount)), "must be non-negative")
	}

	return nil
}

// VisualizeOptions configures visualization behavior.
type VisualizeOptions struct {
	// Format specifies the output format (html, json, csv)
	// Default: "html"
	Format string

	// ShowConfidence includes confidence scores in visualization
	// Default: true
	ShowConfidence bool

	// ShowAlignment includes alignment status in visualization
	// Default: true
	ShowAlignment bool

	// IncludeContext shows surrounding text context
	// Default: true
	IncludeContext bool

	// ContextWindow specifies characters of context to show around extractions
	// Default: 50
	ContextWindow int

	// GroupByClass groups extractions by their class in output
	// Default: false
	GroupByClass bool

	// SortByPosition sorts extractions by their position in text
	// Default: true
	SortByPosition bool
}

// NewVisualizeOptions creates VisualizeOptions with sensible defaults.
func NewVisualizeOptions() *VisualizeOptions {
	return &VisualizeOptions{
		Format:         "html",
		ShowConfidence: true,
		ShowAlignment:  true,
		IncludeContext: true,
		ContextWindow:  50,
		GroupByClass:   false,
		SortByPosition: true,
	}
}

// WithFormat sets the output format.
func (opts *VisualizeOptions) WithFormat(format string) *VisualizeOptions {
	opts.Format = format
	return opts
}

// WithConfidence enables or disables confidence display.
func (opts *VisualizeOptions) WithConfidence(show bool) *VisualizeOptions {
	opts.ShowConfidence = show
	return opts
}

// WithAlignment enables or disables alignment status display.
func (opts *VisualizeOptions) WithAlignment(show bool) *VisualizeOptions {
	opts.ShowAlignment = show
	return opts
}

// WithContext enables or disables context display.
func (opts *VisualizeOptions) WithContext(include bool) *VisualizeOptions {
	opts.IncludeContext = include
	return opts
}

// WithContextWindow sets the context window size.
func (opts *VisualizeOptions) WithContextWindow(window int) *VisualizeOptions {
	opts.ContextWindow = window
	return opts
}

// WithGroupByClass enables or disables grouping by class.
func (opts *VisualizeOptions) WithGroupByClass(group bool) *VisualizeOptions {
	opts.GroupByClass = group
	return opts
}

// WithSortByPosition enables or disables sorting by position.
func (opts *VisualizeOptions) WithSortByPosition(sort bool) *VisualizeOptions {
	opts.SortByPosition = sort
	return opts
}

// Validate checks if the options are valid.
func (opts *VisualizeOptions) Validate() error {
	validFormats := map[string]bool{
		"html": true,
		"json": true,
		"csv":  true,
	}

	if !validFormats[opts.Format] {
		return NewValidationError("Format", opts.Format, "must be one of: html, json, csv")
	}

	if opts.ContextWindow < 0 {
		return NewValidationError("ContextWindow", string(rune(opts.ContextWindow)), "must be non-negative")
	}

	return nil
}