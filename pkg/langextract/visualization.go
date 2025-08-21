package langextract

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html"
	"sort"
	"strings"

	"github.com/sehwan505/langextract-go/pkg/document"
	"github.com/sehwan505/langextract-go/pkg/extraction"
)

// generateHTMLVisualization creates an HTML visualization of extractions.
func generateHTMLVisualization(doc *document.AnnotatedDocument, opts *VisualizeOptions) (string, error) {
	var builder strings.Builder

	// HTML header
	builder.WriteString("<!DOCTYPE html>\n<html>\n<head>\n")
	builder.WriteString("<title>LangExtract Results</title>\n")
	builder.WriteString("<style>\n")
	builder.WriteString(getHTMLStyle())
	builder.WriteString("</style>\n")
	builder.WriteString("</head>\n<body>\n")

	// Title
	builder.WriteString("<h1>Extraction Results</h1>\n")

	// Document info
	builder.WriteString("<div class=\"document-info\">\n")
	builder.WriteString(fmt.Sprintf("<p><strong>Document ID:</strong> %s</p>\n", doc.DocumentID()))
	builder.WriteString(fmt.Sprintf("<p><strong>Text Length:</strong> %d characters</p>\n", doc.Length()))
	builder.WriteString(fmt.Sprintf("<p><strong>Total Extractions:</strong> %d</p>\n", doc.ExtractionCount()))
	if doc.ExtractionCount() > 0 {
		builder.WriteString(fmt.Sprintf("<p><strong>Text Coverage:</strong> %.1f%%</p>\n", doc.GetCoverage()))
	}
	builder.WriteString("</div>\n")

	// Get extractions
	extractions := doc.Extractions
	if len(extractions) == 0 {
		builder.WriteString("<p>No extractions found.</p>\n")
		builder.WriteString("</body>\n</html>")
		return builder.String(), nil
	}

	// Sort extractions if requested
	if opts.SortByPosition {
		sort.Slice(extractions, func(i, j int) bool {
			if extractions[i].CharInterval != nil && extractions[j].CharInterval != nil {
				return extractions[i].CharInterval.StartPos < extractions[j].CharInterval.StartPos
			}
			return i < j // Fallback to original order
		})
	}

	// Group by class if requested
	if opts.GroupByClass {
		builder.WriteString(generateGroupedHTML(extractions, opts))
	} else {
		builder.WriteString(generateSequentialHTML(extractions, opts))
	}

	// Highlighted text
	if opts.IncludeContext {
		builder.WriteString("<h2>Highlighted Text</h2>\n")
		highlighted, err := generateHighlightedText(doc, extractions, opts)
		if err != nil {
			return "", fmt.Errorf("failed to generate highlighted text: %w", err)
		}
		builder.WriteString("<div class=\"highlighted-text\">\n")
		builder.WriteString(highlighted)
		builder.WriteString("</div>\n")
	}

	// Legend
	builder.WriteString("<h2>Legend</h2>\n")
	builder.WriteString(generateLegend(extractions))

	builder.WriteString("</body>\n</html>")
	return builder.String(), nil
}

// generateJSONVisualization creates a JSON representation of extractions.
func generateJSONVisualization(doc *document.AnnotatedDocument, opts *VisualizeOptions) (string, error) {
	result := map[string]interface{}{
		"document_id":     doc.DocumentID(),
		"text_length":     doc.Length(),
		"extraction_count": doc.ExtractionCount(),
		"text_coverage":   doc.GetCoverage(),
		"extractions":     formatExtractionsForJSON(doc.Extractions, opts),
	}

	if opts.GroupByClass {
		result["extractions_by_class"] = groupExtractionsByClass(doc.Extractions)
	}

	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return string(jsonBytes), nil
}

// generateCSVVisualization creates a CSV representation of extractions.
func generateCSVVisualization(doc *document.AnnotatedDocument, opts *VisualizeOptions) (string, error) {
	var output strings.Builder
	writer := csv.NewWriter(&output)

	// Header
	headers := []string{"extraction_class", "extraction_text"}
	if opts.ShowConfidence {
		headers = append(headers, "confidence")
	}
	if opts.ShowAlignment {
		headers = append(headers, "char_start", "char_end", "alignment_status")
	}
	headers = append(headers, "attributes")

	if err := writer.Write(headers); err != nil {
		return "", fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Data rows
	for _, ext := range doc.Extractions {
		row := []string{ext.ExtractionClass, ext.ExtractionText}

		if opts.ShowConfidence {
			if conf, ok := ext.GetConfidence(); ok {
				row = append(row, fmt.Sprintf("%.3f", conf))
			} else {
				row = append(row, "")
			}
		}

		if opts.ShowAlignment {
			if ext.CharInterval != nil {
				row = append(row, 
					fmt.Sprintf("%d", ext.CharInterval.StartPos),
					fmt.Sprintf("%d", ext.CharInterval.EndPos))
			} else {
				row = append(row, "", "")
			}

			if ext.AlignmentStatus != nil {
				row = append(row, ext.AlignmentStatus.String())
			} else {
				row = append(row, "")
			}
		}

		// Serialize attributes
		attrs, _ := json.Marshal(ext.Attributes)
		row = append(row, string(attrs))

		if err := writer.Write(row); err != nil {
			return "", fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	writer.Flush()
	return output.String(), nil
}

// getHTMLStyle returns CSS styles for HTML visualization.
func getHTMLStyle() string {
	return `
body {
	font-family: Arial, sans-serif;
	margin: 20px;
	line-height: 1.6;
}

.document-info {
	background: #f5f5f5;
	padding: 15px;
	border-radius: 5px;
	margin-bottom: 20px;
}

.extraction-group {
	margin-bottom: 20px;
	border: 1px solid #ddd;
	border-radius: 5px;
	overflow: hidden;
}

.group-header {
	background: #e9e9e9;
	padding: 10px 15px;
	font-weight: bold;
	border-bottom: 1px solid #ddd;
}

.extraction-item {
	padding: 10px 15px;
	border-bottom: 1px solid #eee;
}

.extraction-item:last-child {
	border-bottom: none;
}

.extraction-text {
	font-weight: bold;
	color: #333;
}

.extraction-meta {
	color: #666;
	font-size: 0.9em;
	margin-top: 5px;
}

.confidence {
	background: #e3f2fd;
	padding: 2px 6px;
	border-radius: 3px;
	font-size: 0.8em;
}

.alignment-exact { color: #4caf50; }
.alignment-fuzzy { color: #ff9800; }
.alignment-partial { color: #f44336; }
.alignment-semantic { color: #9c27b0; }
.alignment-approximate { color: #795548; }
.alignment-none { color: #9e9e9e; }

.highlighted-text {
	background: #fafafa;
	padding: 20px;
	border: 1px solid #ddd;
	border-radius: 5px;
	white-space: pre-wrap;
	font-family: monospace;
	line-height: 1.8;
}

.highlight {
	padding: 2px 4px;
	border-radius: 3px;
	font-weight: bold;
}

.highlight-person { background: #ffcdd2; }
.highlight-organization { background: #c8e6c9; }
.highlight-location { background: #bbdefb; }
.highlight-date { background: #f8bbd9; }
.highlight-money { background: #dcedc8; }
.highlight-default { background: #fff3e0; }

.legend {
	display: flex;
	flex-wrap: wrap;
	gap: 10px;
	margin-top: 10px;
}

.legend-item {
	display: flex;
	align-items: center;
	gap: 5px;
}

.legend-color {
	width: 20px;
	height: 20px;
	border-radius: 3px;
	border: 1px solid #ccc;
}
`
}

// generateGroupedHTML creates HTML grouped by extraction class.
func generateGroupedHTML(extractions []*extraction.Extraction, opts *VisualizeOptions) string {
	var builder strings.Builder

	// Group extractions by class
	groups := make(map[string][]*extraction.Extraction)
	for _, ext := range extractions {
		groups[ext.ExtractionClass] = append(groups[ext.ExtractionClass], ext)
	}

	builder.WriteString("<h2>Extractions by Class</h2>\n")

	for class, classExtractions := range groups {
		builder.WriteString("<div class=\"extraction-group\">\n")
		builder.WriteString(fmt.Sprintf("<div class=\"group-header\">%s (%d)</div>\n", 
			html.EscapeString(class), len(classExtractions)))

		for _, ext := range classExtractions {
			builder.WriteString("<div class=\"extraction-item\">\n")
			builder.WriteString(generateExtractionHTML(ext, opts))
			builder.WriteString("</div>\n")
		}

		builder.WriteString("</div>\n")
	}

	return builder.String()
}

// generateSequentialHTML creates HTML in sequential order.
func generateSequentialHTML(extractions []*extraction.Extraction, opts *VisualizeOptions) string {
	var builder strings.Builder

	builder.WriteString("<h2>All Extractions</h2>\n")
	builder.WriteString("<div class=\"extraction-group\">\n")

	for i, ext := range extractions {
		builder.WriteString(fmt.Sprintf("<div class=\"extraction-item\">\n"))
		builder.WriteString(fmt.Sprintf("<div class=\"extraction-text\">%d. %s</div>\n", 
			i+1, html.EscapeString(ext.ExtractionText)))
		builder.WriteString(fmt.Sprintf("<div class=\"extraction-meta\">Class: %s</div>\n", 
			html.EscapeString(ext.ExtractionClass)))

		if opts.ShowConfidence {
			if conf, ok := ext.GetConfidence(); ok {
				builder.WriteString(fmt.Sprintf("<div class=\"extraction-meta\">Confidence: <span class=\"confidence\">%.1f%%</span></div>\n", 
					conf*100))
			}
		}

		if opts.ShowAlignment && ext.AlignmentStatus != nil {
			builder.WriteString(fmt.Sprintf("<div class=\"extraction-meta\">Alignment: <span class=\"alignment-%s\">%s</span></div>\n", 
				strings.ToLower(ext.AlignmentStatus.String()), ext.AlignmentStatus.String()))
		}

		builder.WriteString("</div>\n")
	}

	builder.WriteString("</div>\n")
	return builder.String()
}

// generateExtractionHTML creates HTML for a single extraction.
func generateExtractionHTML(ext *extraction.Extraction, opts *VisualizeOptions) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("<div class=\"extraction-text\">%s</div>\n", 
		html.EscapeString(ext.ExtractionText)))

	if opts.ShowConfidence {
		if conf, ok := ext.GetConfidence(); ok {
			builder.WriteString(fmt.Sprintf("<div class=\"extraction-meta\">Confidence: <span class=\"confidence\">%.1f%%</span></div>\n", 
				conf*100))
		}
	}

	if opts.ShowAlignment && ext.AlignmentStatus != nil {
		builder.WriteString(fmt.Sprintf("<div class=\"extraction-meta\">Alignment: <span class=\"alignment-%s\">%s</span></div>\n", 
			strings.ToLower(ext.AlignmentStatus.String()), ext.AlignmentStatus.String()))
	}

	return builder.String()
}

// generateHighlightedText creates highlighted version of the source text.
func generateHighlightedText(doc *document.AnnotatedDocument, extractions []*extraction.Extraction, opts *VisualizeOptions) (string, error) {
	text := doc.Text
	if text == "" {
		return "No text available", nil
	}

	// Create highlights for extractions with position information
	highlights := make([]highlight, 0)
	for _, ext := range extractions {
		if ext.CharInterval != nil {
			highlights = append(highlights, highlight{
				Start: ext.CharInterval.StartPos,
				End:   ext.CharInterval.EndPos,
				Class: ext.ExtractionClass,
				Text:  ext.ExtractionText,
			})
		}
	}

	// Sort highlights by position
	sort.Slice(highlights, func(i, j int) bool {
		return highlights[i].Start < highlights[j].Start
	})

	// Apply highlights
	result := applyHighlights(text, highlights)
	return html.EscapeString(result), nil
}

// generateLegend creates a legend for extraction classes.
func generateLegend(extractions []*extraction.Extraction) string {
	var builder strings.Builder

	// Get unique classes
	classes := make(map[string]bool)
	for _, ext := range extractions {
		classes[ext.ExtractionClass] = true
	}

	builder.WriteString("<div class=\"legend\">\n")
	for class := range classes {
		colorClass := getColorClass(class)
		builder.WriteString(fmt.Sprintf("<div class=\"legend-item\">\n"))
		builder.WriteString(fmt.Sprintf("<div class=\"legend-color %s\"></div>\n", colorClass))
		builder.WriteString(fmt.Sprintf("<span>%s</span>\n", html.EscapeString(class)))
		builder.WriteString("</div>\n")
	}
	builder.WriteString("</div>\n")

	return builder.String()
}

// formatExtractionsForJSON formats extractions for JSON output.
func formatExtractionsForJSON(extractions []*extraction.Extraction, opts *VisualizeOptions) []map[string]interface{} {
	result := make([]map[string]interface{}, len(extractions))

	for i, ext := range extractions {
		item := map[string]interface{}{
			"extraction_class": ext.ExtractionClass,
			"extraction_text":  ext.ExtractionText,
		}

		if opts.ShowConfidence {
			if conf, ok := ext.GetConfidence(); ok {
				item["confidence"] = conf
			}
		}

		if opts.ShowAlignment {
			if ext.CharInterval != nil {
				item["char_interval"] = map[string]int{
					"start": ext.CharInterval.StartPos,
					"end":   ext.CharInterval.EndPos,
				}
			}
			if ext.AlignmentStatus != nil {
				item["alignment_status"] = ext.AlignmentStatus.String()
			}
		}

		if len(ext.Attributes) > 0 {
			item["attributes"] = ext.Attributes
		}

		result[i] = item
	}

	return result
}

// groupExtractionsByClass groups extractions by their class.
func groupExtractionsByClass(extractions []*extraction.Extraction) map[string][]map[string]interface{} {
	groups := make(map[string][]map[string]interface{})

	for _, ext := range extractions {
		item := map[string]interface{}{
			"extraction_text": ext.ExtractionText,
		}

		if conf, ok := ext.GetConfidence(); ok {
			item["confidence"] = conf
		}

		if ext.CharInterval != nil {
			item["char_interval"] = map[string]int{
				"start": ext.CharInterval.StartPos,
				"end":   ext.CharInterval.EndPos,
			}
		}

		groups[ext.ExtractionClass] = append(groups[ext.ExtractionClass], item)
	}

	return groups
}

// highlight represents a text highlight.
type highlight struct {
	Start int
	End   int
	Class string
	Text  string
}

// applyHighlights applies highlights to text.
func applyHighlights(text string, highlights []highlight) string {
	if len(highlights) == 0 {
		return text
	}

	var result strings.Builder
	lastPos := 0

	for _, h := range highlights {
		// Add text before highlight
		if h.Start > lastPos {
			result.WriteString(text[lastPos:h.Start])
		}

		// Add highlighted text
		colorClass := getColorClass(h.Class)
		result.WriteString(fmt.Sprintf("<span class=\"highlight %s\">", colorClass))
		if h.End <= len(text) {
			result.WriteString(text[h.Start:h.End])
		}
		result.WriteString("</span>")

		lastPos = h.End
	}

	// Add remaining text
	if lastPos < len(text) {
		result.WriteString(text[lastPos:])
	}

	return result.String()
}

// getColorClass returns a CSS class for highlighting based on extraction class.
func getColorClass(class string) string {
	switch strings.ToLower(class) {
	case "person", "people", "name", "names":
		return "highlight-person"
	case "organization", "org", "company", "corp":
		return "highlight-organization"
	case "location", "place", "city", "country":
		return "highlight-location"
	case "date", "time", "datetime":
		return "highlight-date"
	case "money", "currency", "price", "amount":
		return "highlight-money"
	default:
		return "highlight-default"
	}
}