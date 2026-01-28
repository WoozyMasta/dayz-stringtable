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

const (
	deeplPaidBaseURL = "https://api.deepl.com/v2/translate"
	deeplFreeBaseURL = "https://api-free.deepl.com/v2/translate"
)

// deeplRequest matches the DeepL JSON translation payload.
type deeplRequest struct {
	TargetLang         string   `json:"target_lang"`
	SourceLang         string   `json:"source_lang,omitempty"`
	Formality          string   `json:"formality,omitempty"`
	SplitSentences     string   `json:"split_sentences,omitempty"`
	Texts              []string `json:"text"`
	PreserveFormatting int      `json:"preserve_formatting,omitempty"`
}

// DeeplClient implements the DeepL API.
type DeeplClient struct {
	HTTPClient         *http.Client
	URL                string
	AuthKey            string
	SourceLang         string
	Formality          string
	SplitSentences     string
	PreserveFormatting bool
	UseFreeAPI         bool
}

func (c *DeeplClient) httpClient() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}

	return &http.Client{Timeout: 60 * time.Second}
}

// Translate sends a request to DeepL and returns translated strings in order.
func (c *DeeplClient) Translate(ctx context.Context, req Request) ([]string, error) {
	if c.AuthKey == "" {
		return nil, fmt.Errorf("deepl auth key is required")
	}
	if req.TargetLang == "" {
		return nil, fmt.Errorf("deepl target language is required")
	}
	if len(req.Texts) == 0 {
		return nil, fmt.Errorf("deepl translation request is empty")
	}

	endpoint := c.URL
	if endpoint == "" {
		if c.UseFreeAPI {
			endpoint = deeplFreeBaseURL
		} else {
			endpoint = deeplPaidBaseURL
		}
	}

	source := req.SourceLang
	if source == "" {
		source = c.SourceLang
	}

	payload := deeplRequest{
		Texts:              req.Texts,
		TargetLang:         req.TargetLang,
		SourceLang:         source,
		Formality:          c.Formality,
		PreserveFormatting: boolToInt(c.PreserveFormatting),
		SplitSentences:     c.SplitSentences,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("deepl marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("deepl request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "DeepL-Auth-Key "+c.AuthKey)

	resp, err := c.httpClient().Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("deepl request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("deepl read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("deepl response %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var parsed struct {
		Translations []struct {
			Text string `json:"text"`
		} `json:"translations"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("deepl parse response: %w", err)
	}
	if len(parsed.Translations) != len(req.Texts) {
		return nil, fmt.Errorf("deepl response size mismatch: got %d, want %d", len(parsed.Translations), len(req.Texts))
	}

	out := make([]string, 0, len(parsed.Translations))
	for _, item := range parsed.Translations {
		out = append(out, item.Text)
	}

	return out, nil
}

// boolToInt converts booleans to DeepL's 0/1 toggle values.
func boolToInt(value bool) int {
	if value {
		return 1
	}

	return 0
}
