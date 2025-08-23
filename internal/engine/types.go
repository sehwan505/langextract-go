package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/sehwan505/langextract-go/pkg/document"
	"github.com/sehwan505/langextract-go/pkg/extraction"
	"github.com/sehwan505/langextract-go/pkg/providers"
)

// ExtractionRequest represents an internal extraction request.
// This encapsulates all information needed for a single extraction operation.
type ExtractionRequest struct {
	// Request identification
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`

	// Input data
	Document *document.Document `json:"document"`
	Text     string             `json:"text"`

	// Extraction configuration
	TaskDescription string                    `json:"task_description"`
	Examples        []*extraction.ExampleData `json:"examples"`
	Schema          extraction.ExtractionSchema `json:"schema,omitempty"`

	// Provider configuration
	ProviderID     string                  `json:"provider_id"`
	ModelID        string                  `json:"model_id"`
	ModelConfig    *providers.ModelConfig  `json:"model_config,omitempty"`
	Provider       providers.BaseLanguageModel `json:"-"` // Not serialized

	// Processing options
	MaxTokens        int           `json:"max_tokens,omitempty"`
	Temperature      float64       `json:"temperature"`
	Timeout          time.Duration `json:"timeout"`
	RetryCount       int           `json:"retry_count"`
	ValidateOutput   bool          `json:"validate_output"`
	ExtractionPasses int           `json:"extraction_passes"`

	// Context and cancellation
	Context context.Context `json:"-"` // Not serialized

	// Progress tracking
	ProgressCallback func(progress ExtractionProgress) `json:"-"` // Not serialized
}

// ExtractionResponse represents the result of an extraction operation.
type ExtractionResponse struct {
	// Request identification
	RequestID string    `json:"request_id"`
	Timestamp time.Time `json:"timestamp"`

	// Results
	AnnotatedDocument *document.AnnotatedDocument `json:"annotated_document,omitempty"`
	Extractions       []*extraction.Extraction    `json:"extractions"`

	// Execution metadata
	ExecutionTime    time.Duration `json:"execution_time"`
	TokensUsed       int           `json:"tokens_used,omitempty"`
	ProviderUsed     string        `json:"provider_used"`
	ModelUsed        string        `json:"model_used"`
	PassesCompleted  int           `json:"passes_completed"`
	
	// Quality metrics
	ExtractionCount  int     `json:"extraction_count"`
	TextCoverage     float64 `json:"text_coverage"`
	ConfidenceScore  float64 `json:"confidence_score,omitempty"`

	// Error information
	Error           error                   `json:"-"` // Not serialized
	ErrorMessage    string                  `json:"error_message,omitempty"`
	ErrorCode       string                  `json:"error_code,omitempty"`
	ValidationErrors []ValidationError      `json:"validation_errors,omitempty"`

	// Debug information
	DebugInfo *ExtractionDebugInfo `json:"debug_info,omitempty"`
}

// ExtractionProgress represents progress information for long-running extractions.
type ExtractionProgress struct {
	RequestID       string        `json:"request_id"`
	Stage           string        `json:"stage"`
	Progress        float64       `json:"progress"` // 0.0 to 1.0
	Message         string        `json:"message"`
	ElapsedTime     time.Duration `json:"elapsed_time"`
	EstimatedTotal  time.Duration `json:"estimated_total,omitempty"`
	CurrentPass     int           `json:"current_pass"`
	TotalPasses     int           `json:"total_passes"`
	CurrentChunk    int           `json:"current_chunk"`
	ChunksProcessed int           `json:"chunks_processed"`
	TotalChunks     int           `json:"total_chunks"`
}

// ExtractionDebugInfo contains detailed debug information about the extraction process.
type ExtractionDebugInfo struct {
	// Prompt information
	GeneratedPrompt string            `json:"generated_prompt,omitempty"`
	PromptTokens    int               `json:"prompt_tokens,omitempty"`
	
	// Provider responses
	RawResponses    []string          `json:"raw_responses,omitempty"`
	ResponseTokens  int               `json:"response_tokens,omitempty"`
	
	// Processing details
	ProcessingSteps []ProcessingStep  `json:"processing_steps,omitempty"`
	RetryAttempts   int               `json:"retry_attempts"`
	FailoverEvents  []FailoverEvent   `json:"failover_events,omitempty"`
	
	// Performance metrics
	ProviderLatency time.Duration     `json:"provider_latency"`
	ProcessingTime  time.Duration     `json:"processing_time"`
	
	// Schema validation details
	SchemaValidation *SchemaValidationResult `json:"schema_validation,omitempty"`
}

// ProcessingStep represents a single step in the extraction pipeline.
type ProcessingStep struct {
	Name        string        `json:"name"`
	StartTime   time.Time     `json:"start_time"`
	Duration    time.Duration `json:"duration"`
	Status      string        `json:"status"` // "success", "error", "skipped"
	Message     string        `json:"message,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// FailoverEvent represents a provider failover event.
type FailoverEvent struct {
	Timestamp       time.Time `json:"timestamp"`
	OriginalProvider string    `json:"original_provider"`
	FailureReason   string    `json:"failure_reason"`
	FallbackProvider string    `json:"fallback_provider"`
	Success         bool      `json:"success"`
}

// ValidationError represents a schema validation error.
type ValidationError struct {
	Field       string `json:"field"`
	Value       string `json:"value"`
	Constraint  string `json:"constraint"`
	Message     string `json:"message"`
}

// SchemaValidationResult contains the results of schema validation.
type SchemaValidationResult struct {
	Valid              bool              `json:"valid"`
	ErrorCount         int               `json:"error_count"`
	WarningCount       int               `json:"warning_count"`
	ValidationErrors   []ValidationError `json:"validation_errors,omitempty"`
	ValidationWarnings []ValidationError `json:"validation_warnings,omitempty"`
	ValidationTime     time.Duration     `json:"validation_time"`
}

// ExtractionStage represents the current stage of extraction processing.
type ExtractionStage string

const (
	StageInitialization  ExtractionStage = "initialization"
	StagePreprocessing   ExtractionStage = "preprocessing"
	StagePromptBuilding  ExtractionStage = "prompt_building"
	StageProviderCall    ExtractionStage = "provider_call"
	StageChunkProcessing ExtractionStage = "chunk_processing"
	StageResponseParsing ExtractionStage = "response_parsing"
	StageValidation      ExtractionStage = "validation"
	StageAlignment       ExtractionStage = "alignment"
	StageAggregation     ExtractionStage = "aggregation"
	StageFinalization    ExtractionStage = "finalization"
	StageComplete        ExtractionStage = "complete"
	StageError           ExtractionStage = "error"
)

// ExtractionStatus represents the status of an extraction operation.
type ExtractionStatus string

const (
	StatusPending    ExtractionStatus = "pending"
	StatusProcessing ExtractionStatus = "processing"
	StatusComplete   ExtractionStatus = "complete"
	StatusFailed     ExtractionStatus = "failed"
	StatusCancelled  ExtractionStatus = "cancelled"
)

// NewExtractionRequest creates a new extraction request with default values.
func NewExtractionRequest(doc *document.Document, taskDescription string) *ExtractionRequest {
	return &ExtractionRequest{
		ID:               generateRequestID(),
		Timestamp:        time.Now(),
		Document:         doc,
		Text:             doc.Text,
		TaskDescription:  taskDescription,
		Examples:         make([]*extraction.ExampleData, 0),
		Temperature:      0.0,
		Timeout:          60 * time.Second,
		RetryCount:       2,
		ValidateOutput:   true,
		ExtractionPasses: 1,
		Context:          context.Background(),
	}
}

// NewExtractionResponse creates a new extraction response.
func NewExtractionResponse(requestID string) *ExtractionResponse {
	return &ExtractionResponse{
		RequestID:        requestID,
		Timestamp:        time.Now(),
		Extractions:      make([]*extraction.Extraction, 0),
		ValidationErrors: make([]ValidationError, 0),
	}
}

// IsSuccessful returns true if the extraction was successful.
func (r *ExtractionResponse) IsSuccessful() bool {
	return r.Error == nil && r.ErrorMessage == ""
}

// AddValidationError adds a validation error to the response.
func (r *ExtractionResponse) AddValidationError(field, value, constraint, message string) {
	r.ValidationErrors = append(r.ValidationErrors, ValidationError{
		Field:      field,
		Value:      value,
		Constraint: constraint,
		Message:    message,
	})
}

// SetError sets the error information for the response.
func (r *ExtractionResponse) SetError(err error, code string) {
	r.Error = err
	r.ErrorCode = code
	if err != nil {
		r.ErrorMessage = err.Error()
	}
}

// AddProcessingStep adds a processing step to the debug information.
func (r *ExtractionResponse) AddProcessingStep(name, status, message string, duration time.Duration, metadata map[string]any) {
	if r.DebugInfo == nil {
		r.DebugInfo = &ExtractionDebugInfo{
			ProcessingSteps: make([]ProcessingStep, 0),
		}
	}

	step := ProcessingStep{
		Name:      name,
		StartTime: time.Now().Add(-duration),
		Duration:  duration,
		Status:    status,
		Message:   message,
		Metadata:  metadata,
	}

	r.DebugInfo.ProcessingSteps = append(r.DebugInfo.ProcessingSteps, step)
}

// AddFailoverEvent adds a failover event to the debug information.
func (r *ExtractionResponse) AddFailoverEvent(originalProvider, failureReason, fallbackProvider string, success bool) {
	if r.DebugInfo == nil {
		r.DebugInfo = &ExtractionDebugInfo{
			FailoverEvents: make([]FailoverEvent, 0),
		}
	}

	event := FailoverEvent{
		Timestamp:        time.Now(),
		OriginalProvider: originalProvider,
		FailureReason:    failureReason,
		FallbackProvider: fallbackProvider,
		Success:          success,
	}

	r.DebugInfo.FailoverEvents = append(r.DebugInfo.FailoverEvents, event)
}

// generateRequestID generates a unique request ID.
func generateRequestID() string {
	// Use timestamp + random suffix for uniqueness
	return fmt.Sprintf("req_%d_%d", time.Now().UnixNano(), time.Now().Nanosecond()%10000)
}