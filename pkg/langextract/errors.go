package langextract

import (
	"fmt"
)

// ExtractError represents errors that occur during extraction operations.
type ExtractError struct {
	Op      string // Operation that caused the error
	Message string // Human-readable error message
	Err     error  // Underlying error, if any
}

// Error implements the error interface.
func (e *ExtractError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("langextract: %s: %s: %v", e.Op, e.Message, e.Err)
	}
	return fmt.Sprintf("langextract: %s: %s", e.Op, e.Message)
}

// Unwrap returns the underlying error.
func (e *ExtractError) Unwrap() error {
	return e.Err
}

// ValidationError represents validation errors in extraction configuration.
type ValidationError struct {
	Field   string // Field that failed validation
	Value   string // Invalid value
	Message string // Validation error message
}

// Error implements the error interface.
func (v *ValidationError) Error() string {
	return fmt.Sprintf("validation error: field %s with value %q: %s", v.Field, v.Value, v.Message)
}

// ProviderError represents errors from language model providers.
type ProviderError struct {
	Provider string // Provider name (openai, gemini, ollama)
	Status   string // HTTP status or error code
	Message  string // Provider error message
	Err      error  // Underlying error
}

// Error implements the error interface.
func (p *ProviderError) Error() string {
	if p.Err != nil {
		return fmt.Sprintf("provider %s error [%s]: %s: %v", p.Provider, p.Status, p.Message, p.Err)
	}
	return fmt.Sprintf("provider %s error [%s]: %s", p.Provider, p.Status, p.Message)
}

// Unwrap returns the underlying error.
func (p *ProviderError) Unwrap() error {
	return p.Err
}

// AlignmentError represents errors during text alignment operations.
type AlignmentError struct {
	ExtractedText string // Text that failed to align
	SourceText    string // Source text snippet
	Message       string // Alignment error message
}

// Error implements the error interface.
func (a *AlignmentError) Error() string {
	return fmt.Sprintf("alignment error: failed to align %q in source: %s", a.ExtractedText, a.Message)
}

// Error creation helpers

// NewExtractError creates a new ExtractError.
func NewExtractError(op, message string, err error) *ExtractError {
	return &ExtractError{
		Op:      op,
		Message: message,
		Err:     err,
	}
}

// NewValidationError creates a new ValidationError.
func NewValidationError(field, value, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Value:   value,
		Message: message,
	}
}

// NewProviderError creates a new ProviderError.
func NewProviderError(provider, status, message string, err error) *ProviderError {
	return &ProviderError{
		Provider: provider,
		Status:   status,
		Message:  message,
		Err:      err,
	}
}

// NewAlignmentError creates a new AlignmentError.
func NewAlignmentError(extractedText, sourceText, message string) *AlignmentError {
	return &AlignmentError{
		ExtractedText: extractedText,
		SourceText:    sourceText,
		Message:       message,
	}
}