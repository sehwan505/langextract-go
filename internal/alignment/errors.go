package alignment

import (
	"fmt"
	"strings"
)

// AlignmentError represents errors that occur during text alignment operations.
type AlignmentError struct {
	Type     string                 // Error type (e.g., "validation", "processing", "timeout")
	Message  string                 // Human-readable error message
	Details  map[string]interface{} // Additional error context
	Cause    error                  // Underlying error if any
	Position int                    // Character position where error occurred (-1 if not applicable)
	Method   string                 // Alignment method that caused the error
}

// Error implements the error interface.
func (e *AlignmentError) Error() string {
	var parts []string
	
	if e.Type != "" {
		parts = append(parts, fmt.Sprintf("[%s]", e.Type))
	}
	
	if e.Method != "" {
		parts = append(parts, fmt.Sprintf("(%s)", e.Method))
	}
	
	parts = append(parts, e.Message)
	
	if e.Position >= 0 {
		parts = append(parts, fmt.Sprintf("at position %d", e.Position))
	}
	
	if len(e.Details) > 0 {
		var details []string
		for k, v := range e.Details {
			details = append(details, fmt.Sprintf("%s=%v", k, v))
		}
		parts = append(parts, fmt.Sprintf("details: %s", strings.Join(details, ", ")))
	}
	
	return strings.Join(parts, " ")
}

// Unwrap returns the underlying error for error wrapping support.
func (e *AlignmentError) Unwrap() error {
	return e.Cause
}

// IsType checks if the error is of a specific type.
func (e *AlignmentError) IsType(errorType string) bool {
	return e.Type == errorType
}

// NewAlignmentError creates a new alignment error.
func NewAlignmentError(errorType, message string) *AlignmentError {
	return &AlignmentError{
		Type:     errorType,
		Message:  message,
		Details:  make(map[string]interface{}),
		Position: -1,
	}
}

// NewAlignmentErrorWithCause creates a new alignment error wrapping another error.
func NewAlignmentErrorWithCause(errorType, message string, cause error) *AlignmentError {
	return &AlignmentError{
		Type:     errorType,
		Message:  message,
		Details:  make(map[string]interface{}),
		Cause:    cause,
		Position: -1,
	}
}

// WithPosition adds position information to the error.
func (e *AlignmentError) WithPosition(pos int) *AlignmentError {
	e.Position = pos
	return e
}

// WithMethod adds the alignment method information to the error.
func (e *AlignmentError) WithMethod(method string) *AlignmentError {
	e.Method = method
	return e
}

// WithDetail adds a detail key-value pair to the error.
func (e *AlignmentError) WithDetail(key string, value interface{}) *AlignmentError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// Common error types
const (
	ErrorTypeValidation    = "validation"
	ErrorTypeProcessing    = "processing"
	ErrorTypeTimeout       = "timeout"
	ErrorTypeNotFound      = "not_found"
	ErrorTypeAmbiguous     = "ambiguous"
	ErrorTypeLowConfidence = "low_confidence"
	ErrorTypeInternal      = "internal"
)

// Common alignment errors
var (
	ErrInvalidOptions     = NewAlignmentError(ErrorTypeValidation, "invalid alignment options")
	ErrEmptyExtracted     = NewAlignmentError(ErrorTypeValidation, "extracted text cannot be empty")
	ErrEmptySource        = NewAlignmentError(ErrorTypeValidation, "source text cannot be empty")
	ErrNoAlignment        = NewAlignmentError(ErrorTypeNotFound, "no valid alignment found")
	ErrAmbiguousAlignment = NewAlignmentError(ErrorTypeAmbiguous, "multiple equally valid alignments found")
	ErrLowConfidence      = NewAlignmentError(ErrorTypeLowConfidence, "alignment confidence below threshold")
	ErrContextExpired     = NewAlignmentError(ErrorTypeTimeout, "context expired during alignment")
)