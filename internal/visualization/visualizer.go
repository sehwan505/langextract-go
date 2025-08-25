package visualization

import (
	"context"
	"fmt"
	"sync"

	"github.com/sehwan505/langextract-go/pkg/document"
)

// DefaultVisualizer is the main visualizer that orchestrates different generators and exporters
type DefaultVisualizer struct {
	htmlGenerator *HTMLGenerator
	jsonExporter  *JSONExporter
	csvExporter   *CSVExporter
	mdExporter    *MarkdownExporter
	colorManager  ColorManager
	options       *VisualizationOptions
	mutex         sync.RWMutex
}

// NewDefaultVisualizer creates a new default visualizer with all exporters
func NewDefaultVisualizer(opts *VisualizationOptions) *DefaultVisualizer {
	if opts == nil {
		opts = DefaultVisualizationOptions()
	}
	
	colorManager := NewDefaultColorManager()
	
	return &DefaultVisualizer{
		htmlGenerator: NewHTMLGeneratorWithColorManager(colorManager, opts),
		jsonExporter:  NewJSONExporter(nil),
		csvExporter:   NewCSVExporter(nil),
		mdExporter:    NewMarkdownExporter(nil),
		colorManager:  colorManager,
		options:      opts,
	}
}

// Generate creates a visualization from an annotated document
func (v *DefaultVisualizer) Generate(ctx context.Context, doc *document.AnnotatedDocument, opts *VisualizationOptions) (string, error) {
	if doc == nil {
		return "", NewValidationError("document cannot be nil", nil)
	}
	
	// Use provided options or fall back to visualizer options
	if opts == nil {
		opts = v.options
	}
	
	if err := opts.Validate(); err != nil {
		return "", err
	}
	
	// Route to appropriate generator based on format
	switch opts.Format {
	case OutputFormatHTML:
		return v.htmlGenerator.Generate(ctx, doc, opts)
		
	case OutputFormatJSON:
		exportOpts := v.convertToExportOptions(opts)
		data, err := v.jsonExporter.Export(ctx, doc, exportOpts)
		if err != nil {
			return "", err
		}
		return string(data), nil
		
	case OutputFormatCSV:
		exportOpts := v.convertToExportOptions(opts)
		data, err := v.csvExporter.Export(ctx, doc, exportOpts)
		if err != nil {
			return "", err
		}
		return string(data), nil
		
	case OutputFormatMarkdown:
		exportOpts := v.convertToExportOptions(opts)
		data, err := v.mdExporter.Export(ctx, doc, exportOpts)
		if err != nil {
			return "", err
		}
		return string(data), nil
		
	case OutputFormatPlainText:
		return v.generatePlainText(doc, opts), nil
		
	default:
		return "", NewValidationError("unsupported output format", map[string]interface{}{
			"format": opts.Format,
		})
	}
}

// GetSupportedFormats returns all supported formats
func (v *DefaultVisualizer) GetSupportedFormats() []OutputFormat {
	return []OutputFormat{
		OutputFormatHTML,
		OutputFormatJSON,
		OutputFormatCSV,
		OutputFormatMarkdown,
		OutputFormatPlainText,
	}
}

// Name returns the name of the visualizer
func (v *DefaultVisualizer) Name() string {
	return "DefaultVisualizer"
}

// Validate validates the visualizer configuration
func (v *DefaultVisualizer) Validate() error {
	if v.htmlGenerator == nil {
		return NewValidationError("HTML generator cannot be nil", nil)
	}
	
	if v.jsonExporter == nil {
		return NewValidationError("JSON exporter cannot be nil", nil)
	}
	
	if v.csvExporter == nil {
		return NewValidationError("CSV exporter cannot be nil", nil)
	}
	
	if v.mdExporter == nil {
		return NewValidationError("Markdown exporter cannot be nil", nil)
	}
	
	if v.colorManager == nil {
		return NewValidationError("color manager cannot be nil", nil)
	}
	
	return nil
}

// Export exports a document using the specified exporter
func (v *DefaultVisualizer) Export(ctx context.Context, doc *document.AnnotatedDocument, format OutputFormat, opts *ExportOptions) ([]byte, error) {
	if doc == nil {
		return nil, NewValidationError("document cannot be nil", nil)
	}
	
	if opts == nil {
		opts = DefaultExportOptions().WithFormat(format)
	}
	
	if err := opts.Validate(); err != nil {
		return nil, err
	}
	
	// Route to appropriate exporter
	switch format {
	case OutputFormatJSON:
		return v.jsonExporter.Export(ctx, doc, opts)
		
	case OutputFormatCSV:
		return v.csvExporter.Export(ctx, doc, opts)
		
	case OutputFormatMarkdown:
		return v.mdExporter.Export(ctx, doc, opts)
		
	case OutputFormatHTML:
		// Convert to visualization options
		vizOpts := v.convertToVisualizationOptions(opts)
		html, err := v.htmlGenerator.Generate(ctx, doc, vizOpts)
		if err != nil {
			return nil, err
		}
		return []byte(html), nil
		
	case OutputFormatPlainText:
		text := v.generatePlainText(doc, v.convertToVisualizationOptions(opts))
		return []byte(text), nil
		
	default:
		return nil, NewValidationError("unsupported export format", map[string]interface{}{
			"format": format,
		})
	}
}

// GetColorManager returns the color manager
func (v *DefaultVisualizer) GetColorManager() ColorManager {
	v.mutex.RLock()
	defer v.mutex.RUnlock()
	return v.colorManager
}

// SetColorManager sets a custom color manager
func (v *DefaultVisualizer) SetColorManager(colorManager ColorManager) error {
	if colorManager == nil {
		return NewValidationError("color manager cannot be nil", nil)
	}
	
	v.mutex.Lock()
	defer v.mutex.Unlock()
	
	v.colorManager = colorManager
	
	// Update HTML generator with new color manager
	v.htmlGenerator = NewHTMLGeneratorWithColorManager(colorManager, v.options)
	
	return nil
}

// UpdateOptions updates the visualizer options
func (v *DefaultVisualizer) UpdateOptions(opts *VisualizationOptions) error {
	if opts == nil {
		return NewValidationError("options cannot be nil", nil)
	}
	
	if err := opts.Validate(); err != nil {
		return err
	}
	
	v.mutex.Lock()
	defer v.mutex.Unlock()
	
	v.options = opts
	
	// Update HTML generator with new options
	v.htmlGenerator = NewHTMLGeneratorWithColorManager(v.colorManager, opts)
	
	return nil
}

// GetOptions returns a copy of the current options
func (v *DefaultVisualizer) GetOptions() *VisualizationOptions {
	v.mutex.RLock()
	defer v.mutex.RUnlock()
	
	if v.options == nil {
		return DefaultVisualizationOptions()
	}
	
	// Return a copy to prevent external modifications
	opts := *v.options
	return &opts
}

// GetExporter returns an exporter for the specified format
func (v *DefaultVisualizer) GetExporter(format OutputFormat) (Exporter, error) {
	switch format {
	case OutputFormatJSON:
		return v.jsonExporter, nil
	case OutputFormatCSV:
		return v.csvExporter, nil
	case OutputFormatMarkdown:
		return v.mdExporter, nil
	default:
		return nil, NewValidationError("unsupported format for exporter", map[string]interface{}{
			"format": format,
		})
	}
}

// Helper methods

// convertToExportOptions converts visualization options to export options
func (v *DefaultVisualizer) convertToExportOptions(vizOpts *VisualizationOptions) *ExportOptions {
	exportOpts := DefaultExportOptions()
	
	if vizOpts != nil {
		exportOpts.Format = vizOpts.Format
		exportOpts.IncludeMetadata = vizOpts.IncludeMetadata
		exportOpts.FilterClasses = vizOpts.FilterClasses
		exportOpts.Pretty = true // Default to pretty for visualizations
		
		// Set sort options if extractions should be sorted
		if vizOpts.SortExtractions {
			exportOpts.SortBy = "position"
			exportOpts.SortOrder = SortOrderAsc
		}
	}
	
	return exportOpts
}

// convertToVisualizationOptions converts export options to visualization options
func (v *DefaultVisualizer) convertToVisualizationOptions(exportOpts *ExportOptions) *VisualizationOptions {
	vizOpts := DefaultVisualizationOptions()
	
	if exportOpts != nil {
		vizOpts.Format = exportOpts.Format
		vizOpts.IncludeMetadata = exportOpts.IncludeMetadata
		vizOpts.FilterClasses = exportOpts.FilterClasses
		vizOpts.SortExtractions = exportOpts.SortBy != ""
	}
	
	return vizOpts
}

// generatePlainText generates a simple plain text representation
func (v *DefaultVisualizer) generatePlainText(doc *document.AnnotatedDocument, opts *VisualizationOptions) string {
	if doc == nil || len(doc.Extractions()) == 0 {
		return "No extractions found."
	}
	
	var result []string
	result = append(result, "=== LangExtract Results ===")
	result = append(result, "")
	
	// Filter extractions if needed
	extractions := doc.Extractions()
	if len(opts.FilterClasses) > 0 {
		var filtered []*document.Extraction
		classSet := make(map[string]bool)
		for _, class := range opts.FilterClasses {
			classSet[class] = true
		}
		
		for _, ext := range extractions {
			if ext != nil && classSet[ext.Class()] {
				filtered = append(filtered, ext)
			}
		}
		extractions = filtered
	}
	
	result = append(result, fmt.Sprintf("Total Extractions: %d", len(extractions)))
	result = append(result, "")
	
	for i, ext := range extractions {
		if ext == nil {
			continue
		}
		
		result = append(result, fmt.Sprintf("%d. %s", i+1, ext.Text()))
		result = append(result, fmt.Sprintf("   Class: %s", ext.Class()))
		
		if ext.Interval() != nil {
			interval := ext.Interval()
			result = append(result, fmt.Sprintf("   Position: %d-%d", interval.Start(), interval.End()))
		}
		
		if ext.Confidence() > 0 {
			result = append(result, fmt.Sprintf("   Confidence: %.4f", ext.Confidence()))
		}
		
		result = append(result, "")
	}
	
	return fmt.Sprintf("%s\n", strings.Join(result, "\n"))
}

// VisualizationRegistry provides a centralized registry for visualizers
type VisualizationRegistry struct {
	visualizers map[string]Visualizer
	mutex       sync.RWMutex
}

// NewVisualizationRegistry creates a new visualization registry
func NewVisualizationRegistry() *VisualizationRegistry {
	registry := &VisualizationRegistry{
		visualizers: make(map[string]Visualizer),
	}
	
	// Register default visualizer
	registry.visualizers["default"] = NewDefaultVisualizer(nil)
	
	return registry
}

// RegisterVisualizer registers a visualizer
func (r *VisualizationRegistry) RegisterVisualizer(name string, visualizer Visualizer) error {
	if name == "" {
		return NewValidationError("visualizer name cannot be empty", nil)
	}
	
	if visualizer == nil {
		return NewValidationError("visualizer cannot be nil", nil)
	}
	
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	r.visualizers[name] = visualizer
	return nil
}

// GetVisualizer returns a visualizer by name
func (r *VisualizationRegistry) GetVisualizer(name string) (Visualizer, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	visualizer, exists := r.visualizers[name]
	if !exists {
		return nil, NewValidationError("visualizer not found", map[string]interface{}{
			"name": name,
		})
	}
	
	return visualizer, nil
}

// GetAvailableVisualizers returns all available visualizer names
func (r *VisualizationRegistry) GetAvailableVisualizers() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	var names []string
	for name := range r.visualizers {
		names = append(names, name)
	}
	
	return names
}

// DefaultVisualizationRegistry is the default global visualization registry
var DefaultVisualizationRegistry = NewVisualizationRegistry()

// Convenience functions for working with the default registry

// RegisterVisualizer registers a visualizer with the default registry
func RegisterVisualizer(name string, visualizer Visualizer) error {
	return DefaultVisualizationRegistry.RegisterVisualizer(name, visualizer)
}

// GetVisualizer returns a visualizer from the default registry
func GetVisualizer(name string) (Visualizer, error) {
	return DefaultVisualizationRegistry.GetVisualizer(name)
}

// GetDefaultVisualizer returns the default visualizer
func GetDefaultVisualizer() Visualizer {
	visualizer, _ := DefaultVisualizationRegistry.GetVisualizer("default")
	return visualizer
}