package main

import (
	"log"
	"net/http"

	"github.com/butlerwang/project-seed/backend/internal/config"
	"github.com/butlerwang/project-seed/backend/internal/handler"
	"github.com/butlerwang/project-seed/backend/internal/repository"
	"github.com/butlerwang/project-seed/backend/internal/service"
)

func main() {
	cfg := config.Load()

	var repos repository.Repository = repository.NewMemoryRepository()

	svc := service.New(cfg, repos)
	router := handler.NewRouter(cfg, svc, repos)

	log.Printf("listening on %s (env=%s)", cfg.Address, cfg.Environment)
	if err := http.ListenAndServe(cfg.Address, router); err != nil {
		log.Fatal(err)
	}
}
