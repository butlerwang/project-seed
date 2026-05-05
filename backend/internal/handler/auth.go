package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/butlerwang/project-seed/backend/internal/middleware"
	"github.com/butlerwang/project-seed/backend/internal/service"
)

type authHandler struct{ svc *service.Services }

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func (h *authHandler) register(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	result, err := h.svc.Auth.Register(body.Email, body.Password)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	setRefreshCookie(w, result.RefreshToken)
	writeJSON(w, http.StatusCreated, map[string]any{"token": result.Token, "user": result.User})
}

func (h *authHandler) login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	result, err := h.svc.Auth.Login(body.Email, body.Password)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	setRefreshCookie(w, result.RefreshToken)
	writeJSON(w, http.StatusOK, map[string]any{"token": result.Token, "user": result.User})
}

func (h *authHandler) refresh(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("refresh_token")
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "missing refresh token")
		return
	}
	result, err := h.svc.Auth.Refresh(c.Value)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "refresh token expired")
		return
	}
	setRefreshCookie(w, result.RefreshToken)
	writeJSON(w, http.StatusOK, map[string]any{"token": result.Token, "user": result.User})
}

func (h *authHandler) logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie("refresh_token"); err == nil {
		_ = h.svc.Auth.Logout(c.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: "refresh_token", Value: "", MaxAge: -1, HttpOnly: true, Path: "/"})
	w.WriteHeader(http.StatusNoContent)
}

func (h *authHandler) me(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromCtx(r.Context())
	writeJSON(w, http.StatusOK, map[string]any{
		"id":    claims.Subject,
		"email": claims.Email,
		"role":  claims.Role,
	})
}

func setRefreshCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		Expires:  time.Now().Add(30 * 24 * time.Hour),
	})
}
