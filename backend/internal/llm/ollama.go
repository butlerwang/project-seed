package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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

func (p *OllamaProvider) Stream(ctx context.Context, req Request, w io.Writer) error {
	prompt := req.UserPrompt
	if req.System != "" {
		prompt = req.System + "\n\n" + prompt
	}
	body := map[string]any{
		"model":  req.Model,
		"prompt": prompt,
		"stream": true,
	}
	if req.JSONMode {
		body["format"] = "json"
	}

	b, _ := json.Marshal(body)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.BaseURL+"/api/generate", bytes.NewReader(b))
	if err != nil {
		return err
	}
	httpReq.Header.Set("content-type", "application/json")

	resp, err := p.Client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	type chunk struct {
		Response string `json:"response"`
		Done     bool   `json:"done"`
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		var c chunk
		if err := json.Unmarshal(scanner.Bytes(), &c); err != nil {
			continue
		}
		if c.Done {
			break
		}
		_, _ = io.WriteString(w, c.Response)
	}
	return scanner.Err()
}
