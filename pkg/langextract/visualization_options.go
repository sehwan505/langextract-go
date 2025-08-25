package langextract

import (
	"github.com/sehwan505/langextract-go/internal/visualization"
)

// VisualizeOptions contains options for generating visualizations.
// This provides a public interface to the internal visualization options.
type VisualizeOptions struct {
	// Format specifies the output format (html, json, csv, markdown, text)
	Format string `json:"format"`
	
	// ShowLegend controls whether to show the class color legend (HTML only)
	ShowLegend bool `json:"show_legend"`
	
	// AnimationSpeed controls the speed of animations in seconds (HTML only)
	AnimationSpeed float64 `json:"animation_speed,omitempty"`
	
	// GIFOptimized applies optimizations for GIF creation (HTML only)
	GIFOptimized bool `json:"gif_optimized,omitempty"`
	
	// ContextChars specifies the number of context characters around extractions
	ContextChars int `json:"context_chars,omitempty"`
	
	// MaxTextLength limits the length of text to display
	MaxTextLength int `json:"max_text_length,omitempty"`
	
	// CustomColors allows overriding the default color palette
	CustomColors map[string]string `json:"custom_colors,omitempty"`
	
	// IncludeMetadata controls whether to include extraction metadata
	IncludeMetadata bool `json:"include_metadata"`
	
	// SortExtractions controls whether to sort extractions by position
	SortExtractions bool `json:"sort_extractions"`
	
	// FilterClasses allows filtering to specific extraction classes
	FilterClasses []string `json:"filter_classes,omitempty"`
	
	// Pretty controls pretty-printing for structured formats
	Pretty bool `json:"pretty"`
	
	// CSVDelimiter specifies the delimiter for CSV format
	CSVDelimiter string `json:"csv_delimiter,omitempty"`
	
	// IncludeText controls whether to include the full document text
	IncludeText bool `json:"include_text"`
	
	// IncludePositions controls whether to include character positions
	IncludePositions bool `json:"include_positions"`
	
	// Debug enables debug mode for troubleshooting
	Debug bool `json:"debug,omitempty"`
}

// NewVisualizeOptions creates a new VisualizeOptions with sensible defaults.
func NewVisualizeOptions() *VisualizeOptions {
	return &VisualizeOptions{
		Format:           "html",
		ShowLegend:       true,
		AnimationSpeed:   1.0,
		GIFOptimized:     false,
		ContextChars:     150,
		MaxTextLength:    0, // No limit
		IncludeMetadata:  true,
		SortExtractions:  true,
		Pretty:           true,
		CSVDelimiter:     ",",
		IncludeText:      true,
		IncludePositions: true,
		Debug:           false,
	}
}

// WithFormat sets the output format.
// Supported formats: "html", "json", "csv", "markdown", "text"
func (opts *VisualizeOptions) WithFormat(format string) *VisualizeOptions {
	opts.Format = format
	return opts
}

// WithShowLegend sets whether to show the legend.
func (opts *VisualizeOptions) WithShowLegend(show bool) *VisualizeOptions {
	opts.ShowLegend = show
	return opts
}

// WithAnimationSpeed sets the animation speed for HTML visualizations.
func (opts *VisualizeOptions) WithAnimationSpeed(speed float64) *VisualizeOptions {
	opts.AnimationSpeed = speed
	return opts
}

// WithGIFOptimized enables or disables GIF optimizations.
func (opts *VisualizeOptions) WithGIFOptimized(optimized bool) *VisualizeOptions {
	opts.GIFOptimized = optimized
	return opts
}

// WithContextChars sets the number of context characters around extractions.
func (opts *VisualizeOptions) WithContextChars(chars int) *VisualizeOptions {
	opts.ContextChars = chars
	return opts
}

// WithMaxTextLength sets the maximum text length to display.
func (opts *VisualizeOptions) WithMaxTextLength(length int) *VisualizeOptions {
	opts.MaxTextLength = length
	return opts
}

// WithCustomColors sets custom colors for extraction classes.
func (opts *VisualizeOptions) WithCustomColors(colors map[string]string) *VisualizeOptions {
	opts.CustomColors = colors
	return opts
}

// WithIncludeMetadata sets whether to include extraction metadata.
func (opts *VisualizeOptions) WithIncludeMetadata(include bool) *VisualizeOptions {
	opts.IncludeMetadata = include
	return opts
}

// WithSortExtractions sets whether to sort extractions by position.
func (opts *VisualizeOptions) WithSortExtractions(sort bool) *VisualizeOptions {
	opts.SortExtractions = sort
	return opts
}

// WithFilterClasses sets the extraction classes to filter by.
func (opts *VisualizeOptions) WithFilterClasses(classes []string) *VisualizeOptions {
	opts.FilterClasses = classes
	return opts
}

// WithPretty sets whether to use pretty formatting.
func (opts *VisualizeOptions) WithPretty(pretty bool) *VisualizeOptions {
	opts.Pretty = pretty
	return opts
}

// WithCSVDelimiter sets the CSV delimiter.
func (opts *VisualizeOptions) WithCSVDelimiter(delimiter string) *VisualizeOptions {
	opts.CSVDelimiter = delimiter
	return opts
}

// WithIncludeText sets whether to include the full document text.
func (opts *VisualizeOptions) WithIncludeText(include bool) *VisualizeOptions {
	opts.IncludeText = include
	return opts
}

// WithIncludePositions sets whether to include character positions.
func (opts *VisualizeOptions) WithIncludePositions(include bool) *VisualizeOptions {
	opts.IncludePositions = include
	return opts
}

// WithDebug enables or disables debug mode.
func (opts *VisualizeOptions) WithDebug(debug bool) *VisualizeOptions {
	opts.Debug = debug
	return opts
}

// convertToVisualizationOptions converts public options to internal options
func convertToVisualizationOptions(opts *VisualizeOptions) *visualization.VisualizationOptions {
	if opts == nil {
		opts = NewVisualizeOptions()
	}
	
	// Map format string to internal format enum
	var format visualization.OutputFormat
	switch opts.Format {
	case "html":
		format = visualization.OutputFormatHTML
	case "json":
		format = visualization.OutputFormatJSON
	case "csv":
		format = visualization.OutputFormatCSV
	case "markdown", "md":
		format = visualization.OutputFormatMarkdown
	case "text", "txt":
		format = visualization.OutputFormatPlainText
	default:
		format = visualization.OutputFormatHTML // Default fallback
	}
	
	return &visualization.VisualizationOptions{
		Format:          format,
		ShowLegend:      opts.ShowLegend,
		AnimationSpeed:  opts.AnimationSpeed,
		GIFOptimized:    opts.GIFOptimized,
		ContextChars:    opts.ContextChars,
		MaxTextLength:   opts.MaxTextLength,
		CustomColors:    opts.CustomColors,
		IncludeMetadata: opts.IncludeMetadata,
		SortExtractions: opts.SortExtractions,
		FilterClasses:   opts.FilterClasses,
		Debug:          opts.Debug,
	}
}

// ExportOptions provides options for exporting annotated documents.
type ExportOptions struct {
	// Format specifies the output format
	Format string `json:"format"`
	
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
	
	// SortOrder specifies ascending ("asc") or descending ("desc") sort
	SortOrder string `json:"sort_order,omitempty"`
}

// NewExportOptions creates a new ExportOptions with sensible defaults.
func NewExportOptions() *ExportOptions {
	return &ExportOptions{
		Format:            "json",
		IncludeText:       true,
		IncludeMetadata:   true,
		IncludeAttributes: true,
		IncludePositions:  true,
		Pretty:            true,
		CSVDelimiter:      ",",
		SortOrder:         "asc",
	}
}

// WithFormat sets the export format.
func (opts *ExportOptions) WithFormat(format string) *ExportOptions {
	opts.Format = format
	return opts
}

// WithIncludeText sets whether to include the full document text.
func (opts *ExportOptions) WithIncludeText(include bool) *ExportOptions {
	opts.IncludeText = include
	return opts
}

// WithIncludeMetadata sets whether to include extraction metadata.
func (opts *ExportOptions) WithIncludeMetadata(include bool) *ExportOptions {
	opts.IncludeMetadata = include
	return opts
}

// WithIncludeAttributes sets whether to include extraction attributes.
func (opts *ExportOptions) WithIncludeAttributes(include bool) *ExportOptions {
	opts.IncludeAttributes = include
	return opts
}

// WithIncludePositions sets whether to include character positions.
func (opts *ExportOptions) WithIncludePositions(include bool) *ExportOptions {
	opts.IncludePositions = include
	return opts
}

// WithPretty sets whether to use pretty formatting.
func (opts *ExportOptions) WithPretty(pretty bool) *ExportOptions {
	opts.Pretty = pretty
	return opts
}

// WithCSVDelimiter sets the CSV delimiter.
func (opts *ExportOptions) WithCSVDelimiter(delimiter string) *ExportOptions {
	opts.CSVDelimiter = delimiter
	return opts
}

// WithCSVHeaders sets custom CSV headers.
func (opts *ExportOptions) WithCSVHeaders(headers []string) *ExportOptions {
	opts.CSVHeaders = headers
	return opts
}

// WithFilterClasses sets the extraction classes to filter by.
func (opts *ExportOptions) WithFilterClasses(classes []string) *ExportOptions {
	opts.FilterClasses = classes
	return opts
}

// WithSortBy sets the field to sort by.
func (opts *ExportOptions) WithSortBy(field string) *ExportOptions {
	opts.SortBy = field
	return opts
}

// WithSortOrder sets the sort order.
func (opts *ExportOptions) WithSortOrder(order string) *ExportOptions {
	opts.SortOrder = order
	return opts
}

// Export exports an annotated document to the specified format.
// This provides direct access to the export functionality.
//
// Example usage:
//
//	doc, err := langextract.Extract("John works at Google", opts)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Export as JSON
//	jsonData, err := langextract.Export(doc, NewExportOptions().WithFormat("json"))
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Export as CSV
//	csvData, err := langextract.Export(doc, NewExportOptions().WithFormat("csv"))
//	if err != nil {
//		log.Fatal(err)
//	}
func Export(input interface{}, opts *ExportOptions) ([]byte, error) {
	if opts == nil {
		opts = NewExportOptions()
	}

	// Convert input to AnnotatedDocument
	var doc *document.AnnotatedDocument
	switch v := input.(type) {
	case *document.AnnotatedDocument:
		doc = v
	default:
		return nil, fmt.Errorf("unsupported input type: %T", input)
	}

	if doc == nil {
		return nil, fmt.Errorf("no document to export")
	}

	// Convert options to internal format
	exportOpts := convertToExportOptions(opts)

	// Get the default visualizer
	visualizer := visualization.GetDefaultVisualizer()
	if visualizer == nil {
		return nil, fmt.Errorf("no visualizer available")
	}

	// Map format string to internal format
	var format visualization.OutputFormat
	switch opts.Format {
	case "json":
		format = visualization.OutputFormatJSON
	case "csv":
		format = visualization.OutputFormatCSV
	case "markdown", "md":
		format = visualization.OutputFormatMarkdown
	case "html":
		format = visualization.OutputFormatHTML
	case "text", "txt":
		format = visualization.OutputFormatPlainText
	default:
		return nil, fmt.Errorf("unsupported export format: %s", opts.Format)
	}

	// Export document
	result, err := visualizer.Export(context.Background(), doc, format, exportOpts)
	if err != nil {
		return nil, fmt.Errorf("export failed: %w", err)
	}

	return result, nil
}

// convertToExportOptions converts public export options to internal options
func convertToExportOptions(opts *ExportOptions) *visualization.ExportOptions {
	if opts == nil {
		opts = NewExportOptions()
	}
	
	// Map format string to internal format enum
	var format visualization.OutputFormat
	switch opts.Format {
	case "json":
		format = visualization.OutputFormatJSON
	case "csv":
		format = visualization.OutputFormatCSV
	case "markdown", "md":
		format = visualization.OutputFormatMarkdown
	case "html":
		format = visualization.OutputFormatHTML
	case "text", "txt":
		format = visualization.OutputFormatPlainText
	default:
		format = visualization.OutputFormatJSON // Default fallback
	}
	
	// Map sort order
	var sortOrder visualization.SortOrder
	switch opts.SortOrder {
	case "desc", "descending":
		sortOrder = visualization.SortOrderDesc
	default:
		sortOrder = visualization.SortOrderAsc
	}
	
	return &visualization.ExportOptions{
		Format:            format,
		IncludeText:       opts.IncludeText,
		IncludeMetadata:   opts.IncludeMetadata,
		IncludeAttributes: opts.IncludeAttributes,
		IncludePositions:  opts.IncludePositions,
		Pretty:            opts.Pretty,
		CSVDelimiter:      opts.CSVDelimiter,
		CSVHeaders:        opts.CSVHeaders,
		FilterClasses:     opts.FilterClasses,
		SortBy:            opts.SortBy,
		SortOrder:         sortOrder,
	}
}