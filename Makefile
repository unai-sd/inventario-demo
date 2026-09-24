# Development tasks. Run `make help` to list them.

DATABASE_URL ?= postgres://inventario:inventario@localhost:5433/inventario?sslmode=disable
PORT ?= 8080
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
export DATABASE_URL PORT

.PHONY: help deps db-up db-down migrate seed test test-integration lint fmt build run

help: ## List the available targets
	@grep -E '^[a-z-]+:.*## ' $(MAKEFILE_LIST) | awk -F ':.*## ' '{printf "  %-18s %s\n", $$1, $$2}'

deps: ## Download the locked Go modules
	go mod download

db-up: ## Start PostgreSQL 16 with docker compose and wait until it is healthy
	docker compose up -d --wait db

db-down: ## Stop PostgreSQL and delete its data
	docker compose down -v

migrate: ## Apply pending SQL migrations to DATABASE_URL
	go run ./cmd/migrate

seed: ## Load demo warehouses, products and stock (needs db-up and migrate)
	docker compose exec -T db psql -v ON_ERROR_STOP=1 -U inventario -d inventario < scripts/seed.sql

test: ## Unit tests (no database needed)
	go test ./...

test-integration: ## Integration tests against DATABASE_URL (needs db-up)
	go test -tags integration -p 1 -count=1 ./...

lint: ## go vet plus gofmt check
	go vet ./...
	go vet -tags integration ./...
	@test -z "$$(gofmt -l .)" || (echo "gofmt needed in:"; gofmt -l .; exit 1)

fmt: ## Format the code
	gofmt -w .

build: ## Build the binaries into bin/
	go build -ldflags "-X main.version=$(VERSION)" -o bin/ ./cmd/...

run: ## Run the API on PORT (default 8080)
	go run -ldflags "-X main.version=$(VERSION)" ./cmd/api
