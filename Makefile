.PHONY: up down dev test build logs help docker-build docker-push docker-publish docker-verify

REGISTRY = africa-south1-docker.pkg.dev/kanso-496617/kanso-repo
TAG = latest

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

docker-build: ## Build all images for GCP Artifact Registry
	docker build -t $(REGISTRY)/api:$(TAG) -f backend/Dockerfile --target runtime backend
	docker build -t $(REGISTRY)/scheduler:$(TAG) -f scheduler/Dockerfile scheduler
	docker build -t $(REGISTRY)/nlp:$(TAG) -f nlp-service/Dockerfile nlp-service
	docker build -t $(REGISTRY)/chromedp:$(TAG) -f infra/chromedp/Dockerfile infra/chromedp

docker-push: ## Push all images to GCP Artifact Registry
	docker push $(REGISTRY)/api:$(TAG)
	docker push $(REGISTRY)/scheduler:$(TAG)
	docker push $(REGISTRY)/nlp:$(TAG)
	docker push $(REGISTRY)/chromedp:$(TAG)

docker-publish: docker-build docker-push ## Build + Push all images
	@echo "Published: $(REGISTRY)/{api,scheduler,nlp,chromedp}:$(TAG)"

docker-verify: ## Verify images exist in the registry
	@echo "Checking images in $(REGISTRY):"
	@for img in api scheduler nlp chromedp; do \
		docker manifest inspect $(REGISTRY)/$$img:$(TAG) > /dev/null 2>&1 && \
		echo "  ✓ $$img:$(TAG)" || echo "  ✗ $$img:$(TAG)"; \
	done
