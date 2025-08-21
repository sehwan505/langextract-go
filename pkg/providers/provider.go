package providers

import (
	"context"
)

// BaseLanguageModel defines the core interface that all language model providers must implement.
// This mirrors the BaseLanguageModel abstract class from the Python implementation.
type BaseLanguageModel interface {
	// Infer generates model output for the given prompts
	Infer(ctx context.Context, prompts []string, options map[string]any) ([][]ScoredOutput, error)
	
	// ParseOutput processes raw model output into structured format
	ParseOutput(output string) (any, error)
	
	// ApplySchema applies schema constraints to the model
	ApplySchema(schema any)
	
	// SetFenceOutput configures whether output should be fenced (e.g., ```json)
	SetFenceOutput(enabled bool)
	
	// GetModelID returns the model identifier
	GetModelID() string
	
	// IsAvailable checks if the provider is ready for use
	IsAvailable() bool
}

// ScoredOutput represents a single output with an optional score.
type ScoredOutput struct {
	Output string  `json:"output"`
	Score  float64 `json:"score,omitempty"`
}

// ModelConfig holds language model configuration.
// This mirrors the ModelConfig from the Python implementation.
type ModelConfig struct {
	ModelID         string                 `json:"model_id"`
	Provider        string                 `json:"provider,omitempty"`
	ProviderKwargs  map[string]any `json:"provider_kwargs,omitempty"`
	Temperature     float64                `json:"temperature,omitempty"`
	MaxTokens       int                    `json:"max_tokens,omitempty"`
	TopP            float64                `json:"top_p,omitempty"`
	FrequencyPenalty float64               `json:"frequency_penalty,omitempty"`
	PresencePenalty  float64               `json:"presence_penalty,omitempty"`
}

// NewModelConfig creates a new ModelConfig with defaults.
func NewModelConfig(modelID string) *ModelConfig {
	return &ModelConfig{
		ModelID:     modelID,
		Temperature: 0.0,
		MaxTokens:   1024,
		TopP:        1.0,
	}
}

// WithProvider sets the provider for the model config.
func (c *ModelConfig) WithProvider(provider string) *ModelConfig {
	c.Provider = provider
	return c
}

// WithProviderKwargs sets provider-specific arguments.
func (c *ModelConfig) WithProviderKwargs(kwargs map[string]any) *ModelConfig {
	c.ProviderKwargs = kwargs
	return c
}

// WithTemperature sets the temperature parameter.
func (c *ModelConfig) WithTemperature(temp float64) *ModelConfig {
	c.Temperature = temp
	return c
}

// WithMaxTokens sets the maximum tokens parameter.
func (c *ModelConfig) WithMaxTokens(maxTokens int) *ModelConfig {
	c.MaxTokens = maxTokens
	return c
}