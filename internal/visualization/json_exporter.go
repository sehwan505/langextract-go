package visualization

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	"github.com/sehwan505/langextract-go/pkg/document"
	"github.com/sehwan505/langextract-go/pkg/extraction"
)

// JSONExporter implements the Exporter interface for JSON format
type JSONExporter struct {
	options *ExportOptions
}

// NewJSONExporter creates a new JSON exporter
func NewJSONExporter(opts *ExportOptions) *JSONExporter {
	if opts == nil {
		opts = DefaultExportOptions().WithFormat(OutputFormatJSON)
	}
	
	return &JSONExporter{
		options: opts,
	}
}

// Export exports an annotated document to JSON format
func (e *JSONExporter) Export(ctx context.Context, doc *document.AnnotatedDocument, opts *ExportOptions) ([]byte, error) {
	if doc == nil {
		return nil, NewExportError("document cannot be nil", OutputFormatJSON, nil)
	}
	
	// Use provided options or fall back to exporter options
	if opts == nil {
		opts = e.options
	}
	
	if err := opts.Validate(); err != nil {
		return nil, err
	}
	
	// Create export data structure
	exportData := &JSONExportData{
		Format:    string(OutputFormatJSON),
		Version:   "1.0",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Metadata:  make(map[string]interface{}),
	}
	
	// Add document text if requested
	if opts.IncludeText {
		exportData.Text = doc.Text
		exportData.TextLength = len(doc.Text)
	}
	
	// Process extractions
	validExtractions := e.filterValidExtractions(doc.Extractions, opts.FilterClasses)
	if opts.SortBy != "" {
		validExtractions = e.sortExtractions(validExtractions, opts.SortBy, opts.SortOrder)
	}
	
	// Convert extractions to export format
	exportExtractions, err := e.convertExtractions(validExtractions, opts)
	if err != nil {
		return nil, NewExportError("failed to convert extractions", OutputFormatJSON, err)
	}
	
	exportData.Extractions = exportExtractions
	exportData.ExtractionCount = len(exportExtractions)
	
	// Add metadata if requested
	if opts.IncludeMetadata {
		exportData.Metadata = e.buildMetadata(doc, validExtractions)
	}
	
	// Generate statistics
	exportData.Statistics = e.generateStatistics(validExtractions)
	
	// Marshal to JSON
	var jsonData []byte
	
	if opts.Pretty {
		jsonData, err = json.MarshalIndent(exportData, "", "  ")
	} else {
		jsonData, err = json.Marshal(exportData)
	}
	
	if err != nil {
		return nil, NewExportError("failed to marshal JSON", OutputFormatJSON, err)
	}
	
	return jsonData, nil
}

// GetFormat returns the output format this exporter handles
func (e *JSONExporter) GetFormat() OutputFormat {
	return OutputFormatJSON
}

// GetMIMEType returns the MIME type for JSON
func (e *JSONExporter) GetMIMEType() string {
	return "application/json"
}

// Name returns the name of the exporter
func (e *JSONExporter) Name() string {
	return "JSONExporter"
}

// JSONExportData represents the structure of exported JSON data
type JSONExportData struct {
	Format          string                    `json:"format"`
	Version         string                    `json:"version"`
	Timestamp       string                    `json:"timestamp"`
	Text            string                    `json:"text,omitempty"`
	TextLength      int                       `json:"text_length,omitempty"`
	Extractions     []JSONExtractionData      `json:"extractions"`
	ExtractionCount int                       `json:"extraction_count"`
	Statistics      *ExtractionStatistics     `json:"statistics,omitempty"`
	Metadata        map[string]interface{}    `json:"metadata,omitempty"`
}

// JSONExtractionData represents an extraction in JSON format
type JSONExtractionData struct {
	Index      int                    `json:"index,omitempty"`
	Text       string                 `json:"text"`
	Class      string                 `json:"class"`
	StartPos   int                    `json:"start_pos,omitempty"`
	EndPos     int                    `json:"end_pos,omitempty"`
	Length     int                    `json:"length,omitempty"`
	Confidence float64                `json:"confidence,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// ExtractionStatistics provides statistics about the extractions
type ExtractionStatistics struct {
	TotalExtractions  int                    `json:"total_extractions"`
	UniqueClasses     int                    `json:"unique_classes"`
	ClassCounts       map[string]int         `json:"class_counts"`
	AverageLength     float64                `json:"average_length"`
	AverageConfidence float64                `json:"average_confidence,omitempty"`
	TextCoverage      float64                `json:"text_coverage,omitempty"`
	Overlaps          int                    `json:"overlaps,omitempty"`
	Gaps              int                    `json:"gaps,omitempty"`
}

// filterValidExtractions filters extractions based on validity and class filters
func (e *JSONExporter) filterValidExtractions(extractions []*extraction.Extraction, filterClasses []string) []*extraction.Extraction {
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
func (e *JSONExporter) sortExtractions(extractions []*extraction.Extraction, sortBy string, order SortOrder) []*extraction.Extraction {
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
				posI = sorted[i].Interval().StartPos
			}
			if sorted[j].Interval() != nil {
				posJ = sorted[j].Interval().StartPos
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

// convertExtractions converts extractions to JSON export format
func (e *JSONExporter) convertExtractions(extractions []*extraction.Extraction, opts *ExportOptions) ([]JSONExtractionData, error) {
	var jsonExtractions []JSONExtractionData
	
	for i, ext := range extractions {
		jsonExt := JSONExtractionData{
			Text:  ext.Text(),
			Class: ext.Class(),
		}
		
		// Add index if requested
		if i >= 0 {
			jsonExt.Index = i
		}
		
		// Add position information if requested and available
		if opts.IncludePositions && ext.Interval() != nil {
			interval := ext.Interval()
			jsonExt.StartPos = interval.StartPos
			jsonExt.EndPos = interval.EndPos
			jsonExt.Length = interval.EndPos - interval.StartPos
		}
		
		// Add confidence if available
		if ext.Confidence() > 0 {
			jsonExt.Confidence = ext.Confidence()
		}
		
		// Add attributes if requested
		if opts.IncludeAttributes {
			jsonExt.Attributes = e.extractAttributes(ext)
		}
		
		// Add metadata if requested
		if opts.IncludeMetadata {
			jsonExt.Metadata = e.extractMetadata(ext)
		}
		
		jsonExtractions = append(jsonExtractions, jsonExt)
	}
	
	return jsonExtractions, nil
}

// extractAttributes extracts attributes from an extraction
func (e *JSONExporter) extractAttributes(ext *extraction.Extraction) map[string]interface{} {
	// This would be implemented based on the actual extraction interface
	// For now, return empty map
	return make(map[string]interface{})
}

// extractMetadata extracts metadata from an extraction
func (e *JSONExporter) extractMetadata(ext *extraction.Extraction) map[string]interface{} {
	metadata := make(map[string]interface{})
	
	// Add basic information
	metadata["class"] = ext.Class()
	metadata["text_length"] = len(ext.Text())
	
	if ext.Interval() != nil {
		interval := ext.Interval()
		metadata["char_span"] = map[string]int{
			"start": interval.StartPos,
			"end":   interval.EndPos,
		}
	}
	
	if ext.Confidence() > 0 {
		metadata["confidence_score"] = ext.Confidence()
	}
	
	return metadata
}

// buildMetadata builds document-level metadata
func (e *JSONExporter) buildMetadata(doc *document.AnnotatedDocument, extractions []*extraction.Extraction) map[string]interface{} {
	metadata := make(map[string]interface{})
	
	// Document information
	metadata["document_text_length"] = len(doc.Text)
	metadata["total_extractions"] = len(extractions)
	
	// Extraction classes
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
	metadata["extraction_classes"] = classes
	
	// Processing information
	metadata["export_timestamp"] = time.Now().UTC().Format(time.RFC3339)
	metadata["export_format"] = string(OutputFormatJSON)
	
	return metadata
}

// generateStatistics generates statistics about the extractions
func (e *JSONExporter) generateStatistics(extractions []*extraction.Extraction) *ExtractionStatistics {
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
	var totalCoverage int
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
		
		// Coverage statistics (if position information is available)
		if ext.Interval() != nil {
			interval := ext.Interval()
			totalCoverage += interval.EndPos - interval.StartPos
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
	
	// Text coverage would require document text length
	// This is a simplified calculation
	if totalCoverage > 0 {
		stats.TextCoverage = float64(totalCoverage)
	}
	
	// TODO: Implement overlap and gap detection
	stats.Overlaps = 0
	stats.Gaps = 0
	
	return stats
}

// ExportToFile exports a document to a JSON file
func (e *JSONExporter) ExportToFile(ctx context.Context, doc *document.AnnotatedDocument, filename string, opts *ExportOptions) error {
	data, err := e.Export(ctx, doc, opts)
	if err != nil {
		return err
	}
	
	// This would typically write to file system
	// For now, we'll just return nil as file operations are handled externally
	_ = data
	_ = filename
	
	return nil
}

// ExportToString exports a document to a JSON string
func (e *JSONExporter) ExportToString(ctx context.Context, doc *document.AnnotatedDocument, opts *ExportOptions) (string, error) {
	data, err := e.Export(ctx, doc, opts)
	if err != nil {
		return "", err
	}
	
	return string(data), nil
}