package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OllamaProvider implements the BaseLanguageModel interface for Ollama local models.
type OllamaProvider struct {
	config      *ModelConfig
	baseURL     string
	client      *http.Client
	schema      any
	fenceOutput bool
}

// OllamaRequest represents an Ollama API request.
type OllamaRequest struct {
	Model    string                 `json:"model"`
	Prompt   string                 `json:"prompt"`
	Stream   bool                   `json:"stream"`
	Options  map[string]interface{} `json:"options,omitempty"`
	Template string                 `json:"template,omitempty"`
}

// OllamaResponse represents an Ollama API response.
type OllamaResponse struct {
	Model              string `json:"model"`
	CreatedAt          string `json:"created_at"`
	Response           string `json:"response"`
	Done               bool   `json:"done"`
	DoneReason         string `json:"done_reason,omitempty"`
	Context            []int  `json:"context,omitempty"`
	TotalDuration      int64  `json:"total_duration,omitempty"`
	LoadDuration       int64  `json:"load_duration,omitempty"`
	PromptEvalCount    int    `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64  `json:"prompt_eval_duration,omitempty"`
	EvalCount          int    `json:"eval_count,omitempty"`
	EvalDuration       int64  `json:"eval_duration,omitempty"`
}

// OllamaTagsResponse represents the response from /api/tags endpoint.
type OllamaTagsResponse struct {
	Models []OllamaModel `json:"models"`
}

// OllamaModel represents a model in the tags response.
type OllamaModel struct {
	Name       string    `json:"name"`
	Size       int64     `json:"size"`
	Digest     string    `json:"digest"`
	ModifiedAt time.Time `json:"modified_at"`
}

// NewOllamaProvider creates a new Ollama provider instance.
func NewOllamaProvider(config *ModelConfig) (BaseLanguageModel, error) {
	baseURL := "http://localhost:11434"
	if config.ProviderKwargs != nil {
		if url, ok := config.ProviderKwargs["base_url"].(string); ok {
			baseURL = url
		}
	}

	// Remove trailing slash if present
	baseURL = strings.TrimSuffix(baseURL, "/")

	provider := &OllamaProvider{
		config:  config,
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 60 * time.Second, // Longer timeout for local models
		},
	}

	// Check if Ollama is available
	if !provider.checkAvailability() {
		return nil, fmt.Errorf("Ollama server is not available at %s", baseURL)
	}

	return provider, nil
}

// checkAvailability checks if Ollama server is running.
func (p *OllamaProvider) checkAvailability() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/api/tags", nil)
	if err != nil {
		return false
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// Infer generates model output for the given prompts.
func (p *OllamaProvider) Infer(ctx context.Context, prompts []string, options map[string]any) ([][]ScoredOutput, error) {
	results := make([][]ScoredOutput, len(prompts))
	
	for i, prompt := range prompts {
		response, err := p.generateCompletion(ctx, prompt, options)
		if err != nil {
			return nil, fmt.Errorf("failed to generate completion for prompt %d: %w", i, err)
		}
		
		outputs := []ScoredOutput{
			{
				Output: response.Response,
				Score:  1.0, // Ollama doesn't provide scores, use default
			},
		}
		results[i] = outputs
	}
	
	return results, nil
}

// generateCompletion makes a request to the Ollama API.
func (p *OllamaProvider) generateCompletion(ctx context.Context, prompt string, options map[string]any) (*OllamaResponse, error) {
	request := OllamaRequest{
		Model:  p.config.ModelID,
		Prompt: prompt,
		Stream: false, // Non-streaming for simplicity
	}

	// Build options from config
	ollamaOptions := make(map[string]any)
	if p.config.Temperature != 0 {
		ollamaOptions["temperature"] = p.config.Temperature
	}
	if p.config.MaxTokens != 0 {
		ollamaOptions["num_predict"] = p.config.MaxTokens
	}
	if p.config.TopP != 0 {
		ollamaOptions["top_p"] = p.config.TopP
	}

	// Add any additional options
	if options != nil {
		for k, v := range options {
			ollamaOptions[k] = v
		}
	}

	if len(ollamaOptions) > 0 {
		request.Options = ollamaOptions
	}

	// Add JSON output instruction if schema is applied
	if p.schema != nil {
		request.Prompt = prompt + "\n\nPlease respond with valid JSON only."
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/generate", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// ParseOutput processes raw model output into structured format.
func (p *OllamaProvider) ParseOutput(output string) (any, error) {
	// Try to parse as JSON first
	var result any
	
	// Remove code fences if present
	cleanOutput := p.cleanOutput(output)
	
	if err := json.Unmarshal([]byte(cleanOutput), &result); err != nil {
		// If JSON parsing fails, return as string
		return cleanOutput, nil
	}
	
	return result, nil
}

// cleanOutput removes markdown code fences and extra whitespace.
func (p *OllamaProvider) cleanOutput(output string) string {
	output = strings.TrimSpace(output)
	
	// Remove ```json and ``` fences
	if strings.HasPrefix(output, "```json") {
		output = strings.TrimPrefix(output, "```json")
		output = strings.TrimSpace(output)
	}
	if strings.HasPrefix(output, "```") {
		output = strings.TrimPrefix(output, "```")
		output = strings.TrimSpace(output)
	}
	if strings.HasSuffix(output, "```") {
		output = strings.TrimSuffix(output, "```")
		output = strings.TrimSpace(output)
	}
	
	return output
}

// ApplySchema applies schema constraints to the model.
func (p *OllamaProvider) ApplySchema(schema any) {
	p.schema = schema
}

// SetFenceOutput configures whether output should be fenced.
func (p *OllamaProvider) SetFenceOutput(enabled bool) {
	p.fenceOutput = enabled
}

// GetModelID returns the model identifier.
func (p *OllamaProvider) GetModelID() string {
	return p.config.ModelID
}

// IsAvailable checks if the provider is ready for use.
func (p *OllamaProvider) IsAvailable() bool {
	return p.checkAvailability()
}

// GetAvailableModels returns a list of locally available models.
func (p *OllamaProvider) GetAvailableModels(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tagsResponse OllamaTagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&tagsResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	models := make([]string, len(tagsResponse.Models))
	for i, model := range tagsResponse.Models {
		models[i] = model.Name
	}

	return models, nil
}

// init registers the Ollama provider with the global registry.
func init() {
	Register("ollama", NewOllamaProvider)
	
	// Register common Ollama model aliases
	RegisterAlias("llama3.2", "ollama")
	RegisterAlias("llama3.2:1b", "ollama")
	RegisterAlias("llama3.2:3b", "ollama")
	RegisterAlias("llama3.1", "ollama")
	RegisterAlias("llama3.1:8b", "ollama")
	RegisterAlias("llama3.1:70b", "ollama")
	RegisterAlias("codellama", "ollama")
	RegisterAlias("mistral", "ollama")
	RegisterAlias("qwen2.5", "ollama")
}