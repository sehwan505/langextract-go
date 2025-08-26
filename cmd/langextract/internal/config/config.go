package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// GlobalConfig contains application-wide configuration settings
type GlobalConfig struct {
	// Logging configuration
	LogLevel  string `mapstructure:"log_level" json:"log_level"`
	LogFormat string `mapstructure:"log_format" json:"log_format"`
	
	// Default provider settings
	DefaultProvider string `mapstructure:"default_provider" json:"default_provider"`
	
	// API configuration
	OpenAIAPIKey    string `mapstructure:"openai_api_key" json:"-"`
	GeminiAPIKey    string `mapstructure:"gemini_api_key" json:"-"`
	OllamaEndpoint  string `mapstructure:"ollama_endpoint" json:"ollama_endpoint"`
	
	// Request configuration
	RequestTimeout  int `mapstructure:"request_timeout" json:"request_timeout"`
	MaxRetries     int `mapstructure:"max_retries" json:"max_retries"`
	RetryDelay     int `mapstructure:"retry_delay" json:"retry_delay"`
	
	// Output configuration
	DefaultFormat    string `mapstructure:"default_format" json:"default_format"`
	PrettyPrint     bool   `mapstructure:"pretty_print" json:"pretty_print"`
	ColorOutput     bool   `mapstructure:"color_output" json:"color_output"`
	
	// Performance settings
	Concurrency     int  `mapstructure:"concurrency" json:"concurrency"`
	CacheEnabled    bool `mapstructure:"cache_enabled" json:"cache_enabled"`
	CacheSize       int  `mapstructure:"cache_size" json:"cache_size"`
	
	// Progress and UI
	ShowProgress    bool `mapstructure:"show_progress" json:"show_progress"`
	Quiet          bool `mapstructure:"quiet" json:"quiet"`
	Verbose        bool `mapstructure:"verbose" json:"verbose"`
	
	// Validation settings
	StrictValidation bool `mapstructure:"strict_validation" json:"strict_validation"`
	
	// File paths
	ConfigDir   string `mapstructure:"config_dir" json:"config_dir"`
	CacheDir    string `mapstructure:"cache_dir" json:"cache_dir"`
	OutputDir   string `mapstructure:"output_dir" json:"output_dir"`
}

// LoadGlobalConfig loads configuration from multiple sources in order of precedence:
// 1. Command line flags (highest priority)
// 2. Environment variables  
// 3. Configuration file
// 4. Default values (lowest priority)
func LoadGlobalConfig() (*GlobalConfig, error) {
	v := viper.New()
	
	// Set configuration file properties
	v.SetConfigName("langextract")
	v.SetConfigType("yaml")
	
	// Add configuration file search paths
	if err := addConfigPaths(v); err != nil {
		return nil, fmt.Errorf("failed to add config paths: %w", err)
	}
	
	// Set up environment variable handling
	v.SetEnvPrefix("LANGEXTRACT")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	
	// Set default values
	setDefaults(v)
	
	// Read configuration file (ignore if not found)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}
	
	// Unmarshal into struct
	var config GlobalConfig
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	
	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	
	return &config, nil
}

// addConfigPaths adds configuration file search paths
func addConfigPaths(v *viper.Viper) error {
	// Current directory
	v.AddConfigPath(".")
	
	// User config directory
	homeDir, err := os.UserHomeDir()
	if err == nil {
		v.AddConfigPath(filepath.Join(homeDir, ".config", "langextract"))
		v.AddConfigPath(filepath.Join(homeDir, ".langextract"))
	}
	
	// System config directory
	v.AddConfigPath("/etc/langextract")
	
	// XDG config directory
	if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); xdgConfigHome != "" {
		v.AddConfigPath(filepath.Join(xdgConfigHome, "langextract"))
	}
	
	return nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// Logging defaults
	v.SetDefault("log_level", "info")
	v.SetDefault("log_format", "text")
	
	// Provider defaults
	v.SetDefault("default_provider", "openai")
	v.SetDefault("ollama_endpoint", "http://localhost:11434")
	
	// Request defaults
	v.SetDefault("request_timeout", 30)
	v.SetDefault("max_retries", 3)
	v.SetDefault("retry_delay", 1)
	
	// Output defaults
	v.SetDefault("default_format", "json")
	v.SetDefault("pretty_print", true)
	v.SetDefault("color_output", true)
	
	// Performance defaults
	v.SetDefault("concurrency", 4)
	v.SetDefault("cache_enabled", true)
	v.SetDefault("cache_size", 100)
	
	// Progress defaults
	v.SetDefault("show_progress", true)
	v.SetDefault("quiet", false)
	v.SetDefault("verbose", false)
	
	// Validation defaults
	v.SetDefault("strict_validation", false)
	
	// Directory defaults
	homeDir, _ := os.UserHomeDir()
	if homeDir != "" {
		v.SetDefault("config_dir", filepath.Join(homeDir, ".config", "langextract"))
		v.SetDefault("cache_dir", filepath.Join(homeDir, ".cache", "langextract"))
		v.SetDefault("output_dir", filepath.Join(homeDir, "langextract-output"))
	}
}

// Validate validates the configuration
func (c *GlobalConfig) Validate() error {
	// Validate log level
	validLogLevels := []string{"debug", "info", "warn", "error"}
	if !contains(validLogLevels, c.LogLevel) {
		return fmt.Errorf("invalid log_level '%s', must be one of: %v", c.LogLevel, validLogLevels)
	}
	
	// Validate log format
	validLogFormats := []string{"text", "json"}
	if !contains(validLogFormats, c.LogFormat) {
		return fmt.Errorf("invalid log_format '%s', must be one of: %v", c.LogFormat, validLogFormats)
	}
	
	// Validate provider
	validProviders := []string{"openai", "gemini", "ollama"}
	if !contains(validProviders, c.DefaultProvider) {
		return fmt.Errorf("invalid default_provider '%s', must be one of: %v", c.DefaultProvider, validProviders)
	}
	
	// Validate timeout values
	if c.RequestTimeout < 1 || c.RequestTimeout > 300 {
		return fmt.Errorf("request_timeout must be between 1 and 300 seconds")
	}
	
	if c.MaxRetries < 0 || c.MaxRetries > 10 {
		return fmt.Errorf("max_retries must be between 0 and 10")
	}
	
	if c.RetryDelay < 0 || c.RetryDelay > 60 {
		return fmt.Errorf("retry_delay must be between 0 and 60 seconds")
	}
	
	// Validate concurrency
	if c.Concurrency < 1 || c.Concurrency > 100 {
		return fmt.Errorf("concurrency must be between 1 and 100")
	}
	
	// Validate cache size
	if c.CacheSize < 0 || c.CacheSize > 10000 {
		return fmt.Errorf("cache_size must be between 0 and 10000")
	}
	
	// Validate output format
	validFormats := []string{"json", "yaml", "csv", "html", "markdown", "text"}
	if !contains(validFormats, c.DefaultFormat) {
		return fmt.Errorf("invalid default_format '%s', must be one of: %v", c.DefaultFormat, validFormats)
	}
	
	return nil
}

// GetAPIKey returns the API key for the specified provider
func (c *GlobalConfig) GetAPIKey(provider string) string {
	switch strings.ToLower(provider) {
	case "openai":
		return c.OpenAIAPIKey
	case "gemini":
		return c.GeminiAPIKey
	default:
		return ""
	}
}

// HasAPIKey checks if an API key is configured for the provider
func (c *GlobalConfig) HasAPIKey(provider string) bool {
	return c.GetAPIKey(provider) != ""
}

// CreateConfigTemplate creates a configuration file template
func CreateConfigTemplate(path string) error {
	template := `# LangExtract Configuration File
# This file contains default settings for the langextract CLI

# Logging configuration
log_level: info          # debug, info, warn, error
log_format: text         # text, json

# Default provider (openai, gemini, ollama)
default_provider: openai

# API Keys (can also be set via environment variables)
# openai_api_key: sk-...
# gemini_api_key: ...
ollama_endpoint: http://localhost:11434

# Request configuration
request_timeout: 30      # seconds
max_retries: 3
retry_delay: 1          # seconds

# Output configuration
default_format: json    # json, yaml, csv, html, markdown, text
pretty_print: true
color_output: true

# Performance settings
concurrency: 4          # number of parallel requests
cache_enabled: true
cache_size: 100        # number of cached responses

# Progress and UI
show_progress: true
quiet: false
verbose: false

# Validation settings
strict_validation: false

# Directory paths (relative to home directory)
# config_dir: ~/.config/langextract
# cache_dir: ~/.cache/langextract
# output_dir: ~/langextract-output
`

	return os.WriteFile(path, []byte(template), 0644)
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}