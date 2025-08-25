package visualization

import (
	"fmt"
)

// VisualizationError represents a base error type for visualization operations
type VisualizationError struct {
	Operation string                 `json:"operation"`
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Cause     error                  `json:"-"`
}

// Error implements the error interface
func (e *VisualizationError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("visualization %s failed: %s: %v", e.Operation, e.Message, e.Cause)
	}
	return fmt.Sprintf("visualization %s failed: %s", e.Operation, e.Message)
}

// Unwrap returns the wrapped error
func (e *VisualizationError) Unwrap() error {
	return e.Cause
}

// ValidationError represents validation-related visualization errors
type ValidationError struct {
	*VisualizationError
}

// NewValidationError creates a new validation error
func NewValidationError(message string, details map[string]interface{}) *ValidationError {
	return &ValidationError{
		VisualizationError: &VisualizationError{
			Operation: "validation",
			Message:   message,
			Details:   details,
		},
	}
}

// RenderingError represents template rendering errors
type RenderingError struct {
	*VisualizationError
}

// NewRenderingError creates a new rendering error
func NewRenderingError(message string, cause error) *RenderingError {
	return &RenderingError{
		VisualizationError: &VisualizationError{
			Operation: "rendering",
			Message:   message,
			Cause:     cause,
		},
	}
}

// ExportError represents data export errors
type ExportError struct {
	*VisualizationError
}

// NewExportError creates a new export error
func NewExportError(message string, format OutputFormat, cause error) *ExportError {
	return &ExportError{
		VisualizationError: &VisualizationError{
			Operation: "export",
			Message:   message,
			Details: map[string]interface{}{
				"format": format,
			},
			Cause: cause,
		},
	}
}

// TemplateError represents template-related errors
type TemplateError struct {
	*VisualizationError
}

// NewTemplateError creates a new template error
func NewTemplateError(templateName, message string, cause error) *TemplateError {
	return &TemplateError{
		VisualizationError: &VisualizationError{
			Operation: "template",
			Message:   message,
			Details: map[string]interface{}{
				"template": templateName,
			},
			Cause: cause,
		},
	}
}

// ColorError represents color assignment errors
type ColorError struct {
	*VisualizationError
}

// NewColorError creates a new color error
func NewColorError(message string, details map[string]interface{}) *ColorError {
	return &ColorError{
		VisualizationError: &VisualizationError{
			Operation: "color_assignment",
			Message:   message,
			Details:   details,
		},
	}
}

// HTMLGenerationError represents HTML generation errors
type HTMLGenerationError struct {
	*VisualizationError
}

// NewHTMLGenerationError creates a new HTML generation error
func NewHTMLGenerationError(message string, cause error) *HTMLGenerationError {
	return &HTMLGenerationError{
		VisualizationError: &VisualizationError{
			Operation: "html_generation",
			Message:   message,
			Cause:     cause,
		},
	}
}

// DataProcessingError represents data processing errors
type DataProcessingError struct {
	*VisualizationError
}

// NewDataProcessingError creates a new data processing error
func NewDataProcessingError(message string, cause error) *DataProcessingError {
	return &DataProcessingError{
		VisualizationError: &VisualizationError{
			Operation: "data_processing",
			Message:   message,
			Cause:     cause,
		},
	}
}

// Error type predicates for easy error type checking

// IsValidationError checks if the error is a validation error
func IsValidationError(err error) bool {
	_, ok := err.(*ValidationError)
	return ok
}

// IsRenderingError checks if the error is a rendering error
func IsRenderingError(err error) bool {
	_, ok := err.(*RenderingError)
	return ok
}

// IsExportError checks if the error is an export error
func IsExportError(err error) bool {
	_, ok := err.(*ExportError)
	return ok
}

// IsTemplateError checks if the error is a template error
func IsTemplateError(err error) bool {
	_, ok := err.(*TemplateError)
	return ok
}

// IsColorError checks if the error is a color error
func IsColorError(err error) bool {
	_, ok := err.(*ColorError)
	return ok
}

// IsHTMLGenerationError checks if the error is an HTML generation error
func IsHTMLGenerationError(err error) bool {
	_, ok := err.(*HTMLGenerationError)
	return ok
}

// IsDataProcessingError checks if the error is a data processing error
func IsDataProcessingError(err error) bool {
	_, ok := err.(*DataProcessingError)
	return ok
}