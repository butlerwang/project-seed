NAME ?= project-seed

.PHONY: up down logs test migrate seed db-shell minio-shell rename help

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

up: ## Start all services
	docker compose up --build -d

down: ## Stop all services
	docker compose down

logs: ## Tail logs
	docker compose logs -f

test: ## Run backend tests
	cd backend && go test ./...

migrate: ## Run DB migrations (requires DATABASE_URL)
	cd backend && go run ./cmd/migrate

seed: ## Seed DB with admin user (requires DATABASE_URL)
	cd backend && go run ./cmd/seed

db-shell: ## Open psql shell
	docker compose exec postgres psql -U seed seed

minio-shell: ## Open MinIO mc shell
	docker compose exec minio mc alias set local http://localhost:9000 seed seed-dev-secret

rename: ## Rename project-seed → NAME (e.g. make rename NAME=myapp)
	@echo "Renaming project-seed → $(NAME)"
	find . -type f \( -name "*.go" -o -name "*.toml" -o -name "*.yml" -o -name "*.yaml" -o -name "*.json" -o -name "*.mjs" -o -name "*.ts" -o -name "*.tsx" -o -name "*.md" -o -name ".env.example" -o -name "Makefile" -o -name "Dockerfile" \) \
		-not -path "*/node_modules/*" -not -path "*/.git/*" \
		| xargs sed -i.bak "s/project-seed/$(NAME)/g"
	find . -name "*.bak" -delete
	@echo "Done — review with: git diff"
