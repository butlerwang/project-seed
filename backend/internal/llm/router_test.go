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

type mockProvider struct {
	chatResult   string
	streamTokens []string
}

func (m *mockProvider) Chat(_ context.Context, _ llm.Request) (string, error) {
	return m.chatResult, nil
}

func (m *mockProvider) Stream(_ context.Context, _ llm.Request, w io.Writer) error {
	for _, tok := range m.streamTokens {
		_, _ = io.WriteString(w, tok)
	}
	return nil
}

func TestRouterStreamWritesTokens(t *testing.T) {
	cfg := config.Config{}
	router := llm.NewRouterWithProvider(cfg, &mockProvider{streamTokens: []string{"Hello", " world"}})

	var buf bytes.Buffer
	err := router.Stream(context.Background(), "research", "system", "prompt", &buf)
	require.NoError(t, err)
	assert.Equal(t, "Hello world", buf.String())
}
