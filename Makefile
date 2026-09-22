.DEFAULT_GOAL := help

.PHONY: help fmt vet test tidy

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
