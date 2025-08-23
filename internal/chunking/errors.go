package chunking

import (
	"fmt"
	"strings"
)

// ChunkingError represents errors that occur during text chunking operations.
type ChunkingError struct {
	Type     string                 // Error type (e.g., "validation", "processing", "memory")
	Message  string                 // Human-readable error message
	Details  map[string]interface{} // Additional error context
	Cause    error                  // Underlying error if any
	Position int                    // Character position where error occurred (-1 if not applicable)
}

// Error implements the error interface.
func (e *ChunkingError) Error() string {
	var parts []string
	
	if e.Type != "" {
		parts = append(parts, fmt.Sprintf("[%s]", e.Type))
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
func (e *ChunkingError) Unwrap() error {
	return e.Cause
}

// IsType checks if the error is of a specific type.
func (e *ChunkingError) IsType(errorType string) bool {
	return e.Type == errorType
}

// NewChunkingError creates a new chunking error.
func NewChunkingError(errorType, message string) *ChunkingError {
	return &ChunkingError{
		Type:     errorType,
		Message:  message,
		Details:  make(map[string]interface{}),
		Position: -1,
	}
}

// NewChunkingErrorWithCause creates a new chunking error wrapping another error.
func NewChunkingErrorWithCause(errorType, message string, cause error) *ChunkingError {
	return &ChunkingError{
		Type:     errorType,
		Message:  message,
		Details:  make(map[string]interface{}),
		Cause:    cause,
		Position: -1,
	}
}

// WithPosition adds position information to the error.
func (e *ChunkingError) WithPosition(pos int) *ChunkingError {
	e.Position = pos
	return e
}

// WithDetail adds a detail key-value pair to the error.
func (e *ChunkingError) WithDetail(key string, value interface{}) *ChunkingError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// Common error types
const (
	ErrorTypeValidation = "validation"
	ErrorTypeProcessing = "processing"
	ErrorTypeMemory     = "memory"
	ErrorTypeTimeout    = "timeout"
	ErrorTypeInternal   = "internal"
)

// Common chunking errors
var (
	ErrInvalidOptions = NewChunkingError(ErrorTypeValidation, "invalid chunking options")
	ErrEmptyText      = NewChunkingError(ErrorTypeValidation, "text cannot be empty")
	ErrContextExpired = NewChunkingError(ErrorTypeTimeout, "context expired during chunking")
	ErrMemoryLimit    = NewChunkingError(ErrorTypeMemory, "memory limit exceeded during chunking")
)