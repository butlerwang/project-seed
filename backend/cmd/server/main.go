package main

import (
	"log"
	"net/http"

	"github.com/butlerwang/project-seed/backend/internal/config"
	"github.com/butlerwang/project-seed/backend/internal/handler"
)

func main() {
	cfg := config.Load()
	router := handler.NewRouter(cfg, nil, nil)
	log.Printf("listening on %s", cfg.Address)
	if err := http.ListenAndServe(cfg.Address, router); err != nil {
		log.Fatal(err)
	}
}
