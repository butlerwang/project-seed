package handler

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/butlerwang/project-seed/backend/internal/config"
	"github.com/butlerwang/project-seed/backend/internal/middleware"
	"github.com/butlerwang/project-seed/backend/internal/service"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type authHandler struct {
	svc         *service.Services
	frontendURL string
	googleCfg   *oauth2.Config
}

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

func (h *authHandler) verifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		writeErr(w, http.StatusBadRequest, "missing token")
		return
	}
	if err := h.svc.Auth.VerifyEmail(token); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	http.Redirect(w, r, h.frontendURL+"/login?verified=1", http.StatusFound)
}

func (h *authHandler) forgotPassword(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := h.svc.Auth.ForgotPassword(body.Email); err != nil {
		writeErr(w, http.StatusInternalServerError, "failed to send reset email")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *authHandler) resetPassword(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := h.svc.Auth.ResetPassword(body.Token, body.Password); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func googleOAuthConfig(cfg config.Config) *oauth2.Config {
	if cfg.GoogleClientID == "" || cfg.GoogleClientSecret == "" {
		return nil
	}
	return &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  strings.TrimRight(cfg.AppURL, "/") + "/api/v1/auth/google/callback",
		Scopes:       []string{"openid", "email"},
		Endpoint:     google.Endpoint,
	}
}

func (h *authHandler) googleRedirect(w http.ResponseWriter, r *http.Request) {
	if h.googleCfg == nil {
		writeErr(w, http.StatusNotImplemented, "google oauth not configured")
		return
	}
	state := generateState()
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   300,
	})
	http.Redirect(w, r, h.googleCfg.AuthCodeURL(state), http.StatusFound)
}

func (h *authHandler) googleCallback(w http.ResponseWriter, r *http.Request) {
	if h.googleCfg == nil {
		writeErr(w, http.StatusNotImplemented, "google oauth not configured")
		return
	}
	cookie, err := r.Cookie("oauth_state")
	if err != nil || cookie.Value != r.URL.Query().Get("state") {
		writeErr(w, http.StatusBadRequest, "invalid oauth state")
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "oauth_state", Value: "", MaxAge: -1, Path: "/"})

	token, err := h.googleCfg.Exchange(r.Context(), r.URL.Query().Get("code"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "oauth exchange failed")
		return
	}
	email, err := fetchGoogleEmail(r.Context(), h.googleCfg, token)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "failed to fetch user info")
		return
	}
	result, err := h.svc.Auth.FindOrCreateGoogleUser(email)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "failed to create user")
		return
	}
	setRefreshCookie(w, result.RefreshToken)
	http.Redirect(w, r, h.frontendURL+"/dashboard?token="+result.Token, http.StatusFound)
}

func fetchGoogleEmail(ctx context.Context, cfg *oauth2.Config, token *oauth2.Token) (string, error) {
	client := cfg.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var info struct {
		Email string `json:"email"`
	}
	if err := json.Unmarshal(data, &info); err != nil {
		return "", err
	}
	if info.Email == "" {
		return "", errors.New("no email in google response")
	}
	return info.Email, nil
}

func generateState() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
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
