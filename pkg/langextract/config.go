package langextract

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Config represents global library configuration.
// This mirrors the configuration approach from Google's langextract Python library.
type Config struct {
	// API Keys for different providers
	OpenAIAPIKey string
	GeminiAPIKey string
	OllamaURL    string

	// Default provider settings
	DefaultModelID   string
	DefaultProvider  string
	DefaultTimeout   time.Duration
	DefaultRetries   int
	DefaultDebugMode bool

	// Global behavior settings
	EnableCaching    bool
	CacheDirectory   string
	LogLevel         string
	MaxConcurrency   int
	ConfigFilePath   string
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		DefaultModelID:   "gemini-2.5-flash",
		DefaultProvider:  "",
		DefaultTimeout:   60 * time.Second,
		DefaultRetries:   2,
		DefaultDebugMode: false,
		EnableCaching:    false,
		CacheDirectory:   "",
		LogLevel:         "info",
		MaxConcurrency:   10,
		ConfigFilePath:   "",
	}
}

// LoadConfig loads configuration from environment variables and optional config file.
// This follows the Google langextract pattern of environment-first configuration.
func LoadConfig() (*Config, error) {
	config := DefaultConfig()

	// Load from environment variables (highest priority)
	if err := config.loadFromEnvironment(); err != nil {
		return nil, fmt.Errorf("failed to load environment configuration: %w", err)
	}

	// Load from .env file if present (lower priority)
	if err := config.loadFromEnvFile(); err != nil {
		// .env file is optional, so we only warn on parse errors, not missing file
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load .env file: %w", err)
		}
	}

	// Load from config file if specified
	if config.ConfigFilePath != "" {
		if err := config.loadFromFile(config.ConfigFilePath); err != nil {
			return nil, fmt.Errorf("failed to load config file %s: %w", config.ConfigFilePath, err)
		}
	}

	return config, nil
}

// loadFromEnvironment loads configuration from environment variables.
func (c *Config) loadFromEnvironment() error {
	// API Keys (following Google langextract naming convention)
	if key := os.Getenv("LANGEXTRACT_API_KEY"); key != "" {
		// Generic API key - try to determine provider or set as default
		c.OpenAIAPIKey = key
	}
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		c.OpenAIAPIKey = key
	}
	if key := os.Getenv("GEMINI_API_KEY"); key != "" {
		c.GeminiAPIKey = key
	}
	if key := os.Getenv("GOOGLE_API_KEY"); key != "" {
		c.GeminiAPIKey = key
	}
	if url := os.Getenv("OLLAMA_URL"); url != "" {
		c.OllamaURL = url
	}

	// Default settings
	if modelID := os.Getenv("LANGEXTRACT_MODEL_ID"); modelID != "" {
		c.DefaultModelID = modelID
	}
	if provider := os.Getenv("LANGEXTRACT_PROVIDER"); provider != "" {
		c.DefaultProvider = provider
	}
	if timeoutStr := os.Getenv("LANGEXTRACT_TIMEOUT"); timeoutStr != "" {
		if timeout, err := time.ParseDuration(timeoutStr); err == nil {
			c.DefaultTimeout = timeout
		}
	}
	if retriesStr := os.Getenv("LANGEXTRACT_RETRIES"); retriesStr != "" {
		if retries, err := strconv.Atoi(retriesStr); err == nil {
			c.DefaultRetries = retries
		}
	}
	if debugStr := os.Getenv("LANGEXTRACT_DEBUG"); debugStr != "" {
		if debug, err := strconv.ParseBool(debugStr); err == nil {
			c.DefaultDebugMode = debug
		}
	}

	// Global settings
	if cachingStr := os.Getenv("LANGEXTRACT_ENABLE_CACHING"); cachingStr != "" {
		if caching, err := strconv.ParseBool(cachingStr); err == nil {
			c.EnableCaching = caching
		}
	}
	if cacheDir := os.Getenv("LANGEXTRACT_CACHE_DIR"); cacheDir != "" {
		c.CacheDirectory = cacheDir
	}
	if logLevel := os.Getenv("LANGEXTRACT_LOG_LEVEL"); logLevel != "" {
		c.LogLevel = logLevel
	}
	if concurrencyStr := os.Getenv("LANGEXTRACT_MAX_CONCURRENCY"); concurrencyStr != "" {
		if concurrency, err := strconv.Atoi(concurrencyStr); err == nil {
			c.MaxConcurrency = concurrency
		}
	}
	if configFile := os.Getenv("LANGEXTRACT_CONFIG_FILE"); configFile != "" {
		c.ConfigFilePath = configFile
	}

	return nil
}

// loadFromEnvFile loads configuration from a .env file in the current directory.
func (c *Config) loadFromEnvFile() error {
	envFile := ".env"
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		return err
	}

	file, err := os.Open(envFile)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE format
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		if len(value) >= 2 {
			if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
				(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
				value = value[1 : len(value)-1]
			}
		}

		// Set environment variable if not already set
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	// Re-load from environment now that .env is loaded
	return c.loadFromEnvironment()
}

// loadFromFile loads configuration from a specified file path.
func (c *Config) loadFromFile(filePath string) error {
	// For now, we'll support the same .env format
	// This can be extended to support JSON/YAML in the future
	
	if !filepath.IsAbs(filePath) {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		filePath = filepath.Join(wd, filePath)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE format
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		if len(value) >= 2 {
			if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
				(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
				value = value[1 : len(value)-1]
			}
		}

		// Apply configuration values directly
		c.applyConfigValue(key, value)
	}

	return scanner.Err()
}

// applyConfigValue applies a single configuration key-value pair.
func (c *Config) applyConfigValue(key, value string) {
	switch key {
	case "LANGEXTRACT_API_KEY", "OPENAI_API_KEY":
		c.OpenAIAPIKey = value
	case "GEMINI_API_KEY", "GOOGLE_API_KEY":
		c.GeminiAPIKey = value
	case "OLLAMA_URL":
		c.OllamaURL = value
	case "LANGEXTRACT_MODEL_ID":
		c.DefaultModelID = value
	case "LANGEXTRACT_PROVIDER":
		c.DefaultProvider = value
	case "LANGEXTRACT_TIMEOUT":
		if timeout, err := time.ParseDuration(value); err == nil {
			c.DefaultTimeout = timeout
		}
	case "LANGEXTRACT_RETRIES":
		if retries, err := strconv.Atoi(value); err == nil {
			c.DefaultRetries = retries
		}
	case "LANGEXTRACT_DEBUG":
		if debug, err := strconv.ParseBool(value); err == nil {
			c.DefaultDebugMode = debug
		}
	case "LANGEXTRACT_ENABLE_CACHING":
		if caching, err := strconv.ParseBool(value); err == nil {
			c.EnableCaching = caching
		}
	case "LANGEXTRACT_CACHE_DIR":
		c.CacheDirectory = value
	case "LANGEXTRACT_LOG_LEVEL":
		c.LogLevel = value
	case "LANGEXTRACT_MAX_CONCURRENCY":
		if concurrency, err := strconv.Atoi(value); err == nil {
			c.MaxConcurrency = concurrency
		}
	}
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c.DefaultTimeout <= 0 {
		return NewValidationError("DefaultTimeout", c.DefaultTimeout.String(), "must be positive")
	}

	if c.DefaultRetries < 0 {
		return NewValidationError("DefaultRetries", strconv.Itoa(c.DefaultRetries), "must be non-negative")
	}

	if c.MaxConcurrency <= 0 {
		return NewValidationError("MaxConcurrency", strconv.Itoa(c.MaxConcurrency), "must be positive")
	}

	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
		"fatal": true,
	}
	if !validLogLevels[strings.ToLower(c.LogLevel)] {
		return NewValidationError("LogLevel", c.LogLevel, "must be one of: debug, info, warn, error, fatal")
	}

	return nil
}

// HasProviderCredentials checks if credentials are available for the specified provider.
func (c *Config) HasProviderCredentials(provider string) bool {
	switch strings.ToLower(provider) {
	case "openai":
		return c.OpenAIAPIKey != ""
	case "gemini":
		return c.GeminiAPIKey != ""
	case "ollama":
		return c.OllamaURL != ""
	default:
		return false
	}
}

// GetAPIKey returns the API key for the specified provider.
func (c *Config) GetAPIKey(provider string) string {
	switch strings.ToLower(provider) {
	case "openai":
		return c.OpenAIAPIKey
	case "gemini":
		return c.GeminiAPIKey
	default:
		return ""
	}
}

// Global configuration instance
var globalConfig *Config

// GetGlobalConfig returns the global configuration instance.
// If not initialized, it loads from environment and .env file.
func GetGlobalConfig() (*Config, error) {
	if globalConfig == nil {
		config, err := LoadConfig()
		if err != nil {
			return nil, err
		}
		globalConfig = config
	}
	return globalConfig, nil
}

// SetGlobalConfig sets the global configuration instance.
func SetGlobalConfig(config *Config) error {
	if err := config.Validate(); err != nil {
		return err
	}
	globalConfig = config
	return nil
}

// ResetGlobalConfig resets the global configuration to nil.
// Useful for testing or forcing re-initialization.
func ResetGlobalConfig() {
	globalConfig = nil
}