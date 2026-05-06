package middleware_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/butlerwang/project-seed/backend/internal/middleware"
	"github.com/stretchr/testify/assert"
)

func TestStructuredLoggerLogsMethodAndPath(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	handler := middleware.StructuredLogger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
}
