package main

import (
	"fmt"
	"log"

	"github.com/sehwan505/langextract-go/pkg/document"
	"github.com/sehwan505/langextract-go/pkg/extraction"
	"github.com/sehwan505/langextract-go/pkg/types"
)

func main() {
	// Example 1: Basic Document and Extraction
	fmt.Println("=== Basic Usage Example ===")
	basicExample()

	fmt.Println("\n=== Schema Definition Example ===")
	schemaExample()

	fmt.Println("\n=== Complex Extraction Example ===")
	complexExample()
}

func basicExample() {
	// Create a document
	text := "John Doe is a software engineer at Google Inc. He lives in Mountain View, California."
	doc := document.NewDocument(text)

	fmt.Printf("Document ID: %s\n", doc.DocumentID())
	fmt.Printf("Text length: %d characters\n", doc.Length())
	fmt.Printf("Token count: %d\n", doc.TokenCount())

	// Create extractions
	personExt := extraction.NewExtraction("person", "John Doe")
	if interval, err := types.NewCharInterval(0, 8); err == nil {
		personExt.SetCharInterval(interval)
	}
	personExt.SetAlignmentStatus(types.AlignmentExact)
	personExt.SetConfidence(0.95)

	orgExt := extraction.NewExtraction("organization", "Google Inc.")
	if interval, err := types.NewCharInterval(41, 52); err == nil {
		orgExt.SetCharInterval(interval)
	}
	orgExt.SetAlignmentStatus(types.AlignmentExact)
	orgExt.SetConfidence(0.90)

	locationExt := extraction.NewExtraction("location", "Mountain View, California")
	if interval, err := types.NewCharInterval(65, 89); err == nil {
		locationExt.SetCharInterval(interval)
	}
	locationExt.SetAlignmentStatus(types.AlignmentExact)
	locationExt.SetConfidence(0.88)

	// Create annotated document
	annotated := document.NewAnnotatedDocument(doc)
	annotated.AddExtraction(personExt)
	annotated.AddExtraction(orgExt)
	annotated.AddExtraction(locationExt)

	// Display results
	fmt.Printf("Extractions: %d\n", annotated.ExtractionCount())
	fmt.Printf("Unique classes: %v\n", annotated.GetUniqueClasses())
	fmt.Printf("Text coverage: %.1f%%\n", annotated.GetCoverage())

	// Print each extraction
	for i, ext := range annotated.Extractions {
		fmt.Printf("Extraction %d: %s\n", i+1, ext.String())
	}
}

func schemaExample() {
	// Create a schema for extracting business information
	schema := extraction.NewBasicExtractionSchema("business_entities", "Extract business-related entities")

	// Define extraction classes
	personClass := &extraction.ClassDefinition{
		Name:        "person",
		Description: "Person names and titles",
		Required:    true,
		MinCount:    intPtr(1),
	}

	orgClass := &extraction.ClassDefinition{
		Name:        "organization",
		Description: "Company and organization names",
		Required:    false,
	}

	locationClass := &extraction.ClassDefinition{
		Name:        "location",
		Description: "Geographic locations",
		Required:    false,
	}

	schema.AddClass(personClass)
	schema.AddClass(orgClass)
	schema.AddClass(locationClass)

	// Add global fields
	confidenceField := &extraction.FieldDefinition{
		Name:        "confidence",
		Type:        "number",
		Description: "Confidence score between 0 and 1",
		Required:    true,
		Minimum:     floatPtr(0.0),
		Maximum:     floatPtr(1.0),
	}

	schema.AddGlobalField(confidenceField)

	fmt.Printf("Schema: %s\n", schema.GetName())
	fmt.Printf("Description: %s\n", schema.GetDescription())
	fmt.Printf("Classes: %v\n", schema.GetClasses())

	// Create sample extraction and validate
	ext := extraction.NewExtraction("person", "Jane Smith")
	ext.SetConfidence(0.92)

	if err := schema.ValidateExtraction(ext); err != nil {
		log.Printf("Validation error: %v", err)
	} else {
		fmt.Println("Extraction is valid according to schema")
	}
}

func complexExample() {
	// More complex document with multiple entities
	text := `Dr. Sarah Johnson, Chief Technology Officer at Microsoft Corporation, announced today 
that the company will be opening a new research facility in Seattle, Washington. The facility 
will focus on artificial intelligence and machine learning research. Ms. Johnson, who holds 
a PhD from Stanford University, has been with Microsoft for over 10 years.`

	doc := document.NewDocument(text)
	annotated := document.NewAnnotatedDocument(doc)

	// Extract multiple entities with grouping
	extractions := []*extraction.Extraction{
		createExtraction("person", "Dr. Sarah Johnson", 0, 17, 0, 0),
		createExtraction("title", "Chief Technology Officer", 19, 43, 0, 0),
		createExtraction("organization", "Microsoft Corporation", 47, 68, 0, 1),
		createExtraction("location", "Seattle, Washington", 140, 159, 0, 2),
		createExtraction("facility_type", "research facility", 122, 139, 0, 2),
		createExtraction("research_area", "artificial intelligence", 180, 203, 1, 3),
		createExtraction("research_area", "machine learning research", 208, 233, 1, 3),
		createExtraction("person", "Ms. Johnson", 235, 246, 2, 0),
		createExtraction("degree", "PhD", 259, 262, 2, 4),
		createExtraction("organization", "Stanford University", 268, 287, 2, 4),
		createExtraction("organization", "Microsoft", 304, 313, 2, 1),
		createExtraction("duration", "over 10 years", 318, 332, 2, 5),
	}

	annotated.AddExtractions(extractions)

	// Sort by position in text
	annotated.SortExtractionsByPosition()

	fmt.Printf("Document length: %d characters\n", doc.Length())
	fmt.Printf("Total extractions: %d\n", annotated.ExtractionCount())
	fmt.Printf("Text coverage: %.1f%%\n", annotated.GetCoverage())

	// Group extractions by class
	for _, class := range annotated.GetUniqueClasses() {
		classExtractions := annotated.GetExtractionsByClass(class)
		fmt.Printf("\n%s (%d):\n", class, len(classExtractions))
		for _, ext := range classExtractions {
			conf, _ := ext.GetConfidence()
			fmt.Printf("  - %q (confidence: %.2f)\n", ext.ExtractionText, conf)
		}
	}

	// Filter by confidence
	highConfidenceExts := annotated.FilterExtractionsByConfidence(0.90)
	fmt.Printf("\nHigh confidence extractions (>0.90): %d\n", len(highConfidenceExts))
	for _, ext := range highConfidenceExts {
		conf, _ := ext.GetConfidence()
		fmt.Printf("  - %s: %q (%.2f)\n", ext.ExtractionClass, ext.ExtractionText, conf)
	}
}

func createExtraction(class, text string, start, end, groupIndex, sentenceIndex int) *extraction.Extraction {
	ext := extraction.NewExtraction(class, text)

	if interval, err := types.NewCharInterval(start, end); err == nil {
		ext.SetCharInterval(interval)
	}

	ext.SetAlignmentStatus(types.AlignmentExact)
	ext.SetGroupIndex(groupIndex)
	ext.SetExtractionIndex(sentenceIndex)

	// Simulate confidence scores
	confidence := 0.85 + float64(len(text)%10)/100.0 // Varies between 0.85-0.94
	ext.SetConfidence(confidence)

	// Add metadata
	ext.AddAttribute("sentence_index", sentenceIndex)
	ext.AddAttribute("word_count", len(types.TokenInterval{}.String()))

	return ext
}

// Helper functions for pointers
func intPtr(i int) *int {
	return &i
}

func floatPtr(f float64) *float64 {
	return &f
}
