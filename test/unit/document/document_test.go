package document_test

import (
	"strings"
	"testing"

	"github.com/sehwan505/langextract-go/pkg/document"
)

func TestNewDocument(t *testing.T) {
	text := "This is a test document."
	doc := document.NewDocument(text)
	
	if doc.Text != text {
		t.Errorf("Text = %q, want %q", doc.Text, text)
	}
	
	if doc.AdditionalContext != "" {
		t.Errorf("AdditionalContext = %q, want empty string", doc.AdditionalContext)
	}
}

func TestNewDocumentWithContext(t *testing.T) {
	text := "This is a test document."
	context := "Test context information."
	doc := document.NewDocumentWithContext(text, context)
	
	if doc.Text != text {
		t.Errorf("Text = %q, want %q", doc.Text, text)
	}
	
	if doc.AdditionalContext != context {
		t.Errorf("AdditionalContext = %q, want %q", doc.AdditionalContext, context)
	}
}

func TestDocument_DocumentID(t *testing.T) {
	doc := document.NewDocument("test text")
	
	// First call should generate an ID
	id1 := doc.DocumentID()
	if id1 == "" {
		t.Error("DocumentID() returned empty string")
	}
	
	// Second call should return the same ID
	id2 := doc.DocumentID()
	if id1 != id2 {
		t.Errorf("DocumentID() inconsistent: first=%q, second=%q", id1, id2)
	}
	
	// ID should start with "doc_"
	if !strings.HasPrefix(id1, "doc_") {
		t.Errorf("DocumentID() = %q, expected to start with 'doc_'", id1)
	}
}

func TestDocument_SetDocumentID(t *testing.T) {
	doc := document.NewDocument("test text")
	customID := "custom_id_123"
	
	doc.SetDocumentID(customID)
	
	if doc.DocumentID() != customID {
		t.Errorf("DocumentID() = %q, want %q", doc.DocumentID(), customID)
	}
}

func TestDocument_TokenizedText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected []string
	}{
		{"simple text", "hello world", []string{"hello", "world"}},
		{"multiple spaces", "hello   world  test", []string{"hello", "world", "test"}},
		{"empty text", "", []string{}},
		{"single word", "hello", []string{"hello"}},
		{"punctuation", "hello, world!", []string{"hello,", "world!"}},
		{"tabs and newlines", "hello\tworld\ntest", []string{"hello", "world", "test"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := document.NewDocument(tt.text)
			got := doc.TokenizedText()
			
			if len(got) != len(tt.expected) {
				t.Errorf("TokenizedText() length = %d, want %d", len(got), len(tt.expected))
				return
			}
			
			for i, token := range got {
				if token != tt.expected[i] {
					t.Errorf("TokenizedText()[%d] = %q, want %q", i, token, tt.expected[i])
				}
			}
		})
	}
}

func TestDocument_SetTokenizedText(t *testing.T) {
	doc := document.NewDocument("original text")
	customTokens := []string{"custom", "tokens", "here"}
	
	doc.SetTokenizedText(customTokens)
	got := doc.TokenizedText()
	
	if len(got) != len(customTokens) {
		t.Errorf("TokenizedText() length = %d, want %d", len(got), len(customTokens))
		return
	}
	
	for i, token := range got {
		if token != customTokens[i] {
			t.Errorf("TokenizedText()[%d] = %q, want %q", i, token, customTokens[i])
		}
	}
}

func TestDocument_Length(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{"empty text", "", 0},
		{"single character", "a", 1},
		{"simple text", "hello", 5},
		{"text with spaces", "hello world", 11},
		{"unicode text", "héllo wörld", 13}, // UTF-8 byte length
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := document.NewDocument(tt.text)
			if got := doc.Length(); got != tt.expected {
				t.Errorf("Length() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestDocument_TokenCount(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{"empty text", "", 0},
		{"single word", "hello", 1},
		{"two words", "hello world", 2},
		{"multiple spaces", "hello   world  test", 3},
		{"tabs and newlines", "hello\tworld\ntest", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := document.NewDocument(tt.text)
			if got := doc.TokenCount(); got != tt.expected {
				t.Errorf("TokenCount() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestDocument_IsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected bool
	}{
		{"empty text", "", true},
		{"whitespace only", "   \t\n  ", true},
		{"single character", "a", false},
		{"normal text", "hello world", false},
		{"text with leading/trailing spaces", "  hello  ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := document.NewDocument(tt.text)
			if got := doc.IsEmpty(); got != tt.expected {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDocument_String(t *testing.T) {
	// Test normal length text
	doc := document.NewDocument("This is a short document.")
	str := doc.String()
	
	if !strings.Contains(str, "Document{") {
		t.Errorf("String() should contain 'Document{', got %q", str)
	}
	
	if !strings.Contains(str, "This is a short document.") {
		t.Errorf("String() should contain full text, got %q", str)
	}
	
	// Test long text (should be truncated)
	longText := strings.Repeat("a", 200)
	longDoc := document.NewDocument(longText)
	longStr := longDoc.String()
	
	if !strings.Contains(longStr, "...") {
		t.Errorf("String() should truncate long text with '...', got %q", longStr)
	}
	
	// Test that preview is not longer than 100 characters
	startIdx := strings.Index(longStr, "preview=\"") + len("preview=\"")
	endIdx := strings.Index(longStr[startIdx:], "\"")
	preview := longStr[startIdx : startIdx+endIdx]
	
	if len(preview) > 100 {
		t.Errorf("String() preview should be max 100 chars, got %d chars", len(preview))
	}
}

func TestDocument_TokenizedTextCaching(t *testing.T) {
	doc := document.NewDocument("hello world test")
	
	// First call should tokenize
	tokens1 := doc.TokenizedText()
	
	// Second call should return cached result
	tokens2 := doc.TokenizedText()
	
	// Should be the same slice (same memory address)
	if &tokens1[0] != &tokens2[0] {
		t.Error("TokenizedText() should cache results")
	}
	
	// Setting new tokens should update cache
	newTokens := []string{"new", "tokens"}
	doc.SetTokenizedText(newTokens)
	tokens3 := doc.TokenizedText()
	
	if len(tokens3) != 2 || tokens3[0] != "new" || tokens3[1] != "tokens" {
		t.Errorf("SetTokenizedText() should update cached tokens")
	}
}

func TestDocument_IDGeneration(t *testing.T) {
	// Same text should generate same ID
	doc1 := document.NewDocument("identical text")
	doc2 := document.NewDocument("identical text")
	
	if doc1.DocumentID() != doc2.DocumentID() {
		t.Error("Documents with identical text should have same ID")
	}
	
	// Different text should generate different IDs
	doc3 := document.NewDocument("different text")
	
	if doc1.DocumentID() == doc3.DocumentID() {
		t.Error("Documents with different text should have different IDs")
	}
	
	// Same text with different context should generate different IDs
	doc4 := document.NewDocumentWithContext("identical text", "context")
	
	if doc1.DocumentID() == doc4.DocumentID() {
		t.Error("Documents with same text but different context should have different IDs")
	}
}