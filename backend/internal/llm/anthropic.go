package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type AnthropicProvider struct {
	APIKey  string
	Client  *http.Client
	BaseURL string
}

func (p *AnthropicProvider) Chat(ctx context.Context, req Request) (string, error) {
	body := p.buildBody(req, false)
	b, _ := json.Marshal(body)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL()+"/v1/messages", bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	p.setHeaders(httpReq)

	resp, err := p.Client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("anthropic %d: %s", resp.StatusCode, string(raw))
	}

	var out struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if len(out.Content) == 0 {
		return "", fmt.Errorf("empty response")
	}
	return out.Content[0].Text, nil
}

func (p *AnthropicProvider) Stream(ctx context.Context, req Request, w io.Writer) error {
	body := p.buildBody(req, true)
	b, _ := json.Marshal(body)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL()+"/v1/messages", bytes.NewReader(b))
	if err != nil {
		return err
	}
	p.setHeaders(httpReq)
	httpReq.Header.Set("accept", "text/event-stream")

	resp, err := p.Client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("anthropic stream %d: %s", resp.StatusCode, string(raw))
	}

	type delta struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	type event struct {
		Type  string `json:"type"`
		Delta delta  `json:"delta"`
	}

	scanner := bufio.NewScanner(resp.Body)
	var eventType string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "event: ") {
			eventType = strings.TrimPrefix(line, "event: ")
			continue
		}
		if strings.HasPrefix(line, "data: ") && eventType == "content_block_delta" {
			data := strings.TrimPrefix(line, "data: ")
			var ev event
			if err := json.Unmarshal([]byte(data), &ev); err == nil && ev.Delta.Type == "text_delta" {
				_, _ = io.WriteString(w, ev.Delta.Text)
			}
		}
		if line == "" {
			eventType = ""
		}
	}
	return scanner.Err()
}

func (p *AnthropicProvider) baseURL() string {
	if p.BaseURL != "" {
		return p.BaseURL
	}
	return "https://api.anthropic.com"
}

func (p *AnthropicProvider) buildBody(req Request, stream bool) map[string]any {
	body := map[string]any{
		"model":      req.Model,
		"max_tokens": req.MaxTokens,
		"stream":     stream,
		"messages": []map[string]string{
			{"role": "user", "content": req.UserPrompt},
		},
	}
	if req.System != "" {
		body["system"] = req.System
	}
	if req.JSONMode {
		sys, _ := body["system"].(string)
		body["system"] = sys + "\nRespond with valid JSON only."
	}
	return body
}

func (p *AnthropicProvider) setHeaders(r *http.Request) {
	r.Header.Set("x-api-key", p.APIKey)
	r.Header.Set("anthropic-version", "2023-06-01")
	r.Header.Set("content-type", "application/json")
}
