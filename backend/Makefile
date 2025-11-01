.PHONY: help install dev build run-api run-worker db-setup db-migrate clean

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install: ## Install Go dependencies
	go mod download
	go mod tidy

dev: ## Run API server in development mode
	go run cmd/api/main.go

build: ## Build binaries
	@echo "Building API server..."
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/api cmd/api/main.go
	@echo "Building worker..."
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/worker cmd/worker/main.go
	@echo "✅ Build complete"

run-api: build ## Run API server
	./bin/api

run-worker: build ## Run background worker
	./bin/worker

db-setup: ## Setup MySQL database
	mysql -u root -p < scripts/schema.sql
	@echo "✅ Database setup complete"

db-migrate: ## Run database migrations (if needed)
	@echo "Running migrations..."
	# Add migration tool if needed

clean: ## Clean build artifacts
	rm -rf bin/
	go clean

test: ## Run tests
	go test -v ./...

lint: ## Run linter
	golangci-lint run
