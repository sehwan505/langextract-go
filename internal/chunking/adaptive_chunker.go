package chunking

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/sehwan505/langextract-go/pkg/document"
	"github.com/sehwan505/langextract-go/pkg/types"
)

// AdaptiveChunker implements dynamic chunk sizing based on content complexity,
// density, and structural characteristics.
type AdaptiveChunker struct {
	sentencePattern  *regexp.Regexp
	paragraphPattern *regexp.Regexp
	complexityThresholds ContentComplexityThresholds
}

// ContentComplexityThresholds defines thresholds for different complexity levels.
type ContentComplexityThresholds struct {
	LowComplexity    float64 // Below this = simple content (larger chunks)
	MediumComplexity float64 // Between low and medium = moderate content
	HighComplexity   float64 // Above medium = complex content (smaller chunks)
}

// ContentMetrics holds various metrics used to assess content complexity.
type ContentMetrics struct {
	AverageWordsPerSentence  float64
	AverageSyllablesPerWord  float64
	UniqueWordRatio          float64
	PunctuationDensity       float64
	NumberDensity           float64
	StructuralComplexity     float64
	OverallComplexity        float64
}

// NewAdaptiveChunker creates a new AdaptiveChunker with default complexity thresholds.
func NewAdaptiveChunker() *AdaptiveChunker {
	return &AdaptiveChunker{
		sentencePattern:  regexp.MustCompile(`[.!?]+\s+`),
		paragraphPattern: regexp.MustCompile(`\n\s*\n`),
		complexityThresholds: ContentComplexityThresholds{
			LowComplexity:    0.3,
			MediumComplexity: 0.6,
			HighComplexity:   0.8,
		},
	}
}

// WithComplexityThresholds allows customization of complexity thresholds.
func (ac *AdaptiveChunker) WithComplexityThresholds(thresholds ContentComplexityThresholds) *AdaptiveChunker {
	ac.complexityThresholds = thresholds
	return ac
}

// Name returns the name of this chunking algorithm.
func (ac *AdaptiveChunker) Name() string {
	return "AdaptiveChunker"
}

// ChunkDocument splits a document into adaptively-sized chunks.
func (ac *AdaptiveChunker) ChunkDocument(ctx context.Context, doc *document.Document, opts ChunkingOptions) ([]TextChunk, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	if doc == nil {
		return nil, NewChunkingError(ErrorTypeValidation, "document cannot be nil")
	}

	return ac.ChunkText(ctx, doc.Text, opts)
}

// ChunkText splits raw text into adaptively-sized chunks.
func (ac *AdaptiveChunker) ChunkText(ctx context.Context, text string, opts ChunkingOptions) ([]TextChunk, error) {
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

	// Analyze content complexity for the entire text
	globalMetrics := ac.analyzeContentComplexity(text)
	
	// Create adaptive chunks based on local complexity
	chunks, err := ac.createAdaptiveChunks(ctx, text, globalMetrics, opts)
	if err != nil {
		return nil, err
	}

	// Apply complexity-aware overlap
	if opts.OverlapRatio > 0 {
		chunks = ac.applyAdaptiveOverlap(chunks, text, opts)
	}

	// Finalize chunks with adaptive metadata
	return ac.finalizeAdaptiveChunks(chunks, globalMetrics, opts), nil
}

// EstimateChunks provides an estimate considering adaptive sizing.
func (ac *AdaptiveChunker) EstimateChunks(text string, opts ChunkingOptions) int {
	if text == "" || opts.MaxCharBuffer <= 0 {
		return 0
	}
	
	// Analyze complexity to estimate adaptive behavior
	metrics := ac.analyzeContentComplexity(text)
	
	// Adjust effective chunk size based on complexity
	adaptiveFactor := ac.calculateAdaptiveFactor(metrics.OverallComplexity)
	effectiveChunkSize := float64(opts.MaxCharBuffer) * adaptiveFactor
	
	if opts.OverlapRatio > 0 {
		effectiveChunkSize *= (1.0 - opts.OverlapRatio)
	}
	
	estimate := int(math.Ceil(float64(len(text)) / effectiveChunkSize))
	if estimate == 0 && len(text) > 0 {
		return 1
	}
	return estimate
}

// analyzeContentComplexity performs comprehensive complexity analysis.
func (ac *AdaptiveChunker) analyzeContentComplexity(text string) ContentMetrics {
	metrics := ContentMetrics{}
	
	// Basic text statistics
	words := strings.Fields(text)
	sentences := ac.findSentences(text)
	
	if len(sentences) > 0 {
		metrics.AverageWordsPerSentence = float64(len(words)) / float64(len(sentences))
	}
	
	// Calculate unique word ratio
	if len(words) > 0 {
		uniqueWords := make(map[string]bool)
		for _, word := range words {
			uniqueWords[strings.ToLower(strings.Trim(word, ".,!?;:"))] = true
		}
		metrics.UniqueWordRatio = float64(len(uniqueWords)) / float64(len(words))
	}
	
	// Estimate syllables per word (simplified heuristic)
	if len(words) > 0 {
		totalSyllables := 0
		for _, word := range words {
			totalSyllables += ac.estimateSyllables(word)
		}
		metrics.AverageSyllablesPerWord = float64(totalSyllables) / float64(len(words))
	}
	
	// Calculate punctuation density
	punctuationCount := 0
	numberCount := 0
	for _, r := range text {
		if strings.ContainsRune(".,!?;:()[]{}\"'-", r) {
			punctuationCount++
		}
		if r >= '0' && r <= '9' {
			numberCount++
		}
	}
	
	if len(text) > 0 {
		metrics.PunctuationDensity = float64(punctuationCount) / float64(len(text))
		metrics.NumberDensity = float64(numberCount) / float64(len(text))
	}
	
	// Calculate structural complexity
	metrics.StructuralComplexity = ac.calculateStructuralComplexity(text)
	
	// Calculate overall complexity score
	metrics.OverallComplexity = ac.calculateOverallComplexity(metrics)
	
	return metrics
}

// createAdaptiveChunks creates chunks with sizes adapted to content complexity.
func (ac *AdaptiveChunker) createAdaptiveChunks(ctx context.Context, text string, globalMetrics ContentMetrics, opts ChunkingOptions) ([]TextChunk, error) {
	var chunks []TextChunk
	
	// Split text into initial segments for analysis
	segments := ac.createInitialSegments(text, opts)
	
	for i, segment := range segments {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return nil, NewChunkingErrorWithCause(ErrorTypeTimeout, "chunking cancelled", ctx.Err())
		default:
		}
		
		// Analyze local complexity
		localMetrics := ac.analyzeContentComplexity(segment.Text)
		
		// Determine adaptive chunk size for this segment
		adaptiveSize := ac.calculateAdaptiveChunkSize(localMetrics, opts.MaxCharBuffer)
		
		// Create chunks from this segment
		segmentChunks, err := ac.createChunksFromSegment(ctx, segment, adaptiveSize, len(chunks), opts)
		if err != nil {
			return nil, err
		}
		
		// Add segment-specific metadata
		for j := range segmentChunks {
			segmentChunks[j].Metadata.Properties["local_complexity"] = localMetrics.OverallComplexity
			segmentChunks[j].Metadata.Properties["adaptive_size"] = adaptiveSize
			segmentChunks[j].Metadata.Properties["segment_index"] = i
		}
		
		chunks = append(chunks, segmentChunks...)
	}
	
	return chunks, nil
}

// TextSegment represents a segment of text with position information.
type TextSegment struct {
	Text      string
	StartPos  int
	EndPos    int
}

// createInitialSegments splits text into segments for local analysis.
func (ac *AdaptiveChunker) createInitialSegments(text string, opts ChunkingOptions) []TextSegment {
	var segments []TextSegment
	
	// Use paragraph boundaries as initial segment boundaries
	paragraphs := ac.paragraphPattern.Split(text, -1)
	currentPos := 0
	
	for _, paragraph := range paragraphs {
		paragraph = strings.TrimSpace(paragraph)
		if paragraph == "" {
			continue
		}
		
		// Find the actual position of this paragraph in the original text
		segmentStart := strings.Index(text[currentPos:], paragraph) + currentPos
		segmentEnd := segmentStart + len(paragraph)
		
		segments = append(segments, TextSegment{
			Text:     paragraph,
			StartPos: segmentStart,
			EndPos:   segmentEnd,
		})
		
		currentPos = segmentEnd
	}
	
	// If no paragraphs found, treat entire text as one segment
	if len(segments) == 0 {
		segments = append(segments, TextSegment{
			Text:     text,
			StartPos: 0,
			EndPos:   len(text),
		})
	}
	
	return segments
}

// calculateAdaptiveChunkSize determines the optimal chunk size based on complexity.
func (ac *AdaptiveChunker) calculateAdaptiveChunkSize(metrics ContentMetrics, maxBuffer int) int {
	adaptiveFactor := ac.calculateAdaptiveFactor(metrics.OverallComplexity)
	return int(float64(maxBuffer) * adaptiveFactor)
}

// calculateAdaptiveFactor determines how to scale chunk size based on complexity.
func (ac *AdaptiveChunker) calculateAdaptiveFactor(complexity float64) float64 {
	// More complex content gets smaller chunks, simpler content gets larger chunks
	if complexity <= ac.complexityThresholds.LowComplexity {
		return 1.2 // 20% larger chunks for simple content
	} else if complexity <= ac.complexityThresholds.MediumComplexity {
		return 1.0 // Standard size for medium complexity
	} else if complexity <= ac.complexityThresholds.HighComplexity {
		return 0.8 // 20% smaller chunks for complex content
	} else {
		return 0.6 // 40% smaller chunks for very complex content
	}
}

// createChunksFromSegment creates appropriately-sized chunks from a text segment.
func (ac *AdaptiveChunker) createChunksFromSegment(ctx context.Context, segment TextSegment, targetSize int, startIndex int, opts ChunkingOptions) ([]TextChunk, error) {
	var chunks []TextChunk
	
	if len(segment.Text) <= targetSize {
		// Segment fits in one chunk
		chunk := ac.createAdaptiveChunk(segment.Text, segment.StartPos, startIndex, opts)
		chunks = append(chunks, chunk)
		return chunks, nil
	}
	
	// Split segment into multiple chunks
	sentences := ac.findSentences(segment.Text)
	currentChunk := ""
	currentStart := segment.StartPos
	chunkIndex := startIndex
	
	for _, sentence := range sentences {
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
		
		// If adding this sentence would exceed target size, finalize current chunk
		if len(currentChunk)+len(sentence) > targetSize && currentChunk != "" {
			chunk := ac.createAdaptiveChunk(currentChunk, currentStart, chunkIndex, opts)
			chunks = append(chunks, chunk)
			
			// Start new chunk
			currentChunk = sentence
			currentStart = strings.Index(segment.Text[currentStart:], sentence) + segment.StartPos
			chunkIndex++
		} else {
			if currentChunk != "" {
				currentChunk += " " + sentence
			} else {
				currentChunk = sentence
				currentStart = strings.Index(segment.Text, sentence) + segment.StartPos
			}
		}
	}
	
	// Add final chunk if any content remains
	if currentChunk != "" {
		chunk := ac.createAdaptiveChunk(currentChunk, currentStart, chunkIndex, opts)
		chunks = append(chunks, chunk)
	}
	
	return chunks, nil
}

// createAdaptiveChunk creates a TextChunk with adaptive metadata.
func (ac *AdaptiveChunker) createAdaptiveChunk(text string, startPos, index int, opts ChunkingOptions) TextChunk {
	charInterval, _ := types.NewCharInterval(startPos, startPos+len(text))
	
	// Analyze this specific chunk
	chunkMetrics := ac.analyzeContentComplexity(text)
	
	chunk := TextChunk{
		ID:           fmt.Sprintf("adaptive_chunk_%d", index),
		Text:         text,
		CharInterval: charInterval,
		ChunkIndex:   index,
		Metadata: ChunkMetadata{
			ChunkerName:   ac.Name(),
			SentenceCount: ac.countSentences(text),
			WordCount:     len(strings.Fields(text)),
			Quality:       ac.calculateAdaptiveQuality(text, chunkMetrics),
			Properties: map[string]interface{}{
				"complexity_score":        chunkMetrics.OverallComplexity,
				"avg_words_per_sentence":  chunkMetrics.AverageWordsPerSentence,
				"unique_word_ratio":       chunkMetrics.UniqueWordRatio,
				"punctuation_density":     chunkMetrics.PunctuationDensity,
				"structural_complexity":   chunkMetrics.StructuralComplexity,
				"adaptive_sizing":         true,
			},
		},
	}
	
	return chunk
}

// applyAdaptiveOverlap applies intelligent overlap based on complexity.
func (ac *AdaptiveChunker) applyAdaptiveOverlap(chunks []TextChunk, text string, opts ChunkingOptions) []TextChunk {
	if len(chunks) <= 1 {
		return chunks
	}
	
	for i := 1; i < len(chunks); i++ {
		// Calculate adaptive overlap size based on complexity
		complexity := chunks[i].Metadata.Properties["complexity_score"].(float64)
		adaptiveOverlapRatio := opts.OverlapRatio
		
		// More complex content gets larger overlap for better context preservation
		if complexity > ac.complexityThresholds.HighComplexity {
			adaptiveOverlapRatio *= 1.5
		} else if complexity < ac.complexityThresholds.LowComplexity {
			adaptiveOverlapRatio *= 0.75
		}
		
		overlapSize := int(float64(opts.MaxCharBuffer) * adaptiveOverlapRatio)
		
		// Apply the overlap
		overlapStart := chunks[i].CharInterval.StartPos - overlapSize
		if overlapStart < chunks[i-1].CharInterval.StartPos {
			overlapStart = chunks[i-1].CharInterval.StartPos
		}
		
		// Try to align with sentence boundary for better context
		overlapText := text[overlapStart:chunks[i].CharInterval.StartPos]
		sentenceBoundary := strings.LastIndex(overlapText, ". ")
		if sentenceBoundary > 0 {
			overlapStart += sentenceBoundary + 2
		}
		
		// Update chunk
		newText := text[overlapStart:chunks[i].CharInterval.EndPos]
		chunks[i].Text = newText
		chunks[i].CharInterval.StartPos = overlapStart
		chunks[i].Metadata.HasOverlap = true
		chunks[i].Metadata.OverlapStart = chunks[i].CharInterval.StartPos - overlapStart
		chunks[i].Metadata.Properties["adaptive_overlap_ratio"] = adaptiveOverlapRatio
	}
	
	return chunks
}

// Helper methods for complexity analysis

func (ac *AdaptiveChunker) estimateSyllables(word string) int {
	// Simple heuristic for syllable estimation
	word = strings.ToLower(strings.Trim(word, ".,!?;:"))
	if len(word) == 0 {
		return 0
	}
	
	vowels := "aeiouy"
	syllableCount := 0
	previousWasVowel := false
	
	for _, r := range word {
		isVowel := strings.ContainsRune(vowels, r)
		if isVowel && !previousWasVowel {
			syllableCount++
		}
		previousWasVowel = isVowel
	}
	
	// Adjust for silent e
	if strings.HasSuffix(word, "e") && syllableCount > 1 {
		syllableCount--
	}
	
	// Ensure at least one syllable
	if syllableCount == 0 {
		syllableCount = 1
	}
	
	return syllableCount
}

func (ac *AdaptiveChunker) calculateStructuralComplexity(text string) float64 {
	lines := strings.Split(text, "\n")
	structuralElements := 0
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Count various structural elements
		if regexp.MustCompile(`^\d+[\.\)]\s+`).MatchString(line) {
			structuralElements++ // Numbered lists
		}
		if regexp.MustCompile(`^[•\-\*]\s+`).MatchString(line) {
			structuralElements++ // Bullet points
		}
		if regexp.MustCompile(`^[A-Z][A-Z\s]+$`).MatchString(line) {
			structuralElements++ // Headers
		}
		if strings.Contains(line, "\t") {
			structuralElements++ // Indentation
		}
	}
	
	// Normalize by total lines
	if len(lines) == 0 {
		return 0
	}
	return float64(structuralElements) / float64(len(lines))
}

func (ac *AdaptiveChunker) calculateOverallComplexity(metrics ContentMetrics) float64 {
	// Weighted combination of complexity factors
	weights := map[string]float64{
		"sentence_length": 0.25,
		"syllable_complexity": 0.20,
		"vocabulary_diversity": 0.20,
		"punctuation_density": 0.15,
		"number_density": 0.10,
		"structural_complexity": 0.10,
	}
	
	// Normalize individual metrics to 0-1 scale
	normalizedSentenceLength := math.Min(metrics.AverageWordsPerSentence/25.0, 1.0)
	normalizedSyllableComplexity := math.Min(metrics.AverageSyllablesPerWord/3.0, 1.0)
	normalizedVocabularyDiversity := metrics.UniqueWordRatio
	normalizedPunctuationDensity := math.Min(metrics.PunctuationDensity*20.0, 1.0)
	normalizedNumberDensity := math.Min(metrics.NumberDensity*50.0, 1.0)
	normalizedStructuralComplexity := math.Min(metrics.StructuralComplexity*5.0, 1.0)
	
	// Calculate weighted score
	complexity := weights["sentence_length"]*normalizedSentenceLength +
		weights["syllable_complexity"]*normalizedSyllableComplexity +
		weights["vocabulary_diversity"]*normalizedVocabularyDiversity +
		weights["punctuation_density"]*normalizedPunctuationDensity +
		weights["number_density"]*normalizedNumberDensity +
		weights["structural_complexity"]*normalizedStructuralComplexity
	
	return math.Min(complexity, 1.0)
}

func (ac *AdaptiveChunker) calculateAdaptiveQuality(text string, metrics ContentMetrics) float64 {
	// Quality assessment considering adaptive sizing appropriateness
	baseQuality := 0.8 // Start with good base quality
	
	// Adjust based on complexity handling
	complexityAlignment := 1.0 - math.Abs(metrics.OverallComplexity-0.5)*2.0
	if complexityAlignment < 0 {
		complexityAlignment = 0
	}
	
	// Consider sentence completeness
	sentenceComplete := 0.0
	text = strings.TrimSpace(text)
	if len(text) > 0 {
		lastChar := text[len(text)-1]
		if lastChar == '.' || lastChar == '!' || lastChar == '?' {
			sentenceComplete = 1.0
		} else {
			sentenceComplete = 0.6
		}
	}
	
	// Weighted quality score
	quality := 0.5*baseQuality + 0.3*complexityAlignment + 0.2*sentenceComplete
	return math.Min(quality, 1.0)
}

func (ac *AdaptiveChunker) findSentences(text string) []string {
	sentences := ac.sentencePattern.Split(text, -1)
	var result []string
	
	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if sentence != "" {
			result = append(result, sentence)
		}
	}
	
	return result
}

func (ac *AdaptiveChunker) countSentences(text string) int {
	return len(ac.findSentences(text))
}

func (ac *AdaptiveChunker) finalizeAdaptiveChunks(chunks []TextChunk, globalMetrics ContentMetrics, opts ChunkingOptions) []TextChunk {
	totalChunks := len(chunks)
	
	for i := range chunks {
		chunks[i].TotalChunks = totalChunks
		
		// Add global context information
		chunks[i].Metadata.Properties["global_complexity"] = globalMetrics.OverallComplexity
		chunks[i].Metadata.Properties["adaptive_chunking"] = true
		chunks[i].Metadata.Properties["complexity_thresholds"] = ac.complexityThresholds
	}
	
	return chunks
}