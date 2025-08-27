# prompt

Package: `github.com/sehwan505/langextract-go/internal/prompt`

```go
package prompt // import "github.com/sehwan505/langextract-go/internal/prompt"


FUNCTIONS

func ErrExampleSelectionFailed(reason string, cause error) error
    ErrExampleSelectionFailed creates an example selection error

func ErrInvalidExample(exampleID, reason string) error
    ErrInvalidExample creates an example validation error

func ErrInvalidTemplate(templateName, reason string) error
    ErrInvalidTemplate creates a template validation error

func ErrPromptBuildFailed(reason string, cause error) error
    ErrPromptBuildFailed creates a prompt building error

func ErrPromptValidationFailed(reason string) error
    ErrPromptValidationFailed creates a prompt validation error

func ErrTemplateNotFound(templateName string) error
    ErrTemplateNotFound creates a template not found error

func ErrTemplateRenderFailed(templateName, reason string, cause error) error
    ErrTemplateRenderFailed creates a template rendering error


TYPES

type BasicExampleValidator struct{}
    BasicExampleValidator implements basic example validation

func NewBasicExampleValidator() *BasicExampleValidator
    NewBasicExampleValidator creates a new basic example validator

func (v *BasicExampleValidator) Name() string
    Name returns the name of the validator

func (v *BasicExampleValidator) ValidateExample(ctx context.Context, example *extraction.ExampleData) error
    ValidateExample validates basic example requirements

type ComplexityExampleScorer struct{}
    ComplexityExampleScorer implements complexity-based example scoring

func NewComplexityExampleScorer() *ComplexityExampleScorer
    NewComplexityExampleScorer creates a new complexity example scorer

func (s *ComplexityExampleScorer) Name() string
    Name returns the name of the scorer

func (s *ComplexityExampleScorer) ScoreExample(ctx context.Context, example *extraction.ExampleData) (float64, error)
    ScoreExample scores an example's complexity

type ComplexityStats struct {
	Average float64 `json:"average"`
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
	StdDev  float64 `json:"std_dev"`
}
    ComplexityStats contains complexity-related statistics

type DefaultTemplateRenderer struct {
	// Has unexported fields.
}
    DefaultTemplateRenderer implements template rendering using Go's
    text/template

func NewDefaultTemplateRenderer() *DefaultTemplateRenderer
    NewDefaultTemplateRenderer creates a new default template renderer

func (r *DefaultTemplateRenderer) ClearFunctions()
    ClearFunctions removes all custom functions (keeps default ones)

func (r *DefaultTemplateRenderer) GetFunctions() map[string]interface{}
    GetFunctions returns all registered functions

func (r *DefaultTemplateRenderer) GetTemplateVariables(templateContent string) ([]string, error)
    GetTemplateVariables extracts variable names from a template

func (r *DefaultTemplateRenderer) HasFunction(name string) bool
    HasFunction checks if a function is registered

func (r *DefaultTemplateRenderer) RegisterFunction(name string, fn interface{}) error
    RegisterFunction registers a custom function for use in templates

func (r *DefaultTemplateRenderer) Render(ctx context.Context, template *PromptTemplate, data interface{}) (string, error)
    Render renders a template with the given context

func (r *DefaultTemplateRenderer) UnregisterFunction(name string)
    UnregisterFunction removes a function from the renderer

func (r *DefaultTemplateRenderer) ValidateTemplate(promptTemplate *PromptTemplate) error
    ValidateTemplate validates a template without rendering it

type ExampleCollection struct {
	Name        string                    `json:"name"`
	Description string                    `json:"description"`
	Domain      string                    `json:"domain"`
	Version     string                    `json:"version"`
	Examples    []*extraction.ExampleData `json:"examples"`
	Metadata    map[string]interface{}    `json:"metadata"`
	Tags        []string                  `json:"tags"`
	CreatedAt   int64                     `json:"created_at"`
	UpdatedAt   int64                     `json:"updated_at"`
	Statistics  *ExampleCollectionStats   `json:"statistics"`
}
    ExampleCollection represents a collection of examples for a specific domain
    or task

type ExampleCollectionStats struct {
	TotalExamples    int              `json:"total_examples"`
	ClassCounts      map[string]int   `json:"class_counts"`
	AverageLength    float64          `json:"average_length"`
	QualityScores    *QualityStats    `json:"quality_scores"`
	ComplexityScores *ComplexityStats `json:"complexity_scores"`
	ValidationStatus *ValidationStats `json:"validation_status"`
}
    ExampleCollectionStats contains statistics about an example collection

type ExampleError struct {
	ExampleID string
	Operation string
	Message   string
	Cause     error
}
    ExampleError represents an error in example processing

func NewExampleError(operation, message string, cause error) *ExampleError
    NewExampleError creates a new example error

func NewExampleErrorWithID(exampleID, operation, message string, cause error) *ExampleError
    NewExampleErrorWithID creates a new example error with ID

func (e *ExampleError) Error() string
    Error implements the error interface

func (e *ExampleError) Unwrap() error
    Unwrap returns the underlying cause

type ExampleLoadOptions struct {
	ValidateOnLoad  bool     `json:"validate_on_load"`
	ComputeScores   bool     `json:"compute_scores"`
	FilterTags      []string `json:"filter_tags"`
	FilterClasses   []string `json:"filter_classes"`
	MinQualityScore float64  `json:"min_quality_score"`
	MaxExamples     int      `json:"max_examples"`
	ShuffleExamples bool     `json:"shuffle_examples"`
}
    ExampleLoadOptions contains options for loading examples

type ExampleManager struct {
	// Has unexported fields.
}
    ExampleManager manages loading, validation, and organization of extraction
    examples

func NewExampleManager() *ExampleManager
    NewExampleManager creates a new example manager

func (m *ExampleManager) GetExampleCollection(name string) (*ExampleCollection, error)
    GetExampleCollection retrieves an example collection by name

func (m *ExampleManager) ListExampleCollections() []string
    ListExampleCollections returns all available example collection names

func (m *ExampleManager) LoadExamplesFromDirectory(ctx context.Context, dir string, opts *ExampleLoadOptions) ([]*ExampleCollection, error)
    LoadExamplesFromDirectory loads all example files from a directory

func (m *ExampleManager) LoadExamplesFromFile(ctx context.Context, filePath string, opts *ExampleLoadOptions) (*ExampleCollection, error)
    LoadExamplesFromFile loads examples from a JSON file

func (m *ExampleManager) LoadExamplesFromReader(ctx context.Context, reader io.Reader, name string, opts *ExampleLoadOptions) (*ExampleCollection, error)
    LoadExamplesFromReader loads examples from an io.Reader

func (m *ExampleManager) RegisterScorer(scorer ExampleScorer)
    RegisterScorer registers an example scorer

func (m *ExampleManager) RegisterValidator(validator ExampleValidator)
    RegisterValidator registers an example validator

func (m *ExampleManager) ScoreExample(ctx context.Context, example *extraction.ExampleData) (map[string]float64, error)
    ScoreExample scores a single example using all registered scorers

func (m *ExampleManager) ValidateExample(ctx context.Context, example *extraction.ExampleData) error
    ValidateExample validates a single example using all registered validators

func (m *ExampleManager) ValidateExampleCollection(ctx context.Context, collection *ExampleCollection) []error
    ValidateExampleCollection validates all examples in a collection

type ExampleScorer interface {
	ScoreExample(ctx context.Context, example *extraction.ExampleData) (float64, error)
	Name() string
}
    ExampleScorer defines the interface for scoring examples

type ExampleSelectionStrategy int
    ExampleSelectionStrategy defines strategies for selecting examples

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
func (s ExampleSelectionStrategy) String() string
    String returns the string representation of the selection strategy

type ExampleSelector interface {
	// SelectExamples selects the best examples for the given task and context
	SelectExamples(ctx context.Context, task *ExtractionTask, allExamples []*extraction.ExampleData, maxExamples int) ([]*extraction.ExampleData, error)

	// ScoreExample scores an example's quality for the given task
	ScoreExample(ctx context.Context, task *ExtractionTask, example *extraction.ExampleData) (float64, error)

	// Name returns the name of the example selector
	Name() string
}
    ExampleSelector defines the interface for selecting optimal examples for
    few-shot learning

type ExampleValidator interface {
	ValidateExample(ctx context.Context, example *extraction.ExampleData) error
	Name() string
}
    ExampleValidator defines the interface for validating examples

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
    ExtractionTask represents a task for extracting structured information

type FewShotPromptBuilder struct {
	// Has unexported fields.
}
    FewShotPromptBuilder implements example-driven prompt construction

func NewFewShotPromptBuilder(opts *PromptOptions) *FewShotPromptBuilder
    NewFewShotPromptBuilder creates a new few-shot prompt builder

func (b *FewShotPromptBuilder) BuildPrompt(ctx context.Context, task *ExtractionTask, text string) (string, error)
    BuildPrompt constructs a prompt for the given task and text

func (b *FewShotPromptBuilder) BuildPromptWithExamples(ctx context.Context, task *ExtractionTask, text string, examples []*extraction.ExampleData) (string, error)
    BuildPromptWithExamples constructs a prompt with specific examples

func (b *FewShotPromptBuilder) GetOptions() *PromptOptions
    GetOptions returns the current options

func (b *FewShotPromptBuilder) Name() string
    Name returns the name of the prompt builder

func (b *FewShotPromptBuilder) SetExampleSelector(selector ExampleSelector)
    SetExampleSelector sets a custom example selector

func (b *FewShotPromptBuilder) SetTemplateRenderer(renderer TemplateRenderer)
    SetTemplateRenderer sets a custom template renderer

func (b *FewShotPromptBuilder) UpdateOptions(opts *PromptOptions) error
    UpdateOptions updates the prompt options

func (b *FewShotPromptBuilder) Validate() error
    Validate validates the prompt builder configuration

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
    PromptBuilder defines the interface for building prompts for language models

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
    PromptContext contains context data for prompt rendering

type PromptError struct {
	Operation string
	Message   string
	Cause     error
}
    PromptError represents an error in prompt processing

func NewPromptError(operation, message string, cause error) *PromptError
    NewPromptError creates a new prompt error

func (e *PromptError) Error() string
    Error implements the error interface

func (e *PromptError) Unwrap() error
    Unwrap returns the underlying cause

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
    PromptOptions contains configuration options for prompt builders

func DefaultPromptOptions() *PromptOptions
    DefaultPromptOptions returns default prompt options

func (opts *PromptOptions) Validate() error
    Validate validates the prompt options

func (opts *PromptOptions) WithCustomTemplate(template *PromptTemplate) *PromptOptions
    WithCustomTemplate sets a custom template

func (opts *PromptOptions) WithDebug(debug bool) *PromptOptions
    WithDebug enables debug mode

func (opts *PromptOptions) WithExampleSelection(strategy ExampleSelectionStrategy) *PromptOptions
    WithExampleSelection sets the example selection strategy

func (opts *PromptOptions) WithMaxExamples(max int) *PromptOptions
    WithMaxExamples sets the maximum number of examples

func (opts *PromptOptions) WithOutputFormat(format string) *PromptOptions
    WithOutputFormat sets the output format

func (opts *PromptOptions) WithVariable(key string, value interface{}) *PromptOptions
    WithVariable sets a template variable

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
    PromptTemplate represents a prompt template with structured components

type QualityExampleScorer struct {
	// Has unexported fields.
}
    QualityExampleScorer implements quality-based example scoring

func NewQualityExampleScorer() *QualityExampleScorer
    NewQualityExampleScorer creates a new quality example scorer

func (s *QualityExampleScorer) Name() string
    Name returns the name of the scorer

func (s *QualityExampleScorer) ScoreExample(ctx context.Context, example *extraction.ExampleData) (float64, error)
    ScoreExample scores an example's quality

type QualityExampleSelector struct{}
    QualityExampleSelector implements example selection based on quality scores

func NewQualityExampleSelector() *QualityExampleSelector
    NewQualityExampleSelector creates a new quality-based example selector

func (s *QualityExampleSelector) Name() string
    Name returns the name of the example selector

func (s *QualityExampleSelector) ScoreExample(ctx context.Context, task *ExtractionTask, example *extraction.ExampleData) (float64, error)
    ScoreExample scores an example's quality for the given task

func (s *QualityExampleSelector) SelectExamples(ctx context.Context, task *ExtractionTask, allExamples []*extraction.ExampleData, maxExamples int) ([]*extraction.ExampleData, error)
    SelectExamples selects the best examples for the given task and context

type QualityStats struct {
	Average float64 `json:"average"`
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
	StdDev  float64 `json:"std_dev"`
}
    QualityStats contains quality-related statistics

type SchemaAwareExampleSelector struct {
	// Has unexported fields.
}
    SchemaAwareExampleSelector implements schema-aware example selection

func NewSchemaAwareExampleSelector() *SchemaAwareExampleSelector
    NewSchemaAwareExampleSelector creates a new schema-aware example selector

func (s *SchemaAwareExampleSelector) Name() string
    Name returns the name of the example selector

func (s *SchemaAwareExampleSelector) ScoreExample(ctx context.Context, task *ExtractionTask, example *extraction.ExampleData) (float64, error)
    ScoreExample scores an example's quality with emphasis on schema compliance

func (s *SchemaAwareExampleSelector) SelectExamples(ctx context.Context, task *ExtractionTask, allExamples []*extraction.ExampleData, maxExamples int) ([]*extraction.ExampleData, error)
    SelectExamples selects examples that best demonstrate schema compliance

type SchemaPromptBuilder struct {
	// Has unexported fields.
}
    SchemaPromptBuilder implements JSON schema-guided prompt construction

func NewSchemaPromptBuilder(opts *PromptOptions) *SchemaPromptBuilder
    NewSchemaPromptBuilder creates a new schema-based prompt builder

func (b *SchemaPromptBuilder) BuildPrompt(ctx context.Context, task *ExtractionTask, text string) (string, error)
    BuildPrompt constructs a prompt for the given task and text

func (b *SchemaPromptBuilder) BuildPromptWithExamples(ctx context.Context, task *ExtractionTask, text string, examples []*extraction.ExampleData) (string, error)
    BuildPromptWithExamples constructs a prompt with specific examples

func (b *SchemaPromptBuilder) Name() string
    Name returns the name of the prompt builder

func (b *SchemaPromptBuilder) SetExampleSelector(selector ExampleSelector)
    SetExampleSelector sets a custom example selector

func (b *SchemaPromptBuilder) SetTemplateRenderer(renderer TemplateRenderer)
    SetTemplateRenderer sets a custom template renderer

func (b *SchemaPromptBuilder) Validate() error
    Validate validates the prompt builder configuration

type SchemaPromptContext struct {
	*PromptContext
	Schema            extraction.ExtractionSchema `json:"schema"`
	SchemaString      string                      `json:"schema_string"`
	RequiredClasses   []string                    `json:"required_classes"`
	OptionalClasses   []string                    `json:"optional_classes"`
	ClassDescriptions map[string]string           `json:"class_descriptions"`
}
    SchemaPromptContext extends PromptContext with schema-specific information

type TemplateError struct {
	TemplateName string
	Line         int
	Column       int
	Message      string
	Cause        error
}
    TemplateError represents an error in template processing

func NewTemplateError(templateName, message string, cause error) *TemplateError
    NewTemplateError creates a new template error

func NewTemplateErrorWithLocation(templateName, message string, line, column int, cause error) *TemplateError
    NewTemplateErrorWithLocation creates a new template error with location

func (e *TemplateError) Error() string
    Error implements the error interface

func (e *TemplateError) Unwrap() error
    Unwrap returns the underlying cause

type TemplateRenderer interface {
	// Render renders a template with the given context
	Render(ctx context.Context, template *PromptTemplate, data interface{}) (string, error)

	// RegisterFunction registers a custom function for use in templates
	RegisterFunction(name string, fn interface{}) error

	// GetFunctions returns all registered functions
	GetFunctions() map[string]interface{}
}
    TemplateRenderer defines the interface for rendering prompt templates

type ValidationStats struct {
	ValidCount   int     `json:"valid_count"`
	InvalidCount int     `json:"invalid_count"`
	ValidRatio   float64 `json:"valid_ratio"`
}
    ValidationStats contains validation-related statistics

```
