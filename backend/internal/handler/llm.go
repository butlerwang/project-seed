package handler

import (
	"encoding/json"
	"fmt"
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

	flusher, _ := w.(http.Flusher)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	if flusher != nil {
		flusher.Flush()
	}

	sw := &sseWriter{w: w, flusher: flusher}
	if err := h.svc.LLM.Stream(r.Context(), req.Operation, req.System, req.Prompt, sw); err != nil {
		_, _ = fmt.Fprint(w, "data: [ERROR]\n\n")
		if flusher != nil {
			flusher.Flush()
		}
		return
	}
	_, _ = fmt.Fprint(w, "data: [DONE]\n\n")
	if flusher != nil {
		flusher.Flush()
	}
}

type sseWriter struct {
	w       http.ResponseWriter
	flusher http.Flusher
}

func (sw *sseWriter) Write(p []byte) (int, error) {
	n, err := fmt.Fprintf(sw.w, "data: %s\n\n", string(p))
	if sw.flusher != nil {
		sw.flusher.Flush()
	}
	return n, err
}
