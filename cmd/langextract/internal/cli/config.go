package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/sehwan505/langextract-go/cmd/langextract/internal/config"
	"github.com/sehwan505/langextract-go/cmd/langextract/internal/logger"
)

// ConfigOptions contains all options for the config command
type ConfigOptions struct {
	// Operations
	Operation string // get, set, unset, list, init, edit, validate

	// Target configuration
	Key   string // Configuration key to operate on
	Value string // Value to set

	// Scope
	Global bool // Operate on global configuration
	Local  bool // Operate on local (project) configuration
	User   bool // Operate on user configuration

	// File operations
	ConfigFile   string // Specific configuration file
	CreateConfig bool   // Create configuration file if it doesn't exist
	BackupConfig bool   // Create backup before modifying

	// Output options
	Output     string // Output file
	Format     string // Output format: yaml, json, table
	ShowAll    bool   // Show all configuration values
	ShowOrigin bool   // Show where each value is defined

	// Validation
	Validate bool // Validate configuration after changes
	DryRun   bool // Show what would be changed without applying
}

// NewConfigCommand creates the config command with subcommands
func NewConfigCommand(cfg *config.GlobalConfig, log *logger.Logger) *cobra.Command {
	opts := &ConfigOptions{
		Format:       "yaml",
		ShowAll:      false,
		ShowOrigin:   false,
		Global:       true,
		Local:        false,
		User:         false,
		CreateConfig: false,
		BackupConfig: true,
		Validate:     true,
		DryRun:       false,
	}

	cmd := &cobra.Command{
		Use:   "config [command]",
		Short: "Manage configuration settings",
		Long: `Manage configuration settings for langextract. Configuration can be stored
at different levels: global (system-wide), user (per-user), or local (per-project).

Configuration precedence (highest to lowest):
1. Command line flags
2. Environment variables
3. Local configuration file (./langextract.yaml)
4. User configuration file (~/.config/langextract/langextract.yaml)
5. Global configuration file (/etc/langextract/langextract.yaml)
6. Default values

Examples:
  # Show current configuration
  langextract config list
  
  # Get a specific value
  langextract config get log_level
  
  # Set a configuration value
  langextract config set log_level debug
  
  # Initialize configuration with defaults
  langextract config init
  
  # Edit configuration interactively
  langextract config edit
  
  # Validate current configuration
  langextract config validate`,
	}

	// Add subcommands
	cmd.AddCommand(NewConfigListCommand(cfg, log, opts))
	cmd.AddCommand(NewConfigGetCommand(cfg, log, opts))
	cmd.AddCommand(NewConfigSetCommand(cfg, log, opts))
	cmd.AddCommand(NewConfigUnsetCommand(cfg, log, opts))
	cmd.AddCommand(NewConfigInitCommand(cfg, log, opts))
	cmd.AddCommand(NewConfigEditCommand(cfg, log, opts))
	cmd.AddCommand(NewConfigValidateCommand(cfg, log, opts))

	return cmd
}

// NewConfigListCommand creates the config list subcommand
func NewConfigListCommand(cfg *config.GlobalConfig, log *logger.Logger, opts *ConfigOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [flags]",
		Short: "List all configuration settings",
		Long: `List all configuration settings with their current values and sources.
Shows where each configuration value is defined (command line, environment, file, or default).

Examples:
  # List all configuration
  langextract config list
  
  # Show configuration with origins
  langextract config list --show-origin
  
  # Export configuration to file
  langextract config list --format json --output config.json
  
  # Show only user-level configuration
  langextract config list --user`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigList(cmd.Context(), opts, cfg, log)
		},
	}

	// Output flags
	cmd.Flags().StringVarP(&opts.Output, "output", "o", "", "Output file")
	cmd.Flags().StringVar(&opts.Format, "format", opts.Format, "Output format (yaml, json, table)")
	cmd.Flags().BoolVar(&opts.ShowAll, "all", opts.ShowAll, "Show all configuration values including defaults")
	cmd.Flags().BoolVar(&opts.ShowOrigin, "show-origin", opts.ShowOrigin, "Show where each value is defined")

	// Scope flags
	cmd.Flags().BoolVar(&opts.Global, "global", false, "Show global configuration only")
	cmd.Flags().BoolVar(&opts.User, "user", false, "Show user configuration only")
	cmd.Flags().BoolVar(&opts.Local, "local", false, "Show local configuration only")

	return cmd
}

// NewConfigGetCommand creates the config get subcommand
func NewConfigGetCommand(cfg *config.GlobalConfig, log *logger.Logger, opts *ConfigOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [key]",
		Short: "Get a configuration value",
		Long: `Get the value of a specific configuration key. Shows the effective value
after applying all configuration sources in order of precedence.

Examples:
  # Get log level
  langextract config get log_level
  
  # Get default provider
  langextract config get default_provider
  
  # Get with origin information
  langextract config get log_level --show-origin`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Key = args[0]
			return runConfigGet(cmd.Context(), opts, cfg, log)
		},
	}

	// Output flags
	cmd.Flags().BoolVar(&opts.ShowOrigin, "show-origin", opts.ShowOrigin, "Show where the value is defined")

	return cmd
}

// NewConfigSetCommand creates the config set subcommand
func NewConfigSetCommand(cfg *config.GlobalConfig, log *logger.Logger, opts *ConfigOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set [key] [value]",
		Short: "Set a configuration value",
		Long: `Set the value of a specific configuration key. By default, values are set
in the user configuration file. Use --global or --local to change the scope.

Examples:
  # Set log level
  langextract config set log_level debug
  
  # Set provider with validation
  langextract config set default_provider openai --validate
  
  # Set in global configuration
  langextract config set log_level info --global
  
  # Dry run to see what would be changed
  langextract config set concurrency 8 --dry-run`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Key = args[0]
			opts.Value = args[1]
			return runConfigSet(cmd.Context(), opts, cfg, log)
		},
	}

	// Scope flags
	cmd.Flags().BoolVar(&opts.Global, "global", false, "Set in global configuration")
	cmd.Flags().BoolVar(&opts.User, "user", true, "Set in user configuration (default)")
	cmd.Flags().BoolVar(&opts.Local, "local", false, "Set in local configuration")

	// Operation flags
	cmd.Flags().StringVar(&opts.ConfigFile, "config-file", "", "Specific configuration file to modify")
	cmd.Flags().BoolVar(&opts.CreateConfig, "create", opts.CreateConfig, "Create configuration file if it doesn't exist")
	cmd.Flags().BoolVar(&opts.BackupConfig, "backup", opts.BackupConfig, "Create backup before modifying")
	cmd.Flags().BoolVar(&opts.Validate, "validate", opts.Validate, "Validate configuration after changes")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", opts.DryRun, "Show what would be changed without applying")

	return cmd
}

// NewConfigUnsetCommand creates the config unset subcommand
func NewConfigUnsetCommand(cfg *config.GlobalConfig, log *logger.Logger, opts *ConfigOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unset [key]",
		Short: "Unset a configuration value",
		Long: `Remove a configuration key from the configuration file. The key will
revert to its default value or the value from a lower-priority configuration source.

Examples:
  # Unset log level (reverts to default)
  langextract config unset log_level
  
  # Unset from global configuration
  langextract config unset default_provider --global
  
  # Dry run to see what would be changed
  langextract config unset concurrency --dry-run`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Key = args[0]
			opts.Operation = "unset"
			return runConfigSet(cmd.Context(), opts, cfg, log)
		},
	}

	// Scope flags
	cmd.Flags().BoolVar(&opts.Global, "global", false, "Unset from global configuration")
	cmd.Flags().BoolVar(&opts.User, "user", true, "Unset from user configuration (default)")
	cmd.Flags().BoolVar(&opts.Local, "local", false, "Unset from local configuration")

	// Operation flags
	cmd.Flags().StringVar(&opts.ConfigFile, "config-file", "", "Specific configuration file to modify")
	cmd.Flags().BoolVar(&opts.BackupConfig, "backup", opts.BackupConfig, "Create backup before modifying")
	cmd.Flags().BoolVar(&opts.Validate, "validate", opts.Validate, "Validate configuration after changes")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", opts.DryRun, "Show what would be changed without applying")

	return cmd
}

// NewConfigInitCommand creates the config init subcommand
func NewConfigInitCommand(cfg *config.GlobalConfig, log *logger.Logger, opts *ConfigOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [flags]",
		Short: "Initialize configuration file with defaults",
		Long: `Initialize a configuration file with default values and helpful comments.
Creates a template configuration file that can be customized as needed.

Examples:
  # Initialize user configuration
  langextract config init
  
  # Initialize local configuration
  langextract config init --local
  
  # Initialize with specific file
  langextract config init --config-file ./my-config.yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigInit(cmd.Context(), opts, cfg, log)
		},
	}

	// Scope flags
	cmd.Flags().BoolVar(&opts.Global, "global", false, "Initialize global configuration")
	cmd.Flags().BoolVar(&opts.User, "user", true, "Initialize user configuration (default)")
	cmd.Flags().BoolVar(&opts.Local, "local", false, "Initialize local configuration")

	// Operation flags
	cmd.Flags().StringVar(&opts.ConfigFile, "config-file", "", "Specific configuration file to create")
	cmd.Flags().BoolVar(&opts.CreateConfig, "force", false, "Overwrite existing configuration file")

	return cmd
}

// NewConfigEditCommand creates the config edit subcommand
func NewConfigEditCommand(cfg *config.GlobalConfig, log *logger.Logger, opts *ConfigOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit [flags]",
		Short: "Edit configuration file interactively",
		Long: `Open the configuration file in an editor for interactive editing.
Uses the EDITOR environment variable or falls back to default editors.

Examples:
  # Edit user configuration
  langextract config edit
  
  # Edit global configuration
  langextract config edit --global
  
  # Edit specific file
  langextract config edit --config-file ./custom-config.yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigEdit(cmd.Context(), opts, cfg, log)
		},
	}

	// Scope flags
	cmd.Flags().BoolVar(&opts.Global, "global", false, "Edit global configuration")
	cmd.Flags().BoolVar(&opts.User, "user", true, "Edit user configuration (default)")
	cmd.Flags().BoolVar(&opts.Local, "local", false, "Edit local configuration")

	// Operation flags
	cmd.Flags().StringVar(&opts.ConfigFile, "config-file", "", "Specific configuration file to edit")
	cmd.Flags().BoolVar(&opts.CreateConfig, "create", true, "Create configuration file if it doesn't exist")
	cmd.Flags().BoolVar(&opts.BackupConfig, "backup", opts.BackupConfig, "Create backup before editing")
	cmd.Flags().BoolVar(&opts.Validate, "validate", opts.Validate, "Validate configuration after editing")

	return cmd
}

// NewConfigValidateCommand creates the config validate subcommand
func NewConfigValidateCommand(cfg *config.GlobalConfig, log *logger.Logger, opts *ConfigOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate [flags]",
		Short: "Validate configuration settings",
		Long: `Validate the current configuration for correctness and completeness.
Checks for invalid values, missing required settings, and potential conflicts.

Examples:
  # Validate current configuration
  langextract config validate
  
  # Validate specific file
  langextract config validate --config-file ./my-config.yaml
  
  # Validate with detailed output
  langextract config validate --format table --output validation-report.txt`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigValidate(cmd.Context(), opts, cfg, log)
		},
	}

	// Operation flags
	cmd.Flags().StringVar(&opts.ConfigFile, "config-file", "", "Specific configuration file to validate")

	// Output flags
	cmd.Flags().StringVarP(&opts.Output, "output", "o", "", "Output file for validation report")
	cmd.Flags().StringVar(&opts.Format, "format", "table", "Output format (table, json, yaml)")

	return cmd
}

// runConfigList executes the config list command
func runConfigList(ctx context.Context, opts *ConfigOptions, cfg *config.GlobalConfig, log *logger.Logger) error {
	log.WithOperation("config-list").Info("Listing configuration")

	// Generate configuration output
	output, err := formatConfiguration(cfg, opts)
	if err != nil {
		return fmt.Errorf("failed to format configuration: %w", err)
	}

	// Write output
	if opts.Output != "" {
		if err := os.WriteFile(opts.Output, []byte(output), 0644); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
	} else {
		fmt.Print(output)
	}

	return nil
}

// runConfigGet executes the config get command
func runConfigGet(ctx context.Context, opts *ConfigOptions, cfg *config.GlobalConfig, log *logger.Logger) error {
	log.WithOperation("config-get").Info("Getting configuration value")

	value, origin, err := getConfigValue(cfg, opts.Key)
	if err != nil {
		return fmt.Errorf("failed to get configuration value: %w", err)
	}

	if opts.ShowOrigin {
		fmt.Printf("%s (from %s)\n", value, origin)
	} else {
		fmt.Println(value)
	}

	return nil
}

// runConfigSet executes the config set/unset commands
func runConfigSet(ctx context.Context, opts *ConfigOptions, cfg *config.GlobalConfig, log *logger.Logger) error {
	operation := "set"
	if opts.Operation == "unset" {
		operation = "unset"
	}

	log.WithOperation("config-"+operation).Info(fmt.Sprintf("Configuration %s", operation))

	// Determine configuration file path
	configFile, err := determineConfigFile(opts)
	if err != nil {
		return fmt.Errorf("failed to determine configuration file: %w", err)
	}

	// Show dry run if requested
	if opts.DryRun {
		if opts.Operation == "unset" {
			fmt.Printf("Would unset %s in %s\n", opts.Key, configFile)
		} else {
			fmt.Printf("Would set %s = %s in %s\n", opts.Key, opts.Value, configFile)
		}
		return nil
	}

	// Create backup if requested
	if opts.BackupConfig {
		if err := createConfigBackup(configFile); err != nil {
			log.WithError(err).Warning("Failed to create backup")
		}
	}

	// Apply configuration change
	if err := applyConfigChange(configFile, opts); err != nil {
		return fmt.Errorf("failed to apply configuration change: %w", err)
	}

	// Validate if requested
	if opts.Validate {
		if err := validateConfigFile(configFile); err != nil {
			return fmt.Errorf("configuration validation failed: %w", err)
		}
	}

	log.Success(fmt.Sprintf("Configuration %s completed", operation))
	return nil
}

// runConfigInit executes the config init command
func runConfigInit(ctx context.Context, opts *ConfigOptions, cfg *config.GlobalConfig, log *logger.Logger) error {
	log.WithOperation("config-init").Info("Initializing configuration")

	// Determine configuration file path
	configFile, err := determineConfigFile(opts)
	if err != nil {
		return fmt.Errorf("failed to determine configuration file: %w", err)
	}

	// Check if file already exists
	if _, err := os.Stat(configFile); err == nil && !opts.CreateConfig {
		return fmt.Errorf("configuration file already exists: %s (use --force to overwrite)", configFile)
	}

	// Create directory if needed
	if dir := filepath.Dir(configFile); dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create configuration directory: %w", err)
		}
	}

	// Create configuration template
	if err := config.CreateConfigTemplate(configFile); err != nil {
		return fmt.Errorf("failed to create configuration template: %w", err)
	}

	log.WithFile(configFile).Success("Configuration file initialized")
	return nil
}

// runConfigEdit executes the config edit command
func runConfigEdit(ctx context.Context, opts *ConfigOptions, cfg *config.GlobalConfig, log *logger.Logger) error {
	log.WithOperation("config-edit").Info("Editing configuration")

	// Determine configuration file path
	configFile, err := determineConfigFile(opts)
	if err != nil {
		return fmt.Errorf("failed to determine configuration file: %w", err)
	}

	// Create file if it doesn't exist and create is enabled
	if _, err := os.Stat(configFile); os.IsNotExist(err) && opts.CreateConfig {
		if err := config.CreateConfigTemplate(configFile); err != nil {
			return fmt.Errorf("failed to create configuration file: %w", err)
		}
	}

	// Create backup if requested
	if opts.BackupConfig {
		if err := createConfigBackup(configFile); err != nil {
			log.WithError(err).Warning("Failed to create backup")
		}
	}

	// Open editor - placeholder implementation
	log.WithFile(configFile).Info("Opening configuration file in editor")
	fmt.Printf("Would open %s in editor (not implemented)\n", configFile)

	// Validate after editing if requested
	if opts.Validate {
		if err := validateConfigFile(configFile); err != nil {
			return fmt.Errorf("configuration validation failed: %w", err)
		}
	}

	return nil
}

// runConfigValidate executes the config validate command
func runConfigValidate(ctx context.Context, opts *ConfigOptions, cfg *config.GlobalConfig, log *logger.Logger) error {
	log.WithOperation("config-validate").Info("Validating configuration")

	// Validate current configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Generate validation report
	report := generateConfigValidationReport(cfg)

	// Write output
	if opts.Output != "" {
		if err := os.WriteFile(opts.Output, []byte(report), 0644); err != nil {
			return fmt.Errorf("failed to write validation report: %w", err)
		}
	} else {
		fmt.Print(report)
	}

	log.Success("Configuration validation completed")
	return nil
}

// formatConfiguration formats configuration for output
func formatConfiguration(cfg *config.GlobalConfig, opts *ConfigOptions) (string, error) {
	switch opts.Format {
	case "json":
		return formatConfigJSON(cfg)
	case "table":
		return formatConfigTable(cfg)
	default:
		return formatConfigYAML(cfg)
	}
}

// formatConfigYAML formats configuration as YAML
func formatConfigYAML(cfg *config.GlobalConfig) (string, error) {
	// Placeholder YAML formatting
	var output strings.Builder
	output.WriteString("# LangExtract Configuration\n\n")
	output.WriteString(fmt.Sprintf("log_level: %s\n", cfg.LogLevel))
	output.WriteString(fmt.Sprintf("log_format: %s\n", cfg.LogFormat))
	output.WriteString(fmt.Sprintf("default_provider: %s\n", cfg.DefaultProvider))
	output.WriteString(fmt.Sprintf("concurrency: %d\n", cfg.Concurrency))
	output.WriteString(fmt.Sprintf("cache_enabled: %t\n", cfg.CacheEnabled))
	
	return output.String(), nil
}

// formatConfigJSON formats configuration as JSON
func formatConfigJSON(cfg *config.GlobalConfig) (string, error) {
	// Placeholder JSON formatting
	return fmt.Sprintf(`{"log_level": "%s", "default_provider": "%s", "concurrency": %d}`,
		cfg.LogLevel, cfg.DefaultProvider, cfg.Concurrency), nil
}

// formatConfigTable formats configuration as a table
func formatConfigTable(cfg *config.GlobalConfig) (string, error) {
	var output strings.Builder
	
	output.WriteString("KEY\tVALUE\tTYPE\n")
	output.WriteString("---\t-----\t----\n")
	output.WriteString(fmt.Sprintf("log_level\t%s\tstring\n", cfg.LogLevel))
	output.WriteString(fmt.Sprintf("log_format\t%s\tstring\n", cfg.LogFormat))
	output.WriteString(fmt.Sprintf("default_provider\t%s\tstring\n", cfg.DefaultProvider))
	output.WriteString(fmt.Sprintf("concurrency\t%d\tint\n", cfg.Concurrency))
	output.WriteString(fmt.Sprintf("cache_enabled\t%t\tbool\n", cfg.CacheEnabled))

	return output.String(), nil
}

// getConfigValue gets a configuration value by key
func getConfigValue(cfg *config.GlobalConfig, key string) (string, string, error) {
	// Map configuration keys to values
	switch strings.ToLower(key) {
	case "log_level":
		return cfg.LogLevel, "configuration", nil
	case "log_format":
		return cfg.LogFormat, "configuration", nil
	case "default_provider":
		return cfg.DefaultProvider, "configuration", nil
	case "concurrency":
		return fmt.Sprintf("%d", cfg.Concurrency), "configuration", nil
	case "cache_enabled":
		return fmt.Sprintf("%t", cfg.CacheEnabled), "configuration", nil
	default:
		return "", "", fmt.Errorf("unknown configuration key: %s", key)
	}
}

// determineConfigFile determines which configuration file to use
func determineConfigFile(opts *ConfigOptions) (string, error) {
	if opts.ConfigFile != "" {
		return opts.ConfigFile, nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	if opts.Global {
		return "/etc/langextract/langextract.yaml", nil
	} else if opts.Local {
		return "./langextract.yaml", nil
	} else {
		// User configuration (default)
		return filepath.Join(homeDir, ".config", "langextract", "langextract.yaml"), nil
	}
}

// createConfigBackup creates a backup of the configuration file
func createConfigBackup(configFile string) error {
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil // No backup needed for non-existent file
	}

	backupFile := configFile + ".backup"
	
	input, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read configuration file: %w", err)
	}

	if err := os.WriteFile(backupFile, input, 0644); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	return nil
}

// applyConfigChange applies a configuration change to a file
func applyConfigChange(configFile string, opts *ConfigOptions) error {
	// Placeholder implementation
	if opts.Operation == "unset" {
		fmt.Printf("Unsetting %s in %s\n", opts.Key, configFile)
	} else {
		fmt.Printf("Setting %s = %s in %s\n", opts.Key, opts.Value, configFile)
	}
	return nil
}

// validateConfigFile validates a configuration file
func validateConfigFile(configFile string) error {
	// Placeholder validation
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return fmt.Errorf("configuration file does not exist: %s", configFile)
	}
	return nil
}

// generateConfigValidationReport generates a validation report for configuration
func generateConfigValidationReport(cfg *config.GlobalConfig) string {
	var report strings.Builder
	
	report.WriteString("Configuration Validation Report\n")
	report.WriteString("===============================\n\n")
	report.WriteString("Status: VALID\n")
	report.WriteString("No validation errors found.\n\n")
	
	report.WriteString("Configuration Summary:\n")
	report.WriteString(fmt.Sprintf("  Log Level: %s\n", cfg.LogLevel))
	report.WriteString(fmt.Sprintf("  Default Provider: %s\n", cfg.DefaultProvider))
	report.WriteString(fmt.Sprintf("  Concurrency: %d\n", cfg.Concurrency))
	
	return report.String()
}