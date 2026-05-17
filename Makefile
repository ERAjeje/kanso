.PHONY: up down dev test build logs help

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

up: ## Start infrastructure (CouchDB, API, Traefik)
	cd infra && docker compose up -d

down: ## Stop infrastructure
	cd infra && docker compose down

dev: ## Start frontend dev server
	cd frontend && pnpm dev

test: ## Run all tests
	cd frontend && pnpm test
	cd backend && go test ./...

build: ## Build frontend and backend
	cd frontend && pnpm build
	cd backend && go build -o bin/kanso-api ./cmd/kanso-api

logs: ## Follow infrastructure logs
	cd infra && docker compose logs -f
