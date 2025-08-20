package providers_test

import (
	"context"
	"testing"

	"github.com/sehwan505/langextract-go/pkg/providers"
)

func TestModelConfig(t *testing.T) {
	config := providers.NewModelConfig("gpt-4")
	
	if config.ModelID != "gpt-4" {
		t.Errorf("ModelID = %q, want 'gpt-4'", config.ModelID)
	}
	
	if config.Temperature != 0.0 {
		t.Errorf("Temperature = %f, want 0.0", config.Temperature)
	}
	
	if config.MaxTokens != 1024 {
		t.Errorf("MaxTokens = %d, want 1024", config.MaxTokens)
	}
}

func TestModelConfigChaining(t *testing.T) {
	config := providers.NewModelConfig("gpt-4").
		WithProvider("openai").
		WithTemperature(0.5).
		WithProviderKwargs(map[string]interface{}{
			"api_key": "test-key",
		})
	
	if config.Provider != "openai" {
		t.Errorf("Provider = %q, want 'openai'", config.Provider)
	}
	
	if config.Temperature != 0.5 {
		t.Errorf("Temperature = %f, want 0.5", config.Temperature)
	}
	
	if config.ProviderKwargs["api_key"] != "test-key" {
		t.Errorf("ProviderKwargs[api_key] = %v, want 'test-key'", config.ProviderKwargs["api_key"])
	}
}

func TestProviderRegistry(t *testing.T) {
	registry := providers.NewProviderRegistry()
	
	// Test registration
	mockFactory := func(config *providers.ModelConfig) (providers.BaseLanguageModel, error) {
		return &MockProvider{config: config}, nil
	}
	
	registry.Register("mock", mockFactory)
	
	if !registry.HasProvider("mock") {
		t.Error("Provider 'mock' should be registered")
	}
	
	providersList := registry.GetAvailableProviders()
	found := false
	for _, provider := range providersList {
		if provider == "mock" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Provider 'mock' should be in available providers list")
	}
}

func TestProviderRegistryAliases(t *testing.T) {
	registry := providers.NewProviderRegistry()
	
	mockFactory := func(config *providers.ModelConfig) (providers.BaseLanguageModel, error) {
		return &MockProvider{config: config}, nil
	}
	
	registry.Register("mock", mockFactory)
	registry.RegisterAlias("test-model", "mock")
	
	// Test creating model with alias
	config := providers.NewModelConfig("test-model")
	model, err := registry.CreateModel(config)
	if err != nil {
		t.Fatalf("CreateModel() error = %v", err)
	}
	
	if model.GetModelID() != "test-model" {
		t.Errorf("GetModelID() = %q, want 'test-model'", model.GetModelID())
	}
}

func TestCreateModelErrors(t *testing.T) {
	registry := providers.NewProviderRegistry()
	
	// Test unknown provider
	config := providers.NewModelConfig("unknown-model").WithProvider("unknown")
	_, err := registry.CreateModel(config)
	if err == nil {
		t.Error("CreateModel() should return error for unknown provider")
	}
	
	// Test no provider and no alias
	config = providers.NewModelConfig("unknown-model")
	_, err = registry.CreateModel(config)
	if err == nil {
		t.Error("CreateModel() should return error when no provider specified and no alias found")
	}
}

// MockProvider implements BaseLanguageModel for testing
type MockProvider struct {
	config      *providers.ModelConfig
	schema      interface{}
	fenceOutput bool
}

func (m *MockProvider) Infer(ctx context.Context, prompts []string, options map[string]interface{}) ([][]providers.ScoredOutput, error) {
	results := make([][]providers.ScoredOutput, len(prompts))
	for i, prompt := range prompts {
		results[i] = []providers.ScoredOutput{
			{Output: "Mock response for: " + prompt, Score: 1.0},
		}
	}
	return results, nil
}

func (m *MockProvider) ParseOutput(output string) (interface{}, error) {
	return output, nil
}

func (m *MockProvider) ApplySchema(schema interface{}) {
	m.schema = schema
}

func (m *MockProvider) SetFenceOutput(enabled bool) {
	m.fenceOutput = enabled
}

func (m *MockProvider) GetModelID() string {
	return m.config.ModelID
}

func (m *MockProvider) IsAvailable() bool {
	return true
}

func TestGeminiProviderCreation(t *testing.T) {
	config := providers.NewModelConfig("gemini-2.5-flash").
		WithProvider("gemini").
		WithProviderKwargs(map[string]interface{}{
			"api_key": "test-gemini-key",
		})
	
	provider, err := providers.NewGeminiProvider(config)
	if err != nil {
		t.Fatalf("NewGeminiProvider() error = %v", err)
	}
	
	if provider.GetModelID() != "gemini-2.5-flash" {
		t.Errorf("GetModelID() = %q, want 'gemini-2.5-flash'", provider.GetModelID())
	}
	
	if !provider.IsAvailable() {
		t.Error("IsAvailable() should return true when API key is provided")
	}
}

func TestGeminiProviderWithoutAPIKey(t *testing.T) {
	config := providers.NewModelConfig("gemini-2.5-flash").WithProvider("gemini")
	
	_, err := providers.NewGeminiProvider(config)
	if err == nil {
		t.Error("NewGeminiProvider() should return error when no API key is provided")
	}
}

func TestGeminiProviderParseOutput(t *testing.T) {
	config := providers.NewModelConfig("gemini-2.5-flash").
		WithProvider("gemini").
		WithProviderKwargs(map[string]interface{}{
			"api_key": "test-key",
		})
	
	provider, err := providers.NewGeminiProvider(config)
	if err != nil {
		t.Fatalf("NewGeminiProvider() error = %v", err)
	}
	
	// Test JSON parsing
	jsonOutput := `{"name": "test", "value": 123}`
	parsed, err := provider.ParseOutput(jsonOutput)
	if err != nil {
		t.Fatalf("ParseOutput() error = %v", err)
	}
	
	result, ok := parsed.(map[string]interface{})
	if !ok {
		t.Error("ParseOutput() should return map[string]interface{} for valid JSON")
	}
	
	if result["name"] != "test" {
		t.Errorf("ParseOutput() name = %v, want 'test'", result["name"])
	}
	
	// Test non-JSON parsing
	textOutput := "plain text"
	parsed, err = provider.ParseOutput(textOutput)
	if err != nil {
		t.Fatalf("ParseOutput() error = %v", err)
	}
	
	if parsed != textOutput {
		t.Errorf("ParseOutput() = %v, want %v", parsed, textOutput)
	}
}

func TestOllamaProviderCreation(t *testing.T) {
	config := providers.NewModelConfig("llama3.2").
		WithProvider("ollama").
		WithProviderKwargs(map[string]interface{}{
			"base_url": "http://localhost:11434",
		})
	
	// Note: This will fail if Ollama is not running, but we can still test the basic creation
	_, err := providers.NewOllamaProvider(config)
	// We expect this to potentially fail since Ollama might not be running
	// so we'll just check that the function doesn't panic
	if err != nil {
		t.Logf("NewOllamaProvider() error = %v (expected if Ollama is not running)", err)
	}
}

func TestOllamaProviderParseOutput(t *testing.T) {
	config := providers.NewModelConfig("llama3.2").
		WithProvider("ollama").
		WithProviderKwargs(map[string]interface{}{
			"base_url": "http://localhost:11434",
		})
	
	// Create provider without checking availability for unit testing
	provider := &providers.OllamaProvider{
		Config:  config,
		BaseURL: "http://localhost:11434",
	}
	
	// Test JSON parsing
	jsonOutput := `{"name": "test", "value": 123}`
	parsed, err := provider.ParseOutput(jsonOutput)
	if err != nil {
		t.Fatalf("ParseOutput() error = %v", err)
	}
	
	result, ok := parsed.(map[string]interface{})
	if !ok {
		t.Error("ParseOutput() should return map[string]interface{} for valid JSON")
	}
	
	if result["name"] != "test" {
		t.Errorf("ParseOutput() name = %v, want 'test'", result["name"])
	}
	
	// Test non-JSON parsing
	textOutput := "plain text"
	parsed, err = provider.ParseOutput(textOutput)
	if err != nil {
		t.Fatalf("ParseOutput() error = %v", err)
	}
	
	if parsed != textOutput {
		t.Errorf("ParseOutput() = %v, want %v", parsed, textOutput)
	}
}

func TestFactoryFunctions(t *testing.T) {
	// Test CreateGeminiPro
	t.Run("CreateGeminiPro", func(t *testing.T) {
		_, err := providers.CreateGeminiPro("test-key")
		// We expect this to create without error since we provided an API key
		if err != nil {
			t.Errorf("CreateGeminiPro() error = %v", err)
		}
	})
	
	// Test CreateGeminiFlash
	t.Run("CreateGeminiFlash", func(t *testing.T) {
		_, err := providers.CreateGeminiFlash("test-key")
		if err != nil {
			t.Errorf("CreateGeminiFlash() error = %v", err)
		}
	})
	
	// Test CreateOllamaLlama
	t.Run("CreateOllamaLlama", func(t *testing.T) {
		_, err := providers.CreateOllamaLlama("http://localhost:11434")
		// This will likely fail if Ollama is not running, but that's expected
		if err != nil {
			t.Logf("CreateOllamaLlama() error = %v (expected if Ollama is not running)", err)
		}
	})
}