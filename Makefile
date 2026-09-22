.DEFAULT_GOAL := help

COMPOSE_FILE := deploy/docker-compose.yml

.PHONY: help fmt vet test tidy up down clean ps logs proto-lint proto-fmt

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

fmt: ## Format Go source files
	go fmt ./...

vet: ## Run go vet static analysis
	go vet ./...

test: ## Run tests
	go test ./...

tidy: ## Tidy go.mod dependencies
	go mod tidy

up: ## Start local infrastructure (Docker Compose)
	docker compose -f $(COMPOSE_FILE) up -d --wait

down: ## Stop local infrastructure, keep volumes
	docker compose -f $(COMPOSE_FILE) down

clean: ## Stop local infrastructure and remove volumes
	docker compose -f $(COMPOSE_FILE) down -v

ps: ## Show local infrastructure container status
	docker compose -f $(COMPOSE_FILE) ps

logs: ## Follow container logs (use s=<service> for one service)
	docker compose -f $(COMPOSE_FILE) logs -f $(s)

proto-lint: ## Lint proto files
	go tool buf lint

proto-fmt: ## Format proto files
	go tool buf format -w
