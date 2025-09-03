package prompt_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/sehwan505/langextract-go/internal/prompt"
	"github.com/sehwan505/langextract-go/pkg/extraction"
	"github.com/sehwan505/langextract-go/pkg/types"
)

// TestPromptOptions tests the prompt options functionality
func TestPromptOptions(t *testing.T) {
	t.Run("DefaultOptions", func(t *testing.T) {
		opts := prompt.DefaultPromptOptions()

		if opts.MaxExamples != 5 {
			t.Errorf("Expected max examples 5, got %d", opts.MaxExamples)
		}
		if opts.ExampleSelectionStrategy != prompt.SelectionStrategyBest {
			t.Errorf("Expected best selection strategy, got %v", opts.ExampleSelectionStrategy)
		}
		if !opts.IncludeInstructions {
			t.Error("Expected include instructions to be true")
		}
		if !opts.UseSchemaConstraints {
			t.Error("Expected use schema constraints to be true")
		}
		if opts.OutputFormat != "json" {
			t.Errorf("Expected JSON output format, got %s", opts.OutputFormat)
		}
	})

	t.Run("FluentAPI", func(t *testing.T) {
		opts := prompt.DefaultPromptOptions().
			WithMaxExamples(3).
			WithExampleSelection(prompt.SelectionStrategyDiverse).
			WithOutputFormat("yaml").
			WithVariable("custom", "value").
			WithDebug(true)

		if opts.MaxExamples != 3 {
			t.Errorf("Expected max examples 3, got %d", opts.MaxExamples)
		}
		if opts.ExampleSelectionStrategy != prompt.SelectionStrategyDiverse {
			t.Error("Expected diverse selection strategy")
		}
		if opts.OutputFormat != "yaml" {
			t.Errorf("Expected YAML output format, got %s", opts.OutputFormat)
		}
		if opts.Variables["custom"] != "value" {
			t.Error("Expected custom variable to be set")
		}
		if !opts.Debug {
			t.Error("Expected debug to be enabled")
		}
	})

	t.Run("Validation", func(t *testing.T) {
		opts := &prompt.PromptOptions{
			MaxExamples: -1,
			Temperature: 3.0,
		}

		err := opts.Validate()
		if err == nil {
			t.Error("Expected validation error for invalid options")
		}

		opts.MaxExamples = 5
		opts.Temperature = 1.0
		opts.OutputFormat = ""

		err = opts.Validate()
		if err != nil {
			t.Errorf("Expected no validation error, got: %v", err)
		}

		if opts.OutputFormat != "json" {
			t.Error("Expected default output format to be set to json")
		}
	})
}

// TestExtractionTask tests the extraction task functionality
func TestExtractionTask(t *testing.T) {
	t.Run("BasicTask", func(t *testing.T) {
		task := &prompt.ExtractionTask{
			Description:      "Extract people and places",
			Classes:          []string{"person", "location"},
			Instructions:     []string{"Use exact text", "Include context"},
			OutputFormat:     "json",
			MaxExtractions:   10,
			RequireGrounding: true,
		}

		if task.Description != "Extract people and places" {
			t.Error("Expected correct description")
		}
		if len(task.Classes) != 2 {
			t.Error("Expected 2 classes")
		}
		if !task.RequireGrounding {
			t.Error("Expected grounding to be required")
		}
	})
}

// TestFewShotPromptBuilder tests the few-shot prompt builder
func TestFewShotPromptBuilder(t *testing.T) {
	t.Run("NewFewShotPromptBuilder", func(t *testing.T) {
		builder := prompt.NewFewShotPromptBuilder(nil)

		if builder.Name() != "FewShotPromptBuilder" {
			t.Errorf("Expected correct name, got %s", builder.Name())
		}

		if err := builder.Validate(); err != nil {
			t.Errorf("Expected validation to pass, got: %v", err)
		}
	})

	t.Run("BuildPrompt", func(t *testing.T) {
		builder := prompt.NewFewShotPromptBuilder(prompt.DefaultPromptOptions())

		task := &prompt.ExtractionTask{
			Description: "Extract person names",
			Classes:     []string{"person"},
		}

		text := "John Smith visited Paris yesterday."

		ctx := context.Background()
		prompt, err := builder.BuildPrompt(ctx, task, text)

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if prompt == "" {
			t.Error("Expected non-empty prompt")
		}

		// Check that prompt contains task description
		if !strings.Contains(prompt, task.Description) {
			t.Error("Expected prompt to contain task description")
		}

		// Check that prompt contains the input text
		if !strings.Contains(prompt, text) {
			t.Error("Expected prompt to contain input text")
		}
	})

	t.Run("BuildPromptWithExamples", func(t *testing.T) {
		builder := prompt.NewFewShotPromptBuilder(prompt.DefaultPromptOptions())

		task := &prompt.ExtractionTask{
			Description: "Extract person names",
			Classes:     []string{"person"},
		}

		text := "John Smith visited Paris yesterday."

		examples := []*extraction.ExampleData{
			{
				Text: "Alice Johnson works at Google.",
				Extractions: []*extraction.Extraction{
					extraction.NewExtractionWithInterval(
						"person", 
						"Alice Johnson",
						&types.CharInterval{StartPos: 0, EndPos: 13},
					),
				},
			},
		}
		
		// Set confidence
		examples[0].Extractions[0].SetConfidence(0.95)

		ctx := context.Background()
		prompt, err := builder.BuildPromptWithExamples(ctx, task, text, examples)

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if prompt == "" {
			t.Error("Expected non-empty prompt")
		}

		// Check that prompt contains example
		if !strings.Contains(prompt, "Alice Johnson") {
			t.Error("Expected prompt to contain example text")
		}
	})

	t.Run("ErrorHandling", func(t *testing.T) {
		builder := prompt.NewFewShotPromptBuilder(prompt.DefaultPromptOptions())

		ctx := context.Background()

		// Test nil task
		_, err := builder.BuildPrompt(ctx, nil, "test")
		if err == nil {
			t.Error("Expected error for nil task")
		}

		// Test empty text
		task := &prompt.ExtractionTask{Description: "test"}
		_, err = builder.BuildPrompt(ctx, task, "")
		if err == nil {
			t.Error("Expected error for empty text")
		}
	})
}

// TestSchemaPromptBuilder tests the schema prompt builder
func TestSchemaPromptBuilder(t *testing.T) {
	t.Run("NewSchemaPromptBuilder", func(t *testing.T) {
		builder := prompt.NewSchemaPromptBuilder(nil)

		if builder.Name() != "SchemaPromptBuilder" {
			t.Errorf("Expected correct name, got %s", builder.Name())
		}

		if err := builder.Validate(); err != nil {
			t.Errorf("Expected validation to pass, got: %v", err)
		}
	})

	t.Run("BuildPromptWithSchema", func(t *testing.T) {
		// Create a mock schema
		schema := &MockExtractionSchema{}

		builder := prompt.NewSchemaPromptBuilder(prompt.DefaultPromptOptions())

		task := &prompt.ExtractionTask{
			Description: "Extract structured information",
			Schema:      schema,
			Classes:     []string{"person", "location"},
		}

		text := "John Smith visited Paris yesterday."

		ctx := context.Background()
		prompt, err := builder.BuildPrompt(ctx, task, text)

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if prompt == "" {
			t.Error("Expected non-empty prompt")
		}

		// Check schema-specific content
		if !strings.Contains(prompt, "JSON Schema") {
			t.Error("Expected prompt to contain schema information")
		}
	})

	t.Run("ErrorHandling", func(t *testing.T) {
		builder := prompt.NewSchemaPromptBuilder(prompt.DefaultPromptOptions())

		ctx := context.Background()

		// Test task without schema
		task := &prompt.ExtractionTask{Description: "test"}
		_, err := builder.BuildPrompt(ctx, task, "test text")
		if err == nil {
			t.Error("Expected error for task without schema")
		}
	})
}

// TestTemplateRenderer tests the template renderer
func TestTemplateRenderer(t *testing.T) {
	t.Run("NewDefaultTemplateRenderer", func(t *testing.T) {
		renderer := prompt.NewDefaultTemplateRenderer()

		functions := renderer.GetFunctions()
		if len(functions) == 0 {
			t.Error("Expected default functions to be registered")
		}

		// Check for some expected functions
		expectedFunctions := []string{"add", "upper", "lower", "join"}
		for _, fn := range expectedFunctions {
			if !renderer.HasFunction(fn) {
				t.Errorf("Expected function %s to be registered", fn)
			}
		}
	})

	t.Run("RegisterFunction", func(t *testing.T) {
		renderer := prompt.NewDefaultTemplateRenderer()

		testFn := func(s string) string {
			return "test_" + s
		}

		err := renderer.RegisterFunction("test", testFn)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if !renderer.HasFunction("test") {
			t.Error("Expected test function to be registered")
		}

		// Test error cases
		err = renderer.RegisterFunction("", testFn)
		if err == nil {
			t.Error("Expected error for empty function name")
		}

		err = renderer.RegisterFunction("nil_test", nil)
		if err == nil {
			t.Error("Expected error for nil function")
		}
	})

	t.Run("RenderTemplate", func(t *testing.T) {
		renderer := prompt.NewDefaultTemplateRenderer()

		template := &prompt.PromptTemplate{
			Name:     "test_template",
			Template: "Hello {{.Name}}! You have {{len .Items}} items.",
		}

		data := struct {
			Name  string
			Items []string
		}{
			Name:  "Alice",
			Items: []string{"item1", "item2", "item3"},
		}

		ctx := context.Background()
		result, err := renderer.Render(ctx, template, data)

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		expected := "Hello Alice! You have 3 items."
		if !strings.Contains(result, expected) {
			t.Errorf("Expected result to contain '%s', got: %s", expected, result)
		}
	})

	t.Run("ValidateTemplate", func(t *testing.T) {
		renderer := prompt.NewDefaultTemplateRenderer()

		// Valid template
		validTemplate := &prompt.PromptTemplate{
			Name:     "valid",
			Template: "Hello {{.Name}}",
		}

		err := renderer.ValidateTemplate(validTemplate)
		if err != nil {
			t.Errorf("Expected no error for valid template, got: %v", err)
		}

		// Invalid template
		invalidTemplate := &prompt.PromptTemplate{
			Name:     "invalid",
			Template: "Hello {{.Name",
		}

		err = renderer.ValidateTemplate(invalidTemplate)
		if err == nil {
			t.Error("Expected error for invalid template")
		}
	})
}

// TestExampleManager tests the example manager
func TestExampleManager(t *testing.T) {
	t.Run("NewExampleManager", func(t *testing.T) {
		manager := prompt.NewExampleManager()

		collections := manager.ListExampleCollections()
		if len(collections) != 0 {
			t.Error("Expected no collections initially")
		}
	})

	t.Run("LoadExamplesFromReader", func(t *testing.T) {
		manager := prompt.NewExampleManager()

		jsonData := `{
			"name": "test_collection",
			"description": "Test examples",
			"examples": [
				{
					"text": "John Smith works at Google.",
					"extractions": [
						{
							"extraction_text": "John Smith",
							"extraction_class": "person",
							"attributes": {
								"confidence": 0.95
							}
						}
					]
				}
			]
		}`

		reader := strings.NewReader(jsonData)
		ctx := context.Background()

		collection, err := manager.LoadExamplesFromReader(ctx, reader, "test", nil)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if collection.Name != "test_collection" {
			t.Errorf("Expected collection name 'test_collection', got %s", collection.Name)
		}

		if len(collection.Examples) != 1 {
			t.Errorf("Expected 1 example, got %d", len(collection.Examples))
		}
	})

	t.Run("ValidateExample", func(t *testing.T) {
		manager := prompt.NewExampleManager()

		// Valid example
		validExample := &extraction.ExampleData{
			Text: "Test input",
			Extractions: []*extraction.Extraction{
				extraction.NewExtraction("test_class", "Test"),
			},
		}

		ctx := context.Background()
		err := manager.ValidateExample(ctx, validExample)
		if err != nil {
			t.Errorf("Expected no error for valid example, got: %v", err)
		}

		// Invalid example
		invalidExample := &extraction.ExampleData{
			Text:        "",
			Extractions: []*extraction.Extraction{},
		}

		err = manager.ValidateExample(ctx, invalidExample)
		if err == nil {
			t.Error("Expected error for invalid example")
		}
	})

	t.Run("ScoreExample", func(t *testing.T) {
		manager := prompt.NewExampleManager()

		example := &extraction.ExampleData{
			Text: "This is a test input with multiple sentences. It has good length.",
			Extractions: []*extraction.Extraction{
				extraction.NewExtraction("example", "test"),
				extraction.NewExtraction("feature", "sentences"),
			},
		}

		ctx := context.Background()
		scores, err := manager.ScoreExample(ctx, example)

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if len(scores) == 0 {
			t.Error("Expected at least one score")
		}

		if _, exists := scores["quality"]; !exists {
			t.Error("Expected quality score")
		}

		if _, exists := scores["complexity"]; !exists {
			t.Error("Expected complexity score")
		}
	})
}

// TestQualityExampleSelector tests the quality example selector
func TestQualityExampleSelector(t *testing.T) {
	t.Run("SelectExamples", func(t *testing.T) {
		selector := prompt.NewQualityExampleSelector()

		// Create test examples with varying quality
		examples := []*extraction.ExampleData{
			{
				Text: "Short",
				Extractions: []*extraction.Extraction{
					extraction.NewExtraction("test", "Short"),
				},
			},
			{
				Text: "This is a longer, higher quality example with multiple sentences. It demonstrates better extraction scenarios.",
				Extractions: []*extraction.Extraction{
					extraction.NewExtraction("feature", "quality"),
					extraction.NewExtraction("structure", "multiple sentences"),
					extraction.NewExtraction("use_case", "extraction scenarios"),
				},
			},
			{
				Text: "Medium length example for testing.",
				Extractions: []*extraction.Extraction{
					extraction.NewExtraction("size", "Medium"),
					extraction.NewExtraction("purpose", "testing"),
				},
			},
		}

		task := &prompt.ExtractionTask{
			Classes: []string{"feature", "structure", "use_case"},
		}

		ctx := context.Background()
		selected, err := selector.SelectExamples(ctx, task, examples, 2)

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if len(selected) != 2 {
			t.Errorf("Expected 2 selected examples, got %d", len(selected))
		}

		// The highest quality example should be selected first
		if len(selected[0].Text) < 50 {
			t.Error("Expected highest quality (longest) example to be selected first")
		}
	})

	t.Run("ScoreExample", func(t *testing.T) {
		selector := prompt.NewQualityExampleSelector()

		// High quality example
		highQuality := &extraction.ExampleData{
			Text: "This is a comprehensive example with multiple sentences and good structure. It provides excellent demonstration value.",
			Extractions: []*extraction.Extraction{
				extraction.NewExtraction("quality", "comprehensive"),
				extraction.NewExtraction("structure", "multiple sentences"),
				extraction.NewExtraction("purpose", "demonstration"),
			},
		}

		// Low quality example
		lowQuality := &extraction.ExampleData{
			Text: "Test",
			Extractions: []*extraction.Extraction{
				extraction.NewExtraction("simple", "Test"),
			},
		}

		task := &prompt.ExtractionTask{
			Classes: []string{"quality", "structure", "purpose"},
		}

		ctx := context.Background()

		highScore, err := selector.ScoreExample(ctx, task, highQuality)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		lowScore, err := selector.ScoreExample(ctx, task, lowQuality)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if highScore <= lowScore {
			t.Errorf("Expected high quality example to score higher: %f vs %f", highScore, lowScore)
		}
	})
}

// TestConcurrency tests concurrent access to prompt components
func TestPromptConcurrency(t *testing.T) {
	t.Run("ConcurrentPromptBuilding", func(t *testing.T) {
		builder := prompt.NewFewShotPromptBuilder(prompt.DefaultPromptOptions())

		task := &prompt.ExtractionTask{
			Description: "Extract names",
			Classes:     []string{"person"},
		}

		const numGoroutines = 10
		results := make(chan string, numGoroutines)
		errors := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				ctx := context.Background()
				text := fmt.Sprintf("Person %d is working.", id)

				prompt, err := builder.BuildPrompt(ctx, task, text)
				if err != nil {
					errors <- err
					return
				}
				results <- prompt
			}(i)
		}

		// Collect results
		for i := 0; i < numGoroutines; i++ {
			select {
			case prompt := <-results:
				if prompt == "" {
					t.Errorf("Goroutine %d returned empty prompt", i)
				}
			case err := <-errors:
				t.Errorf("Goroutine %d returned error: %v", i, err)
			case <-time.After(5 * time.Second):
				t.Fatal("Concurrent prompt building timed out")
			}
		}
	})

	t.Run("ConcurrentExampleManagement", func(t *testing.T) {
		manager := prompt.NewExampleManager()

		example := &extraction.ExampleData{
			Text: "Test concurrent access",
			Extractions: []*extraction.Extraction{
				extraction.NewExtraction("example", "Test"),
			},
		}

		const numGoroutines = 10
		results := make(chan bool, numGoroutines)
		errors := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func() {
				ctx := context.Background()

				// Test validation
				err := manager.ValidateExample(ctx, example)
				if err != nil {
					errors <- err
					return
				}

				// Test scoring
				_, err = manager.ScoreExample(ctx, example)
				if err != nil {
					errors <- err
					return
				}

				results <- true
			}()
		}

		// Collect results
		for i := 0; i < numGoroutines; i++ {
			select {
			case <-results:
				// Success
			case err := <-errors:
				t.Errorf("Goroutine returned error: %v", err)
			case <-time.After(5 * time.Second):
				t.Fatal("Concurrent example management timed out")
			}
		}
	})
}

// MockExtractionSchema is a mock implementation of ExtractionSchema
type MockExtractionSchema struct{}

func (s *MockExtractionSchema) GetName() string {
	return "mock_schema"
}

func (s *MockExtractionSchema) GetDescription() string {
	return "Mock schema for testing"
}

func (s *MockExtractionSchema) GetClasses() []string {
	return []string{"person", "location", "organization"}
}

func (s *MockExtractionSchema) ValidateExtraction(extraction *extraction.Extraction) error {
	return nil
}

func (s *MockExtractionSchema) ToJSONSchema() (map[string]interface{}, error) {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"class": map[string]interface{}{
				"type": "string",
				"enum": s.GetClasses(),
			},
			"text": map[string]interface{}{
				"type": "string",
			},
		},
		"required": []string{"class", "text"},
	}, nil
}