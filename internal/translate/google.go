package translate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// googleBaseURL is the default Google Translate v2 endpoint.
const googleBaseURL = "https://translation.googleapis.com/language/translate/v2"

// googleRequest matches the Google Translate v2 payload.
type googleRequest struct {
	Target string   `json:"target"`
	Source string   `json:"source,omitempty"`
	Format string   `json:"format,omitempty"`
	Texts  []string `json:"q"`
}

// GoogleClient implements the Google Translate v2 API.
type GoogleClient struct {
	HTTPClient *http.Client
	URL        string
	APIKey     string
	SourceLang string
	Format     string
}

func (c *GoogleClient) httpClient() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}

	return &http.Client{Timeout: 60 * time.Second}
}

// Translate sends a request to Google Translate and returns translated strings in order.
func (c *GoogleClient) Translate(ctx context.Context, req Request) ([]string, error) {
	if c.APIKey == "" {
		return nil, fmt.Errorf("google api key is required")
	}
	if req.TargetLang == "" {
		return nil, fmt.Errorf("google target language is required")
	}
	if len(req.Texts) == 0 {
		return nil, fmt.Errorf("google translation request is empty")
	}

	endpoint := c.URL
	if endpoint == "" {
		endpoint = googleBaseURL
	}

	source := req.SourceLang
	if source == "" {
		source = c.SourceLang
	}

	payload := googleRequest{
		Texts:  req.Texts,
		Target: req.TargetLang,
		Source: source,
		Format: c.Format,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("google marshal request: %w", err)
	}

	endpointURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("google parse url: %w", err)
	}
	query := endpointURL.Query()
	query.Set("key", c.APIKey)
	endpointURL.RawQuery = query.Encode()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpointURL.String(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("google request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient().Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("google request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("google read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("google response %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var parsed struct {
		Data struct {
			Translations []struct {
				TranslatedText string `json:"translatedText"`
			} `json:"translations"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("google parse response: %w", err)
	}
	if len(parsed.Data.Translations) != len(req.Texts) {
		return nil, fmt.Errorf("google response size mismatch: got %d, want %d", len(parsed.Data.Translations), len(req.Texts))
	}

	out := make([]string, 0, len(parsed.Data.Translations))
	for _, item := range parsed.Data.Translations {
		out = append(out, item.TranslatedText)
	}

	return out, nil
}
