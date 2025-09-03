package providers_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

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

// MockProviderWithPriority extends MockProvider with priority information
type MockProviderWithPriority struct {
	config   *providers.ModelConfig
	priority string
	schema   interface{}
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

// MockProviderWithPriority methods
func (m *MockProviderWithPriority) Infer(ctx context.Context, prompts []string, options map[string]interface{}) ([][]providers.ScoredOutput, error) {
	results := make([][]providers.ScoredOutput, len(prompts))
	for i, prompt := range prompts {
		results[i] = []providers.ScoredOutput{
			{Output: "Mock response with " + m.priority + " priority for: " + prompt, Score: 1.0},
		}
	}
	return results, nil
}

func (m *MockProviderWithPriority) ParseOutput(output string) (interface{}, error) {
	return output, nil
}

func (m *MockProviderWithPriority) ApplySchema(schema interface{}) {
	m.schema = schema
}

func (m *MockProviderWithPriority) SetFenceOutput(enabled bool) {
	m.fenceOutput = enabled
}

func (m *MockProviderWithPriority) GetModelID() string {
	return m.config.ModelID
}

func (m *MockProviderWithPriority) IsAvailable() bool {
	return true
}

// Helper functions for environment management
func saveEnvironment() map[string]string {
	env := make(map[string]string)
	for _, kv := range os.Environ() {
		if idx := strings.Index(kv, "="); idx != -1 {
			key := kv[:idx]
			value := kv[idx+1:]
			env[key] = value
		}
	}
	return env
}

func restoreEnvironment(env map[string]string) {
	// Clear current environment
	os.Clearenv()
	
	// Restore original environment
	for key, value := range env {
		os.Setenv(key, value)
	}
}

func clearTestEnvironment() {
	testKeys := []string{
		"OPENAI_API_KEY", "OPENAI_BASE_URL",
		"GEMINI_API_KEY", "GOOGLE_API_KEY", "GEMINI_BASE_URL",
		"OLLAMA_BASE_URL", "LANGEXTRACT_API_KEY",
	}
	
	for _, key := range testKeys {
		os.Unsetenv(key)
	}
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
	
	// Create provider using constructor
	provider, err := providers.NewOllamaProvider(config)
	if err != nil {
		if strings.Contains(err.Error(), "Ollama server is not available") {
			t.Skip("Ollama server not available - skipping test")
		}
		t.Fatalf("Failed to create OllamaProvider: %v", err)
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

// TestEnvironmentVariableHandling tests provider creation with environment variables
// Following patterns from Python factory_test.py
func TestEnvironmentVariableHandling(t *testing.T) {
	tests := []struct {
		name     string
		modelID  string
		envVars  map[string]string
		expected string // expected API key or base URL
		wantErr  bool
	}{
		{
			name:    "Gemini with GEMINI_API_KEY",
			modelID: "gemini-2.5-flash",
			envVars: map[string]string{"GEMINI_API_KEY": "test-gemini-key"},
			expected: "test-gemini-key",
			wantErr:  false,
		},
		{
			name:    "Gemini with GOOGLE_API_KEY fallback",
			modelID: "gemini-2.5-flash",
			envVars: map[string]string{"GOOGLE_API_KEY": "test-google-key"},
			expected: "test-google-key",
			wantErr:  false,
		},
		{
			name:    "OpenAI with OPENAI_API_KEY",
			modelID: "gpt-4",
			envVars: map[string]string{"OPENAI_API_KEY": "test-openai-key"},
			expected: "test-openai-key",
			wantErr:  false,
		},
		{
			name:    "Ollama with OLLAMA_BASE_URL",
			modelID: "llama3.2",
			envVars: map[string]string{"OLLAMA_BASE_URL": "http://custom:11434"},
			expected: "http://custom:11434",
			wantErr:  false,
		},
		{
			name:    "Ollama with default localhost",
			modelID: "llama3.2",
			envVars: map[string]string{},
			expected: "http://localhost:11434",
			wantErr:  false,
		},
		{
			name:    "Unknown model ID",
			modelID: "unknown-model",
			envVars: map[string]string{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original environment
			originalEnv := saveEnvironment()
			defer restoreEnvironment(originalEnv)
			
			// Clear environment and set test variables
			clearTestEnvironment()
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Test CreateModelFromEnv
			model, err := providers.CreateModelFromEnv(tt.modelID)

			if tt.wantErr {
				if err == nil {
					t.Errorf("CreateModelFromEnv() error = nil, wantErr %v", tt.wantErr)
				}
				return
			}

			if err != nil {
				// Skip Ollama tests if server is not available
				if strings.Contains(tt.modelID, "llama") && strings.Contains(err.Error(), "Ollama server is not available") {
					t.Skipf("Ollama server not available - skipping test: %v", err)
				}
				t.Fatalf("CreateModelFromEnv() error = %v, wantErr %v", err, tt.wantErr)
			}

			if model == nil {
				t.Error("CreateModelFromEnv() returned nil model")
				return
			}

			// Verify model properties based on provider type
			if model.GetModelID() != tt.modelID {
				t.Errorf("Model ID = %s, want %s", model.GetModelID(), tt.modelID)
			}
		})
	}
}

// TestProviderPrioritySystem tests provider selection priority
// Following patterns from Python factory_test.py
func TestProviderPrioritySystem(t *testing.T) {
	registry := providers.NewProviderRegistry()

	// Register multiple providers for the same model pattern with different priorities
	highPriorityFactory := func(config *providers.ModelConfig) (providers.BaseLanguageModel, error) {
		return &MockProviderWithPriority{config: config, priority: "high"}, nil
	}
	lowPriorityFactory := func(config *providers.ModelConfig) (providers.BaseLanguageModel, error) {
		return &MockProviderWithPriority{config: config, priority: "low"}, nil
	}

	// Register providers (in Go, we simulate priority with explicit provider names)
	registry.Register("high-priority-gemini", highPriorityFactory)
	registry.Register("low-priority-gemini", lowPriorityFactory)

	// Test explicit provider selection overrides priority
	config := providers.NewModelConfig("gemini-test").WithProvider("low-priority-gemini")
	model, err := registry.CreateModel(config)
	if err != nil {
		t.Fatalf("CreateModel() error = %v", err)
	}

	mockModel, ok := model.(*MockProviderWithPriority)
	if !ok {
		t.Fatalf("Expected MockProviderWithPriority, got %T", model)
	}

	if mockModel.priority != "low" {
		t.Errorf("Expected low priority provider, got %s", mockModel.priority)
	}
}

// TestProviderConfigurationOptions tests various provider configuration scenarios
func TestProviderConfigurationOptions(t *testing.T) {
	tests := []struct {
		name           string
		config         *providers.ModelConfig
		expectProvider string
		wantErr        bool
	}{
		{
			name: "OpenAI with all options",
			config: providers.NewModelConfig("gpt-4").
				WithProvider("openai").
				WithTemperature(0.7).
				WithMaxTokens(2048).
				WithProviderKwargs(map[string]interface{}{
					"api_key": "test-key",
					"timeout": 30,
				}),
			expectProvider: "openai",
			wantErr:        false,
		},
		{
			name: "Gemini with response schema",
			config: providers.NewModelConfig("gemini-2.5-flash").
				WithProvider("gemini").
				WithProviderKwargs(map[string]interface{}{
					"api_key":        "test-key",
					"response_schema": map[string]interface{}{"type": "object"},
				}),
			expectProvider: "gemini",
			wantErr:        false,
		},
		{
			name: "Ollama with JSON format",
			config: providers.NewModelConfig("llama3.2").
				WithProvider("ollama").
				WithProviderKwargs(map[string]interface{}{
					"base_url": "http://localhost:11434",
					"format":   "json",
				}),
			expectProvider: "ollama",
			wantErr:        false,
		},
		{
			name:    "Missing provider and no alias",
			config:  providers.NewModelConfig("unknown-model"),
			wantErr: true,
		},
		{
			name:    "Unknown provider",
			config:  providers.NewModelConfig("test-model").WithProvider("unknown-provider"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := providers.NewProviderRegistry()
			providers.RegisterDefaultProviders(registry)

			model, err := registry.CreateModel(tt.config)

			if tt.wantErr {
				if err == nil {
					t.Errorf("CreateModel() error = nil, wantErr %v", tt.wantErr)
				}
				return
			}

			if err != nil {
				// Skip Ollama tests if server is not available
				if tt.expectProvider == "ollama" && strings.Contains(err.Error(), "Ollama server is not available") {
					t.Skipf("Ollama server not available - skipping test: %v", err)
				}
				t.Errorf("CreateModel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if model == nil {
				t.Error("CreateModel() returned nil model")
				return
			}

			// Additional validations can be added here
			if model.GetModelID() != tt.config.ModelID {
				t.Errorf("Model ID = %s, want %s", model.GetModelID(), tt.config.ModelID)
			}
		})
	}
}

// TestProviderAvailabilityChecks tests provider availability checking
func TestProviderAvailabilityChecks(t *testing.T) {
	tests := []struct {
		name          string
		setupEnv      func()
		cleanupEnv    func()
		modelID       string
		expectAvailable bool
	}{
		{
			name: "Gemini available with API key",
			setupEnv: func() {
				os.Setenv("GEMINI_API_KEY", "test-key")
			},
			cleanupEnv: func() {
				os.Unsetenv("GEMINI_API_KEY")
			},
			modelID:         "gemini-2.5-flash",
			expectAvailable: true,
		},
		{
			name: "OpenAI available with API key",
			setupEnv: func() {
				os.Setenv("OPENAI_API_KEY", "test-key")
			},
			cleanupEnv: func() {
				os.Unsetenv("OPENAI_API_KEY")
			},
			modelID:         "gpt-4",
			expectAvailable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupEnv != nil {
				tt.setupEnv()
			}
			if tt.cleanupEnv != nil {
				defer tt.cleanupEnv()
			}

			model, err := providers.CreateModelFromEnv(tt.modelID)
			if err != nil {
				if tt.expectAvailable {
					t.Errorf("CreateModelFromEnv() error = %v, expected available", err)
				}
				return
			}

			if model == nil {
				t.Error("CreateModelFromEnv() returned nil model")
				return
			}

			isAvailable := model.IsAvailable()
			if isAvailable != tt.expectAvailable {
				t.Errorf("IsAvailable() = %v, want %v", isAvailable, tt.expectAvailable)
			}
		})
	}
}

// TestConcurrentProviderCreation tests thread-safe provider creation
func TestConcurrentProviderCreation(t *testing.T) {
	registry := providers.NewProviderRegistry()
	providers.RegisterDefaultProviders(registry)

	// Test concurrent model creation
	const numGoroutines = 10
	const numIterations = 10

	done := make(chan bool, numGoroutines)
	errors := make(chan error, numGoroutines*numIterations)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			for j := 0; j < numIterations; j++ {
				config := providers.NewModelConfig("gemini-2.5-flash").
					WithProvider("gemini").
					WithProviderKwargs(map[string]interface{}{
						"api_key": "test-key",
					})

				_, err := registry.CreateModel(config)
				if err != nil {
					errors <- err
				}
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		select {
		case <-done:
			// Success
		case <-time.After(5 * time.Second):
			t.Fatal("Test timed out")
		}
	}

	// Check for errors
	close(errors)
	for err := range errors {
		t.Errorf("Concurrent creation error: %v", err)
	}
}