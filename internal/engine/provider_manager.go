package engine

import (
	"context"
	"crypto/md5"
	"fmt"
	"sync"
	"time"

	"github.com/sehwan505/langextract-go/pkg/providers"
)

// ProviderManager manages multiple language model providers with health monitoring,
// load balancing, and failover capabilities.
type ProviderManager struct {
	registry    *providers.ProviderRegistry
	healthStats map[string]*ProviderHealth
	cache       *ResponseCache
	config      *ProviderManagerConfig
	mu          sync.RWMutex
}

// ProviderManagerConfig configures the provider manager behavior.
type ProviderManagerConfig struct {
	// Health monitoring
	HealthCheckInterval time.Duration
	UnhealthyThreshold  int
	RecoveryThreshold   int
	HealthTimeout       time.Duration

	// Load balancing
	LoadBalanceStrategy LoadBalanceStrategy
	MaxConcurrentRequests int

	// Response caching
	EnableCaching   bool
	CacheTimeout    time.Duration
	MaxCacheSize    int

	// Failover
	EnableFailover    bool
	FailoverTimeout   time.Duration
	MaxFailoverAttempts int
}

// ProviderHealth tracks the health status of a provider.
type ProviderHealth struct {
	ProviderName      string        `json:"provider_name"`
	IsHealthy         bool          `json:"is_healthy"`
	LastHealthCheck   time.Time     `json:"last_health_check"`
	ConsecutiveFailures int         `json:"consecutive_failures"`
	ConsecutiveSuccesses int        `json:"consecutive_successes"`
	AverageLatency    time.Duration `json:"average_latency"`
	TotalRequests     int64         `json:"total_requests"`
	SuccessfulRequests int64        `json:"successful_requests"`
	FailedRequests    int64         `json:"failed_requests"`
	LastError         string        `json:"last_error,omitempty"`
}

// LoadBalanceStrategy defines the load balancing strategy.
type LoadBalanceStrategy string

const (
	RoundRobin     LoadBalanceStrategy = "round_robin"
	LeastLatency   LoadBalanceStrategy = "least_latency"
	HealthyOnly    LoadBalanceStrategy = "healthy_only"
	WeightedRandom LoadBalanceStrategy = "weighted_random"
)

// ResponseCache implements a simple in-memory response cache.
type ResponseCache struct {
	cache    map[string]*CacheEntry
	maxSize  int
	timeout  time.Duration
	mu       sync.RWMutex
}

// CacheEntry represents a cached response.
type CacheEntry struct {
	Response  *CacheableResponse `json:"response"`
	Timestamp time.Time          `json:"timestamp"`
	HitCount  int                `json:"hit_count"`
}

// CacheableResponse represents a cacheable provider response.
type CacheableResponse struct {
	Output      string    `json:"output"`
	TokensUsed  int       `json:"tokens_used"`
	Latency     time.Duration `json:"latency"`
	ProviderID  string    `json:"provider_id"`
	ModelID     string    `json:"model_id"`
}

// NewProviderManager creates a new provider manager with the given configuration.
func NewProviderManager(config *ProviderManagerConfig) *ProviderManager {
	if config == nil {
		config = DefaultProviderManagerConfig()
	}

	pm := &ProviderManager{
		registry:    providers.NewProviderRegistry(),
		healthStats: make(map[string]*ProviderHealth),
		cache:       NewResponseCache(config.MaxCacheSize, config.CacheTimeout),
		config:      config,
	}

	// Register default providers
	providers.RegisterDefaultProviders(pm.registry)

	// Initialize health monitoring
	pm.initializeHealthMonitoring()

	return pm
}

// DefaultProviderManagerConfig returns a default configuration.
func DefaultProviderManagerConfig() *ProviderManagerConfig {
	return &ProviderManagerConfig{
		HealthCheckInterval:   30 * time.Second,
		UnhealthyThreshold:    3,
		RecoveryThreshold:     2,
		HealthTimeout:         10 * time.Second,
		LoadBalanceStrategy:   HealthyOnly,
		MaxConcurrentRequests: 10,
		EnableCaching:         true,
		CacheTimeout:          5 * time.Minute,
		MaxCacheSize:          1000,
		EnableFailover:        true,
		FailoverTimeout:       5 * time.Second,
		MaxFailoverAttempts:   2,
	}
}

// SelectProvider selects the best available provider based on the configuration.
func (pm *ProviderManager) SelectProvider(modelID string, excludeProviders []string) (providers.BaseLanguageModel, string, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	availableProviders := pm.getAvailableProviders(excludeProviders)
	if len(availableProviders) == 0 {
		return nil, "", fmt.Errorf("no available providers for model %s", modelID)
	}

	var selectedProvider string
	switch pm.config.LoadBalanceStrategy {
	case RoundRobin:
		selectedProvider = pm.selectRoundRobin(availableProviders)
	case LeastLatency:
		selectedProvider = pm.selectLeastLatency(availableProviders)
	case HealthyOnly:
		selectedProvider = pm.selectHealthyOnly(availableProviders)
	case WeightedRandom:
		selectedProvider = pm.selectWeightedRandom(availableProviders)
	default:
		selectedProvider = pm.selectHealthyOnly(availableProviders)
	}

	if selectedProvider == "" {
		return nil, "", fmt.Errorf("no suitable provider found for model %s", modelID)
	}

	// Create provider instance
	config := providers.NewModelConfig(modelID)
	provider, err := pm.registry.CreateModel(config)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create provider %s: %w", selectedProvider, err)
	}

	return provider, selectedProvider, nil
}

// ExecuteWithFailover executes a request with automatic failover on failure.
func (pm *ProviderManager) ExecuteWithFailover(ctx context.Context, request *ExtractionRequest) (*CacheableResponse, error) {
	// Check cache first
	if pm.config.EnableCaching {
		if cached := pm.getCachedResponse(request); cached != nil {
			return cached, nil
		}
	}

	var lastErr error
	excludeProviders := make([]string, 0)

	for attempt := 0; attempt <= pm.config.MaxFailoverAttempts; attempt++ {
		// Select provider
		provider, providerName, err := pm.SelectProvider(request.ModelID, excludeProviders)
		if err != nil {
			return nil, fmt.Errorf("provider selection failed: %w", err)
		}

		// Execute request
		startTime := time.Now()
		response, err := pm.executeRequest(ctx, provider, request)
		latency := time.Since(startTime)

		// Update health stats
		pm.updateProviderHealth(providerName, err == nil, latency, err)

		if err == nil {
			// Cache successful response
			if pm.config.EnableCaching {
				pm.cacheResponse(request, response)
			}
			return response, nil
		}

		// Record failure and prepare for retry
		lastErr = err
		excludeProviders = append(excludeProviders, providerName)

		// Log failover event
		if request.ProgressCallback != nil {
			// This would be logged in the response later
		}

		// Check if we should retry
		if attempt >= pm.config.MaxFailoverAttempts {
			break
		}

		// Wait before retry
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Duration(attempt+1) * time.Second):
		}
	}

	return nil, fmt.Errorf("all provider attempts failed, last error: %w", lastErr)
}

// GetProviderHealth returns the health status of all providers.
func (pm *ProviderManager) GetProviderHealth() map[string]*ProviderHealth {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	result := make(map[string]*ProviderHealth)
	for name, health := range pm.healthStats {
		// Create a copy to avoid concurrent modification
		healthCopy := *health
		result[name] = &healthCopy
	}
	return result
}

// GetCacheStats returns cache statistics.
func (pm *ProviderManager) GetCacheStats() map[string]interface{} {
	return pm.cache.GetStats()
}

// initializeHealthMonitoring starts the health monitoring goroutine.
func (pm *ProviderManager) initializeHealthMonitoring() {
	// Initialize health stats for all available providers
	availableProviders := pm.registry.GetAvailableProviders()
	for _, providerName := range availableProviders {
		pm.healthStats[providerName] = &ProviderHealth{
			ProviderName:    providerName,
			IsHealthy:       true,
			LastHealthCheck: time.Now(),
		}
	}

	// Start health monitoring goroutine
	go pm.healthMonitorLoop()
}

// healthMonitorLoop continuously monitors provider health.
func (pm *ProviderManager) healthMonitorLoop() {
	ticker := time.NewTicker(pm.config.HealthCheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		pm.performHealthChecks()
	}
}

// performHealthChecks checks the health of all providers.
func (pm *ProviderManager) performHealthChecks() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for _, health := range pm.healthStats {
		// Create a simple health check request
		config := providers.NewModelConfig("test-model")
		provider, err := pm.registry.CreateModel(config)
		
		isHealthy := err == nil && provider != nil && provider.IsAvailable()

		// Update health status
		health.LastHealthCheck = time.Now()
		if isHealthy {
			health.ConsecutiveSuccesses++
			health.ConsecutiveFailures = 0
			if health.ConsecutiveSuccesses >= pm.config.RecoveryThreshold {
				health.IsHealthy = true
			}
		} else {
			health.ConsecutiveFailures++
			health.ConsecutiveSuccesses = 0
			if health.ConsecutiveFailures >= pm.config.UnhealthyThreshold {
				health.IsHealthy = false
			}
			if err != nil {
				health.LastError = err.Error()
			}
		}
	}
}

// getAvailableProviders returns a list of available providers, excluding specified ones.
func (pm *ProviderManager) getAvailableProviders(excludeProviders []string) []string {
	excludeMap := make(map[string]bool)
	for _, provider := range excludeProviders {
		excludeMap[provider] = true
	}

	var available []string
	for providerName, health := range pm.healthStats {
		if !excludeMap[providerName] && health.IsHealthy {
			available = append(available, providerName)
		}
	}
	return available
}

// Provider selection strategies
func (pm *ProviderManager) selectRoundRobin(providers []string) string {
	if len(providers) == 0 {
		return ""
	}
	// Simple round-robin implementation
	now := time.Now().Unix()
	return providers[now%int64(len(providers))]
}

func (pm *ProviderManager) selectLeastLatency(providers []string) string {
	if len(providers) == 0 {
		return ""
	}

	var bestProvider string
	var lowestLatency time.Duration = time.Hour // Start with a high value

	for _, provider := range providers {
		if health, exists := pm.healthStats[provider]; exists {
			if health.AverageLatency < lowestLatency {
				lowestLatency = health.AverageLatency
				bestProvider = provider
			}
		}
	}

	if bestProvider == "" {
		return providers[0] // Fallback to first provider
	}
	return bestProvider
}

func (pm *ProviderManager) selectHealthyOnly(providers []string) string {
	for _, provider := range providers {
		if health, exists := pm.healthStats[provider]; exists && health.IsHealthy {
			return provider
		}
	}
	return ""
}

func (pm *ProviderManager) selectWeightedRandom(providers []string) string {
	if len(providers) == 0 {
		return ""
	}
	// Simple weighted random based on success rate
	// For now, just return the first healthy provider
	return pm.selectHealthyOnly(providers)
}

// updateProviderHealth updates the health statistics for a provider.
func (pm *ProviderManager) updateProviderHealth(providerName string, success bool, latency time.Duration, err error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	health, exists := pm.healthStats[providerName]
	if !exists {
		health = &ProviderHealth{
			ProviderName: providerName,
			IsHealthy:    true,
		}
		pm.healthStats[providerName] = health
	}

	health.TotalRequests++
	if success {
		health.SuccessfulRequests++
		health.ConsecutiveSuccesses++
		health.ConsecutiveFailures = 0
	} else {
		health.FailedRequests++
		health.ConsecutiveFailures++
		health.ConsecutiveSuccesses = 0
		if err != nil {
			health.LastError = err.Error()
		}
	}

	// Update average latency
	if health.TotalRequests == 1 {
		health.AverageLatency = latency
	} else {
		// Simple moving average
		health.AverageLatency = (health.AverageLatency*time.Duration(health.TotalRequests-1) + latency) / time.Duration(health.TotalRequests)
	}

	// Update health status
	if success && health.ConsecutiveSuccesses >= pm.config.RecoveryThreshold {
		health.IsHealthy = true
	} else if !success && health.ConsecutiveFailures >= pm.config.UnhealthyThreshold {
		health.IsHealthy = false
	}
}

// executeRequest executes a request with the given provider.
func (pm *ProviderManager) executeRequest(ctx context.Context, provider providers.BaseLanguageModel, request *ExtractionRequest) (*CacheableResponse, error) {
	// Build prompt (simplified for now)
	prompt := fmt.Sprintf("Task: %s\nText: %s", request.TaskDescription, request.Text)
	
	// Execute the request
	results, err := provider.Infer(ctx, []string{prompt}, nil)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 || len(results[0]) == 0 {
		return nil, fmt.Errorf("no results returned from provider")
	}

	response := &CacheableResponse{
		Output:     results[0][0].Output,
		TokensUsed: 0, // Would be calculated from actual usage
		ProviderID: request.ProviderID,
		ModelID:    request.ModelID,
	}

	return response, nil
}

// Response caching methods
func (pm *ProviderManager) getCachedResponse(request *ExtractionRequest) *CacheableResponse {
	if !pm.config.EnableCaching {
		return nil
	}

	key := pm.generateCacheKey(request)
	return pm.cache.Get(key)
}

func (pm *ProviderManager) cacheResponse(request *ExtractionRequest, response *CacheableResponse) {
	if !pm.config.EnableCaching {
		return
	}

	key := pm.generateCacheKey(request)
	pm.cache.Set(key, response)
}

func (pm *ProviderManager) generateCacheKey(request *ExtractionRequest) string {
	// Generate a cache key based on request parameters
	data := fmt.Sprintf("%s|%s|%s|%f|%d", 
		request.TaskDescription, 
		request.Text, 
		request.ModelID, 
		request.Temperature, 
		request.MaxTokens)
	
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// NewResponseCache creates a new response cache.
func NewResponseCache(maxSize int, timeout time.Duration) *ResponseCache {
	cache := &ResponseCache{
		cache:   make(map[string]*CacheEntry),
		maxSize: maxSize,
		timeout: timeout,
	}

	// Start cleanup goroutine
	go cache.cleanupLoop()

	return cache
}

// Get retrieves a cached response.
func (c *ResponseCache) Get(key string) *CacheableResponse {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.cache[key]
	if !exists {
		return nil
	}

	// Check if entry has expired
	if time.Since(entry.Timestamp) > c.timeout {
		return nil
	}

	entry.HitCount++
	return entry.Response
}

// Set stores a response in the cache.
func (c *ResponseCache) Set(key string, response *CacheableResponse) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if cache is full
	if len(c.cache) >= c.maxSize {
		c.evictOldest()
	}

	c.cache[key] = &CacheEntry{
		Response:  response,
		Timestamp: time.Now(),
		HitCount:  0,
	}
}

// GetStats returns cache statistics.
func (c *ResponseCache) GetStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	totalHits := 0
	for _, entry := range c.cache {
		totalHits += entry.HitCount
	}

	return map[string]interface{}{
		"size":       len(c.cache),
		"max_size":   c.maxSize,
		"total_hits": totalHits,
	}
}

// cleanupLoop periodically removes expired entries.
func (c *ResponseCache) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanup()
	}
}

// cleanup removes expired entries from the cache.
func (c *ResponseCache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.cache {
		if now.Sub(entry.Timestamp) > c.timeout {
			delete(c.cache, key)
		}
	}
}

// evictOldest removes the oldest entry from the cache.
func (c *ResponseCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range c.cache {
		if oldestKey == "" || entry.Timestamp.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.Timestamp
		}
	}

	if oldestKey != "" {
		delete(c.cache, oldestKey)
	}
}