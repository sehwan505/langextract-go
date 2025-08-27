package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/sehwan505/langextract-go/pkg/extraction"
)

// loadSchema loads extraction schema from file
func loadSchema(schemaPath string) (*extraction.BasicExtractionSchema, error) {
	data, err := os.ReadFile(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema file %s: %w", schemaPath, err)
	}

	// Parse JSON schema
	var rawSchema map[string]interface{}
	if err := json.Unmarshal(data, &rawSchema); err != nil {
		return nil, fmt.Errorf("failed to parse schema JSON: %w", err)
	}

	// Convert to BasicExtractionSchema
	schema := extraction.NewBasicExtractionSchema(
		rawSchema["name"].(string),
		rawSchema["description"].(string),
	)

	// Add classes
	if classes, ok := rawSchema["classes"].([]interface{}); ok {
		for _, classRaw := range classes {
			if classMap, ok := classRaw.(map[string]interface{}); ok {
				classDef := &extraction.ClassDefinition{
					Name:        classMap["name"].(string),
					Description: getStringValue(classMap, "description"),
					Required:    getBoolValue(classMap, "required"),
				}

				// Add fields if present
				if fields, ok := classMap["fields"].([]interface{}); ok {
					for _, fieldRaw := range fields {
						if fieldMap, ok := fieldRaw.(map[string]interface{}); ok {
							fieldDef := &extraction.FieldDefinition{
								Name:        fieldMap["name"].(string),
								Type:        fieldMap["type"].(string),
								Description: getStringValue(fieldMap, "description"),
								Required:    getBoolValue(fieldMap, "required"),
							}
							
							// Add enum values if present
							if enumValues, ok := fieldMap["enum"].([]interface{}); ok {
								fieldDef.Enum = make([]string, len(enumValues))
								for i, val := range enumValues {
									fieldDef.Enum[i] = val.(string)
								}
							}
							
							classDef.Fields = append(classDef.Fields, fieldDef)
						}
					}
				}
				
				schema.AddClass(classDef)
			}
		}
	}

	// Add global fields
	if globalFields, ok := rawSchema["globalFields"].([]interface{}); ok {
		for _, fieldRaw := range globalFields {
			if fieldMap, ok := fieldRaw.(map[string]interface{}); ok {
				fieldDef := &extraction.FieldDefinition{
					Name:        fieldMap["name"].(string),
					Type:        fieldMap["type"].(string),
					Description: getStringValue(fieldMap, "description"),
					Required:    getBoolValue(fieldMap, "required"),
				}
				schema.AddGlobalField(fieldDef)
			}
		}
	}

	return schema, nil
}

// Helper functions for safe type conversion
func getStringValue(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getBoolValue(m map[string]interface{}, key string) bool {
	if val, ok := m[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}