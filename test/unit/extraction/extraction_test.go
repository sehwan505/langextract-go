package extraction_test

import (
	"testing"

	"github.com/sehwan505/langextract-go/pkg/extraction"
	"github.com/sehwan505/langextract-go/pkg/types"
)

func TestNewExtraction(t *testing.T) {
	class := "person"
	text := "John Doe"

	ext := extraction.NewExtraction(class, text)

	if ext.ExtractionClass != class {
		t.Errorf("ExtractionClass = %q, want %q", ext.ExtractionClass, class)
	}

	if ext.ExtractionText != text {
		t.Errorf("ExtractionText = %q, want %q", ext.ExtractionText, text)
	}

	if ext.Attributes == nil {
		t.Error("Attributes should be initialized")
	}

	if len(ext.Attributes) != 0 {
		t.Errorf("Attributes should be empty initially, got %d items", len(ext.Attributes))
	}
}

func TestNewExtractionWithInterval(t *testing.T) {
	class := "person"
	text := "John Doe"
	interval := &types.CharInterval{StartPos: 0, EndPos: 8}

	ext := extraction.NewExtractionWithInterval(class, text, interval)

	if ext.ExtractionClass != class {
		t.Errorf("ExtractionClass = %q, want %q", ext.ExtractionClass, class)
	}

	if ext.ExtractionText != text {
		t.Errorf("ExtractionText = %q, want %q", ext.ExtractionText, text)
	}

	if ext.CharInterval != interval {
		t.Errorf("CharInterval = %v, want %v", ext.CharInterval, interval)
	}
}

func TestExtraction_SetCharInterval(t *testing.T) {
	ext := extraction.NewExtraction("test", "text")
	interval := &types.CharInterval{StartPos: 5, EndPos: 10}

	ext.SetCharInterval(interval)

	if ext.CharInterval != interval {
		t.Errorf("CharInterval = %v, want %v", ext.CharInterval, interval)
	}
}

func TestExtraction_SetTokenInterval(t *testing.T) {
	ext := extraction.NewExtraction("test", "text")
	interval := &types.TokenInterval{StartToken: 2, EndToken: 5}

	ext.SetTokenInterval(interval)

	if ext.TokenInterval != interval {
		t.Errorf("TokenInterval = %v, want %v", ext.TokenInterval, interval)
	}
}

func TestExtraction_SetAlignmentStatus(t *testing.T) {
	ext := extraction.NewExtraction("test", "text")
	status := types.AlignmentExact

	ext.SetAlignmentStatus(status)

	if ext.AlignmentStatus == nil {
		t.Error("AlignmentStatus should not be nil")
	}

	if *ext.AlignmentStatus != status {
		t.Errorf("AlignmentStatus = %v, want %v", *ext.AlignmentStatus, status)
	}
}

func TestExtraction_SetExtractionIndex(t *testing.T) {
	ext := extraction.NewExtraction("test", "text")
	index := 5

	ext.SetExtractionIndex(index)

	if ext.ExtractionIndex == nil {
		t.Error("ExtractionIndex should not be nil")
	}

	if *ext.ExtractionIndex != index {
		t.Errorf("ExtractionIndex = %d, want %d", *ext.ExtractionIndex, index)
	}
}

func TestExtraction_SetGroupIndex(t *testing.T) {
	ext := extraction.NewExtraction("test", "text")
	index := 3

	ext.SetGroupIndex(index)

	if ext.GroupIndex == nil {
		t.Error("GroupIndex should not be nil")
	}

	if *ext.GroupIndex != index {
		t.Errorf("GroupIndex = %d, want %d", *ext.GroupIndex, index)
	}
}

func TestExtraction_SetDescription(t *testing.T) {
	ext := extraction.NewExtraction("test", "text")
	desc := "Test description"

	ext.SetDescription(desc)

	if ext.Description == nil {
		t.Error("Description should not be nil")
	}

	if *ext.Description != desc {
		t.Errorf("Description = %q, want %q", *ext.Description, desc)
	}
}

func TestExtraction_AddAttribute(t *testing.T) {
	ext := extraction.NewExtraction("test", "text")

	ext.AddAttribute("confidence", 0.95)
	ext.AddAttribute("source", "test_source")

	if len(ext.Attributes) != 2 {
		t.Errorf("Attributes length = %d, want 2", len(ext.Attributes))
	}

	conf, exists := ext.Attributes["confidence"]
	if !exists {
		t.Error("confidence attribute should exist")
	}
	if conf != 0.95 {
		t.Errorf("confidence = %v, want 0.95", conf)
	}

	source, exists := ext.Attributes["source"]
	if !exists {
		t.Error("source attribute should exist")
	}
	if source != "test_source" {
		t.Errorf("source = %v, want 'test_source'", source)
	}
}

func TestExtraction_GetAttribute(t *testing.T) {
	ext := extraction.NewExtraction("test", "text")
	ext.AddAttribute("test_key", "test_value")

	// Test existing attribute
	value, exists := ext.GetAttribute("test_key")
	if !exists {
		t.Error("GetAttribute() should return true for existing attribute")
	}
	if value != "test_value" {
		t.Errorf("GetAttribute() = %v, want 'test_value'", value)
	}

	// Test non-existing attribute
	_, exists = ext.GetAttribute("non_existing")
	if exists {
		t.Error("GetAttribute() should return false for non-existing attribute")
	}
}

func TestExtraction_GetStringAttribute(t *testing.T) {
	ext := extraction.NewExtraction("test", "text")
	ext.AddAttribute("string_attr", "string_value")
	ext.AddAttribute("int_attr", 123)

	// Test string attribute
	str, ok := ext.GetStringAttribute("string_attr")
	if !ok {
		t.Error("GetStringAttribute() should return true for string attribute")
	}
	if str != "string_value" {
		t.Errorf("GetStringAttribute() = %q, want 'string_value'", str)
	}

	// Test non-string attribute
	_, ok = ext.GetStringAttribute("int_attr")
	if ok {
		t.Error("GetStringAttribute() should return false for non-string attribute")
	}

	// Test non-existing attribute
	_, ok = ext.GetStringAttribute("non_existing")
	if ok {
		t.Error("GetStringAttribute() should return false for non-existing attribute")
	}
}

func TestExtraction_GetFloatAttribute(t *testing.T) {
	ext := extraction.NewExtraction("test", "text")
	ext.AddAttribute("float64_attr", 123.45)
	ext.AddAttribute("float32_attr", float32(67.89))
	ext.AddAttribute("int_attr", 100)
	ext.AddAttribute("int64_attr", int64(200))
	ext.AddAttribute("string_attr", "not_a_number")

	tests := []struct {
		name     string
		key      string
		expected float64
		wantOk   bool
	}{
		{"float64", "float64_attr", 123.45, true},
		{"float32", "float32_attr", 67.889999, true}, // float32 precision
		{"int", "int_attr", 100.0, true},
		{"int64", "int64_attr", 200.0, true},
		{"string", "string_attr", 0, false},
		{"non-existing", "non_existing", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ext.GetFloatAttribute(tt.key)
			if ok != tt.wantOk {
				t.Errorf("GetFloatAttribute() ok = %v, want %v", ok, tt.wantOk)
			}
			if tt.wantOk {
				// Use approximate comparison for floating point
				if abs(got-tt.expected) > 0.0001 {
					t.Errorf("GetFloatAttribute() = %f, want %f", got, tt.expected)
				}
			}
		})
	}
}

func TestExtraction_GetIntAttribute(t *testing.T) {
	ext := extraction.NewExtraction("test", "text")
	ext.AddAttribute("int_attr", 123)
	ext.AddAttribute("int64_attr", int64(456))
	ext.AddAttribute("int32_attr", int32(789))
	ext.AddAttribute("float64_attr", 100.0)
	ext.AddAttribute("string_attr", "not_a_number")

	tests := []struct {
		name     string
		key      string
		expected int
		wantOk   bool
	}{
		{"int", "int_attr", 123, true},
		{"int64", "int64_attr", 456, true},
		{"int32", "int32_attr", 789, true},
		{"float64", "float64_attr", 100, true},
		{"string", "string_attr", 0, false},
		{"non-existing", "non_existing", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ext.GetIntAttribute(tt.key)
			if ok != tt.wantOk {
				t.Errorf("GetIntAttribute() ok = %v, want %v", ok, tt.wantOk)
			}
			if tt.wantOk && got != tt.expected {
				t.Errorf("GetIntAttribute() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestExtraction_HasCharInterval(t *testing.T) {
	ext := extraction.NewExtraction("test", "text")

	if ext.HasCharInterval() {
		t.Error("HasCharInterval() should return false when CharInterval is nil")
	}

	ext.SetCharInterval(&types.CharInterval{StartPos: 0, EndPos: 5})

	if !ext.HasCharInterval() {
		t.Error("HasCharInterval() should return true when CharInterval is set")
	}
}

func TestExtraction_HasTokenInterval(t *testing.T) {
	ext := extraction.NewExtraction("test", "text")

	if ext.HasTokenInterval() {
		t.Error("HasTokenInterval() should return false when TokenInterval is nil")
	}

	ext.SetTokenInterval(&types.TokenInterval{StartToken: 0, EndToken: 2})

	if !ext.HasTokenInterval() {
		t.Error("HasTokenInterval() should return true when TokenInterval is set")
	}
}

func TestExtraction_IsWellGrounded(t *testing.T) {
	ext := extraction.NewExtraction("test", "text")

	if ext.IsWellGrounded() {
		t.Error("IsWellGrounded() should return false when AlignmentStatus is nil")
	}

	// Test low quality alignment
	ext.SetAlignmentStatus(types.AlignmentApproximate)
	if ext.IsWellGrounded() {
		t.Error("IsWellGrounded() should return false for low quality alignment")
	}

	// Test high quality alignment
	ext.SetAlignmentStatus(types.AlignmentExact)
	if !ext.IsWellGrounded() {
		t.Error("IsWellGrounded() should return true for high quality alignment")
	}
}

func TestExtraction_GetSetConfidence(t *testing.T) {
	ext := extraction.NewExtraction("test", "text")

	// Test when confidence is not set
	_, ok := ext.GetConfidence()
	if ok {
		t.Error("GetConfidence() should return false when confidence is not set")
	}

	// Set confidence
	ext.SetConfidence(0.85)

	// Test getting confidence
	conf, ok := ext.GetConfidence()
	if !ok {
		t.Error("GetConfidence() should return true when confidence is set")
	}
	if conf != 0.85 {
		t.Errorf("GetConfidence() = %f, want 0.85", conf)
	}
}

func TestExtraction_Length(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{"empty text", "", 0},
		{"single character", "a", 1},
		{"normal text", "hello", 5},
		{"unicode text", "héllo", 6}, // UTF-8 byte length
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ext := extraction.NewExtraction("test", tt.text)
			if got := ext.Length(); got != tt.expected {
				t.Errorf("Length() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestExtraction_IsEmpty(t *testing.T) {
	emptyExt := extraction.NewExtraction("test", "")
	if !emptyExt.IsEmpty() {
		t.Error("IsEmpty() should return true for empty text")
	}

	nonEmptyExt := extraction.NewExtraction("test", "text")
	if nonEmptyExt.IsEmpty() {
		t.Error("IsEmpty() should return false for non-empty text")
	}
}

func TestExtraction_Copy(t *testing.T) {
	// Create original extraction with all fields set
	original := extraction.NewExtraction("person", "John Doe")
	original.SetCharInterval(&types.CharInterval{StartPos: 0, EndPos: 8})
	original.SetTokenInterval(&types.TokenInterval{StartToken: 0, EndToken: 2})
	original.SetAlignmentStatus(types.AlignmentExact)
	original.SetExtractionIndex(1)
	original.SetGroupIndex(2)
	original.SetDescription("Test person")
	original.AddAttribute("confidence", 0.95)
	original.AddAttribute("source", "test")

	// Create copy
	copy := original.Copy()

	// Verify all fields are copied
	if copy.ExtractionClass != original.ExtractionClass {
		t.Errorf("Copy ExtractionClass = %q, want %q", copy.ExtractionClass, original.ExtractionClass)
	}

	if copy.ExtractionText != original.ExtractionText {
		t.Errorf("Copy ExtractionText = %q, want %q", copy.ExtractionText, original.ExtractionText)
	}

	// Verify intervals are deep copied
	if copy.CharInterval == original.CharInterval {
		t.Error("CharInterval should be deep copied, not reference copied")
	}
	if *copy.CharInterval != *original.CharInterval {
		t.Error("CharInterval values should be equal")
	}

	if copy.TokenInterval == original.TokenInterval {
		t.Error("TokenInterval should be deep copied, not reference copied")
	}
	if *copy.TokenInterval != *original.TokenInterval {
		t.Error("TokenInterval values should be equal")
	}

	// Verify other pointer fields
	if *copy.AlignmentStatus != *original.AlignmentStatus {
		t.Error("AlignmentStatus should be copied")
	}

	if *copy.ExtractionIndex != *original.ExtractionIndex {
		t.Error("ExtractionIndex should be copied")
	}

	if *copy.GroupIndex != *original.GroupIndex {
		t.Error("GroupIndex should be copied")
	}

	if *copy.Description != *original.Description {
		t.Error("Description should be copied")
	}

	// Verify attributes are deep copied (compare addresses, not maps directly)
	if &copy.Attributes == &original.Attributes {
		t.Error("Attributes should be deep copied, not reference copied")
	}

	if len(copy.Attributes) != len(original.Attributes) {
		t.Error("Attributes length should be equal")
	}

	for key, value := range original.Attributes {
		if copy.Attributes[key] != value {
			t.Errorf("Attribute %s should be copied", key)
		}
	}
}

func TestExtraction_String(t *testing.T) {
	// Test extraction without interval
	ext1 := extraction.NewExtraction("person", "John Doe")
	str1 := ext1.String()

	if !contains(str1, "person") || !contains(str1, "John Doe") || !contains(str1, "no-position") {
		t.Errorf("String() = %q, should contain class, text, and no-position", str1)
	}

	// Test extraction with interval
	ext2 := extraction.NewExtractionWithInterval("person", "John Doe", &types.CharInterval{StartPos: 0, EndPos: 8})
	str2 := ext2.String()

	if !contains(str2, "person") || !contains(str2, "John Doe") || !contains(str2, "[0:8)") {
		t.Errorf("String() = %q, should contain class, text, and position", str2)
	}

	// Test long text truncation
	longText := "This is a very long extraction text that should be truncated in the string representation"
	ext3 := extraction.NewExtraction("long", longText)
	str3 := ext3.String()

	if !contains(str3, "...") {
		t.Errorf("String() should truncate long text with '...', got %q", str3)
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || len(s) > len(substr) &&
			(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
				containsAt(s, substr)))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// abs returns the absolute value of a float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}