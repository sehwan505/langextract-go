package alignment_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/sehwan505/langextract-go/internal/alignment"
	"github.com/sehwan505/langextract-go/pkg/types"
)

// TestExactMatcher tests the ExactMatcher implementation
func TestExactMatcher(t *testing.T) {
	matcher := alignment.NewExactMatcher()
	
	t.Run("Name", func(t *testing.T) {
		if matcher.Name() != "ExactMatcher" {
			t.Errorf("Expected 'ExactMatcher', got %s", matcher.Name())
		}
	})
	
	t.Run("AlignExtraction_ExactMatch", func(t *testing.T) {
		extracted := "test phrase"
		source := "This is a test phrase in the document."
		opts := alignment.DefaultAlignmentOptions()
		
		interval, result, err := matcher.AlignExtraction(context.Background(), extracted, source, opts)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		if interval == nil {
			t.Fatal("Expected non-nil interval")
		}
		
		if result == nil {
			t.Fatal("Expected non-nil result")
		}
		
		if result.Status != types.AlignmentExact {
			t.Errorf("Expected exact alignment, got %s", result.Status)
		}
		
		if result.Confidence < 0.9 {
			t.Errorf("Expected high confidence for exact match, got %f", result.Confidence)
		}
		
		// Check that the interval correctly identifies the phrase
		alignedText := source[interval.StartPos:interval.EndPos]
		if !strings.Contains(strings.ToLower(alignedText), strings.ToLower(extracted)) {
			t.Errorf("Aligned text '%s' doesn't contain extracted text '%s'", alignedText, extracted)
		}
	})
	
	t.Run("AlignExtraction_CaseInsensitive", func(t *testing.T) {
		extracted := "TEST PHRASE"
		source := "This is a test phrase in the document."
		opts := alignment.DefaultAlignmentOptions().WithCaseSensitive(false)
		
		interval, result, err := matcher.AlignExtraction(context.Background(), extracted, source, opts)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		if interval == nil {
			t.Fatal("Expected non-nil interval")
		}
		
		if result.Status != types.AlignmentExact {
			t.Errorf("Expected exact alignment with case insensitive, got %s", result.Status)
		}
	})
	
	t.Run("AlignExtraction_IgnoreWhitespace", func(t *testing.T) {
		extracted := "test   phrase"
		source := "This is a test phrase in the document."
		opts := alignment.DefaultAlignmentOptions().WithIgnoreWhitespace(true)
		
		interval, result, err := matcher.AlignExtraction(context.Background(), extracted, source, opts)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		if interval == nil {
			t.Fatal("Expected non-nil interval")
		}
		
		if result.Status != types.AlignmentExact {
			t.Errorf("Expected exact alignment with whitespace normalization, got %s", result.Status)
		}
	})
	
	t.Run("AlignExtraction_NoMatch", func(t *testing.T) {
		extracted := "nonexistent phrase"
		source := "This is a test phrase in the document."
		opts := alignment.DefaultAlignmentOptions()
		
		_, _, err := matcher.AlignExtraction(context.Background(), extracted, source, opts)
		if err == nil {
			t.Error("Expected error for non-existent phrase")
		}
		
		// Check that it's a "not found" error
		if alignErr, ok := err.(*alignment.AlignmentError); ok {
			if !alignErr.IsType("not_found") {
				t.Errorf("Expected 'not_found' error type, got %s", alignErr.Type)
			}
		} else {
			t.Error("Expected AlignmentError type")
		}
	})
	
	t.Run("AlignExtraction_EmptyInputs", func(t *testing.T) {
		opts := alignment.DefaultAlignmentOptions()
		
		// Empty extracted text
		_, _, err := matcher.AlignExtraction(context.Background(), "", "some source", opts)
		if err == nil {
			t.Error("Expected error for empty extracted text")
		}
		
		// Empty source text
		_, _, err = matcher.AlignExtraction(context.Background(), "some text", "", opts)
		if err == nil {
			t.Error("Expected error for empty source text")
		}
	})
	
	t.Run("AlignExtractions_Batch", func(t *testing.T) {
		extractions := []string{"test phrase", "document", "another phrase"}
		source := "This is a test phrase in the document with another phrase."
		opts := alignment.DefaultAlignmentOptions()
		
		results, err := matcher.AlignExtractions(context.Background(), extractions, source, opts)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		if len(results) != len(extractions) {
			t.Errorf("Expected %d results, got %d", len(extractions), len(results))
		}
		
		// Check that most extractions were successfully aligned
		successCount := 0
		for _, result := range results {
			if result.IsValid() {
				successCount++
			}
		}
		
		if successCount < 2 {
			t.Errorf("Expected at least 2 successful alignments, got %d", successCount)
		}
	})
	
	t.Run("ValidateAlignment", func(t *testing.T) {
		extracted := "test phrase"
		source := "This is a test phrase in the document."
		
		// Create a valid interval
		interval := &types.CharInterval{StartPos: 10, EndPos: 21}
		
		confidence, err := matcher.ValidateAlignment(extracted, source, interval)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		if confidence < 0.9 {
			t.Errorf("Expected high confidence for valid alignment, got %f", confidence)
		}
		
		// Test invalid interval
		invalidInterval := &types.CharInterval{StartPos: -1, EndPos: 5}
		_, err = matcher.ValidateAlignment(extracted, source, invalidInterval)
		if err == nil {
			t.Error("Expected error for invalid interval")
		}
	})
	
	t.Run("ContextCancellation", func(t *testing.T) {
		extracted := "test phrase"
		source := "This is a test phrase in the document."
		opts := alignment.DefaultAlignmentOptions()
		
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately
		
		_, _, err := matcher.AlignExtraction(ctx, extracted, source, opts)
		if err == nil {
			t.Error("Expected error for cancelled context")
		}
	})
}

// TestFuzzyMatcher tests the FuzzyMatcher implementation
func TestFuzzyMatcher(t *testing.T) {
	matcher := alignment.NewFuzzyMatcher()
	
	t.Run("Name", func(t *testing.T) {
		if matcher.Name() != "FuzzyMatcher" {
			t.Errorf("Expected 'FuzzyMatcher', got %s", matcher.Name())
		}
	})
	
	t.Run("AlignExtraction_ExactMatch", func(t *testing.T) {
		extracted := "test phrase"
		source := "This is a test phrase in the document."
		opts := alignment.DefaultAlignmentOptions()
		
		interval, result, err := matcher.AlignExtraction(context.Background(), extracted, source, opts)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		if interval == nil {
			t.Fatal("Expected non-nil interval")
		}
		
		if result == nil {
			t.Fatal("Expected non-nil result")
		}
		
		if result.Status != types.AlignmentExact {
			t.Errorf("Expected exact alignment, got %s", result.Status)
		}
	})
	
	t.Run("AlignExtraction_FuzzyMatch", func(t *testing.T) {
		extracted := "tast phrase" // Typo: 'e' -> 'a'
		source := "This is a test phrase in the document."
		opts := alignment.DefaultAlignmentOptions().WithMaxDistance(2)
		
		interval, result, err := matcher.AlignExtraction(context.Background(), extracted, source, opts)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		if interval == nil {
			t.Fatal("Expected non-nil interval")
		}
		
		if result.Status == types.AlignmentNone {
			t.Error("Expected some alignment for fuzzy match")
		}
		
		// Fuzzy match should have lower confidence than exact match
		if result.Confidence > 0.95 {
			t.Errorf("Expected lower confidence for fuzzy match, got %f", result.Confidence)
		}
	})
	
	t.Run("AlignExtraction_MaxDistance", func(t *testing.T) {
		extracted := "completely different text"
		source := "This is a test phrase in the document."
		opts := alignment.DefaultAlignmentOptions().WithMaxDistance(1) // Very low tolerance
		
		_, _, err := matcher.AlignExtraction(context.Background(), extracted, source, opts)
		if err == nil {
			t.Error("Expected error for text that exceeds max distance")
		}
	})
	
	t.Run("AlignExtraction_WithCustomDistance", func(t *testing.T) {
		customMatcher := alignment.NewFuzzyMatcherWithDistance(3)
		
		extracted := "tset phraze" // Multiple typos
		source := "This is a test phrase in the document."
		opts := alignment.DefaultAlignmentOptions().WithMaxDistance(5)
		
		interval, result, err := customMatcher.AlignExtraction(context.Background(), extracted, source, opts)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		if interval == nil {
			t.Fatal("Expected non-nil interval for custom distance matcher")
		}
		
		if result.Status == types.AlignmentNone {
			t.Error("Expected some alignment with higher distance tolerance")
		}
	})
	
	t.Run("FindBestAlignment", func(t *testing.T) {
		extracted := "tast phrase"
		source := "This is a test phrase in the document."
		opts := alignment.DefaultAlignmentOptions()
		
		interval, result, err := matcher.FindBestAlignment(context.Background(), extracted, source, opts)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		if interval == nil {
			t.Fatal("Expected non-nil interval")
		}
		
		// Best alignment should find a reasonable match
		if result.Score < 0.5 {
			t.Errorf("Expected reasonable score for best alignment, got %f", result.Score)
		}
	})
	
	t.Run("ValidateAlignment_EditDistance", func(t *testing.T) {
		extracted := "tast phrase"
		source := "This is a test phrase in the document."
		
		// Create interval that should match the corrected phrase
		interval := &types.CharInterval{StartPos: 10, EndPos: 21}
		
		confidence, err := matcher.ValidateAlignment(extracted, source, interval)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		// Should have good confidence even with edit distance
		if confidence < 0.7 {
			t.Errorf("Expected reasonable confidence for fuzzy alignment, got %f", confidence)
		}
	})
}

// TestMultiAligner tests the MultiAligner implementation
func TestMultiAligner(t *testing.T) {
	aligner := alignment.NewMultiAligner()
	
	t.Run("GetAvailableAligners", func(t *testing.T) {
		aligners := aligner.GetAvailableAligners()
		if len(aligners) < 2 {
			t.Errorf("Expected at least 2 default aligners, got %d", len(aligners))
		}
		
		expectedAligners := map[string]bool{
			"ExactMatcher": false,
			"FuzzyMatcher": false,
		}
		
		for _, name := range aligners {
			if _, exists := expectedAligners[name]; exists {
				expectedAligners[name] = true
			}
		}
		
		for name, found := range expectedAligners {
			if !found {
				t.Errorf("Expected to find aligner %s", name)
			}
		}
	})
	
	t.Run("RegisterAligner", func(t *testing.T) {
		customAligner := alignment.NewExactMatcher()
		
		err := aligner.RegisterAligner(customAligner, 150) // Higher priority
		if err != nil {
			t.Fatalf("Unexpected error registering aligner: %v", err)
		}
		
		// Test registering nil aligner
		err = aligner.RegisterAligner(nil, 100)
		if err == nil {
			t.Error("Expected error for nil aligner")
		}
	})
	
	t.Run("AlignWithBestMethod", func(t *testing.T) {
		extracted := "test phrase"
		source := "This is a test phrase in the document."
		opts := alignment.DefaultAlignmentOptions()
		
		interval, result, err := aligner.AlignWithBestMethod(context.Background(), extracted, source, opts)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		if interval == nil {
			t.Fatal("Expected non-nil interval")
		}
		
		if result == nil {
			t.Fatal("Expected non-nil result")
		}
		
		// Should find a good alignment
		if result.Score < 0.9 {
			t.Errorf("Expected high score for best method, got %f", result.Score)
		}
	})
	
	t.Run("AlignWithAllMethods", func(t *testing.T) {
		extracted := "test phrase"
		source := "This is a test phrase in the document."
		opts := alignment.DefaultAlignmentOptions()
		
		results, err := aligner.AlignWithAllMethods(context.Background(), extracted, source, opts)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		if len(results) < 2 {
			t.Errorf("Expected at least 2 results, got %d", len(results))
		}
		
		// At least one result should be successful
		successCount := 0
		for _, result := range results {
			if result.IsValid() {
				successCount++
			}
		}
		
		if successCount == 0 {
			t.Error("Expected at least one successful alignment")
		}
	})
	
	t.Run("AlignWithStrategy", func(t *testing.T) {
		extracted := "test phrase"
		source := "This is a test phrase in the document with test phrase again."
		opts := alignment.DefaultAlignmentOptions()
		
		// Test different strategies
		strategies := []alignment.AlignmentStrategy{
			alignment.StrategyBestScore,
			alignment.StrategyFirstFound,
			alignment.StrategyMostConfident,
			alignment.StrategyExactPreferred,
		}
		
		for _, strategy := range strategies {
			interval, result, err := aligner.AlignWithStrategy(context.Background(), extracted, source, opts, strategy)
			if err != nil {
				t.Errorf("Unexpected error for strategy %s: %v", strategy.String(), err)
				continue
			}
			
			if interval == nil {
				t.Errorf("Expected non-nil interval for strategy %s", strategy.String())
				continue
			}
			
			if result == nil {
				t.Errorf("Expected non-nil result for strategy %s", strategy.String())
				continue
			}
		}
	})
	
	t.Run("RemoveAligner", func(t *testing.T) {
		// Try to remove an existing aligner
		removed := aligner.RemoveAligner("ExactMatcher")
		if !removed {
			t.Error("Expected to successfully remove ExactMatcher")
		}
		
		// Try to remove non-existent aligner
		removed = aligner.RemoveAligner("NonExistentMatcher")
		if removed {
			t.Error("Expected to fail removing non-existent aligner")
		}
	})
	
	t.Run("PriorityManagement", func(t *testing.T) {
		// Test getting priority
		if priority, exists := aligner.GetAlignerPriority("FuzzyMatcher"); exists {
			if priority <= 0 {
				t.Errorf("Expected positive priority, got %d", priority)
			}
		} else {
			t.Error("Expected to find FuzzyMatcher")
		}
		
		// Test setting priority
		success := aligner.SetAlignerPriority("FuzzyMatcher", 200)
		if !success {
			t.Error("Expected to successfully set priority")
		}
		
		// Verify priority was set
		if priority, exists := aligner.GetAlignerPriority("FuzzyMatcher"); exists {
			if priority != 200 {
				t.Errorf("Expected priority 200, got %d", priority)
			}
		}
	})
	
	t.Run("GetStats", func(t *testing.T) {
		stats := aligner.GetStats()
		
		if totalAligners, ok := stats["total_aligners"].(int); ok {
			if totalAligners <= 0 {
				t.Errorf("Expected positive number of aligners, got %d", totalAligners)
			}
		} else {
			t.Error("Expected total_aligners in stats")
		}
		
		if alignerList, ok := stats["aligners"].([]map[string]interface{}); ok {
			if len(alignerList) == 0 {
				t.Error("Expected non-empty aligner list")
			}
		} else {
			t.Error("Expected aligners list in stats")
		}
	})
	
	t.Run("ClearAligners", func(t *testing.T) {
		aligner.ClearAligners()
		
		aligners := aligner.GetAvailableAligners()
		if len(aligners) != 0 {
			t.Errorf("Expected 0 aligners after clear, got %d", len(aligners))
		}
	})
}

// TestAlignmentOptions tests the alignment options and validation
func TestAlignmentOptions(t *testing.T) {
	t.Run("DefaultOptions", func(t *testing.T) {
		opts := alignment.DefaultAlignmentOptions()
		
		if opts.MaxDistance < 0 {
			t.Error("Expected non-negative max distance")
		}
		if opts.MinConfidence < 0 || opts.MinConfidence > 1 {
			t.Error("Expected min confidence between 0 and 1")
		}
		if opts.MaxCandidates <= 0 {
			t.Error("Expected positive max candidates")
		}
		if opts.TimeoutMs <= 0 {
			t.Error("Expected positive timeout")
		}
	})
	
	t.Run("OptionsValidation", func(t *testing.T) {
		// Test invalid max distance
		opts := alignment.AlignmentOptions{
			MaxDistance:     -1,
			MinConfidence:   0.7,
			MaxCandidates:   10,
			TimeoutMs:       5000,
		}
		if err := opts.Validate(); err == nil {
			t.Error("Expected validation error for negative max distance")
		}
		
		// Test invalid min confidence
		opts = alignment.AlignmentOptions{
			MaxDistance:     5,
			MinConfidence:   1.5, // Too high
			MaxCandidates:   10,
			TimeoutMs:       5000,
		}
		if err := opts.Validate(); err == nil {
			t.Error("Expected validation error for high min confidence")
		}
		
		// Test invalid max candidates
		opts = alignment.AlignmentOptions{
			MaxDistance:     5,
			MinConfidence:   0.7,
			MaxCandidates:   0, // Zero or negative
			TimeoutMs:       5000,
		}
		if err := opts.Validate(); err == nil {
			t.Error("Expected validation error for zero max candidates")
		}
		
		// Test invalid timeout
		opts = alignment.AlignmentOptions{
			MaxDistance:     5,
			MinConfidence:   0.7,
			MaxCandidates:   10,
			TimeoutMs:       0, // Zero or negative
		}
		if err := opts.Validate(); err == nil {
			t.Error("Expected validation error for zero timeout")
		}
	})
	
	t.Run("FluentAPI", func(t *testing.T) {
		opts := alignment.DefaultAlignmentOptions().
			WithCaseSensitive(true).
			WithIgnoreWhitespace(false).
			WithMaxDistance(10).
			WithMinConfidence(0.8).
			WithTimeout(10000)
		
		if !opts.CaseSensitive {
			t.Error("Expected case sensitive to be true")
		}
		if opts.IgnoreWhitespace {
			t.Error("Expected ignore whitespace to be false")
		}
		if opts.MaxDistance != 10 {
			t.Errorf("Expected max distance 10, got %d", opts.MaxDistance)
		}
		if opts.MinConfidence != 0.8 {
			t.Errorf("Expected min confidence 0.8, got %f", opts.MinConfidence)
		}
		if opts.TimeoutMs != 10000 {
			t.Errorf("Expected timeout 10000, got %d", opts.TimeoutMs)
		}
	})
}

// TestAlignmentResult tests the AlignmentResult methods
func TestAlignmentResult(t *testing.T) {
	t.Run("NewAlignmentResult", func(t *testing.T) {
		interval := &types.CharInterval{StartPos: 0, EndPos: 10}
		alignmentInfo, _ := types.NewAlignmentResult(types.AlignmentExact, 0.95, 0.95, "TestMethod")
		
		result, err := alignment.NewAlignmentResult(interval, alignmentInfo, "extracted", "aligned", 0.95)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		if result == nil {
			t.Fatal("Expected non-nil result")
		}
		
		if !result.IsValid() {
			t.Error("Expected result to be valid")
		}
		
		if result.GetScore() != 0.95 {
			t.Errorf("Expected score 0.95, got %f", result.GetScore())
		}
	})
	
	t.Run("NewAlignmentResult_InvalidInputs", func(t *testing.T) {
		// Nil interval
		_, err := alignment.NewAlignmentResult(nil, nil, "extracted", "aligned", 0.95)
		if err == nil {
			t.Error("Expected error for nil interval")
		}
		
		// Invalid confidence
		interval := &types.CharInterval{StartPos: 0, EndPos: 10}
		_, err = alignment.NewAlignmentResult(interval, nil, "extracted", "aligned", 1.5)
		if err == nil {
			t.Error("Expected error for invalid confidence")
		}
	})
	
	t.Run("String", func(t *testing.T) {
		interval := &types.CharInterval{StartPos: 0, EndPos: 10}
		alignmentInfo, _ := types.NewAlignmentResult(types.AlignmentExact, 0.95, 0.95, "TestMethod")
		result, _ := alignment.NewAlignmentResult(interval, alignmentInfo, "extracted", "aligned", 0.95)
		
		str := result.String()
		if !strings.Contains(str, "[0:10)") {
			t.Errorf("Expected string to contain position, got %s", str)
		}
		if !strings.Contains(str, "0.95") {
			t.Errorf("Expected string to contain confidence, got %s", str)
		}
	})
}

// BenchmarkAlignment benchmarks different alignment algorithms
func BenchmarkAlignment(b *testing.B) {
	extracted := "test phrase"
	source := "This is a long document with many words and phrases. It contains the test phrase somewhere in the middle. There are many other words and sentences that make this a realistic test case for alignment algorithms."
	opts := alignment.DefaultAlignmentOptions()
	ctx := context.Background()
	
	b.Run("ExactMatcher", func(b *testing.B) {
		matcher := alignment.NewExactMatcher()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, err := matcher.AlignExtraction(ctx, extracted, source, opts)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	
	b.Run("FuzzyMatcher", func(b *testing.B) {
		matcher := alignment.NewFuzzyMatcher()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, err := matcher.AlignExtraction(ctx, extracted, source, opts)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
	
	b.Run("MultiAligner", func(b *testing.B) {
		aligner := alignment.NewMultiAligner()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, err := aligner.AlignWithBestMethod(ctx, extracted, source, opts)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// TestAlignmentConcurrency tests concurrent alignment operations
func TestAlignmentConcurrency(t *testing.T) {
	matcher := alignment.NewExactMatcher()
	extracted := "test phrase"
	source := "This is a test phrase in the document for concurrent testing."
	opts := alignment.DefaultAlignmentOptions()
	
	// Run multiple alignment operations concurrently
	const numGoroutines = 10
	results := make(chan error, numGoroutines)
	
	for i := 0; i < numGoroutines; i++ {
		go func() {
			_, _, err := matcher.AlignExtraction(context.Background(), extracted, source, opts)
			results <- err
		}()
	}
	
	// Check all results
	for i := 0; i < numGoroutines; i++ {
		select {
		case err := <-results:
			if err != nil {
				t.Errorf("Concurrent alignment failed: %v", err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("Concurrent alignment timed out")
		}
	}
}