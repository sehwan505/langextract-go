package alignment

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/sehwan505/langextract-go/pkg/types"
)

// DefaultMultiAligner implements the MultiAligner interface using multiple
// alignment algorithms with priority-based selection.
type DefaultMultiAligner struct {
	aligners []prioritizedAligner
	mutex    sync.RWMutex
}

// prioritizedAligner wraps an aligner with its priority.
type prioritizedAligner struct {
	aligner  TextAligner
	priority int
	name     string
}

// NewMultiAligner creates a new MultiAligner with default aligners.
func NewMultiAligner() *DefaultMultiAligner {
	ma := &DefaultMultiAligner{
		aligners: make([]prioritizedAligner, 0),
	}
	
	// Register default aligners with priorities
	ma.RegisterAligner(NewExactMatcher(), 100)    // Highest priority for exact matches
	ma.RegisterAligner(NewFuzzyMatcher(), 50)     // Medium priority for fuzzy matches
	
	return ma
}

// RegisterAligner adds a new alignment algorithm to the pool.
func (ma *DefaultMultiAligner) RegisterAligner(aligner TextAligner, priority int) error {
	if aligner == nil {
		return NewAlignmentError(ErrorTypeValidation, "aligner cannot be nil")
	}
	
	ma.mutex.Lock()
	defer ma.mutex.Unlock()
	
	// Check if aligner with same name already exists
	name := aligner.Name()
	for i, existing := range ma.aligners {
		if existing.name == name {
			// Update existing aligner
			ma.aligners[i] = prioritizedAligner{
				aligner:  aligner,
				priority: priority,
				name:     name,
			}
			ma.sortAlignersLocked()
			return nil
		}
	}
	
	// Add new aligner
	ma.aligners = append(ma.aligners, prioritizedAligner{
		aligner:  aligner,
		priority: priority,
		name:     name,
	})
	
	ma.sortAlignersLocked()
	return nil
}

// AlignWithBestMethod tries multiple alignment methods and returns the best result.
func (ma *DefaultMultiAligner) AlignWithBestMethod(ctx context.Context, extracted, source string, opts AlignmentOptions) (*types.CharInterval, *types.AlignmentResult, error) {
	ma.mutex.RLock()
	aligners := make([]prioritizedAligner, len(ma.aligners))
	copy(aligners, ma.aligners)
	ma.mutex.RUnlock()
	
	if len(aligners) == 0 {
		return nil, nil, NewAlignmentError(ErrorTypeInternal, "no aligners registered")
	}
	
	// Try aligners in priority order
	var bestInterval *types.CharInterval
	var bestResult *types.AlignmentResult
	var bestScore float64
	var firstError error
	
	for _, prioritized := range aligners {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return nil, nil, NewAlignmentErrorWithCause(ErrorTypeTimeout, "multi-alignment cancelled", ctx.Err())
		default:
		}
		
		interval, result, err := prioritized.aligner.AlignExtraction(ctx, extracted, source, opts)
		
		if err == nil && result != nil {
			score := result.Score
			
			// Apply priority bonus to score for comparison
			priorityBonus := float64(prioritized.priority) / 1000.0 // Small bonus to break ties
			adjustedScore := score + priorityBonus
			
			if adjustedScore > bestScore {
				bestInterval = interval
				bestResult = result
				bestScore = score // Store original score, not adjusted
			}
			
			// If we found a high-quality exact match, we can stop early
			if result.Status == types.AlignmentExact && score >= 0.95 {
				break
			}
		} else if firstError == nil {
			firstError = err
		}
	}
	
	if bestInterval == nil {
		if firstError != nil {
			return nil, nil, firstError
		}
		return nil, nil, ErrNoAlignment.WithDetail("methods_tried", len(aligners))
	}
	
	return bestInterval, bestResult, nil
}

// AlignWithAllMethods tries all registered methods and returns all results.
func (ma *DefaultMultiAligner) AlignWithAllMethods(ctx context.Context, extracted, source string, opts AlignmentOptions) ([]AlignmentResult, error) {
	ma.mutex.RLock()
	aligners := make([]prioritizedAligner, len(ma.aligners))
	copy(aligners, ma.aligners)
	ma.mutex.RUnlock()
	
	if len(aligners) == 0 {
		return nil, NewAlignmentError(ErrorTypeInternal, "no aligners registered")
	}
	
	results := make([]AlignmentResult, 0, len(aligners))
	
	// Try all aligners
	for _, prioritized := range aligners {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return results, NewAlignmentErrorWithCause(ErrorTypeTimeout, "multi-alignment cancelled", ctx.Err())
		default:
		}
		
		startTime := time.Now()
		interval, alignmentInfo, err := prioritized.aligner.AlignExtraction(ctx, extracted, source, opts)
		processingTime := time.Since(startTime).Milliseconds()
		
		var result AlignmentResult
		
		if err == nil && interval != nil && alignmentInfo != nil {
			// Successful alignment
			alignedText := ""
			if interval.EndPos <= len(source) {
				alignedText = source[interval.StartPos:interval.EndPos]
			}
			
			result = AlignmentResult{
				CharInterval:  interval,
				AlignmentInfo: alignmentInfo,
				ExtractedText: extracted,
				AlignedText:   alignedText,
				Confidence:    alignmentInfo.Confidence,
				Metadata: AlignmentMetadata{
					AlgorithmName:    prioritized.name,
					ProcessingTimeMs: processingTime,
					AttemptedMethods: []string{prioritized.name},
					ErrorCount:       0,
					FallbackUsed:     false,
					Properties: map[string]interface{}{
						"priority":     prioritized.priority,
						"method_index": len(results),
					},
				},
			}
		} else {
			// Failed alignment
			result = AlignmentResult{
				ExtractedText: extracted,
				Confidence:    0.0,
				Metadata: AlignmentMetadata{
					AlgorithmName:    prioritized.name,
					ProcessingTimeMs: processingTime,
					AttemptedMethods: []string{prioritized.name},
					ErrorCount:       1,
					FallbackUsed:     false,
					Properties: map[string]interface{}{
						"priority":     prioritized.priority,
						"method_index": len(results),
						"error":        err.Error(),
					},
				},
			}
		}
		
		results = append(results, result)
	}
	
	return results, nil
}

// GetAvailableAligners returns the list of registered alignment algorithms.
func (ma *DefaultMultiAligner) GetAvailableAligners() []string {
	ma.mutex.RLock()
	defer ma.mutex.RUnlock()
	
	names := make([]string, len(ma.aligners))
	for i, aligner := range ma.aligners {
		names[i] = aligner.name
	}
	
	return names
}

// AlignWithStrategy aligns using a specific strategy for result selection.
func (ma *DefaultMultiAligner) AlignWithStrategy(ctx context.Context, extracted, source string, opts AlignmentOptions, strategy AlignmentStrategy) (*types.CharInterval, *types.AlignmentResult, error) {
	// Get all results first
	allResults, err := ma.AlignWithAllMethods(ctx, extracted, source, opts)
	if err != nil {
		return nil, nil, err
	}
	
	// Filter valid results
	validResults := make([]AlignmentResult, 0)
	for _, result := range allResults {
		if result.IsValid() {
			validResults = append(validResults, result)
		}
	}
	
	if len(validResults) == 0 {
		return nil, nil, ErrNoAlignment.WithDetail("strategy", strategy.String())
	}
	
	// Apply strategy to select best result
	selectedResult := ma.applySelectionStrategy(validResults, strategy)
	
	return selectedResult.CharInterval, selectedResult.AlignmentInfo, nil
}

// applySelectionStrategy applies the specified strategy to select the best result.
func (ma *DefaultMultiAligner) applySelectionStrategy(results []AlignmentResult, strategy AlignmentStrategy) AlignmentResult {
	if len(results) == 0 {
		return AlignmentResult{}
	}
	
	if len(results) == 1 {
		return results[0]
	}
	
	switch strategy {
	case StrategyBestScore:
		return ma.selectByBestScore(results)
	case StrategyFirstFound:
		return results[0] // Assumes results are in priority order
	case StrategyMostConfident:
		return ma.selectByConfidence(results)
	case StrategyExactPreferred:
		return ma.selectExactPreferred(results)
	case StrategyPositionBased:
		return ma.selectByPosition(results)
	default:
		return ma.selectByBestScore(results) // Default to best score
	}
}

// selectByBestScore selects the result with the highest score.
func (ma *DefaultMultiAligner) selectByBestScore(results []AlignmentResult) AlignmentResult {
	best := results[0]
	for _, result := range results[1:] {
		if result.GetScore() > best.GetScore() {
			best = result
		}
	}
	return best
}

// selectByConfidence selects the result with the highest confidence.
func (ma *DefaultMultiAligner) selectByConfidence(results []AlignmentResult) AlignmentResult {
	best := results[0]
	for _, result := range results[1:] {
		if result.Confidence > best.Confidence {
			best = result
		}
	}
	return best
}

// selectExactPreferred prefers exact matches over other types.
func (ma *DefaultMultiAligner) selectExactPreferred(results []AlignmentResult) AlignmentResult {
	// First, try to find exact matches
	for _, result := range results {
		if result.AlignmentInfo != nil && result.AlignmentInfo.Status == types.AlignmentExact {
			return result
		}
	}
	
	// If no exact matches, fall back to best score
	return ma.selectByBestScore(results)
}

// selectByPosition selects based on position preferences (earlier positions preferred).
func (ma *DefaultMultiAligner) selectByPosition(results []AlignmentResult) AlignmentResult {
	best := results[0]
	for _, result := range results[1:] {
		// Prefer earlier positions for same quality
		if result.GetScore() >= best.GetScore() {
			if result.CharInterval != nil && best.CharInterval != nil {
				if result.CharInterval.StartPos < best.CharInterval.StartPos {
					best = result
				}
			}
		}
	}
	return best
}

// sortAlignersLocked sorts aligners by priority (highest first).
// Must be called while holding a write lock.
func (ma *DefaultMultiAligner) sortAlignersLocked() {
	sort.Slice(ma.aligners, func(i, j int) bool {
		return ma.aligners[i].priority > ma.aligners[j].priority
	})
}

// RemoveAligner removes an aligner by name.
func (ma *DefaultMultiAligner) RemoveAligner(name string) bool {
	ma.mutex.Lock()
	defer ma.mutex.Unlock()
	
	for i, aligner := range ma.aligners {
		if aligner.name == name {
			// Remove by swapping with last element and truncating
			ma.aligners[i] = ma.aligners[len(ma.aligners)-1]
			ma.aligners = ma.aligners[:len(ma.aligners)-1]
			ma.sortAlignersLocked()
			return true
		}
	}
	
	return false
}

// GetAlignerPriority returns the priority of a registered aligner.
func (ma *DefaultMultiAligner) GetAlignerPriority(name string) (int, bool) {
	ma.mutex.RLock()
	defer ma.mutex.RUnlock()
	
	for _, aligner := range ma.aligners {
		if aligner.name == name {
			return aligner.priority, true
		}
	}
	
	return 0, false
}

// SetAlignerPriority updates the priority of a registered aligner.
func (ma *DefaultMultiAligner) SetAlignerPriority(name string, priority int) bool {
	ma.mutex.Lock()
	defer ma.mutex.Unlock()
	
	for i, aligner := range ma.aligners {
		if aligner.name == name {
			ma.aligners[i].priority = priority
			ma.sortAlignersLocked()
			return true
		}
	}
	
	return false
}

// ClearAligners removes all registered aligners.
func (ma *DefaultMultiAligner) ClearAligners() {
	ma.mutex.Lock()
	defer ma.mutex.Unlock()
	
	ma.aligners = make([]prioritizedAligner, 0)
}

// GetStats returns statistics about the registered aligners.
func (ma *DefaultMultiAligner) GetStats() map[string]interface{} {
	ma.mutex.RLock()
	defer ma.mutex.RUnlock()
	
	stats := map[string]interface{}{
		"total_aligners": len(ma.aligners),
		"aligners":       make([]map[string]interface{}, len(ma.aligners)),
	}
	
	for i, aligner := range ma.aligners {
		stats["aligners"].([]map[string]interface{})[i] = map[string]interface{}{
			"name":     aligner.name,
			"priority": aligner.priority,
		}
	}
	
	return stats
}