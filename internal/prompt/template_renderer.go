package prompt

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"strings"
	"text/template"

	"github.com/sehwan505/langextract-go/pkg/extraction"
)

// DefaultTemplateRenderer implements template rendering using Go's text/template
type DefaultTemplateRenderer struct {
	functions map[string]interface{}
}

// NewDefaultTemplateRenderer creates a new default template renderer
func NewDefaultTemplateRenderer() *DefaultTemplateRenderer {
	renderer := &DefaultTemplateRenderer{
		functions: make(map[string]interface{}),
	}

	// Register default template functions
	renderer.registerDefaultFunctions()

	return renderer
}

// registerDefaultFunctions registers the default template functions
func (r *DefaultTemplateRenderer) registerDefaultFunctions() {
	r.functions["add"] = func(a, b int) int { return a + b }
	r.functions["sub"] = func(a, b int) int { return a - b }
	r.functions["mul"] = func(a, b int) int { return a * b }
	r.functions["div"] = func(a, b int) int {
		if b == 0 {
			return 0
		}
		return a / b
	}

	r.functions["len"] = func(v interface{}) int {
		if v == nil {
			return 0
		}
		
		// Use reflection to handle any slice, array, map, or string
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.String, reflect.Slice, reflect.Array, reflect.Map, reflect.Chan:
			return rv.Len()
		default:
			return 0
		}
	}

	r.functions["upper"] = strings.ToUpper
	r.functions["lower"] = strings.ToLower
	r.functions["title"] = strings.Title
	r.functions["trim"] = strings.TrimSpace

	r.functions["join"] = func(sep string, items []string) string {
		return strings.Join(items, sep)
	}

	r.functions["split"] = func(sep, str string) []string {
		return strings.Split(str, sep)
	}

	r.functions["contains"] = func(substr, str string) bool {
		return strings.Contains(str, substr)
	}

	r.functions["hasPrefix"] = func(prefix, str string) bool {
		return strings.HasPrefix(str, prefix)
	}

	r.functions["hasSuffix"] = func(suffix, str string) bool {
		return strings.HasSuffix(str, suffix)
	}

	r.functions["replace"] = func(old, new, str string) string {
		return strings.ReplaceAll(str, old, new)
	}

	r.functions["truncate"] = func(length int, str string) string {
		if len(str) <= length {
			return str
		}
		if length < 3 {
			return str[:length]
		}
		return str[:length-3] + "..."
	}

	r.functions["formatExample"] = func(example *extraction.ExampleData, format interface{}) string {
		formatStr := "json"
		if f, ok := format.(string); ok {
			formatStr = f
		}
		return formatExample(example, formatStr)
	}

	r.functions["formatExtractions"] = func(extractions []*extraction.Extraction, format interface{}) string {
		formatStr := "json"
		if f, ok := format.(string); ok {
			formatStr = f
		}
		example := &extraction.ExampleData{Extractions: extractions}
		return formatExample(example, formatStr)
	}

	r.functions["extractionClasses"] = func(extractions []*extraction.Extraction) []string {
		classSet := make(map[string]bool)
		for _, ext := range extractions {
			classSet[ext.Class()] = true
		}

		var classes []string
		for class := range classSet {
			classes = append(classes, class)
		}
		return classes
	}

	r.functions["countByClass"] = func(class string, extractions []*extraction.Extraction) int {
		count := 0
		for _, ext := range extractions {
			if ext.Class() == class {
				count++
			}
		}
		return count
	}

	r.functions["filterByClass"] = func(class string, extractions []*extraction.Extraction) []*extraction.Extraction {
		var filtered []*extraction.Extraction
		for _, ext := range extractions {
			if ext.Class() == class {
				filtered = append(filtered, ext)
			}
		}
		return filtered
	}

	r.functions["indent"] = func(spaces int, text string) string {
		if spaces <= 0 {
			return text
		}

		prefix := strings.Repeat(" ", spaces)
		lines := strings.Split(text, "\n")
		for i, line := range lines {
			if line != "" {
				lines[i] = prefix + line
			}
		}
		return strings.Join(lines, "\n")
	}

	r.functions["nl"] = func() string { return "\n" }
	r.functions["br"] = func() string { return "<br>" }
	r.functions["tab"] = func() string { return "\t" }

	r.functions["escape"] = func(str string) string {
		str = strings.ReplaceAll(str, "\\", "\\\\")
		str = strings.ReplaceAll(str, "\"", "\\\"")
		str = strings.ReplaceAll(str, "\n", "\\n")
		str = strings.ReplaceAll(str, "\r", "\\r")
		str = strings.ReplaceAll(str, "\t", "\\t")
		return str
	}

	r.functions["quote"] = func(str string) string {
		return fmt.Sprintf("%q", str)
	}

	r.functions["default"] = func(defaultValue, value interface{}) interface{} {
		if value == nil {
			return defaultValue
		}

		switch v := value.(type) {
		case string:
			if v == "" {
				return defaultValue
			}
		case []interface{}:
			if len(v) == 0 {
				return defaultValue
			}
		}

		return value
	}

	r.functions["coalesce"] = func(values ...interface{}) interface{} {
		for _, value := range values {
			if value != nil {
				switch v := value.(type) {
				case string:
					if v != "" {
						return v
					}
				case []interface{}:
					if len(v) > 0 {
						return v
					}
				default:
					return v
				}
			}
		}
		return nil
	}
}

// Render renders a template with the given context
func (r *DefaultTemplateRenderer) Render(ctx context.Context, template *PromptTemplate, data interface{}) (string, error) {
	if template == nil {
		return "", NewTemplateError("", "template cannot be nil", nil)
	}

	if template.Template == "" {
		return "", NewTemplateError(template.Name, "template content cannot be empty", nil)
	}

	// Create Go template
	tmpl, err := r.createTemplate(template)
	if err != nil {
		return "", NewTemplateError(template.Name, "failed to create template", err)
	}

	// Check for context cancellation
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	// Render template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", NewTemplateError(template.Name, "failed to execute template", err)
	}

	result := buf.String()

	// Post-process the result
	result = r.postProcess(result)

	return result, nil
}

// createTemplate creates a Go text/template from the prompt template
func (r *DefaultTemplateRenderer) createTemplate(promptTemplate *PromptTemplate) (*template.Template, error) {
	tmpl := template.New(promptTemplate.Name)

	// Add custom functions
	tmpl = tmpl.Funcs(r.functions)

	// Parse template
	tmpl, err := tmpl.Parse(promptTemplate.Template)
	if err != nil {
		return nil, err
	}

	return tmpl, nil
}

// postProcess performs post-processing on the rendered template
func (r *DefaultTemplateRenderer) postProcess(content string) string {
	// Remove excessive whitespace
	content = strings.TrimSpace(content)

	// Remove multiple consecutive empty lines
	lines := strings.Split(content, "\n")
	var processedLines []string
	emptyLineCount := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			emptyLineCount++
			if emptyLineCount <= 2 { // Allow max 2 consecutive empty lines
				processedLines = append(processedLines, line)
			}
		} else {
			emptyLineCount = 0
			processedLines = append(processedLines, line)
		}
	}

	return strings.Join(processedLines, "\n")
}

// RegisterFunction registers a custom function for use in templates
func (r *DefaultTemplateRenderer) RegisterFunction(name string, fn interface{}) error {
	if name == "" {
		return fmt.Errorf("function name cannot be empty")
	}

	if fn == nil {
		return fmt.Errorf("function cannot be nil")
	}

	r.functions[name] = fn
	return nil
}

// GetFunctions returns all registered functions
func (r *DefaultTemplateRenderer) GetFunctions() map[string]interface{} {
	// Return a copy to prevent external modification
	funcs := make(map[string]interface{})
	for name, fn := range r.functions {
		funcs[name] = fn
	}
	return funcs
}

// UnregisterFunction removes a function from the renderer
func (r *DefaultTemplateRenderer) UnregisterFunction(name string) {
	delete(r.functions, name)
}

// HasFunction checks if a function is registered
func (r *DefaultTemplateRenderer) HasFunction(name string) bool {
	_, exists := r.functions[name]
	return exists
}

// ClearFunctions removes all custom functions (keeps default ones)
func (r *DefaultTemplateRenderer) ClearFunctions() {
	r.functions = make(map[string]interface{})
	r.registerDefaultFunctions()
}

// ValidateTemplate validates a template without rendering it
func (r *DefaultTemplateRenderer) ValidateTemplate(promptTemplate *PromptTemplate) error {
	if promptTemplate == nil {
		return NewTemplateError("", "template cannot be nil", nil)
	}

	if promptTemplate.Template == "" {
		return NewTemplateError(promptTemplate.Name, "template content cannot be empty", nil)
	}

	// Try to parse the template
	_, err := r.createTemplate(promptTemplate)
	if err != nil {
		return NewTemplateError(promptTemplate.Name, "template validation failed", err)
	}

	return nil
}

// GetTemplateVariables extracts variable names from a template
func (r *DefaultTemplateRenderer) GetTemplateVariables(templateContent string) ([]string, error) {
	// This is a simplified implementation
	// A more sophisticated version would parse the template AST
	var variables []string
	lines := strings.Split(templateContent, "\n")

	for _, line := range lines {
		// Look for {{.Variable}} patterns
		start := 0
		for {
			openIndex := strings.Index(line[start:], "{{.")
			if openIndex == -1 {
				break
			}
			openIndex += start

			closeIndex := strings.Index(line[openIndex:], "}}")
			if closeIndex == -1 {
				break
			}
			closeIndex += openIndex

			// Extract variable name
			varContent := line[openIndex+3 : closeIndex]
			parts := strings.Fields(varContent)
			if len(parts) > 0 {
				// Handle nested access like .Task.Description
				varName := parts[0]
				if !contains(variables, varName) {
					variables = append(variables, varName)
				}
			}

			start = closeIndex + 2
		}
	}

	return variables, nil
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}