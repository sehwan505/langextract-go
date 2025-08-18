package document

import (
	"fmt"
	"sort"

	"github.com/sehwan505/langextract-go/pkg/extraction"
)

// AnnotatedDocument represents a document with extracted entities and annotations.
type AnnotatedDocument struct {
	*Document                            // Embedded document
	Extractions []*extraction.Extraction `json:"extractions,omitempty"` // List of extracted entities
}

// NewAnnotatedDocument creates a new AnnotatedDocument from a Document.
func NewAnnotatedDocument(doc *Document) *AnnotatedDocument {
	return &AnnotatedDocument{
		Document:    doc,
		Extractions: make([]*extraction.Extraction, 0),
	}
}

// NewAnnotatedDocumentWithText creates a new AnnotatedDocument with the given text.
func NewAnnotatedDocumentWithText(text string) *AnnotatedDocument {
	return &AnnotatedDocument{
		Document:    NewDocument(text),
		Extractions: make([]*extraction.Extraction, 0),
	}
}

// AddExtraction adds an extraction to the document.
func (ad *AnnotatedDocument) AddExtraction(ext *extraction.Extraction) {
	if ext == nil {
		return
	}
	ad.Extractions = append(ad.Extractions, ext)
}

// AddExtractions adds multiple extractions to the document.
func (ad *AnnotatedDocument) AddExtractions(extractions []*extraction.Extraction) {
	for _, ext := range extractions {
		ad.AddExtraction(ext)
	}
}

// GetExtractionsByClass returns all extractions of a specific class.
func (ad *AnnotatedDocument) GetExtractionsByClass(class string) []*extraction.Extraction {
	var result []*extraction.Extraction
	for _, ext := range ad.Extractions {
		if ext.ExtractionClass == class {
			result = append(result, ext)
		}
	}
	return result
}

// GetExtractionsByGroup returns all extractions with a specific group index.
func (ad *AnnotatedDocument) GetExtractionsByGroup(groupIndex int) []*extraction.Extraction {
	var result []*extraction.Extraction
	for _, ext := range ad.Extractions {
		if ext.GroupIndex != nil && *ext.GroupIndex == groupIndex {
			result = append(result, ext)
		}
	}
	return result
}

// GetUniqueClasses returns all unique extraction classes in the document.
func (ad *AnnotatedDocument) GetUniqueClasses() []string {
	classSet := make(map[string]bool)
	for _, ext := range ad.Extractions {
		classSet[ext.ExtractionClass] = true
	}

	classes := make([]string, 0, len(classSet))
	for class := range classSet {
		classes = append(classes, class)
	}
	sort.Strings(classes)
	return classes
}

// ExtractionCount returns the total number of extractions.
func (ad *AnnotatedDocument) ExtractionCount() int {
	return len(ad.Extractions)
}

// HasExtractions returns true if the document has any extractions.
func (ad *AnnotatedDocument) HasExtractions() bool {
	return len(ad.Extractions) > 0
}

// SortExtractionsByPosition sorts extractions by their character position in the text.
func (ad *AnnotatedDocument) SortExtractionsByPosition() {
	sort.Slice(ad.Extractions, func(i, j int) bool {
		extI, extJ := ad.Extractions[i], ad.Extractions[j]

		// Handle nil character intervals
		if extI.CharInterval == nil && extJ.CharInterval == nil {
			return false // Keep original order if both are nil
		}
		if extI.CharInterval == nil {
			return false // Put nil intervals at the end
		}
		if extJ.CharInterval == nil {
			return true // Put nil intervals at the end
		}

		return extI.CharInterval.StartPos < extJ.CharInterval.StartPos
	})
}

// SortExtractionsByIndex sorts extractions by their extraction index.
func (ad *AnnotatedDocument) SortExtractionsByIndex() {
	sort.Slice(ad.Extractions, func(i, j int) bool {
		extI, extJ := ad.Extractions[i], ad.Extractions[j]

		// Handle nil extraction indices
		if extI.ExtractionIndex == nil && extJ.ExtractionIndex == nil {
			return false // Keep original order if both are nil
		}
		if extI.ExtractionIndex == nil {
			return false // Put nil indices at the end
		}
		if extJ.ExtractionIndex == nil {
			return true // Put nil indices at the end
		}

		return *extI.ExtractionIndex < *extJ.ExtractionIndex
	})
}

// FilterExtractionsByConfidence returns extractions with confidence above the threshold.
// This assumes extractions have confidence scores in their attributes.
func (ad *AnnotatedDocument) FilterExtractionsByConfidence(threshold float64) []*extraction.Extraction {
	var result []*extraction.Extraction
	for _, ext := range ad.Extractions {
		if ext.Attributes != nil {
			if confVal, exists := ext.Attributes["confidence"]; exists {
				if conf, ok := confVal.(float64); ok && conf >= threshold {
					result = append(result, ext)
				}
			}
		}
	}
	return result
}

// GetCoverage returns the percentage of document text covered by extractions.
func (ad *AnnotatedDocument) GetCoverage() float64 {
	if ad.Document == nil || ad.Document.Length() == 0 {
		return 0.0
	}

	totalCovered := 0
	for _, ext := range ad.Extractions {
		if ext.CharInterval != nil {
			totalCovered += ext.CharInterval.Length()
		}
	}

	return float64(totalCovered) / float64(ad.Document.Length()) * 100.0
}

// String returns a string representation of the annotated document.
func (ad *AnnotatedDocument) String() string {
	if ad.Document == nil {
		return "AnnotatedDocument{document=nil, extractions=0}"
	}

	return fmt.Sprintf("AnnotatedDocument{id=%s, length=%d, extractions=%d, classes=%v}",
		ad.Document.DocumentID(), ad.Document.Length(), len(ad.Extractions), ad.GetUniqueClasses())
}
