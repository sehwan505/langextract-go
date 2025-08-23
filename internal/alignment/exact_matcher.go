package alignment

import (
	"context"
	"strings"
	"time"

	"github.com/sehwan505/langextract-go/pkg/types"
)

// ExactMatcher implements exact string matching for text alignment.
// It finds exact matches between extracted text and source text with
// optional normalization for case and whitespace.
type ExactMatcher struct{}

// NewExactMatcher creates a new ExactMatcher.
func NewExactMatcher() *ExactMatcher {
	return &ExactMatcher{}
}

// Name returns the name of this alignment algorithm.
func (em *ExactMatcher) Name() string {
	return "ExactMatcher"
}

// AlignExtraction finds exact matches of extracted text in the source text.
func (em *ExactMatcher) AlignExtraction(ctx context.Context, extracted, source string, opts AlignmentOptions) (*types.CharInterval, *types.AlignmentResult, error) {
	startTime := time.Now()
	
	if err := opts.Validate(); err != nil {
		return nil, nil, err
	}
	
	if extracted == "" {
		return nil, nil, ErrEmptyExtracted.WithMethod(em.Name())
	}
	
	if source == "" {
		return nil, nil, ErrEmptySource.WithMethod(em.Name())
	}
	
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return nil, nil, NewAlignmentErrorWithCause(ErrorTypeTimeout, "alignment cancelled", ctx.Err()).WithMethod(em.Name())
	default:
	}
	
	// Normalize texts based on options
	normalizedExtracted := em.normalizeText(extracted, opts)
	normalizedSource := em.normalizeText(source, opts)
	
	// Find all exact matches
	matches := em.findExactMatches(normalizedExtracted, normalizedSource, extracted, source, opts)
	
	if len(matches) == 0 {
		return nil, nil, ErrNoAlignment.WithMethod(em.Name()).WithDetail("extracted_text", extracted)
	}
	
	// Select the best match based on strategy
	bestMatch := em.selectBestMatch(matches, opts)
	
	// Create character interval
	charInterval, err := types.NewCharInterval(bestMatch.Position, bestMatch.Position+bestMatch.Length)
	if err != nil {
		return nil, nil, NewAlignmentErrorWithCause(ErrorTypeProcessing, "failed to create character interval", err).WithMethod(em.Name())
	}
	
	// Create alignment result
	alignmentResult, err := types.NewAlignmentResult(
		types.AlignmentExact,
		bestMatch.Score,
		bestMatch.Score,
		em.Name(),
	)
	if err != nil {
		return nil, nil, NewAlignmentErrorWithCause(ErrorTypeProcessing, "failed to create alignment result", err).WithMethod(em.Name())
	}
	
	// Check confidence threshold
	if bestMatch.Score < opts.MinConfidence {
		return nil, nil, ErrLowConfidence.WithMethod(em.Name()).
			WithDetail("confidence", bestMatch.Score).
			WithDetail("threshold", opts.MinConfidence)
	}
	
	// Log processing time
	processingTime := time.Since(startTime).Milliseconds()
	if processingTime > opts.TimeoutMs {
		return nil, nil, NewAlignmentError(ErrorTypeTimeout, "alignment exceeded timeout").
			WithMethod(em.Name()).
			WithDetail("processing_time_ms", processingTime).
			WithDetail("timeout_ms", opts.TimeoutMs)
	}
	
	return charInterval, alignmentResult, nil
}

// AlignExtractions aligns multiple extractions in batch.
func (em *ExactMatcher) AlignExtractions(ctx context.Context, extractions []string, source string, opts AlignmentOptions) ([]AlignmentResult, error) {
	var results []AlignmentResult
	
	for i, extracted := range extractions {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return results, NewAlignmentErrorWithCause(ErrorTypeTimeout, "batch alignment cancelled", ctx.Err()).WithMethod(em.Name())
		default:
		}
		
		charInterval, alignmentInfo, err := em.AlignExtraction(ctx, extracted, source, opts)
		if err != nil {
			// Create a failed result
			result := AlignmentResult{
				ExtractedText: extracted,
				Confidence:    0.0,
				Metadata: AlignmentMetadata{
					AlgorithmName: em.Name(),
					ErrorCount:    1,
					Properties: map[string]interface{}{
						"error":           err.Error(),
						"extraction_index": i,
					},
				},
			}
			results = append(results, result)
			continue
		}
		
		// Get the aligned text
		alignedText := ""
		if charInterval != nil && charInterval.EndPos <= len(source) {
			alignedText = source[charInterval.StartPos:charInterval.EndPos]
		}
		
		// Create successful result
		result := AlignmentResult{
			CharInterval:  charInterval,
			AlignmentInfo: alignmentInfo,
			ExtractedText: extracted,
			AlignedText:   alignedText,
			Confidence:    alignmentInfo.Confidence,
			Metadata: AlignmentMetadata{
				AlgorithmName: em.Name(),
				Properties: map[string]interface{}{
					"extraction_index": i,
					"exact_match":      true,
				},
			},
		}
		
		results = append(results, result)
	}
	
	return results, nil
}

// FindBestAlignment finds the best exact match using multiple strategies.
func (em *ExactMatcher) FindBestAlignment(ctx context.Context, extracted, source string, opts AlignmentOptions) (*types.CharInterval, *types.AlignmentResult, error) {
	// For exact matching, this is the same as regular alignment
	return em.AlignExtraction(ctx, extracted, source, opts)
}

// ValidateAlignment checks if a proposed alignment is valid for exact matching.
func (em *ExactMatcher) ValidateAlignment(extracted, source string, interval *types.CharInterval) (float64, error) {
	if interval == nil {
		return 0.0, NewAlignmentError(ErrorTypeValidation, "interval cannot be nil").WithMethod(em.Name())
	}
	
	if interval.StartPos < 0 || interval.EndPos > len(source) {
		return 0.0, NewAlignmentError(ErrorTypeValidation, "interval out of bounds").
			WithMethod(em.Name()).
			WithDetail("interval", interval.String()).
			WithDetail("source_length", len(source))
	}
	
	alignedText := source[interval.StartPos:interval.EndPos]
	
	// Check for exact match
	if strings.EqualFold(extracted, alignedText) {
		return 1.0, nil
	}
	
	// Check for match with whitespace normalization
	if strings.EqualFold(strings.TrimSpace(extracted), strings.TrimSpace(alignedText)) {
		return 0.95, nil
	}
	
	// No match
	return 0.0, nil
}

// normalizeText applies normalization based on alignment options.
func (em *ExactMatcher) normalizeText(text string, opts AlignmentOptions) string {
	normalized := text
	
	if !opts.CaseSensitive {
		normalized = strings.ToLower(normalized)
	}
	
	if opts.IgnoreWhitespace {
		// Normalize whitespace (replace multiple spaces with single space, trim)
		normalized = strings.TrimSpace(normalized)
		normalized = strings.Join(strings.Fields(normalized), " ")
	}
	
	if opts.IgnorePunctuation {
		// Remove common punctuation marks
		punctuation := ".,!?;:()[]{}\"'-"
		for _, p := range punctuation {
			normalized = strings.ReplaceAll(normalized, string(p), "")
		}
		// Clean up any extra spaces created by punctuation removal
		normalized = strings.TrimSpace(normalized)
		normalized = strings.Join(strings.Fields(normalized), " ")
	}
	
	return normalized
}

// findExactMatches finds all exact matches in the source text.
func (em *ExactMatcher) findExactMatches(normalizedExtracted, normalizedSource, originalExtracted, originalSource string, opts AlignmentOptions) []AlignmentCandidate {
	var candidates []AlignmentCandidate
	
	if normalizedExtracted == "" {
		return candidates
	}
	
	// Find all occurrences of the normalized extracted text
	searchText := normalizedExtracted
	sourceText := normalizedSource
	
	start := 0
	for {
		pos := strings.Index(sourceText[start:], searchText)
		if pos == -1 {
			break
		}
		
		actualPos := start + pos
		
		// Map back to original text position
		originalPos := em.mapNormalizedToOriginal(actualPos, normalizedSource, originalSource, opts)
		originalLength := em.calculateOriginalLength(actualPos, len(searchText), normalizedSource, originalSource, opts)
		
		// Validate the match in the original text
		if originalPos >= 0 && originalPos+originalLength <= len(originalSource) {
			matchedText := originalSource[originalPos : originalPos+originalLength]
			score := em.calculateExactMatchScore(originalExtracted, matchedText, opts)
			
			candidate := AlignmentCandidate{
				Position:     originalPos,
				Length:       originalLength,
				Score:        score,
				Method:       em.Name(),
				MatchedText:  matchedText,
				EditDistance: 0, // Exact match
				Properties: map[string]interface{}{
					"normalized_position": actualPos,
					"exact_match":         true,
				},
			}
			
			candidates = append(candidates, candidate)
			
			// Limit the number of candidates
			if len(candidates) >= opts.MaxCandidates {
				break
			}
		}
		
		start += pos + 1
	}
	
	return candidates
}

// selectBestMatch selects the best match from candidates.
func (em *ExactMatcher) selectBestMatch(candidates []AlignmentCandidate, opts AlignmentOptions) AlignmentCandidate {
	if len(candidates) == 0 {
		return AlignmentCandidate{}
	}
	
	if len(candidates) == 1 {
		return candidates[0]
	}
	
	// For exact matching, prefer the first match unless there are tie-breaking criteria
	bestCandidate := candidates[0]
	
	for _, candidate := range candidates[1:] {
		// Prefer higher scores
		if candidate.Score > bestCandidate.Score {
			bestCandidate = candidate
			continue
		}
		
		// If scores are equal, prefer earlier positions (for consistency)
		if candidate.Score == bestCandidate.Score && candidate.Position < bestCandidate.Position {
			bestCandidate = candidate
		}
	}
	
	return bestCandidate
}

// calculateExactMatchScore calculates the quality score for an exact match.
func (em *ExactMatcher) calculateExactMatchScore(extracted, matched string, opts AlignmentOptions) float64 {
	// Start with perfect score for exact match
	score := 1.0
	
	// Check if it's truly exact (without normalization)
	if extracted == matched {
		return 1.0
	}
	
	// Apply penalties for normalization differences
	if !opts.CaseSensitive && strings.ToLower(extracted) == strings.ToLower(matched) {
		score = 0.98 // Slight penalty for case differences
	}
	
	if opts.IgnoreWhitespace {
		extractedNorm := strings.Join(strings.Fields(extracted), " ")
		matchedNorm := strings.Join(strings.Fields(matched), " ")
		if strings.EqualFold(extractedNorm, matchedNorm) {
			score = 0.95 // Penalty for whitespace differences
		}
	}
	
	if opts.IgnorePunctuation {
		// Additional penalty for punctuation differences
		score *= 0.92
	}
	
	return score
}

// mapNormalizedToOriginal maps a position in normalized text to original text.
func (em *ExactMatcher) mapNormalizedToOriginal(normalizedPos int, normalizedText, originalText string, opts AlignmentOptions) int {
	// Simple implementation: assume character-by-character mapping with adjustments
	// This is a simplified approach; a more sophisticated implementation would track
	// the exact mapping during normalization
	
	if normalizedPos == 0 {
		return 0
	}
	
	if normalizedPos >= len(normalizedText) {
		return len(originalText)
	}
	
	// For case-only changes, position mapping is direct
	if !opts.IgnoreWhitespace && !opts.IgnorePunctuation {
		if normalizedPos < len(originalText) {
			return normalizedPos
		}
		return len(originalText) - 1
	}
	
	// More complex mapping needed for whitespace/punctuation changes
	// This is a simplified approximation
	ratio := float64(normalizedPos) / float64(len(normalizedText))
	originalPos := int(ratio * float64(len(originalText)))
	
	if originalPos >= len(originalText) {
		originalPos = len(originalText) - 1
	}
	
	return originalPos
}

// calculateOriginalLength calculates the length in original text corresponding to normalized length.
func (em *ExactMatcher) calculateOriginalLength(normalizedPos, normalizedLength int, normalizedText, originalText string, opts AlignmentOptions) int {
	// Simple implementation: estimate based on text length ratios
	if normalizedLength == 0 {
		return 0
	}
	
	// For case-only changes, length is preserved
	if !opts.IgnoreWhitespace && !opts.IgnorePunctuation {
		return normalizedLength
	}
	
	// Estimate based on text compression ratio
	if len(normalizedText) == 0 {
		return normalizedLength
	}
	
	ratio := float64(len(originalText)) / float64(len(normalizedText))
	originalLength := int(float64(normalizedLength) * ratio)
	
	// Ensure we don't exceed bounds
	startPos := em.mapNormalizedToOriginal(normalizedPos, normalizedText, originalText, opts)
	if startPos+originalLength > len(originalText) {
		originalLength = len(originalText) - startPos
	}
	
	return originalLength
}