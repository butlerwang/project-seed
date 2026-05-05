package handler

import (
	"net/http"

	"github.com/butlerwang/project-seed/backend/internal/config"
	"github.com/butlerwang/project-seed/backend/internal/middleware"
	"github.com/butlerwang/project-seed/backend/internal/repository"
	"github.com/butlerwang/project-seed/backend/internal/service"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func NewRouter(cfg config.Config, svc *service.Services, repos repository.Repository) http.Handler {
	r := chi.NewRouter()
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	r.Get("/health", handleHealth)

	if svc == nil {
		return r
	}

	auth := &authHandler{svc: svc}
	admin := &adminHandler{repos: repos}

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/register", auth.register)
		r.Post("/auth/login", auth.login)
		r.Post("/auth/refresh", auth.refresh)
		r.Post("/auth/logout", auth.logout)

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAuth(svc.Auth))
			r.Get("/auth/me", auth.me)

			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireAdmin())
				r.Get("/admin/users", admin.listUsers)
			})
		})
	})

	return r
}
