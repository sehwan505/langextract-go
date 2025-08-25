package prompt

import (
	"context"
	"fmt"

	"github.com/sehwan505/langextract-go/pkg/extraction"
)

// PromptBuilder defines the interface for building prompts for language models
type PromptBuilder interface {
	// BuildPrompt constructs a prompt for the given task and text
	BuildPrompt(ctx context.Context, task *ExtractionTask, text string) (string, error)
	
	// BuildPromptWithExamples constructs a prompt with specific examples
	BuildPromptWithExamples(ctx context.Context, task *ExtractionTask, text string, examples []*extraction.ExampleData) (string, error)
	
	// Name returns the name of the prompt builder
	Name() string
	
	// Validate validates the prompt builder configuration
	Validate() error
}

// TemplateRenderer defines the interface for rendering prompt templates
type TemplateRenderer interface {
	// Render renders a template with the given context
	Render(ctx context.Context, template *PromptTemplate, data interface{}) (string, error)
	
	// RegisterFunction registers a custom function for use in templates
	RegisterFunction(name string, fn interface{}) error
	
	// GetFunctions returns all registered functions
	GetFunctions() map[string]interface{}
}

// ExampleSelector defines the interface for selecting optimal examples for few-shot learning
type ExampleSelector interface {
	// SelectExamples selects the best examples for the given task and context
	SelectExamples(ctx context.Context, task *ExtractionTask, allExamples []*extraction.ExampleData, maxExamples int) ([]*extraction.ExampleData, error)
	
	// ScoreExample scores an example's quality for the given task
	ScoreExample(ctx context.Context, task *ExtractionTask, example *extraction.ExampleData) (float64, error)
	
	// Name returns the name of the example selector
	Name() string
}

// ExtractionTask represents a task for extracting structured information
type ExtractionTask struct {
	// Description is the human-readable task description
	Description string `json:"description"`
	
	// Schema defines the expected extraction schema (optional)
	Schema extraction.ExtractionSchema `json:"schema,omitempty"`
	
	// Classes are the extraction classes to look for
	Classes []string `json:"classes,omitempty"`
	
	// Instructions are additional instructions for the model
	Instructions []string `json:"instructions,omitempty"`
	
	// OutputFormat specifies the desired output format (json, yaml, etc.)
	OutputFormat string `json:"output_format,omitempty"`
	
	// MaxExtractions limits the number of extractions (0 = unlimited)
	MaxExtractions int `json:"max_extractions,omitempty"`
	
	// RequireGrounding specifies if source grounding is required
	RequireGrounding bool `json:"require_grounding"`
	
	// Metadata contains additional task-specific data
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// PromptTemplate represents a prompt template with structured components
type PromptTemplate struct {
	// Name is the template identifier
	Name string `json:"name"`
	
	// Description is the main instruction for the model
	Description string `json:"description"`
	
	// SystemMessage is the system-level instruction (for models that support it)
	SystemMessage string `json:"system_message,omitempty"`
	
	// Template is the template string with placeholders
	Template string `json:"template"`
	
	// Examples are few-shot learning examples
	Examples []*extraction.ExampleData `json:"examples,omitempty"`
	
	// Variables are template variables and their descriptions
	Variables map[string]string `json:"variables,omitempty"`
	
	// OutputFormat specifies the expected output format
	OutputFormat string `json:"output_format,omitempty"`
	
	// Version is the template version for compatibility
	Version string `json:"version,omitempty"`
	
	// Tags are template tags for categorization
	Tags []string `json:"tags,omitempty"`
	
	// Metadata contains additional template data
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// PromptContext contains context data for prompt rendering
type PromptContext struct {
	// Task is the extraction task
	Task *ExtractionTask `json:"task"`
	
	// Text is the input text to process
	Text string `json:"text"`
	
	// Examples are the selected examples for few-shot learning
	Examples []*extraction.ExampleData `json:"examples,omitempty"`
	
	// Variables are template variables
	Variables map[string]interface{} `json:"variables,omitempty"`
	
	// TextLength is the length of the input text in characters
	TextLength int `json:"text_length"`
	
	// ExampleCount is the number of examples provided
	ExampleCount int `json:"example_count"`
	
	// Timestamp is when the prompt was generated
	Timestamp int64 `json:"timestamp"`
}

// PromptOptions contains configuration options for prompt builders
type PromptOptions struct {
	// MaxExamples limits the number of examples to include
	MaxExamples int `json:"max_examples,omitempty"`
	
	// ExampleSelectionStrategy specifies how to select examples
	ExampleSelectionStrategy ExampleSelectionStrategy `json:"example_selection_strategy,omitempty"`
	
	// IncludeInstructions specifies whether to include detailed instructions
	IncludeInstructions bool `json:"include_instructions"`
	
	// UseSchemaConstraints specifies whether to use schema-based constraints
	UseSchemaConstraints bool `json:"use_schema_constraints"`
	
	// OutputFormat specifies the desired output format
	OutputFormat string `json:"output_format,omitempty"`
	
	// Temperature for example selection randomness
	Temperature float64 `json:"temperature,omitempty"`
	
	// CustomTemplate allows overriding the default template
	CustomTemplate *PromptTemplate `json:"custom_template,omitempty"`
	
	// Variables are additional template variables
	Variables map[string]interface{} `json:"variables,omitempty"`
	
	// Debug enables debug mode for prompt generation
	Debug bool `json:"debug,omitempty"`
}

// ExampleSelectionStrategy defines strategies for selecting examples
type ExampleSelectionStrategy int

const (
	// SelectionStrategyRandom selects examples randomly
	SelectionStrategyRandom ExampleSelectionStrategy = iota
	
	// SelectionStrategyDiverse selects diverse examples
	SelectionStrategyDiverse
	
	// SelectionStrategySimilar selects examples similar to the task
	SelectionStrategySimilar
	
	// SelectionStrategyQuality selects highest quality examples
	SelectionStrategyQuality
	
	// SelectionStrategyBest selects best examples based on multiple criteria
	SelectionStrategyBest
)

// String returns the string representation of the selection strategy
func (s ExampleSelectionStrategy) String() string {
	switch s {
	case SelectionStrategyRandom:
		return "random"
	case SelectionStrategyDiverse:
		return "diverse"
	case SelectionStrategySimilar:
		return "similar"
	case SelectionStrategyQuality:
		return "quality"
	case SelectionStrategyBest:
		return "best"
	default:
		return "unknown"
	}
}

// DefaultPromptOptions returns default prompt options
func DefaultPromptOptions() *PromptOptions {
	return &PromptOptions{
		MaxExamples:              5,
		ExampleSelectionStrategy: SelectionStrategyBest,
		IncludeInstructions:      true,
		UseSchemaConstraints:     true,
		OutputFormat:             "json",
		Temperature:              0.0,
		Variables:                make(map[string]interface{}),
		Debug:                    false,
	}
}

// WithMaxExamples sets the maximum number of examples
func (opts *PromptOptions) WithMaxExamples(max int) *PromptOptions {
	opts.MaxExamples = max
	return opts
}

// WithExampleSelection sets the example selection strategy
func (opts *PromptOptions) WithExampleSelection(strategy ExampleSelectionStrategy) *PromptOptions {
	opts.ExampleSelectionStrategy = strategy
	return opts
}

// WithOutputFormat sets the output format
func (opts *PromptOptions) WithOutputFormat(format string) *PromptOptions {
	opts.OutputFormat = format
	return opts
}

// WithCustomTemplate sets a custom template
func (opts *PromptOptions) WithCustomTemplate(template *PromptTemplate) *PromptOptions {
	opts.CustomTemplate = template
	return opts
}

// WithVariable sets a template variable
func (opts *PromptOptions) WithVariable(key string, value interface{}) *PromptOptions {
	if opts.Variables == nil {
		opts.Variables = make(map[string]interface{})
	}
	opts.Variables[key] = value
	return opts
}

// WithDebug enables debug mode
func (opts *PromptOptions) WithDebug(debug bool) *PromptOptions {
	opts.Debug = debug
	return opts
}

// Validate validates the prompt options
func (opts *PromptOptions) Validate() error {
	if opts.MaxExamples < 0 {
		return fmt.Errorf("max_examples: must be non-negative")
	}
	
	if opts.Temperature < 0.0 || opts.Temperature > 2.0 {
		return fmt.Errorf("temperature: must be between 0.0 and 2.0")
	}
	
	if opts.OutputFormat == "" {
		opts.OutputFormat = "json"
	}
	
	return nil
}