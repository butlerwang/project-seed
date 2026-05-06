package llm_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/butlerwang/project-seed/backend/internal/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOllamaProviderStream(t *testing.T) {
	ndjson := "{\"response\":\"Hello\",\"done\":false}\n" +
		"{\"response\":\" world\",\"done\":false}\n" +
		"{\"response\":\"\",\"done\":true}\n"

	provider := &llm.OllamaProvider{
		BaseURL: "http://ollama.test",
		Client: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/x-ndjson"}},
				Body:       io.NopCloser(bytes.NewBufferString(ndjson)),
			}, nil
		})},
	}

	var buf bytes.Buffer
	err := provider.Stream(context.Background(), llm.Request{
		Model:      "test-model",
		UserPrompt: "say hello",
	}, &buf)
	require.NoError(t, err)
	assert.Equal(t, "Hello world", buf.String())
}
