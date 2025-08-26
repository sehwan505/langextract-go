package extraction_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/sehwan505/langextract-go/pkg/extraction"
)

// TestBasicExtractionSchema tests the basic schema implementation
// Following patterns from Python schema_test.py
func TestBasicExtractionSchema(t *testing.T) {
	schema := extraction.NewBasicExtractionSchema("test_schema", "Test schema for unit testing")

	if schema.GetName() != "test_schema" {
		t.Errorf("GetName() = %q, want 'test_schema'", schema.GetName())
	}

	if schema.GetDescription() != "Test schema for unit testing" {
		t.Errorf("GetDescription() = %q, want 'Test schema for unit testing'", schema.GetDescription())
	}

	// Initially, no classes should exist
	classes := schema.GetClasses()
	if len(classes) != 0 {
		t.Errorf("GetClasses() length = %d, want 0", len(classes))
	}
}

// TestSchemaClassManagement tests adding and retrieving classes
func TestSchemaClassManagement(t *testing.T) {
	schema := extraction.NewBasicExtractionSchema("test_schema", "Test schema")

	// Add first class
	personClass := &extraction.ClassDefinition{
		Name:        "person",
		Description: "Person entity",
		Required:    true,
	}
	schema.AddClass(personClass)

	// Add second class
	locationClass := &extraction.ClassDefinition{
		Name:        "location",
		Description: "Location entity",
		Required:    false,
	}
	schema.AddClass(locationClass)

	// Test GetClasses
	classes := schema.GetClasses()
	if len(classes) != 2 {
		t.Errorf("GetClasses() length = %d, want 2", len(classes))
	}

	expectedClasses := []string{"person", "location"}
	for i, expected := range expectedClasses {
		if classes[i] != expected {
			t.Errorf("GetClasses()[%d] = %q, want %q", i, classes[i], expected)
		}
	}

	// Test GetClass
	retrieved := schema.GetClass("person")
	if retrieved == nil {
		t.Error("GetClass('person') returned nil")
	} else if retrieved.Name != "person" {
		t.Errorf("GetClass('person').Name = %q, want 'person'", retrieved.Name)
	}

	// Test non-existent class
	nonExistent := schema.GetClass("non_existent")
	if nonExistent != nil {
		t.Error("GetClass('non_existent') should return nil")
	}
}

// TestSchemaFieldDefinitions tests field definition handling
func TestSchemaFieldDefinitions(t *testing.T) {
	schema := extraction.NewBasicExtractionSchema("test_schema", "Test schema")

	// Add global field
	confidenceField := &extraction.FieldDefinition{
		Name:        "confidence",
		Type:        "number",
		Description: "Extraction confidence score",
		Required:    false,
		Minimum:     ptrFloat64(0.0),
		Maximum:     ptrFloat64(1.0),
	}
	schema.AddGlobalField(confidenceField)

	// Add class with specific fields
	personClass := &extraction.ClassDefinition{
		Name:        "person",
		Description: "Person entity",
		Fields: []*extraction.FieldDefinition{
			{
				Name:        "age",
				Type:        "number",
				Description: "Person's age",
				Required:    false,
				Minimum:     ptrFloat64(0.0),
				Maximum:     ptrFloat64(150.0),
			},
			{
				Name:        "gender",
				Type:        "string",
				Description: "Person's gender",
				Required:    false,
				Enum:        []string{"male", "female", "other"},
			},
		},
	}
	schema.AddClass(personClass)

	// Test field validation scenarios
	tests := []struct {
		name        string
		extraction  *extraction.Extraction
		expectError bool
		errorSubstr string
	}{
		{
			name: "valid extraction with all fields",
			extraction: createTestExtraction("person", "John Doe", map[string]interface{}{
				"age":        30,
				"gender":     "male",
				"confidence": 0.95,
			}),
			expectError: false,
		},
		{
			name: "valid extraction with minimal fields",
			extraction: createTestExtraction("person", "Jane Smith", map[string]interface{}{
				"confidence": 0.8,
			}),
			expectError: false,
		},
		{
			name: "invalid extraction with out-of-range age",
			extraction: createTestExtraction("person", "Old Person", map[string]interface{}{
				"age": 200,
			}),
			expectError: true,
			errorSubstr: "too large",
		},
		{
			name: "invalid extraction with invalid gender enum",
			extraction: createTestExtraction("person", "Person", map[string]interface{}{
				"gender": "invalid_gender",
			}),
			expectError: true,
			errorSubstr: "not in allowed values",
		},
		{
			name: "invalid extraction with wrong type",
			extraction: createTestExtraction("person", "Person", map[string]interface{}{
				"age": "thirty",
			}),
			expectError: true,
			errorSubstr: "expected number",
		},
		{
			name: "invalid extraction with unknown class",
			extraction: createTestExtraction("unknown_class", "Entity", map[string]interface{}{
				"confidence": 0.9,
			}),
			expectError: true,
			errorSubstr: "unknown extraction class",
		},
		{
			name:        "nil extraction",
			extraction:  nil,
			expectError: true,
			errorSubstr: "cannot be nil",
		},
		{
			name:        "empty extraction text",
			extraction:  createTestExtraction("person", "", map[string]interface{}{}),
			expectError: true,
			errorSubstr: "cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := schema.ValidateExtraction(tt.extraction)

			if tt.expectError {
				if err == nil {
					t.Errorf("ValidateExtraction() error = nil, want error containing %q", tt.errorSubstr)
				} else if !strings.Contains(err.Error(), tt.errorSubstr) {
					t.Errorf("ValidateExtraction() error = %q, want error containing %q", err.Error(), tt.errorSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateExtraction() error = %v, want nil", err)
				}
			}
		})
	}
}

// TestSchemaJSONSchemaGeneration tests JSON Schema generation
// Following patterns from Python schema_test.py parameterized tests
func TestSchemaJSONSchemaGeneration(t *testing.T) {
	tests := []struct {
		name           string
		setupSchema    func() *extraction.BasicExtractionSchema
		expectedSchema map[string]interface{}
	}{
		{
			name: "empty schema",
			setupSchema: func() *extraction.BasicExtractionSchema {
				return extraction.NewBasicExtractionSchema("empty_schema", "Empty test schema")
			},
			expectedSchema: map[string]interface{}{
				"$schema":     "http://json-schema.org/draft-07/schema#",
				"type":        "object",
				"title":       "empty_schema",
				"description": "Empty test schema",
				"properties": map[string]interface{}{
					"extractions": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"extraction_class": map[string]interface{}{
									"type": "string",
									"enum": []string{},
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
			},
		},
		{
			name: "single class schema",
			setupSchema: func() *extraction.BasicExtractionSchema {
				schema := extraction.NewBasicExtractionSchema("single_class", "Single class schema")
				schema.AddClass(&extraction.ClassDefinition{
					Name:        "person",
					Description: "Person entity",
				})
				return schema
			},
			expectedSchema: map[string]interface{}{
				"$schema":     "http://json-schema.org/draft-07/schema#",
				"type":        "object",
				"title":       "single_class",
				"description": "Single class schema",
				"properties": map[string]interface{}{
					"extractions": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"extraction_class": map[string]interface{}{
									"type": "string",
									"enum": []string{"person"},
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
			},
		},
		{
			name: "multiple classes schema",
			setupSchema: func() *extraction.BasicExtractionSchema {
				schema := extraction.NewBasicExtractionSchema("multi_class", "Multi class schema")
				schema.AddClass(&extraction.ClassDefinition{
					Name:        "person",
					Description: "Person entity",
				})
				schema.AddClass(&extraction.ClassDefinition{
					Name:        "location",
					Description: "Location entity",
				})
				return schema
			},
			expectedSchema: map[string]interface{}{
				"$schema":     "http://json-schema.org/draft-07/schema#",
				"type":        "object",
				"title":       "multi_class",
				"description": "Multi class schema",
				"properties": map[string]interface{}{
					"extractions": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"extraction_class": map[string]interface{}{
									"type": "string",
									"enum": []string{"person", "location"},
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
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := tt.setupSchema()
			
			jsonSchema, err := schema.ToJSONSchema()
			if err != nil {
				t.Fatalf("ToJSONSchema() error = %v", err)
			}

			// Compare JSON representations for easier debugging
			expectedJSON, err := json.MarshalIndent(tt.expectedSchema, "", "  ")
			if err != nil {
				t.Fatalf("Failed to marshal expected schema: %v", err)
			}

			actualJSON, err := json.MarshalIndent(jsonSchema, "", "  ")
			if err != nil {
				t.Fatalf("Failed to marshal actual schema: %v", err)
			}

			if string(expectedJSON) != string(actualJSON) {
				t.Errorf("ToJSONSchema() mismatch:\nExpected:\n%s\n\nActual:\n%s", 
					string(expectedJSON), string(actualJSON))
			}
		})
	}
}

// TestExtractionTaskValidation tests extraction task validation
func TestExtractionTaskValidation(t *testing.T) {
	// Create a valid schema
	schema := extraction.NewBasicExtractionSchema("test_schema", "Test schema")
	schema.AddClass(&extraction.ClassDefinition{
		Name:        "person",
		Description: "Person entity",
		Fields: []*extraction.FieldDefinition{
			{
				Name:     "name",
				Type:     "string",
				Required: true,
			},
		},
	})

	tests := []struct {
		name        string
		task        *extraction.ExtractionTask
		expectError bool
		errorSubstr string
	}{
		{
			name: "valid task",
			task: &extraction.ExtractionTask{
				Schema: schema,
				Prompt: "Extract person entities",
				Examples: []*extraction.ExampleData{
					{
						Text: "John is a developer",
						Extractions: []*extraction.Extraction{
							createTestExtraction("person", "John", map[string]interface{}{
								"name": "John",
							}),
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "nil schema",
			task: &extraction.ExtractionTask{
				Schema: nil,
				Prompt: "Extract entities",
			},
			expectError: true,
			errorSubstr: "schema cannot be nil",
		},
		{
			name: "empty prompt",
			task: &extraction.ExtractionTask{
				Schema: schema,
				Prompt: "   ",
			},
			expectError: true,
			errorSubstr: "prompt cannot be empty",
		},
		{
			name: "invalid example extraction",
			task: &extraction.ExtractionTask{
				Schema: schema,
				Prompt: "Extract person entities",
				Examples: []*extraction.ExampleData{
					{
						Text: "John is a developer",
						Extractions: []*extraction.Extraction{
							createTestExtraction("unknown_class", "John", map[string]interface{}{}),
						},
					},
				},
			},
			expectError: true,
			errorSubstr: "unknown extraction class",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.task.Validate()

			if tt.expectError {
				if err == nil {
					t.Errorf("Validate() error = nil, want error containing %q", tt.errorSubstr)
				} else if !strings.Contains(err.Error(), tt.errorSubstr) {
					t.Errorf("Validate() error = %q, want error containing %q", err.Error(), tt.errorSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() error = %v, want nil", err)
				}
			}
		})
	}
}

// TestSchemaJSONSerialization tests schema JSON serialization
func TestSchemaJSONSerialization(t *testing.T) {
	// Create a complex schema
	schema := extraction.NewBasicExtractionSchema("complex_schema", "Complex test schema")
	
	// Add global field
	schema.AddGlobalField(&extraction.FieldDefinition{
		Name:        "confidence",
		Type:        "number",
		Description: "Extraction confidence",
		Required:    false,
		Minimum:     ptrFloat64(0.0),
		Maximum:     ptrFloat64(1.0),
	})

	// Add class with fields
	schema.AddClass(&extraction.ClassDefinition{
		Name:        "person",
		Description: "Person entity",
		Required:    true,
		Fields: []*extraction.FieldDefinition{
			{
				Name:        "age",
				Type:        "number",
				Description: "Person age",
				Required:    false,
				Minimum:     ptrFloat64(0.0),
			},
		},
	})

	// Test serialization
	jsonData, err := schema.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}

	// Test deserialization
	deserializedSchema, err := extraction.SchemaFromJSON(jsonData)
	if err != nil {
		t.Fatalf("SchemaFromJSON() error = %v", err)
	}

	// Verify deserialized schema
	if deserializedSchema.GetName() != schema.GetName() {
		t.Errorf("Deserialized schema name = %q, want %q", 
			deserializedSchema.GetName(), schema.GetName())
	}

	if deserializedSchema.GetDescription() != schema.GetDescription() {
		t.Errorf("Deserialized schema description = %q, want %q",
			deserializedSchema.GetDescription(), schema.GetDescription())
	}

	deserializedClasses := deserializedSchema.GetClasses()
	originalClasses := schema.GetClasses()
	
	if len(deserializedClasses) != len(originalClasses) {
		t.Errorf("Deserialized classes length = %d, want %d",
			len(deserializedClasses), len(originalClasses))
	}
}

// TestFieldDefinitionConstraints tests various field constraint validations
func TestFieldDefinitionConstraints(t *testing.T) {
	schema := extraction.NewBasicExtractionSchema("constraint_test", "Field constraint testing")
	
	// Add class with various field constraints
	schema.AddClass(&extraction.ClassDefinition{
		Name: "test_entity",
		Fields: []*extraction.FieldDefinition{
			{
				Name:      "string_length",
				Type:      "string",
				MinLength: ptrInt(3),
				MaxLength: ptrInt(10),
			},
			{
				Name: "enum_field",
				Type: "string",
				Enum: []string{"option1", "option2", "option3"},
			},
			{
				Name:    "number_range",
				Type:    "number",
				Minimum: ptrFloat64(10.0),
				Maximum: ptrFloat64(100.0),
			},
			{
				Name: "boolean_field",
				Type: "boolean",
			},
			{
				Name: "array_field",
				Type: "array",
			},
		},
	})

	tests := []struct {
		name        string
		attributes  map[string]interface{}
		expectError bool
		errorSubstr string
	}{
		{
			name: "valid all constraints",
			attributes: map[string]interface{}{
				"string_length": "valid",
				"enum_field":    "option1",
				"number_range":  50.0,
				"boolean_field": true,
				"array_field":   []interface{}{"item1", "item2"},
			},
			expectError: false,
		},
		{
			name: "string too short",
			attributes: map[string]interface{}{
				"string_length": "ab",
			},
			expectError: true,
			errorSubstr: "too short",
		},
		{
			name: "string too long",
			attributes: map[string]interface{}{
				"string_length": "this_is_too_long",
			},
			expectError: true,
			errorSubstr: "too long",
		},
		{
			name: "invalid enum value",
			attributes: map[string]interface{}{
				"enum_field": "invalid_option",
			},
			expectError: true,
			errorSubstr: "not in allowed values",
		},
		{
			name: "number too small",
			attributes: map[string]interface{}{
				"number_range": 5.0,
			},
			expectError: true,
			errorSubstr: "too small",
		},
		{
			name: "number too large",
			attributes: map[string]interface{}{
				"number_range": 150.0,
			},
			expectError: true,
			errorSubstr: "too large",
		},
		{
			name: "wrong boolean type",
			attributes: map[string]interface{}{
				"boolean_field": "not_a_boolean",
			},
			expectError: true,
			errorSubstr: "expected boolean",
		},
		{
			name: "wrong array type",
			attributes: map[string]interface{}{
				"array_field": "not_an_array",
			},
			expectError: true,
			errorSubstr: "expected array",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extraction := createTestExtraction("test_entity", "test text", tt.attributes)
			err := schema.ValidateExtraction(extraction)

			if tt.expectError {
				if err == nil {
					t.Errorf("ValidateExtraction() error = nil, want error containing %q", tt.errorSubstr)
				} else if !strings.Contains(err.Error(), tt.errorSubstr) {
					t.Errorf("ValidateExtraction() error = %q, want error containing %q", err.Error(), tt.errorSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateExtraction() error = %v, want nil", err)
				}
			}
		})
	}
}

// Helper functions
func createTestExtraction(class, text string, attributes map[string]interface{}) *extraction.Extraction {
	ext := extraction.NewExtraction(class, text)
	for key, value := range attributes {
		ext.AddAttribute(key, value)
	}
	return ext
}

func ptrFloat64(v float64) *float64 {
	return &v
}

func ptrInt(v int) *int {
	return &v
}