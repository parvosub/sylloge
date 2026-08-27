package summarize

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// OpenAICompatibleSummarizer implements the Summarizer interface using any
// OpenAI-compatible chat completions API (e.g. Ollama's /v1 endpoint, or a
// cloud provider such as OpenAI, Together, or Groq).
type OpenAICompatibleSummarizer struct {
	client       *http.Client
	baseURL      string
	model        string
	systemPrompt string
	apiKey       string
}

// OpenAIConfig holds the configuration for an OpenAICompatibleSummarizer.
// BaseURL and Model are required; SystemPrompt, APIKey, and Client are optional.
type OpenAIConfig struct {
	BaseURL      string
	Model        string
	SystemPrompt string
	APIKey       string
	Client       *http.Client
}

// NewOpenAICompatibleSummarizer creates a new summarizer from the given config.
func NewOpenAICompatibleSummarizer(cfg OpenAIConfig) (*OpenAICompatibleSummarizer, error) {
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("OpenAIConfig.BaseURL is required")
	}
	if cfg.Model == "" {
		return nil, fmt.Errorf("OpenAIConfig.Model is required")
	}
	client := cfg.Client
	if client == nil {
		client = &http.Client{}
	}
	return &OpenAICompatibleSummarizer{
		client:       client,
		baseURL:      cfg.BaseURL,
		model:        cfg.Model,
		systemPrompt: cfg.SystemPrompt,
		apiKey:       cfg.APIKey,
	}, nil
}

// Summarize takes raw notes and returns a coherent summary using the
// OpenAI-compatible chat completions API.
func (o *OpenAICompatibleSummarizer) Summarize(notes string) (string, error) {
	messages := []map[string]string{
		{"role": "system", "content": o.systemPrompt},
		{"role": "user", "content": fmt.Sprintf("Summarize the following student work notes into a coherent, professional summary for a report card comment:\n\n%s", notes)},
	}
	payload := map[string]interface{}{
		"model":    o.model,
		"messages": messages,
		"stream":   false,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := o.baseURL + "/chat/completions"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if o.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+o.apiKey)
	}

	resp, err := o.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("API response contained no choices")
	}

	return result.Choices[0].Message.Content, nil
}
