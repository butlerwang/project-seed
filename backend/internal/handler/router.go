package handler

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/butlerwang/project-seed/backend/internal/config"
	"github.com/butlerwang/project-seed/backend/internal/middleware"
	"github.com/butlerwang/project-seed/backend/internal/repository"
	"github.com/butlerwang/project-seed/backend/internal/service"
	"github.com/butlerwang/project-seed/backend/internal/storage"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func NewRouter(cfg config.Config, svc *service.Services, repos repository.Repository, store storage.Storage) http.Handler {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	r := chi.NewRouter()
	r.Use(middleware.StructuredLogger(log))
	r.Use(chimw.Recoverer)
	r.Use(middleware.RateLimit(cfg.RateLimitRPS, cfg.RateLimitBurst))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	r.Get("/health", handleHealth)

	var auth *authHandler
	var admin *adminHandler
	if svc != nil {
		auth = &authHandler{
			svc:         svc,
			frontendURL: cfg.FrontendURL,
			googleCfg:   googleOAuthConfig(cfg),
		}
		admin = &adminHandler{repos: repos}
	}

	var upload *uploadHandler
	if store != nil {
		upload = &uploadHandler{store: store}
	}
	webhook := &webhookHandler{secret: cfg.StripeWebhookSecret}
	var llmH *llmHandler
	if svc != nil && svc.LLM != nil {
		llmH = &llmHandler{svc: svc}
	}

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/webhooks/stripe", webhook.stripe)

		if auth != nil {
			r.Post("/auth/register", auth.register)
			r.Post("/auth/login", auth.login)
			r.Post("/auth/refresh", auth.refresh)
			r.Post("/auth/logout", auth.logout)
			r.Get("/auth/verify-email", auth.verifyEmail)
			r.Post("/auth/forgot-password", auth.forgotPassword)
			r.Post("/auth/reset-password", auth.resetPassword)
			r.Get("/auth/google", auth.googleRedirect)
			r.Get("/auth/google/callback", auth.googleCallback)
		}

		if upload != nil && auth == nil {
			r.Post("/files", upload.upload)
			r.Get("/files/{key}", upload.download)
		}

		if auth != nil && admin != nil {
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireAuth(svc.Auth))
				r.Get("/auth/me", auth.me)
				if upload != nil {
					r.Post("/files", upload.upload)
					r.Get("/files/{key}", upload.download)
				}
				if llmH != nil {
					r.Post("/llm/stream", llmH.stream)
				}

				r.Group(func(r chi.Router) {
					r.Use(middleware.RequireAdmin())
					r.Get("/admin/users", admin.listUsers)
				})
			})
		}
	})

	return r
}
