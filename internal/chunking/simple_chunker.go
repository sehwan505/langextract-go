package chunking

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/sehwan505/langextract-go/pkg/document"
	"github.com/sehwan505/langextract-go/pkg/types"
)

// SimpleChunker implements basic sentence and paragraph-based text chunking.
// It attempts to preserve sentence boundaries while respecting character limits.
type SimpleChunker struct {
	sentencePattern *regexp.Regexp
	paragraphPattern *regexp.Regexp
}

// NewSimpleChunker creates a new SimpleChunker with default sentence detection.
func NewSimpleChunker() *SimpleChunker {
	return &SimpleChunker{
		// Basic sentence ending patterns (can be enhanced with more sophisticated rules)
		sentencePattern: regexp.MustCompile(`[.!?]+\s+`),
		// Paragraph boundaries (double newline or more)
		paragraphPattern: regexp.MustCompile(`\n\s*\n`),
	}
}

// Name returns the name of this chunking algorithm.
func (sc *SimpleChunker) Name() string {
	return "SimpleChunker"
}

// ChunkDocument splits a document into chunks based on the specified options.
func (sc *SimpleChunker) ChunkDocument(ctx context.Context, doc *document.Document, opts ChunkingOptions) ([]TextChunk, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	if doc == nil {
		return nil, NewChunkingError(ErrorTypeValidation, "document cannot be nil")
	}

	return sc.ChunkText(ctx, doc.Text, opts)
}

// ChunkText splits raw text into chunks based on the specified options.
func (sc *SimpleChunker) ChunkText(ctx context.Context, text string, opts ChunkingOptions) ([]TextChunk, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	if text == "" {
		return []TextChunk{}, nil
	}

	// Check for context cancellation
	select {
	case <-ctx.Done():
		return nil, NewChunkingErrorWithCause(ErrorTypeTimeout, "chunking cancelled", ctx.Err())
	default:
	}

	var chunks []TextChunk
	var err error

	if opts.PreserveParagraphs {
		chunks, err = sc.chunkByParagraphs(ctx, text, opts)
	} else if opts.PreserveSentences {
		chunks, err = sc.chunkBySentences(ctx, text, opts)
	} else {
		chunks, err = sc.chunkByCharacters(ctx, text, opts)
	}

	if err != nil {
		return nil, err
	}

	// Add metadata and finalize chunks
	return sc.finalizeChunks(chunks, text, opts), nil
}

// EstimateChunks provides an estimate of the number of chunks that will be created.
func (sc *SimpleChunker) EstimateChunks(text string, opts ChunkingOptions) int {
	if text == "" || opts.MaxCharBuffer <= 0 {
		return 0
	}
	
	// Simple estimation based on text length and overlap
	effectiveChunkSize := float64(opts.MaxCharBuffer)
	if opts.OverlapRatio > 0 {
		effectiveChunkSize *= (1.0 - opts.OverlapRatio)
	}
	
	estimate := int(float64(len(text)) / effectiveChunkSize)
	if estimate == 0 && len(text) > 0 {
		return 1
	}
	return estimate
}

// chunkByParagraphs splits text into chunks while preserving paragraph boundaries.
func (sc *SimpleChunker) chunkByParagraphs(ctx context.Context, text string, opts ChunkingOptions) ([]TextChunk, error) {
	paragraphs := sc.paragraphPattern.Split(text, -1)
	var chunks []TextChunk
	currentChunk := ""
	startPos := 0
	
	for i, paragraph := range paragraphs {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return nil, NewChunkingErrorWithCause(ErrorTypeTimeout, "chunking cancelled", ctx.Err())
		default:
		}
		
		paragraph = strings.TrimSpace(paragraph)
		if paragraph == "" {
			continue
		}
		
		// If adding this paragraph would exceed the limit, finalize current chunk
		if len(currentChunk)+len(paragraph) > opts.MaxCharBuffer && currentChunk != "" {
			chunk := sc.createChunk(currentChunk, startPos, len(chunks), opts)
			chunks = append(chunks, chunk)
			
			// Start new chunk
			currentChunk = paragraph
			startPos = strings.Index(text[startPos:], paragraph) + startPos
		} else {
			if currentChunk != "" {
				currentChunk += "\n\n" + paragraph
			} else {
				currentChunk = paragraph
				startPos = strings.Index(text, paragraph)
			}
		}
		
		// Handle last paragraph
		if i == len(paragraphs)-1 && currentChunk != "" {
			chunk := sc.createChunk(currentChunk, startPos, len(chunks), opts)
			chunks = append(chunks, chunk)
		}
	}
	
	return chunks, nil
}

// chunkBySentences splits text into chunks while preserving sentence boundaries.
func (sc *SimpleChunker) chunkBySentences(ctx context.Context, text string, opts ChunkingOptions) ([]TextChunk, error) {
	sentences := sc.findSentences(text)
	var chunks []TextChunk
	currentChunk := ""
	startPos := 0
	
	for i, sentence := range sentences {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return nil, NewChunkingErrorWithCause(ErrorTypeTimeout, "chunking cancelled", ctx.Err())
		default:
		}
		
		sentence = strings.TrimSpace(sentence)
		if sentence == "" {
			continue
		}
		
		// If this sentence alone exceeds the limit, create it as its own chunk
		if len(sentence) > opts.MaxCharBuffer {
			// First, finalize any existing chunk
			if currentChunk != "" {
				chunk := sc.createChunk(currentChunk, startPos, len(chunks), opts)
				chunks = append(chunks, chunk)
				currentChunk = ""
			}
			
			// Create chunk for the oversized sentence
			sentencePos := strings.Index(text[startPos:], sentence) + startPos
			chunk := sc.createChunk(sentence, sentencePos, len(chunks), opts)
			chunks = append(chunks, chunk)
			startPos = sentencePos + len(sentence)
			continue
		}
		
		// If adding this sentence would exceed the limit, finalize current chunk
		if len(currentChunk)+len(sentence) > opts.MaxCharBuffer && currentChunk != "" {
			chunk := sc.createChunk(currentChunk, startPos, len(chunks), opts)
			chunks = append(chunks, chunk)
			
			// Start new chunk with current sentence
			currentChunk = sentence
			startPos = strings.Index(text[startPos:], sentence) + startPos
		} else {
			if currentChunk != "" {
				currentChunk += " " + sentence
			} else {
				currentChunk = sentence
				startPos = strings.Index(text[startPos:], sentence) + startPos
			}
		}
		
		// Handle last sentence
		if i == len(sentences)-1 && currentChunk != "" {
			chunk := sc.createChunk(currentChunk, startPos, len(chunks), opts)
			chunks = append(chunks, chunk)
		}
	}
	
	return chunks, nil
}

// chunkByCharacters splits text into fixed-size character chunks without regard to boundaries.
func (sc *SimpleChunker) chunkByCharacters(ctx context.Context, text string, opts ChunkingOptions) ([]TextChunk, error) {
	var chunks []TextChunk
	overlapSize := int(float64(opts.MaxCharBuffer) * opts.OverlapRatio)
	step := opts.MaxCharBuffer - overlapSize
	
	for i := 0; i < len(text); i += step {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return nil, NewChunkingErrorWithCause(ErrorTypeTimeout, "chunking cancelled", ctx.Err())
		default:
		}
		
		end := i + opts.MaxCharBuffer
		if end > len(text) {
			end = len(text)
		}
		
		chunkText := text[i:end]
		if len(chunkText) < opts.MinChunkSize && i > 0 {
			break // Skip chunks that are too small (except the first chunk)
		}
		
		chunk := sc.createChunk(chunkText, i, len(chunks), opts)
		chunks = append(chunks, chunk)
		
		if end >= len(text) {
			break
		}
	}
	
	return chunks, nil
}

// findSentences identifies sentence boundaries in the text.
func (sc *SimpleChunker) findSentences(text string) []string {
	// Split by sentence endings, then clean up
	sentences := sc.sentencePattern.Split(text, -1)
	var result []string
	
	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if sentence != "" {
			result = append(result, sentence)
		}
	}
	
	return result
}

// createChunk creates a TextChunk with the specified parameters.
func (sc *SimpleChunker) createChunk(text string, startPos, index int, opts ChunkingOptions) TextChunk {
	charInterval, _ := types.NewCharInterval(startPos, startPos+len(text))
	
	chunk := TextChunk{
		ID:           fmt.Sprintf("chunk_%d", index),
		Text:         text,
		CharInterval: charInterval,
		ChunkIndex:   index,
		Metadata: ChunkMetadata{
			ChunkerName:   sc.Name(),
			SentenceCount: sc.countSentences(text),
			WordCount:     len(strings.Fields(text)),
			Quality:       sc.assessQuality(text, opts),
			Properties:    make(map[string]interface{}),
		},
	}
	
	return chunk
}

// finalizeChunks adds final metadata and calculates overlaps.
func (sc *SimpleChunker) finalizeChunks(chunks []TextChunk, originalText string, opts ChunkingOptions) []TextChunk {
	totalChunks := len(chunks)
	
	for i := range chunks {
		chunks[i].TotalChunks = totalChunks
		
		// Calculate overlaps if needed
		if opts.OverlapRatio > 0 && opts.IncludeMetadata {
			chunks[i].Metadata.HasOverlap = i > 0 || i < totalChunks-1
			
			if i > 0 {
				chunks[i].Metadata.OverlapStart = sc.calculateOverlap(chunks[i-1], chunks[i])
			}
			if i < totalChunks-1 {
				chunks[i].Metadata.OverlapEnd = sc.calculateOverlap(chunks[i], chunks[i+1])
			}
		}
		
		// Add algorithm-specific properties
		chunks[i].Metadata.Properties["preserve_sentences"] = opts.PreserveSentences
		chunks[i].Metadata.Properties["preserve_paragraphs"] = opts.PreserveParagraphs
		chunks[i].Metadata.Properties["max_char_buffer"] = opts.MaxCharBuffer
	}
	
	return chunks
}

// countSentences provides a simple sentence count for the text.
func (sc *SimpleChunker) countSentences(text string) int {
	sentences := sc.findSentences(text)
	return len(sentences)
}

// assessQuality provides a simple quality assessment of the chunk.
func (sc *SimpleChunker) assessQuality(text string, opts ChunkingOptions) float64 {
	// Simple quality assessment based on:
	// - Length utilization
	// - Sentence completeness
	// - Character distribution
	
	lengthUtilization := float64(len(text)) / float64(opts.MaxCharBuffer)
	if lengthUtilization > 1.0 {
		lengthUtilization = 1.0
	}
	
	// Check if text ends with sentence-ending punctuation
	text = strings.TrimSpace(text)
	sentenceComplete := 0.0
	if len(text) > 0 {
		lastChar := text[len(text)-1]
		if lastChar == '.' || lastChar == '!' || lastChar == '?' {
			sentenceComplete = 1.0
		}
	}
	
	// Simple character distribution check (avoid too many whitespace/special chars)
	alphaCount := 0
	for _, r := range text {
		if unicode.IsLetter(r) {
			alphaCount++
		}
	}
	alphaRatio := float64(alphaCount) / float64(len(text))
	
	// Weighted average of quality factors
	quality := 0.5*lengthUtilization + 0.3*sentenceComplete + 0.2*alphaRatio
	if quality > 1.0 {
		quality = 1.0
	}
	
	return quality
}

// calculateOverlap calculates the number of overlapping characters between two chunks.
func (sc *SimpleChunker) calculateOverlap(chunk1, chunk2 TextChunk) int {
	if chunk1.CharInterval == nil || chunk2.CharInterval == nil {
		return 0
	}
	
	intersection := chunk1.CharInterval.Intersection(*chunk2.CharInterval)
	if intersection == nil {
		return 0
	}
	
	return intersection.Length()
}