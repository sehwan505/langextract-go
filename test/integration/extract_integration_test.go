package integration_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/sehwan505/langextract-go/pkg/document"
	"github.com/sehwan505/langextract-go/pkg/extraction"
	"github.com/sehwan505/langextract-go/pkg/langextract"
	"github.com/sehwan505/langextract-go/pkg/providers"
)

// TestExtractWithSchemaConstraints tests extraction with schema system integration
// Following patterns from Python extract_schema_integration_test.py
func TestExtractWithSchemaConstraints(t *testing.T) {
	// Skip if no API keys are available
	if !hasAnyAPIKey() {
		t.Skip("Skipping integration test: no API keys available")
	}

	// Create test schema
	schema := extraction.NewBasicExtractionSchema("medical_entities", "Medical entity extraction")
	schema.AddClass(&extraction.ClassDefinition{
		Name:        "condition",
		Description: "Medical condition",
		Fields: []*extraction.FieldDefinition{
			{
				Name:        "severity",
				Type:        "string",
				Description: "Condition severity",
				Enum:        []string{"mild", "moderate", "severe"},
			},
		},
	})

	// Create example data
	examples := []*extraction.ExampleData{
		{
			Text: "Patient has diabetes",
			Extractions: []*extraction.Extraction{
				func() *extraction.Extraction {
					ext := extraction.NewExtraction("condition", "diabetes")
					ext.AddAttribute("severity", "moderate")
					return ext
				}(),
			},
		},
	}

	testCases := []struct {
		name                  string
		modelID               string
		useSchemaConstraints  bool
		expectSchemaUsed      bool
		setupEnv              func()
		cleanupEnv            func()
	}{
		{
			name:                 "Gemini with schema constraints",
			modelID:              "gemini-2.5-flash",
			useSchemaConstraints: true,
			expectSchemaUsed:     true,
			setupEnv: func() {
				if key := getGeminiAPIKey(); key != "" {
					os.Setenv("GEMINI_API_KEY", key)
				}
			},
			cleanupEnv: func() {
				os.Unsetenv("GEMINI_API_KEY")
			},
		},
		{
			name:                 "Gemini without schema constraints",
			modelID:              "gemini-2.5-flash",
			useSchemaConstraints: false,
			expectSchemaUsed:     false,
			setupEnv: func() {
				if key := getGeminiAPIKey(); key != "" {
					os.Setenv("GEMINI_API_KEY", key)
				}
			},
			cleanupEnv: func() {
				os.Unsetenv("GEMINI_API_KEY")
			},
		},
		{
			name:                 "OpenAI with schema constraints",
			modelID:              "gpt-4o-mini",
			useSchemaConstraints: true,
			expectSchemaUsed:     true,
			setupEnv: func() {
				if key := getOpenAIAPIKey(); key != "" {
					os.Setenv("OPENAI_API_KEY", key)
				}
			},
			cleanupEnv: func() {
				os.Unsetenv("OPENAI_API_KEY")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Skip if this specific provider's API key is not available
			tc.setupEnv()
			defer tc.cleanupEnv()

			if !isProviderAvailable(tc.modelID) {
				t.Skipf("Skipping test: %s provider not available", tc.modelID)
			}

			// Create extraction options
			options := langextract.NewExtractOptions().
				WithModelID(tc.modelID).
				WithPromptDescription("Extract medical conditions").
				WithExamples(examples).
				WithValidation(tc.useSchemaConstraints).
				WithTimeout(30 * time.Second)

			// If using schema constraints, add the schema
			if tc.useSchemaConstraints {
				options.WithSchema(schema)
			}

			// Run extraction
			testDoc := document.NewDocument("Patient has hypertension and shows signs of anxiety")
			
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			options.WithContext(ctx)
			result, err := langextract.Extract(testDoc, options)
			if err != nil {
				// For integration tests, we might expect some failures due to API limits
				if strings.Contains(err.Error(), "rate limit") || 
				   strings.Contains(err.Error(), "quota") ||
				   strings.Contains(err.Error(), "timeout") {
					t.Skipf("Skipping due to API limitation: %v", err)
				}
				t.Fatalf("Extract() error = %v", err)
			}

			if result == nil {
				t.Fatal("Extract() returned nil result")
			}

			// Verify result is an AnnotatedDocument
			if result == nil {
				t.Fatal("Extract() returned nil result")
			}
			annotated := result

			// Basic validation of results
			if annotated.Text != testDoc.Text {
				t.Errorf("AnnotatedDocument.Text = %q, want %q", annotated.Text, testDoc.Text)
			}

			// Verify extractions exist (schema should help get better results)
			if tc.useSchemaConstraints && len(annotated.Extractions) == 0 {
				t.Log("Warning: No extractions found with schema constraints")
			}

			// If schema constraints are used, validate extractions against schema
			if tc.useSchemaConstraints {
				for i, ext := range annotated.Extractions {
					if err := schema.ValidateExtraction(ext); err != nil {
						t.Errorf("Extraction %d failed schema validation: %v", i, err)
					}
				}
			}
		})
	}
}

// TestExtractWithMockProvider tests extraction with mock providers for unit testing
func TestExtractWithMockProvider(t *testing.T) {
	// Create a mock provider for testing
	mockProvider := &MockLanguageModelProvider{
		modelID: "mock-model",
		responses: []string{
			`{"extractions": [{"extraction_class": "condition", "extraction_text": "hypertension"}]}`,
		},
	}

	// Register mock provider
	registry := providers.NewProviderRegistry()
	registry.Register("mock", func(config *providers.ModelConfig) (providers.BaseLanguageModel, error) {
		return mockProvider, nil
	})

	// Create schema
	schema := extraction.NewBasicExtractionSchema("test_schema", "Test schema")
	schema.AddClass(&extraction.ClassDefinition{
		Name:        "condition",
		Description: "Medical condition",
	})

	// Create extraction options with mock provider
	options := langextract.NewExtractOptions().
		WithModelID("mock-model").
		WithPromptDescription("Extract conditions").
		WithSchema(schema).
		WithValidation(true)

	// Test extraction
	testDoc := document.NewDocument("Patient has hypertension")
	
	ctx := context.Background()
	options.WithContext(ctx)
	result, err := langextract.Extract(testDoc, options)
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	if result == nil {
		t.Fatal("Extract() returned nil result")
	}
	annotated := result

	// Verify extractions
	if len(annotated.Extractions) != 1 {
		t.Errorf("Expected 1 extraction, got %d", len(annotated.Extractions))
	}

	if len(annotated.Extractions) > 0 {
		ext := annotated.Extractions[0]
		if ext.ExtractionClass != "condition" {
			t.Errorf("Extraction class = %q, want 'condition'", ext.ExtractionClass)
		}
		if ext.ExtractionText != "hypertension" {
			t.Errorf("Extraction text = %q, want 'hypertension'", ext.ExtractionText)
		}
	}
}

// TestExtractWithDifferentDocumentTypes tests extraction with various document types
func TestExtractWithDifferentDocumentTypes(t *testing.T) {
	if !hasAnyAPIKey() {
		t.Skip("Skipping integration test: no API keys available")
	}

	// Create simple schema
	schema := extraction.NewBasicExtractionSchema("entities", "Entity extraction")
	schema.AddClass(&extraction.ClassDefinition{
		Name:        "person",
		Description: "Person name",
	})
	schema.AddClass(&extraction.ClassDefinition{
		Name:        "location",
		Description: "Location name",
	})

	testCases := []struct {
		name     string
		document interface{}
		wantType string
	}{
		{
			name:     "Document object",
			document: document.NewDocument("John works in New York"),
			wantType: "*document.AnnotatedDocument",
		},
		{
			name:     "Plain text string",
			document: "Alice lives in London",
			wantType: "*document.AnnotatedDocument",
		},
		{
			name:     "Document with context",
			document: document.NewDocumentWithContext("Bob visits Paris", "Travel context"),
			wantType: "*document.AnnotatedDocument",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Use first available provider
			modelID := getFirstAvailableModel()
			if modelID == "" {
				t.Skip("No available model for testing")
			}

			options := langextract.NewExtractOptions().
				WithModelID(modelID).
				WithPromptDescription("Extract person and location entities").
				WithSchema(schema).
				WithValidation(true).
				WithTimeout(30 * time.Second)

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			options.WithContext(ctx)
			result, err := langextract.Extract(tc.document, options)
			if err != nil {
				if isAPILimitError(err) {
					t.Skipf("Skipping due to API limitation: %v", err)
				}
				t.Fatalf("Extract() error = %v", err)
			}

			// Verify result type
			resultType := getTypeName(result)
			if resultType != tc.wantType {
				t.Errorf("Extract() returned %s, want %s", resultType, tc.wantType)
			}

			// Basic validation
			if result == nil {
				t.Fatal("Result is nil")
			}
			annotated := result

			if annotated.Text == "" {
				t.Error("AnnotatedDocument has empty text")
			}
		})
	}
}

// TestExtractWithBatchProcessing tests batch document processing
func TestExtractWithBatchProcessing(t *testing.T) {
	if !hasAnyAPIKey() {
		t.Skip("Skipping integration test: no API keys available")
	}

	// Create batch of documents
	documents := []interface{}{
		document.NewDocument("Dr. Smith works at Memorial Hospital"),
		document.NewDocument("Patient John has diabetes"),
		document.NewDocument("The clinic is located in downtown Seattle"),
	}

	// Create schema
	schema := extraction.NewBasicExtractionSchema("medical_entities", "Medical entity extraction")
	schema.AddClass(&extraction.ClassDefinition{Name: "person"})
	schema.AddClass(&extraction.ClassDefinition{Name: "condition"})
	schema.AddClass(&extraction.ClassDefinition{Name: "location"})
	schema.AddClass(&extraction.ClassDefinition{Name: "organization"})

	modelID := getFirstAvailableModel()
	if modelID == "" {
		t.Skip("No available model for testing")
	}

	options := langextract.NewExtractOptions().
		WithModelID(modelID).
		WithPromptDescription("Extract medical entities").
		WithSchema(schema).
		WithValidation(true).
		WithTimeout(60 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Process documents in batch
	results := make([]*document.AnnotatedDocument, len(documents))
	for i, doc := range documents {
		options.WithContext(ctx)
		result, err := langextract.Extract(doc, options)
		if err != nil {
			if isAPILimitError(err) {
				t.Skipf("Skipping batch test due to API limitation: %v", err)
			}
			t.Errorf("Extract() for document %d error = %v", i, err)
			continue
		}

		if result == nil {
			t.Errorf("Document %d: result is nil", i)
			continue
		}
		annotated := result

		results[i] = annotated
	}

	// Verify all results
	for i, result := range results {
		if result == nil {
			continue // Skip failed extractions
		}

		if result.Text == "" {
			t.Errorf("Document %d: empty text", i)
		}

		// Validate extractions against schema
		for j, ext := range result.Extractions {
			if err := schema.ValidateExtraction(ext); err != nil {
				t.Errorf("Document %d, extraction %d: schema validation failed: %v", i, j, err)
			}
		}
	}

	t.Logf("Successfully processed %d documents with batch processing", len(documents))
}

// TestExtractWithErrorHandling tests various error scenarios
func TestExtractWithErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		document    interface{}
		options     *langextract.ExtractOptions
		expectError bool
		errorSubstr string
	}{
		{
			name:        "nil document",
			document:    nil,
			options:     langextract.NewExtractOptions().WithModelID("gpt-4"),
			expectError: true,
			errorSubstr: "document cannot be nil",
		},
		{
			name:        "empty options",
			document:    document.NewDocument("test"),
			options:     nil,
			expectError: true,
			errorSubstr: "options cannot be nil",
		},
		{
			name:        "invalid model ID",
			document:    document.NewDocument("test"),
			options:     langextract.NewExtractOptions().WithModelID("invalid-model"),
			expectError: true,
			errorSubstr: "provider",
		},
		{
			name:        "empty document",
			document:    document.NewDocument(""),
			options:     langextract.NewExtractOptions().WithModelID("gpt-4"),
			expectError: true,
			errorSubstr: "empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.options != nil {
				tt.options.WithContext(ctx)
			}
			result, err := langextract.Extract(tt.document, tt.options)

			if tt.expectError {
				if err == nil {
					t.Errorf("Extract() error = nil, want error containing %q", tt.errorSubstr)
				} else if !strings.Contains(err.Error(), tt.errorSubstr) {
					t.Errorf("Extract() error = %q, want error containing %q", err.Error(), tt.errorSubstr)
				}
				if result != nil {
					t.Error("Extract() result should be nil when error occurs")
				}
			} else {
				if err != nil {
					t.Errorf("Extract() error = %v, want nil", err)
				}
				if result == nil {
					t.Error("Extract() result should not be nil when no error")
				}
			}
		})
	}
}

// Mock provider for testing
type MockLanguageModelProvider struct {
	modelID   string
	responses []string
	callCount int
	schema    interface{}
	fenceOutput bool
}

func (m *MockLanguageModelProvider) Infer(ctx context.Context, prompts []string, options map[string]interface{}) ([][]providers.ScoredOutput, error) {
	results := make([][]providers.ScoredOutput, len(prompts))
	for i := range prompts {
		responseIdx := m.callCount % len(m.responses)
		results[i] = []providers.ScoredOutput{
			{Output: m.responses[responseIdx], Score: 1.0},
		}
		m.callCount++
	}
	return results, nil
}

func (m *MockLanguageModelProvider) ParseOutput(output string) (interface{}, error) {
	return output, nil
}

func (m *MockLanguageModelProvider) ApplySchema(schema interface{}) {
	m.schema = schema
}

func (m *MockLanguageModelProvider) SetFenceOutput(enabled bool) {
	m.fenceOutput = enabled
}

func (m *MockLanguageModelProvider) GetModelID() string {
	return m.modelID
}

func (m *MockLanguageModelProvider) IsAvailable() bool {
	return true
}

// Helper functions
func hasAnyAPIKey() bool {
	return getGeminiAPIKey() != "" || getOpenAIAPIKey() != "" || getOllamaBaseURL() != ""
}

func getGeminiAPIKey() string {
	if key := os.Getenv("GEMINI_API_KEY"); key != "" {
		return key
	}
	if key := os.Getenv("GOOGLE_API_KEY"); key != "" {
		return key
	}
	return os.Getenv("LANGEXTRACT_API_KEY")
}

func getOpenAIAPIKey() string {
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		return key
	}
	return os.Getenv("LANGEXTRACT_API_KEY")
}

func getOllamaBaseURL() string {
	if url := os.Getenv("OLLAMA_BASE_URL"); url != "" {
		return url
	}
	return "http://localhost:11434"
}

func isProviderAvailable(modelID string) bool {
	switch {
	case strings.Contains(modelID, "gemini"):
		return getGeminiAPIKey() != ""
	case strings.Contains(modelID, "gpt") || strings.Contains(modelID, "openai"):
		return getOpenAIAPIKey() != ""
	case strings.Contains(modelID, "llama") || strings.Contains(modelID, "ollama"):
		// For Ollama, we assume it's available if OLLAMA_BASE_URL is set or default
		return true
	default:
		return false
	}
}

func getFirstAvailableModel() string {
	if getGeminiAPIKey() != "" {
		return "gemini-2.5-flash"
	}
	if getOpenAIAPIKey() != "" {
		return "gpt-4o-mini"
	}
	if getOllamaBaseURL() != "" {
		return "llama3.2"
	}
	return ""
}

func isAPILimitError(err error) bool {
	errStr := err.Error()
	return strings.Contains(errStr, "rate limit") ||
		   strings.Contains(errStr, "quota") ||
		   strings.Contains(errStr, "timeout") ||
		   strings.Contains(errStr, "429") ||
		   strings.Contains(errStr, "503")
}

func getTypeName(v interface{}) string {
	if v == nil {
		return "nil"
	}
	switch v.(type) {
	case *document.AnnotatedDocument:
		return "*document.AnnotatedDocument"
	case *document.Document:
		return "*document.Document"
	default:
		return "unknown"
	}
}