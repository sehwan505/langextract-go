package visualization

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/sehwan505/langextract-go/pkg/document"
	"github.com/sehwan505/langextract-go/pkg/extraction"
)

// CSVExporter implements the Exporter interface for CSV format
type CSVExporter struct {
	options *ExportOptions
}

// NewCSVExporter creates a new CSV exporter
func NewCSVExporter(opts *ExportOptions) *CSVExporter {
	if opts == nil {
		opts = DefaultExportOptions().WithFormat(OutputFormatCSV)
	}
	
	// Ensure CSV-specific defaults
	if opts.CSVDelimiter == "" {
		opts.CSVDelimiter = ","
	}
	
	return &CSVExporter{
		options: opts,
	}
}

// Export exports an annotated document to CSV format
func (e *CSVExporter) Export(ctx context.Context, doc *document.AnnotatedDocument, opts *ExportOptions) ([]byte, error) {
	if doc == nil {
		return nil, NewExportError("document cannot be nil", OutputFormatCSV, nil)
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
	
	// Prepare CSV data
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	
	// Set custom delimiter if specified
	if len(opts.CSVDelimiter) == 1 {
		writer.Comma = rune(opts.CSVDelimiter[0])
	}
	
	// Write headers
	headers := e.buildHeaders(opts)
	if err := writer.Write(headers); err != nil {
		return nil, NewExportError("failed to write CSV headers", OutputFormatCSV, err)
	}
	
	// Write extraction data
	for i, ext := range validExtractions {
		record := e.buildRecord(ext, i, doc, opts)
		if err := writer.Write(record); err != nil {
			return nil, NewExportError("failed to write CSV record", OutputFormatCSV, err)
		}
	}
	
	writer.Flush()
	
	if err := writer.Error(); err != nil {
		return nil, NewExportError("CSV writer error", OutputFormatCSV, err)
	}
	
	return buf.Bytes(), nil
}

// GetFormat returns the output format this exporter handles
func (e *CSVExporter) GetFormat() OutputFormat {
	return OutputFormatCSV
}

// GetMIMEType returns the MIME type for CSV
func (e *CSVExporter) GetMIMEType() string {
	return "text/csv"
}

// Name returns the name of the exporter
func (e *CSVExporter) Name() string {
	return "CSVExporter"
}

// buildHeaders builds the CSV headers based on export options
func (e *CSVExporter) buildHeaders(opts *ExportOptions) []string {
	// Use custom headers if provided
	if len(opts.CSVHeaders) > 0 {
		return opts.CSVHeaders
	}
	
	// Build default headers based on options
	var headers []string
	
	// Always include basic fields
	headers = append(headers, "index", "text", "class")
	
	// Add position fields if requested
	if opts.IncludePositions {
		headers = append(headers, "start_pos", "end_pos", "length")
	}
	
	// Add confidence if requested
	headers = append(headers, "confidence")
	
	// Add metadata fields if requested
	if opts.IncludeMetadata {
		headers = append(headers, "metadata")
	}
	
	// Add attributes if requested
	if opts.IncludeAttributes {
		headers = append(headers, "attributes")
	}
	
	// Additional useful fields
	headers = append(headers, "char_count", "word_count")
	
	return headers
}

// buildRecord builds a CSV record for an extraction
func (e *CSVExporter) buildRecord(ext *extraction.Extraction, index int, doc *document.AnnotatedDocument, opts *ExportOptions) []string {
	var record []string
	
	// Index
	record = append(record, strconv.Itoa(index))
	
	// Text (escaped for CSV)
	record = append(record, e.escapeCSVField(ext.Text()))
	
	// Class
	record = append(record, e.escapeCSVField(ext.Class()))
	
	// Position information
	if opts.IncludePositions {
		if ext.Interval() != nil {
			interval := ext.Interval()
			record = append(record, 
				strconv.Itoa(interval.Start()),
				strconv.Itoa(interval.End()),
				strconv.Itoa(interval.End()-interval.Start()))
		} else {
			record = append(record, "", "", "")
		}
	}
	
	// Confidence
	if ext.Confidence() > 0 {
		record = append(record, fmt.Sprintf("%.4f", ext.Confidence()))
	} else {
		record = append(record, "")
	}
	
	// Metadata
	if opts.IncludeMetadata {
		metadata := e.extractMetadata(ext)
		metadataStr := e.formatMapAsCSVField(metadata)
		record = append(record, metadataStr)
	}
	
	// Attributes
	if opts.IncludeAttributes {
		attributes := e.extractAttributes(ext)
		attributesStr := e.formatMapAsCSVField(attributes)
		record = append(record, attributesStr)
	}
	
	// Additional fields
	text := ext.Text()
	record = append(record, 
		strconv.Itoa(len(text)),                    // char_count
		strconv.Itoa(len(strings.Fields(text))))    // word_count
	
	return record
}

// escapeCSVField escapes a field for CSV output
func (e *CSVExporter) escapeCSVField(field string) string {
	// Remove newlines and tabs for cleaner CSV
	cleaned := strings.ReplaceAll(field, "\n", " ")
	cleaned = strings.ReplaceAll(cleaned, "\t", " ")
	cleaned = strings.ReplaceAll(cleaned, "\r", " ")
	
	// Trim extra spaces
	cleaned = strings.TrimSpace(cleaned)
	
	return cleaned
}

// formatMapAsCSVField formats a map as a CSV field
func (e *CSVExporter) formatMapAsCSVField(data map[string]interface{}) string {
	if len(data) == 0 {
		return ""
	}
	
	var pairs []string
	
	// Sort keys for consistent output
	var keys []string
	for key := range data {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	
	for _, key := range keys {
		value := data[key]
		valueStr := fmt.Sprintf("%v", value)
		
		// Clean the value string for CSV
		valueStr = e.escapeCSVField(valueStr)
		
		pairs = append(pairs, fmt.Sprintf("%s=%s", key, valueStr))
	}
	
	return strings.Join(pairs, "; ")
}

// filterValidExtractions filters extractions based on class filters
func (e *CSVExporter) filterValidExtractions(extractions []*extraction.Extraction, filterClasses []string) []*extraction.Extraction {
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
func (e *CSVExporter) sortExtractions(extractions []*extraction.Extraction, sortBy string, order SortOrder) []*extraction.Extraction {
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
		
	case "text":
		sort.Slice(sorted, func(i, j int) bool {
			if order == SortOrderDesc {
				return sorted[i].Text() > sorted[j].Text()
			}
			return sorted[i].Text() < sorted[j].Text()
		})
	}
	
	return sorted
}

// extractMetadata extracts metadata from an extraction for CSV export
func (e *CSVExporter) extractMetadata(ext *extraction.Extraction) map[string]interface{} {
	metadata := make(map[string]interface{})
	
	// Basic information
	metadata["class"] = ext.Class()
	metadata["text_length"] = len(ext.Text())
	
	if ext.Interval() != nil {
		interval := ext.Interval()
		metadata["start"] = interval.Start()
		metadata["end"] = interval.End()
	}
	
	if ext.Confidence() > 0 {
		metadata["confidence"] = fmt.Sprintf("%.4f", ext.Confidence())
	}
	
	return metadata
}

// extractAttributes extracts attributes from an extraction for CSV export
func (e *CSVExporter) extractAttributes(ext *extraction.Extraction) map[string]interface{} {
	// This would be implemented based on the actual extraction interface
	// For now, return empty map
	return make(map[string]interface{})
}

// ExportWithCustomHeaders exports with custom column headers
func (e *CSVExporter) ExportWithCustomHeaders(ctx context.Context, doc *document.AnnotatedDocument, headers []string, opts *ExportOptions) ([]byte, error) {
	if opts == nil {
		opts = DefaultExportOptions().WithFormat(OutputFormatCSV)
	}
	
	// Override headers
	opts.CSVHeaders = headers
	
	return e.Export(ctx, doc, opts)
}

// ExportSummary exports a summary CSV with aggregated statistics
func (e *CSVExporter) ExportSummary(ctx context.Context, doc *document.AnnotatedDocument, opts *ExportOptions) ([]byte, error) {
	if doc == nil {
		return nil, NewExportError("document cannot be nil", OutputFormatCSV, nil)
	}
	
	if opts == nil {
		opts = e.options
	}
	
	validExtractions := e.filterValidExtractions(doc.Extractions(), opts.FilterClasses)
	
	// Group extractions by class
	classCounts := make(map[string]int)
	classLengths := make(map[string][]int)
	classConfidences := make(map[string][]float64)
	
	for _, ext := range validExtractions {
		if ext == nil {
			continue
		}
		
		class := ext.Class()
		classCounts[class]++
		classLengths[class] = append(classLengths[class], len(ext.Text()))
		
		if ext.Confidence() > 0 {
			classConfidences[class] = append(classConfidences[class], ext.Confidence())
		}
	}
	
	// Build summary CSV
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	
	if len(opts.CSVDelimiter) == 1 {
		writer.Comma = rune(opts.CSVDelimiter[0])
	}
	
	// Write headers
	summaryHeaders := []string{"class", "count", "avg_length", "min_length", "max_length", "avg_confidence"}
	if err := writer.Write(summaryHeaders); err != nil {
		return nil, NewExportError("failed to write summary CSV headers", OutputFormatCSV, err)
	}
	
	// Sort classes for consistent output
	var classes []string
	for class := range classCounts {
		classes = append(classes, class)
	}
	sort.Strings(classes)
	
	// Write summary data
	for _, class := range classes {
		count := classCounts[class]
		lengths := classLengths[class]
		confidences := classConfidences[class]
		
		// Calculate statistics
		avgLength := 0.0
		minLength := 0
		maxLength := 0
		
		if len(lengths) > 0 {
			total := 0
			minLength = lengths[0]
			maxLength = lengths[0]
			
			for _, length := range lengths {
				total += length
				if length < minLength {
					minLength = length
				}
				if length > maxLength {
					maxLength = length
				}
			}
			
			avgLength = float64(total) / float64(len(lengths))
		}
		
		avgConfidence := 0.0
		if len(confidences) > 0 {
			total := 0.0
			for _, conf := range confidences {
				total += conf
			}
			avgConfidence = total / float64(len(confidences))
		}
		
		record := []string{
			class,
			strconv.Itoa(count),
			fmt.Sprintf("%.2f", avgLength),
			strconv.Itoa(minLength),
			strconv.Itoa(maxLength),
			fmt.Sprintf("%.4f", avgConfidence),
		}
		
		if err := writer.Write(record); err != nil {
			return nil, NewExportError("failed to write summary CSV record", OutputFormatCSV, err)
		}
	}
	
	writer.Flush()
	
	if err := writer.Error(); err != nil {
		return nil, NewExportError("summary CSV writer error", OutputFormatCSV, err)
	}
	
	return buf.Bytes(), nil
}

// ExportToString exports a document to a CSV string
func (e *CSVExporter) ExportToString(ctx context.Context, doc *document.AnnotatedDocument, opts *ExportOptions) (string, error) {
	data, err := e.Export(ctx, doc, opts)
	if err != nil {
		return "", err
	}
	
	return string(data), nil
}