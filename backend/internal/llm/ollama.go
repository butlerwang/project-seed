package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type OllamaProvider struct {
	BaseURL string
	Client  *http.Client
}

func (p *OllamaProvider) Chat(ctx context.Context, req Request) (string, error) {
	prompt := req.UserPrompt
	if req.System != "" {
		prompt = req.System + "\n\n" + prompt
	}
	body := map[string]any{
		"model":  req.Model,
		"prompt": prompt,
		"stream": false,
	}
	if req.JSONMode {
		body["format"] = "json"
	}

	b, _ := json.Marshal(body)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.BaseURL+"/api/generate", bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("content-type", "application/json")

	resp, err := p.Client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var out struct {
		Response string `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("ollama decode: %w", err)
	}
	return out.Response, nil
}
