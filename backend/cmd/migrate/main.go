package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/butlerwang/project-seed/backend/internal/config"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func main() {
	cfg := config.Load()
	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	dir := migrationsDir()
	goose.SetBaseFS(nil)
	if err := goose.Up(db, dir); err != nil {
		log.Fatal(err)
	}
	log.Println("migrations complete")
}

func migrationsDir() string {
	for _, d := range []string{"./migrations", "../migrations", "backend/migrations", "/app/migrations"} {
		if info, err := os.Stat(d); err == nil && info.IsDir() {
			return d
		}
	}
	log.Fatal("migrations directory not found")
	return ""
}
