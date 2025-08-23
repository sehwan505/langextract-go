package chunking

import (
	"context"
	"fmt"
	"strings"

	"github.com/sehwan505/langextract-go/pkg/document"
	"github.com/sehwan505/langextract-go/pkg/types"
)

// TextChunker defines the interface for text chunking algorithms.
// Different implementations can provide various chunking strategies
// such as sentence-based, paragraph-based, or semantic chunking.
type TextChunker interface {
	// ChunkDocument splits a document into smaller text chunks for processing.
	// Returns a slice of TextChunk objects with position information.
	ChunkDocument(ctx context.Context, doc *document.Document, opts ChunkingOptions) ([]TextChunk, error)
	
	// ChunkText splits raw text into chunks with the given options.
	// This is a convenience method for simple text chunking without document metadata.
	ChunkText(ctx context.Context, text string, opts ChunkingOptions) ([]TextChunk, error)
	
	// EstimateChunks returns an estimate of how many chunks will be created
	// for the given text. Useful for progress tracking and memory planning.
	EstimateChunks(text string, opts ChunkingOptions) int
	
	// Name returns the name of the chunking algorithm for logging and debugging.
	Name() string
}

// TextChunk represents a chunk of text with position and metadata information.
type TextChunk struct {
	// ID is a unique identifier for this chunk within the document
	ID string
	
	// Text is the actual text content of the chunk
	Text string
	
	// CharInterval represents the character position in the original document
	CharInterval *types.CharInterval
	
	// TokenInterval represents the token position if tokenization is available
	TokenInterval *types.TokenInterval
	
	// Metadata contains additional information about the chunk
	Metadata ChunkMetadata
	
	// SourceDocument is a reference to the original document
	SourceDocument *document.Document
	
	// ChunkIndex is the sequential index of this chunk (0-based)
	ChunkIndex int
	
	// TotalChunks is the total number of chunks in the document
	TotalChunks int
}

// ChunkMetadata contains additional information about a text chunk.
type ChunkMetadata struct {
	// ChunkerName is the name of the algorithm that created this chunk
	ChunkerName string
	
	// SentenceCount is the number of sentences in this chunk
	SentenceCount int
	
	// WordCount is the approximate number of words in this chunk
	WordCount int
	
	// Language is the detected language of the chunk (if available)
	Language string
	
	// Quality is a score indicating the quality of the chunking (0.0-1.0)
	Quality float64
	
	// HasOverlap indicates if this chunk overlaps with adjacent chunks
	HasOverlap bool
	
	// OverlapStart is the number of characters overlapping with the previous chunk
	OverlapStart int
	
	// OverlapEnd is the number of characters overlapping with the next chunk
	OverlapEnd int
	
	// Properties contains algorithm-specific metadata
	Properties map[string]interface{}
}

// ChunkingOptions configures the behavior of text chunking algorithms.
type ChunkingOptions struct {
	// MaxCharBuffer is the maximum number of characters per chunk
	MaxCharBuffer int
	
	// MaxTokenBuffer is the maximum number of tokens per chunk (if tokenization is available)
	MaxTokenBuffer int
	
	// OverlapRatio is the fraction of overlap between adjacent chunks (0.0-0.5)
	OverlapRatio float64
	
	// PreserveSentences indicates whether to avoid breaking sentences
	PreserveSentences bool
	
	// PreserveParagraphs indicates whether to avoid breaking paragraphs
	PreserveParagraphs bool
	
	// MinChunkSize is the minimum characters required for a valid chunk
	MinChunkSize int
	
	// Language hint for language-specific chunking rules
	Language string
	
	// CustomDelimiters are additional delimiters to consider for chunking
	CustomDelimiters []string
	
	// IncludeMetadata determines whether to compute detailed chunk metadata
	IncludeMetadata bool
	
	// Properties contains algorithm-specific options
	Properties map[string]interface{}
}

// DefaultChunkingOptions returns the default options for text chunking.
func DefaultChunkingOptions() ChunkingOptions {
	return ChunkingOptions{
		MaxCharBuffer:     1000,
		MaxTokenBuffer:    0, // 0 means token-based chunking is disabled
		OverlapRatio:      0.1,
		PreserveSentences: true,
		PreserveParagraphs: false,
		MinChunkSize:      50,
		Language:          "",
		CustomDelimiters:  []string{},
		IncludeMetadata:   true,
		Properties:        make(map[string]interface{}),
	}
}

// WithMaxCharBuffer sets the maximum character buffer size.
func (opts ChunkingOptions) WithMaxCharBuffer(size int) ChunkingOptions {
	opts.MaxCharBuffer = size
	return opts
}

// WithOverlap sets the overlap ratio between chunks.
func (opts ChunkingOptions) WithOverlap(ratio float64) ChunkingOptions {
	opts.OverlapRatio = ratio
	return opts
}

// WithSentencePreservation enables or disables sentence preservation.
func (opts ChunkingOptions) WithSentencePreservation(preserve bool) ChunkingOptions {
	opts.PreserveSentences = preserve
	return opts
}

// WithLanguage sets a language hint for chunking.
func (opts ChunkingOptions) WithLanguage(language string) ChunkingOptions {
	opts.Language = language
	return opts
}

// Validate checks if the chunking options are valid.
func (opts ChunkingOptions) Validate() error {
	if opts.MaxCharBuffer <= 0 {
		return &ChunkingError{
			Type:    "validation",
			Message: "max character buffer must be positive",
			Details: map[string]interface{}{"max_char_buffer": opts.MaxCharBuffer},
		}
	}
	if opts.OverlapRatio < 0.0 || opts.OverlapRatio >= 0.5 {
		return &ChunkingError{
			Type:    "validation",
			Message: "overlap ratio must be between 0.0 and 0.5",
			Details: map[string]interface{}{"overlap_ratio": opts.OverlapRatio},
		}
	}
	if opts.MinChunkSize < 0 {
		return &ChunkingError{
			Type:    "validation",
			Message: "minimum chunk size cannot be negative",
			Details: map[string]interface{}{"min_chunk_size": opts.MinChunkSize},
		}
	}
	if opts.MinChunkSize >= opts.MaxCharBuffer {
		return &ChunkingError{
			Type:    "validation",
			Message: "minimum chunk size must be less than maximum chunk buffer",
			Details: map[string]interface{}{
				"min_chunk_size":  opts.MinChunkSize,
				"max_char_buffer": opts.MaxCharBuffer,
			},
		}
	}
	return nil
}

// Length returns the number of characters in the chunk.
func (tc TextChunk) Length() int {
	return len(tc.Text)
}

// IsEmpty returns true if the chunk contains no text.
func (tc TextChunk) IsEmpty() bool {
	return len(tc.Text) == 0
}

// EstimateWords returns an approximate word count for the chunk.
func (tc TextChunk) EstimateWords() int {
	if tc.Metadata.WordCount > 0 {
		return tc.Metadata.WordCount
	}
	// Simple word estimation by splitting on whitespace
	return len(strings.Fields(tc.Text))
}

// HasPosition returns true if the chunk has position information.
func (tc TextChunk) HasPosition() bool {
	return tc.CharInterval != nil
}

// String returns a string representation of the chunk.
func (tc TextChunk) String() string {
	if tc.CharInterval != nil {
		return fmt.Sprintf("TextChunk{id=%s, pos=%s, len=%d}", 
			tc.ID, tc.CharInterval.String(), tc.Length())
	}
	return fmt.Sprintf("TextChunk{id=%s, len=%d}", tc.ID, tc.Length())
}