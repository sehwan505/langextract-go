package extraction

import (
	"encoding/json"
	"fmt"
)

// ExampleData represents a training example with input text and expected extractions.
type ExampleData struct {
	Text        string        `json:"text"`                  // Input text for the example
	Extractions []*Extraction `json:"extractions,omitempty"` // Expected extractions from the text
}

// NewExampleData creates a new ExampleData with the given text.
func NewExampleData(text string) *ExampleData {
	return &ExampleData{
		Text:        text,
		Extractions: make([]*Extraction, 0),
	}
}

// NewExampleDataWithExtractions creates a new ExampleData with text and extractions.
func NewExampleDataWithExtractions(text string, extractions []*Extraction) *ExampleData {
	return &ExampleData{
		Text:        text,
		Extractions: extractions,
	}
}

// AddExtraction adds an extraction to the example.
func (ed *ExampleData) AddExtraction(extraction *Extraction) {
	if extraction == nil {
		return
	}
	ed.Extractions = append(ed.Extractions, extraction)
}

// AddExtractions adds multiple extractions to the example.
func (ed *ExampleData) AddExtractions(extractions []*Extraction) {
	for _, ext := range extractions {
		ed.AddExtraction(ext)
	}
}

// ExtractionCount returns the number of extractions in the example.
func (ed *ExampleData) ExtractionCount() int {
	return len(ed.Extractions)
}

// HasExtractions returns true if the example has any extractions.
func (ed *ExampleData) HasExtractions() bool {
	return len(ed.Extractions) > 0
}

// GetExtractionsByClass returns all extractions of a specific class.
func (ed *ExampleData) GetExtractionsByClass(class string) []*Extraction {
	var result []*Extraction
	for _, ext := range ed.Extractions {
		if ext.ExtractionClass == class {
			result = append(result, ext)
		}
	}
	return result
}

// ToJSON converts the example data to JSON format.
func (ed *ExampleData) ToJSON() ([]byte, error) {
	return json.MarshalIndent(ed, "", "  ")
}

// FromJSON creates ExampleData from JSON bytes.
func FromJSON(data []byte) (*ExampleData, error) {
	var example ExampleData
	err := json.Unmarshal(data, &example)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal example data: %w", err)
	}
	return &example, nil
}

// Validate checks if the example data is valid.
func (ed *ExampleData) Validate() error {
	if ed.Text == "" {
		return fmt.Errorf("example text cannot be empty")
	}
	
	// Validate each extraction
	for i, ext := range ed.Extractions {
		if ext.ExtractionClass == "" {
			return fmt.Errorf("extraction %d: class cannot be empty", i)
		}
		if ext.ExtractionText == "" {
			return fmt.Errorf("extraction %d: text cannot be empty", i)
		}
		
		// Validate character intervals if present
		if ext.CharInterval != nil {
			if ext.CharInterval.StartPos < 0 {
				return fmt.Errorf("extraction %d: start position cannot be negative", i)
			}
			if ext.CharInterval.EndPos > len(ed.Text) {
				return fmt.Errorf("extraction %d: end position exceeds text length", i)
			}
			if ext.CharInterval.StartPos >= ext.CharInterval.EndPos {
				return fmt.Errorf("extraction %d: invalid interval range", i)
			}
		}
	}
	
	return nil
}

// Copy creates a deep copy of the example data.
func (ed *ExampleData) Copy() *ExampleData {
	copy := &ExampleData{
		Text:        ed.Text,
		Extractions: make([]*Extraction, len(ed.Extractions)),
	}
	
	for i, ext := range ed.Extractions {
		copy.Extractions[i] = ext.Copy()
	}
	
	return copy
}

// String returns a string representation of the example data.
func (ed *ExampleData) String() string {
	preview := ed.Text
	if len(preview) > 100 {
		preview = preview[:97] + "..."
	}
	return fmt.Sprintf("ExampleData{text=%q, extractions=%d}", preview, len(ed.Extractions))
}