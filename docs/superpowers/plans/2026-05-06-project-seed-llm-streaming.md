# project-seed LLM Streaming Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `Stream()` support to the LLM router (Anthropic SSE + Ollama streaming) and expose an SSE endpoint so the frontend can stream token-by-token LLM responses.

**Architecture:** The existing `Provider` interface gains a `Stream()` method that writes tokens to an `io.Writer`. The `Router` gains a `Stream()` method using the same `pick()` logic. A new HTTP handler at `POST /api/v1/llm/stream` reads the request body and writes `text/event-stream` SSE to the client. The frontend gets a `useStream()` hook that reads from the SSE endpoint and appends tokens to a state string.

**Tech Stack:** Go stdlib `bufio.Scanner` + `http.Flusher` for SSE; `EventSource` API / `fetch` with `ReadableStream` for the frontend hook.

---

## File Map

| Action | Path | Responsibility |
|--------|------|---------------|
| Modify | `backend/internal/llm/provider.go` | Add `Stream(ctx, req, io.Writer) error` to `Provider` interface |
| Modify | `backend/internal/llm/anthropic.go` | Implement `AnthropicProvider.Stream()` |
| Modify | `backend/internal/llm/ollama.go` | Implement `OllamaProvider.Stream()` |
| Modify | `backend/internal/llm/router.go` | Add `Router.Stream()` method |
| Create | `backend/internal/llm/router_test.go` | Unit tests for Router.Stream with mock provider |
| Create | `backend/internal/handler/llm.go` | SSE HTTP handler for streaming |
| Create | `backend/internal/handler/llm_test.go` | Handler test with mock LLM |
| Modify | `backend/internal/handler/router.go` | Register `POST /api/v1/llm/stream` |
| Modify | `backend/internal/service/services.go` | Expose LLM router on Services struct |
| Create | `frontend/lib/useStream.ts` | React hook: POST → SSE ReadableStream |

---

## Task 1: Extend the Provider interface with Stream

**Files:**
- Modify: `backend/internal/llm/provider.go`

- [ ] **Step 1: Update `backend/internal/llm/provider.go`**

The current file is:
```go
package llm

import "context"

type Request struct {
	Model      string
	System     string
	UserPrompt string
	MaxTokens  int
	JSONMode   bool
}

type Provider interface {
	Chat(ctx context.Context, req Request) (string, error)
}
```

Replace with:
```go
package llm

import (
	"context"
	"io"
)

type Request struct {
	Model      string
	System     string
	UserPrompt string
	MaxTokens  int
	JSONMode   bool
}

// Provider is an LLM backend (Anthropic or Ollama).
type Provider interface {
	// Chat returns the full response as a string.
	Chat(ctx context.Context, req Request) (string, error)
	// Stream writes token chunks to w as they arrive, then closes.
	// Each write is one or more tokens; w must be flushed by the caller.
	Stream(ctx context.Context, req Request, w io.Writer) error
}
```

- [ ] **Step 2: Build to see compile errors (both providers now break)**

```bash
cd backend && go build ./...
```

Expected: FAIL — `AnthropicProvider` and `OllamaProvider` do not implement `Provider`.

- [ ] **Step 3: Commit the interface change alone**

```bash
git add internal/llm/provider.go
git commit -m "feat(llm): add Stream() to Provider interface"
```

---

## Task 2: Implement AnthropicProvider.Stream

**Files:**
- Modify: `backend/internal/llm/anthropic.go`

Anthropic streaming uses SSE. Each event looks like:
```
event: content_block_delta
data: {"type":"content_block_delta","delta":{"type":"text_delta","text":"Hello"}}
```
We only care about `content_block_delta` events where `delta.type == "text_delta"`.

- [ ] **Step 1: Write the failing test — create `backend/internal/llm/anthropic_test.go`**

This test uses `httptest.Server` to mock the Anthropic SSE API:

```go
package llm_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/butlerwang/project-seed/backend/internal/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnthropicProvider_Stream(t *testing.T) {
	sseBody := strings.Join([]string{
		`event: message_start`,
		`data: {"type":"message_start"}`,
		``,
		`event: content_block_start`,
		`data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`,
		``,
		`event: content_block_delta`,
		`data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello"}}`,
		``,
		`event: content_block_delta`,
		`data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":" world"}}`,
		``,
		`event: message_delta`,
		`data: {"type":"message_delta","delta":{"stop_reason":"end_turn"}}`,
		``,
		`event: message_stop`,
		`data: {"type":"message_stop"}`,
	}, "\n")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(sseBody))
	}))
	defer srv.Close()

	provider := &llm.AnthropicProvider{
		APIKey:      "test-key",
		Client:      srv.Client(),
		BaseURL:     srv.URL, // override for testing
	}

	var buf bytes.Buffer
	err := provider.Stream(context.Background(), llm.Request{
		Model:      "claude-haiku-test",
		UserPrompt: "say hello",
		MaxTokens:  100,
	}, &buf)
	require.NoError(t, err)
	assert.Equal(t, "Hello world", buf.String())
}
```

Note: we add a `BaseURL` field to `AnthropicProvider` so tests can point at a mock server.

- [ ] **Step 2: Run to confirm failure**

```bash
go test ./internal/llm/... -run TestAnthropicProvider_Stream -v
```

Expected: FAIL — `AnthropicProvider` has no `Stream` method and no `BaseURL` field.

- [ ] **Step 3: Update `AnthropicProvider` struct and implement `Stream` in `backend/internal/llm/anthropic.go`**

Replace the entire file:

```go
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
	BaseURL string // defaults to "https://api.anthropic.com" if empty
}

func (p *AnthropicProvider) baseURL() string {
	if p.BaseURL != "" {
		return p.BaseURL
	}
	return "https://api.anthropic.com"
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

// Stream writes token chunks to w as they arrive via Anthropic SSE.
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

	// Parse SSE: look for content_block_delta events with text_delta
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
				io.WriteString(w, ev.Delta.Text) //nolint:errcheck
			}
		}
		if line == "" {
			eventType = ""
		}
	}
	return scanner.Err()
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
```

- [ ] **Step 4: Run test**

```bash
go test ./internal/llm/... -run TestAnthropicProvider_Stream -v
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/llm/anthropic.go internal/llm/anthropic_test.go
git commit -m "feat(llm): implement AnthropicProvider.Stream() via SSE"
```

---

## Task 3: Implement OllamaProvider.Stream

**Files:**
- Modify: `backend/internal/llm/ollama.go`

Ollama's `/api/generate` with `"stream": true` returns newline-delimited JSON (NDJSON):
```json
{"model":"qwen3:0.6b","response":"Hello","done":false}
{"model":"qwen3:0.6b","response":" world","done":false}
{"model":"qwen3:0.6b","response":"","done":true}
```

- [ ] **Step 1: Write failing test — create `backend/internal/llm/ollama_test.go`**

```go
package llm_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/butlerwang/project-seed/backend/internal/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOllamaProvider_Stream(t *testing.T) {
	ndjson := "{\"response\":\"Hello\",\"done\":false}\n" +
		"{\"response\":\" world\",\"done\":false}\n" +
		"{\"response\":\"\",\"done\":true}\n"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-ndjson")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(ndjson))
	}))
	defer srv.Close()

	provider := &llm.OllamaProvider{BaseURL: srv.URL, Client: srv.Client()}

	var buf bytes.Buffer
	err := provider.Stream(context.Background(), llm.Request{
		Model:      "test-model",
		UserPrompt: "say hello",
	}, &buf)
	require.NoError(t, err)
	assert.Equal(t, "Hello world", buf.String())
}
```

- [ ] **Step 2: Run to confirm failure**

```bash
go test ./internal/llm/... -run TestOllamaProvider_Stream -v
```

Expected: FAIL — `OllamaProvider` has no `Stream` method.

- [ ] **Step 3: Implement `Stream` in `backend/internal/llm/ollama.go`**

Replace the entire file:

```go
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

// Stream writes token chunks to w as they arrive from Ollama's NDJSON stream.
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
		io.WriteString(w, c.Response) //nolint:errcheck
	}
	return scanner.Err()
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/llm/... -v
```

Expected: all PASS (Chat and Stream for both providers).

- [ ] **Step 5: Commit**

```bash
git add internal/llm/ollama.go internal/llm/ollama_test.go
git commit -m "feat(llm): implement OllamaProvider.Stream() via NDJSON"
```

---

## Task 4: Add Router.Stream and update Router.pick for new interface

**Files:**
- Modify: `backend/internal/llm/router.go`
- Create: `backend/internal/llm/router_test.go`

The `Router.pick()` currently returns `(Provider, string)` but constructs providers inline. The providers now have a `BaseURL` field on `AnthropicProvider`; the `pick()` method needs no change for baseURL (defaults to real API). We just need to add `Router.Stream()`.

- [ ] **Step 1: Write failing test — create `backend/internal/llm/router_test.go`**

```go
package llm_test

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/butlerwang/project-seed/backend/internal/config"
	"github.com/butlerwang/project-seed/backend/internal/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockProvider implements llm.Provider for tests.
type mockProvider struct {
	chatResult   string
	streamTokens []string
}

func (m *mockProvider) Chat(_ context.Context, _ llm.Request) (string, error) {
	return m.chatResult, nil
}

func (m *mockProvider) Stream(_ context.Context, _ llm.Request, w io.Writer) error {
	for _, tok := range m.streamTokens {
		io.WriteString(w, tok)
	}
	return nil
}

func TestRouter_Stream_writesTokens(t *testing.T) {
	cfg := config.Config{} // no API keys → falls back to Ollama
	router := llm.NewRouterWithProvider(cfg, &mockProvider{streamTokens: []string{"Hello", " world"}})

	var buf bytes.Buffer
	err := router.Stream(context.Background(), "research", "system", "prompt", &buf)
	require.NoError(t, err)
	assert.Equal(t, "Hello world", buf.String())
}
```

Note: we add `llm.NewRouterWithProvider(cfg, provider)` — a constructor that accepts an injected provider, for testability.

- [ ] **Step 2: Run to confirm failure**

```bash
go test ./internal/llm/... -run TestRouter_Stream -v
```

Expected: FAIL — `NewRouterWithProvider` and `Router.Stream` undefined.

- [ ] **Step 3: Update `backend/internal/llm/router.go`**

Replace the entire file:

```go
package llm

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/butlerwang/project-seed/backend/internal/config"
)

var capableOps = map[string]bool{
	"research": true,
	"generate": true,
	"analyze":  true,
}

type Router struct {
	cfg      config.Config
	client   *http.Client
	override Provider // non-nil only in tests
}

func NewRouter(cfg config.Config) *Router {
	return &Router{cfg: cfg, client: &http.Client{Timeout: 90 * time.Second}}
}

// NewRouterWithProvider creates a Router that always uses the given provider.
// Used in tests to inject a mock.
func NewRouterWithProvider(cfg config.Config, p Provider) *Router {
	return &Router{cfg: cfg, client: &http.Client{Timeout: 90 * time.Second}, override: p}
}

// Chat sends a single-turn LLM request and returns the full response.
func (r *Router) Chat(ctx context.Context, operation, system, userPrompt string, jsonMode bool) (string, error) {
	provider, model := r.pick(operation)
	return provider.Chat(ctx, Request{
		Model:      model,
		System:     system,
		UserPrompt: userPrompt,
		MaxTokens:  2048,
		JSONMode:   jsonMode,
	})
}

// Stream sends a streaming LLM request and writes tokens to w as they arrive.
func (r *Router) Stream(ctx context.Context, operation, system, userPrompt string, w io.Writer) error {
	provider, model := r.pick(operation)
	return provider.Stream(ctx, Request{
		Model:      model,
		System:     system,
		UserPrompt: userPrompt,
		MaxTokens:  2048,
	}, w)
}

func (r *Router) pick(operation string) (Provider, string) {
	if r.override != nil {
		return r.override, "mock"
	}
	capable := capableOps[operation]
	if r.cfg.AnthropicKey != "" {
		model := r.cfg.AnthropicFast
		if capable {
			model = r.cfg.AnthropicCapable
		}
		return &AnthropicProvider{APIKey: r.cfg.AnthropicKey, Client: r.client}, model
	}
	model := r.cfg.OllamaFast
	if capable {
		model = r.cfg.OllamaCapable
	}
	return &OllamaProvider{BaseURL: r.cfg.OllamaBaseURL, Client: r.client}, model
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/llm/... -v
```

Expected: all PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/llm/router.go internal/llm/router_test.go
git commit -m "feat(llm): add Router.Stream() and NewRouterWithProvider for test injection"
```

---

## Task 5: Wire LLM router into Services

**Files:**
- Modify: `backend/internal/service/services.go`

- [ ] **Step 1: Expose the LLM router on the Services struct**

Replace `backend/internal/service/services.go`:

```go
package service

import (
	"github.com/butlerwang/project-seed/backend/internal/config"
	"github.com/butlerwang/project-seed/backend/internal/llm"
	"github.com/butlerwang/project-seed/backend/internal/repository"
)

type Services struct {
	Auth *AuthService
	LLM  *llm.Router
}

func New(cfg config.Config, repos repository.Repository) *Services {
	return &Services{
		Auth: NewAuthService(cfg, repos),
		LLM:  llm.NewRouter(cfg),
	}
}
```

- [ ] **Step 2: Build check**

```bash
go build ./...
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add internal/service/services.go
git commit -m "feat: expose LLM router on Services struct"
```

---

## Task 6: SSE HTTP handler for streaming

**Files:**
- Create: `backend/internal/handler/llm.go`
- Create: `backend/internal/handler/llm_test.go`
- Modify: `backend/internal/handler/router.go`

The handler accepts `POST /api/v1/llm/stream` with JSON body `{"operation":"...", "system":"...", "prompt":"..."}` and responds with `text/event-stream`. Each token is sent as:
```
data: <token text>

```
A final `data: [DONE]` signals completion.

- [ ] **Step 1: Write failing test — create `backend/internal/handler/llm_test.go`**

```go
package handler_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/butlerwang/project-seed/backend/internal/handler"
	"github.com/butlerwang/project-seed/backend/internal/llm"
	"github.com/butlerwang/project-seed/backend/internal/service"
	"github.com/butlerwang/project-seed/backend/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLLMStreamHandler_streams(t *testing.T) {
	mock := &mockLLMProvider{tokens: []string{"Hello", " world"}}
	svc := &service.Services{
		LLM: llm.NewRouterWithProvider(testConfig(), mock),
	}
	r := handler.NewRouter(testConfig(), svc, nil, storage.NewMemoryStorage())

	body, _ := json.Marshal(map[string]string{
		"operation": "chat",
		"system":    "You are helpful.",
		"prompt":    "say hello",
	})
	req := httptest.NewRequest("POST", "/api/v1/llm/stream", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "text/event-stream", rr.Header().Get("Content-Type"))

	var tokens []string
	scanner := bufio.NewScanner(rr.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			tok := strings.TrimPrefix(line, "data: ")
			if tok != "[DONE]" {
				tokens = append(tokens, tok)
			}
		}
	}
	assert.Equal(t, []string{"Hello", " world"}, tokens)
}

// mockLLMProvider is a local mock — implements llm.Provider
type mockLLMProvider struct {
	tokens []string
}

func (m *mockLLMProvider) Chat(_ context.Context, _ llm.Request) (string, error) {
	return strings.Join(m.tokens, ""), nil
}

func (m *mockLLMProvider) Stream(_ context.Context, _ llm.Request, w io.Writer) error {
	for _, tok := range m.tokens {
		io.WriteString(w, tok)
	}
	return nil
}
```

Add missing imports at top: `"context"`, `"io"`.

- [ ] **Step 2: Run to confirm failure**

```bash
go test ./internal/handler/... -run TestLLMStreamHandler -v
```

Expected: FAIL — no `/api/v1/llm/stream` route.

- [ ] **Step 3: Implement `backend/internal/handler/llm.go`**

```go
package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type llmHandler struct {
	svc interface {
		StreamLLM(operation, system, prompt string, w io.Writer) error
	}
}

type streamRequest struct {
	Operation string `json:"operation"`
	System    string `json:"system"`
	Prompt    string `json:"prompt"`
}

func (h *llmHandler) stream(w http.ResponseWriter, r *http.Request) {
	var req streamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Prompt == "" {
		http.Error(w, "prompt is required", http.StatusBadRequest)
		return
	}
	if req.Operation == "" {
		req.Operation = "chat"
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	// sseWriter flushes after each write so the client receives tokens immediately
	sw := &sseWriter{w: w, flusher: flusher}
	if err := h.svc.StreamLLM(req.Operation, req.System, req.Prompt, sw); err != nil {
		fmt.Fprintf(w, "data: [ERROR]\n\n")
		flusher.Flush()
		return
	}
	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()
}

type sseWriter struct {
	w       http.ResponseWriter
	flusher http.Flusher
}

func (sw *sseWriter) Write(p []byte) (int, error) {
	n, err := fmt.Fprintf(sw.w, "data: %s\n\n", string(p))
	sw.flusher.Flush()
	return n, err
}
```

Wait — this approach writes one SSE event per `Stream()` write call, but `Stream()` in the providers writes token by token already. This is correct.

However, the `llmHandler` uses an interface `StreamLLM(...)` that doesn't exist on `*service.Services`. Let's instead access `svc.LLM.Stream()` directly. Update the handler to take `*service.Services`:

```go
package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/butlerwang/project-seed/backend/internal/service"
)

type llmHandler struct {
	svc *service.Services
}

type streamRequest struct {
	Operation string `json:"operation"`
	System    string `json:"system"`
	Prompt    string `json:"prompt"`
}

func (h *llmHandler) stream(w http.ResponseWriter, r *http.Request) {
	var req streamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Prompt == "" {
		http.Error(w, "prompt is required", http.StatusBadRequest)
		return
	}
	if req.Operation == "" {
		req.Operation = "chat"
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	sw := &sseWriter{w: w, flusher: flusher}
	if err := h.svc.LLM.Stream(r.Context(), req.Operation, req.System, req.Prompt, sw); err != nil {
		fmt.Fprintf(w, "data: [ERROR]\n\n")
		flusher.Flush()
		return
	}
	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()
}

type sseWriter struct {
	w       http.ResponseWriter
	flusher http.Flusher
}

func (sw *sseWriter) Write(p []byte) (int, error) {
	n, err := fmt.Fprintf(sw.w, "data: %s\n\n", string(p))
	sw.flusher.Flush()
	return n, err
}
```

- [ ] **Step 4: Register route in `backend/internal/handler/router.go`**

Add inside the authenticated group:

```go
llmH := &llmHandler{svc: svc}

r.Group(func(r chi.Router) {
    r.Use(middleware.RequireAuth(svc.Auth))
    r.Get("/auth/me", auth.me)
    r.Post("/files", upload.upload)
    r.Get("/files/{key}", upload.download)
    r.Post("/llm/stream", llmH.stream)   // ← add this

    r.Group(func(r chi.Router) {
        r.Use(middleware.RequireAdmin())
        r.Get("/admin/users", admin.listUsers)
    })
})
```

- [ ] **Step 5: Run tests**

```bash
go test ./internal/handler/... -run TestLLMStreamHandler -v
go test ./internal/llm/... -v
```

Expected: all PASS.

- [ ] **Step 6: Build check**

```bash
go build ./...
```

Expected: no errors.

- [ ] **Step 7: Commit**

```bash
git add internal/handler/llm.go internal/handler/llm_test.go internal/handler/router.go
git commit -m "feat: add POST /api/v1/llm/stream SSE endpoint"
```

---

## Task 7: Frontend useStream hook

**Files:**
- Create: `frontend/lib/useStream.ts`

- [ ] **Step 1: Implement `frontend/lib/useStream.ts`**

```typescript
import { useCallback, useRef, useState } from "react";
import { getToken } from "./auth";

interface StreamOptions {
  operation?: string;
  system?: string;
}

interface StreamState {
  text: string;
  loading: boolean;
  error: string | null;
}

/**
 * useStream — sends a prompt to POST /api/v1/llm/stream and accumulates
 * the SSE token stream into `text`. Call `stream(prompt)` to start.
 * Call `abort()` to cancel mid-stream.
 */
export function useStream(opts: StreamOptions = {}) {
  const [state, setState] = useState<StreamState>({ text: "", loading: false, error: null });
  const abortRef = useRef<AbortController | null>(null);

  const abort = useCallback(() => {
    abortRef.current?.abort();
    setState((s) => ({ ...s, loading: false }));
  }, []);

  const stream = useCallback(
    async (prompt: string) => {
      abortRef.current?.abort();
      const controller = new AbortController();
      abortRef.current = controller;

      setState({ text: "", loading: true, error: null });

      try {
        const resp = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/api/v1/llm/stream`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${getToken() ?? ""}`,
          },
          body: JSON.stringify({
            operation: opts.operation ?? "chat",
            system: opts.system ?? "",
            prompt,
          }),
          signal: controller.signal,
        });

        if (!resp.ok) {
          const msg = await resp.text();
          setState({ text: "", loading: false, error: msg || "Stream failed" });
          return;
        }

        const reader = resp.body?.getReader();
        if (!reader) {
          setState({ text: "", loading: false, error: "No response body" });
          return;
        }

        const decoder = new TextDecoder();
        let buffer = "";

        while (true) {
          const { done, value } = await reader.read();
          if (done) break;

          buffer += decoder.decode(value, { stream: true });
          const lines = buffer.split("\n");
          buffer = lines.pop() ?? "";

          for (const line of lines) {
            if (!line.startsWith("data: ")) continue;
            const token = line.slice(6);
            if (token === "[DONE]") {
              setState((s) => ({ ...s, loading: false }));
              return;
            }
            if (token === "[ERROR]") {
              setState((s) => ({ ...s, loading: false, error: "LLM error" }));
              return;
            }
            setState((s) => ({ ...s, text: s.text + token }));
          }
        }

        setState((s) => ({ ...s, loading: false }));
      } catch (err: unknown) {
        if (err instanceof Error && err.name === "AbortError") return;
        setState({ text: "", loading: false, error: err instanceof Error ? err.message : "Unknown error" });
      }
    },
    [opts.operation, opts.system]
  );

  return { ...state, stream, abort };
}
```

- [ ] **Step 2: Verify TypeScript types**

```bash
cd frontend && npx tsc --noEmit
```

Expected: no errors in `lib/useStream.ts`.

- [ ] **Step 3: Usage example (in a component)**

To verify the hook compiles correctly, add a quick smoke test in a page comment:

```typescript
// Usage in a component:
// const { text, loading, error, stream, abort } = useStream({ operation: "research" });
// <button onClick={() => stream("Summarize the latest trends in AI")}>Ask</button>
// <pre>{text}</pre>
```

- [ ] **Step 4: Commit**

```bash
cd ..  # back to repo root
git add frontend/lib/useStream.ts
git commit -m "feat(frontend): add useStream() hook for SSE LLM streaming"
```

---

## Task 8: Final integration check

- [ ] **Step 1: Run full backend test suite**

```bash
cd backend && go test ./... -count=1 -v 2>&1 | tail -30
```

Expected: all PASS.

- [ ] **Step 2: Run frontend type check**

```bash
cd frontend && npx tsc --noEmit
```

Expected: no errors.

- [ ] **Step 3: Verify the stream endpoint is reachable in local dev**

```bash
# Start local stack
make up

# Wait for backend to start (check health)
curl -s http://localhost:8080/health | jq

# Register a test user and get a token
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"test@example.com","password":"password123"}' | jq -r '.token')

# Stream a prompt
curl -s -N \
  -X POST http://localhost:8080/api/v1/llm/stream \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"operation":"chat","prompt":"Say hello in 5 words"}'
```

Expected: SSE events printed to terminal, ending with `data: [DONE]`.

- [ ] **Step 4: Final commit**

```bash
git add .
git commit -m "chore: verify LLM streaming integration complete" --allow-empty
```
