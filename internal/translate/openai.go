package translate

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

// openAIRequest matches the OpenAI-compatible chat completion payload.
type openAIRequest struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	Temperature float64         `json:"temperature,omitempty"`
}

// openAIMessage is a single chat message for OpenAI-compatible APIs.
type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// openAIResponse extracts the assistant content from chat completions.
type openAIResponse struct {
	Choices []struct {
		Message openAIMessage `json:"message"`
	} `json:"choices"`
}

const (
	openAIPromptPreserve = "Preserve punctuation, spacing, and placeholders like {name}, %%s, {0}, or <tag>."
	openAIPromptJSONOnly = "Return ONLY a JSON array of strings in the same order."
)

// OpenAIClient implements OpenAI-compatible chat completions.
type OpenAIClient struct {
	HTTPClient  *http.Client
	BaseURL     string
	APIKey      string
	Model       string
	Temperature float64
}

func (c *OpenAIClient) httpClient() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}

	return &http.Client{Timeout: 60 * time.Second}
}

// Translate sends a chat completion request and returns translated strings in order.
func (c *OpenAIClient) Translate(ctx context.Context, req Request) ([]string, error) {
	if c.APIKey == "" {
		return nil, fmt.Errorf("openai api key is required")
	}
	if req.TargetLang == "" {
		return nil, fmt.Errorf("openai target language is required")
	}
	if len(req.Texts) == 0 {
		return nil, fmt.Errorf("openai translation request is empty")
	}

	baseURL := strings.TrimRight(c.BaseURL, "/")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	model := c.Model
	if model == "" {
		model = "gpt-4o-mini"
	}

	system := buildSystemPrompt(req.SourceLang, req.TargetLang)
	input, err := json.Marshal(req.Texts)
	if err != nil {
		return nil, fmt.Errorf("openai marshal input: %w", err)
	}

	payload := openAIRequest{
		Model: model,
		Messages: []openAIMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: string(input)},
		},
	}
	if c.Temperature != 0 {
		payload.Temperature = c.Temperature
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("openai marshal request: %w", err)
	}

	endpoint := baseURL + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("openai request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.httpClient().Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("openai request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("openai read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("openai response %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var parsed openAIResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("openai parse response: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return nil, fmt.Errorf("openai response missing choices")
	}

	content := strings.TrimSpace(parsed.Choices[0].Message.Content)
	translations, err := parseJSONArray(content)
	if err != nil {
		return nil, fmt.Errorf("openai parse translations: %w", err)
	}
	if len(translations) != len(req.Texts) {
		return nil, fmt.Errorf("openai response size mismatch: got %d, want %d", len(translations), len(req.Texts))
	}

	return translations, nil
}

// buildSystemPrompt creates the translation instruction for the LLM.
func buildSystemPrompt(sourceLang, targetLang string) string {
	if sourceLang != "" {
		return fmt.Sprintf(
			"Translate the following strings from %s to %s. %s %s",
			sourceLang,
			targetLang,
			openAIPromptPreserve,
			openAIPromptJSONOnly,
		)
	}
	return fmt.Sprintf(
		"Translate the following strings into %s. %s %s",
		targetLang,
		openAIPromptPreserve,
		openAIPromptJSONOnly,
	)
}
