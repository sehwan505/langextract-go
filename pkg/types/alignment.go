package types

import (
	"fmt"
	"strings"
)

// AlignmentStatus represents the quality of alignment between extracted text
// and the source text from which it was extracted.
type AlignmentStatus int

const (
	// AlignmentNone indicates no alignment could be found
	AlignmentNone AlignmentStatus = iota
	
	// AlignmentExact indicates the extracted text matches exactly
	AlignmentExact
	
	// AlignmentFuzzy indicates a close but not exact match
	AlignmentFuzzy
	
	// AlignmentSemantic indicates a semantic match but different surface form
	AlignmentSemantic
	
	// AlignmentPartial indicates only part of the text could be aligned
	AlignmentPartial
	
	// AlignmentApproximate indicates a rough approximation of the location
	AlignmentApproximate
)

// String returns the string representation of the alignment status.
func (as AlignmentStatus) String() string {
	switch as {
	case AlignmentNone:
		return "none"
	case AlignmentExact:
		return "exact"
	case AlignmentFuzzy:
		return "fuzzy"
	case AlignmentSemantic:
		return "semantic"
	case AlignmentPartial:
		return "partial"
	case AlignmentApproximate:
		return "approximate"
	default:
		return fmt.Sprintf("unknown(%d)", int(as))
	}
}

// IsValid returns true if the alignment status is a known value.
func (as AlignmentStatus) IsValid() bool {
	return as >= AlignmentNone && as <= AlignmentApproximate
}

// Quality returns a numeric score (0-100) representing the quality of alignment.
// Higher scores indicate better alignment quality.
func (as AlignmentStatus) Quality() int {
	switch as {
	case AlignmentExact:
		return 100
	case AlignmentFuzzy:
		return 80
	case AlignmentSemantic:
		return 60
	case AlignmentPartial:
		return 40
	case AlignmentApproximate:
		return 20
	case AlignmentNone:
		return 0
	default:
		return 0
	}
}

// ParseAlignmentStatus parses a string into an AlignmentStatus.
func ParseAlignmentStatus(s string) (AlignmentStatus, error) {
	switch strings.ToLower(s) {
	case "none":
		return AlignmentNone, nil
	case "exact":
		return AlignmentExact, nil
	case "fuzzy":
		return AlignmentFuzzy, nil
	case "semantic":
		return AlignmentSemantic, nil
	case "partial":
		return AlignmentPartial, nil
	case "approximate":
		return AlignmentApproximate, nil
	default:
		return AlignmentNone, fmt.Errorf("invalid alignment status: %s", s)
	}
}

// AlignmentResult represents the result of an alignment operation.
type AlignmentResult struct {
	Status     AlignmentStatus // Quality of the alignment
	Confidence float64         // Confidence score (0.0-1.0)
	Score      float64         // Alignment score (algorithm-specific)
	Method     string          // Name of the alignment method used
}

// NewAlignmentResult creates a new alignment result with validation.
func NewAlignmentResult(status AlignmentStatus, confidence, score float64, method string) (*AlignmentResult, error) {
	if !status.IsValid() {
		return nil, fmt.Errorf("invalid alignment status: %v", status)
	}
	if confidence < 0.0 || confidence > 1.0 {
		return nil, fmt.Errorf("confidence must be between 0.0 and 1.0, got %f", confidence)
	}
	if method == "" {
		return nil, fmt.Errorf("alignment method cannot be empty")
	}
	
	return &AlignmentResult{
		Status:     status,
		Confidence: confidence,
		Score:      score,
		Method:     method,
	}, nil
}

// IsGoodAlignment returns true if this represents a high-quality alignment.
func (ar AlignmentResult) IsGoodAlignment() bool {
	return ar.Status.Quality() >= 60 && ar.Confidence >= 0.7
}

// String returns a string representation of the alignment result.
func (ar AlignmentResult) String() string {
	return fmt.Sprintf("AlignmentResult{status=%s, confidence=%.2f, score=%.2f, method=%s}",
		ar.Status, ar.Confidence, ar.Score, ar.Method)
}