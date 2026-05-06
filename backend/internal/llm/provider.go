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

type Provider interface {
	Chat(ctx context.Context, req Request) (string, error)
	Stream(ctx context.Context, req Request, w io.Writer) error
}
