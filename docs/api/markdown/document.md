# document

Package: `github.com/sehwan505/langextract-go/pkg/document`

```go
package document // import "github.com/sehwan505/langextract-go/pkg/document"


TYPES

type AnnotatedDocument struct {
	*Document                            // Embedded document
	Extractions []*extraction.Extraction `json:"extractions,omitempty"` // List of extracted entities
}
    AnnotatedDocument represents a document with extracted entities and
    annotations.

func NewAnnotatedDocument(doc *Document) *AnnotatedDocument
    NewAnnotatedDocument creates a new AnnotatedDocument from a Document.

func NewAnnotatedDocumentWithText(text string) *AnnotatedDocument
    NewAnnotatedDocumentWithText creates a new AnnotatedDocument with the given
    text.

func (ad *AnnotatedDocument) AddExtraction(ext *extraction.Extraction)
    AddExtraction adds an extraction to the document.

func (ad *AnnotatedDocument) AddExtractions(extractions []*extraction.Extraction)
    AddExtractions adds multiple extractions to the document.

func (ad *AnnotatedDocument) ExtractionCount() int
    ExtractionCount returns the total number of extractions.

func (ad *AnnotatedDocument) FilterExtractionsByConfidence(threshold float64) []*extraction.Extraction
    FilterExtractionsByConfidence returns extractions with confidence above
    the threshold. This assumes extractions have confidence scores in their
    attributes.

func (ad *AnnotatedDocument) GetCoverage() float64
    GetCoverage returns the percentage of document text covered by extractions.

func (ad *AnnotatedDocument) GetExtractionsByClass(class string) []*extraction.Extraction
    GetExtractionsByClass returns all extractions of a specific class.

func (ad *AnnotatedDocument) GetExtractionsByGroup(groupIndex int) []*extraction.Extraction
    GetExtractionsByGroup returns all extractions with a specific group index.

func (ad *AnnotatedDocument) GetUniqueClasses() []string
    GetUniqueClasses returns all unique extraction classes in the document.

func (ad *AnnotatedDocument) HasExtractions() bool
    HasExtractions returns true if the document has any extractions.

func (ad *AnnotatedDocument) SortExtractionsByIndex()
    SortExtractionsByIndex sorts extractions by their extraction index.

func (ad *AnnotatedDocument) SortExtractionsByPosition()
    SortExtractionsByPosition sorts extractions by their character position in
    the text.

func (ad *AnnotatedDocument) String() string
    String returns a string representation of the annotated document.

type Document struct {
	Text              string `json:"text"`                         // Raw text content
	AdditionalContext string `json:"additional_context,omitempty"` // Optional context metadata

	// Has unexported fields.
}
    Document represents a text document with optional metadata and tokenization.

func NewDocument(text string) *Document
    NewDocument creates a new Document with the given text.

func NewDocumentWithContext(text, context string) *Document
    NewDocumentWithContext creates a new Document with text and additional
    context.

func (d *Document) DocumentID() string
    DocumentID returns the document identifier, generating one if not set.

func (d *Document) IsEmpty() bool
    IsEmpty returns true if the document has no text content.

func (d *Document) Length() int
    Length returns the character length of the document text.

func (d *Document) SetDocumentID(id string)
    SetDocumentID sets the document identifier.

func (d *Document) SetTokenizedText(tokens []string)
    SetTokenizedText sets the tokenized text directly.

func (d *Document) String() string
    String returns a string representation of the document.

func (d *Document) TokenCount() int
    TokenCount returns the number of tokens in the document.

func (d *Document) TokenizedText() []string
    TokenizedText returns the tokenized version of the text. If not already
    tokenized, performs basic whitespace tokenization.

```
