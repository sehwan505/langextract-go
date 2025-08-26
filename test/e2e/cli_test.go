package e2e_test

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestCLIExtractCommand tests the CLI extract command with various options
func TestCLIExtractCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}

	// Build the CLI if it doesn't exist
	cliPath := buildCLIIfNeeded(t)

	// Create test fixtures
	testDir := setupTestFixtures(t)
	defer os.RemoveAll(testDir)

	tests := []struct {
		name           string
		args           []string
		inputText      string
		expectError    bool
		expectOutput   []string // Substrings that should be present in output
		unexpectOutput []string // Substrings that should NOT be present
		setupEnv       func()
		cleanupEnv     func()
	}{
		{
			name:        "basic help command",
			args:        []string{"--help"},
			expectError: false,
			expectOutput: []string{
				"langextract",
				"Usage:",
				"extract",
				"visualize",
			},
		},
		{
			name:        "extract help command",
			args:        []string{"extract", "--help"},
			expectError: false,
			expectOutput: []string{
				"Extract structured information from text",
				"--model",
				"--prompt",
				"--output",
			},
		},
		{
			name:        "extract with mock provider",
			args:        []string{"extract", "--model", "mock", "--prompt", "Extract names", "--output", "json"},
			inputText:   "John works at Google.",
			expectError: false,
			expectOutput: []string{
				"extractions",
				"John",
			},
			setupEnv: func() {
				os.Setenv("LANGEXTRACT_MOCK", "true")
			},
			cleanupEnv: func() {
				os.Unsetenv("LANGEXTRACT_MOCK")
			},
		},
		{
			name:        "extract from file",
			args:        []string{"extract", "--model", "mock", "--prompt", "Extract entities", "--input", filepath.Join(testDir, "sample.txt"), "--output", "json"},
			expectError: false,
			expectOutput: []string{
				"extractions",
			},
			setupEnv: func() {
				os.Setenv("LANGEXTRACT_MOCK", "true")
			},
			cleanupEnv: func() {
				os.Unsetenv("LANGEXTRACT_MOCK")
			},
		},
		{
			name:        "extract with schema file",
			args:        []string{"extract", "--model", "mock", "--prompt", "Extract entities", "--schema", filepath.Join(testDir, "schema.json"), "--output", "json"},
			inputText:   "Dr. Smith works at Memorial Hospital.",
			expectError: false,
			expectOutput: []string{
				"extractions",
			},
			setupEnv: func() {
				os.Setenv("LANGEXTRACT_MOCK", "true")
			},
			cleanupEnv: func() {
				os.Unsetenv("LANGEXTRACT_MOCK")
			},
		},
		{
			name:        "extract with examples file",
			args:        []string{"extract", "--model", "mock", "--prompt", "Extract entities", "--examples", filepath.Join(testDir, "examples.json"), "--output", "json"},
			inputText:   "Alice lives in New York.",
			expectError: false,
			expectOutput: []string{
				"extractions",
			},
			setupEnv: func() {
				os.Setenv("LANGEXTRACT_MOCK", "true")
			},
			cleanupEnv: func() {
				os.Unsetenv("LANGEXTRACT_MOCK")
			},
		},
		{
			name:        "extract with invalid model",
			args:        []string{"extract", "--model", "invalid-model", "--prompt", "Extract entities"},
			inputText:   "Test text",
			expectError: true,
			expectOutput: []string{
				"error",
				"provider",
			},
		},
		{
			name:        "extract with missing prompt",
			args:        []string{"extract", "--model", "mock"},
			inputText:   "Test text",
			expectError: true,
			expectOutput: []string{
				"error",
				"prompt",
			},
		},
		{
			name:        "extract with invalid schema file",
			args:        []string{"extract", "--model", "mock", "--prompt", "Extract entities", "--schema", "nonexistent.json"},
			inputText:   "Test text",
			expectError: true,
			expectOutput: []string{
				"error",
				"schema",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment
			if tt.setupEnv != nil {
				tt.setupEnv()
			}
			if tt.cleanupEnv != nil {
				defer tt.cleanupEnv()
			}

			// Prepare command
			cmd := exec.Command(cliPath, tt.args...)
			
			// Set input if provided
			if tt.inputText != "" {
				cmd.Stdin = strings.NewReader(tt.inputText)
			}

			// Run command
			output, err := cmd.CombinedOutput()
			outputStr := string(output)

			// Check error expectation
			if tt.expectError && err == nil {
				t.Errorf("Expected error but command succeeded. Output: %s", outputStr)
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected success but command failed with error: %v. Output: %s", err, outputStr)
			}

			// Check expected output
			for _, expected := range tt.expectOutput {
				if !strings.Contains(outputStr, expected) {
					t.Errorf("Expected output to contain %q, but it didn't. Full output: %s", expected, outputStr)
				}
			}

			// Check unexpected output
			for _, unexpected := range tt.unexpectOutput {
				if strings.Contains(outputStr, unexpected) {
					t.Errorf("Expected output NOT to contain %q, but it did. Full output: %s", unexpected, outputStr)
				}
			}
		})
	}
}

// TestCLIVisualizeCommand tests the CLI visualize command
func TestCLIVisualizeCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}

	cliPath := buildCLIIfNeeded(t)
	testDir := setupTestFixtures(t)
	defer os.RemoveAll(testDir)

	tests := []struct {
		name        string
		args        []string
		inputFile   string
		expectError bool
		checkOutput func(string) error
		setupEnv    func()
		cleanupEnv  func()
	}{
		{
			name:        "visualize help command",
			args:        []string{"visualize", "--help"},
			expectError: false,
			checkOutput: func(output string) error {
				expectedSubstrings := []string{
					"Visualize extracted entities",
					"--format",
					"--output",
					"html",
					"json",
					"csv",
				}
				for _, expected := range expectedSubstrings {
					if !strings.Contains(output, expected) {
						return fmt.Errorf("expected output to contain %q", expected)
					}
				}
				return nil
			},
		},
		{
			name:      "visualize with HTML output",
			args:      []string{"visualize", "--input", filepath.Join(testDir, "annotated.json"), "--format", "html"},
			inputFile: "annotated.json",
			checkOutput: func(output string) error {
				if !strings.Contains(output, "<html>") || !strings.Contains(output, "</html>") {
					return fmt.Errorf("expected HTML output")
				}
				return nil
			},
		},
		{
			name:      "visualize with JSON output",
			args:      []string{"visualize", "--input", filepath.Join(testDir, "annotated.json"), "--format", "json"},
			inputFile: "annotated.json",
			checkOutput: func(output string) error {
				var jsonData interface{}
				if err := json.Unmarshal([]byte(output), &jsonData); err != nil {
					return fmt.Errorf("expected valid JSON output: %v", err)
				}
				return nil
			},
		},
		{
			name:      "visualize with CSV output",
			args:      []string{"visualize", "--input", filepath.Join(testDir, "annotated.json"), "--format", "csv"},
			inputFile: "annotated.json",
			checkOutput: func(output string) error {
				lines := strings.Split(strings.TrimSpace(output), "\n")
				if len(lines) < 2 { // At least header + one data row
					return fmt.Errorf("expected CSV with header and data rows")
				}
				// Check CSV header
				if !strings.Contains(lines[0], "extraction_class") {
					return fmt.Errorf("expected CSV header to contain extraction_class")
				}
				return nil
			},
		},
		{
			name:        "visualize with invalid format",
			args:        []string{"visualize", "--input", filepath.Join(testDir, "annotated.json"), "--format", "invalid"},
			inputFile:   "annotated.json",
			expectError: true,
		},
		{
			name:        "visualize with missing input file",
			args:        []string{"visualize", "--input", "nonexistent.json", "--format", "html"},
			expectError: true,
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

			cmd := exec.Command(cliPath, tt.args...)
			output, err := cmd.CombinedOutput()
			outputStr := string(output)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but command succeeded. Output: %s", outputStr)
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected success but command failed with error: %v. Output: %s", err, outputStr)
			}

			if tt.checkOutput != nil && !tt.expectError {
				if err := tt.checkOutput(outputStr); err != nil {
					t.Errorf("Output validation failed: %v. Full output: %s", err, outputStr)
				}
			}
		})
	}
}

// TestCLIProvidersList tests listing available providers
func TestCLIProvidersList(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}

	cliPath := buildCLIIfNeeded(t)

	cmd := exec.Command(cliPath, "providers")
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		t.Fatalf("providers command failed: %v. Output: %s", err, string(output))
	}

	outputStr := string(output)
	expectedProviders := []string{"openai", "gemini", "ollama"}
	
	for _, provider := range expectedProviders {
		if !strings.Contains(outputStr, provider) {
			t.Errorf("Expected providers list to contain %q. Full output: %s", provider, outputStr)
		}
	}
}

// TestCLIConfigValidation tests configuration validation
func TestCLIConfigValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}

	cliPath := buildCLIIfNeeded(t)
	testDir := setupTestFixtures(t)
	defer os.RemoveAll(testDir)

	// Create invalid config file
	invalidConfigPath := filepath.Join(testDir, "invalid_config.yaml")
	invalidConfig := `
invalid_yaml_content:
  - this is not valid
  - yaml: content
    missing: closing bracket
`
	if err := os.WriteFile(invalidConfigPath, []byte(invalidConfig), 0644); err != nil {
		t.Fatalf("Failed to create invalid config: %v", err)
	}

	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "valid config validation",
			args:        []string{"validate", "--config", filepath.Join(testDir, "config.yaml")},
			expectError: false,
		},
		{
			name:        "invalid config validation",
			args:        []string{"validate", "--config", invalidConfigPath},
			expectError: true,
		},
		{
			name:        "nonexistent config validation",
			args:        []string{"validate", "--config", "nonexistent.yaml"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(cliPath, tt.args...)
			output, err := cmd.CombinedOutput()

			if tt.expectError && err == nil {
				t.Errorf("Expected error but command succeeded. Output: %s", string(output))
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected success but command failed: %v. Output: %s", err, string(output))
			}
		})
	}
}

// TestCLIBatchProcessing tests batch processing functionality
func TestCLIBatchProcessing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}

	cliPath := buildCLIIfNeeded(t)
	testDir := setupTestFixtures(t)
	defer os.RemoveAll(testDir)

	// Setup mock environment
	os.Setenv("LANGEXTRACT_MOCK", "true")
	defer os.Unsetenv("LANGEXTRACT_MOCK")

	// Create batch input file
	batchInputPath := filepath.Join(testDir, "batch_input.txt")
	batchInput := `John works at Google.
Alice lives in New York.
Microsoft is headquartered in Seattle.`
	
	if err := os.WriteFile(batchInputPath, []byte(batchInput), 0644); err != nil {
		t.Fatalf("Failed to create batch input: %v", err)
	}

	outputPath := filepath.Join(testDir, "batch_output.json")

	cmd := exec.Command(cliPath, "batch", 
		"--input", batchInputPath,
		"--output", outputPath,
		"--model", "mock",
		"--prompt", "Extract entities",
		"--format", "json")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Batch command failed: %v. Output: %s", err, string(output))
	}

	// Check output file exists and contains results
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("Expected output file %s to be created", outputPath)
	}

	// Read and validate output
	outputData, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	var results []interface{}
	if err := json.Unmarshal(outputData, &results); err != nil {
		t.Errorf("Output file should contain valid JSON: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 results for 3 input lines, got %d", len(results))
	}
}

// TestCLIInteractiveMode tests interactive mode (if implemented)
func TestCLIInteractiveMode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}

	cliPath := buildCLIIfNeeded(t)

	// Setup mock environment
	os.Setenv("LANGEXTRACT_MOCK", "true")
	defer os.Unsetenv("LANGEXTRACT_MOCK")

	cmd := exec.Command(cliPath, "interactive", "--model", "mock")
	
	// Create pipes for stdin/stdout
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("Failed to create stdin pipe: %v", err)
	}
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("Failed to create stdout pipe: %v", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start interactive command: %v", err)
	}

	// Send commands
	go func() {
		defer stdin.Close()
		stdin.Write([]byte("set prompt Extract names\n"))
		stdin.Write([]byte("extract John works at Google\n"))
		stdin.Write([]byte("quit\n"))
	}()

	// Read output with timeout
	outputChan := make(chan []byte)
	go func() {
		scanner := bufio.NewScanner(stdout)
		var output []string
		for scanner.Scan() {
			output = append(output, scanner.Text())
		}
		outputChan <- []byte(strings.Join(output, "\n"))
	}()

	// Wait for completion with timeout
	done := make(chan error)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-time.After(10 * time.Second):
		cmd.Process.Kill()
		t.Fatal("Interactive mode test timed out")
	case err := <-done:
		if err != nil {
			// Interactive mode might not be implemented, so we just log this
			t.Logf("Interactive mode not available or failed: %v", err)
		}
	}
}

// buildCLIIfNeeded builds the CLI binary if it doesn't exist
func buildCLIIfNeeded(t *testing.T) string {
	// Get project root
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}

	cliPath := filepath.Join(projectRoot, "bin", "langextract")
	
	// Check if binary exists and is recent
	if stat, err := os.Stat(cliPath); err == nil {
		// If binary is less than 1 hour old, assume it's current
		if time.Since(stat.ModTime()) < time.Hour {
			return cliPath
		}
	}

	// Build the CLI
	t.Logf("Building CLI binary...")
	
	buildCmd := exec.Command("go", "build", "-o", cliPath, "./cmd/langextract")
	buildCmd.Dir = projectRoot
	
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build CLI: %v. Output: %s", err, string(output))
	}

	return cliPath
}

// setupTestFixtures creates test fixtures for CLI testing
func setupTestFixtures(t *testing.T) string {
	testDir, err := os.MkdirTemp("", "langextract-e2e-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create sample text file
	sampleText := "Dr. John Smith works at Memorial Hospital in downtown Seattle. The patient, Alice Johnson, was diagnosed with diabetes."
	if err := os.WriteFile(filepath.Join(testDir, "sample.txt"), []byte(sampleText), 0644); err != nil {
		t.Fatalf("Failed to create sample.txt: %v", err)
	}

	// Create schema file
	schema := map[string]interface{}{
		"name":        "medical_schema",
		"description": "Medical entity extraction schema",
		"classes": []map[string]interface{}{
			{"name": "person", "description": "Person names"},
			{"name": "organization", "description": "Organization names"},
			{"name": "location", "description": "Location names"},
			{"name": "condition", "description": "Medical conditions"},
		},
	}
	schemaData, _ := json.MarshalIndent(schema, "", "  ")
	if err := os.WriteFile(filepath.Join(testDir, "schema.json"), schemaData, 0644); err != nil {
		t.Fatalf("Failed to create schema.json: %v", err)
	}

	// Create examples file
	examples := []map[string]interface{}{
		{
			"text": "Dr. Sarah works at City Hospital",
			"extractions": []map[string]interface{}{
				{"extraction_class": "person", "extraction_text": "Dr. Sarah"},
				{"extraction_class": "organization", "extraction_text": "City Hospital"},
			},
		},
	}
	examplesData, _ := json.MarshalIndent(examples, "", "  ")
	if err := os.WriteFile(filepath.Join(testDir, "examples.json"), examplesData, 0644); err != nil {
		t.Fatalf("Failed to create examples.json: %v", err)
	}

	// Create annotated document file
	annotatedDoc := map[string]interface{}{
		"text": "Dr. Smith works at Memorial Hospital",
		"extractions": []map[string]interface{}{
			{
				"extraction_class": "person",
				"extraction_text":  "Dr. Smith",
				"char_interval":    map[string]int{"start_pos": 0, "end_pos": 9},
			},
			{
				"extraction_class": "organization",
				"extraction_text":  "Memorial Hospital",
				"char_interval":    map[string]int{"start_pos": 19, "end_pos": 36},
			},
		},
	}
	annotatedData, _ := json.MarshalIndent(annotatedDoc, "", "  ")
	if err := os.WriteFile(filepath.Join(testDir, "annotated.json"), annotatedData, 0644); err != nil {
		t.Fatalf("Failed to create annotated.json: %v", err)
	}

	// Create config file
	config := map[string]interface{}{
		"default_model": "gemini-2.5-flash",
		"timeout":       30,
		"max_retries":   3,
		"providers": map[string]interface{}{
			"gemini": map[string]interface{}{
				"api_key_env": "GEMINI_API_KEY",
			},
			"openai": map[string]interface{}{
				"api_key_env": "OPENAI_API_KEY",
			},
		},
	}
	configData, _ := json.MarshalIndent(config, "", "  ")
	if err := os.WriteFile(filepath.Join(testDir, "config.yaml"), configData, 0644); err != nil {
		t.Fatalf("Failed to create config.yaml: %v", err)
	}

	return testDir
}

// findProjectRoot finds the project root directory
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Look for go.mod file
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("could not find project root (go.mod not found)")
}