// Package langextract provides a high-level API for extracting structured information
// from unstructured text using Large Language Models (LLMs).
//
// This package mirrors the functionality of Google's langextract Python library,
// providing precise source grounding, schema-driven extraction, and support for
// multiple language model providers.
//
// Basic usage:
//
//	import "github.com/sehwan505/langextract-go/pkg/langextract"
//
//	// Extract entities from text
//	opts := langextract.NewExtractOptions().
//		WithPromptDescription("Extract person names and organizations").
//		WithModelID("gemini-2.5-flash")
//
//	result, err := langextract.Extract("John Doe works at Google Inc.", opts)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Visualize results
//	html, err := langextract.Visualize(result, langextract.NewVisualizeOptions())
//	if err != nil {
//		log.Fatal(err)
//	}
package langextract

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/sehwan505/langextract-go/internal/visualization"
	"github.com/sehwan505/langextract-go/pkg/document"
	"github.com/sehwan505/langextract-go/pkg/extraction"
	"github.com/sehwan505/langextract-go/pkg/providers"
)

// TextOrDocuments represents flexible input types for extraction.
type TextOrDocuments interface{}

// Extract extracts structured information from text using language models.
// This is the main entry point that mirrors the Python langextract.extract() function.
//
// The input can be:
//   - A string of text
//   - A URL (must start with http:// or https://)
//   - A Document object
//   - A slice of any of the above
//
// Returns an AnnotatedDocument with extracted entities and their source grounding.
func Extract(input TextOrDocuments, opts *ExtractOptions) (*document.AnnotatedDocument, error) {
	if opts == nil {
		opts = NewExtractOptions()
	}

	// Validate options
	if err := opts.Validate(); err != nil {
		return nil, NewExtractError("validate_options", "invalid extraction options", err)
	}

	// Set up context with timeout
	ctx := opts.Context
	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	// Convert input to documents
	docs, err := parseInput(input)
	if err != nil {
		return nil, NewExtractError("parse_input", "failed to parse input", err)
	}

	// For now, handle single document (multi-document support in future)
	if len(docs) == 0 {
		return nil, NewExtractError("parse_input", "no documents to process", nil)
	}

	if len(docs) > 1 {
		return nil, NewExtractError("multi_document", "multi-document extraction not yet implemented", nil)
	}

	doc := docs[0]

	// Validate required parameters
	if len(opts.Examples) == 0 {
		if opts.DebugMode {
			log.Printf("Warning: No examples provided for few-shot learning")
		}
	}

	// Create provider
	provider, err := createProvider(opts)
	if err != nil {
		return nil, NewExtractError("create_provider", "failed to create language model provider", err)
	}

	// Perform extraction
	annotatedDoc, err := performExtraction(ctx, doc, provider, opts)
	if err != nil {
		return nil, NewExtractError("perform_extraction", "extraction failed", err)
	}

	return annotatedDoc, nil
}

// Visualize generates visualization output for extracted data.
// This mirrors the Python langextract.visualize() function.
//
// Returns formatted output according to the specified options.
func Visualize(doc *document.AnnotatedDocument, opts *VisualizeOptions) (string, error) {
	if doc == nil {
		return "", NewExtractError("visualize", "document cannot be nil", nil)
	}

	if opts == nil {
		opts = NewVisualizeOptions()
	}

	if err := opts.Validate(); err != nil {
		return "", NewExtractError("validate_options", "invalid visualization options", err)
	}

	switch opts.Format {
	case "html":
		return generateHTMLVisualization(doc, opts)
	case "json":
		return generateJSONVisualization(doc, opts)
	case "csv":
		return generateCSVVisualization(doc, opts)
	default:
		return "", NewExtractError("visualize", fmt.Sprintf("unsupported format: %s", opts.Format), nil)
	}
}

// parseInput converts various input types to Document objects.
func parseInput(input TextOrDocuments) ([]*document.Document, error) {
	switch v := input.(type) {
	case string:
		// Check if it's a URL
		if isURL(v) {
			return nil, NewExtractError("url_input", "URL input not yet implemented", nil)
		}
		return []*document.Document{document.NewDocument(v)}, nil

	case *document.Document:
		return []*document.Document{v}, nil

	case []*document.Document:
		return v, nil

	case []string:
		docs := make([]*document.Document, len(v))
		for i, text := range v {
			if isURL(text) {
				return nil, NewExtractError("url_input", "URL input not yet implemented", nil)
			}
			docs[i] = document.NewDocument(text)
		}
		return docs, nil

	default:
		return nil, NewExtractError("parse_input", fmt.Sprintf("unsupported input type: %T", input), nil)
	}
}

// isURL checks if a string looks like a URL.
func isURL(s string) bool {
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		_, err := url.Parse(s)
		return err == nil
	}
	return false
}

// createProvider creates a language model provider based on options.
func createProvider(opts *ExtractOptions) (providers.BaseLanguageModel, error) {
	var config *providers.ModelConfig
	
	if opts.ModelConfig != nil {
		config = opts.ModelConfig
	} else {
		config = providers.NewModelConfig(opts.ModelID)
	}

	// Apply options to config
	if opts.Temperature > 0 {
		config = config.WithTemperature(opts.Temperature)
	}
	if opts.MaxTokens > 0 {
		config = config.WithMaxTokens(opts.MaxTokens)
	}

	// Create provider using registry
	registry := providers.NewProviderRegistry()
	providers.RegisterDefaultProviders(registry)

	provider, err := registry.CreateModel(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider for model %s: %w", opts.ModelID, err)
	}

	// Check if provider is available
	if !provider.IsAvailable() {
		return nil, NewProviderError(config.Provider, "unavailable", "provider is not available", nil)
	}

	return provider, nil
}

// performExtraction executes the extraction process.
func performExtraction(ctx context.Context, doc *document.Document, provider providers.BaseLanguageModel, opts *ExtractOptions) (*document.AnnotatedDocument, error) {
	// Create annotated document
	annotatedDoc := document.NewAnnotatedDocument(doc)

	// Build prompt
	prompt, err := buildPrompt(doc.Text, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to build prompt: %w", err)
	}

	if opts.DebugMode {
		log.Printf("Generated prompt: %s", prompt)
	}

	// Apply schema if provided
	if opts.Schema != nil {
		jsonSchema, err := opts.Schema.ToJSONSchema()
		if err != nil {
			return nil, fmt.Errorf("failed to convert schema: %w", err)
		}
		provider.ApplySchema(jsonSchema)
	}

	// Perform extraction with retries
	var lastErr error
	for attempt := 0; attempt <= opts.RetryCount; attempt++ {
		if attempt > 0 {
			if opts.DebugMode {
				log.Printf("Retry attempt %d/%d", attempt, opts.RetryCount)
			}
			time.Sleep(time.Duration(attempt) * time.Second) // Exponential backoff
		}

		// Check context cancellation
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// Call provider
		results, err := provider.Infer(ctx, []string{prompt}, nil)
		if err != nil {
			lastErr = err
			if opts.DebugMode {
				log.Printf("Provider error (attempt %d): %v", attempt+1, err)
			}
			continue
		}

		if len(results) == 0 || len(results[0]) == 0 {
			lastErr = fmt.Errorf("no results returned from provider")
			continue
		}

		// Parse the response
		response := results[0][0].Output
		if opts.DebugMode {
			log.Printf("Provider response: %s", response)
		}

		// Parse extractions from response
		extractions, err := parseExtractions(response, doc.Text, provider)
		if err != nil {
			lastErr = fmt.Errorf("failed to parse extractions: %w", err)
			if opts.DebugMode {
				log.Printf("Parse error: %v", err)
			}
			continue
		}

		// Validate extractions if schema is provided
		if opts.ValidateOutput && opts.Schema != nil {
			validExtractions := make([]*extraction.Extraction, 0, len(extractions))
			for _, ext := range extractions {
				if err := opts.Schema.ValidateExtraction(ext); err != nil {
					if opts.DebugMode {
						log.Printf("Validation failed for extraction %s: %v", ext.ExtractionText, err)
					}
					continue
				}
				validExtractions = append(validExtractions, ext)
			}
			extractions = validExtractions
		}

		// Add extractions to document
		annotatedDoc.AddExtractions(extractions)

		if opts.DebugMode {
			log.Printf("Successfully extracted %d entities", len(extractions))
		}

		return annotatedDoc, nil
	}

	return nil, fmt.Errorf("extraction failed after %d retries: %w", opts.RetryCount+1, lastErr)
}

// buildPrompt constructs the prompt for the language model.
func buildPrompt(text string, opts *ExtractOptions) (string, error) {
	var prompt strings.Builder

	// Add task description
	prompt.WriteString("Extract structured information from the following text.\n\n")
	
	if opts.PromptDescription != "" {
		prompt.WriteString("Task: ")
		prompt.WriteString(opts.PromptDescription)
		prompt.WriteString("\n\n")
	}

	// Add examples if provided
	if len(opts.Examples) > 0 {
		prompt.WriteString("Examples:\n")
		for i, example := range opts.Examples {
			prompt.WriteString(fmt.Sprintf("\nExample %d:\n", i+1))
			prompt.WriteString("Text: ")
			prompt.WriteString(example.Text)
			prompt.WriteString("\n")
			
			if len(example.Extractions) > 0 {
				prompt.WriteString("Extractions:\n")
				for _, ext := range example.Extractions {
					prompt.WriteString(fmt.Sprintf("- %s: %s\n", ext.ExtractionClass, ext.ExtractionText))
				}
			}
		}
		prompt.WriteString("\n")
	}

	// Add schema information if provided
	if opts.Schema != nil {
		prompt.WriteString("Expected extraction classes: ")
		classes := opts.Schema.GetClasses()
		prompt.WriteString(strings.Join(classes, ", "))
		prompt.WriteString("\n\n")
	}

	// Add the text to process
	prompt.WriteString("Text to process:\n")
	prompt.WriteString(text)
	prompt.WriteString("\n\n")

	// Add format instructions
	prompt.WriteString("Please extract entities in the following JSON format:\n")
	prompt.WriteString("{\n")
	prompt.WriteString("  \"extractions\": [\n")
	prompt.WriteString("    {\n")
	prompt.WriteString("      \"extraction_class\": \"class_name\",\n")
	prompt.WriteString("      \"extraction_text\": \"extracted_text\",\n")
	prompt.WriteString("      \"confidence\": 0.95\n")
	prompt.WriteString("    }\n")
	prompt.WriteString("  ]\n")
	prompt.WriteString("}")

	return prompt.String(), nil
}

// parseExtractions parses the model response into Extraction objects.
func parseExtractions(response, sourceText string, provider providers.BaseLanguageModel) ([]*extraction.Extraction, error) {
	// Parse JSON response
	parsed, err := provider.ParseOutput(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse provider response: %w", err)
	}

	// Convert to map
	data, ok := parsed.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("expected JSON object, got %T", parsed)
	}

	// Get extractions array
	extractionsData, ok := data["extractions"]
	if !ok {
		return nil, fmt.Errorf("no 'extractions' field found in response")
	}

	extractionsArray, ok := extractionsData.([]interface{})
	if !ok {
		return nil, fmt.Errorf("'extractions' field is not an array")
	}

	// Parse each extraction
	extractions := make([]*extraction.Extraction, 0, len(extractionsArray))
	for _, item := range extractionsArray {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue // Skip invalid items
		}

		class, _ := itemMap["extraction_class"].(string)
		text, _ := itemMap["extraction_text"].(string)
		
		if class == "" || text == "" {
			continue // Skip incomplete extractions
		}

		ext := extraction.NewExtraction(class, text)

		// Add confidence if present
		if conf, ok := itemMap["confidence"].(float64); ok {
			ext.SetConfidence(conf)
		}

		// Add other attributes
		for key, value := range itemMap {
			if key != "extraction_class" && key != "extraction_text" && key != "confidence" {
				ext.AddAttribute(key, value)
			}
		}

		// TODO: Add text alignment to find source positions
		// This will be implemented when the alignment package is created

		extractions = append(extractions, ext)
	}

	return extractions, nil
}

// Visualize generates visualizations from annotated documents.
// This function provides multiple output formats including interactive HTML, JSON, CSV, and Markdown.
//
// The input can be:
//   - An AnnotatedDocument (the most common case)
//   - A path to a JSONL file containing annotated documents
//
// Example usage:
//
//	doc, err := langextract.Extract("John works at Google", opts)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Generate interactive HTML
//	html, err := langextract.Visualize(doc, NewVisualizeOptions().WithFormat("html"))
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Export as JSON
//	json, err := langextract.Visualize(doc, NewVisualizeOptions().WithFormat("json"))
//	if err != nil {
//		log.Fatal(err)
//	}
func Visualize(input interface{}, opts *VisualizeOptions) (string, error) {
	if opts == nil {
		opts = NewVisualizeOptions()
	}

	// Convert input to AnnotatedDocument
	var doc *document.AnnotatedDocument
	var err error

	switch v := input.(type) {
	case *document.AnnotatedDocument:
		doc = v
	case string:
		// Try to interpret as file path
		return "", fmt.Errorf("file path input not yet supported")
	default:
		return "", fmt.Errorf("unsupported input type: %T", input)
	}

	if doc == nil {
		return "", fmt.Errorf("no document to visualize")
	}

	// Convert options to internal format
	vizOpts := convertToVisualizationOptions(opts)

	// Get the default visualizer
	visualizer := visualization.GetDefaultVisualizer()
	if visualizer == nil {
		return "", fmt.Errorf("no visualizer available")
	}

	// Generate visualization
	result, err := visualizer.Generate(context.Background(), doc, vizOpts)
	if err != nil {
		return "", fmt.Errorf("visualization failed: %w", err)
	}

	return result, nil
}