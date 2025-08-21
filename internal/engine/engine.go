package engine

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/sehwan505/langextract-go/pkg/document"
	"github.com/sehwan505/langextract-go/pkg/extraction"
)

// ExtractionEngine is the main processing pipeline for document extraction.
// It coordinates document preprocessing, provider management, multi-pass extraction,
// and result aggregation.
type ExtractionEngine struct {
	providerManager *ProviderManager
	config          *ExtractionEngineConfig
	activeRequests  map[string]*ExtractionRequest
	requestMutex    sync.RWMutex
}

// ExtractionEngineConfig configures the extraction engine behavior.
type ExtractionEngineConfig struct {
	// Processing settings
	MaxConcurrentRequests int
	DefaultTimeout        time.Duration
	EnableDebugMode       bool

	// Multi-pass extraction
	EnableMultiPass       bool
	MaxPasses            int
	PassImprovementThreshold float64

	// Result aggregation
	EnableDeduplication   bool
	ConfidenceThreshold   float64
	OverlapResolution     OverlapResolutionStrategy

	// Progress tracking
	EnableProgressTracking bool
	ProgressUpdateInterval time.Duration

	// Provider management
	ProviderConfig *ProviderManagerConfig
}

// OverlapResolutionStrategy defines how overlapping extractions are handled.
type OverlapResolutionStrategy string

const (
	KeepHighestConfidence OverlapResolutionStrategy = "highest_confidence"
	KeepLongest          OverlapResolutionStrategy = "longest"
	KeepFirst           OverlapResolutionStrategy = "first"
	MergeOverlapping    OverlapResolutionStrategy = "merge"
)

// NewExtractionEngine creates a new extraction engine with the given configuration.
func NewExtractionEngine(config *ExtractionEngineConfig) *ExtractionEngine {
	if config == nil {
		config = DefaultExtractionEngineConfig()
	}

	if config.ProviderConfig == nil {
		config.ProviderConfig = DefaultProviderManagerConfig()
	}

	engine := &ExtractionEngine{
		providerManager: NewProviderManager(config.ProviderConfig),
		config:          config,
		activeRequests:  make(map[string]*ExtractionRequest),
	}

	return engine
}

// DefaultExtractionEngineConfig returns a default configuration.
func DefaultExtractionEngineConfig() *ExtractionEngineConfig {
	return &ExtractionEngineConfig{
		MaxConcurrentRequests:    10,
		DefaultTimeout:           60 * time.Second,
		EnableDebugMode:          false,
		EnableMultiPass:          false,
		MaxPasses:               3,
		PassImprovementThreshold: 0.1,
		EnableDeduplication:      true,
		ConfidenceThreshold:      0.5,
		OverlapResolution:        KeepHighestConfidence,
		EnableProgressTracking:   true,
		ProgressUpdateInterval:   time.Second,
	}
}

// ProcessExtraction processes a single extraction request through the complete pipeline.
func (e *ExtractionEngine) ProcessExtraction(request *ExtractionRequest) (*ExtractionResponse, error) {
	// Initialize response
	response := NewExtractionResponse(request.ID)
	startTime := time.Now()

	// Add to active requests
	e.trackRequest(request)
	defer e.untrackRequest(request.ID)

	// Start progress tracking if enabled
	if e.config.EnableProgressTracking && request.ProgressCallback != nil {
		go e.trackProgress(request, response)
	}

	// Execute the extraction pipeline
	err := e.executePipeline(request, response)
	
	// Set execution metadata
	response.ExecutionTime = time.Since(startTime)
	if err != nil {
		response.SetError(err, "pipeline_error")
	}

	return response, err
}

// executePipeline executes the complete extraction pipeline.
func (e *ExtractionEngine) executePipeline(request *ExtractionRequest, response *ExtractionResponse) error {
	ctx := request.Context
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), request.Timeout)
		defer cancel()
	}

	// Stage 1: Initialization
	if err := e.stageInitialization(request, response); err != nil {
		return fmt.Errorf("initialization failed: %w", err)
	}

	// Stage 2: Document preprocessing
	if err := e.stagePreprocessing(request, response); err != nil {
		return fmt.Errorf("preprocessing failed: %w", err)
	}

	// Stage 3: Multi-pass extraction
	if err := e.stageExtraction(ctx, request, response); err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}

	// Stage 4: Result aggregation and deduplication
	if err := e.stageAggregation(request, response); err != nil {
		return fmt.Errorf("aggregation failed: %w", err)
	}

	// Stage 5: Validation
	if err := e.stageValidation(request, response); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Stage 6: Finalization
	if err := e.stageFinalization(request, response); err != nil {
		return fmt.Errorf("finalization failed: %w", err)
	}

	return nil
}

// stageInitialization performs initial setup and validation.
func (e *ExtractionEngine) stageInitialization(request *ExtractionRequest, response *ExtractionResponse) error {
	stepStart := time.Now()
	
	// Validate request
	if request.Document == nil && request.Text == "" {
		return fmt.Errorf("no input document or text provided")
	}

	if request.TaskDescription == "" {
		return fmt.Errorf("task description is required")
	}

	// Set defaults if needed
	if request.RetryCount < 0 {
		request.RetryCount = 2
	}
	if request.ExtractionPasses < 1 {
		request.ExtractionPasses = 1
	}

	// Log processing step
	response.AddProcessingStep(
		"initialization", 
		"success", 
		"Request initialized and validated", 
		time.Since(stepStart), 
		map[string]any{
			"text_length": len(request.Text),
			"has_examples": len(request.Examples) > 0,
			"has_schema": request.Schema != nil,
		},
	)

	return nil
}

// stagePreprocessing performs document preprocessing and validation.
func (e *ExtractionEngine) stagePreprocessing(request *ExtractionRequest, response *ExtractionResponse) error {
	stepStart := time.Now()

	// Ensure we have a Document object
	if request.Document == nil && request.Text != "" {
		request.Document = document.NewDocument(request.Text)
	}

	// Validate document
	if request.Document == nil {
		return fmt.Errorf("failed to create document from input")
	}

	// Update request text if needed
	if request.Text == "" {
		request.Text = request.Document.Text
	}

	// Basic text preprocessing
	text := strings.TrimSpace(request.Text)
	if text == "" {
		return fmt.Errorf("document text is empty after preprocessing")
	}
	request.Text = text

	response.AddProcessingStep(
		"preprocessing", 
		"success", 
		"Document preprocessed and validated", 
		time.Since(stepStart),
		map[string]any{
			"document_id": request.Document.DocumentID(),
			"text_length": len(request.Text),
			"token_count": request.Document.TokenCount(),
		},
	)

	return nil
}

// stageExtraction performs the main extraction using providers.
func (e *ExtractionEngine) stageExtraction(ctx context.Context, request *ExtractionRequest, response *ExtractionResponse) error {
	stepStart := time.Now()

	var allExtractions []*extraction.Extraction
	passCount := request.ExtractionPasses

	// If multi-pass is enabled and not explicitly set, determine optimal passes
	if e.config.EnableMultiPass && request.ExtractionPasses == 1 {
		passCount = e.determineOptimalPasses(request)
	}

	// Execute extraction passes
	for pass := 1; pass <= passCount; pass++ {
		passExtractions, err := e.executeExtractionPass(ctx, request, response, pass, passCount)
		if err != nil {
			if pass == 1 {
				// First pass failure is critical
				return fmt.Errorf("extraction pass %d failed: %w", pass, err)
			}
			// Later pass failures are logged but not critical
			response.AddProcessingStep(
				fmt.Sprintf("extraction_pass_%d", pass),
				"error",
				fmt.Sprintf("Pass %d failed: %v", pass, err),
				time.Since(stepStart),
				map[string]any{"pass": pass, "error": err.Error()},
			)
			break
		}

		allExtractions = append(allExtractions, passExtractions...)
		response.PassesCompleted = pass

		// Check if additional passes are beneficial
		if e.config.EnableMultiPass && pass > 1 {
			improvement := e.calculatePassImprovement(allExtractions, passExtractions)
			if improvement < e.config.PassImprovementThreshold {
				break
			}
		}
	}

	// Store all extractions
	response.Extractions = allExtractions
	response.ExtractionCount = len(allExtractions)

	response.AddProcessingStep(
		"extraction", 
		"success", 
		fmt.Sprintf("Completed %d extraction passes with %d extractions", response.PassesCompleted, len(allExtractions)), 
		time.Since(stepStart),
		map[string]any{
			"passes_completed": response.PassesCompleted,
			"extraction_count": len(allExtractions),
		},
	)

	return nil
}

// executeExtractionPass executes a single extraction pass.
func (e *ExtractionEngine) executeExtractionPass(ctx context.Context, request *ExtractionRequest, response *ExtractionResponse, passNum, totalPasses int) ([]*extraction.Extraction, error) {
	// Update progress
	if request.ProgressCallback != nil {
		progress := ExtractionProgress{
			RequestID:   request.ID,
			Stage:       string(StageProviderCall),
			Progress:    float64(passNum-1) / float64(totalPasses),
			Message:     fmt.Sprintf("Executing extraction pass %d of %d", passNum, totalPasses),
			CurrentPass: passNum,
			TotalPasses: totalPasses,
		}
		request.ProgressCallback(progress)
	}

	// Execute request with provider manager (includes failover)
	cachedResponse, err := e.providerManager.ExecuteWithFailover(ctx, request)
	if err != nil {
		return nil, err
	}

	// Update response metadata
	response.ProviderUsed = cachedResponse.ProviderID
	response.ModelUsed = cachedResponse.ModelID
	response.TokensUsed += cachedResponse.TokensUsed

	// Parse extractions from response
	extractions, err := e.parseExtractions(cachedResponse.Output, request.Text)
	if err != nil {
		return nil, fmt.Errorf("failed to parse extractions: %w", err)
	}

	return extractions, nil
}

// stageAggregation aggregates and deduplicates results from multiple passes.
func (e *ExtractionEngine) stageAggregation(request *ExtractionRequest, response *ExtractionResponse) error {
	stepStart := time.Now()

	if !e.config.EnableDeduplication || len(response.Extractions) == 0 {
		response.AddProcessingStep(
			"aggregation", 
			"skipped", 
			"Deduplication disabled or no extractions", 
			time.Since(stepStart),
			nil,
		)
		return nil
	}

	originalCount := len(response.Extractions)
	
	// Remove duplicates and resolve overlaps
	deduplicatedExtractions := e.deduplicateExtractions(response.Extractions)
	resolvedExtractions := e.resolveOverlaps(deduplicatedExtractions)
	
	// Filter by confidence threshold
	filteredExtractions := e.filterByConfidence(resolvedExtractions, e.config.ConfidenceThreshold)

	response.Extractions = filteredExtractions
	response.ExtractionCount = len(filteredExtractions)

	response.AddProcessingStep(
		"aggregation", 
		"success", 
		fmt.Sprintf("Aggregated extractions: %d -> %d", originalCount, len(filteredExtractions)), 
		time.Since(stepStart),
		map[string]any{
			"original_count": originalCount,
			"final_count": len(filteredExtractions),
			"duplicates_removed": originalCount - len(deduplicatedExtractions),
			"overlaps_resolved": len(deduplicatedExtractions) - len(resolvedExtractions),
			"low_confidence_filtered": len(resolvedExtractions) - len(filteredExtractions),
		},
	)

	return nil
}

// stageValidation validates extractions against schema if provided.
func (e *ExtractionEngine) stageValidation(request *ExtractionRequest, response *ExtractionResponse) error {
	stepStart := time.Now()

	if !request.ValidateOutput || request.Schema == nil {
		response.AddProcessingStep(
			"validation", 
			"skipped", 
			"Schema validation disabled or no schema provided", 
			time.Since(stepStart),
			nil,
		)
		return nil
	}

	validExtractions := make([]*extraction.Extraction, 0, len(response.Extractions))
	validationErrors := make([]ValidationError, 0)

	for _, ext := range response.Extractions {
		if err := request.Schema.ValidateExtraction(ext); err != nil {
			validationErrors = append(validationErrors, ValidationError{
				Field:      "extraction",
				Value:      ext.ExtractionText,
				Constraint: "schema_validation",
				Message:    err.Error(),
			})
		} else {
			validExtractions = append(validExtractions, ext)
		}
	}

	response.Extractions = validExtractions
	response.ExtractionCount = len(validExtractions)
	response.ValidationErrors = validationErrors

	status := "success"
	if len(validationErrors) > 0 {
		status = "warning"
	}

	response.AddProcessingStep(
		"validation", 
		status, 
		fmt.Sprintf("Schema validation completed: %d valid, %d invalid", len(validExtractions), len(validationErrors)), 
		time.Since(stepStart),
		map[string]any{
			"valid_count": len(validExtractions),
			"invalid_count": len(validationErrors),
		},
	)

	return nil
}

// stageFinalization creates the final annotated document and calculates metrics.
func (e *ExtractionEngine) stageFinalization(request *ExtractionRequest, response *ExtractionResponse) error {
	stepStart := time.Now()

	// Create annotated document
	annotatedDoc := document.NewAnnotatedDocument(request.Document)
	annotatedDoc.AddExtractions(response.Extractions)

	response.AnnotatedDocument = annotatedDoc
	response.TextCoverage = annotatedDoc.GetCoverage()

	// Calculate confidence score
	if len(response.Extractions) > 0 {
		totalConfidence := 0.0
		count := 0
		for _, ext := range response.Extractions {
			if conf, ok := ext.GetConfidence(); ok {
				totalConfidence += conf
				count++
			}
		}
		if count > 0 {
			response.ConfidenceScore = totalConfidence / float64(count)
		}
	}

	response.AddProcessingStep(
		"finalization", 
		"success", 
		"Extraction pipeline completed successfully", 
		time.Since(stepStart),
		map[string]any{
			"text_coverage": response.TextCoverage,
			"confidence_score": response.ConfidenceScore,
		},
	)

	return nil
}

// Helper methods

func (e *ExtractionEngine) trackRequest(request *ExtractionRequest) {
	e.requestMutex.Lock()
	defer e.requestMutex.Unlock()
	e.activeRequests[request.ID] = request
}

func (e *ExtractionEngine) untrackRequest(requestID string) {
	e.requestMutex.Lock()
	defer e.requestMutex.Unlock()
	delete(e.activeRequests, requestID)
}

func (e *ExtractionEngine) trackProgress(request *ExtractionRequest, response *ExtractionResponse) {
	ticker := time.NewTicker(e.config.ProgressUpdateInterval)
	defer ticker.Stop()

	startTime := time.Now()
	for range ticker.C {
		// Check if request is still active
		e.requestMutex.RLock()
		_, active := e.activeRequests[request.ID]
		e.requestMutex.RUnlock()

		if !active {
			break
		}

		progress := ExtractionProgress{
			RequestID:   request.ID,
			Stage:       string(StageProviderCall),
			ElapsedTime: time.Since(startTime),
			Message:     "Processing extraction...",
		}

		request.ProgressCallback(progress)
	}
}

func (e *ExtractionEngine) determineOptimalPasses(request *ExtractionRequest) int {
	// Simple heuristic: more passes for longer documents
	textLength := len(request.Text)
	if textLength > 10000 {
		return 3
	} else if textLength > 5000 {
		return 2
	}
	return 1
}

func (e *ExtractionEngine) calculatePassImprovement(allExtractions, newExtractions []*extraction.Extraction) float64 {
	if len(allExtractions) == 0 {
		return 1.0
	}
	return float64(len(newExtractions)) / float64(len(allExtractions))
}

func (e *ExtractionEngine) parseExtractions(output, sourceText string) ([]*extraction.Extraction, error) {
	// This is a simplified implementation - would need proper JSON parsing
	// For now, return empty list if we can't parse
	return make([]*extraction.Extraction, 0), nil
}

func (e *ExtractionEngine) deduplicateExtractions(extractions []*extraction.Extraction) []*extraction.Extraction {
	seen := make(map[string]*extraction.Extraction)
	result := make([]*extraction.Extraction, 0)

	for _, ext := range extractions {
		key := fmt.Sprintf("%s:%s", ext.ExtractionClass, ext.ExtractionText)
		if existing, exists := seen[key]; exists {
			// Keep the one with higher confidence
			if extConf, ok := ext.GetConfidence(); ok {
				if existingConf, ok := existing.GetConfidence(); ok {
					if extConf > existingConf {
						seen[key] = ext
					}
				}
			}
		} else {
			seen[key] = ext
			result = append(result, ext)
		}
	}

	return result
}

func (e *ExtractionEngine) resolveOverlaps(extractions []*extraction.Extraction) []*extraction.Extraction {
	// Simple implementation - would need proper overlap detection
	return extractions
}

func (e *ExtractionEngine) filterByConfidence(extractions []*extraction.Extraction, threshold float64) []*extraction.Extraction {
	result := make([]*extraction.Extraction, 0)

	for _, ext := range extractions {
		if conf, ok := ext.GetConfidence(); ok {
			if conf >= threshold {
				result = append(result, ext)
			}
		} else {
			// Include extractions without confidence scores
			result = append(result, ext)
		}
	}

	return result
}

// GetActiveRequests returns information about currently active requests.
func (e *ExtractionEngine) GetActiveRequests() map[string]*ExtractionRequest {
	e.requestMutex.RLock()
	defer e.requestMutex.RUnlock()

	result := make(map[string]*ExtractionRequest)
	for id, request := range e.activeRequests {
		result[id] = request
	}
	return result
}

// GetProviderHealth returns the health status of all providers.
func (e *ExtractionEngine) GetProviderHealth() map[string]*ProviderHealth {
	return e.providerManager.GetProviderHealth()
}

// GetCacheStats returns cache statistics.
func (e *ExtractionEngine) GetCacheStats() map[string]interface{} {
	return e.providerManager.GetCacheStats()
}