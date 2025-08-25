package extraction

import (
	"fmt"

	"github.com/sehwan505/langextract-go/pkg/types"
)

// Extraction represents a single extracted entity from text with source grounding.
type Extraction struct {
	ExtractionClass string                 `json:"extraction_class"`           // Type/category of the extraction
	ExtractionText  string                 `json:"extraction_text"`            // The actual extracted text
	CharInterval    *types.CharInterval    `json:"char_interval,omitempty"`    // Character position in source text
	TokenInterval   *types.TokenInterval   `json:"token_interval,omitempty"`   // Token position in source text
	AlignmentStatus *types.AlignmentStatus `json:"alignment_status,omitempty"` // Quality of source grounding
	ExtractionIndex *int                   `json:"extraction_index,omitempty"` // Order in extraction results
	GroupIndex      *int                   `json:"group_index,omitempty"`      // Grouping identifier
	Description     *string                `json:"description,omitempty"`      // Human-readable description
	Attributes      map[string]interface{} `json:"attributes,omitempty"`       // Additional metadata
}

// NewExtraction creates a new Extraction with required fields.
func NewExtraction(class, text string) *Extraction {
	return &Extraction{
		ExtractionClass: class,
		ExtractionText:  text,
		Attributes:      make(map[string]interface{}),
	}
}

// NewExtractionWithInterval creates a new Extraction with character interval.
func NewExtractionWithInterval(class, text string, interval *types.CharInterval) *Extraction {
	return &Extraction{
		ExtractionClass: class,
		ExtractionText:  text,
		CharInterval:    interval,
		Attributes:      make(map[string]interface{}),
	}
}

// SetCharInterval sets the character interval for this extraction.
func (e *Extraction) SetCharInterval(interval *types.CharInterval) {
	e.CharInterval = interval
}

// SetTokenInterval sets the token interval for this extraction.
func (e *Extraction) SetTokenInterval(interval *types.TokenInterval) {
	e.TokenInterval = interval
}

// SetAlignmentStatus sets the alignment status for this extraction.
func (e *Extraction) SetAlignmentStatus(status types.AlignmentStatus) {
	e.AlignmentStatus = &status
}

// SetExtractionIndex sets the extraction index.
func (e *Extraction) SetExtractionIndex(index int) {
	e.ExtractionIndex = &index
}

// SetGroupIndex sets the group index.
func (e *Extraction) SetGroupIndex(index int) {
	e.GroupIndex = &index
}

// SetDescription sets the description.
func (e *Extraction) SetDescription(desc string) {
	e.Description = &desc
}

// AddAttribute adds a key-value pair to the attributes map.
func (e *Extraction) AddAttribute(key string, value interface{}) {
	if e.Attributes == nil {
		e.Attributes = make(map[string]interface{})
	}
	e.Attributes[key] = value
}

// GetAttribute retrieves an attribute value by key.
func (e *Extraction) GetAttribute(key string) (interface{}, bool) {
	if e.Attributes == nil {
		return nil, false
	}
	value, exists := e.Attributes[key]
	return value, exists
}

// GetStringAttribute retrieves a string attribute value by key.
func (e *Extraction) GetStringAttribute(key string) (string, bool) {
	value, exists := e.GetAttribute(key)
	if !exists {
		return "", false
	}
	str, ok := value.(string)
	return str, ok
}

// GetFloatAttribute retrieves a float64 attribute value by key.
func (e *Extraction) GetFloatAttribute(key string) (float64, bool) {
	value, exists := e.GetAttribute(key)
	if !exists {
		return 0, false
	}

	// Handle both float64 and float32
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	default:
		return 0, false
	}
}

// GetIntAttribute retrieves an integer attribute value by key.
func (e *Extraction) GetIntAttribute(key string) (int, bool) {
	value, exists := e.GetAttribute(key)
	if !exists {
		return 0, false
	}

	// Handle different integer types
	switch v := value.(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	case int32:
		return int(v), true
	case float64:
		return int(v), true
	default:
		return 0, false
	}
}

// HasCharInterval returns true if the extraction has character position information.
func (e *Extraction) HasCharInterval() bool {
	return e.CharInterval != nil
}

// HasTokenInterval returns true if the extraction has token position information.
func (e *Extraction) HasTokenInterval() bool {
	return e.TokenInterval != nil
}

// IsWellGrounded returns true if the extraction has good source grounding.
func (e *Extraction) IsWellGrounded() bool {
	if e.AlignmentStatus == nil {
		return false
	}
	return e.AlignmentStatus.Quality() >= 60
}

// GetConfidence returns the confidence score if available in attributes.
func (e *Extraction) GetConfidence() (float64, bool) {
	return e.GetFloatAttribute("confidence")
}

// SetConfidence sets the confidence score in attributes.
func (e *Extraction) SetConfidence(confidence float64) {
	e.AddAttribute("confidence", confidence)
}

// Length returns the character length of the extracted text.
func (e *Extraction) Length() int {
	return len(e.ExtractionText)
}

// IsEmpty returns true if the extraction text is empty.
func (e *Extraction) IsEmpty() bool {
	return e.ExtractionText == ""
}

// Copy creates a deep copy of the extraction.
func (e *Extraction) Copy() *Extraction {
	copy := &Extraction{
		ExtractionClass: e.ExtractionClass,
		ExtractionText:  e.ExtractionText,
	}

	if e.CharInterval != nil {
		copy.CharInterval = &types.CharInterval{
			StartPos: e.CharInterval.StartPos,
			EndPos:   e.CharInterval.EndPos,
		}
	}

	if e.TokenInterval != nil {
		copy.TokenInterval = &types.TokenInterval{
			StartToken: e.TokenInterval.StartToken,
			EndToken:   e.TokenInterval.EndToken,
		}
	}

	if e.AlignmentStatus != nil {
		status := *e.AlignmentStatus
		copy.AlignmentStatus = &status
	}

	if e.ExtractionIndex != nil {
		index := *e.ExtractionIndex
		copy.ExtractionIndex = &index
	}

	if e.GroupIndex != nil {
		index := *e.GroupIndex
		copy.GroupIndex = &index
	}

	if e.Description != nil {
		desc := *e.Description
		copy.Description = &desc
	}

	if e.Attributes != nil {
		copy.Attributes = make(map[string]interface{})
		for k, v := range e.Attributes {
			copy.Attributes[k] = v
		}
	}

	return copy
}

// String returns a string representation of the extraction.
func (e *Extraction) String() string {
	var pos string
	if e.CharInterval != nil {
		pos = e.CharInterval.String()
	} else {
		pos = "no-position"
	}

	preview := e.ExtractionText
	if len(preview) > 50 {
		preview = preview[:47] + "..."
	}

	return fmt.Sprintf("Extraction{class=%s, text=%q, pos=%s}",
		e.ExtractionClass, preview, pos)
}

// Class returns the extraction class (alias for ExtractionClass for compatibility)
func (e *Extraction) Class() string {
	return e.ExtractionClass
}

// Text returns the extraction text (alias for ExtractionText for compatibility) 
func (e *Extraction) Text() string {
	return e.ExtractionText
}

// Interval returns the character interval (alias for CharInterval for compatibility)
func (e *Extraction) Interval() *types.CharInterval {
	return e.CharInterval
}

// Confidence returns the confidence score from attributes if available
func (e *Extraction) Confidence() float64 {
	conf, _ := e.GetConfidence()
	return conf
}
