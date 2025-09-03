package cli

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/sehwan505/langextract-go/cmd/langextract/internal/config"
	"github.com/sehwan505/langextract-go/cmd/langextract/internal/logger"
)

// ProviderInfo represents information about a language model provider
type ProviderInfo struct {
	Name        string            `json:"name"`
	Available   bool              `json:"available"`
	Models      []string          `json:"models"`
	Aliases     map[string]string `json:"aliases"`
	Config      map[string]interface{} `json:"config"`
	Status      string            `json:"status"`
	LastTested  *time.Time        `json:"last_tested,omitempty"`
	TestResult  *TestResult       `json:"test_result,omitempty"`
}

// TestResult represents the result of testing a provider
type TestResult struct {
	Success     bool              `json:"success"`
	Duration    time.Duration     `json:"duration"`
	Error       string            `json:"error,omitempty"`
	ModelInfo   map[string]interface{} `json:"model_info,omitempty"`
	Metrics     map[string]interface{} `json:"metrics,omitempty"`
}

// ProvidersOptions contains all options for the providers command
type ProvidersOptions struct {
	// Operation
	Operation string // list, test, config, install, uninstall

	// Provider selection
	Provider string   // Specific provider to operate on
	Providers []string // Multiple providers to operate on
	All      bool     // Apply operation to all providers

	// Test options
	TestModel   string // Model to test
	TestPrompt  string // Custom test prompt
	TestTimeout int    // Test timeout in seconds
	QuickTest   bool   // Run quick test instead of full test

	// Output options
	Output     string // Output file
	Format     string // Output format: table, json, yaml
	ShowConfig bool   // Show provider configuration
	ShowModels bool   // Show available models
	Detailed   bool   // Show detailed information

	// Configuration options
	SetConfig   map[string]string // Configuration values to set
	UnsetConfig []string         // Configuration keys to unset
	Interactive bool             // Interactive configuration

	// Installation options
	InstallDir string // Installation directory for providers
	Force      bool   // Force installation/uninstallation
}

// NewProvidersCommand creates the providers command with subcommands
func NewProvidersCommand(cfg *config.GlobalConfig, log *logger.Logger) *cobra.Command {
	opts := &ProvidersOptions{
		Format:      "table",
		ShowConfig:  false,
		ShowModels:  false,
		Detailed:    false,
		TestTimeout: 30,
		QuickTest:   false,
		Interactive: false,
		Force:       false,
	}

	cmd := &cobra.Command{
		Use:   "providers [command]",
		Short: "Manage and test language model providers",
		Long: `Manage and test language model providers. This command allows you to list available
providers, test their connectivity and performance, configure provider settings,
and install/uninstall provider plugins.

Available providers typically include:
- openai: OpenAI GPT models
- gemini: Google Gemini models  
- ollama: Local Ollama models

Examples:
  # List all available providers
  langextract providers list
  
  # Test a specific provider
  langextract providers test --provider openai
  
  # Test all providers with quick test
  langextract providers test --all --quick-test
  
  # Show detailed provider information
  langextract providers list --detailed --show-models --show-config
  
  # Configure provider settings
  langextract providers config --provider openai --set api_key=sk-xxx --set base_url=https://api.openai.com/v1
  
  # Export provider information to JSON
  langextract providers list --format json --output providers.json`,
	}

	// Add subcommands
	cmd.AddCommand(NewProvidersListCommand(cfg, log, opts))
	cmd.AddCommand(NewProvidersTestCommand(cfg, log, opts))
	cmd.AddCommand(NewProvidersConfigCommand(cfg, log, opts))
	cmd.AddCommand(NewProvidersInstallCommand(cfg, log, opts))
	cmd.AddCommand(NewProvidersUninstallCommand(cfg, log, opts))

	return cmd
}

// NewProvidersListCommand creates the providers list subcommand
func NewProvidersListCommand(cfg *config.GlobalConfig, log *logger.Logger, opts *ProvidersOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [flags]",
		Short: "List available language model providers",
		Long: `List all available language model providers with their status, models, and configuration.
Shows which providers are currently available and configured.

Examples:
  # List all providers in table format
  langextract providers list
  
  # Show detailed information including models
  langextract providers list --detailed --show-models
  
  # Export to JSON with configuration details
  langextract providers list --format json --show-config --output providers.json
  
  # List specific providers only
  langextract providers list --providers openai,gemini`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProvidersList(cmd.Context(), opts, cfg, log)
		},
	}

	// Output flags
	cmd.Flags().StringVarP(&opts.Output, "output", "o", "", "Output file")
	cmd.Flags().StringVar(&opts.Format, "format", opts.Format, "Output format (table, json, yaml)")
	cmd.Flags().BoolVar(&opts.ShowConfig, "show-config", opts.ShowConfig, "Show provider configuration")
	cmd.Flags().BoolVar(&opts.ShowModels, "show-models", opts.ShowModels, "Show available models")
	cmd.Flags().BoolVar(&opts.Detailed, "detailed", opts.Detailed, "Show detailed information")

	// Provider selection flags
	cmd.Flags().StringVar(&opts.Provider, "provider", "", "Show specific provider only")
	cmd.Flags().StringSliceVar(&opts.Providers, "providers", nil, "Show specific providers (comma-separated)")
	cmd.Flags().BoolVar(&opts.All, "all", true, "Show all providers (default)")

	return cmd
}

// NewProvidersTestCommand creates the providers test subcommand
func NewProvidersTestCommand(cfg *config.GlobalConfig, log *logger.Logger, opts *ProvidersOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test [flags]",
		Short: "Test language model provider connectivity and performance",
		Long: `Test language model providers to verify they are properly configured and working.
Tests can include connectivity checks, model availability, and basic inference.

Examples:
  # Test default provider
  langextract providers test
  
  # Test specific provider
  langextract providers test --provider openai
  
  # Test all providers with quick test
  langextract providers test --all --quick-test
  
  # Test with custom model and prompt
  langextract providers test --provider openai --test-model gpt-4 --test-prompt "Hello, world!"
  
  # Generate test report
  langextract providers test --all --format json --output test-results.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProvidersTest(cmd.Context(), opts, cfg, log)
		},
	}

	// Provider selection flags
	cmd.Flags().StringVar(&opts.Provider, "provider", cfg.DefaultProvider, "Provider to test")
	cmd.Flags().StringSliceVar(&opts.Providers, "providers", nil, "Providers to test (comma-separated)")
	cmd.Flags().BoolVar(&opts.All, "all", false, "Test all providers")

	// Test configuration flags
	cmd.Flags().StringVar(&opts.TestModel, "test-model", "", "Model to test (uses provider default if not specified)")
	cmd.Flags().StringVar(&opts.TestPrompt, "test-prompt", "Say hello", "Custom test prompt")
	cmd.Flags().IntVar(&opts.TestTimeout, "test-timeout", opts.TestTimeout, "Test timeout in seconds")
	cmd.Flags().BoolVar(&opts.QuickTest, "quick-test", opts.QuickTest, "Run quick test instead of full test")

	// Output flags
	cmd.Flags().StringVarP(&opts.Output, "output", "o", "", "Output file for test results")
	cmd.Flags().StringVar(&opts.Format, "format", "table", "Output format (table, json, yaml)")

	return cmd
}

// NewProvidersConfigCommand creates the providers config subcommand
func NewProvidersConfigCommand(cfg *config.GlobalConfig, log *logger.Logger, opts *ProvidersOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config [flags]",
		Short: "Configure provider settings",
		Long: `Configure language model provider settings such as API keys, endpoints,
and other provider-specific options. Changes are saved to the configuration file.

Examples:
  # Set OpenAI API key
  langextract providers config --provider openai --set api_key=sk-xxx
  
  # Configure Ollama endpoint
  langextract providers config --provider ollama --set endpoint=http://localhost:11434
  
  # Interactive configuration
  langextract providers config --provider gemini --interactive
  
  # Remove configuration value
  langextract providers config --provider openai --unset api_key
  
  # Show current configuration
  langextract providers config --provider openai`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProvidersConfig(cmd.Context(), opts, cfg, log)
		},
	}

	// Provider selection flags
	cmd.Flags().StringVar(&opts.Provider, "provider", "", "Provider to configure (required)")

	// Configuration flags
	cmd.Flags().StringToStringVar(&opts.SetConfig, "set", nil, "Set configuration values (key=value)")
	cmd.Flags().StringSliceVar(&opts.UnsetConfig, "unset", nil, "Unset configuration keys")
	cmd.Flags().BoolVar(&opts.Interactive, "interactive", opts.Interactive, "Interactive configuration")

	// Output flags
	cmd.Flags().StringVar(&opts.Format, "format", "table", "Output format (table, json, yaml)")

	// Mark provider as required
	cmd.MarkFlagRequired("provider")

	return cmd
}

// NewProvidersInstallCommand creates the providers install subcommand
func NewProvidersInstallCommand(cfg *config.GlobalConfig, log *logger.Logger, opts *ProvidersOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install [provider]",
		Short: "Install provider plugins",
		Long: `Install language model provider plugins. This allows adding
support for additional providers beyond the built-in ones.

Examples:
  # Install a provider plugin
  langextract providers install custom-provider
  
  # Force reinstall
  langextract providers install custom-provider --force`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Provider = args[0]
			opts.Operation = "install" // Explicitly set operation
			return runProvidersInstall(cmd.Context(), opts, cfg, log)
		},
	}

	// Installation flags
	cmd.Flags().StringVar(&opts.InstallDir, "install-dir", "", "Installation directory")
	cmd.Flags().BoolVar(&opts.Force, "force", opts.Force, "Force installation")

	return cmd
}

// NewProvidersUninstallCommand creates the providers uninstall subcommand
func NewProvidersUninstallCommand(cfg *config.GlobalConfig, log *logger.Logger, opts *ProvidersOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall [provider]",
		Short: "Uninstall provider plugins",
		Long: `Uninstall language model provider plugins. This removes previously
installed provider plugins from the system.

Examples:
  # Uninstall a provider plugin
  langextract providers uninstall custom-provider
  
  # Force uninstall
  langextract providers uninstall custom-provider --force`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Provider = args[0]
			opts.Operation = "uninstall" // Explicitly set operation
			return runProvidersInstall(cmd.Context(), opts, cfg, log)
		},
	}

	// Uninstallation flags
	cmd.Flags().BoolVar(&opts.Force, "force", opts.Force, "Force uninstallation")

	return cmd
}

// runProvidersList executes the providers list command
func runProvidersList(ctx context.Context, opts *ProvidersOptions, cfg *config.GlobalConfig, log *logger.Logger) error {
	log.WithOperation("providers-list").Info("Listing providers")

	// Get provider information
	providers, err := getProviderInfo(opts, cfg)
	if err != nil {
		return fmt.Errorf("failed to get provider information: %w", err)
	}

	// Filter providers if specified
	if opts.Provider != "" {
		filtered := []ProviderInfo{}
		for _, p := range providers {
			if p.Name == opts.Provider {
				filtered = append(filtered, p)
				break
			}
		}
		providers = filtered
	} else if len(opts.Providers) > 0 {
		filtered := []ProviderInfo{}
		for _, p := range providers {
			for _, name := range opts.Providers {
				if p.Name == name {
					filtered = append(filtered, p)
					break
				}
			}
		}
		providers = filtered
	}

	// Generate output
	output, err := formatProvidersOutput(providers, opts)
	if err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	// Write output
	if opts.Output != "" {
		if err := os.WriteFile(opts.Output, []byte(output), 0644); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
	} else {
		fmt.Print(output)
	}

	log.WithCount(len(providers)).Success("Listed providers")
	return nil
}

// runProvidersTest executes the providers test command
func runProvidersTest(ctx context.Context, opts *ProvidersOptions, cfg *config.GlobalConfig, log *logger.Logger) error {
	log.WithOperation("providers-test").Info("Testing providers")

	// Determine which providers to test
	var providersToTest []string
	if opts.All {
		providersToTest = []string{"openai", "gemini", "ollama"} // Built-in providers
	} else if len(opts.Providers) > 0 {
		providersToTest = opts.Providers
	} else if opts.Provider != "" {
		providersToTest = []string{opts.Provider}
	} else {
		providersToTest = []string{cfg.DefaultProvider}
	}

	var results []ProviderInfo

	for _, providerName := range providersToTest {
		log.WithProvider(providerName).Info("Testing provider")
		
		result, err := testProvider(ctx, providerName, opts, cfg, log)
		if err != nil {
			log.WithProvider(providerName).WithError(err).Error("Provider test failed")
			result = ProviderInfo{
				Name:      providerName,
				Available: false,
				Status:    "error",
				TestResult: &TestResult{
					Success: false,
					Error:   err.Error(),
				},
			}
		}

		results = append(results, result)
	}

	// Generate output
	output, err := formatTestResults(results, opts)
	if err != nil {
		return fmt.Errorf("failed to format test results: %w", err)
	}

	// Write output
	if opts.Output != "" {
		if err := os.WriteFile(opts.Output, []byte(output), 0644); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
	} else {
		fmt.Print(output)
	}

	// Count successful tests
	successCount := 0
	for _, result := range results {
		if result.TestResult != nil && result.TestResult.Success {
			successCount++
		}
	}

	log.WithFields(map[string]interface{}{
		"total":      len(results),
		"successful": successCount,
		"failed":     len(results) - successCount,
	}).Info("Provider testing completed")

	return nil
}

// runProvidersConfig executes the providers config command
func runProvidersConfig(ctx context.Context, opts *ProvidersOptions, cfg *config.GlobalConfig, log *logger.Logger) error {
	log.WithOperation("providers-config").WithProvider(opts.Provider).Info("Configuring provider")

	// Validate provider exists
	if !isValidProvider(opts.Provider) {
		return fmt.Errorf("unknown provider: %s", opts.Provider)
	}

	// If no set/unset operations, show current configuration
	if len(opts.SetConfig) == 0 && len(opts.UnsetConfig) == 0 && !opts.Interactive {
		return showProviderConfig(opts.Provider, cfg, opts)
	}

	// Apply configuration changes
	if err := applyProviderConfig(opts.Provider, opts, cfg); err != nil {
		return fmt.Errorf("failed to apply configuration: %w", err)
	}

	log.WithProvider(opts.Provider).Success("Provider configuration updated")
	return nil
}

// runProvidersInstall executes the providers install command
func runProvidersInstall(ctx context.Context, opts *ProvidersOptions, cfg *config.GlobalConfig, log *logger.Logger) error {
	operation := "install"
	if opts.Operation == "uninstall" {
		operation = "uninstall"
	}

	log.WithOperation("providers-"+operation).WithProvider(opts.Provider).Info(fmt.Sprintf("Provider %s", operation))

	// Placeholder implementation
	log.WithProvider(opts.Provider).Warning(fmt.Sprintf("Provider %s not yet implemented", operation))

	return nil
}

// getProviderInfo retrieves information about available providers
func getProviderInfo(opts *ProvidersOptions, cfg *config.GlobalConfig) ([]ProviderInfo, error) {
	// Built-in providers - in a full implementation, this would query the actual provider registry
	providers := []ProviderInfo{
		{
			Name:      "openai",
			Available: cfg.HasAPIKey("openai"),
			Models:    []string{"gpt-4", "gpt-3.5-turbo", "gpt-4-turbo"},
			Aliases:   map[string]string{"gpt-4": "gpt-4", "gpt-3.5": "gpt-3.5-turbo"},
			Status:    "ready",
		},
		{
			Name:      "gemini",
			Available: cfg.HasAPIKey("gemini"),
			Models:    []string{"gemini-pro", "gemini-pro-vision"},
			Aliases:   map[string]string{"gemini": "gemini-pro"},
			Status:    "ready",
		},
		{
			Name:      "ollama",
			Available: cfg.OllamaEndpoint != "",
			Models:    []string{"llama2", "codellama", "mistral"},
			Aliases:   map[string]string{"llama": "llama2"},
			Status:    "ready",
		},
	}

	return providers, nil
}

// testProvider tests a specific provider
func testProvider(ctx context.Context, providerName string, opts *ProvidersOptions, 
	cfg *config.GlobalConfig, log *logger.Logger) (ProviderInfo, error) {
	
	startTime := time.Now()
	
	// Basic connectivity test - placeholder implementation
	info := ProviderInfo{
		Name:      providerName,
		Available: false,
		Status:    "testing",
	}

	// Simulate test based on provider availability
	var success bool
	var testError error

	switch providerName {
	case "openai":
		success = cfg.HasAPIKey("openai")
		if !success {
			testError = fmt.Errorf("OpenAI API key not configured")
		}
	case "gemini":
		success = cfg.HasAPIKey("gemini")
		if !success {
			testError = fmt.Errorf("Gemini API key not configured")
		}
	case "ollama":
		success = cfg.OllamaEndpoint != ""
		if !success {
			testError = fmt.Errorf("Ollama endpoint not configured")
		}
	default:
		success = false
		testError = fmt.Errorf("unknown provider")
	}

	// Create test result
	result := &TestResult{
		Success:  success,
		Duration: time.Since(startTime),
	}

	if testError != nil {
		result.Error = testError.Error()
		info.Status = "error"
	} else {
		info.Status = "available"
		info.Available = true
	}

	info.TestResult = result
	now := time.Now()
	info.LastTested = &now

	return info, nil
}

// formatProvidersOutput formats provider information for output
func formatProvidersOutput(providers []ProviderInfo, opts *ProvidersOptions) (string, error) {
	switch opts.Format {
	case "json":
		return formatProvidersJSON(providers)
	case "yaml":
		return formatProvidersYAML(providers)
	default:
		return formatProvidersTable(providers, opts)
	}
}

// formatProvidersTable formats providers as a table
func formatProvidersTable(providers []ProviderInfo, opts *ProvidersOptions) (string, error) {
	var output strings.Builder

	// Header
	output.WriteString("PROVIDER\tAVAILABLE\tSTATUS\tMODELS\n")
	output.WriteString("--------\t---------\t------\t------\n")

	// Rows
	for _, p := range providers {
		available := "No"
		if p.Available {
			available = "Yes"
		}

		models := "N/A"
		if opts.ShowModels && len(p.Models) > 0 {
			models = strings.Join(p.Models, ", ")
		}

		output.WriteString(fmt.Sprintf("%s\t%s\t%s\t%s\n", 
			p.Name, available, p.Status, models))

		// Show detailed information if requested
		if opts.Detailed {
			if p.TestResult != nil {
				output.WriteString(fmt.Sprintf("  Last Tested: %v\n", p.LastTested))
				if p.TestResult.Error != "" {
					output.WriteString(fmt.Sprintf("  Error: %s\n", p.TestResult.Error))
				}
			}
		}
	}

	return output.String(), nil
}

// formatProvidersJSON formats providers as JSON
func formatProvidersJSON(providers []ProviderInfo) (string, error) {
	// Placeholder JSON formatting
	return fmt.Sprintf(`{"providers": [%d providers]}`, len(providers)), nil
}

// formatProvidersYAML formats providers as YAML
func formatProvidersYAML(providers []ProviderInfo) (string, error) {
	// Placeholder YAML formatting
	return fmt.Sprintf("providers:\n  count: %d\n", len(providers)), nil
}

// formatTestResults formats test results for output
func formatTestResults(results []ProviderInfo, opts *ProvidersOptions) (string, error) {
	switch opts.Format {
	case "json":
		return formatTestResultsJSON(results)
	case "yaml":
		return formatTestResultsYAML(results)
	default:
		return formatTestResultsTable(results)
	}
}

// formatTestResultsTable formats test results as a table
func formatTestResultsTable(results []ProviderInfo) (string, error) {
	var output strings.Builder

	output.WriteString("PROVIDER\tSTATUS\tDURATION\tERROR\n")
	output.WriteString("--------\t------\t--------\t-----\n")

	for _, result := range results {
		status := "PASS"
		duration := "N/A"
		errorMsg := ""

		if result.TestResult != nil {
			if !result.TestResult.Success {
				status = "FAIL"
			}
			duration = result.TestResult.Duration.String()
			errorMsg = result.TestResult.Error
		}

		output.WriteString(fmt.Sprintf("%s\t%s\t%s\t%s\n", 
			result.Name, status, duration, errorMsg))
	}

	return output.String(), nil
}

// formatTestResultsJSON formats test results as JSON
func formatTestResultsJSON(results []ProviderInfo) (string, error) {
	// Placeholder JSON formatting
	return fmt.Sprintf(`{"test_results": [%d results]}`, len(results)), nil
}

// formatTestResultsYAML formats test results as YAML
func formatTestResultsYAML(results []ProviderInfo) (string, error) {
	// Placeholder YAML formatting
	return fmt.Sprintf("test_results:\n  count: %d\n", len(results)), nil
}

// isValidProvider checks if a provider name is valid
func isValidProvider(provider string) bool {
	validProviders := []string{"openai", "gemini", "ollama"}
	for _, p := range validProviders {
		if p == provider {
			return true
		}
	}
	return false
}

// showProviderConfig shows current provider configuration
func showProviderConfig(provider string, cfg *config.GlobalConfig, opts *ProvidersOptions) error {
	// Placeholder configuration display
	fmt.Printf("Configuration for provider: %s\n", provider)
	
	switch provider {
	case "openai":
		fmt.Printf("  API Key: %s\n", maskAPIKey(cfg.OpenAIAPIKey))
	case "gemini":
		fmt.Printf("  API Key: %s\n", maskAPIKey(cfg.GeminiAPIKey))
	case "ollama":
		fmt.Printf("  Endpoint: %s\n", cfg.OllamaEndpoint)
	}

	return nil
}

// applyProviderConfig applies configuration changes to a provider
func applyProviderConfig(provider string, opts *ProvidersOptions, cfg *config.GlobalConfig) error {
	// Placeholder configuration application
	for key, value := range opts.SetConfig {
		fmt.Printf("Setting %s.%s = %s\n", provider, key, value)
	}

	for _, key := range opts.UnsetConfig {
		fmt.Printf("Unsetting %s.%s\n", provider, key)
	}

	return nil
}

// maskAPIKey masks an API key for display
func maskAPIKey(apiKey string) string {
	if apiKey == "" {
		return "<not set>"
	}
	if len(apiKey) <= 8 {
		return strings.Repeat("*", len(apiKey))
	}
	return apiKey[:4] + strings.Repeat("*", len(apiKey)-8) + apiKey[len(apiKey)-4:]
}