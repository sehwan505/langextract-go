package prompt

import (
	"fmt"
)

// PromptError represents an error in prompt processing
type PromptError struct {
	Operation string
	Message   string
	Cause     error
}

// Error implements the error interface
func (e *PromptError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("prompt %s: %s: %v", e.Operation, e.Message, e.Cause)
	}
	return fmt.Sprintf("prompt %s: %s", e.Operation, e.Message)
}

// Unwrap returns the underlying cause
func (e *PromptError) Unwrap() error {
	return e.Cause
}

// NewPromptError creates a new prompt error
func NewPromptError(operation, message string, cause error) *PromptError {
	return &PromptError{
		Operation: operation,
		Message:   message,
		Cause:     cause,
	}
}

// TemplateError represents an error in template processing
type TemplateError struct {
	TemplateName string
	Line         int
	Column       int
	Message      string
	Cause        error
}

// Error implements the error interface
func (e *TemplateError) Error() string {
	location := ""
	if e.Line > 0 {
		if e.Column > 0 {
			location = fmt.Sprintf(" at line %d, column %d", e.Line, e.Column)
		} else {
			location = fmt.Sprintf(" at line %d", e.Line)
		}
	}
	
	if e.Cause != nil {
		return fmt.Sprintf("template '%s'%s: %s: %v", e.TemplateName, location, e.Message, e.Cause)
	}
	return fmt.Sprintf("template '%s'%s: %s", e.TemplateName, location, e.Message)
}

// Unwrap returns the underlying cause
func (e *TemplateError) Unwrap() error {
	return e.Cause
}

// NewTemplateError creates a new template error
func NewTemplateError(templateName, message string, cause error) *TemplateError {
	return &TemplateError{
		TemplateName: templateName,
		Message:      message,
		Cause:        cause,
	}
}

// NewTemplateErrorWithLocation creates a new template error with location
func NewTemplateErrorWithLocation(templateName, message string, line, column int, cause error) *TemplateError {
	return &TemplateError{
		TemplateName: templateName,
		Line:         line,
		Column:       column,
		Message:      message,
		Cause:        cause,
	}
}

// ExampleError represents an error in example processing
type ExampleError struct {
	ExampleID string
	Operation string
	Message   string
	Cause     error
}

// Error implements the error interface
func (e *ExampleError) Error() string {
	if e.ExampleID != "" {
		if e.Cause != nil {
			return fmt.Sprintf("example '%s' %s: %s: %v", e.ExampleID, e.Operation, e.Message, e.Cause)
		}
		return fmt.Sprintf("example '%s' %s: %s", e.ExampleID, e.Operation, e.Message)
	}
	
	if e.Cause != nil {
		return fmt.Sprintf("example %s: %s: %v", e.Operation, e.Message, e.Cause)
	}
	return fmt.Sprintf("example %s: %s", e.Operation, e.Message)
}

// Unwrap returns the underlying cause
func (e *ExampleError) Unwrap() error {
	return e.Cause
}

// NewExampleError creates a new example error
func NewExampleError(operation, message string, cause error) *ExampleError {
	return &ExampleError{
		Operation: operation,
		Message:   message,
		Cause:     cause,
	}
}

// NewExampleErrorWithID creates a new example error with ID
func NewExampleErrorWithID(exampleID, operation, message string, cause error) *ExampleError {
	return &ExampleError{
		ExampleID: exampleID,
		Operation: operation,
		Message:   message,
		Cause:     cause,
	}
}

// Common error constructors

// ErrInvalidTemplate creates a template validation error
func ErrInvalidTemplate(templateName, reason string) error {
	return NewTemplateError(templateName, fmt.Sprintf("invalid template: %s", reason), nil)
}

// ErrTemplateNotFound creates a template not found error
func ErrTemplateNotFound(templateName string) error {
	return NewTemplateError(templateName, "template not found", nil)
}

// ErrTemplateRenderFailed creates a template rendering error
func ErrTemplateRenderFailed(templateName, reason string, cause error) error {
	return NewTemplateError(templateName, fmt.Sprintf("rendering failed: %s", reason), cause)
}

// ErrInvalidExample creates an example validation error
func ErrInvalidExample(exampleID, reason string) error {
	return NewExampleErrorWithID(exampleID, "validation", fmt.Sprintf("invalid example: %s", reason), nil)
}

// ErrExampleSelectionFailed creates an example selection error
func ErrExampleSelectionFailed(reason string, cause error) error {
	return NewExampleError("selection", fmt.Sprintf("selection failed: %s", reason), cause)
}

// ErrPromptBuildFailed creates a prompt building error
func ErrPromptBuildFailed(reason string, cause error) error {
	return NewPromptError("build", fmt.Sprintf("build failed: %s", reason), cause)
}

// ErrPromptValidationFailed creates a prompt validation error
func ErrPromptValidationFailed(reason string) error {
	return NewPromptError("validation", fmt.Sprintf("validation failed: %s", reason), nil)
}