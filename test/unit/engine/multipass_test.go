package engine_test

import (
	"context"
	"testing"
	"time"

	"github.com/sehwan505/langextract-go/internal/alignment"
	"github.com/sehwan505/langextract-go/internal/chunking"
	"github.com/sehwan505/langextract-go/internal/engine"
	"github.com/sehwan505/langextract-go/pkg/document"
	"github.com/sehwan505/langextract-go/pkg/extraction"
	"github.com/sehwan505/langextract-go/pkg/types"
)

// TestMultiPassCoordinator tests the MultiPassCoordinator implementation
func TestMultiPassCoordinator(t *testing.T) {
	t.Run("NewMultiPassCoordinator", func(t *testing.T) {
		config := engine.DefaultMultiPassConfig()
		coordinator := engine.NewMultiPassCoordinator(config)
		
		if coordinator == nil {
			t.Fatal("Expected non-nil coordinator")
		}
		
		// Test with nil config (should use defaults)
		coordinatorDefault := engine.NewMultiPassCoordinator(nil)
		if coordinatorDefault == nil {
			t.Fatal("Expected non-nil coordinator with default config")
		}
	})
	
	t.Run("DefaultMultiPassConfig", func(t *testing.T) {
		config := engine.DefaultMultiPassConfig()
		
		if config.MaxPasses <= 0 {
			t.Error("Expected positive max passes")
		}
		if config.MinPasses <= 0 {
			t.Error("Expected positive min passes")
		}
		if config.ImprovementThreshold < 0 || config.ImprovementThreshold > 1 {
			t.Error("Expected improvement threshold between 0 and 1")
		}
		if config.ConfidenceThreshold < 0 || config.ConfidenceThreshold > 1 {
			t.Error("Expected confidence threshold between 0 and 1")
		}
		if config.QualityThreshold < 0 || config.QualityThreshold > 1 {
			t.Error("Expected quality threshold between 0 and 1")
		}
		if config.ConcurrentChunks <= 0 {
			t.Error("Expected positive concurrent chunks")
		}
		// Remove timeout check as it's not in MultiPassConfig
	})
	
	t.Run("ConfigValidation", func(t *testing.T) {
		// Test various config scenarios
		configs := []*engine.MultiPassConfig{
			// Valid config
			{
				MaxPasses:          3,
				MinPasses:          1,
				ImprovementThreshold: 0.1,
				ConfidenceThreshold:  0.7,
				EnableChunking:      true,
				ChunkingOptions:     chunking.DefaultChunkingOptions(),
				EnableAlignment:     true,
				AlignmentOptions:    alignment.DefaultAlignmentOptions(),
				QualityThreshold:    0.6,
				ConcurrentChunks:    2,
				EnableCaching:       true,
				PassStrategy:        engine.AdaptivePasses,
				MergingStrategy:     engine.OverlapResolution,
			},
			// Chunking disabled
			{
				MaxPasses:       2,
				MinPasses:       1,
				EnableChunking:  false,
				EnableAlignment: false,
				ConcurrentChunks: 1,
				PassStrategy:    engine.FixedPasses,
				MergingStrategy: engine.UnionMerge,
			},
		}
		
		for i, config := range configs {
			coordinator := engine.NewMultiPassCoordinator(config)
			if coordinator == nil {
				t.Errorf("Config %d: Expected non-nil coordinator", i)
			}
		}
	})
}

// TestMultiPassConfigStrategies tests different multi-pass strategies
func TestMultiPassConfigStrategies(t *testing.T) {
	t.Run("ChunkOverlapStrategies", func(t *testing.T) {
		strategies := []engine.ChunkOverlapStrategy{
			engine.NoOverlap,
			engine.FixedOverlap,
			engine.AdaptiveOverlap,
			engine.SemanticOverlap,
		}
		
		for _, strategy := range strategies {
			config := engine.DefaultMultiPassConfig()
			config.OverlapStrategy = strategy
			
			_ = engine.NewMultiPassCoordinator(config)
		}
	})
	
	t.Run("PassStrategies", func(t *testing.T) {
		strategies := []engine.PassStrategy{
			engine.FixedPasses,
			engine.AdaptivePasses,
			engine.QualityDriven,
			engine.CoverageDriven,
		}
		
		for _, strategy := range strategies {
			config := engine.DefaultMultiPassConfig()
			config.PassStrategy = strategy
			
			_ = engine.NewMultiPassCoordinator(config)
		}
	})
	
	t.Run("MergingStrategies", func(t *testing.T) {
		strategies := []engine.MergingStrategy{
			engine.UnionMerge,
			engine.HighestConfidence,
			engine.VotingMerge,
			engine.OverlapResolution,
		}
		
		for _, strategy := range strategies {
			config := engine.DefaultMultiPassConfig()
			config.MergingStrategy = strategy
			
			_ = engine.NewMultiPassCoordinator(config)
		}
	})
}

// TestMultiPassMetrics tests the metrics tracking functionality
func TestMultiPassMetrics(t *testing.T) {
	t.Run("PassMetrics", func(t *testing.T) {
		metrics := engine.PassMetrics{
			PassNumber:        1,
			ChunksProcessed:   5,
			ExtractionsFound:  10,
			AverageConfidence: 0.85,
			ProcessingTime:    time.Second,
			ImprovementScore:  0.15,
			ErrorCount:        0,
		}
		
		if metrics.PassNumber != 1 {
			t.Errorf("Expected pass number 1, got %d", metrics.PassNumber)
		}
		if metrics.AverageConfidence != 0.85 {
			t.Errorf("Expected confidence 0.85, got %f", metrics.AverageConfidence)
		}
	})
	
	t.Run("ChunkMetrics", func(t *testing.T) {
		metrics := engine.ChunkMetrics{
			ChunkID:          "chunk_1",
			ChunkSize:        500,
			ExtractionsFound: 3,
			ProcessingTime:   100 * time.Millisecond,
			AlignmentSuccess: true,
			QualityScore:     0.9,
		}
		
		if metrics.ChunkID != "chunk_1" {
			t.Errorf("Expected chunk ID 'chunk_1', got %s", metrics.ChunkID)
		}
		if !metrics.AlignmentSuccess {
			t.Error("Expected alignment success to be true")
		}
	})
	
	t.Run("AlignmentMetrics", func(t *testing.T) {
		metrics := engine.AlignmentMetrics{
			ExtractedText:   "test phrase",
			AlignmentStatus: types.AlignmentExact,
			Confidence:      0.95,
			ProcessingTime:  50 * time.Millisecond,
			Method:          "ExactMatcher",
		}
		
		if metrics.ExtractedText != "test phrase" {
			t.Errorf("Expected extracted text 'test phrase', got %s", metrics.ExtractedText)
		}
		if metrics.Method != "ExactMatcher" {
			t.Errorf("Expected method 'ExactMatcher', got %s", metrics.Method)
		}
	})
	
	t.Run("MultiPassMetrics", func(t *testing.T) {
		metrics := engine.MultiPassMetrics{
			TotalPasses:         3,
			TotalChunks:         10,
			TotalExtractions:    25,
			TotalAlignments:     20,
			ProcessingTime:      5 * time.Second,
			OverallConfidence:   0.82,
			CoverageImprovement: 0.15,
			QualityScore:        0.88,
		}
		
		if metrics.TotalPasses != 3 {
			t.Errorf("Expected 3 total passes, got %d", metrics.TotalPasses)
		}
		if metrics.OverallConfidence != 0.82 {
			t.Errorf("Expected confidence 0.82, got %f", metrics.OverallConfidence)
		}
	})
}

// TestMultiPassResult tests the result data structures
func TestMultiPassResult(t *testing.T) {
	t.Run("MultiPassResult", func(t *testing.T) {
		startTime := time.Now()
		result := &engine.MultiPassResult{
			RequestID:        "test_request",
			StartTime:        startTime,
			ProcessingTime:   2 * time.Second,
			Passes:           make([]engine.PassResult, 0),
			AllExtractions:   make([]*extraction.Extraction, 0),
			FinalExtractions: make([]*extraction.Extraction, 0),
			TotalExtractions: 0,
			Success:          true,
			AlignmentErrors:  make([]string, 0),
		}
		
		if result.RequestID != "test_request" {
			t.Errorf("Expected request ID 'test_request', got %s", result.RequestID)
		}
		if !result.Success {
			t.Error("Expected success to be true")
		}
	})
	
	t.Run("PassResult", func(t *testing.T) {
		result := engine.PassResult{
			PassNumber:      2,
			ChunksProcessed: 8,
			Extractions:     make([]*extraction.Extraction, 0),
			ChunkResults:    make([]engine.ChunkResult, 0),
			ProcessingTime:  time.Second,
			Success:         true,
			Error:           nil,
		}
		
		if result.PassNumber != 2 {
			t.Errorf("Expected pass number 2, got %d", result.PassNumber)
		}
		if !result.Success {
			t.Error("Expected success to be true")
		}
	})
	
	t.Run("ChunkResult", func(t *testing.T) {
		result := engine.ChunkResult{
			ChunkID:        "chunk_test",
			Extractions:    make([]*extraction.Extraction, 0),
			ProcessingTime: 500 * time.Millisecond,
			TokensUsed:     150,
			Success:        true,
			CacheHit:       false,
			Error:          nil,
		}
		
		if result.ChunkID != "chunk_test" {
			t.Errorf("Expected chunk ID 'chunk_test', got %s", result.ChunkID)
		}
		if result.TokensUsed != 150 {
			t.Errorf("Expected 150 tokens used, got %d", result.TokensUsed)
		}
	})
}

// TestMultiPassIntegration tests integration scenarios
func TestMultiPassIntegration(t *testing.T) {
	t.Run("ChunkingIntegration", func(t *testing.T) {
		config := engine.DefaultMultiPassConfig()
		config.EnableChunking = true
		config.ChunkingOptions = chunking.DefaultChunkingOptions().WithMaxCharBuffer(100)
		
		coordinator := engine.NewMultiPassCoordinator(config)
		if coordinator == nil {
			t.Fatal("Expected non-nil coordinator")
		}
		
		// Test that chunking configuration is properly initialized
		// This is implicit testing since we can't access private fields
		// But we can verify the coordinator was created successfully
	})
	
	t.Run("AlignmentIntegration", func(t *testing.T) {
		config := engine.DefaultMultiPassConfig()
		config.EnableAlignment = true
		config.AlignmentOptions = alignment.DefaultAlignmentOptions().WithMaxDistance(5)
		config.AlignmentStrategy = alignment.StrategyBestScore
		
		coordinator := engine.NewMultiPassCoordinator(config)
		if coordinator == nil {
			t.Fatal("Expected non-nil coordinator")
		}
		
		// Test that alignment configuration is properly initialized
		// This is implicit testing since we can't access private fields
		// But we can verify the coordinator was created successfully
	})
	
	t.Run("CachingIntegration", func(t *testing.T) {
		config := engine.DefaultMultiPassConfig()
		config.EnableCaching = true
		config.CacheExpirationMinutes = 30
		
		coordinator := engine.NewMultiPassCoordinator(config)
		if coordinator == nil {
			t.Fatal("Expected non-nil coordinator")
		}
		
		// Test that caching configuration is properly initialized
		// This is implicit testing since we can't access private fields
		// But we can verify the coordinator was created successfully
	})
	
	t.Run("ConcurrencySettings", func(t *testing.T) {
		config := engine.DefaultMultiPassConfig()
		config.ConcurrentChunks = 5
		
		coordinator := engine.NewMultiPassCoordinator(config)
		if coordinator == nil {
			t.Fatal("Expected non-nil coordinator")
		}
		
		// Test that concurrency settings are properly handled
		// This is implicit testing since we can't access private fields
		// But we can verify the coordinator was created successfully
	})
}

// MockProviderManager for testing (simplified)
type MockProviderManager struct {
	responses map[string]*engine.CacheableResponse
}

func NewMockProviderManager() *MockProviderManager {
	return &MockProviderManager{
		responses: make(map[string]*engine.CacheableResponse),
	}
}

func (m *MockProviderManager) ExecuteWithFailover(ctx context.Context, request *engine.ExtractionRequest) (*engine.CacheableResponse, error) {
	// Return a mock response
	return &engine.CacheableResponse{
		Output:     `{"extractions": [{"text": "test extraction", "class": "test"}]}`,
		ProviderID: "mock_provider",
		ModelID:    "mock_model",
		TokensUsed: 100,
		Latency:    100 * time.Millisecond,
	}, nil
}

// TestMultiPassExecution tests the actual execution logic (simplified)
func TestMultiPassExecution(t *testing.T) {
	t.Run("BasicExecution", func(t *testing.T) {
		// This test is more conceptual since ExecuteMultiPass requires
		// a complex setup with provider managers and real extraction logic
		config := engine.DefaultMultiPassConfig()
		config.MaxPasses = 2
		config.EnableChunking = false // Simplify for testing
		config.EnableAlignment = false // Simplify for testing
		
		coordinator := engine.NewMultiPassCoordinator(config)
		if coordinator == nil {
			t.Fatal("Expected non-nil coordinator")
		}
		
		// Create a simple test request
		doc := document.NewDocument("This is a test document for multi-pass extraction.")
		request := engine.NewExtractionRequest(doc, "Extract test phrases")
		
		// Note: We can't actually test ExecuteMultiPass without a real provider manager
		// This test serves as a structural validation
		if request.Document == nil {
			t.Error("Expected non-nil document in request")
		}
		if request.TaskDescription == "" {
			t.Error("Expected non-empty task description")
		}
	})
}

// BenchmarkMultiPassCoordinator benchmarks the coordinator operations
func BenchmarkMultiPassCoordinator(b *testing.B) {
	config := engine.DefaultMultiPassConfig()
	
	b.Run("NewMultiPassCoordinator", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			coordinator := engine.NewMultiPassCoordinator(config)
			if coordinator == nil {
				b.Fatal("Expected non-nil coordinator")
			}
		}
	})
	
	b.Run("ConfigCreation", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			config := engine.DefaultMultiPassConfig()
			if config == nil {
				b.Fatal("Expected non-nil config")
			}
		}
	})
}

// TestMultiPassConcurrency tests concurrent access to the coordinator
func TestMultiPassConcurrency(t *testing.T) {
	_ = engine.NewMultiPassCoordinator(engine.DefaultMultiPassConfig())
	
	// Test concurrent creation of multiple coordinators
	const numGoroutines = 10
	results := make(chan *engine.MultiPassCoordinator, numGoroutines)
	
	for i := 0; i < numGoroutines; i++ {
		go func() {
			config := engine.DefaultMultiPassConfig()
			coord := engine.NewMultiPassCoordinator(config)
			results <- coord
		}()
	}
	
	// Check all results
	for i := 0; i < numGoroutines; i++ {
		select {
		case coord := <-results:
			if coord == nil {
				t.Errorf("Goroutine %d returned nil coordinator", i)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("Concurrent coordinator creation timed out")
		}
	}
}