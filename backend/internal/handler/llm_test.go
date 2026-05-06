package handler_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/butlerwang/project-seed/backend/internal/handler"
	"github.com/butlerwang/project-seed/backend/internal/llm"
	"github.com/butlerwang/project-seed/backend/internal/repository"
	"github.com/butlerwang/project-seed/backend/internal/service"
	"github.com/butlerwang/project-seed/backend/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockLLMProvider struct {
	tokens []string
}

func (m *mockLLMProvider) Chat(_ context.Context, _ llm.Request) (string, error) {
	return strings.Join(m.tokens, ""), nil
}

func (m *mockLLMProvider) Stream(_ context.Context, _ llm.Request, w io.Writer) error {
	for _, tok := range m.tokens {
		_, _ = io.WriteString(w, tok)
	}
	return nil
}

func TestLLMStreamHandlerStreams(t *testing.T) {
	repos := repository.NewMemoryRepository()
	svc := service.New(testConfig(), repos)
	svc.LLM = llm.NewRouterWithProvider(testConfig(), &mockLLMProvider{tokens: []string{"Hello", " world"}})

	authResult, err := svc.Auth.Register("stream@example.com", "password123")
	require.NoError(t, err)

	r := handler.NewRouter(testConfig(), svc, repos, storage.NewMemoryStorage())

	body, err := json.Marshal(map[string]string{
		"operation": "chat",
		"system":    "You are helpful.",
		"prompt":    "say hello",
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/llm/stream", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authResult.Token)
	rr := newFlushRecorder()
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

type flushRecorder struct {
	*httptest.ResponseRecorder
}

func newFlushRecorder() *flushRecorder {
	return &flushRecorder{ResponseRecorder: httptest.NewRecorder()}
}

func (r *flushRecorder) Flush() {}
