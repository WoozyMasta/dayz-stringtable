// Package translate provides translation clients for PO file content.
package translate

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// Request describes a translation request.
type Request struct {
	SourceLang string
	TargetLang string
	Texts      []string
}

// Client translates batches of strings.
type Client interface {
	Translate(ctx context.Context, req Request) ([]string, error)
}

// parseJSONArray extracts a JSON array of strings from a response payload.
func parseJSONArray(content string) ([]string, error) {
	var out []string
	if err := json.Unmarshal([]byte(content), &out); err == nil {
		return out, nil
	}

	// Some models may wrap the JSON in extra text; try to salvage the array.
	start := strings.Index(content, "[")
	end := strings.LastIndex(content, "]")
	if start >= 0 && end > start {
		if err := json.Unmarshal([]byte(content[start:end+1]), &out); err == nil {
			return out, nil
		}
	}

	return nil, fmt.Errorf("expected JSON array, got: %s", content)
}
