package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"

	"trisse-home/internal/modules/summary/domain"
)

// Ollama implements usecase.Summarizer with the Ollama HTTP API.
type Ollama struct{ client *http.Client }

// NewOllama returns the adapter; the client needs a generous timeout, a
// model on a CPU takes a while.
func NewOllama(client *http.Client) *Ollama { return &Ollama{client: client} }

const maxOllamaResponse = 1 << 20

// Summarize implements usecase.Summarizer (POST /api/generate, no streaming).
func (o *Ollama) Summarize(ctx context.Context, s domain.Settings, prompt string) (string, error) {
	body, err := json.Marshal(map[string]any{
		"model": s.Model, "prompt": prompt, "stream": false,
		"options": map[string]any{"temperature": 0.2, "num_predict": 400},
	})
	if err != nil {
		return "", fmt.Errorf("encode request: %w", err)
	}
	var out struct {
		Response string `json:"response"`
		Error    string `json:"error"`
	}
	if err := o.do(ctx, http.MethodPost, s.URL+"/api/generate", body, &out); err != nil {
		return "", err
	}
	if out.Error != "" {
		return "", fmt.Errorf("%s", out.Error)
	}
	return out.Response, nil
}

// Models implements usecase.Summarizer (GET /api/tags).
func (o *Ollama) Models(ctx context.Context, url string) ([]string, error) {
	var out struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := o.do(ctx, http.MethodGet, url+"/api/tags", nil, &out); err != nil {
		return nil, err
	}
	names := make([]string, 0, len(out.Models))
	for _, m := range out.Models {
		names = append(names, m.Name)
	}
	sort.Strings(names)
	return names, nil
}

func (o *Ollama) do(ctx context.Context, method, url string, body []byte, dst any) error {
	var rd io.Reader
	if body != nil {
		rd = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, rd)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := o.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()
	data, err := io.ReadAll(io.LimitReader(res.Body, maxOllamaResponse))
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	if res.StatusCode != http.StatusOK {
		var e struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(data, &e) == nil && e.Error != "" {
			return fmt.Errorf("ollama: %s", e.Error)
		}
		return fmt.Errorf("ollama answered %s", res.Status)
	}
	if err := json.Unmarshal(data, dst); err != nil {
		return fmt.Errorf("unexpected response (is this an Ollama server?): %w", err)
	}
	return nil
}
