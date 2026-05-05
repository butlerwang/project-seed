package llm

import (
	"context"
	"net/http"
	"time"

	"github.com/butlerwang/project-seed/backend/internal/config"
)

// capableOps lists operations that need the larger/more capable model.
var capableOps = map[string]bool{
	"research": true,
	"generate": true,
	"analyze":  true,
}

type Router struct {
	cfg    config.Config
	client *http.Client
}

func NewRouter(cfg config.Config) *Router {
	return &Router{cfg: cfg, client: &http.Client{Timeout: 90 * time.Second}}
}

// Chat sends a single-turn LLM request.
// Uses Anthropic if ANTHROPIC_API_KEY is set, otherwise falls back to Ollama.
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

func (r *Router) pick(operation string) (Provider, string) {
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
