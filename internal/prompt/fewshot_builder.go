package prompt

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/sehwan505/langextract-go/pkg/extraction"
)

// FewShotPromptBuilder implements example-driven prompt construction
type FewShotPromptBuilder struct {
	options          *PromptOptions
	templateRenderer TemplateRenderer
	exampleSelector  ExampleSelector
}

// NewFewShotPromptBuilder creates a new few-shot prompt builder
func NewFewShotPromptBuilder(opts *PromptOptions) *FewShotPromptBuilder {
	if opts == nil {
		opts = DefaultPromptOptions()
	}

	return &FewShotPromptBuilder{
		options:          opts,
		templateRenderer: NewDefaultTemplateRenderer(),
		exampleSelector:  NewQualityExampleSelector(),
	}
}

// Name returns the name of the prompt builder
func (b *FewShotPromptBuilder) Name() string {
	return "FewShotPromptBuilder"
}

// Validate validates the prompt builder configuration
func (b *FewShotPromptBuilder) Validate() error {
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
func (b *FewShotPromptBuilder) BuildPrompt(ctx context.Context, task *ExtractionTask, text string) (string, error) {
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
func (b *FewShotPromptBuilder) BuildPromptWithExamples(ctx context.Context, task *ExtractionTask, text string, examples []*extraction.ExampleData) (string, error) {
	if task == nil {
		return "", ErrPromptBuildFailed("task cannot be nil", nil)
	}

	if text == "" {
		return "", ErrPromptBuildFailed("text cannot be empty", nil)
	}

	// Select examples if not provided or need filtering
	selectedExamples := examples
	if len(examples) > b.options.MaxExamples {
		var err error
		selectedExamples, err = b.exampleSelector.SelectExamples(ctx, task, examples, b.options.MaxExamples)
		if err != nil {
			return "", ErrPromptBuildFailed("example selection failed", err)
		}
	}

	// Create prompt template
	template := b.createPromptTemplate(task, selectedExamples)

	// Create prompt context
	promptContext := &PromptContext{
		Task:         task,
		Text:         text,
		Examples:     selectedExamples,
		Variables:    b.options.Variables,
		TextLength:   len(text),
		ExampleCount: len(selectedExamples),
		Timestamp:    getCurrentTimestamp(),
		OutputFormat: b.options.OutputFormat,
	}

	// Render the prompt
	prompt, err := b.templateRenderer.Render(ctx, template, promptContext)
	if err != nil {
		return "", ErrPromptBuildFailed("template rendering failed", err)
	}

	return prompt, nil
}

// createPromptTemplate creates a prompt template for the task
func (b *FewShotPromptBuilder) createPromptTemplate(task *ExtractionTask, examples []*extraction.ExampleData) *PromptTemplate {
	if b.options.CustomTemplate != nil {
		template := *b.options.CustomTemplate
		template.Examples = examples
		return &template
	}

	return &PromptTemplate{
		Name:        "fewshot_extraction",
		Description: b.buildTaskDescription(task),
		Template:    b.buildTemplate(task),
		Examples:    examples,
		Variables: map[string]string{
			"task_description": "The extraction task description",
			"text":             "The input text to process",
			"examples":         "Few-shot learning examples",
			"output_format":    "The expected output format",
		},
		OutputFormat: b.options.OutputFormat,
		Version:      "1.0",
		Tags:         []string{"few-shot", "extraction"},
	}
}

// buildTaskDescription creates a comprehensive task description
func (b *FewShotPromptBuilder) buildTaskDescription(task *ExtractionTask) string {
	var parts []string

	// Main description
	if task.Description != "" {
		parts = append(parts, task.Description)
	}

	// Extraction classes
	if len(task.Classes) > 0 {
		classesStr := strings.Join(task.Classes, ", ")
		parts = append(parts, fmt.Sprintf("Extract the following types of information: %s", classesStr))
	}

	// Additional instructions
	if b.options.IncludeInstructions {
		parts = append(parts, b.getDefaultInstructions(task))
	}

	// Custom instructions
	if len(task.Instructions) > 0 {
		parts = append(parts, "Additional instructions:")
		for _, instruction := range task.Instructions {
			parts = append(parts, fmt.Sprintf("- %s", instruction))
		}
	}

	// Schema constraints
	if b.options.UseSchemaConstraints && task.Schema != nil {
		parts = append(parts, b.buildSchemaInstructions(task.Schema))
	}

	// Source grounding requirement
	if task.RequireGrounding {
		parts = append(parts, "IMPORTANT: Use exact text from the source for all extractions. Do not paraphrase or modify the extracted text.")
	}

	return strings.Join(parts, "\n\n")
}

// getDefaultInstructions returns default extraction instructions
func (b *FewShotPromptBuilder) getDefaultInstructions(task *ExtractionTask) string {
	instructions := []string{
		"Extract structured information from the provided text.",
		"Return results in valid " + strings.ToUpper(b.options.OutputFormat) + " format.",
		"Include only information that is explicitly stated or clearly implied in the text.",
		"If no relevant information is found, return an empty result.",
	}

	if task.MaxExtractions > 0 {
		instructions = append(instructions, fmt.Sprintf("Extract at most %d items.", task.MaxExtractions))
	}

	return strings.Join(instructions, " ")
}

// buildSchemaInstructions creates instructions based on the schema
func (b *FewShotPromptBuilder) buildSchemaInstructions(schema extraction.ExtractionSchema) string {
	if schema == nil {
		return ""
	}

	var instructions []string
	instructions = append(instructions, "Follow this schema for extractions:")

	// Add class descriptions if available
	// Note: This assumes the schema interface has methods to get class information
	// This would need to be implemented based on the actual schema interface

	return strings.Join(instructions, "\n")
}

// buildTemplate creates the prompt template string
func (b *FewShotPromptBuilder) buildTemplate(task *ExtractionTask) string {
	var templateParts []string

	// System message (for models that support it)
	templateParts = append(templateParts, "You are an expert information extraction system.")

	// Task description
	templateParts = append(templateParts, "{{.Task.Description}}")

	// Examples section
	templateParts = append(templateParts, b.buildExamplesSection())

	// Input text section
	templateParts = append(templateParts, "Now extract information from the following text:")
	templateParts = append(templateParts, "Text: {{.Text}}")

	// Output format instruction
	outputInstruction := fmt.Sprintf("Output (in %s format):", strings.ToUpper(b.options.OutputFormat))
	templateParts = append(templateParts, outputInstruction)

	return strings.Join(templateParts, "\n\n")
}

// buildExamplesSection creates the examples section of the template
func (b *FewShotPromptBuilder) buildExamplesSection() string {
	return `{{if .Examples}}Here are some examples:

{{range $i, $example := .Examples}}Example {{add $i 1}}:
Input: {{$example.Input}}
Output: {{formatExample $example $.OutputFormat}}

{{end}}{{end}}`
}

// SetTemplateRenderer sets a custom template renderer
func (b *FewShotPromptBuilder) SetTemplateRenderer(renderer TemplateRenderer) {
	b.templateRenderer = renderer
}

// SetExampleSelector sets a custom example selector
func (b *FewShotPromptBuilder) SetExampleSelector(selector ExampleSelector) {
	b.exampleSelector = selector
}

// GetOptions returns the current options
func (b *FewShotPromptBuilder) GetOptions() *PromptOptions {
	return b.options
}

// UpdateOptions updates the prompt options
func (b *FewShotPromptBuilder) UpdateOptions(opts *PromptOptions) error {
	if opts == nil {
		return ErrPromptValidationFailed("options cannot be nil")
	}

	if err := opts.Validate(); err != nil {
		return ErrPromptValidationFailed(fmt.Sprintf("invalid options: %v", err))
	}

	b.options = opts
	return nil
}

// QualityExampleSelector implements example selection based on quality scores
type QualityExampleSelector struct{}

// NewQualityExampleSelector creates a new quality-based example selector
func NewQualityExampleSelector() *QualityExampleSelector {
	return &QualityExampleSelector{}
}

// Name returns the name of the example selector
func (s *QualityExampleSelector) Name() string {
	return "QualityExampleSelector"
}

// SelectExamples selects the best examples for the given task and context
func (s *QualityExampleSelector) SelectExamples(ctx context.Context, task *ExtractionTask, allExamples []*extraction.ExampleData, maxExamples int) ([]*extraction.ExampleData, error) {
	if len(allExamples) <= maxExamples {
		return allExamples, nil
	}

	// Score all examples
	type scoredExample struct {
		example *extraction.ExampleData
		score   float64
	}

	var scoredExamples []scoredExample
	for _, example := range allExamples {
		score, err := s.ScoreExample(ctx, task, example)
		if err != nil {
			continue // Skip examples that can't be scored
		}
		scoredExamples = append(scoredExamples, scoredExample{
			example: example,
			score:   score,
		})
	}

	// Sort by score (highest first)
	sort.Slice(scoredExamples, func(i, j int) bool {
		return scoredExamples[i].score > scoredExamples[j].score
	})

	// Select top examples
	var selectedExamples []*extraction.ExampleData
	for i := 0; i < maxExamples && i < len(scoredExamples); i++ {
		selectedExamples = append(selectedExamples, scoredExamples[i].example)
	}

	return selectedExamples, nil
}

// ScoreExample scores an example's quality for the given task
func (s *QualityExampleSelector) ScoreExample(ctx context.Context, task *ExtractionTask, example *extraction.ExampleData) (float64, error) {
	if example == nil {
		return 0.0, NewExampleError("scoring", "example cannot be nil", nil)
	}

	var score float64

	// Input quality (length, clarity)
	if example.Text != "" {
		score += s.scoreInputQuality(example.Text)
	}

	// Output quality (number of extractions, completeness)
	if len(example.Extractions) > 0 {
		score += s.scoreOutputQuality(example.Extractions)
	}

	// Relevance to task classes
	if len(task.Classes) > 0 {
		score += s.scoreRelevance(example, task.Classes)
	}

	// Complexity bonus (more complex examples are often more valuable)
	score += s.scoreComplexity(example)

	// Normalize score to 0-1 range
	return score / 4.0, nil
}

// scoreInputQuality scores the quality of the input text
func (s *QualityExampleSelector) scoreInputQuality(input string) float64 {
	if len(input) == 0 {
		return 0.0
	}

	var score float64

	// Length score (optimal range 100-500 characters)
	length := len(input)
	if length >= 100 && length <= 500 {
		score += 1.0
	} else if length >= 50 && length <= 1000 {
		score += 0.7
	} else if length >= 20 {
		score += 0.4
	}

	return score
}

// scoreOutputQuality scores the quality of the extractions
func (s *QualityExampleSelector) scoreOutputQuality(extractions []*extraction.Extraction) float64 {
	if len(extractions) == 0 {
		return 0.0
	}

	var score float64

	// Number of extractions (2-5 is optimal)
	numExtractions := len(extractions)
	if numExtractions >= 2 && numExtractions <= 5 {
		score += 1.0
	} else if numExtractions == 1 || (numExtractions >= 6 && numExtractions <= 10) {
		score += 0.7
	} else if numExtractions <= 15 {
		score += 0.4
	}

	return score
}

// scoreRelevance scores how relevant the example is to the task classes
func (s *QualityExampleSelector) scoreRelevance(example *extraction.ExampleData, taskClasses []string) float64 {
	if len(taskClasses) == 0 || len(example.Extractions) == 0 {
		return 0.0
	}

	// Create a set of task classes for quick lookup
	taskClassSet := make(map[string]bool)
	for _, class := range taskClasses {
		taskClassSet[strings.ToLower(class)] = true
	}

	// Count matching classes
	matchingClasses := 0
	totalClasses := 0

	for _, extraction := range example.Extractions {
		totalClasses++
		if taskClassSet[strings.ToLower(extraction.Class())] {
			matchingClasses++
		}
	}

	if totalClasses == 0 {
		return 0.0
	}

	return float64(matchingClasses) / float64(totalClasses)
}

// scoreComplexity scores the complexity of the example
func (s *QualityExampleSelector) scoreComplexity(example *extraction.ExampleData) float64 {
	var complexity float64

	// Input complexity (sentence count, word variety)
	if example.Text != "" {
		sentences := strings.Count(example.Text, ".") + strings.Count(example.Text, "!") + strings.Count(example.Text, "?")
		if sentences > 1 {
			complexity += 0.5
		}

		words := len(strings.Fields(example.Text))
		if words > 20 {
			complexity += 0.3
		}
	}

	// Output complexity (variety of extraction classes)
	classSet := make(map[string]bool)
	for _, extraction := range example.Extractions {
		classSet[extraction.Class()] = true
	}

	if len(classSet) > 1 {
		complexity += 0.2
	}

	return complexity
}

// getCurrentTimestamp returns the current Unix timestamp
func getCurrentTimestamp() int64 {
	// This would typically use time.Now().Unix()
	// Using a placeholder for now
	return 1640995200 // 2022-01-01 00:00:00 UTC
}

// Helper function to format examples (would be registered as template function)
func formatExample(example *extraction.ExampleData, format string) string {
	if len(example.Extractions) == 0 {
		return "{}"
	}

	switch strings.ToLower(format) {
	case "json":
		data, err := json.MarshalIndent(example.Extractions, "", "  ")
		if err != nil {
			return "{}"
		}
		return string(data)
	case "yaml":
		// YAML formatting would be implemented here
		return formatExampleAsYAML(example.Extractions)
	default:
		return formatExampleAsJSON(example.Extractions)
	}
}

// formatExampleAsJSON formats extractions as JSON
func formatExampleAsJSON(extractions []*extraction.Extraction) string {
	data, err := json.MarshalIndent(extractions, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(data)
}

// formatExampleAsYAML formats extractions as YAML (simplified)
func formatExampleAsYAML(extractions []*extraction.Extraction) string {
	var lines []string
	for i, ext := range extractions {
		lines = append(lines, fmt.Sprintf("- text: %q", ext.Text()))
		lines = append(lines, fmt.Sprintf("  class: %q", ext.Class()))
		if ext.Confidence() > 0 {
			lines = append(lines, fmt.Sprintf("  confidence: %.2f", ext.Confidence()))
		}
		if i < len(extractions)-1 {
			lines = append(lines, "")
		}
	}
	return strings.Join(lines, "\n")
}