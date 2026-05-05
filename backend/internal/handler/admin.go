package handler

import (
	"net/http"
	"strconv"

	"github.com/butlerwang/project-seed/backend/internal/repository"
)

type adminHandler struct{ repos repository.Repository }

func (h *adminHandler) listUsers(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	users, err := h.repos.ListUsers(limit, offset)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "failed to list users")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"users": users, "limit": limit, "offset": offset})
}
