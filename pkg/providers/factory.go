package providers

import (
	"fmt"
	"os"
)

// ProviderOptions holds options for creating providers.
type ProviderOptions struct {
	APIKey     string
	BaseURL    string
	Timeout    int
	MaxRetries int
}

// CreateOpenAI creates an OpenAI provider with the given model ID and options.
func CreateOpenAI(modelID string, opts *ProviderOptions) (BaseLanguageModel, error) {
	config := NewModelConfig(modelID).WithProvider("openai")
	
	if opts != nil {
		kwargs := make(map[string]interface{})
		if opts.APIKey != "" {
			kwargs["api_key"] = opts.APIKey
		}
		if opts.BaseURL != "" {
			kwargs["base_url"] = opts.BaseURL
		}
		config.WithProviderKwargs(kwargs)
	}
	
	return CreateModel(config)
}

// CreateGPT4 creates a GPT-4 provider instance.
func CreateGPT4(apiKey string) (BaseLanguageModel, error) {
	opts := &ProviderOptions{APIKey: apiKey}
	return CreateOpenAI("gpt-4", opts)
}

// CreateGPT35Turbo creates a GPT-3.5-turbo provider instance.
func CreateGPT35Turbo(apiKey string) (BaseLanguageModel, error) {
	opts := &ProviderOptions{APIKey: apiKey}
	return CreateOpenAI("gpt-3.5-turbo", opts)
}

// MustCreateModel creates a model and panics on error.
// Useful for initialization when the model must be available.
func MustCreateModel(config *ModelConfig) BaseLanguageModel {
	model, err := CreateModel(config)
	if err != nil {
		panic(fmt.Sprintf("failed to create model: %v", err))
	}
	return model
}

// CreateModelFromEnv creates a model using environment variables for configuration.
func CreateModelFromEnv(modelID string) (BaseLanguageModel, error) {
	// Detect provider from model ID
	var provider string
	switch {
	case isOpenAIModel(modelID):
		provider = "openai"
	default:
		return nil, fmt.Errorf("cannot determine provider for model ID: %s", modelID)
	}
	
	config := NewModelConfig(modelID).WithProvider(provider)
	
	// Add environment-based configuration
	kwargs := make(map[string]interface{})
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" && provider == "openai" {
		kwargs["api_key"] = apiKey
	}
	if baseURL := os.Getenv("OPENAI_BASE_URL"); baseURL != "" && provider == "openai" {
		kwargs["base_url"] = baseURL
	}
	
	if len(kwargs) > 0 {
		config.WithProviderKwargs(kwargs)
	}
	
	return CreateModel(config)
}

// isOpenAIModel checks if a model ID belongs to OpenAI.
func isOpenAIModel(modelID string) bool {
	openaiModels := []string{
		"gpt-4", "gpt-4-turbo", "gpt-3.5-turbo", 
		"gpt-4o", "gpt-4o-mini", "text-davinci-003",
	}
	
	for _, model := range openaiModels {
		if modelID == model {
			return true
		}
	}
	return false
}