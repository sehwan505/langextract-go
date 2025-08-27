# chunking

Package: `github.com/sehwan505/langextract-go/internal/chunking`

```go
package chunking // import "github.com/sehwan505/langextract-go/internal/chunking"


CONSTANTS

const (
	ErrorTypeValidation = "validation"
	ErrorTypeProcessing = "processing"
	ErrorTypeMemory     = "memory"
	ErrorTypeTimeout    = "timeout"
	ErrorTypeInternal   = "internal"
)
    Common error types


VARIABLES

var (
	ErrInvalidOptions = NewChunkingError(ErrorTypeValidation, "invalid chunking options")
	ErrEmptyText      = NewChunkingError(ErrorTypeValidation, "text cannot be empty")
	ErrContextExpired = NewChunkingError(ErrorTypeTimeout, "context expired during chunking")
	ErrMemoryLimit    = NewChunkingError(ErrorTypeMemory, "memory limit exceeded during chunking")
)
    Common chunking errors


TYPES

type AdaptiveChunker struct {
	// Has unexported fields.
}
    AdaptiveChunker implements dynamic chunk sizing based on content complexity,
    density, and structural characteristics.

func NewAdaptiveChunker() *AdaptiveChunker
    NewAdaptiveChunker creates a new AdaptiveChunker with default complexity
    thresholds.

func (ac *AdaptiveChunker) ChunkDocument(ctx context.Context, doc *document.Document, opts ChunkingOptions) ([]TextChunk, error)
    ChunkDocument splits a document into adaptively-sized chunks.

func (ac *AdaptiveChunker) ChunkText(ctx context.Context, text string, opts ChunkingOptions) ([]TextChunk, error)
    ChunkText splits raw text into adaptively-sized chunks.

func (ac *AdaptiveChunker) EstimateChunks(text string, opts ChunkingOptions) int
    EstimateChunks provides an estimate considering adaptive sizing.

func (ac *AdaptiveChunker) Name() string
    Name returns the name of this chunking algorithm.

func (ac *AdaptiveChunker) WithComplexityThresholds(thresholds ContentComplexityThresholds) *AdaptiveChunker
    WithComplexityThresholds allows customization of complexity thresholds.

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
    ChunkMetadata contains additional information about a text chunk.

type ChunkingError struct {
	Type     string                 // Error type (e.g., "validation", "processing", "memory")
	Message  string                 // Human-readable error message
	Details  map[string]interface{} // Additional error context
	Cause    error                  // Underlying error if any
	Position int                    // Character position where error occurred (-1 if not applicable)
}
    ChunkingError represents errors that occur during text chunking operations.

func NewChunkingError(errorType, message string) *ChunkingError
    NewChunkingError creates a new chunking error.

func NewChunkingErrorWithCause(errorType, message string, cause error) *ChunkingError
    NewChunkingErrorWithCause creates a new chunking error wrapping another
    error.

func (e *ChunkingError) Error() string
    Error implements the error interface.

func (e *ChunkingError) IsType(errorType string) bool
    IsType checks if the error is of a specific type.

func (e *ChunkingError) Unwrap() error
    Unwrap returns the underlying error for error wrapping support.

func (e *ChunkingError) WithDetail(key string, value interface{}) *ChunkingError
    WithDetail adds a detail key-value pair to the error.

func (e *ChunkingError) WithPosition(pos int) *ChunkingError
    WithPosition adds position information to the error.

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
    ChunkingOptions configures the behavior of text chunking algorithms.

func DefaultChunkingOptions() ChunkingOptions
    DefaultChunkingOptions returns the default options for text chunking.

func (opts ChunkingOptions) Validate() error
    Validate checks if the chunking options are valid.

func (opts ChunkingOptions) WithLanguage(language string) ChunkingOptions
    WithLanguage sets a language hint for chunking.

func (opts ChunkingOptions) WithMaxCharBuffer(size int) ChunkingOptions
    WithMaxCharBuffer sets the maximum character buffer size.

func (opts ChunkingOptions) WithOverlap(ratio float64) ChunkingOptions
    WithOverlap sets the overlap ratio between chunks.

func (opts ChunkingOptions) WithSentencePreservation(preserve bool) ChunkingOptions
    WithSentencePreservation enables or disables sentence preservation.

type ContentComplexityThresholds struct {
	LowComplexity    float64 // Below this = simple content (larger chunks)
	MediumComplexity float64 // Between low and medium = moderate content
	HighComplexity   float64 // Above medium = complex content (smaller chunks)
}
    ContentComplexityThresholds defines thresholds for different complexity
    levels.

type ContentMetrics struct {
	AverageWordsPerSentence float64
	AverageSyllablesPerWord float64
	UniqueWordRatio         float64
	PunctuationDensity      float64
	NumberDensity           float64
	StructuralComplexity    float64
	OverallComplexity       float64
}
    ContentMetrics holds various metrics used to assess content complexity.

type SemanticChunker struct {
	// Has unexported fields.
}
    SemanticChunker implements context-aware text chunking that attempts to
    preserve semantic boundaries and topical coherence.

func NewSemanticChunker() *SemanticChunker
    NewSemanticChunker creates a new SemanticChunker with enhanced boundary
    detection.

func (sc *SemanticChunker) ChunkDocument(ctx context.Context, doc *document.Document, opts ChunkingOptions) ([]TextChunk, error)
    ChunkDocument splits a document into semantically coherent chunks.

func (sc *SemanticChunker) ChunkText(ctx context.Context, text string, opts ChunkingOptions) ([]TextChunk, error)
    ChunkText splits raw text into semantically coherent chunks.

func (sc *SemanticChunker) EstimateChunks(text string, opts ChunkingOptions) int
    EstimateChunks provides an estimate considering semantic boundaries.

func (sc *SemanticChunker) Name() string
    Name returns the name of this chunking algorithm.

type SimpleChunker struct {
	// Has unexported fields.
}
    SimpleChunker implements basic sentence and paragraph-based text chunking.
    It attempts to preserve sentence boundaries while respecting character
    limits.

func NewSimpleChunker() *SimpleChunker
    NewSimpleChunker creates a new SimpleChunker with default sentence
    detection.

func (sc *SimpleChunker) ChunkDocument(ctx context.Context, doc *document.Document, opts ChunkingOptions) ([]TextChunk, error)
    ChunkDocument splits a document into chunks based on the specified options.

func (sc *SimpleChunker) ChunkText(ctx context.Context, text string, opts ChunkingOptions) ([]TextChunk, error)
    ChunkText splits raw text into chunks based on the specified options.

func (sc *SimpleChunker) EstimateChunks(text string, opts ChunkingOptions) int
    EstimateChunks provides an estimate of the number of chunks that will be
    created.

func (sc *SimpleChunker) Name() string
    Name returns the name of this chunking algorithm.

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
    TextChunk represents a chunk of text with position and metadata information.

func (tc TextChunk) EstimateWords() int
    EstimateWords returns an approximate word count for the chunk.

func (tc TextChunk) HasPosition() bool
    HasPosition returns true if the chunk has position information.

func (tc TextChunk) IsEmpty() bool
    IsEmpty returns true if the chunk contains no text.

func (tc TextChunk) Length() int
    Length returns the number of characters in the chunk.

func (tc TextChunk) String() string
    String returns a string representation of the chunk.

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
    TextChunker defines the interface for text chunking algorithms.
    Different implementations can provide various chunking strategies such as
    sentence-based, paragraph-based, or semantic chunking.

type TextSegment struct {
	Text     string
	StartPos int
	EndPos   int
}
    TextSegment represents a segment of text with position information.

type TextStructureAnalysis struct {
	SentenceBoundaries  []int              // Character positions of sentence boundaries
	ParagraphBoundaries []int              // Character positions of paragraph boundaries
	TopicTransitions    []int              // Character positions of likely topic transitions
	SectionHeaders      []int              // Character positions of section headers
	ListItems           []int              // Character positions of list items
	KeywordDensity      map[string]float64 // Density of important keywords
}
    TextStructureAnalysis contains information about the structure of the text.

```
