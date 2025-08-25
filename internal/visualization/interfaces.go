package visualization

import (
	"context"

	"github.com/sehwan505/langextract-go/pkg/document"
	"github.com/sehwan505/langextract-go/pkg/extraction"
)

// Visualizer defines the main interface for generating visualizations
type Visualizer interface {
	// Generate creates a visualization from an annotated document
	Generate(ctx context.Context, doc *document.AnnotatedDocument, opts *VisualizationOptions) (string, error)
	
	// GetSupportedFormats returns the formats this visualizer supports
	GetSupportedFormats() []OutputFormat
	
	// Name returns the name of the visualizer
	Name() string
	
	// Validate validates the visualizer configuration
	Validate() error
}

// Exporter defines the interface for exporting annotated documents to different formats
type Exporter interface {
	// Export exports an annotated document to the specified format
	Export(ctx context.Context, doc *document.AnnotatedDocument, opts *ExportOptions) ([]byte, error)
	
	// GetFormat returns the output format this exporter handles
	GetFormat() OutputFormat
	
	// GetMIMEType returns the MIME type for the exported format
	GetMIMEType() string
	
	// Name returns the name of the exporter
	Name() string
}

// ColorManager defines the interface for managing extraction class colors
type ColorManager interface {
	// AssignColors assigns colors to extraction classes
	AssignColors(extractions []*extraction.Extraction) map[string]string
	
	// GetColor returns the color for a specific class
	GetColor(class string) string
	
	// GetPalette returns the color palette being used
	GetPalette() []string
	
	// SetPalette sets a custom color palette
	SetPalette(palette []string)
}

// TemplateRenderer defines the interface for rendering visualization templates
type TemplateRenderer interface {
	// Render renders a template with the given data
	Render(ctx context.Context, templateName string, data interface{}) (string, error)
	
	// RegisterTemplate registers a new template
	RegisterTemplate(name string, content string) error
	
	// GetTemplate returns a template by name
	GetTemplate(name string) (string, error)
	
	// GetAvailableTemplates returns all available template names
	GetAvailableTemplates() []string
}

// OutputFormat represents supported visualization output formats
type OutputFormat string

const (
	// OutputFormatHTML generates interactive HTML visualization
	OutputFormatHTML OutputFormat = "html"
	
	// OutputFormatJSON exports structured JSON with metadata
	OutputFormatJSON OutputFormat = "json"
	
	// OutputFormatCSV exports tabular CSV data
	OutputFormatCSV OutputFormat = "csv"
	
	// OutputFormatMarkdown exports human-readable Markdown
	OutputFormatMarkdown OutputFormat = "markdown"
	
	// OutputFormatPlainText exports simple text format
	OutputFormatPlainText OutputFormat = "text"
)

// String returns the string representation of the output format
func (f OutputFormat) String() string {
	return string(f)
}

// IsValid checks if the output format is valid
func (f OutputFormat) IsValid() bool {
	switch f {
	case OutputFormatHTML, OutputFormatJSON, OutputFormatCSV, OutputFormatMarkdown, OutputFormatPlainText:
		return true
	default:
		return false
	}
}

// GetMIMEType returns the MIME type for the output format
func (f OutputFormat) GetMIMEType() string {
	switch f {
	case OutputFormatHTML:
		return "text/html"
	case OutputFormatJSON:
		return "application/json"
	case OutputFormatCSV:
		return "text/csv"
	case OutputFormatMarkdown:
		return "text/markdown"
	case OutputFormatPlainText:
		return "text/plain"
	default:
		return "application/octet-stream"
	}
}

// VisualizationOptions contains options for generating visualizations
type VisualizationOptions struct {
	// Format specifies the output format
	Format OutputFormat `json:"format"`
	
	// ShowLegend controls whether to show the class color legend
	ShowLegend bool `json:"show_legend"`
	
	// AnimationSpeed controls the speed of animations (HTML only)
	AnimationSpeed float64 `json:"animation_speed,omitempty"`
	
	// GIFOptimized applies optimizations for GIF creation (HTML only)
	GIFOptimized bool `json:"gif_optimized,omitempty"`
	
	// ContextChars specifies the number of context characters to show around extractions
	ContextChars int `json:"context_chars,omitempty"`
	
	// MaxTextLength limits the length of text to display
	MaxTextLength int `json:"max_text_length,omitempty"`
	
	// CustomColors allows overriding the default color palette
	CustomColors map[string]string `json:"custom_colors,omitempty"`
	
	// TemplateOverrides allows overriding default templates
	TemplateOverrides map[string]string `json:"template_overrides,omitempty"`
	
	// IncludeMetadata controls whether to include extraction metadata
	IncludeMetadata bool `json:"include_metadata"`
	
	// SortExtractions controls whether to sort extractions by position
	SortExtractions bool `json:"sort_extractions"`
	
	// FilterClasses allows filtering to specific extraction classes
	FilterClasses []string `json:"filter_classes,omitempty"`
	
	// Debug enables debug mode for troubleshooting
	Debug bool `json:"debug,omitempty"`
}

// ExportOptions contains options for exporting documents
type ExportOptions struct {
	// Format specifies the output format
	Format OutputFormat `json:"format"`
	
	// IncludeText controls whether to include the full document text
	IncludeText bool `json:"include_text"`
	
	// IncludeMetadata controls whether to include extraction metadata
	IncludeMetadata bool `json:"include_metadata"`
	
	// IncludeAttributes controls whether to include extraction attributes
	IncludeAttributes bool `json:"include_attributes"`
	
	// IncludePositions controls whether to include character positions
	IncludePositions bool `json:"include_positions"`
	
	// Pretty controls pretty-printing for structured formats
	Pretty bool `json:"pretty"`
	
	// CSVDelimiter specifies the delimiter for CSV format
	CSVDelimiter string `json:"csv_delimiter,omitempty"`
	
	// CSVHeaders specifies custom headers for CSV format
	CSVHeaders []string `json:"csv_headers,omitempty"`
	
	// FilterClasses allows filtering to specific extraction classes
	FilterClasses []string `json:"filter_classes,omitempty"`
	
	// SortBy specifies the field to sort by
	SortBy string `json:"sort_by,omitempty"`
	
	// SortOrder specifies ascending or descending sort
	SortOrder SortOrder `json:"sort_order,omitempty"`
}

// SortOrder represents sort order options
type SortOrder string

const (
	// SortOrderAsc sorts in ascending order
	SortOrderAsc SortOrder = "asc"
	
	// SortOrderDesc sorts in descending order
	SortOrderDesc SortOrder = "desc"
)

// SpanPoint represents a span boundary point for HTML generation
type SpanPoint struct {
	// Position is the character position in the text
	Position int `json:"position"`
	
	// TagType indicates whether this is a start or end tag
	TagType TagType `json:"tag_type"`
	
	// SpanIndex is the index of the span for HTML data attributes
	SpanIndex int `json:"span_index"`
	
	// Extraction is the associated extraction data
	Extraction *extraction.Extraction `json:"extraction"`
}

// TagType represents span boundary tag types
type TagType string

const (
	// TagTypeStart indicates a span start tag
	TagTypeStart TagType = "start"
	
	// TagTypeEnd indicates a span end tag
	TagTypeEnd TagType = "end"
)

// ExtractionData represents processed extraction data for visualization
type ExtractionData struct {
	// Index is the extraction index
	Index int `json:"index"`
	
	// Class is the extraction class
	Class string `json:"class"`
	
	// Text is the extracted text
	Text string `json:"text"`
	
	// Color is the assigned color for this class
	Color string `json:"color"`
	
	// StartPos is the start character position
	StartPos int `json:"start_pos"`
	
	// EndPos is the end character position
	EndPos int `json:"end_pos"`
	
	// BeforeText is the context text before the extraction
	BeforeText string `json:"before_text"`
	
	// ExtractionText is the extracted text (HTML-escaped)
	ExtractionText string `json:"extraction_text"`
	
	// AfterText is the context text after the extraction
	AfterText string `json:"after_text"`
	
	// AttributesHTML is the HTML representation of attributes
	AttributesHTML string `json:"attributes_html"`
	
	// Confidence is the extraction confidence score
	Confidence float64 `json:"confidence,omitempty"`
	
	// Metadata contains additional extraction metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// DefaultVisualizationOptions returns default visualization options
func DefaultVisualizationOptions() *VisualizationOptions {
	return &VisualizationOptions{
		Format:          OutputFormatHTML,
		ShowLegend:      true,
		AnimationSpeed:  1.0,
		GIFOptimized:    false,
		ContextChars:    150,
		MaxTextLength:   0, // No limit
		IncludeMetadata: true,
		SortExtractions: true,
		Debug:          false,
	}
}

// WithFormat sets the output format
func (opts *VisualizationOptions) WithFormat(format OutputFormat) *VisualizationOptions {
	opts.Format = format
	return opts
}

// WithShowLegend sets whether to show the legend
func (opts *VisualizationOptions) WithShowLegend(show bool) *VisualizationOptions {
	opts.ShowLegend = show
	return opts
}

// WithAnimationSpeed sets the animation speed
func (opts *VisualizationOptions) WithAnimationSpeed(speed float64) *VisualizationOptions {
	opts.AnimationSpeed = speed
	return opts
}

// WithContextChars sets the number of context characters
func (opts *VisualizationOptions) WithContextChars(chars int) *VisualizationOptions {
	opts.ContextChars = chars
	return opts
}

// WithCustomColors sets custom colors for extraction classes
func (opts *VisualizationOptions) WithCustomColors(colors map[string]string) *VisualizationOptions {
	opts.CustomColors = colors
	return opts
}

// WithDebug enables debug mode
func (opts *VisualizationOptions) WithDebug(debug bool) *VisualizationOptions {
	opts.Debug = debug
	return opts
}

// DefaultExportOptions returns default export options
func DefaultExportOptions() *ExportOptions {
	return &ExportOptions{
		Format:            OutputFormatJSON,
		IncludeText:       true,
		IncludeMetadata:   true,
		IncludeAttributes: true,
		IncludePositions:  true,
		Pretty:            true,
		CSVDelimiter:      ",",
		SortOrder:         SortOrderAsc,
	}
}

// WithFormat sets the export format
func (opts *ExportOptions) WithFormat(format OutputFormat) *ExportOptions {
	opts.Format = format
	return opts
}

// WithPretty sets pretty printing
func (opts *ExportOptions) WithPretty(pretty bool) *ExportOptions {
	opts.Pretty = pretty
	return opts
}

// WithCSVDelimiter sets the CSV delimiter
func (opts *ExportOptions) WithCSVDelimiter(delimiter string) *ExportOptions {
	opts.CSVDelimiter = delimiter
	return opts
}

// WithFilterClasses sets the classes to filter
func (opts *ExportOptions) WithFilterClasses(classes []string) *ExportOptions {
	opts.FilterClasses = classes
	return opts
}

// Validate validates the visualization options
func (opts *VisualizationOptions) Validate() error {
	if !opts.Format.IsValid() {
		return NewValidationError("invalid output format", map[string]interface{}{
			"format": opts.Format,
		})
	}
	
	if opts.AnimationSpeed < 0.1 || opts.AnimationSpeed > 10.0 {
		return NewValidationError("animation speed must be between 0.1 and 10.0 seconds", map[string]interface{}{
			"animation_speed": opts.AnimationSpeed,
		})
	}
	
	if opts.ContextChars < 0 {
		return NewValidationError("context chars must be non-negative", map[string]interface{}{
			"context_chars": opts.ContextChars,
		})
	}
	
	if opts.MaxTextLength < 0 {
		return NewValidationError("max text length must be non-negative", map[string]interface{}{
			"max_text_length": opts.MaxTextLength,
		})
	}
	
	return nil
}

// Validate validates the export options
func (opts *ExportOptions) Validate() error {
	if !opts.Format.IsValid() {
		return NewValidationError("invalid output format", map[string]interface{}{
			"format": opts.Format,
		})
	}
	
	if opts.CSVDelimiter == "" && opts.Format == OutputFormatCSV {
		return NewValidationError("CSV delimiter cannot be empty", nil)
	}
	
	if opts.SortOrder != "" && opts.SortOrder != SortOrderAsc && opts.SortOrder != SortOrderDesc {
		return NewValidationError("invalid sort order", map[string]interface{}{
			"sort_order": opts.SortOrder,
		})
	}
	
	return nil
}