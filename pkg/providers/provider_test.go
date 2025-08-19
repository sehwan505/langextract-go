package providers

import (
	"context"
	"testing"
)

func TestModelConfig(t *testing.T) {
	config := NewModelConfig("gpt-4")
	
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
	config := NewModelConfig("gpt-4").
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
	registry := NewProviderRegistry()
	
	// Test registration
	mockFactory := func(config *ModelConfig) (BaseLanguageModel, error) {
		return &MockProvider{config: config}, nil
	}
	
	registry.Register("mock", mockFactory)
	
	if !registry.HasProvider("mock") {
		t.Error("Provider 'mock' should be registered")
	}
	
	providers := registry.GetAvailableProviders()
	found := false
	for _, provider := range providers {
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
	registry := NewProviderRegistry()
	
	mockFactory := func(config *ModelConfig) (BaseLanguageModel, error) {
		return &MockProvider{config: config}, nil
	}
	
	registry.Register("mock", mockFactory)
	registry.RegisterAlias("test-model", "mock")
	
	// Test creating model with alias
	config := NewModelConfig("test-model")
	model, err := registry.CreateModel(config)
	if err != nil {
		t.Fatalf("CreateModel() error = %v", err)
	}
	
	if model.GetModelID() != "test-model" {
		t.Errorf("GetModelID() = %q, want 'test-model'", model.GetModelID())
	}
}

func TestCreateModelErrors(t *testing.T) {
	registry := NewProviderRegistry()
	
	// Test unknown provider
	config := NewModelConfig("unknown-model").WithProvider("unknown")
	_, err := registry.CreateModel(config)
	if err == nil {
		t.Error("CreateModel() should return error for unknown provider")
	}
	
	// Test no provider and no alias
	config = NewModelConfig("unknown-model")
	_, err = registry.CreateModel(config)
	if err == nil {
		t.Error("CreateModel() should return error when no provider specified and no alias found")
	}
}

// MockProvider implements BaseLanguageModel for testing
type MockProvider struct {
	config      *ModelConfig
	schema      interface{}
	fenceOutput bool
}

func (m *MockProvider) Infer(ctx context.Context, prompts []string, options map[string]interface{}) ([][]ScoredOutput, error) {
	results := make([][]ScoredOutput, len(prompts))
	for i, prompt := range prompts {
		results[i] = []ScoredOutput{
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