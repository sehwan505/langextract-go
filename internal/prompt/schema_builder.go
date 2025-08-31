package prompt

import (
	"context"
	"fmt"
	"strings"

	"github.com/sehwan505/langextract-go/pkg/extraction"
)

// SchemaPromptBuilder implements JSON schema-guided prompt construction
type SchemaPromptBuilder struct {
	options          *PromptOptions
	templateRenderer TemplateRenderer
	exampleSelector  ExampleSelector
}

// NewSchemaPromptBuilder creates a new schema-based prompt builder
func NewSchemaPromptBuilder(opts *PromptOptions) *SchemaPromptBuilder {
	if opts == nil {
		opts = DefaultPromptOptions()
	}

	// Enable schema constraints by default for schema builder
	opts.UseSchemaConstraints = true

	return &SchemaPromptBuilder{
		options:          opts,
		templateRenderer: NewDefaultTemplateRenderer(),
		exampleSelector:  NewSchemaAwareExampleSelector(),
	}
}

// Name returns the name of the prompt builder
func (b *SchemaPromptBuilder) Name() string {
	return "SchemaPromptBuilder"
}

// Validate validates the prompt builder configuration
func (b *SchemaPromptBuilder) Validate() error {
	if b.options == nil {
		return ErrPromptValidationFailed("options cannot be nil")
	}

	if err := b.options.Validate(); err != nil {
		return ErrPromptValidationFailed(fmt.Sprintf("invalid options: %v", err))
	}

	if b.templateRenderer == nil {
		return ErrPromptValidationFailed("template renderer cannot be nil")
	}

	if b.exampleSelector == nil {
		return ErrPromptValidationFailed("example selector cannot be nil")
	}

	return nil
}

// BuildPrompt constructs a prompt for the given task and text
func (b *SchemaPromptBuilder) BuildPrompt(ctx context.Context, task *ExtractionTask, text string) (string, error) {
	if task == nil {
		return "", ErrPromptBuildFailed("task cannot be nil", nil)
	}

	if text == "" {
		return "", ErrPromptBuildFailed("text cannot be empty", nil)
	}

	// Use empty examples if none provided
	return b.BuildPromptWithExamples(ctx, task, text, nil)
}

// BuildPromptWithExamples constructs a prompt with specific examples
func (b *SchemaPromptBuilder) BuildPromptWithExamples(ctx context.Context, task *ExtractionTask, text string, examples []*extraction.ExampleData) (string, error) {
	if task == nil {
		return "", ErrPromptBuildFailed("task cannot be nil", nil)
	}

	if text == "" {
		return "", ErrPromptBuildFailed("text cannot be empty", nil)
	}

	// Validate that task has schema for schema-based building
	if task.Schema == nil {
		return "", ErrPromptBuildFailed("task must have schema for schema-based prompt building", nil)
	}

	// Select schema-aware examples
	selectedExamples := examples
	if len(examples) > b.options.MaxExamples {
		var err error
		selectedExamples, err = b.exampleSelector.SelectExamples(ctx, task, examples, b.options.MaxExamples)
		if err != nil {
			return "", ErrPromptBuildFailed("schema-aware example selection failed", err)
		}
	}

	// Create schema-specific prompt template
	template := b.createSchemaPromptTemplate(task, selectedExamples)

	// Create enhanced prompt context with schema information
	promptContext := &SchemaPromptContext{
		PromptContext: &PromptContext{
			Task:         task,
			Text:         text,
			Examples:     selectedExamples,
			Variables:    b.options.Variables,
			TextLength:   len(text),
			ExampleCount: len(selectedExamples),
			Timestamp:    getCurrentTimestamp(),
		},
		Schema:           task.Schema,
		SchemaString:     b.buildSchemaString(task.Schema),
		RequiredClasses:  b.extractRequiredClasses(task.Schema),
		OptionalClasses:  b.extractOptionalClasses(task.Schema),
		ClassDescriptions: b.buildClassDescriptions(task.Schema),
	}

	// Render the prompt
	prompt, err := b.templateRenderer.Render(ctx, template, promptContext)
	if err != nil {
		return "", ErrPromptBuildFailed("schema template rendering failed", err)
	}

	return prompt, nil
}

// SchemaPromptContext extends PromptContext with schema-specific information
type SchemaPromptContext struct {
	*PromptContext
	Schema            extraction.ExtractionSchema `json:"schema"`
	SchemaString      string                      `json:"schema_string"`
	RequiredClasses   []string                    `json:"required_classes"`
	OptionalClasses   []string                    `json:"optional_classes"`
	ClassDescriptions map[string]string           `json:"class_descriptions"`
}

// createSchemaPromptTemplate creates a schema-specific prompt template
func (b *SchemaPromptBuilder) createSchemaPromptTemplate(task *ExtractionTask, examples []*extraction.ExampleData) *PromptTemplate {
	if b.options.CustomTemplate != nil {
		template := *b.options.CustomTemplate
		template.Examples = examples
		return &template
	}

	return &PromptTemplate{
		Name:        "schema_extraction",
		Description: b.buildSchemaTaskDescription(task),
		Template:    b.buildSchemaTemplate(task),
		Examples:    examples,
		Variables: map[string]string{
			"task_description":    "The extraction task description",
			"text":               "The input text to process",
			"examples":           "Schema-compliant examples",
			"schema":             "The JSON schema for extractions",
			"schema_string":      "The schema as a formatted string",
			"required_classes":   "Required extraction classes",
			"optional_classes":   "Optional extraction classes",
			"class_descriptions": "Descriptions for each extraction class",
		},
		OutputFormat: b.options.OutputFormat,
		Version:      "1.0",
		Tags:         []string{"schema-guided", "extraction", "structured"},
	}
}

// buildSchemaTaskDescription creates a schema-aware task description
func (b *SchemaPromptBuilder) buildSchemaTaskDescription(task *ExtractionTask) string {
	var parts []string

	// Main description
	if task.Description != "" {
		parts = append(parts, task.Description)
	}

	// Schema-specific instructions
	parts = append(parts, "Extract structured information according to the provided JSON schema.")
	parts = append(parts, "Follow the schema constraints exactly and include all required fields.")

	// Required vs optional classes
	if len(task.Classes) > 0 {
		classesStr := strings.Join(task.Classes, ", ")
		parts = append(parts, fmt.Sprintf("Focus on these extraction classes: %s", classesStr))
	}

	// Additional schema-specific instructions
	parts = append(parts, b.getSchemaInstructions(task))

	// Custom instructions
	if len(task.Instructions) > 0 {
		parts = append(parts, "Additional instructions:")
		for _, instruction := range task.Instructions {
			parts = append(parts, fmt.Sprintf("- %s", instruction))
		}
	}

	// Source grounding requirement
	if task.RequireGrounding {
		parts = append(parts, "IMPORTANT: Use exact text from the source for all extractions. Maintain precise character positions when possible.")
	}

	return strings.Join(parts, "\n\n")
}

// getSchemaInstructions returns schema-specific instructions
func (b *SchemaPromptBuilder) getSchemaInstructions(task *ExtractionTask) string {
	instructions := []string{
		"Ensure all extractions conform to the provided schema.",
		"Include all required fields for each extraction class.",
		"Use the exact field names and types specified in the schema.",
		"Validate that your output can be parsed according to the schema.",
	}

	if task.MaxExtractions > 0 {
		instructions = append(instructions, fmt.Sprintf("Extract at most %d items total.", task.MaxExtractions))
	}

	if b.options.OutputFormat == "json" {
		instructions = append(instructions, "Return valid JSON that matches the schema structure.")
	}

	return strings.Join(instructions, " ")
}

// buildSchemaTemplate creates the schema-specific prompt template
func (b *SchemaPromptBuilder) buildSchemaTemplate(task *ExtractionTask) string {
	var templateParts []string

	// System message
	templateParts = append(templateParts, "You are a precise information extraction system that follows JSON schemas exactly.")

	// Task description
	templateParts = append(templateParts, "{{.Task.Description}}")

	// Schema section
	templateParts = append(templateParts, b.buildSchemaSection())

	// Examples section
	templateParts = append(templateParts, b.buildSchemaExamplesSection())

	// Input text section
	templateParts = append(templateParts, "Now extract information from the following text according to the schema:")
	templateParts = append(templateParts, "Text: {{.Text}}")

	// Output format with schema reminder
	templateParts = append(templateParts, b.buildSchemaOutputSection())

	return strings.Join(templateParts, "\n\n")
}

// buildSchemaSection creates the schema documentation section
func (b *SchemaPromptBuilder) buildSchemaSection() string {
	return `JSON Schema for Extractions:
{{.SchemaString}}

{{if .RequiredClasses}}Required extraction classes: {{join ", " .RequiredClasses}}{{end}}
{{if .OptionalClasses}}Optional extraction classes: {{join ", " .OptionalClasses}}{{end}}

{{if .ClassDescriptions}}Class Descriptions:
{{range $class, $desc := .ClassDescriptions}}- {{$class}}: {{$desc}}
{{end}}{{end}}`
}

// buildSchemaExamplesSection creates the examples section for schema prompts
func (b *SchemaPromptBuilder) buildSchemaExamplesSection() string {
	return `{{if .Examples}}Schema-compliant examples:

{{range $i, $example := .Examples}}Example {{add $i 1}}:
Input: {{$example.Input}}
Output: {{formatExample $example .OutputFormat}}
{{if $example.Metadata.SchemaValidation}}✓ Schema compliant{{end}}

{{end}}{{end}}`
}

// buildSchemaOutputSection creates the output section with schema validation reminder
func (b *SchemaPromptBuilder) buildSchemaOutputSection() string {
	format := strings.ToUpper(b.options.OutputFormat)
	return fmt.Sprintf(`Output (valid %s following the schema):`, format) + `
Remember to:
- Follow the schema structure exactly
- Include all required fields
- Use correct data types
- Validate your output against the schema`
}

// buildSchemaString creates a formatted string representation of the schema
func (b *SchemaPromptBuilder) buildSchemaString(schema extraction.ExtractionSchema) string {
	if schema == nil {
		return "{}"
	}

	// This would need to be implemented based on the actual schema interface
	// For now, return a placeholder
	return `{
  "type": "object",
  "properties": {
    "extractions": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "text": {"type": "string"},
          "class": {"type": "string"},
          "confidence": {"type": "number"},
          "interval": {
            "type": "object",
            "properties": {
              "start": {"type": "integer"},
              "end": {"type": "integer"}
            }
          }
        },
        "required": ["text", "class"]
      }
    }
  }
}`
}

// extractRequiredClasses extracts required classes from the schema
func (b *SchemaPromptBuilder) extractRequiredClasses(schema extraction.ExtractionSchema) []string {
	if schema == nil {
		return nil
	}

	// This would be implemented based on the actual schema interface
	// For now, return empty slice
	return []string{}
}

// extractOptionalClasses extracts optional classes from the schema
func (b *SchemaPromptBuilder) extractOptionalClasses(schema extraction.ExtractionSchema) []string {
	if schema == nil {
		return nil
	}

	// This would be implemented based on the actual schema interface
	// For now, return empty slice
	return []string{}
}

// buildClassDescriptions creates descriptions for each extraction class
func (b *SchemaPromptBuilder) buildClassDescriptions(schema extraction.ExtractionSchema) map[string]string {
	if schema == nil {
		return make(map[string]string)
	}

	// This would be implemented based on the actual schema interface
	// For now, return empty map
	return make(map[string]string)
}

// SetTemplateRenderer sets a custom template renderer
func (b *SchemaPromptBuilder) SetTemplateRenderer(renderer TemplateRenderer) {
	b.templateRenderer = renderer
}

// SetExampleSelector sets a custom example selector
func (b *SchemaPromptBuilder) SetExampleSelector(selector ExampleSelector) {
	b.exampleSelector = selector
}

// SchemaAwareExampleSelector implements schema-aware example selection
type SchemaAwareExampleSelector struct {
	qualitySelector *QualityExampleSelector
}

// NewSchemaAwareExampleSelector creates a new schema-aware example selector
func NewSchemaAwareExampleSelector() *SchemaAwareExampleSelector {
	return &SchemaAwareExampleSelector{
		qualitySelector: NewQualityExampleSelector(),
	}
}

// Name returns the name of the example selector
func (s *SchemaAwareExampleSelector) Name() string {
	return "SchemaAwareExampleSelector"
}

// SelectExamples selects examples that best demonstrate schema compliance
func (s *SchemaAwareExampleSelector) SelectExamples(ctx context.Context, task *ExtractionTask, allExamples []*extraction.ExampleData, maxExamples int) ([]*extraction.ExampleData, error) {
	if len(allExamples) <= maxExamples {
		return allExamples, nil
	}

	// First filter for schema-compliant examples
	schemaCompliantExamples := s.filterSchemaCompliantExamples(task.Schema, allExamples)

	// If we have enough schema-compliant examples, use quality selection on them
	if len(schemaCompliantExamples) >= maxExamples {
		return s.qualitySelector.SelectExamples(ctx, task, schemaCompliantExamples, maxExamples)
	}

	// If not enough schema-compliant examples, mix with high-quality ones
	remaining := maxExamples - len(schemaCompliantExamples)
	if remaining > 0 {
		// Get remaining examples from quality selection
		qualityExamples, err := s.qualitySelector.SelectExamples(ctx, task, allExamples, remaining)
		if err != nil {
			return schemaCompliantExamples, nil
		}

		// Combine and deduplicate
		combined := append(schemaCompliantExamples, qualityExamples...)
		return s.deduplicateExamples(combined), nil
	}

	return schemaCompliantExamples, nil
}

// ScoreExample scores an example's quality with emphasis on schema compliance
func (s *SchemaAwareExampleSelector) ScoreExample(ctx context.Context, task *ExtractionTask, example *extraction.ExampleData) (float64, error) {
	// Get base quality score
	baseScore, err := s.qualitySelector.ScoreExample(ctx, task, example)
	if err != nil {
		return 0.0, err
	}

	// Add schema compliance bonus
	schemaScore := s.scoreSchemaCompliance(task.Schema, example)

	// Weight: 60% base quality, 40% schema compliance
	return baseScore*0.6 + schemaScore*0.4, nil
}

// filterSchemaCompliantExamples filters examples that are schema-compliant
func (s *SchemaAwareExampleSelector) filterSchemaCompliantExamples(schema extraction.ExtractionSchema, examples []*extraction.ExampleData) []*extraction.ExampleData {
	var compliantExamples []*extraction.ExampleData

	for _, example := range examples {
		if s.isSchemaCompliant(schema, example) {
			compliantExamples = append(compliantExamples, example)
		}
	}

	return compliantExamples
}

// isSchemaCompliant checks if an example is compliant with the schema
func (s *SchemaAwareExampleSelector) isSchemaCompliant(schema extraction.ExtractionSchema, example *extraction.ExampleData) bool {
	if schema == nil || example == nil {
		return false
	}

	// Check if example metadata indicates schema compliance
	if example.Metadata != nil {
		if compliance, exists := example.Metadata["schema_compliant"]; exists {
			if compliant, ok := compliance.(bool); ok {
				return compliant
			}
		}
	}

	// Basic validation: check if all extractions have required fields
	for _, extraction := range example.Extractions {
		if extraction.Text() == "" || extraction.Class() == "" {
			return false
		}
	}

	// Additional schema validation would be implemented here
	// based on the actual schema interface

	return true
}

// scoreSchemaCompliance scores how well an example demonstrates schema compliance
func (s *SchemaAwareExampleSelector) scoreSchemaCompliance(schema extraction.ExtractionSchema, example *extraction.ExampleData) float64 {
	if schema == nil || example == nil {
		return 0.0
	}

	var score float64

	// Check metadata for schema compliance indicator
	if example.Metadata != nil {
		if compliance, exists := example.Metadata["schema_compliant"]; exists {
			if compliant, ok := compliance.(bool); ok && compliant {
				score += 0.5
			}
		}
	}

	// Score based on extraction completeness
	if len(example.Extractions) > 0 {
		completeExtractions := 0
		for _, extraction := range example.Extractions {
			if extraction.Text() != "" && extraction.Class() != "" {
				completeExtractions++
				// Bonus for having confidence scores
				if extraction.Confidence() > 0 {
					score += 0.1
				}
				// Bonus for having intervals (source grounding)
				if extraction.Interval() != nil {
					score += 0.1
				}
			}
		}

		completenessRatio := float64(completeExtractions) / float64(len(example.Extractions))
		score += completenessRatio * 0.3
	}

	return score
}

// deduplicateExamples removes duplicate examples
func (s *SchemaAwareExampleSelector) deduplicateExamples(examples []*extraction.ExampleData) []*extraction.ExampleData {
	seen := make(map[string]bool)
	var unique []*extraction.ExampleData

	for _, example := range examples {
		// Create a simple hash of the example
		hash := fmt.Sprintf("%s_%d", example.Input(), len(example.Extractions))
		if !seen[hash] {
			seen[hash] = true
			unique = append(unique, example)
		}
	}

	return unique
}