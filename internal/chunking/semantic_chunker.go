package chunking

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strings"
	"unicode"

	"github.com/sehwan505/langextract-go/pkg/document"
	"github.com/sehwan505/langextract-go/pkg/types"
)

// SemanticChunker implements context-aware text chunking that attempts to
// preserve semantic boundaries and topical coherence.
type SemanticChunker struct {
	sentencePattern  *regexp.Regexp
	paragraphPattern *regexp.Regexp
	topicPatterns    []*regexp.Regexp
}

// NewSemanticChunker creates a new SemanticChunker with enhanced boundary detection.
func NewSemanticChunker() *SemanticChunker {
	return &SemanticChunker{
		sentencePattern:  regexp.MustCompile(`[.!?]+\s+`),
		paragraphPattern: regexp.MustCompile(`\n\s*\n`),
		topicPatterns: []*regexp.Regexp{
			// Topic transition indicators
			regexp.MustCompile(`(?i)\b(however|furthermore|moreover|additionally|in contrast|on the other hand|meanwhile|subsequently|therefore|thus|consequently)\b`),
			// Section headers and lists
			regexp.MustCompile(`^\s*\d+[\.\)]\s+`), // Numbered lists
			regexp.MustCompile(`^\s*[•\-\*]\s+`),   // Bullet points
			regexp.MustCompile(`^[A-Z][A-Z\s]+$`),  // ALL CAPS headers
		},
	}
}

// Name returns the name of this chunking algorithm.
func (sc *SemanticChunker) Name() string {
	return "SemanticChunker"
}

// ChunkDocument splits a document into semantically coherent chunks.
func (sc *SemanticChunker) ChunkDocument(ctx context.Context, doc *document.Document, opts ChunkingOptions) ([]TextChunk, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	if doc == nil {
		return nil, NewChunkingError(ErrorTypeValidation, "document cannot be nil")
	}

	return sc.ChunkText(ctx, doc.Text, opts)
}

// ChunkText splits raw text into semantically coherent chunks.
func (sc *SemanticChunker) ChunkText(ctx context.Context, text string, opts ChunkingOptions) ([]TextChunk, error) {
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

	// Analyze text structure first
	textAnalysis := sc.analyzeTextStructure(text)
	
	// Create semantic chunks based on analysis
	chunks, err := sc.createSemanticChunks(ctx, text, textAnalysis, opts)
	if err != nil {
		return nil, err
	}

	// Apply overlap strategy if requested
	if opts.OverlapRatio > 0 {
		chunks = sc.applySemanticOverlap(chunks, text, opts)
	}

	// Finalize chunks with metadata
	return sc.finalizeChunks(chunks, text, opts), nil
}

// EstimateChunks provides an estimate considering semantic boundaries.
func (sc *SemanticChunker) EstimateChunks(text string, opts ChunkingOptions) int {
	if text == "" || opts.MaxCharBuffer <= 0 {
		return 0
	}
	
	analysis := sc.analyzeTextStructure(text)
	
	// Factor in structural elements for more accurate estimation
	structuralComplexity := 1.0
	if len(analysis.ParagraphBoundaries) > 0 {
		structuralComplexity += 0.2 // Paragraphs increase chunk count
	}
	if len(analysis.TopicTransitions) > 0 {
		structuralComplexity += 0.3 // Topic transitions increase chunk count
	}
	
	effectiveChunkSize := float64(opts.MaxCharBuffer) / structuralComplexity
	if opts.OverlapRatio > 0 {
		effectiveChunkSize *= (1.0 - opts.OverlapRatio)
	}
	
	estimate := int(math.Ceil(float64(len(text)) / effectiveChunkSize))
	if estimate == 0 && len(text) > 0 {
		return 1
	}
	return estimate
}

// TextStructureAnalysis contains information about the structure of the text.
type TextStructureAnalysis struct {
	SentenceBoundaries []int            // Character positions of sentence boundaries
	ParagraphBoundaries []int           // Character positions of paragraph boundaries
	TopicTransitions   []int            // Character positions of likely topic transitions
	SectionHeaders     []int            // Character positions of section headers
	ListItems          []int            // Character positions of list items
	KeywordDensity     map[string]float64 // Density of important keywords
}

// analyzeTextStructure performs structural analysis of the input text.
func (sc *SemanticChunker) analyzeTextStructure(text string) *TextStructureAnalysis {
	analysis := &TextStructureAnalysis{
		SentenceBoundaries:  []int{},
		ParagraphBoundaries: []int{},
		TopicTransitions:    []int{},
		SectionHeaders:      []int{},
		ListItems:           []int{},
		KeywordDensity:      make(map[string]float64),
	}

	// Find sentence boundaries
	sentenceMatches := sc.sentencePattern.FindAllStringIndex(text, -1)
	for _, match := range sentenceMatches {
		analysis.SentenceBoundaries = append(analysis.SentenceBoundaries, match[1])
	}

	// Find paragraph boundaries
	paragraphMatches := sc.paragraphPattern.FindAllStringIndex(text, -1)
	for _, match := range paragraphMatches {
		analysis.ParagraphBoundaries = append(analysis.ParagraphBoundaries, match[1])
	}

	// Find topic transitions and structural elements
	lines := strings.Split(text, "\n")
	currentPos := 0
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Check for various structural patterns
		for _, pattern := range sc.topicPatterns {
			if pattern.MatchString(line) {
				// Determine the type of match
				if strings.Contains(pattern.String(), `\d+[\.\)]`) {
					analysis.ListItems = append(analysis.ListItems, currentPos)
				} else if strings.Contains(pattern.String(), `[A-Z]`) {
					analysis.SectionHeaders = append(analysis.SectionHeaders, currentPos)
				} else {
					analysis.TopicTransitions = append(analysis.TopicTransitions, currentPos)
				}
			}
		}
		
		currentPos += len(line) + 1 // +1 for newline
	}

	// Analyze keyword density for better semantic understanding
	analysis.KeywordDensity = sc.calculateKeywordDensity(text)

	return analysis
}

// createSemanticChunks creates chunks based on semantic analysis.
func (sc *SemanticChunker) createSemanticChunks(ctx context.Context, text string, analysis *TextStructureAnalysis, opts ChunkingOptions) ([]TextChunk, error) {
	var chunks []TextChunk
	
	// Determine optimal split points based on structure analysis
	splitPoints := sc.determineSplitPoints(text, analysis, opts)
	
	// Create chunks between split points
	for i := 0; i < len(splitPoints)-1; i++ {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return nil, NewChunkingErrorWithCause(ErrorTypeTimeout, "chunking cancelled", ctx.Err())
		default:
		}
		
		start := splitPoints[i]
		end := splitPoints[i+1]
		
		chunkText := strings.TrimSpace(text[start:end])
		if len(chunkText) < opts.MinChunkSize && len(chunks) > 0 {
			// Merge small chunks with the previous chunk
			chunks[len(chunks)-1].Text += " " + chunkText
			chunks[len(chunks)-1].CharInterval.EndPos = end
			continue
		}
		
		if chunkText != "" {
			chunk := sc.createSemanticChunk(chunkText, start, len(chunks), analysis, opts)
			chunks = append(chunks, chunk)
		}
	}
	
	return chunks, nil
}

// determineSplitPoints finds optimal locations to split the text.
func (sc *SemanticChunker) determineSplitPoints(text string, analysis *TextStructureAnalysis, opts ChunkingOptions) []int {
	splitPoints := []int{0} // Start with beginning of text
	
	// Collect all potential split points with priorities
	type splitCandidate struct {
		position int
		priority float64
		reason   string
	}
	
	var candidates []splitCandidate
	
	// Add paragraph boundaries (high priority)
	for _, pos := range analysis.ParagraphBoundaries {
		candidates = append(candidates, splitCandidate{pos, 0.9, "paragraph"})
	}
	
	// Add topic transitions (high priority)
	for _, pos := range analysis.TopicTransitions {
		candidates = append(candidates, splitCandidate{pos, 0.8, "topic_transition"})
	}
	
	// Add section headers (highest priority)
	for _, pos := range analysis.SectionHeaders {
		candidates = append(candidates, splitCandidate{pos, 1.0, "section_header"})
	}
	
	// Add sentence boundaries (medium priority)
	for _, pos := range analysis.SentenceBoundaries {
		candidates = append(candidates, splitCandidate{pos, 0.6, "sentence"})
	}
	
	// Sort candidates by position
	for i := 0; i < len(candidates)-1; i++ {
		for j := i + 1; j < len(candidates); j++ {
			if candidates[i].position > candidates[j].position {
				candidates[i], candidates[j] = candidates[j], candidates[i]
			}
		}
	}
	
	// Select split points based on chunk size constraints
	lastSplit := 0
	for _, candidate := range candidates {
		distanceFromLast := candidate.position - lastSplit
		
		// If we're approaching the max buffer size, take the best available split
		if distanceFromLast >= int(float64(opts.MaxCharBuffer)*0.8) {
			splitPoints = append(splitPoints, candidate.position)
			lastSplit = candidate.position
		}
		// If we've exceeded the max buffer size, we must split
		if distanceFromLast >= opts.MaxCharBuffer {
			splitPoints = append(splitPoints, candidate.position)
			lastSplit = candidate.position
		}
	}
	
	// Add end of text
	if len(splitPoints) == 0 || splitPoints[len(splitPoints)-1] != len(text) {
		splitPoints = append(splitPoints, len(text))
	}
	
	return splitPoints
}

// createSemanticChunk creates a TextChunk with semantic metadata.
func (sc *SemanticChunker) createSemanticChunk(text string, startPos, index int, analysis *TextStructureAnalysis, opts ChunkingOptions) TextChunk {
	charInterval, _ := types.NewCharInterval(startPos, startPos+len(text))
	
	// Calculate semantic coherence score
	coherenceScore := sc.calculateSemanticCoherence(text, analysis)
	
	chunk := TextChunk{
		ID:           fmt.Sprintf("semantic_chunk_%d", index),
		Text:         text,
		CharInterval: charInterval,
		ChunkIndex:   index,
		Metadata: ChunkMetadata{
			ChunkerName:   sc.Name(),
			SentenceCount: sc.countSentences(text),
			WordCount:     len(strings.Fields(text)),
			Quality:       coherenceScore,
			Properties: map[string]interface{}{
				"semantic_coherence": coherenceScore,
				"has_topic_transition": sc.hasTopicTransition(text),
				"has_section_header":   sc.hasSectionHeader(text),
				"structural_complexity": sc.calculateStructuralComplexity(text),
			},
		},
	}
	
	return chunk
}

// applySemanticOverlap adds intelligent overlap between chunks.
func (sc *SemanticChunker) applySemanticOverlap(chunks []TextChunk, text string, opts ChunkingOptions) []TextChunk {
	if len(chunks) <= 1 {
		return chunks
	}
	
	overlapSize := int(float64(opts.MaxCharBuffer) * opts.OverlapRatio)
	
	for i := 1; i < len(chunks); i++ {
		// Find a good overlap point (prefer sentence boundaries)
		overlapStart := chunks[i].CharInterval.StartPos - overlapSize
		if overlapStart < chunks[i-1].CharInterval.StartPos {
			overlapStart = chunks[i-1].CharInterval.StartPos
		}
		
		// Try to align with sentence boundary
		overlapText := text[overlapStart:chunks[i].CharInterval.StartPos]
		sentenceBoundary := strings.LastIndex(overlapText, ". ")
		if sentenceBoundary > 0 {
			overlapStart += sentenceBoundary + 2
		}
		
		// Update chunk to include overlap
		newText := text[overlapStart:chunks[i].CharInterval.EndPos]
		chunks[i].Text = newText
		chunks[i].CharInterval.StartPos = overlapStart
		chunks[i].Metadata.HasOverlap = true
		chunks[i].Metadata.OverlapStart = chunks[i].CharInterval.StartPos - overlapStart
	}
	
	return chunks
}

// calculateKeywordDensity analyzes keyword frequency for semantic understanding.
func (sc *SemanticChunker) calculateKeywordDensity(text string) map[string]float64 {
	words := strings.Fields(strings.ToLower(text))
	wordCount := make(map[string]int)
	totalWords := len(words)
	
	// Count word frequencies
	for _, word := range words {
		// Clean word of punctuation
		cleanWord := ""
		for _, r := range word {
			if unicode.IsLetter(r) {
				cleanWord += string(r)
			}
		}
		if len(cleanWord) > 3 { // Only consider longer words
			wordCount[cleanWord]++
		}
	}
	
	// Calculate density scores
	density := make(map[string]float64)
	for word, count := range wordCount {
		if count > 1 { // Only include repeated words
			density[word] = float64(count) / float64(totalWords)
		}
	}
	
	return density
}

// calculateSemanticCoherence assesses the semantic coherence of a text chunk.
func (sc *SemanticChunker) calculateSemanticCoherence(text string, analysis *TextStructureAnalysis) float64 {
	// Multiple factors contribute to semantic coherence:
	// 1. Keyword repetition
	// 2. Sentence flow
	// 3. Structural consistency
	
	keywords := sc.calculateKeywordDensity(text)
	keywordScore := 0.0
	for _, density := range keywords {
		keywordScore += density
	}
	if keywordScore > 1.0 {
		keywordScore = 1.0
	}
	
	// Sentence flow (prefer chunks that don't break mid-sentence)
	sentenceFlowScore := 1.0
	if !strings.HasSuffix(strings.TrimSpace(text), ".") &&
		!strings.HasSuffix(strings.TrimSpace(text), "!") &&
		!strings.HasSuffix(strings.TrimSpace(text), "?") {
		sentenceFlowScore = 0.7
	}
	
	// Structural consistency (prefer chunks with consistent structure)
	structuralScore := 0.8 // Default moderate score
	if sc.hasSectionHeader(text) {
		structuralScore = 0.9 // Higher score for chunks with clear structure
	}
	
	// Weighted average
	coherence := 0.4*keywordScore + 0.3*sentenceFlowScore + 0.3*structuralScore
	return coherence
}

// Helper methods for semantic analysis
func (sc *SemanticChunker) hasTopicTransition(text string) bool {
	for _, pattern := range sc.topicPatterns[:1] { // Just check transition words
		if pattern.MatchString(text) {
			return true
		}
	}
	return false
}

func (sc *SemanticChunker) hasSectionHeader(text string) bool {
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		if regexp.MustCompile(`^[A-Z][A-Z\s]+$`).MatchString(strings.TrimSpace(line)) {
			return true
		}
	}
	return false
}

func (sc *SemanticChunker) calculateStructuralComplexity(text string) float64 {
	lines := strings.Split(text, "\n")
	structuralElements := 0
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Count various structural elements
		if regexp.MustCompile(`^\d+[\.\)]\s+`).MatchString(line) ||
			regexp.MustCompile(`^[•\-\*]\s+`).MatchString(line) ||
			regexp.MustCompile(`^[A-Z][A-Z\s]+$`).MatchString(line) {
			structuralElements++
		}
	}
	
	return float64(structuralElements) / float64(len(lines))
}

func (sc *SemanticChunker) countSentences(text string) int {
	sentences := sc.sentencePattern.Split(text, -1)
	count := 0
	for _, sentence := range sentences {
		if strings.TrimSpace(sentence) != "" {
			count++
		}
	}
	return count
}

func (sc *SemanticChunker) finalizeChunks(chunks []TextChunk, originalText string, opts ChunkingOptions) []TextChunk {
	totalChunks := len(chunks)
	
	for i := range chunks {
		chunks[i].TotalChunks = totalChunks
		
		// Add semantic-specific properties
		chunks[i].Metadata.Properties["semantic_analysis"] = true
		chunks[i].Metadata.Properties["overlap_strategy"] = "semantic"
		chunks[i].Metadata.Properties["max_char_buffer"] = opts.MaxCharBuffer
	}
	
	return chunks
}