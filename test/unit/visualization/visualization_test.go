package visualization_test

import (
	"context"
	"strings"
	"testing"

	"github.com/sehwan505/langextract-go/internal/visualization"
	"github.com/sehwan505/langextract-go/pkg/document"
	"github.com/sehwan505/langextract-go/pkg/extraction"
	"github.com/sehwan505/langextract-go/pkg/types"
)

// Test helpers

func createTestDocument() *document.AnnotatedDocument {
	text := "John Smith works at Google Inc. He lives in San Francisco, California."
	
	// Create test extractions
	interval1, _ := types.NewCharInterval(0, 10)
	interval2, _ := types.NewCharInterval(20, 31)
	interval3, _ := types.NewCharInterval(44, 57)
	interval4, _ := types.NewCharInterval(59, 69)
	
	extractions := []*extraction.Extraction{
		extraction.NewExtractionWithInterval(
			"PERSON", 
			"John Smith", 
			interval1,
		),
		extraction.NewExtractionWithInterval(
			"ORGANIZATION", 
			"Google Inc.", 
			interval2,
		),
		extraction.NewExtractionWithInterval(
			"LOCATION", 
			"San Francisco", 
			interval3,
		),
		extraction.NewExtractionWithInterval(
			"LOCATION", 
			"California", 
			interval4,
		),
	}
	
	doc := document.NewDocument(text)
	annotatedDoc := document.NewAnnotatedDocument(doc)
	annotatedDoc.AddExtractions(extractions)
	return annotatedDoc
}

func createEmptyDocument() *document.AnnotatedDocument {
	text := "This is a simple text without any extractions."
	doc := document.NewDocument(text)
	return document.NewAnnotatedDocument(doc)
}

// HTMLGenerator Tests

func TestHTMLGenerator_Generate(t *testing.T) {
	tests := []struct {
		name        string
		doc         *document.AnnotatedDocument
		opts        *visualization.VisualizationOptions
		expectError bool
		contains    []string
	}{
		{
			name:        "Valid document with extractions",
			doc:         createTestDocument(),
			opts:        visualization.DefaultVisualizationOptions(),
			expectError: false,
			contains:    []string{"lx-animated-wrapper", "John Smith", "Google Inc.", "PERSON", "ORGANIZATION"},
		},
		{
			name:        "Empty document",
			doc:         createEmptyDocument(),
			opts:        visualization.DefaultVisualizationOptions(),
			expectError: false,
			contains:    []string{"No valid extractions"},
		},
		{
			name:        "Nil document",
			doc:         nil,
			opts:        visualization.DefaultVisualizationOptions(),
			expectError: true,
			contains:    nil,
		},
		{
			name:        "Custom colors",
			doc:         createTestDocument(),
			opts:        visualization.DefaultVisualizationOptions().WithCustomColors(map[string]string{"PERSON": "#FF0000"}),
			expectError: false,
			contains:    []string{"#FF0000", "John Smith"},
		},
		{
			name:        "No legend",
			doc:         createTestDocument(),
			opts:        visualization.DefaultVisualizationOptions().WithShowLegend(false),
			expectError: false,
			contains:    []string{"lx-animated-wrapper"},
		},
		{
			name:        "GIF optimized",
			doc:         createTestDocument(),
			opts:        &visualization.VisualizationOptions{Format: visualization.OutputFormatHTML, GIFOptimized: true, ShowLegend: true},
			expectError: false,
			contains:    []string{"lx-gif-optimized"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := visualization.NewHTMLGenerator(tt.opts)
			
			result, err := generator.Generate(context.Background(), tt.doc, tt.opts)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if result == "" {
				t.Errorf("Expected non-empty result")
				return
			}
			
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("Result should contain '%s'", expected)
				}
			}
		})
	}
}

func TestHTMLGenerator_GetSupportedFormats(t *testing.T) {
	generator := visualization.NewHTMLGenerator(nil)
	formats := generator.GetSupportedFormats()
	
	if len(formats) != 1 {
		t.Errorf("Expected 1 format, got %d", len(formats))
	}
	
	if formats[0] != visualization.OutputFormatHTML {
		t.Errorf("Expected HTML format, got %s", formats[0])
	}
}

func TestHTMLGenerator_Validate(t *testing.T) {
	generator := visualization.NewHTMLGenerator(nil)
	
	if err := generator.Validate(); err != nil {
		t.Errorf("Validation failed: %v", err)
	}
}

// JSONExporter Tests

func TestJSONExporter_Export(t *testing.T) {
	tests := []struct {
		name        string
		doc         *document.AnnotatedDocument
		opts        *visualization.ExportOptions
		expectError bool
		contains    []string
	}{
		{
			name:        "Valid document with extractions",
			doc:         createTestDocument(),
			opts:        visualization.DefaultExportOptions().WithFormat(visualization.OutputFormatJSON),
			expectError: false,
			contains:    []string{"\"text\":", "\"class\":", "John Smith", "PERSON"},
		},
		{
			name:        "Pretty formatted JSON",
			doc:         createTestDocument(),
			opts:        visualization.DefaultExportOptions().WithFormat(visualization.OutputFormatJSON).WithPretty(true),
			expectError: false,
			contains:    []string{"  \"text\":", "  \"class\":"},
		},
		{
			name:        "Nil document",
			doc:         nil,
			opts:        visualization.DefaultExportOptions().WithFormat(visualization.OutputFormatJSON),
			expectError: true,
			contains:    nil,
		},
		{
			name:        "Filter by class",
			doc:         createTestDocument(),
			opts:        visualization.DefaultExportOptions().WithFormat(visualization.OutputFormatJSON).WithFilterClasses([]string{"PERSON"}),
			expectError: false,
			contains:    []string{"John Smith", "PERSON"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exporter := visualization.NewJSONExporter(tt.opts)
			
			result, err := exporter.Export(context.Background(), tt.doc, tt.opts)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if len(result) == 0 {
				t.Errorf("Expected non-empty result")
				return
			}
			
			resultStr := string(result)
			for _, expected := range tt.contains {
				if !strings.Contains(resultStr, expected) {
					t.Errorf("Result should contain '%s'", expected)
				}
			}
		})
	}
}

func TestJSONExporter_GetFormat(t *testing.T) {
	exporter := visualization.NewJSONExporter(nil)
	
	if exporter.GetFormat() != visualization.OutputFormatJSON {
		t.Errorf("Expected JSON format")
	}
}

func TestJSONExporter_GetMIMEType(t *testing.T) {
	exporter := visualization.NewJSONExporter(nil)
	
	expected := "application/json"
	if exporter.GetMIMEType() != expected {
		t.Errorf("Expected MIME type %s, got %s", expected, exporter.GetMIMEType())
	}
}

// CSVExporter Tests

func TestCSVExporter_Export(t *testing.T) {
	tests := []struct {
		name        string
		doc         *document.AnnotatedDocument
		opts        *visualization.ExportOptions
		expectError bool
		contains    []string
	}{
		{
			name:        "Valid document with extractions",
			doc:         createTestDocument(),
			opts:        visualization.DefaultExportOptions().WithFormat(visualization.OutputFormatCSV),
			expectError: false,
			contains:    []string{"index,text,class", "John Smith,PERSON", "Google Inc.,ORGANIZATION"},
		},
		{
			name:        "Custom delimiter",
			doc:         createTestDocument(),
			opts:        visualization.DefaultExportOptions().WithFormat(visualization.OutputFormatCSV).WithCSVDelimiter(";"),
			expectError: false,
			contains:    []string{"index;text;class", "John Smith;PERSON"},
		},
		{
			name:        "Nil document",
			doc:         nil,
			opts:        visualization.DefaultExportOptions().WithFormat(visualization.OutputFormatCSV),
			expectError: true,
			contains:    nil,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exporter := visualization.NewCSVExporter(tt.opts)
			
			result, err := exporter.Export(context.Background(), tt.doc, tt.opts)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if len(result) == 0 {
				t.Errorf("Expected non-empty result")
				return
			}
			
			resultStr := string(result)
			for _, expected := range tt.contains {
				if !strings.Contains(resultStr, expected) {
					t.Errorf("Result should contain '%s'", expected)
				}
			}
		})
	}
}

func TestCSVExporter_ExportSummary(t *testing.T) {
	exporter := visualization.NewCSVExporter(nil)
	doc := createTestDocument()
	
	result, err := exporter.ExportSummary(context.Background(), doc, nil)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}
	
	resultStr := string(result)
	expected := []string{"class,count", "PERSON,1", "ORGANIZATION,1", "LOCATION,2"}
	
	for _, exp := range expected {
		if !strings.Contains(resultStr, exp) {
			t.Errorf("Summary should contain '%s'", exp)
		}
	}
}

// MarkdownExporter Tests

func TestMarkdownExporter_Export(t *testing.T) {
	tests := []struct {
		name        string
		doc         *document.AnnotatedDocument
		opts        *visualization.ExportOptions
		expectError bool
		contains    []string
	}{
		{
			name:        "Valid document with extractions",
			doc:         createTestDocument(),
			opts:        visualization.DefaultExportOptions().WithFormat(visualization.OutputFormatMarkdown),
			expectError: false,
			contains:    []string{"# LangExtract", "John Smith", "PERSON", "## Summary"},
		},
		{
			name:        "Include text",
			doc:         createTestDocument(),
			opts:        &visualization.ExportOptions{Format: visualization.OutputFormatMarkdown, IncludeText: true},
			expectError: false,
			contains:    []string{"## Highlighted Text", "**John Smith**"},
		},
		{
			name:        "Nil document",
			doc:         nil,
			opts:        visualization.DefaultExportOptions().WithFormat(visualization.OutputFormatMarkdown),
			expectError: true,
			contains:    nil,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exporter := visualization.NewMarkdownExporter(tt.opts)
			
			result, err := exporter.Export(context.Background(), tt.doc, tt.opts)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if len(result) == 0 {
				t.Errorf("Expected non-empty result")
				return
			}
			
			resultStr := string(result)
			for _, expected := range tt.contains {
				if !strings.Contains(resultStr, expected) {
					t.Errorf("Result should contain '%s'", expected)
				}
			}
		})
	}
}

// ColorManager Tests

func TestDefaultColorManager_AssignColors(t *testing.T) {
	doc := createTestDocument()
	colorManager := visualization.NewDefaultColorManager()
	
	colors := colorManager.AssignColors(doc.Extractions)
	
	// Debug: print what we got
	t.Logf("Number of extractions: %d", len(doc.Extractions))
	for i, ext := range doc.Extractions {
		t.Logf("Extraction %d: class='%s', text='%s'", i, ext.Class(), ext.Text())
	}
	t.Logf("Colors assigned: %+v", colors)
	
	// Should have colors for all unique classes
	expectedClasses := []string{"PERSON", "ORGANIZATION", "LOCATION"}
	for _, class := range expectedClasses {
		if _, exists := colors[class]; !exists {
			t.Errorf("Expected color assignment for class %s", class)
		}
	}
	
	// Colors should be from the default palette
	for _, color := range colors {
		found := false
		for _, paletteColor := range visualization.DefaultColorPalette {
			if color == paletteColor {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Color %s not found in default palette", color)
		}
	}
}

func TestDefaultColorManager_GetColor(t *testing.T) {
	colorManager := visualization.NewDefaultColorManager()
	
	// Test getting color for non-existent class
	color1 := colorManager.GetColor("NEWCLASS")
	if color1 == "" {
		t.Errorf("Expected non-empty color")
	}
	
	// Test getting same color for same class
	color2 := colorManager.GetColor("NEWCLASS")
	if color1 != color2 {
		t.Errorf("Expected same color for same class")
	}
}

func TestStaticColorManager(t *testing.T) {
	assignments := map[string]string{
		"PERSON": "#FF0000",
		"PLACE":  "#00FF00",
	}
	
	colorManager := visualization.NewStaticColorManager(assignments)
	
	// Test assigned color
	if colorManager.GetColor("PERSON") != "#FF0000" {
		t.Errorf("Expected assigned color")
	}
	
	// Test fallback color
	fallback := colorManager.GetColor("UNKNOWN")
	if fallback != visualization.DefaultFallbackColor {
		t.Errorf("Expected fallback color")
	}
}

// DefaultVisualizer Tests

func TestDefaultVisualizer_Generate(t *testing.T) {
	tests := []struct {
		name        string
		format      visualization.OutputFormat
		expectError bool
	}{
		{"HTML", visualization.OutputFormatHTML, false},
		{"JSON", visualization.OutputFormatJSON, false},
		{"CSV", visualization.OutputFormatCSV, false},
		{"Markdown", visualization.OutputFormatMarkdown, false},
		{"Plain Text", visualization.OutputFormatPlainText, false},
	}
	
	visualizer := visualization.NewDefaultVisualizer(nil)
	doc := createTestDocument()
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &visualization.VisualizationOptions{
				Format:     tt.format,
				ShowLegend: true,
			}
			
			result, err := visualizer.Generate(context.Background(), doc, opts)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if result == "" {
				t.Errorf("Expected non-empty result")
			}
		})
	}
}

func TestDefaultVisualizer_Export(t *testing.T) {
	visualizer := visualization.NewDefaultVisualizer(nil)
	doc := createTestDocument()
	
	formats := []visualization.OutputFormat{
		visualization.OutputFormatJSON,
		visualization.OutputFormatCSV,
		visualization.OutputFormatMarkdown,
		visualization.OutputFormatHTML,
		visualization.OutputFormatPlainText,
	}
	
	for _, format := range formats {
		t.Run(string(format), func(t *testing.T) {
			opts := visualization.DefaultExportOptions().WithFormat(format)
			
			result, err := visualizer.Export(context.Background(), doc, format, opts)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if len(result) == 0 {
				t.Errorf("Expected non-empty result")
			}
		})
	}
}

func TestDefaultVisualizer_GetSupportedFormats(t *testing.T) {
	visualizer := visualization.NewDefaultVisualizer(nil)
	formats := visualizer.GetSupportedFormats()
	
	expectedFormats := []visualization.OutputFormat{
		visualization.OutputFormatHTML,
		visualization.OutputFormatJSON,
		visualization.OutputFormatCSV,
		visualization.OutputFormatMarkdown,
		visualization.OutputFormatPlainText,
	}
	
	if len(formats) != len(expectedFormats) {
		t.Errorf("Expected %d formats, got %d", len(expectedFormats), len(formats))
	}
	
	for _, expected := range expectedFormats {
		found := false
		for _, actual := range formats {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected format %s not found", expected)
		}
	}
}

func TestDefaultVisualizer_Validate(t *testing.T) {
	visualizer := visualization.NewDefaultVisualizer(nil)
	
	if err := visualizer.Validate(); err != nil {
		t.Errorf("Validation failed: %v", err)
	}
}

// Options Tests

func TestVisualizationOptions_Validate(t *testing.T) {
	tests := []struct {
		name        string
		opts        *visualization.VisualizationOptions
		expectError bool
	}{
		{
			name:        "Valid options",
			opts:        visualization.DefaultVisualizationOptions(),
			expectError: false,
		},
		{
			name: "Invalid format",
			opts: &visualization.VisualizationOptions{
				Format: "invalid",
			},
			expectError: true,
		},
		{
			name: "Invalid animation speed - too low",
			opts: &visualization.VisualizationOptions{
				Format:         visualization.OutputFormatHTML,
				AnimationSpeed: 0.05,
			},
			expectError: true,
		},
		{
			name: "Invalid animation speed - too high",
			opts: &visualization.VisualizationOptions{
				Format:         visualization.OutputFormatHTML,
				AnimationSpeed: 15.0,
			},
			expectError: true,
		},
		{
			name: "Invalid context chars",
			opts: &visualization.VisualizationOptions{
				Format:       visualization.OutputFormatHTML,
				ContextChars: -1,
			},
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.Validate()
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected validation error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected validation error: %v", err)
				}
			}
		})
	}
}

func TestExportOptions_Validate(t *testing.T) {
	tests := []struct {
		name        string
		opts        *visualization.ExportOptions
		expectError bool
	}{
		{
			name:        "Valid options",
			opts:        visualization.DefaultExportOptions(),
			expectError: false,
		},
		{
			name: "Invalid format",
			opts: &visualization.ExportOptions{
				Format: "invalid",
			},
			expectError: true,
		},
		{
			name: "Empty CSV delimiter",
			opts: &visualization.ExportOptions{
				Format:       visualization.OutputFormatCSV,
				CSVDelimiter: "",
			},
			expectError: true,
		},
		{
			name: "Invalid sort order",
			opts: &visualization.ExportOptions{
				Format:    visualization.OutputFormatJSON,
				SortOrder: "invalid",
			},
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.Validate()
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected validation error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected validation error: %v", err)
				}
			}
		})
	}
}

// OutputFormat Tests

func TestOutputFormat_IsValid(t *testing.T) {
	validFormats := []visualization.OutputFormat{
		visualization.OutputFormatHTML,
		visualization.OutputFormatJSON,
		visualization.OutputFormatCSV,
		visualization.OutputFormatMarkdown,
		visualization.OutputFormatPlainText,
	}
	
	for _, format := range validFormats {
		if !format.IsValid() {
			t.Errorf("Format %s should be valid", format)
		}
	}
	
	invalidFormat := visualization.OutputFormat("invalid")
	if invalidFormat.IsValid() {
		t.Errorf("Invalid format should not be valid")
	}
}

func TestOutputFormat_GetMIMEType(t *testing.T) {
	tests := []struct {
		format   visualization.OutputFormat
		expected string
	}{
		{visualization.OutputFormatHTML, "text/html"},
		{visualization.OutputFormatJSON, "application/json"},
		{visualization.OutputFormatCSV, "text/csv"},
		{visualization.OutputFormatMarkdown, "text/markdown"},
		{visualization.OutputFormatPlainText, "text/plain"},
	}
	
	for _, tt := range tests {
		if tt.format.GetMIMEType() != tt.expected {
			t.Errorf("Expected MIME type %s for format %s, got %s", 
				tt.expected, tt.format, tt.format.GetMIMEType())
		}
	}
}

// Registry Tests

func TestVisualizationRegistry(t *testing.T) {
	registry := visualization.NewVisualizationRegistry()
	
	// Test default visualizer exists
	_, err := registry.GetVisualizer("default")
	if err != nil {
		t.Errorf("Expected default visualizer to exist")
	}
	
	// Test registering custom visualizer
	customViz := visualization.NewDefaultVisualizer(nil)
	err = registry.RegisterVisualizer("custom", customViz)
	if err != nil {
		t.Errorf("Failed to register custom visualizer: %v", err)
	}
	
	// Test getting custom visualizer
	retrieved, err := registry.GetVisualizer("custom")
	if err != nil {
		t.Errorf("Failed to get custom visualizer: %v", err)
	}
	
	if retrieved != customViz {
		t.Errorf("Retrieved visualizer is not the same as registered")
	}
	
	// Test getting non-existent visualizer
	_, err = registry.GetVisualizer("nonexistent")
	if err == nil {
		t.Errorf("Expected error for non-existent visualizer")
	}
	
	// Test available visualizers
	available := registry.GetAvailableVisualizers()
	if len(available) < 2 {
		t.Errorf("Expected at least 2 visualizers")
	}
}

// Error Tests

func TestVisualizationErrors(t *testing.T) {
	// Test error types
	validationErr := visualization.NewValidationError("test message", nil)
	if !visualization.IsValidationError(validationErr) {
		t.Errorf("Expected validation error")
	}
	
	renderingErr := visualization.NewRenderingError("test message", nil)
	if !visualization.IsRenderingError(renderingErr) {
		t.Errorf("Expected rendering error")
	}
	
	exportErr := visualization.NewExportError("test message", visualization.OutputFormatJSON, nil)
	if !visualization.IsExportError(exportErr) {
		t.Errorf("Expected export error")
	}
}

// Template Renderer Tests

func TestDefaultTemplateRenderer(t *testing.T) {
	renderer := visualization.NewDefaultTemplateRenderer()
	
	// Test registering template
	template := "Hello {{.Name}}!"
	err := renderer.RegisterTemplate("test", template)
	if err != nil {
		t.Errorf("Failed to register template: %v", err)
	}
	
	// Test rendering template
	data := map[string]interface{}{"Name": "World"}
	result, err := renderer.Render(context.Background(), "test", data)
	if err != nil {
		t.Errorf("Failed to render template: %v", err)
	}
	
	expected := "Hello World!"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
	
	// Test getting template
	_, err = renderer.GetTemplate("test")
	if err != nil {
		t.Errorf("Failed to get template: %v", err)
	}
	
	// Test available templates
	templates := renderer.GetAvailableTemplates()
	if len(templates) != 1 || templates[0] != "test" {
		t.Errorf("Expected one template named 'test'")
	}
}

// Benchmark Tests

func BenchmarkHTMLGeneration(b *testing.B) {
	doc := createTestDocument()
	generator := visualization.NewHTMLGenerator(nil)
	opts := visualization.DefaultVisualizationOptions()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := generator.Generate(context.Background(), doc, opts)
		if err != nil {
			b.Fatalf("Generation failed: %v", err)
		}
	}
}

func BenchmarkJSONExport(b *testing.B) {
	doc := createTestDocument()
	exporter := visualization.NewJSONExporter(nil)
	opts := visualization.DefaultExportOptions().WithFormat(visualization.OutputFormatJSON)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := exporter.Export(context.Background(), doc, opts)
		if err != nil {
			b.Fatalf("Export failed: %v", err)
		}
	}
}

func BenchmarkCSVExport(b *testing.B) {
	doc := createTestDocument()
	exporter := visualization.NewCSVExporter(nil)
	opts := visualization.DefaultExportOptions().WithFormat(visualization.OutputFormatCSV)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := exporter.Export(context.Background(), doc, opts)
		if err != nil {
			b.Fatalf("Export failed: %v", err)
		}
	}
}