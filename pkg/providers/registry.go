package providers

import (
	"fmt"
	"sync"
)

// ProviderFactory creates a new language model provider instance.
type ProviderFactory func(config *ModelConfig) (BaseLanguageModel, error)

// ProviderRegistry manages available language model providers.
// This mirrors the provider discovery mechanism from the Python implementation.
type ProviderRegistry struct {
	mu        sync.RWMutex
	providers map[string]ProviderFactory
	aliases   map[string]string // model_id -> provider_name mappings
}

// NewProviderRegistry creates a new provider registry.
func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{
		providers: make(map[string]ProviderFactory),
		aliases:   make(map[string]string),
	}
}

// Register registers a provider factory with the given name.
func (r *ProviderRegistry) Register(name string, factory ProviderFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[name] = factory
}

// RegisterAlias registers a model ID alias that maps to a specific provider.
func (r *ProviderRegistry) RegisterAlias(modelID, providerName string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.aliases[modelID] = providerName
}

// CreateModel creates a language model instance based on the configuration.
// This mirrors the create_model function from the Python implementation.
func (r *ProviderRegistry) CreateModel(config *ModelConfig) (BaseLanguageModel, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Determine provider name
	providerName := config.Provider
	if providerName == "" {
		// Try to find alias for model ID
		if alias, exists := r.aliases[config.ModelID]; exists {
			providerName = alias
		} else {
			return nil, fmt.Errorf("no provider specified and no alias found for model ID: %s", config.ModelID)
		}
	}

	// Get provider factory
	factory, exists := r.providers[providerName]
	if !exists {
		return nil, fmt.Errorf("unknown provider: %s", providerName)
	}

	// Create model instance
	return factory(config)
}

// GetAvailableProviders returns a list of registered provider names.
func (r *ProviderRegistry) GetAvailableProviders() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	providers := make([]string, 0, len(r.providers))
	for name := range r.providers {
		providers = append(providers, name)
	}
	return providers
}

// HasProvider checks if a provider is registered.
func (r *ProviderRegistry) HasProvider(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.providers[name]
	return exists
}

// Global registry instance
var defaultRegistry = NewProviderRegistry()

// Register registers a provider with the default registry.
func Register(name string, factory ProviderFactory) {
	defaultRegistry.Register(name, factory)
}

// RegisterAlias registers a model alias with the default registry.
func RegisterAlias(modelID, providerName string) {
	defaultRegistry.RegisterAlias(modelID, providerName)
}

// CreateModel creates a model using the default registry.
func CreateModel(config *ModelConfig) (BaseLanguageModel, error) {
	return defaultRegistry.CreateModel(config)
}

// GetAvailableProviders returns available providers from the default registry.
func GetAvailableProviders() []string {
	return defaultRegistry.GetAvailableProviders()
}