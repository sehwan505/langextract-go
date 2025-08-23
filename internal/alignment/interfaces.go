package alignment

import (
	"context"
	"fmt"

	"github.com/sehwan505/langextract-go/pkg/types"
)

// TextAligner defines the interface for aligning extracted text with source text.
// Different implementations provide various alignment strategies such as exact matching,
// fuzzy matching, and semantic alignment.
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

// AlignmentResult represents the result of a text alignment operation.
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

// AlignmentMetadata contains additional information about the alignment process.
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

// AlignmentOptions configures the behavior of text alignment algorithms.
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

// DefaultAlignmentOptions returns the default options for text alignment.
func DefaultAlignmentOptions() AlignmentOptions {
	return AlignmentOptions{
		CaseSensitive:       false,
		IgnoreWhitespace:    true,
		IgnorePunctuation:   false,
		MaxDistance:         5,
		MinConfidence:       0.7,
		MaxCandidates:       10,
		WindowSize:          100,
		AllowPartialMatches: true,
		PreferExactMatches:  true,
		TimeoutMs:           5000,
		Properties:          make(map[string]interface{}),
	}
}

// WithCaseSensitive sets the case sensitivity option.
func (opts AlignmentOptions) WithCaseSensitive(sensitive bool) AlignmentOptions {
	opts.CaseSensitive = sensitive
	return opts
}

// WithIgnoreWhitespace sets the whitespace handling option.
func (opts AlignmentOptions) WithIgnoreWhitespace(ignore bool) AlignmentOptions {
	opts.IgnoreWhitespace = ignore
	return opts
}

// WithMaxDistance sets the maximum edit distance for fuzzy matching.
func (opts AlignmentOptions) WithMaxDistance(distance int) AlignmentOptions {
	opts.MaxDistance = distance
	return opts
}

// WithMinConfidence sets the minimum confidence threshold.
func (opts AlignmentOptions) WithMinConfidence(confidence float64) AlignmentOptions {
	opts.MinConfidence = confidence
	return opts
}

// WithTimeout sets the alignment timeout.
func (opts AlignmentOptions) WithTimeout(timeoutMs int64) AlignmentOptions {
	opts.TimeoutMs = timeoutMs
	return opts
}

// Validate checks if the alignment options are valid.
func (opts AlignmentOptions) Validate() error {
	if opts.MaxDistance < 0 {
		return &AlignmentError{
			Type:    "validation",
			Message: "max distance cannot be negative",
			Details: map[string]interface{}{"max_distance": opts.MaxDistance},
		}
	}
	if opts.MinConfidence < 0.0 || opts.MinConfidence > 1.0 {
		return &AlignmentError{
			Type:    "validation",
			Message: "min confidence must be between 0.0 and 1.0",
			Details: map[string]interface{}{"min_confidence": opts.MinConfidence},
		}
	}
	if opts.MaxCandidates <= 0 {
		return &AlignmentError{
			Type:    "validation",
			Message: "max candidates must be positive",
			Details: map[string]interface{}{"max_candidates": opts.MaxCandidates},
		}
	}
	if opts.WindowSize < 0 {
		return &AlignmentError{
			Type:    "validation",
			Message: "window size cannot be negative",
			Details: map[string]interface{}{"window_size": opts.WindowSize},
		}
	}
	if opts.TimeoutMs <= 0 {
		return &AlignmentError{
			Type:    "validation",
			Message: "timeout must be positive",
			Details: map[string]interface{}{"timeout_ms": opts.TimeoutMs},
		}
	}
	return nil
}

// MultiAligner provides a unified interface for using multiple alignment algorithms.
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

// AlignmentCandidate represents a potential alignment location.
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

// AlignmentStrategy defines different strategies for handling multiple alignment results.
type AlignmentStrategy int

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

// String returns the string representation of the alignment strategy.
func (as AlignmentStrategy) String() string {
	switch as {
	case StrategyBestScore:
		return "best_score"
	case StrategyFirstFound:
		return "first_found"
	case StrategyMostConfident:
		return "most_confident"
	case StrategyExactPreferred:
		return "exact_preferred"
	case StrategyPositionBased:
		return "position_based"
	default:
		return "unknown"
	}
}

// IsValid returns true if the alignment result is considered valid.
func (ar AlignmentResult) IsValid() bool {
	return ar.CharInterval != nil &&
		ar.AlignmentInfo != nil &&
		ar.Confidence >= 0.0 &&
		ar.AlignmentInfo.IsGoodAlignment()
}

// GetScore returns the alignment score, preferring the AlignmentInfo score.
func (ar AlignmentResult) GetScore() float64 {
	if ar.AlignmentInfo != nil {
		return ar.AlignmentInfo.Score
	}
	return ar.Confidence
}

// String returns a string representation of the alignment result.
func (ar AlignmentResult) String() string {
	if ar.CharInterval == nil {
		return "AlignmentResult{invalid}"
	}
	return fmt.Sprintf("AlignmentResult{pos=%s, confidence=%.2f, method=%s}",
		ar.CharInterval.String(), ar.Confidence, ar.Metadata.AlgorithmName)
}

// NewAlignmentResult creates a new alignment result with validation.
func NewAlignmentResult(interval *types.CharInterval, alignmentInfo *types.AlignmentResult, extracted, aligned string, confidence float64) (*AlignmentResult, error) {
	if interval == nil {
		return nil, &AlignmentError{
			Type:    "validation",
			Message: "character interval cannot be nil",
		}
	}
	if confidence < 0.0 || confidence > 1.0 {
		return nil, &AlignmentError{
			Type:    "validation",
			Message: "confidence must be between 0.0 and 1.0",
			Details: map[string]interface{}{"confidence": confidence},
		}
	}
	
	return &AlignmentResult{
		CharInterval:  interval,
		AlignmentInfo: alignmentInfo,
		ExtractedText: extracted,
		AlignedText:   aligned,
		Confidence:    confidence,
		Metadata: AlignmentMetadata{
			Properties: make(map[string]interface{}),
		},
	}, nil
}