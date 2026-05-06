package main

import (
	"log"

	"github.com/butlerwang/project-seed/backend/internal/config"
	"github.com/butlerwang/project-seed/backend/internal/repository"
	"github.com/butlerwang/project-seed/backend/internal/service"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

func main() {
	cfg := config.Load()
	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	db, err := sqlx.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	repos := repository.NewPostgresRepository(db)
	authSvc := service.NewAuthService(cfg, repos, nil)

	email := "admin@example.com"
	if cfg.AdminEmail != "" {
		email = cfg.AdminEmail
	}
	result, err := authSvc.Register(email, "changeme123!")
	if err != nil {
		log.Printf("seed skipped (user may already exist): %v", err)
		return
	}
	log.Printf("seeded admin user: %s (id=%s)", result.User.Email, result.User.ID)
}
