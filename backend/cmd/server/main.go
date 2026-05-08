package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/butlerwang/project-seed/backend/internal/config"
	"github.com/butlerwang/project-seed/backend/internal/handler"
	"github.com/butlerwang/project-seed/backend/internal/repository"
	"github.com/butlerwang/project-seed/backend/internal/service"
	"github.com/butlerwang/project-seed/backend/internal/storage"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

func main() {
	cfg := config.Load()
	if err := cfg.ValidateServerStartup(); err != nil {
		log.Fatal(err)
	}

	var repos repository.Repository = repository.NewMemoryRepository()
	if cfg.DatabaseURL != "" {
		db, err := openDB(cfg.DatabaseURL)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
		repos = repository.NewPostgresRepository(db)
	} else {
		log.Print("DATABASE_URL unset — using in-memory repository")
	}

	svc := service.New(cfg, repos)
	var store storage.Storage
	if cfg.StorageEndpoint != "" && cfg.StorageAccess != "" {
		s3store, err := storage.NewS3Storage(
			context.Background(),
			cfg.StorageEndpoint,
			cfg.StorageAccess,
			cfg.StorageSecret,
			cfg.StorageBucket,
		)
		if err != nil {
			slog.Error("failed to init storage", "err", err)
			os.Exit(1)
		}
		store = s3store
	} else {
		store = storage.NewMemoryStorage()
	}

	router := handler.NewRouter(cfg, svc, repos, store)

	log.Printf("listening on %s (env=%s)", cfg.Address, cfg.Environment)
	if err := http.ListenAndServe(cfg.Address, router); err != nil {
		log.Fatal(err)
	}
}

func openDB(url string) (*sqlx.DB, error) {
	db, err := sqlx.Open("pgx", url)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	return db, db.Ping()
}
