package visualization

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"strings"
	"sync"
)

// DefaultTemplateRenderer implements the TemplateRenderer interface using Go's html/template
type DefaultTemplateRenderer struct {
	templates map[string]*template.Template
	functions template.FuncMap
	mutex     sync.RWMutex
}

// NewDefaultTemplateRenderer creates a new default template renderer
func NewDefaultTemplateRenderer() *DefaultTemplateRenderer {
	renderer := &DefaultTemplateRenderer{
		templates: make(map[string]*template.Template),
		functions: make(template.FuncMap),
	}
	
	// Register default functions
	renderer.registerDefaultFunctions()
	
	return renderer
}

// Render renders a template with the given data
func (r *DefaultTemplateRenderer) Render(ctx context.Context, templateName string, data interface{}) (string, error) {
	r.mutex.RLock()
	tmpl, exists := r.templates[templateName]
	r.mutex.RUnlock()
	
	if !exists {
		return "", NewTemplateError(templateName, "template not found", nil)
	}
	
	var buf bytes.Buffer
	err := tmpl.Execute(&buf, data)
	if err != nil {
		return "", NewTemplateError(templateName, "template execution failed", err)
	}
	
	return buf.String(), nil
}

// RegisterTemplate registers a new template
func (r *DefaultTemplateRenderer) RegisterTemplate(name string, content string) error {
	if name == "" {
		return NewTemplateError(name, "template name cannot be empty", nil)
	}
	
	if content == "" {
		return NewTemplateError(name, "template content cannot be empty", nil)
	}
	
	// Create template with functions
	tmpl, err := template.New(name).Funcs(r.functions).Parse(content)
	if err != nil {
		return NewTemplateError(name, "template parsing failed", err)
	}
	
	r.mutex.Lock()
	r.templates[name] = tmpl
	r.mutex.Unlock()
	
	return nil
}

// GetTemplate returns a template by name
func (r *DefaultTemplateRenderer) GetTemplate(name string) (string, error) {
	r.mutex.RLock()
	_, exists := r.templates[name]
	r.mutex.RUnlock()
	
	if !exists {
		return "", NewTemplateError(name, "template not found", nil)
	}
	
	// Return the template definition (this is a simplified approach)
	// In a real implementation, we might want to store the original content
	return fmt.Sprintf("Template: %s", name), nil
}

// GetAvailableTemplates returns all available template names
func (r *DefaultTemplateRenderer) GetAvailableTemplates() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	var names []string
	for name := range r.templates {
		names = append(names, name)
	}
	
	return names
}

// RegisterFunction registers a custom function for use in templates
func (r *DefaultTemplateRenderer) RegisterFunction(name string, fn interface{}) error {
	if name == "" {
		return NewTemplateError("", "function name cannot be empty", nil)
	}
	
	if fn == nil {
		return NewTemplateError("", "function cannot be nil", nil)
	}
	
	r.mutex.Lock()
	r.functions[name] = fn
	r.mutex.Unlock()
	
	return nil
}

// GetFunctions returns all registered functions
func (r *DefaultTemplateRenderer) GetFunctions() map[string]interface{} {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	result := make(map[string]interface{})
	for name, fn := range r.functions {
		result[name] = fn
	}
	
	return result
}

// registerDefaultFunctions registers commonly used template functions
func (r *DefaultTemplateRenderer) registerDefaultFunctions() {
	r.functions["upper"] = strings.ToUpper
	r.functions["lower"] = strings.ToLower
	r.functions["title"] = strings.Title
	r.functions["trim"] = strings.TrimSpace
	r.functions["join"] = strings.Join
	r.functions["split"] = strings.Split
	r.functions["replace"] = strings.Replace
	r.functions["contains"] = strings.Contains
	r.functions["hasPrefix"] = strings.HasPrefix
	r.functions["hasSuffix"] = strings.HasSuffix
	
	// Arithmetic functions
	r.functions["add"] = func(a, b int) int { return a + b }
	r.functions["sub"] = func(a, b int) int { return a - b }
	r.functions["mul"] = func(a, b int) int { return a * b }
	r.functions["div"] = func(a, b int) int {
		if b == 0 {
			return 0
		}
		return a / b
	}
	
	// Formatting functions
	r.functions["printf"] = fmt.Sprintf
	r.functions["escape"] = template.HTMLEscapeString
	r.functions["unescape"] = template.HTMLEscapeString // Note: Go template doesn't have HTMLUnescapeString
	
	// Array/slice functions
	r.functions["len"] = func(v interface{}) int {
		switch val := v.(type) {
		case []interface{}:
			return len(val)
		case []string:
			return len(val)
		case string:
			return len(val)
		default:
			return 0
		}
	}
	
	r.functions["first"] = func(v interface{}) interface{} {
		switch val := v.(type) {
		case []interface{}:
			if len(val) > 0 {
				return val[0]
			}
		case []string:
			if len(val) > 0 {
				return val[0]
			}
		}
		return nil
	}
	
	r.functions["last"] = func(v interface{}) interface{} {
		switch val := v.(type) {
		case []interface{}:
			if len(val) > 0 {
				return val[len(val)-1]
			}
		case []string:
			if len(val) > 0 {
				return val[len(val)-1]
			}
		}
		return nil
	}
	
	// Conditional functions
	r.functions["default"] = func(defaultValue, value interface{}) interface{} {
		if value == nil || value == "" {
			return defaultValue
		}
		return value
	}
	
	r.functions["empty"] = func(value interface{}) bool {
		if value == nil {
			return true
		}
		switch v := value.(type) {
		case string:
			return v == ""
		case []interface{}:
			return len(v) == 0
		case []string:
			return len(v) == 0
		default:
			return false
		}
	}
	
	r.functions["notEmpty"] = func(value interface{}) bool {
		return !r.functions["empty"].(func(interface{}) bool)(value)
	}
}

// SimpleTemplateRenderer is a minimal template renderer for basic string replacement
type SimpleTemplateRenderer struct {
	templates map[string]string
	mutex     sync.RWMutex
}

// NewSimpleTemplateRenderer creates a new simple template renderer
func NewSimpleTemplateRenderer() *SimpleTemplateRenderer {
	return &SimpleTemplateRenderer{
		templates: make(map[string]string),
	}
}

// Render renders a template using simple string replacement
func (r *SimpleTemplateRenderer) Render(ctx context.Context, templateName string, data interface{}) (string, error) {
	r.mutex.RLock()
	template, exists := r.templates[templateName]
	r.mutex.RUnlock()
	
	if !exists {
		return "", NewTemplateError(templateName, "template not found", nil)
	}
	
	// Simple string replacement based on data map
	result := template
	if dataMap, ok := data.(map[string]interface{}); ok {
		for key, value := range dataMap {
			placeholder := fmt.Sprintf("{{.%s}}", key)
			replacement := fmt.Sprintf("%v", value)
			result = strings.Replace(result, placeholder, replacement, -1)
		}
	}
	
	return result, nil
}

// RegisterTemplate registers a new template
func (r *SimpleTemplateRenderer) RegisterTemplate(name string, content string) error {
	if name == "" {
		return NewTemplateError(name, "template name cannot be empty", nil)
	}
	
	r.mutex.Lock()
	r.templates[name] = content
	r.mutex.Unlock()
	
	return nil
}

// GetTemplate returns a template by name
func (r *SimpleTemplateRenderer) GetTemplate(name string) (string, error) {
	r.mutex.RLock()
	template, exists := r.templates[name]
	r.mutex.RUnlock()
	
	if !exists {
		return "", NewTemplateError(name, "template not found", nil)
	}
	
	return template, nil
}

// GetAvailableTemplates returns all available template names
func (r *SimpleTemplateRenderer) GetAvailableTemplates() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	var names []string
	for name := range r.templates {
		names = append(names, name)
	}
	
	return names
}

// RegisterFunction is not supported by SimpleTemplateRenderer
func (r *SimpleTemplateRenderer) RegisterFunction(name string, fn interface{}) error {
	return NewTemplateError("", "function registration not supported by SimpleTemplateRenderer", nil)
}

// GetFunctions returns an empty map as functions are not supported
func (r *SimpleTemplateRenderer) GetFunctions() map[string]interface{} {
	return make(map[string]interface{})
}

// TemplateRegistry provides a centralized registry for templates
type TemplateRegistry struct {
	renderers map[string]TemplateRenderer
	mutex     sync.RWMutex
}

// NewTemplateRegistry creates a new template registry
func NewTemplateRegistry() *TemplateRegistry {
	registry := &TemplateRegistry{
		renderers: make(map[string]TemplateRenderer),
	}
	
	// Register default renderers
	registry.renderers["default"] = NewDefaultTemplateRenderer()
	registry.renderers["simple"] = NewSimpleTemplateRenderer()
	
	return registry
}

// RegisterRenderer registers a template renderer
func (r *TemplateRegistry) RegisterRenderer(name string, renderer TemplateRenderer) error {
	if name == "" {
		return NewTemplateError("", "renderer name cannot be empty", nil)
	}
	
	if renderer == nil {
		return NewTemplateError("", "renderer cannot be nil", nil)
	}
	
	r.mutex.Lock()
	r.renderers[name] = renderer
	r.mutex.Unlock()
	
	return nil
}

// GetRenderer returns a template renderer by name
func (r *TemplateRegistry) GetRenderer(name string) (TemplateRenderer, error) {
	r.mutex.RLock()
	renderer, exists := r.renderers[name]
	r.mutex.RUnlock()
	
	if !exists {
		return nil, NewTemplateError(name, "renderer not found", nil)
	}
	
	return renderer, nil
}

// GetAvailableRenderers returns all available renderer names
func (r *TemplateRegistry) GetAvailableRenderers() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	var names []string
	for name := range r.renderers {
		names = append(names, name)
	}
	
	return names
}

// DefaultTemplateRegistry is the default global template registry
var DefaultTemplateRegistry = NewTemplateRegistry()

// Convenience functions for working with the default registry

// RegisterRenderer registers a renderer with the default registry
func RegisterRenderer(name string, renderer TemplateRenderer) error {
	return DefaultTemplateRegistry.RegisterRenderer(name, renderer)
}

// GetRenderer returns a renderer from the default registry
func GetRenderer(name string) (TemplateRenderer, error) {
	return DefaultTemplateRegistry.GetRenderer(name)
}

// GetDefaultRenderer returns the default template renderer
func GetDefaultRenderer() TemplateRenderer {
	renderer, _ := DefaultTemplateRegistry.GetRenderer("default")
	return renderer
}