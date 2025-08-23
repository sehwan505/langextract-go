package alignment

import (
	"context"
	"math"
	"strings"
	"time"

	"github.com/sehwan505/langextract-go/pkg/types"
)

// FuzzyMatcher implements fuzzy string matching using Levenshtein distance
// for text alignment. It can find approximate matches when exact matching fails.
type FuzzyMatcher struct {
	maxEditDistance int
}

// NewFuzzyMatcher creates a new FuzzyMatcher with default maximum edit distance.
func NewFuzzyMatcher() *FuzzyMatcher {
	return &FuzzyMatcher{
		maxEditDistance: 5,
	}
}

// NewFuzzyMatcherWithDistance creates a new FuzzyMatcher with specified maximum edit distance.
func NewFuzzyMatcherWithDistance(maxDistance int) *FuzzyMatcher {
	return &FuzzyMatcher{
		maxEditDistance: maxDistance,
	}
}

// Name returns the name of this alignment algorithm.
func (fm *FuzzyMatcher) Name() string {
	return "FuzzyMatcher"
}

// AlignExtraction finds fuzzy matches of extracted text in the source text.
func (fm *FuzzyMatcher) AlignExtraction(ctx context.Context, extracted, source string, opts AlignmentOptions) (*types.CharInterval, *types.AlignmentResult, error) {
	startTime := time.Now()
	
	if err := opts.Validate(); err != nil {
		return nil, nil, err
	}
	
	if extracted == "" {
		return nil, nil, ErrEmptyExtracted.WithMethod(fm.Name())
	}
	
	if source == "" {
		return nil, nil, ErrEmptySource.WithMethod(fm.Name())
	}
	
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return nil, nil, NewAlignmentErrorWithCause(ErrorTypeTimeout, "alignment cancelled", ctx.Err()).WithMethod(fm.Name())
	default:
	}
	
	// Use the smaller of opts.MaxDistance or fm.maxEditDistance
	maxDistance := fm.maxEditDistance
	if opts.MaxDistance > 0 && opts.MaxDistance < maxDistance {
		maxDistance = opts.MaxDistance
	}
	
	// Normalize texts based on options
	normalizedExtracted := fm.normalizeText(extracted, opts)
	normalizedSource := fm.normalizeText(source, opts)
	
	// Find fuzzy matches
	matches := fm.findFuzzyMatches(normalizedExtracted, normalizedSource, extracted, source, maxDistance, opts)
	
	if len(matches) == 0 {
		return nil, nil, ErrNoAlignment.WithMethod(fm.Name()).WithDetail("extracted_text", extracted)
	}
	
	// Select the best match
	bestMatch := fm.selectBestFuzzyMatch(matches, opts)
	
	// Create character interval
	charInterval, err := types.NewCharInterval(bestMatch.Position, bestMatch.Position+bestMatch.Length)
	if err != nil {
		return nil, nil, NewAlignmentErrorWithCause(ErrorTypeProcessing, "failed to create character interval", err).WithMethod(fm.Name())
	}
	
	// Determine alignment status based on edit distance
	alignmentStatus := fm.getAlignmentStatus(bestMatch.EditDistance, len(normalizedExtracted))
	
	// Create alignment result
	alignmentResult, err := types.NewAlignmentResult(
		alignmentStatus,
		bestMatch.Score,
		bestMatch.Score,
		fm.Name(),
	)
	if err != nil {
		return nil, nil, NewAlignmentErrorWithCause(ErrorTypeProcessing, "failed to create alignment result", err).WithMethod(fm.Name())
	}
	
	// Check confidence threshold
	if bestMatch.Score < opts.MinConfidence {
		return nil, nil, ErrLowConfidence.WithMethod(fm.Name()).
			WithDetail("confidence", bestMatch.Score).
			WithDetail("threshold", opts.MinConfidence)
	}
	
	// Check timeout
	processingTime := time.Since(startTime).Milliseconds()
	if processingTime > opts.TimeoutMs {
		return nil, nil, NewAlignmentError(ErrorTypeTimeout, "alignment exceeded timeout").
			WithMethod(fm.Name()).
			WithDetail("processing_time_ms", processingTime).
			WithDetail("timeout_ms", opts.TimeoutMs)
	}
	
	return charInterval, alignmentResult, nil
}

// AlignExtractions aligns multiple extractions using fuzzy matching.
func (fm *FuzzyMatcher) AlignExtractions(ctx context.Context, extractions []string, source string, opts AlignmentOptions) ([]AlignmentResult, error) {
	var results []AlignmentResult
	
	for i, extracted := range extractions {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return results, NewAlignmentErrorWithCause(ErrorTypeTimeout, "batch alignment cancelled", ctx.Err()).WithMethod(fm.Name())
		default:
		}
		
		charInterval, alignmentInfo, err := fm.AlignExtraction(ctx, extracted, source, opts)
		if err != nil {
			// Create a failed result
			result := AlignmentResult{
				ExtractedText: extracted,
				Confidence:    0.0,
				Metadata: AlignmentMetadata{
					AlgorithmName: fm.Name(),
					ErrorCount:    1,
					Properties: map[string]interface{}{
						"error":            err.Error(),
						"extraction_index": i,
						"fuzzy_matching":   true,
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
				AlgorithmName: fm.Name(),
				Properties: map[string]interface{}{
					"extraction_index": i,
					"fuzzy_matching":   true,
					"edit_distance":    fm.calculateEditDistance(extracted, alignedText),
				},
			},
		}
		
		results = append(results, result)
	}
	
	return results, nil
}

// FindBestAlignment finds the best fuzzy match using multiple strategies.
func (fm *FuzzyMatcher) FindBestAlignment(ctx context.Context, extracted, source string, opts AlignmentOptions) (*types.CharInterval, *types.AlignmentResult, error) {
	// Try with different distance thresholds for better results
	originalMaxDistance := opts.MaxDistance
	
	// Try with progressively larger edit distances
	distances := []int{0, 1, 2, 3, opts.MaxDistance}
	if opts.MaxDistance > 5 {
		distances = append(distances, opts.MaxDistance/2)
	}
	
	var bestInterval *types.CharInterval
	var bestResult *types.AlignmentResult
	var bestScore float64
	
	for _, distance := range distances {
		if distance > opts.MaxDistance {
			continue
		}
		
		opts.MaxDistance = distance
		interval, result, err := fm.AlignExtraction(ctx, extracted, source, opts)
		
		if err == nil && result != nil && result.Score > bestScore {
			bestInterval = interval
			bestResult = result
			bestScore = result.Score
			
			// If we found an exact match, stop searching
			if distance == 0 {
				break
			}
		}
	}
	
	// Restore original max distance
	opts.MaxDistance = originalMaxDistance
	
	if bestInterval == nil {
		return nil, nil, ErrNoAlignment.WithMethod(fm.Name())
	}
	
	return bestInterval, bestResult, nil
}

// ValidateAlignment validates a proposed alignment using fuzzy matching.
func (fm *FuzzyMatcher) ValidateAlignment(extracted, source string, interval *types.CharInterval) (float64, error) {
	if interval == nil {
		return 0.0, NewAlignmentError(ErrorTypeValidation, "interval cannot be nil").WithMethod(fm.Name())
	}
	
	if interval.StartPos < 0 || interval.EndPos > len(source) {
		return 0.0, NewAlignmentError(ErrorTypeValidation, "interval out of bounds").
			WithMethod(fm.Name()).
			WithDetail("interval", interval.String()).
			WithDetail("source_length", len(source))
	}
	
	alignedText := source[interval.StartPos:interval.EndPos]
	editDistance := fm.calculateEditDistance(extracted, alignedText)
	
	// Calculate confidence based on edit distance
	maxLength := len(extracted)
	if len(alignedText) > maxLength {
		maxLength = len(alignedText)
	}
	
	if maxLength == 0 {
		return 1.0, nil // Both strings are empty
	}
	
	similarity := 1.0 - (float64(editDistance) / float64(maxLength))
	if similarity < 0 {
		similarity = 0
	}
	
	return similarity, nil
}

// normalizeText applies normalization for fuzzy matching.
func (fm *FuzzyMatcher) normalizeText(text string, opts AlignmentOptions) string {
	normalized := text
	
	if !opts.CaseSensitive {
		normalized = strings.ToLower(normalized)
	}
	
	if opts.IgnoreWhitespace {
		normalized = strings.TrimSpace(normalized)
		normalized = strings.Join(strings.Fields(normalized), " ")
	}
	
	if opts.IgnorePunctuation {
		punctuation := ".,!?;:()[]{}\"'-"
		for _, p := range punctuation {
			normalized = strings.ReplaceAll(normalized, string(p), "")
		}
		normalized = strings.TrimSpace(normalized)
		normalized = strings.Join(strings.Fields(normalized), " ")
	}
	
	return normalized
}

// findFuzzyMatches finds all fuzzy matches within the specified edit distance.
func (fm *FuzzyMatcher) findFuzzyMatches(normalizedExtracted, normalizedSource, originalExtracted, originalSource string, maxDistance int, opts AlignmentOptions) []AlignmentCandidate {
	var candidates []AlignmentCandidate
	
	if normalizedExtracted == "" {
		return candidates
	}
	
	extractedLen := len(normalizedExtracted)
	sourceLen := len(normalizedSource)
	
	// Sliding window approach for fuzzy matching
	windowSizes := fm.getWindowSizes(extractedLen, maxDistance)
	
	for _, windowSize := range windowSizes {
		for start := 0; start <= sourceLen-windowSize; start++ {
			// Check for timeout periodically
			if len(candidates)%100 == 0 {
				select {
				case <-context.Background().Done():
					return candidates
				default:
				}
			}
			
			window := normalizedSource[start : start+windowSize]
			editDistance := fm.calculateEditDistance(normalizedExtracted, window)
			
			if editDistance <= maxDistance {
				// Map back to original positions
				originalStart := fm.mapNormalizedToOriginal(start, normalizedSource, originalSource, opts)
				originalEnd := fm.mapNormalizedToOriginal(start+windowSize, normalizedSource, originalSource, opts)
				originalLength := originalEnd - originalStart
				
				if originalStart >= 0 && originalEnd <= len(originalSource) && originalLength > 0 {
					matchedText := originalSource[originalStart:originalEnd]
					score := fm.calculateFuzzyScore(originalExtracted, matchedText, editDistance)
					
					candidate := AlignmentCandidate{
						Position:     originalStart,
						Length:       originalLength,
						Score:        score,
						Method:       fm.Name(),
						MatchedText:  matchedText,
						EditDistance: editDistance,
						Properties: map[string]interface{}{
							"window_size":         windowSize,
							"normalized_position": start,
							"fuzzy_match":         true,
						},
					}
					
					candidates = append(candidates, candidate)
					
					// Limit candidates to prevent memory issues
					if len(candidates) >= opts.MaxCandidates {
						return candidates
					}
				}
			}
		}
	}
	
	return candidates
}

// getWindowSizes returns appropriate window sizes for fuzzy matching.
func (fm *FuzzyMatcher) getWindowSizes(targetLength, maxDistance int) []int {
	var sizes []int
	
	// Try exact length first
	sizes = append(sizes, targetLength)
	
	// Try lengths within edit distance bounds
	for delta := 1; delta <= maxDistance; delta++ {
		// Shorter windows
		if targetLength-delta > 0 {
			sizes = append(sizes, targetLength-delta)
		}
		// Longer windows
		sizes = append(sizes, targetLength+delta)
	}
	
	// Remove duplicates and sort
	sizeMap := make(map[int]bool)
	var uniqueSizes []int
	for _, size := range sizes {
		if !sizeMap[size] && size > 0 {
			sizeMap[size] = true
			uniqueSizes = append(uniqueSizes, size)
		}
	}
	
	return uniqueSizes
}

// selectBestFuzzyMatch selects the best fuzzy match from candidates.
func (fm *FuzzyMatcher) selectBestFuzzyMatch(candidates []AlignmentCandidate, opts AlignmentOptions) AlignmentCandidate {
	if len(candidates) == 0 {
		return AlignmentCandidate{}
	}
	
	if len(candidates) == 1 {
		return candidates[0]
	}
	
	bestCandidate := candidates[0]
	
	for _, candidate := range candidates[1:] {
		// Prefer lower edit distance
		if candidate.EditDistance < bestCandidate.EditDistance {
			bestCandidate = candidate
			continue
		}
		
		if candidate.EditDistance == bestCandidate.EditDistance {
			// If edit distances are equal, prefer higher score
			if candidate.Score > bestCandidate.Score {
				bestCandidate = candidate
				continue
			}
			
			// If scores are also equal, prefer earlier position
			if candidate.Score == bestCandidate.Score && candidate.Position < bestCandidate.Position {
				bestCandidate = candidate
			}
		}
	}
	
	return bestCandidate
}

// calculateEditDistance computes the Levenshtein distance between two strings.
func (fm *FuzzyMatcher) calculateEditDistance(s1, s2 string) int {
	len1, len2 := len(s1), len(s2)
	
	// Create a matrix to store distances
	matrix := make([][]int, len1+1)
	for i := range matrix {
		matrix[i] = make([]int, len2+1)
	}
	
	// Initialize first row and column
	for i := 0; i <= len1; i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len2; j++ {
		matrix[0][j] = j
	}
	
	// Fill the matrix
	for i := 1; i <= len1; i++ {
		for j := 1; j <= len2; j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}
			
			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}
	
	return matrix[len1][len2]
}

// calculateFuzzyScore calculates a quality score for a fuzzy match.
func (fm *FuzzyMatcher) calculateFuzzyScore(original, matched string, editDistance int) float64 {
	maxLength := len(original)
	if len(matched) > maxLength {
		maxLength = len(matched)
	}
	
	if maxLength == 0 {
		return 1.0
	}
	
	// Base similarity score
	similarity := 1.0 - (float64(editDistance) / float64(maxLength))
	if similarity < 0 {
		similarity = 0
	}
	
	// Bonus for length similarity
	lengthSimilarity := 1.0 - math.Abs(float64(len(original)-len(matched)))/float64(maxLength)
	
	// Weighted combination
	score := 0.8*similarity + 0.2*lengthSimilarity
	
	return score
}

// getAlignmentStatus determines the alignment status based on edit distance.
func (fm *FuzzyMatcher) getAlignmentStatus(editDistance, textLength int) types.AlignmentStatus {
	if editDistance == 0 {
		return types.AlignmentExact
	}
	
	ratio := float64(editDistance) / float64(textLength)
	
	if ratio <= 0.1 {
		return types.AlignmentFuzzy
	} else if ratio <= 0.3 {
		return types.AlignmentPartial
	} else {
		return types.AlignmentApproximate
	}
}

// mapNormalizedToOriginal maps normalized position to original text position.
func (fm *FuzzyMatcher) mapNormalizedToOriginal(normalizedPos int, normalizedText, originalText string, opts AlignmentOptions) int {
	if normalizedPos == 0 {
		return 0
	}
	
	if normalizedPos >= len(normalizedText) {
		return len(originalText)
	}
	
	// Simple ratio-based mapping
	ratio := float64(normalizedPos) / float64(len(normalizedText))
	originalPos := int(ratio * float64(len(originalText)))
	
	if originalPos >= len(originalText) {
		originalPos = len(originalText) - 1
	}
	
	return originalPos
}

// min returns the minimum of three integers.
func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}