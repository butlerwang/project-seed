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
