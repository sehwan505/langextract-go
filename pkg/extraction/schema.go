package extraction

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ExtractionSchema defines the structure and constraints for extractions.
type ExtractionSchema interface {
	// GetName returns the schema name/identifier
	GetName() string
	
	// GetDescription returns a human-readable description of the schema
	GetDescription() string
	
	// GetClasses returns the list of extraction classes supported by this schema
	GetClasses() []string
	
	// ValidateExtraction checks if an extraction conforms to this schema
	ValidateExtraction(extraction *Extraction) error
	
	// ToJSONSchema converts the schema to JSON Schema format for LLM prompts
	ToJSONSchema() (map[string]interface{}, error)
}

// FieldDefinition defines constraints for extraction fields.
type FieldDefinition struct {
	Name        string      `json:"name"`                  // Field name
	Type        string      `json:"type"`                  // Data type (string, number, boolean, array)
	Description string      `json:"description,omitempty"` // Field description
	Required    bool        `json:"required,omitempty"`    // Whether field is required
	Enum        []string    `json:"enum,omitempty"`        // Allowed values for enum types
	Pattern     string      `json:"pattern,omitempty"`     // Regex pattern for string validation
	MinLength   *int        `json:"minLength,omitempty"`   // Minimum string length
	MaxLength   *int        `json:"maxLength,omitempty"`   // Maximum string length
	Minimum     *float64    `json:"minimum,omitempty"`     // Minimum numeric value
	Maximum     *float64    `json:"maximum,omitempty"`     // Maximum numeric value
	Default     interface{} `json:"default,omitempty"`     // Default value
}

// ClassDefinition defines an extraction class with its constraints.
type ClassDefinition struct {
	Name        string             `json:"name"`                  // Class name
	Description string             `json:"description,omitempty"` // Class description
	Fields      []*FieldDefinition `json:"fields,omitempty"`      // Custom fields for this class
	Required    bool               `json:"required,omitempty"`    // Whether this class must appear
	MinCount    *int               `json:"minCount,omitempty"`    // Minimum number of instances
	MaxCount    *int               `json:"maxCount,omitempty"`    // Maximum number of instances
}

// BasicExtractionSchema is a concrete implementation of ExtractionSchema.
type BasicExtractionSchema struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Classes     []*ClassDefinition `json:"classes"`
	GlobalFields []*FieldDefinition `json:"globalFields,omitempty"` // Fields available to all classes
}

// NewBasicExtractionSchema creates a new basic extraction schema.
func NewBasicExtractionSchema(name, description string) *BasicExtractionSchema {
	return &BasicExtractionSchema{
		Name:        name,
		Description: description,
		Classes:     make([]*ClassDefinition, 0),
		GlobalFields: make([]*FieldDefinition, 0),
	}
}

// GetName returns the schema name.
func (s *BasicExtractionSchema) GetName() string {
	return s.Name
}

// GetDescription returns the schema description.
func (s *BasicExtractionSchema) GetDescription() string {
	return s.Description
}

// GetClasses returns the list of extraction classes.
func (s *BasicExtractionSchema) GetClasses() []string {
	classes := make([]string, len(s.Classes))
	for i, class := range s.Classes {
		classes[i] = class.Name
	}
	return classes
}

// AddClass adds a class definition to the schema.
func (s *BasicExtractionSchema) AddClass(class *ClassDefinition) {
	if class == nil {
		return
	}
	s.Classes = append(s.Classes, class)
}

// AddGlobalField adds a global field definition to the schema.
func (s *BasicExtractionSchema) AddGlobalField(field *FieldDefinition) {
	if field == nil {
		return
	}
	s.GlobalFields = append(s.GlobalFields, field)
}

// GetClass returns the class definition for the given name.
func (s *BasicExtractionSchema) GetClass(name string) *ClassDefinition {
	for _, class := range s.Classes {
		if class.Name == name {
			return class
		}
	}
	return nil
}

// ValidateExtraction validates an extraction against the schema.
func (s *BasicExtractionSchema) ValidateExtraction(extraction *Extraction) error {
	if extraction == nil {
		return fmt.Errorf("extraction cannot be nil")
	}
	
	// Check if class is defined in schema
	classDef := s.GetClass(extraction.ExtractionClass)
	if classDef == nil {
		return fmt.Errorf("unknown extraction class: %s", extraction.ExtractionClass)
	}
	
	// Validate extraction text
	if extraction.ExtractionText == "" {
		return fmt.Errorf("extraction text cannot be empty")
	}
	
	// Validate attributes against field definitions
	if err := s.validateAttributes(extraction, classDef); err != nil {
		return fmt.Errorf("attribute validation failed: %w", err)
	}
	
	return nil
}

// validateAttributes validates extraction attributes against field definitions.
func (s *BasicExtractionSchema) validateAttributes(extraction *Extraction, classDef *ClassDefinition) error {
	// Combine global fields and class-specific fields
	allFields := make([]*FieldDefinition, 0, len(s.GlobalFields)+len(classDef.Fields))
	allFields = append(allFields, s.GlobalFields...)
	allFields = append(allFields, classDef.Fields...)
	
	// Check required fields
	for _, field := range allFields {
		if field.Required {
			if extraction.Attributes == nil {
				return fmt.Errorf("required field %s is missing", field.Name)
			}
			if _, exists := extraction.Attributes[field.Name]; !exists {
				return fmt.Errorf("required field %s is missing", field.Name)
			}
		}
	}
	
	// Validate each attribute
	if extraction.Attributes != nil {
		for attrName, attrValue := range extraction.Attributes {
			field := s.findField(attrName, allFields)
			if field != nil {
				if err := s.validateFieldValue(field, attrValue); err != nil {
					return fmt.Errorf("field %s: %w", attrName, err)
				}
			}
		}
	}
	
	return nil
}

// findField finds a field definition by name.
func (s *BasicExtractionSchema) findField(name string, fields []*FieldDefinition) *FieldDefinition {
	for _, field := range fields {
		if field.Name == name {
			return field
		}
	}
	return nil
}

// validateFieldValue validates a field value against its definition.
func (s *BasicExtractionSchema) validateFieldValue(field *FieldDefinition, value interface{}) error {
	switch field.Type {
	case "string":
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
		return s.validateStringField(field, str)
	case "number":
		return s.validateNumberField(field, value)
	case "boolean":
		_, ok := value.(bool)
		if !ok {
			return fmt.Errorf("expected boolean, got %T", value)
		}
	case "array":
		_, ok := value.([]interface{})
		if !ok {
			return fmt.Errorf("expected array, got %T", value)
		}
	}
	return nil
}

// validateStringField validates string field constraints.
func (s *BasicExtractionSchema) validateStringField(field *FieldDefinition, value string) error {
	if field.MinLength != nil && len(value) < *field.MinLength {
		return fmt.Errorf("string too short: %d < %d", len(value), *field.MinLength)
	}
	if field.MaxLength != nil && len(value) > *field.MaxLength {
		return fmt.Errorf("string too long: %d > %d", len(value), *field.MaxLength)
	}
	if len(field.Enum) > 0 {
		found := false
		for _, allowed := range field.Enum {
			if value == allowed {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("value %q not in allowed values: %v", value, field.Enum)
		}
	}
	return nil
}

// validateNumberField validates numeric field constraints.
func (s *BasicExtractionSchema) validateNumberField(field *FieldDefinition, value interface{}) error {
	var num float64
	switch v := value.(type) {
	case float64:
		num = v
	case float32:
		num = float64(v)
	case int:
		num = float64(v)
	case int64:
		num = float64(v)
	default:
		return fmt.Errorf("expected number, got %T", value)
	}
	
	if field.Minimum != nil && num < *field.Minimum {
		return fmt.Errorf("number too small: %f < %f", num, *field.Minimum)
	}
	if field.Maximum != nil && num > *field.Maximum {
		return fmt.Errorf("number too large: %f > %f", num, *field.Maximum)
	}
	return nil
}

// ToJSONSchema converts the schema to JSON Schema format.
func (s *BasicExtractionSchema) ToJSONSchema() (map[string]interface{}, error) {
	schema := map[string]interface{}{
		"$schema":     "http://json-schema.org/draft-07/schema#",
		"type":        "object",
		"title":       s.Name,
		"description": s.Description,
		"properties": map[string]interface{}{
			"extractions": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"extraction_class": map[string]interface{}{
							"type": "string",
							"enum": s.GetClasses(),
						},
						"extraction_text": map[string]interface{}{
							"type": "string",
						},
					},
					"required": []string{"extraction_class", "extraction_text"},
				},
			},
		},
		"required": []string{"extractions"},
	}
	
	return schema, nil
}

// ToJSON converts the schema to JSON format.
func (s *BasicExtractionSchema) ToJSON() ([]byte, error) {
	return json.MarshalIndent(s, "", "  ")
}

// FromJSON creates a schema from JSON bytes.
func SchemaFromJSON(data []byte) (*BasicExtractionSchema, error) {
	var schema BasicExtractionSchema
	err := json.Unmarshal(data, &schema)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal schema: %w", err)
	}
	return &schema, nil
}

// ExtractionTask combines a schema with examples and prompt information.
type ExtractionTask struct {
	Schema      ExtractionSchema `json:"schema"`                // Schema defining expected extractions
	Examples    []*ExampleData   `json:"examples,omitempty"`    // Few-shot learning examples
	Prompt      string           `json:"prompt"`                // Base extraction prompt
	Temperature float64          `json:"temperature,omitempty"` // LLM temperature setting
	MaxTokens   int              `json:"max_tokens,omitempty"`  // Maximum tokens in response
}

// NewExtractionTask creates a new extraction task.
func NewExtractionTask(schema ExtractionSchema, prompt string) *ExtractionTask {
	return &ExtractionTask{
		Schema:   schema,
		Examples: make([]*ExampleData, 0),
		Prompt:   prompt,
	}
}

// AddExample adds an example to the task.
func (task *ExtractionTask) AddExample(example *ExampleData) {
	if example == nil {
		return
	}
	task.Examples = append(task.Examples, example)
}

// Validate validates the extraction task.
func (task *ExtractionTask) Validate() error {
	if task.Schema == nil {
		return fmt.Errorf("schema cannot be nil")
	}
	if strings.TrimSpace(task.Prompt) == "" {
		return fmt.Errorf("prompt cannot be empty")
	}
	
	// Validate examples
	for i, example := range task.Examples {
		if err := example.Validate(); err != nil {
			return fmt.Errorf("example %d: %w", i, err)
		}
		
		// Validate extractions against schema
		for j, extraction := range example.Extractions {
			if err := task.Schema.ValidateExtraction(extraction); err != nil {
				return fmt.Errorf("example %d, extraction %d: %w", i, j, err)
			}
		}
	}
	
	return nil
}

// String returns a string representation of the extraction task.
func (task *ExtractionTask) String() string {
	return fmt.Sprintf("ExtractionTask{schema=%s, examples=%d, prompt_len=%d}", 
		task.Schema.GetName(), len(task.Examples), len(task.Prompt))
}