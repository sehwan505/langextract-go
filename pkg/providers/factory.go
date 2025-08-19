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
		kwargs := make(map[string]any)
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

// CreateGemini creates a Gemini provider with the given model ID and options.
func CreateGemini(modelID string, opts *ProviderOptions) (BaseLanguageModel, error) {
	config := NewModelConfig(modelID).WithProvider("gemini")
	
	if opts != nil {
		kwargs := make(map[string]any)
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

// CreateGeminiPro creates a Gemini Pro provider instance.
func CreateGeminiPro(apiKey string) (BaseLanguageModel, error) {
	opts := &ProviderOptions{APIKey: apiKey}
	return CreateGemini("gemini-2.5-pro", opts)
}

// CreateGeminiFlash creates a Gemini Flash provider instance.
func CreateGeminiFlash(apiKey string) (BaseLanguageModel, error) {
	opts := &ProviderOptions{APIKey: apiKey}
	return CreateGemini("gemini-2.5-flash", opts)
}

// CreateOllama creates an Ollama provider with the given model ID and options.
func CreateOllama(modelID string, opts *ProviderOptions) (BaseLanguageModel, error) {
	config := NewModelConfig(modelID).WithProvider("ollama")
	
	if opts != nil {
		kwargs := make(map[string]any)
		if opts.BaseURL != "" {
			kwargs["base_url"] = opts.BaseURL
		}
		config.WithProviderKwargs(kwargs)
	}
	
	return CreateModel(config)
}

// CreateOllamaLlama creates an Ollama provider with Llama model.
func CreateOllamaLlama(baseURL string) (BaseLanguageModel, error) {
	opts := &ProviderOptions{BaseURL: baseURL}
	return CreateOllama("llama3.2", opts)
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
	case isGeminiModel(modelID):
		provider = "gemini"
	case isOllamaModel(modelID):
		provider = "ollama"
	default:
		return nil, fmt.Errorf("cannot determine provider for model ID: %s", modelID)
	}
	
	config := NewModelConfig(modelID).WithProvider(provider)
	
	// Add environment-based configuration
	kwargs := make(map[string]any)
	
	// OpenAI configuration
	if provider == "openai" {
		if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
			kwargs["api_key"] = apiKey
		}
		if baseURL := os.Getenv("OPENAI_BASE_URL"); baseURL != "" {
			kwargs["base_url"] = baseURL
		}
	}
	
	// Gemini configuration
	if provider == "gemini" {
		if apiKey := os.Getenv("GEMINI_API_KEY"); apiKey != "" {
			kwargs["api_key"] = apiKey
		} else if apiKey := os.Getenv("GOOGLE_API_KEY"); apiKey != "" {
			kwargs["api_key"] = apiKey
		}
		if baseURL := os.Getenv("GEMINI_BASE_URL"); baseURL != "" {
			kwargs["base_url"] = baseURL
		}
	}
	
	// Ollama configuration
	if provider == "ollama" {
		if baseURL := os.Getenv("OLLAMA_BASE_URL"); baseURL != "" {
			kwargs["base_url"] = baseURL
		} else {
			// Default to localhost if not specified
			kwargs["base_url"] = "http://localhost:11434"
		}
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

// isGeminiModel checks if a model ID belongs to Google Gemini.
func isGeminiModel(modelID string) bool {
	geminiModels := []string{
		"gemini-2.5-flash", "gemini-2.5-pro",
		"gemini-1.5-pro", "gemini-1.5-flash", 
		"gemini-pro",
	}
	
	for _, model := range geminiModels {
		if modelID == model {
			return true
		}
	}
	return false
}

// isOllamaModel checks if a model ID belongs to Ollama.
func isOllamaModel(modelID string) bool {
	ollamaModels := []string{
		"llama3.2", "llama3.2:1b", "llama3.2:3b",
		"llama3.1", "llama3.1:8b", "llama3.1:70b",
		"codellama", "mistral", "qwen2.5",
	}
	
	for _, model := range ollamaModels {
		if modelID == model {
			return true
		}
	}
	return false
}