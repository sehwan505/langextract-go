package visualization

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/sehwan505/langextract-go/pkg/document"
	"github.com/sehwan505/langextract-go/pkg/extraction"
)

// MarkdownExporter implements the Exporter interface for Markdown format
type MarkdownExporter struct {
	options *ExportOptions
}

// NewMarkdownExporter creates a new Markdown exporter
func NewMarkdownExporter(opts *ExportOptions) *MarkdownExporter {
	if opts == nil {
		opts = DefaultExportOptions().WithFormat(OutputFormatMarkdown)
	}
	
	return &MarkdownExporter{
		options: opts,
	}
}

// Export exports an annotated document to Markdown format
func (e *MarkdownExporter) Export(ctx context.Context, doc *document.AnnotatedDocument, opts *ExportOptions) ([]byte, error) {
	if doc == nil {
		return nil, NewExportError("document cannot be nil", OutputFormatMarkdown, nil)
	}
	
	// Use provided options or fall back to exporter options
	if opts == nil {
		opts = e.options
	}
	
	if err := opts.Validate(); err != nil {
		return nil, err
	}
	
	// Filter and sort extractions
	validExtractions := e.filterValidExtractions(doc.Extractions(), opts.FilterClasses)
	if opts.SortBy != "" {
		validExtractions = e.sortExtractions(validExtractions, opts.SortBy, opts.SortOrder)
	}
	
	// Build Markdown content
	var md strings.Builder
	
	// Document header
	e.writeHeader(&md, doc, validExtractions, opts)
	
	// Table of contents (if multiple extraction classes)
	if e.hasMultipleClasses(validExtractions) {
		e.writeTableOfContents(&md, validExtractions)
	}
	
	// Document text with highlights (if requested)
	if opts.IncludeText {
		e.writeHighlightedText(&md, doc.Text(), validExtractions, opts)
	}
	
	// Extraction details
	e.writeExtractionDetails(&md, validExtractions, opts)
	
	// Summary statistics
	e.writeSummary(&md, validExtractions, opts)
	
	// Footer
	e.writeFooter(&md, opts)
	
	return []byte(md.String()), nil
}

// GetFormat returns the output format this exporter handles
func (e *MarkdownExporter) GetFormat() OutputFormat {
	return OutputFormatMarkdown
}

// GetMIMEType returns the MIME type for Markdown
func (e *MarkdownExporter) GetMIMEType() string {
	return "text/markdown"
}

// Name returns the name of the exporter
func (e *MarkdownExporter) Name() string {
	return "MarkdownExporter"
}

// writeHeader writes the document header
func (e *MarkdownExporter) writeHeader(md *strings.Builder, doc *document.AnnotatedDocument, extractions []*extraction.Extraction, opts *ExportOptions) {
	md.WriteString("# LangExtract Extraction Report\n\n")
	
	// Basic information
	md.WriteString(fmt.Sprintf("**Generated:** %s\n\n", time.Now().Format("2006-01-02 15:04:05 UTC")))
	md.WriteString(fmt.Sprintf("**Document Length:** %d characters\n\n", len(doc.Text())))
	md.WriteString(fmt.Sprintf("**Total Extractions:** %d\n\n", len(extractions)))
	
	// Extraction classes summary
	classStats := e.calculateClassStatistics(extractions)
	if len(classStats) > 0 {
		md.WriteString("**Extraction Classes:**\n")
		for _, stat := range classStats {
			md.WriteString(fmt.Sprintf("- **%s**: %d extractions\n", stat.Class, stat.Count))
		}
		md.WriteString("\n")
	}
	
	md.WriteString("---\n\n")
}

// writeTableOfContents writes a table of contents for multiple classes
func (e *MarkdownExporter) writeTableOfContents(md *strings.Builder, extractions []*extraction.Extraction) {
	classes := e.getUniqueClasses(extractions)
	if len(classes) <= 1 {
		return
	}
	
	md.WriteString("## Table of Contents\n\n")
	md.WriteString("- [Highlighted Text](#highlighted-text)\n")
	md.WriteString("- [Extraction Details](#extraction-details)\n")
	
	for _, class := range classes {
		anchor := e.createAnchor(class)
		md.WriteString(fmt.Sprintf("  - [%s](#%s)\n", class, anchor))
	}
	
	md.WriteString("- [Summary Statistics](#summary-statistics)\n\n")
}

// writeHighlightedText writes the document text with highlighted extractions
func (e *MarkdownExporter) writeHighlightedText(md *strings.Builder, text string, extractions []*extraction.Extraction, opts *ExportOptions) {
	md.WriteString("## Highlighted Text\n\n")
	
	if len(extractions) == 0 {
		md.WriteString("```\n")
		md.WriteString(text)
		md.WriteString("\n```\n\n")
		return
	}
	
	// Create highlighted version using markdown bold formatting
	highlightedText := e.createHighlightedMarkdown(text, extractions)
	
	// Split into paragraphs for better readability
	paragraphs := strings.Split(highlightedText, "\n\n")
	for i, paragraph := range paragraphs {
		if strings.TrimSpace(paragraph) != "" {
			md.WriteString(paragraph)
			if i < len(paragraphs)-1 {
				md.WriteString("\n\n")
			}
		}
	}
	
	md.WriteString("\n\n")
}

// writeExtractionDetails writes detailed information about each extraction
func (e *MarkdownExporter) writeExtractionDetails(md *strings.Builder, extractions []*extraction.Extraction, opts *ExportOptions) {
	md.WriteString("## Extraction Details\n\n")
	
	if len(extractions) == 0 {
		md.WriteString("*No extractions found.*\n\n")
		return
	}
	
	// Group by class for better organization
	classByExtractions := e.groupByClass(extractions)
	
	for _, class := range e.getUniqueClasses(extractions) {
		classExtractions := classByExtractions[class]
		anchor := e.createAnchor(class)
		
		md.WriteString(fmt.Sprintf("### %s {#%s}\n\n", class, anchor))
		
		for i, ext := range classExtractions {
			e.writeExtractionItem(md, ext, i+1, opts)
		}
		
		md.WriteString("\n")
	}
}

// writeExtractionItem writes a single extraction item
func (e *MarkdownExporter) writeExtractionItem(md *strings.Builder, ext *extraction.Extraction, index int, opts *ExportOptions) {
	md.WriteString(fmt.Sprintf("#### %d. %s\n\n", index, e.escapeMarkdown(ext.Text())))
	
	// Create a table with extraction details
	md.WriteString("| Property | Value |\n")
	md.WriteString("|----------|-------|\n")
	md.WriteString(fmt.Sprintf("| **Text** | %s |\n", e.escapeMarkdown(ext.Text())))
	md.WriteString(fmt.Sprintf("| **Class** | `%s` |\n", ext.Class()))
	
	// Add position information if available
	if opts.IncludePositions && ext.Interval() != nil {
		interval := ext.Interval()
		md.WriteString(fmt.Sprintf("| **Position** | %d-%d (length: %d) |\n", 
			interval.Start(), interval.End(), interval.End()-interval.Start()))
	}
	
	// Add confidence if available
	if ext.Confidence() > 0 {
		md.WriteString(fmt.Sprintf("| **Confidence** | %.4f |\n", ext.Confidence()))
	}
	
	// Add attributes if requested and available
	if opts.IncludeAttributes {
		attributes := e.extractAttributes(ext)
		if len(attributes) > 0 {
			attributesStr := e.formatMapAsMarkdown(attributes)
			md.WriteString(fmt.Sprintf("| **Attributes** | %s |\n", attributesStr))
		}
	}
	
	// Add metadata if requested
	if opts.IncludeMetadata {
		metadata := e.extractMetadata(ext)
		if len(metadata) > 0 {
			metadataStr := e.formatMapAsMarkdown(metadata)
			md.WriteString(fmt.Sprintf("| **Metadata** | %s |\n", metadataStr))
		}
	}
	
	md.WriteString("\n")
	
	// Add context if available
	if opts.IncludeText && ext.Interval() != nil {
		context := e.getExtractionContext(ext, 50) // 50 chars context
		if context != "" {
			md.WriteString("**Context:**\n")
			md.WriteString("```\n")
			md.WriteString(context)
			md.WriteString("\n```\n\n")
		}
	}
}

// writeSummary writes summary statistics
func (e *MarkdownExporter) writeSummary(md *strings.Builder, extractions []*extraction.Extraction, opts *ExportOptions) {
	md.WriteString("## Summary Statistics\n\n")
	
	if len(extractions) == 0 {
		md.WriteString("*No extractions to summarize.*\n\n")
		return
	}
	
	stats := e.calculateDetailedStatistics(extractions)
	
	md.WriteString("### Overview\n\n")
	md.WriteString("| Metric | Value |\n")
	md.WriteString("|--------|-------|\n")
	md.WriteString(fmt.Sprintf("| **Total Extractions** | %d |\n", stats.TotalExtractions))
	md.WriteString(fmt.Sprintf("| **Unique Classes** | %d |\n", stats.UniqueClasses))
	md.WriteString(fmt.Sprintf("| **Average Length** | %.1f characters |\n", stats.AverageLength))
	
	if stats.AverageConfidence > 0 {
		md.WriteString(fmt.Sprintf("| **Average Confidence** | %.4f |\n", stats.AverageConfidence))
	}
	
	md.WriteString("\n")
	
	// Class breakdown
	if len(stats.ClassCounts) > 1 {
		md.WriteString("### Class Breakdown\n\n")
		md.WriteString("| Class | Count | Percentage |\n")
		md.WriteString("|-------|-------|------------|\n")
		
		// Sort classes by count (descending)
		type classCount struct {
			Class string
			Count int
		}
		
		var classCounts []classCount
		for class, count := range stats.ClassCounts {
			classCounts = append(classCounts, classCount{Class: class, Count: count})
		}
		
		sort.Slice(classCounts, func(i, j int) bool {
			return classCounts[i].Count > classCounts[j].Count
		})
		
		for _, cc := range classCounts {
			percentage := float64(cc.Count) / float64(stats.TotalExtractions) * 100
			md.WriteString(fmt.Sprintf("| `%s` | %d | %.1f%% |\n", cc.Class, cc.Count, percentage))
		}
		
		md.WriteString("\n")
	}
}

// writeFooter writes the document footer
func (e *MarkdownExporter) writeFooter(md *strings.Builder, opts *ExportOptions) {
	md.WriteString("---\n\n")
	md.WriteString("*Generated by [LangExtract-Go](https://github.com/sehwan505/langextract-go)*\n")
}

// Helper methods

// createHighlightedMarkdown creates markdown with highlighted extractions
func (e *MarkdownExporter) createHighlightedMarkdown(text string, extractions []*extraction.Extraction) string {
	if len(extractions) == 0 {
		return text
	}
	
	// Sort extractions by position for proper processing
	sortedExtractions := make([]*extraction.Extraction, len(extractions))
	copy(sortedExtractions, extractions)
	
	sort.Slice(sortedExtractions, func(i, j int) bool {
		if sortedExtractions[i].Interval() == nil || sortedExtractions[j].Interval() == nil {
			return false
		}
		return sortedExtractions[i].Interval().Start() < sortedExtractions[j].Interval().Start()
	})
	
	var result strings.Builder
	cursor := 0
	
	for _, ext := range sortedExtractions {
		if ext.Interval() == nil {
			continue
		}
		
		interval := ext.Interval()
		start := interval.Start()
		end := interval.End()
		
		if start < 0 || end > len(text) || start >= end || start < cursor {
			continue
		}
		
		// Add text before extraction
		if start > cursor {
			result.WriteString(text[cursor:start])
		}
		
		// Add highlighted extraction
		extractedText := text[start:end]
		result.WriteString(fmt.Sprintf("**%s**[^%s]", e.escapeMarkdown(extractedText), ext.Class()))
		
		cursor = end
	}
	
	// Add remaining text
	if cursor < len(text) {
		result.WriteString(text[cursor:])
	}
	
	return result.String()
}

// filterValidExtractions filters extractions based on class filters
func (e *MarkdownExporter) filterValidExtractions(extractions []*extraction.Extraction, filterClasses []string) []*extraction.Extraction {
	if len(extractions) == 0 {
		return nil
	}
	
	// Create filter set if classes are specified
	var filterSet map[string]bool
	if len(filterClasses) > 0 {
		filterSet = make(map[string]bool)
		for _, class := range filterClasses {
			filterSet[class] = true
		}
	}
	
	var validExtractions []*extraction.Extraction
	for _, ext := range extractions {
		if ext == nil {
			continue
		}
		
		// Apply class filter if specified
		if filterSet != nil && !filterSet[ext.Class()] {
			continue
		}
		
		validExtractions = append(validExtractions, ext)
	}
	
	return validExtractions
}

// sortExtractions sorts extractions based on the specified field and order
func (e *MarkdownExporter) sortExtractions(extractions []*extraction.Extraction, sortBy string, order SortOrder) []*extraction.Extraction {
	if len(extractions) == 0 {
		return extractions
	}
	
	sorted := make([]*extraction.Extraction, len(extractions))
	copy(sorted, extractions)
	
	switch sortBy {
	case "position", "start_pos":
		sort.Slice(sorted, func(i, j int) bool {
			posI := 0
			posJ := 0
			
			if sorted[i].Interval() != nil {
				posI = sorted[i].Interval().Start()
			}
			if sorted[j].Interval() != nil {
				posJ = sorted[j].Interval().Start()
			}
			
			if order == SortOrderDesc {
				return posI > posJ
			}
			return posI < posJ
		})
		
	case "class":
		sort.Slice(sorted, func(i, j int) bool {
			if order == SortOrderDesc {
				return sorted[i].Class() > sorted[j].Class()
			}
			return sorted[i].Class() < sorted[j].Class()
		})
		
	case "length":
		sort.Slice(sorted, func(i, j int) bool {
			lengthI := len(sorted[i].Text())
			lengthJ := len(sorted[j].Text())
			
			if order == SortOrderDesc {
				return lengthI > lengthJ
			}
			return lengthI < lengthJ
		})
		
	case "confidence":
		sort.Slice(sorted, func(i, j int) bool {
			confI := sorted[i].Confidence()
			confJ := sorted[j].Confidence()
			
			if order == SortOrderDesc {
				return confI > confJ
			}
			return confI < confJ
		})
	}
	
	return sorted
}

// Utility methods

type ClassStatistic struct {
	Class string
	Count int
}

// calculateClassStatistics calculates basic statistics by class
func (e *MarkdownExporter) calculateClassStatistics(extractions []*extraction.Extraction) []ClassStatistic {
	classCounts := make(map[string]int)
	
	for _, ext := range extractions {
		if ext != nil {
			classCounts[ext.Class()]++
		}
	}
	
	var stats []ClassStatistic
	for class, count := range classCounts {
		stats = append(stats, ClassStatistic{Class: class, Count: count})
	}
	
	// Sort by count descending
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Count > stats[j].Count
	})
	
	return stats
}

// calculateDetailedStatistics calculates detailed statistics
func (e *MarkdownExporter) calculateDetailedStatistics(extractions []*extraction.Extraction) *ExtractionStatistics {
	if len(extractions) == 0 {
		return &ExtractionStatistics{
			TotalExtractions: 0,
			UniqueClasses:    0,
			ClassCounts:      make(map[string]int),
		}
	}
	
	stats := &ExtractionStatistics{
		TotalExtractions: len(extractions),
		ClassCounts:      make(map[string]int),
	}
	
	var totalLength int
	var totalConfidence float64
	var confidenceCount int
	classSet := make(map[string]bool)
	
	for _, ext := range extractions {
		if ext == nil {
			continue
		}
		
		// Class statistics
		class := ext.Class()
		classSet[class] = true
		stats.ClassCounts[class]++
		
		// Length statistics
		textLength := len(ext.Text())
		totalLength += textLength
		
		// Confidence statistics
		if ext.Confidence() > 0 {
			totalConfidence += ext.Confidence()
			confidenceCount++
		}
	}
	
	stats.UniqueClasses = len(classSet)
	
	// Calculate averages
	if len(extractions) > 0 {
		stats.AverageLength = float64(totalLength) / float64(len(extractions))
	}
	
	if confidenceCount > 0 {
		stats.AverageConfidence = totalConfidence / float64(confidenceCount)
	}
	
	return stats
}

// getUniqueClasses returns unique classes sorted alphabetically
func (e *MarkdownExporter) getUniqueClasses(extractions []*extraction.Extraction) []string {
	classSet := make(map[string]bool)
	
	for _, ext := range extractions {
		if ext != nil {
			classSet[ext.Class()] = true
		}
	}
	
	var classes []string
	for class := range classSet {
		classes = append(classes, class)
	}
	
	sort.Strings(classes)
	return classes
}

// hasMultipleClasses checks if extractions have multiple classes
func (e *MarkdownExporter) hasMultipleClasses(extractions []*extraction.Extraction) bool {
	return len(e.getUniqueClasses(extractions)) > 1
}

// groupByClass groups extractions by their class
func (e *MarkdownExporter) groupByClass(extractions []*extraction.Extraction) map[string][]*extraction.Extraction {
	grouped := make(map[string][]*extraction.Extraction)
	
	for _, ext := range extractions {
		if ext != nil {
			class := ext.Class()
			grouped[class] = append(grouped[class], ext)
		}
	}
	
	return grouped
}

// createAnchor creates a markdown anchor from text
func (e *MarkdownExporter) createAnchor(text string) string {
	// Convert to lowercase and replace spaces with hyphens
	anchor := strings.ToLower(text)
	anchor = strings.ReplaceAll(anchor, " ", "-")
	anchor = strings.ReplaceAll(anchor, "_", "-")
	
	// Remove non-alphanumeric characters except hyphens
	var result strings.Builder
	for _, r := range anchor {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	
	return result.String()
}

// escapeMarkdown escapes special markdown characters
func (e *MarkdownExporter) escapeMarkdown(text string) string {
	// Escape common markdown characters
	replacements := map[string]string{
		"*":  "\\*",
		"_":  "\\_",
		"`":  "\\`",
		"#":  "\\#",
		"[":  "\\[",
		"]":  "\\]",
		"(":  "\\(",
		")":  "\\)",
		"|":  "\\|",
		"\\": "\\\\",
	}
	
	result := text
	for char, escaped := range replacements {
		result = strings.ReplaceAll(result, char, escaped)
	}
	
	return result
}

// formatMapAsMarkdown formats a map as markdown
func (e *MarkdownExporter) formatMapAsMarkdown(data map[string]interface{}) string {
	if len(data) == 0 {
		return "*None*"
	}
	
	var parts []string
	
	// Sort keys for consistent output
	var keys []string
	for key := range data {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	
	for _, key := range keys {
		value := data[key]
		valueStr := fmt.Sprintf("%v", value)
		parts = append(parts, fmt.Sprintf("**%s**: %s", e.escapeMarkdown(key), e.escapeMarkdown(valueStr)))
	}
	
	return strings.Join(parts, "<br>")
}

// extractMetadata extracts metadata from an extraction
func (e *MarkdownExporter) extractMetadata(ext *extraction.Extraction) map[string]interface{} {
	metadata := make(map[string]interface{})
	
	// Basic information
	metadata["text_length"] = len(ext.Text())
	
	if ext.Interval() != nil {
		interval := ext.Interval()
		metadata["char_span"] = fmt.Sprintf("%d-%d", interval.Start(), interval.End())
	}
	
	if ext.Confidence() > 0 {
		metadata["confidence"] = fmt.Sprintf("%.4f", ext.Confidence())
	}
	
	return metadata
}

// extractAttributes extracts attributes from an extraction
func (e *MarkdownExporter) extractAttributes(ext *extraction.Extraction) map[string]interface{} {
	// This would be implemented based on the actual extraction interface
	// For now, return empty map
	return make(map[string]interface{})
}

// getExtractionContext gets context around an extraction
func (e *MarkdownExporter) getExtractionContext(ext *extraction.Extraction, contextChars int) string {
	if ext.Interval() == nil {
		return ""
	}
	
	// This would need access to the original document text
	// For now, return empty string
	return ""
}

// ExportToString exports a document to a Markdown string
func (e *MarkdownExporter) ExportToString(ctx context.Context, doc *document.AnnotatedDocument, opts *ExportOptions) (string, error) {
	data, err := e.Export(ctx, doc, opts)
	if err != nil {
		return "", err
	}
	
	return string(data), nil
}