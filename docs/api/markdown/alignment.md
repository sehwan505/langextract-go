# alignment

Package: `github.com/sehwan505/langextract-go/internal/alignment`

```go
package alignment // import "github.com/sehwan505/langextract-go/internal/alignment"


CONSTANTS

const (
	ErrorTypeValidation    = "validation"
	ErrorTypeProcessing    = "processing"
	ErrorTypeTimeout       = "timeout"
	ErrorTypeNotFound      = "not_found"
	ErrorTypeAmbiguous     = "ambiguous"
	ErrorTypeLowConfidence = "low_confidence"
	ErrorTypeInternal      = "internal"
)
    Common error types


VARIABLES

var (
	ErrInvalidOptions     = NewAlignmentError(ErrorTypeValidation, "invalid alignment options")
	ErrEmptyExtracted     = NewAlignmentError(ErrorTypeValidation, "extracted text cannot be empty")
	ErrEmptySource        = NewAlignmentError(ErrorTypeValidation, "source text cannot be empty")
	ErrNoAlignment        = NewAlignmentError(ErrorTypeNotFound, "no valid alignment found")
	ErrAmbiguousAlignment = NewAlignmentError(ErrorTypeAmbiguous, "multiple equally valid alignments found")
	ErrLowConfidence      = NewAlignmentError(ErrorTypeLowConfidence, "alignment confidence below threshold")
	ErrContextExpired     = NewAlignmentError(ErrorTypeTimeout, "context expired during alignment")
)
    Common alignment errors


TYPES

type AlignmentCandidate struct {
	// Position is the character position in the source text
	Position int

	// Length is the length of the aligned text
	Length int

	// Score is the alignment quality score
	Score float64

	// Method is the alignment method that found this candidate
	Method string

	// MatchedText is the text that was matched at this position
	MatchedText string

	// EditDistance is the edit distance for fuzzy matches
	EditDistance int

	// Properties contains method-specific properties
	Properties map[string]interface{}
}
    AlignmentCandidate represents a potential alignment location.

type AlignmentError struct {
	Type     string                 // Error type (e.g., "validation", "processing", "timeout")
	Message  string                 // Human-readable error message
	Details  map[string]interface{} // Additional error context
	Cause    error                  // Underlying error if any
	Position int                    // Character position where error occurred (-1 if not applicable)
	Method   string                 // Alignment method that caused the error
}
    AlignmentError represents errors that occur during text alignment
    operations.

func NewAlignmentError(errorType, message string) *AlignmentError
    NewAlignmentError creates a new alignment error.

func NewAlignmentErrorWithCause(errorType, message string, cause error) *AlignmentError
    NewAlignmentErrorWithCause creates a new alignment error wrapping another
    error.

func (e *AlignmentError) Error() string
    Error implements the error interface.

func (e *AlignmentError) IsType(errorType string) bool
    IsType checks if the error is of a specific type.

func (e *AlignmentError) Unwrap() error
    Unwrap returns the underlying error for error wrapping support.

func (e *AlignmentError) WithDetail(key string, value interface{}) *AlignmentError
    WithDetail adds a detail key-value pair to the error.

func (e *AlignmentError) WithMethod(method string) *AlignmentError
    WithMethod adds the alignment method information to the error.

func (e *AlignmentError) WithPosition(pos int) *AlignmentError
    WithPosition adds position information to the error.

type AlignmentMetadata struct {
	// AlgorithmName is the name of the algorithm used for alignment
	AlgorithmName string

	// ProcessingTime is the time taken to perform the alignment
	ProcessingTimeMs int64

	// AttemptedMethods lists all alignment methods that were tried
	AttemptedMethods []string

	// ErrorCount is the number of errors encountered during alignment
	ErrorCount int

	// FallbackUsed indicates if a fallback method was used
	FallbackUsed bool

	// Properties contains algorithm-specific metadata
	Properties map[string]interface{}
}
    AlignmentMetadata contains additional information about the alignment
    process.

type AlignmentOptions struct {
	// CaseSensitive determines if the alignment should be case-sensitive
	CaseSensitive bool

	// IgnoreWhitespace determines if whitespace differences should be ignored
	IgnoreWhitespace bool

	// IgnorePunctuation determines if punctuation differences should be ignored
	IgnorePunctuation bool

	// MaxDistance is the maximum edit distance allowed for fuzzy matching
	MaxDistance int

	// MinConfidence is the minimum confidence score required for valid alignment
	MinConfidence float64

	// MaxCandidates is the maximum number of candidate alignments to consider
	MaxCandidates int

	// WindowSize is the size of the context window for semantic alignment
	WindowSize int

	// AllowPartialMatches determines if partial matches are acceptable
	AllowPartialMatches bool

	// PreferExactMatches gives higher priority to exact matches
	PreferExactMatches bool

	// TimeoutMs is the maximum time allowed for alignment in milliseconds
	TimeoutMs int64

	// Properties contains algorithm-specific options
	Properties map[string]interface{}
}
    AlignmentOptions configures the behavior of text alignment algorithms.

func DefaultAlignmentOptions() AlignmentOptions
    DefaultAlignmentOptions returns the default options for text alignment.

func (opts AlignmentOptions) Validate() error
    Validate checks if the alignment options are valid.

func (opts AlignmentOptions) WithCaseSensitive(sensitive bool) AlignmentOptions
    WithCaseSensitive sets the case sensitivity option.

func (opts AlignmentOptions) WithIgnoreWhitespace(ignore bool) AlignmentOptions
    WithIgnoreWhitespace sets the whitespace handling option.

func (opts AlignmentOptions) WithMaxDistance(distance int) AlignmentOptions
    WithMaxDistance sets the maximum edit distance for fuzzy matching.

func (opts AlignmentOptions) WithMinConfidence(confidence float64) AlignmentOptions
    WithMinConfidence sets the minimum confidence threshold.

func (opts AlignmentOptions) WithTimeout(timeoutMs int64) AlignmentOptions
    WithTimeout sets the alignment timeout.

type AlignmentResult struct {
	// CharInterval represents the location of the aligned text in the source
	CharInterval *types.CharInterval

	// AlignmentResult contains the quality and metadata of the alignment
	AlignmentInfo *types.AlignmentResult

	// ExtractedText is the original extracted text that was aligned
	ExtractedText string

	// AlignedText is the actual text found at the aligned position
	AlignedText string

	// Confidence is the confidence score of this alignment (0.0-1.0)
	Confidence float64

	// Metadata contains additional algorithm-specific information
	Metadata AlignmentMetadata
}
    AlignmentResult represents the result of a text alignment operation.

func NewAlignmentResult(interval *types.CharInterval, alignmentInfo *types.AlignmentResult, extracted, aligned string, confidence float64) (*AlignmentResult, error)
    NewAlignmentResult creates a new alignment result with validation.

func (ar AlignmentResult) GetScore() float64
    GetScore returns the alignment score, preferring the AlignmentInfo score.

func (ar AlignmentResult) IsValid() bool
    IsValid returns true if the alignment result is considered valid.

func (ar AlignmentResult) String() string
    String returns a string representation of the alignment result.

type AlignmentStrategy int
    AlignmentStrategy defines different strategies for handling multiple
    alignment results.

const (
	// StrategyBestScore selects the alignment with the highest score
	StrategyBestScore AlignmentStrategy = iota

	// StrategyFirstFound selects the first valid alignment found
	StrategyFirstFound

	// StrategyMostConfident selects the alignment with highest confidence
	StrategyMostConfident

	// StrategyExactPreferred prefers exact matches over fuzzy matches
	StrategyExactPreferred

	// StrategyPositionBased prefers alignments based on position heuristics
	StrategyPositionBased
)
func (as AlignmentStrategy) String() string
    String returns the string representation of the alignment strategy.

type DefaultMultiAligner struct {
	// Has unexported fields.
}
    DefaultMultiAligner implements the MultiAligner interface using multiple
    alignment algorithms with priority-based selection.

func NewMultiAligner() *DefaultMultiAligner
    NewMultiAligner creates a new MultiAligner with default aligners.

func (ma *DefaultMultiAligner) AlignWithAllMethods(ctx context.Context, extracted, source string, opts AlignmentOptions) ([]AlignmentResult, error)
    AlignWithAllMethods tries all registered methods and returns all results.

func (ma *DefaultMultiAligner) AlignWithBestMethod(ctx context.Context, extracted, source string, opts AlignmentOptions) (*types.CharInterval, *types.AlignmentResult, error)
    AlignWithBestMethod tries multiple alignment methods and returns the best
    result.

func (ma *DefaultMultiAligner) AlignWithStrategy(ctx context.Context, extracted, source string, opts AlignmentOptions, strategy AlignmentStrategy) (*types.CharInterval, *types.AlignmentResult, error)
    AlignWithStrategy aligns using a specific strategy for result selection.

func (ma *DefaultMultiAligner) ClearAligners()
    ClearAligners removes all registered aligners.

func (ma *DefaultMultiAligner) GetAlignerPriority(name string) (int, bool)
    GetAlignerPriority returns the priority of a registered aligner.

func (ma *DefaultMultiAligner) GetAvailableAligners() []string
    GetAvailableAligners returns the list of registered alignment algorithms.

func (ma *DefaultMultiAligner) GetStats() map[string]interface{}
    GetStats returns statistics about the registered aligners.

func (ma *DefaultMultiAligner) RegisterAligner(aligner TextAligner, priority int) error
    RegisterAligner adds a new alignment algorithm to the pool.

func (ma *DefaultMultiAligner) RemoveAligner(name string) bool
    RemoveAligner removes an aligner by name.

func (ma *DefaultMultiAligner) SetAlignerPriority(name string, priority int) bool
    SetAlignerPriority updates the priority of a registered aligner.

type ExactMatcher struct{}
    ExactMatcher implements exact string matching for text alignment.
    It finds exact matches between extracted text and source text with optional
    normalization for case and whitespace.

func NewExactMatcher() *ExactMatcher
    NewExactMatcher creates a new ExactMatcher.

func (em *ExactMatcher) AlignExtraction(ctx context.Context, extracted, source string, opts AlignmentOptions) (*types.CharInterval, *types.AlignmentResult, error)
    AlignExtraction finds exact matches of extracted text in the source text.

func (em *ExactMatcher) AlignExtractions(ctx context.Context, extractions []string, source string, opts AlignmentOptions) ([]AlignmentResult, error)
    AlignExtractions aligns multiple extractions in batch.

func (em *ExactMatcher) FindBestAlignment(ctx context.Context, extracted, source string, opts AlignmentOptions) (*types.CharInterval, *types.AlignmentResult, error)
    FindBestAlignment finds the best exact match using multiple strategies.

func (em *ExactMatcher) Name() string
    Name returns the name of this alignment algorithm.

func (em *ExactMatcher) ValidateAlignment(extracted, source string, interval *types.CharInterval) (float64, error)
    ValidateAlignment checks if a proposed alignment is valid for exact
    matching.

type FuzzyMatcher struct {
	// Has unexported fields.
}
    FuzzyMatcher implements fuzzy string matching using Levenshtein distance for
    text alignment. It can find approximate matches when exact matching fails.

func NewFuzzyMatcher() *FuzzyMatcher
    NewFuzzyMatcher creates a new FuzzyMatcher with default maximum edit
    distance.

func NewFuzzyMatcherWithDistance(maxDistance int) *FuzzyMatcher
    NewFuzzyMatcherWithDistance creates a new FuzzyMatcher with specified
    maximum edit distance.

func (fm *FuzzyMatcher) AlignExtraction(ctx context.Context, extracted, source string, opts AlignmentOptions) (*types.CharInterval, *types.AlignmentResult, error)
    AlignExtraction finds fuzzy matches of extracted text in the source text.

func (fm *FuzzyMatcher) AlignExtractions(ctx context.Context, extractions []string, source string, opts AlignmentOptions) ([]AlignmentResult, error)
    AlignExtractions aligns multiple extractions using fuzzy matching.

func (fm *FuzzyMatcher) FindBestAlignment(ctx context.Context, extracted, source string, opts AlignmentOptions) (*types.CharInterval, *types.AlignmentResult, error)
    FindBestAlignment finds the best fuzzy match using multiple strategies.

func (fm *FuzzyMatcher) Name() string
    Name returns the name of this alignment algorithm.

func (fm *FuzzyMatcher) ValidateAlignment(extracted, source string, interval *types.CharInterval) (float64, error)
    ValidateAlignment validates a proposed alignment using fuzzy matching.

type MultiAligner interface {
	// RegisterAligner adds a new alignment algorithm to the pool
	RegisterAligner(aligner TextAligner, priority int) error

	// AlignWithBestMethod tries multiple alignment methods and returns the best result
	AlignWithBestMethod(ctx context.Context, extracted, source string, opts AlignmentOptions) (*types.CharInterval, *types.AlignmentResult, error)

	// AlignWithAllMethods tries all registered methods and returns all results
	AlignWithAllMethods(ctx context.Context, extracted, source string, opts AlignmentOptions) ([]AlignmentResult, error)

	// GetAvailableAligners returns the list of registered alignment algorithms
	GetAvailableAligners() []string
}
    MultiAligner provides a unified interface for using multiple alignment
    algorithms.

type TextAligner interface {
	// AlignExtraction aligns extracted text with the source text and returns
	// the character interval where the extraction is located.
	AlignExtraction(ctx context.Context, extracted, source string, opts AlignmentOptions) (*types.CharInterval, *types.AlignmentResult, error)

	// AlignExtractions aligns multiple extractions with the source text in batch.
	// This can be more efficient than individual alignments for some algorithms.
	AlignExtractions(ctx context.Context, extractions []string, source string, opts AlignmentOptions) ([]AlignmentResult, error)

	// FindBestAlignment finds the best possible alignment for extracted text
	// by trying multiple strategies and returning the highest quality result.
	FindBestAlignment(ctx context.Context, extracted, source string, opts AlignmentOptions) (*types.CharInterval, *types.AlignmentResult, error)

	// ValidateAlignment checks if a proposed alignment is valid and returns
	// a confidence score for the alignment quality.
	ValidateAlignment(extracted, source string, interval *types.CharInterval) (float64, error)

	// Name returns the name of the alignment algorithm for logging and debugging.
	Name() string
}
    TextAligner defines the interface for aligning extracted text with source
    text. Different implementations provide various alignment strategies such as
    exact matching, fuzzy matching, and semantic alignment.

```
