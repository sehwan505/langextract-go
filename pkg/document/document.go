package document

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

// Document represents a text document with optional metadata and tokenization.
type Document struct {
	Text              string            `json:"text"`                         // Raw text content
	AdditionalContext string            `json:"additional_context,omitempty"` // Optional context metadata
	documentID        string            // Internal document identifier
	tokenizedText     []string          // Cached tokenized text
}

// NewDocument creates a new Document with the given text.
func NewDocument(text string) *Document {
	return &Document{
		Text: text,
	}
}

// NewDocumentWithContext creates a new Document with text and additional context.
func NewDocumentWithContext(text, context string) *Document {
	return &Document{
		Text:              text,
		AdditionalContext: context,
	}
}

// DocumentID returns the document identifier, generating one if not set.
func (d *Document) DocumentID() string {
	if d.documentID == "" {
		d.documentID = d.generateID()
	}
	return d.documentID
}

// SetDocumentID sets the document identifier.
func (d *Document) SetDocumentID(id string) {
	d.documentID = id
}

// TokenizedText returns the tokenized version of the text.
// If not already tokenized, performs basic whitespace tokenization.
func (d *Document) TokenizedText() []string {
	if d.tokenizedText == nil {
		d.tokenizedText = d.tokenize()
	}
	return d.tokenizedText
}

// SetTokenizedText sets the tokenized text directly.
func (d *Document) SetTokenizedText(tokens []string) {
	d.tokenizedText = tokens
}

// generateID creates a unique identifier based on the document content.
func (d *Document) generateID() string {
	hash := sha256.Sum256([]byte(d.Text + d.AdditionalContext))
	return fmt.Sprintf("doc_%x", hash[:8])
}

// tokenize performs basic whitespace tokenization.
// This can be replaced with more sophisticated tokenization later.
func (d *Document) tokenize() []string {
	if d.Text == "" {
		return []string{}
	}
	
	// Basic whitespace tokenization
	fields := strings.Fields(d.Text)
	if len(fields) == 0 {
		return []string{}
	}
	
	return fields
}

// Length returns the character length of the document text.
func (d *Document) Length() int {
	return len(d.Text)
}

// TokenCount returns the number of tokens in the document.
func (d *Document) TokenCount() int {
	return len(d.TokenizedText())
}

// IsEmpty returns true if the document has no text content.
func (d *Document) IsEmpty() bool {
	return strings.TrimSpace(d.Text) == ""
}

// String returns a string representation of the document.
func (d *Document) String() string {
	preview := d.Text
	if len(preview) > 100 {
		preview = preview[:97] + "..."
	}
	return fmt.Sprintf("Document{id=%s, length=%d, preview=%q}", 
		d.DocumentID(), d.Length(), preview)
}