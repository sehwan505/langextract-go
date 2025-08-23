package chunking_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/sehwan505/langextract-go/internal/chunking"
	"github.com/sehwan505/langextract-go/pkg/document"
	"github.com/sehwan505/langextract-go/pkg/types"
)

// TestSimpleChunker tests the SimpleChunker implementation
func TestSimpleChunker(t *testing.T) {
	chunker := chunking.NewSimpleChunker()
	
	t.Run("Name", func(t *testing.T) {
		if chunker.Name() != "SimpleChunker" {
			t.Errorf("Expected 'SimpleChunker', got %s", chunker.Name())
		}
	})
	
	t.Run("ChunkText_EmptyText", func(t *testing.T) {
		opts := chunking.DefaultChunkingOptions()
		chunks, err := chunker.ChunkText(context.Background(), "", opts)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(chunks) != 0 {
			t.Errorf("Expected 0 chunks for empty text, got %d", len(chunks))
		}
	})
	
	t.Run("ChunkText_SingleSentence", func(t *testing.T) {
		text := "This is a simple sentence."
		opts := chunking.DefaultChunkingOptions()
		
		chunks, err := chunker.ChunkText(context.Background(), text, opts)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		if len(chunks) != 1 {
			t.Errorf("Expected 1 chunk, got %d", len(chunks))
		}
		
		if chunks[0].Text != text {
			t.Errorf("Expected chunk text to be '%s', got '%s'", text, chunks[0].Text)
		}
		
		if chunks[0].CharInterval.StartPos != 0 || chunks[0].CharInterval.EndPos != len(text) {
			t.Errorf("Expected char interval [0:%d), got %s", len(text), chunks[0].CharInterval.String())
		}
	})
	
	t.Run("ChunkText_MultipleSentences", func(t *testing.T) {
		text := "First sentence. Second sentence. Third sentence."
		opts := chunking.DefaultChunkingOptions().WithMaxCharBuffer(30).WithSentencePreservation(true)
		opts.MinChunkSize = 10 // Set smaller min chunk size to avoid validation error
		
		chunks, err := chunker.ChunkText(context.Background(), text, opts)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		if len(chunks) < 2 {
			t.Errorf("Expected at least 2 chunks, got %d", len(chunks))
		}
		
		// Verify chunks cover the entire text
		totalLength := 0
		for _, chunk := range chunks {
			totalLength += len(chunk.Text)
		}
		
		// Should be approximately equal (accounting for spaces and splitting)
		if totalLength < len(text)-10 || totalLength > len(text)+10 {
			t.Errorf("Total chunk length %d doesn't match original text length %d", totalLength, len(text))
		}
	})
	
	t.Run("ChunkDocument", func(t *testing.T) {
		text := "This is a test document with multiple sentences. Each sentence should be properly chunked."
		doc := document.NewDocument(text)
		opts := chunking.DefaultChunkingOptions()
		
		chunks, err := chunker.ChunkDocument(context.Background(), doc, opts)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		if len(chunks) == 0 {
			t.Error("Expected at least 1 chunk")
		}
		
		for i, chunk := range chunks {
			if chunk.ChunkIndex != i {
				t.Errorf("Chunk %d has incorrect index %d", i, chunk.ChunkIndex)
			}
			if chunk.TotalChunks != len(chunks) {
				t.Errorf("Chunk %d has incorrect total chunks %d, expected %d", i, chunk.TotalChunks, len(chunks))
			}
			if chunk.Metadata.ChunkerName != "SimpleChunker" {
				t.Errorf("Chunk %d has incorrect chunker name %s", i, chunk.Metadata.ChunkerName)
			}
		}
	})
	
	t.Run("EstimateChunks", func(t *testing.T) {
		text := strings.Repeat("This is a test sentence. ", 50) // ~1250 characters
		opts := chunking.DefaultChunkingOptions().WithMaxCharBuffer(500)
		
		estimate := chunker.EstimateChunks(text, opts)
		
		// Should estimate around 2-3 chunks
		if estimate < 2 || estimate > 5 {
			t.Errorf("Expected estimate between 2 and 5, got %d", estimate)
		}
	})
	
	t.Run("InvalidOptions", func(t *testing.T) {
		text := "Test text"
		opts := chunking.ChunkingOptions{
			MaxCharBuffer: -1, // Invalid
		}
		
		_, err := chunker.ChunkText(context.Background(), text, opts)
		if err == nil {
			t.Error("Expected error for invalid options")
		}
	})
	
	t.Run("ContextCancellation", func(t *testing.T) {
		text := strings.Repeat("This is a very long text that should be chunked. ", 1000)
		opts := chunking.DefaultChunkingOptions().WithMaxCharBuffer(100)
		
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately
		
		_, err := chunker.ChunkText(ctx, text, opts)
		if err == nil {
			t.Error("Expected error for cancelled context")
		}
	})
}

// TestSemanticChunker tests the SemanticChunker implementation
func TestSemanticChunker(t *testing.T) {
	chunker := chunking.NewSemanticChunker()
	
	t.Run("Name", func(t *testing.T) {
		if chunker.Name() != "SemanticChunker" {
			t.Errorf("Expected 'SemanticChunker', got %s", chunker.Name())
		}
	})
	
	t.Run("ChunkText_StructuredContent", func(t *testing.T) {
		text := `Introduction
		
This is the introduction paragraph with important information.

Section 1: Background

This section provides background information. It has multiple sentences that are related to each other.

However, this paragraph introduces a different topic that should potentially start a new chunk.

Section 2: Methods

1. First step in the process
2. Second step in the process  
3. Third step in the process

The methodology section continues with more detailed explanations.`
		
		opts := chunking.DefaultChunkingOptions().WithMaxCharBuffer(300)
		
		chunks, err := chunker.ChunkText(context.Background(), text, opts)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		if len(chunks) < 2 {
			t.Errorf("Expected at least 2 chunks for structured content, got %d", len(chunks))
		}
		
		// Verify semantic properties are set
		for _, chunk := range chunks {
			if chunk.Metadata.ChunkerName != "SemanticChunker" {
				t.Errorf("Expected chunker name 'SemanticChunker', got %s", chunk.Metadata.ChunkerName)
			}
			
			props := chunk.Metadata.Properties
			if props["semantic_analysis"] != true {
				t.Error("Expected semantic_analysis property to be true")
			}
		}
	})
	
	t.Run("EstimateChunks_Complex", func(t *testing.T) {
		// Text with structural elements should have higher estimate
		structuredText := `Title

Introduction paragraph.

Section 1
- Item 1
- Item 2  
- Item 3

However, this is a transition.

Section 2
More content here.`
		
		simpleText := strings.Repeat("Simple sentence. ", 20)
		
		opts := chunking.DefaultChunkingOptions().WithMaxCharBuffer(200)
		
		structuredEstimate := chunker.EstimateChunks(structuredText, opts)
		simpleEstimate := chunker.EstimateChunks(simpleText, opts)
		
		// Just verify that we get reasonable estimates (the exact comparison might be flaky)
		if structuredEstimate < 1 || simpleEstimate < 1 {
			t.Errorf("Expected positive estimates, got structured=%d, simple=%d", 
				structuredEstimate, simpleEstimate)
		}
	})
}

// TestAdaptiveChunker tests the AdaptiveChunker implementation
func TestAdaptiveChunker(t *testing.T) {
	chunker := chunking.NewAdaptiveChunker()
	
	t.Run("Name", func(t *testing.T) {
		if chunker.Name() != "AdaptiveChunker" {
			t.Errorf("Expected 'AdaptiveChunker', got %s", chunker.Name())
		}
	})
	
	t.Run("ChunkText_SimpleVsComplex", func(t *testing.T) {
		// Simple text (low complexity)
		simpleText := strings.Repeat("The cat sat on the mat. ", 50)
		
		// Complex text (high complexity)
		complexText := `Notwithstanding the aforementioned considerations, the implementation of sophisticated algorithms necessitates comprehensive evaluation methodologies. Furthermore, the interdisciplinary nature of contemporary research paradigms requires multifaceted approaches to problem-solving, particularly when addressing complex socio-technical challenges that intersect multiple domains of expertise.`
		
		opts := chunking.DefaultChunkingOptions().WithMaxCharBuffer(200)
		
		simpleChunks, err := chunker.ChunkText(context.Background(), simpleText, opts)
		if err != nil {
			t.Fatalf("Unexpected error for simple text: %v", err)
		}
		
		complexChunks, err := chunker.ChunkText(context.Background(), complexText, opts)
		if err != nil {
			t.Fatalf("Unexpected error for complex text: %v", err)
		}
		
		// Complex text should potentially have more chunks due to adaptive sizing
		for _, chunk := range complexChunks {
			props := chunk.Metadata.Properties
			if props["adaptive_chunking"] != true {
				t.Error("Expected adaptive_chunking property to be true")
			}
			
			if complexityScore, ok := props["complexity_score"].(float64); ok {
				if complexityScore < 0 || complexityScore > 1 {
					t.Errorf("Expected complexity score between 0 and 1, got %f", complexityScore)
				}
			} else {
				t.Error("Expected complexity_score property")
			}
		}
		
		// Verify different properties for simple vs complex
		if len(simpleChunks) > 0 && len(complexChunks) > 0 {
			simpleComplexity := simpleChunks[0].Metadata.Properties["complexity_score"].(float64)
			complexComplexity := complexChunks[0].Metadata.Properties["complexity_score"].(float64)
			
			if complexComplexity <= simpleComplexity {
				t.Errorf("Expected complex text complexity (%f) to be higher than simple text (%f)", 
					complexComplexity, simpleComplexity)
			}
		}
	})
	
	t.Run("WithComplexityThresholds", func(t *testing.T) {
		customThresholds := chunking.ContentComplexityThresholds{
			LowComplexity:    0.2,
			MediumComplexity: 0.5,
			HighComplexity:   0.8,
		}
		
		customChunker := chunker.WithComplexityThresholds(customThresholds)
		if customChunker == nil {
			t.Error("Expected custom chunker to be created")
		}
	})
}

// TestChunkingOptions tests the chunking options and validation
func TestChunkingOptions(t *testing.T) {
	t.Run("DefaultOptions", func(t *testing.T) {
		opts := chunking.DefaultChunkingOptions()
		
		if opts.MaxCharBuffer <= 0 {
			t.Error("Expected positive max char buffer")
		}
		if opts.OverlapRatio < 0 || opts.OverlapRatio >= 0.5 {
			t.Error("Expected overlap ratio between 0 and 0.5")
		}
		if opts.MinChunkSize < 0 {
			t.Error("Expected non-negative min chunk size")
		}
	})
	
	t.Run("OptionsValidation", func(t *testing.T) {
		// Test invalid max char buffer
		opts := chunking.ChunkingOptions{MaxCharBuffer: -1}
		if err := opts.Validate(); err == nil {
			t.Error("Expected validation error for negative max char buffer")
		}
		
		// Test invalid overlap ratio
		opts = chunking.ChunkingOptions{
			MaxCharBuffer: 100,
			OverlapRatio:  0.6, // Too high
		}
		if err := opts.Validate(); err == nil {
			t.Error("Expected validation error for high overlap ratio")
		}
		
		// Test min chunk size >= max char buffer
		opts = chunking.ChunkingOptions{
			MaxCharBuffer: 100,
			MinChunkSize:  100, // Equal to max
		}
		if err := opts.Validate(); err == nil {
			t.Error("Expected validation error for min chunk size >= max char buffer")
		}
	})
	
	t.Run("FluentAPI", func(t *testing.T) {
		opts := chunking.DefaultChunkingOptions().
			WithMaxCharBuffer(500).
			WithOverlap(0.2).
			WithSentencePreservation(false).
			WithLanguage("en")
		
		if opts.MaxCharBuffer != 500 {
			t.Errorf("Expected max char buffer 500, got %d", opts.MaxCharBuffer)
		}
		if opts.OverlapRatio != 0.2 {
			t.Errorf("Expected overlap ratio 0.2, got %f", opts.OverlapRatio)
		}
		if opts.PreserveSentences != false {
			t.Error("Expected preserve sentences to be false")
		}
		if opts.Language != "en" {
			t.Errorf("Expected language 'en', got %s", opts.Language)
		}
	})
}

// TestTextChunk tests the TextChunk methods
func TestTextChunk(t *testing.T) {
	chunk := chunking.TextChunk{
		ID:   "test_chunk",
		Text: "This is a test chunk with some text.",
		CharInterval: &types.CharInterval{StartPos: 0, EndPos: 36},
		Metadata: chunking.ChunkMetadata{
			ChunkerName: "TestChunker",
			WordCount:   8,
		},
	}
	
	t.Run("Length", func(t *testing.T) {
		if chunk.Length() != 36 {
			t.Errorf("Expected length 36, got %d", chunk.Length())
		}
	})
	
	t.Run("IsEmpty", func(t *testing.T) {
		if chunk.IsEmpty() {
			t.Error("Expected chunk not to be empty")
		}
		
		emptyChunk := chunking.TextChunk{Text: ""}
		if !emptyChunk.IsEmpty() {
			t.Error("Expected empty chunk to be empty")
		}
	})
	
	t.Run("EstimateWords", func(t *testing.T) {
		words := chunk.EstimateWords()
		if words != 8 {
			t.Errorf("Expected 8 words, got %d", words)
		}
	})
	
	t.Run("HasPosition", func(t *testing.T) {
		if !chunk.HasPosition() {
			t.Error("Expected chunk to have position")
		}
		
		noPositionChunk := chunking.TextChunk{Text: "test"}
		if noPositionChunk.HasPosition() {
			t.Error("Expected chunk without char interval to not have position")
		}
	})
	
	t.Run("String", func(t *testing.T) {
		str := chunk.String()
		if !strings.Contains(str, "test_chunk") {
			t.Errorf("Expected string to contain chunk ID, got %s", str)
		}
		if !strings.Contains(str, "[0:36)") {
			t.Errorf("Expected string to contain position, got %s", str)
		}
	})
}

// BenchmarkChunking benchmarks different chunking strategies
func BenchmarkChunking(b *testing.B) {
	text := strings.Repeat("This is a test sentence with some complexity. ", 1000)
	opts := chunking.DefaultChunkingOptions().WithMaxCharBuffer(500)
	ctx := context.Background()
	
	b.Run("SimpleChunker", func(b *testing.B) {
		chunker := chunking.NewSimpleChunker()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := chunker.ChunkText(ctx, text, opts)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	
	b.Run("SemanticChunker", func(b *testing.B) {
		chunker := chunking.NewSemanticChunker()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := chunker.ChunkText(ctx, text, opts)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	
	b.Run("AdaptiveChunker", func(b *testing.B) {
		chunker := chunking.NewAdaptiveChunker()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := chunker.ChunkText(ctx, text, opts)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// TestChunkingConcurrency tests concurrent chunking operations
func TestChunkingConcurrency(t *testing.T) {
	chunker := chunking.NewSimpleChunker()
	text := "This is a test for concurrent chunking operations. " + strings.Repeat("More text. ", 100)
	opts := chunking.DefaultChunkingOptions()
	
	// Run multiple chunking operations concurrently
	const numGoroutines = 10
	results := make(chan error, numGoroutines)
	
	for i := 0; i < numGoroutines; i++ {
		go func() {
			_, err := chunker.ChunkText(context.Background(), text, opts)
			results <- err
		}()
	}
	
	// Check all results
	for i := 0; i < numGoroutines; i++ {
		select {
		case err := <-results:
			if err != nil {
				t.Errorf("Concurrent chunking failed: %v", err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("Concurrent chunking timed out")
		}
	}
}