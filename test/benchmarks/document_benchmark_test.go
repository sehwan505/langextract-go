package benchmarks

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/sehwan505/langextract-go/pkg/document"
	"github.com/sehwan505/langextract-go/pkg/extraction"
	"github.com/sehwan505/langextract-go/pkg/providers"
	"github.com/sehwan505/langextract-go/pkg/types"
)

// BenchmarkDocumentCreation benchmarks document creation with various text sizes
func BenchmarkDocumentCreation(b *testing.B) {
	sizes := []struct {
		name string
		size int
	}{
		{"Small_100", 100},
		{"Medium_1KB", 1024},
		{"Large_10KB", 10240},
		{"XLarge_100KB", 102400},
	}

	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			text := strings.Repeat("A", size.size)
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = document.NewDocument(text)
			}
		})
	}
}

// BenchmarkDocumentTokenization benchmarks text tokenization performance
func BenchmarkDocumentTokenization(b *testing.B) {
	sizes := []struct {
		name string
		text string
	}{
		{"Small", "This is a small document with few words."},
		{"Medium", strings.Repeat("This is a medium sized document with many repeated words. ", 50)},
		{"Large", strings.Repeat("This is a large document with extensive content and many tokens to process efficiently during tokenization benchmarking. ", 200)},
		{"VeryLarge", strings.Repeat("This is a very large document designed to stress test the tokenization system with substantial amounts of text content that should provide meaningful performance metrics for optimization purposes. ", 1000)},
	}

	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			doc := document.NewDocument(size.text)
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = doc.TokenizedText()
			}
		})
	}
}

// BenchmarkDocumentTokenizationCached benchmarks cached tokenization performance
func BenchmarkDocumentTokenizationCached(b *testing.B) {
	text := strings.Repeat("This is a test document for tokenization caching benchmarks. ", 1000)
	doc := document.NewDocument(text)
	
	// Prime the cache
	_ = doc.TokenizedText()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = doc.TokenizedText()
	}
}

// BenchmarkExtractionCreation benchmarks extraction object creation
func BenchmarkExtractionCreation(b *testing.B) {
	class := "person"
	text := "John Doe"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ext := extraction.NewExtraction(class, text)
		ext.AddAttribute("confidence", 0.95)
		ext.AddAttribute("source", "test")
		ext.SetCharInterval(&types.CharInterval{StartPos: 0, EndPos: 8})
	}
}

// BenchmarkExtractionCopy benchmarks deep copying of extractions
func BenchmarkExtractionCopy(b *testing.B) {
	// Create a complex extraction to copy
	ext := extraction.NewExtraction("person", "John Doe Smith")
	ext.SetCharInterval(&types.CharInterval{StartPos: 0, EndPos: 13})
	ext.SetTokenInterval(&types.TokenInterval{StartToken: 0, EndToken: 3})
	ext.SetAlignmentStatus(types.AlignmentExact)
	ext.AddAttribute("confidence", 0.95)
	ext.AddAttribute("source", "benchmark")
	ext.AddAttribute("metadata", map[string]interface{}{
		"processed_at": time.Now(),
		"version":      "1.0",
	})
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ext.Copy()
	}
}

// BenchmarkSchemaValidation benchmarks schema validation performance
func BenchmarkSchemaValidation(b *testing.B) {
	// Create schema
	schema := extraction.NewBasicExtractionSchema("benchmark_schema", "Benchmark schema")
	schema.AddClass(&extraction.ClassDefinition{
		Name: "person",
		Fields: []*extraction.FieldDefinition{
			{Name: "age", Type: "number", Minimum: ptrFloat64(0), Maximum: ptrFloat64(150)},
			{Name: "gender", Type: "string", Enum: []string{"male", "female", "other"}},
			{Name: "confidence", Type: "number", Minimum: ptrFloat64(0.0), Maximum: ptrFloat64(1.0)},
		},
	})

	// Create extraction to validate
	ext := extraction.NewExtraction("person", "John Doe")
	ext.AddAttribute("age", 30)
	ext.AddAttribute("gender", "male")
	ext.AddAttribute("confidence", 0.95)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = schema.ValidateExtraction(ext)
	}
}

// BenchmarkSchemaJSONGeneration benchmarks JSON schema generation
func BenchmarkSchemaJSONGeneration(b *testing.B) {
	// Create complex schema
	schema := extraction.NewBasicExtractionSchema("complex_schema", "Complex benchmark schema")
	
	// Add multiple classes
	for i := 0; i < 10; i++ {
		className := fmt.Sprintf("entity_%d", i)
		schema.AddClass(&extraction.ClassDefinition{
			Name: className,
			Fields: []*extraction.FieldDefinition{
				{Name: "confidence", Type: "number"},
				{Name: "category", Type: "string"},
				{Name: "metadata", Type: "array"},
			},
		})
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = schema.ToJSONSchema()
	}
}

// BenchmarkAnnotatedDocumentCreation benchmarks annotated document creation
func BenchmarkAnnotatedDocumentCreation(b *testing.B) {
	baseDoc := document.NewDocument("This is a test document for benchmarking annotated document creation with multiple extractions and complex metadata.")
	
	// Create extractions
	extractions := make([]*extraction.Extraction, 10)
	for i := 0; i < 10; i++ {
		ext := extraction.NewExtraction("entity", fmt.Sprintf("entity_%d", i))
		ext.SetCharInterval(&types.CharInterval{StartPos: i * 5, EndPos: (i + 1) * 5})
		ext.AddAttribute("confidence", 0.95)
		extractions[i] = ext
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		annotatedDoc := document.NewAnnotatedDocument(baseDoc)
		annotatedDoc.AddExtractions(extractions)
		_ = annotatedDoc
	}
}

// BenchmarkMockProviderInference benchmarks mock provider inference
func BenchmarkMockProviderInference(b *testing.B) {
	config := providers.NewModelConfig("mock-model")
	mockProvider := &MockBenchmarkProvider{
		config:   config,
		response: `{"extractions": [{"extraction_class": "person", "extraction_text": "John Doe"}]}`,
	}
	
	prompts := []string{
		"Extract entities from this text: John works at Google.",
		"Find all locations: Alice lives in New York.",
		"Identify organizations: Microsoft and Apple are competitors.",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = mockProvider.Infer(context.Background(), prompts, nil)
	}
}

// BenchmarkProviderRegistryLookup benchmarks provider registry lookup performance
func BenchmarkProviderRegistryLookup(b *testing.B) {
	registry := providers.NewProviderRegistry()
	providers.RegisterDefaultProviders(registry)
	
	// Add many mock providers to test lookup performance
	for i := 0; i < 100; i++ {
		providerName := fmt.Sprintf("mock_provider_%d", i)
		registry.Register(providerName, func(config *providers.ModelConfig) (providers.BaseLanguageModel, error) {
			return &MockBenchmarkProvider{config: config}, nil
		})
		registry.RegisterAlias(fmt.Sprintf("mock_model_%d", i), providerName)
	}
	
	config := providers.NewModelConfig("mock_model_50")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = registry.CreateModel(config)
	}
}

// BenchmarkConcurrentExtractionCreation benchmarks concurrent extraction creation
func BenchmarkConcurrentExtractionCreation(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ext := extraction.NewExtraction("person", "John Doe")
			ext.AddAttribute("confidence", 0.95)
			ext.SetCharInterval(&types.CharInterval{StartPos: 0, EndPos: 8})
		}
	})
}

// BenchmarkConcurrentSchemaValidation benchmarks concurrent schema validation
func BenchmarkConcurrentSchemaValidation(b *testing.B) {
	schema := extraction.NewBasicExtractionSchema("concurrent_schema", "Concurrent validation schema")
	schema.AddClass(&extraction.ClassDefinition{
		Name: "person",
		Fields: []*extraction.FieldDefinition{
			{Name: "confidence", Type: "number"},
		},
	})

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ext := extraction.NewExtraction("person", "Test Person")
			ext.AddAttribute("confidence", 0.95)
			_ = schema.ValidateExtraction(ext)
		}
	})
}

// BenchmarkLargeDocumentProcessing benchmarks processing of large documents
func BenchmarkLargeDocumentProcessing(b *testing.B) {
	// Create a large document (1MB)
	largeText := strings.Repeat("This is a large document with extensive content for benchmarking purposes. It contains multiple sentences and should stress test the document processing pipeline. ", 5000)
	
	sizes := []struct {
		name         string
		textMultiple int
	}{
		{"1MB", 1},
		{"5MB", 5},
		{"10MB", 10},
	}

	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			text := strings.Repeat(largeText, size.textMultiple)
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				doc := document.NewDocument(text)
				_ = doc.TokenizedText()
				_ = doc.TokenCount()
			}
		})
	}
}

// BenchmarkExtractionAttributeOperations benchmarks attribute operations
func BenchmarkExtractionAttributeOperations(b *testing.B) {
	operations := []struct {
		name string
		fn   func(*extraction.Extraction)
	}{
		{
			name: "AddAttribute",
			fn: func(ext *extraction.Extraction) {
				ext.AddAttribute("key", "value")
			},
		},
		{
			name: "GetAttribute",
			fn: func(ext *extraction.Extraction) {
				_, _ = ext.GetAttribute("key")
			},
		},
		{
			name: "GetStringAttribute",
			fn: func(ext *extraction.Extraction) {
				_, _ = ext.GetStringAttribute("key")
			},
		},
		{
			name: "GetFloatAttribute",
			fn: func(ext *extraction.Extraction) {
				_, _ = ext.GetFloatAttribute("number_key")
			},
		},
	}

	for _, op := range operations {
		b.Run(op.name, func(b *testing.B) {
			ext := extraction.NewExtraction("test", "test text")
			ext.AddAttribute("key", "value")
			ext.AddAttribute("number_key", 42.0)
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				op.fn(ext)
			}
		})
	}
}

// BenchmarkMemoryAllocations measures memory allocations for key operations
func BenchmarkMemoryAllocations(b *testing.B) {
	b.Run("DocumentCreation", func(b *testing.B) {
		text := "This is a test document for memory allocation benchmarking."
		
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = document.NewDocument(text)
		}
	})

	b.Run("ExtractionCreation", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ext := extraction.NewExtraction("person", "John Doe")
			ext.AddAttribute("confidence", 0.95)
		}
	})

	b.Run("SchemaValidation", func(b *testing.B) {
		schema := extraction.NewBasicExtractionSchema("memory_schema", "Memory test schema")
		schema.AddClass(&extraction.ClassDefinition{Name: "person"})
		ext := extraction.NewExtraction("person", "Test")
		
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = schema.ValidateExtraction(ext)
		}
	})
}

// Mock provider for benchmarking
type MockBenchmarkProvider struct {
	config   *providers.ModelConfig
	response string
	schema   interface{}
	fenceOutput bool
}

func (m *MockBenchmarkProvider) Infer(ctx context.Context, prompts []string, options map[string]interface{}) ([][]providers.ScoredOutput, error) {
	results := make([][]providers.ScoredOutput, len(prompts))
	for i := range results {
		results[i] = []providers.ScoredOutput{
			{Output: m.response, Score: 1.0},
		}
	}
	return results, nil
}

func (m *MockBenchmarkProvider) ParseOutput(output string) (interface{}, error) {
	return output, nil
}

func (m *MockBenchmarkProvider) ApplySchema(schema interface{}) {
	m.schema = schema
}

func (m *MockBenchmarkProvider) SetFenceOutput(enabled bool) {
	m.fenceOutput = enabled
}

func (m *MockBenchmarkProvider) GetModelID() string {
	return m.config.ModelID
}

func (m *MockBenchmarkProvider) IsAvailable() bool {
	return true
}

// Helper function
func ptrFloat64(v float64) *float64 {
	return &v
}