package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// OpenAIProvider implements the BaseLanguageModel interface for OpenAI models.
type OpenAIProvider struct {
	config      *ModelConfig
	apiKey      string
	baseURL     string
	client      *http.Client
	schema      interface{}
	fenceOutput bool
}

// OpenAIRequest represents an OpenAI API request.
type OpenAIRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	TopP        float64   `json:"top_p,omitempty"`
}

// Message represents a chat message.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIResponse represents an OpenAI API response.
type OpenAIResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice represents a response choice.
type Choice struct {
	Index   int     `json:"index"`
	Message Message `json:"message"`
}

// Usage represents token usage information.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// NewOpenAIProvider creates a new OpenAI provider instance.
func NewOpenAIProvider(config *ModelConfig) (BaseLanguageModel, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		if config.ProviderKwargs != nil {
			if key, ok := config.ProviderKwargs["api_key"].(string); ok {
				apiKey = key
			}
		}
	}
	
	if apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key not found. Set OPENAI_API_KEY environment variable or provide in config")
	}

	baseURL := "https://api.openai.com/v1"
	if config.ProviderKwargs != nil {
		if url, ok := config.ProviderKwargs["base_url"].(string); ok {
			baseURL = url
		}
	}

	return &OpenAIProvider{
		config:  config,
		apiKey:  apiKey,
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// Infer generates model output for the given prompts.
func (p *OpenAIProvider) Infer(ctx context.Context, prompts []string, options map[string]interface{}) ([][]ScoredOutput, error) {
	results := make([][]ScoredOutput, len(prompts))
	
	for i, prompt := range prompts {
		response, err := p.generateCompletion(ctx, prompt)
		if err != nil {
			return nil, fmt.Errorf("failed to generate completion for prompt %d: %w", i, err)
		}
		
		outputs := make([]ScoredOutput, len(response.Choices))
		for j, choice := range response.Choices {
			outputs[j] = ScoredOutput{
				Output: choice.Message.Content,
				Score:  1.0, // OpenAI doesn't provide scores, use default
			}
		}
		results[i] = outputs
	}
	
	return results, nil
}

// generateCompletion makes a request to the OpenAI API.
func (p *OpenAIProvider) generateCompletion(ctx context.Context, prompt string) (*OpenAIResponse, error) {
	request := OpenAIRequest{
		Model: p.config.ModelID,
		Messages: []Message{
			{
				Role:    "user", 
				Content: prompt,
			},
		},
		Temperature: p.config.Temperature,
		MaxTokens:   p.config.MaxTokens,
		TopP:        p.config.TopP,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// ParseOutput processes raw model output into structured format.
func (p *OpenAIProvider) ParseOutput(output string) (interface{}, error) {
	// Try to parse as JSON first
	var result interface{}
	
	// Remove code fences if present
	cleanOutput := p.cleanOutput(output)
	
	if err := json.Unmarshal([]byte(cleanOutput), &result); err != nil {
		// If JSON parsing fails, return as string
		return cleanOutput, nil
	}
	
	return result, nil
}

// cleanOutput removes markdown code fences and extra whitespace.
func (p *OpenAIProvider) cleanOutput(output string) string {
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
func (p *OpenAIProvider) ApplySchema(schema interface{}) {
	p.schema = schema
}

// SetFenceOutput configures whether output should be fenced.
func (p *OpenAIProvider) SetFenceOutput(enabled bool) {
	p.fenceOutput = enabled
}

// GetModelID returns the model identifier.
func (p *OpenAIProvider) GetModelID() string {
	return p.config.ModelID
}

// IsAvailable checks if the provider is ready for use.
func (p *OpenAIProvider) IsAvailable() bool {
	return p.apiKey != ""
}

// init registers the OpenAI provider with the global registry.
func init() {
	Register("openai", NewOpenAIProvider)
	
	// Register common OpenAI model aliases
	RegisterAlias("gpt-4", "openai")
	RegisterAlias("gpt-4-turbo", "openai")
	RegisterAlias("gpt-3.5-turbo", "openai")
	RegisterAlias("gpt-4o", "openai")
	RegisterAlias("gpt-4o-mini", "openai")
}