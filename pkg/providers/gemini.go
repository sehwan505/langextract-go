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

// GeminiProvider implements the BaseLanguageModel interface for Google Gemini models.
type GeminiProvider struct {
	config      *ModelConfig
	apiKey      string
	baseURL     string
	client      *http.Client
	schema      any
	fenceOutput bool
}

// GeminiRequest represents a Gemini API request.
type GeminiRequest struct {
	Contents         []GeminiContent         `json:"contents"`
	GenerationConfig *GeminiGenerationConfig `json:"generationConfig,omitempty"`
}

// GeminiContent represents the content part of a Gemini request.
type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

// GeminiPart represents a part of the content (text, image, etc.).
type GeminiPart struct {
	Text string `json:"text"`
}

// GeminiGenerationConfig represents generation configuration.
type GeminiGenerationConfig struct {
	Temperature      *float64 `json:"temperature,omitempty"`
	TopP             *float64 `json:"topP,omitempty"`
	MaxOutputTokens  *int     `json:"maxOutputTokens,omitempty"`
	ResponseMimeType string   `json:"responseMimeType,omitempty"`
}

// GeminiResponse represents a Gemini API response.
type GeminiResponse struct {
	Candidates    []GeminiCandidate `json:"candidates"`
	UsageMetadata GeminiUsage       `json:"usageMetadata"`
}

// GeminiCandidate represents a response candidate.
type GeminiCandidate struct {
	Content GeminiContent `json:"content"`
	Index   int           `json:"index"`
}

// GeminiUsage represents token usage information.
type GeminiUsage struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

// NewGeminiProvider creates a new Gemini provider instance.
func NewGeminiProvider(config *ModelConfig) (BaseLanguageModel, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		// Also check GOOGLE_API_KEY as alternative
		apiKey = os.Getenv("GOOGLE_API_KEY")
	}
	if apiKey == "" {
		if config.ProviderKwargs != nil {
			if key, ok := config.ProviderKwargs["api_key"].(string); ok {
				apiKey = key
			}
		}
	}
	
	if apiKey == "" {
		return nil, fmt.Errorf("Gemini API key not found. Set GEMINI_API_KEY or GOOGLE_API_KEY environment variable or provide in config")
	}

	baseURL := "https://generativelanguage.googleapis.com/v1beta"
	if config.ProviderKwargs != nil {
		if url, ok := config.ProviderKwargs["base_url"].(string); ok {
			baseURL = url
		}
	}

	return &GeminiProvider{
		config:  config,
		apiKey:  apiKey,
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// Infer generates model output for the given prompts.
func (p *GeminiProvider) Infer(ctx context.Context, prompts []string, options map[string]any) ([][]ScoredOutput, error) {
	results := make([][]ScoredOutput, len(prompts))
	
	for i, prompt := range prompts {
		response, err := p.generateCompletion(ctx, prompt)
		if err != nil {
			return nil, fmt.Errorf("failed to generate completion for prompt %d: %w", i, err)
		}
		
		outputs := make([]ScoredOutput, len(response.Candidates))
		for j, candidate := range response.Candidates {
			text := ""
			if len(candidate.Content.Parts) > 0 {
				text = candidate.Content.Parts[0].Text
			}
			outputs[j] = ScoredOutput{
				Output: text,
				Score:  1.0, // Gemini doesn't provide scores, use default
			}
		}
		results[i] = outputs
	}
	
	return results, nil
}

// generateCompletion makes a request to the Gemini API.
func (p *GeminiProvider) generateCompletion(ctx context.Context, prompt string) (*GeminiResponse, error) {
	request := GeminiRequest{
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{
					{Text: prompt},
				},
			},
		},
	}

	// Add generation config if needed
	if p.config.Temperature != 0 || p.config.MaxTokens != 0 || p.config.TopP != 0 {
		request.GenerationConfig = &GeminiGenerationConfig{}
		
		if p.config.Temperature != 0 {
			request.GenerationConfig.Temperature = &p.config.Temperature
		}
		if p.config.TopP != 0 {
			request.GenerationConfig.TopP = &p.config.TopP
		}
		if p.config.MaxTokens != 0 {
			request.GenerationConfig.MaxOutputTokens = &p.config.MaxTokens
		}
	}

	// Enable JSON output if schema is applied
	if p.schema != nil {
		if request.GenerationConfig == nil {
			request.GenerationConfig = &GeminiGenerationConfig{}
		}
		request.GenerationConfig.ResponseMimeType = "application/json"
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/models/%s:generateContent", p.baseURL, p.config.ModelID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response GeminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// ParseOutput processes raw model output into structured format.
func (p *GeminiProvider) ParseOutput(output string) (any, error) {
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
func (p *GeminiProvider) cleanOutput(output string) string {
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
func (p *GeminiProvider) ApplySchema(schema any) {
	p.schema = schema
}

// SetFenceOutput configures whether output should be fenced.
func (p *GeminiProvider) SetFenceOutput(enabled bool) {
	p.fenceOutput = enabled
}

// GetModelID returns the model identifier.
func (p *GeminiProvider) GetModelID() string {
	return p.config.ModelID
}

// IsAvailable checks if the provider is ready for use.
func (p *GeminiProvider) IsAvailable() bool {
	return p.apiKey != ""
}

// init registers the Gemini provider with the global registry.
func init() {
	Register("gemini", NewGeminiProvider)
	
	// Register common Gemini model aliases
	RegisterAlias("gemini-2.5-flash", "gemini")
	RegisterAlias("gemini-2.5-pro", "gemini")
	RegisterAlias("gemini-1.5-pro", "gemini")
	RegisterAlias("gemini-1.5-flash", "gemini")
	RegisterAlias("gemini-pro", "gemini")
}