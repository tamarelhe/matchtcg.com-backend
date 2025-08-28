.PHONY: help dev-up dev-down run test test-unit test-integration test-coverage lint build clean migrate-up migrate-down migrate-create docker-build docker-run

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development environment
dev-up: ## Start development environment (PostgreSQL + PostGIS)
	docker-compose up -d postgres
	@echo "Waiting for PostgreSQL to be ready..."
	@until docker-compose exec postgres pg_isready -U postgres -d matchtcg; do sleep 1; done
	@echo "Development environment is ready!"

dev-down: ## Stop development environment
	docker-compose down

dev-full: ## Start full development environment including Redis and MailHog
	docker-compose --profile dev --profile cache up -d
	@echo "Waiting for services to be ready..."
	@until docker-compose exec postgres pg_isready -U postgres -d matchtcg; do sleep 1; done
	@echo "Full development environment is ready!"

# Application
run: ## Run the application
	go run ./cmd/server

build: ## Build the application
	CGO_ENABLED=0 go build -a -installsuffix cgo -o bin/server ./cmd/server

clean: ## Clean build artifacts
	rm -rf bin/
	go clean -cache
	go clean -modcache

# Testing
test: ## Run all tests
	go test -v -race ./...

test-unit: ## Run unit tests only
	go test -v -race -short ./...

test-integration: ## Run integration tests
	docker-compose --profile test up -d postgres-test
	@echo "Waiting for test database to be ready..."
	@until docker-compose exec postgres-test pg_isready -U postgres -d matchtcg_test; do sleep 1; done
	DATABASE_URL="postgres://postgres:password@localhost:5433/matchtcg_test?sslmode=disable" \
	go test -v -race ./... -tags=integration
	docker-compose --profile test down

test-coverage: ## Run tests with coverage
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Code quality
lint: ## Run linter
	golangci-lint run

lint-fix: ## Run linter with auto-fix
	golangci-lint run --fix

# Database migrations
migrate-up: ## Apply database migrations
	@if [ ! -f "bin/migrate" ]; then \
		echo "Installing golang-migrate..."; \
		go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest; \
		mkdir -p bin; \
		cp $$(go env GOPATH)/bin/migrate bin/; \
	fi
	bin/migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down: ## Rollback database migrations
	@if [ ! -f "bin/migrate" ]; then \
		echo "Installing golang-migrate..."; \
		go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest; \
		mkdir -p bin; \
		cp $$(go env GOPATH)/bin/migrate bin/; \
	fi
	bin/migrate -path migrations -database "$(DATABASE_URL)" down

migrate-create: ## Create a new migration (usage: make migrate-create name=create_users_table)
	@if [ ! -f "bin/migrate" ]; then \
		echo "Installing golang-migrate..."; \
		go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest; \
		mkdir -p bin; \
		cp $$(go env GOPATH)/bin/migrate bin/; \
	fi
	@if [ -z "$(name)" ]; then \
		echo "Error: name parameter is required. Usage: make migrate-create name=create_users_table"; \
		exit 1; \
	fi
	bin/migrate create -ext sql -dir migrations $(name)

migrate-force: ## Force migration version (usage: make migrate-force version=1)
	@if [ ! -f "bin/migrate" ]; then \
		echo "Installing golang-migrate..."; \
		go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest; \
		mkdir -p bin; \
		cp $$(go env GOPATH)/bin/migrate bin/; \
	fi
	@if [ -z "$(version)" ]; then \
		echo "Error: version parameter is required. Usage: make migrate-force version=1"; \
		exit 1; \
	fi
	bin/migrate -path migrations -database "$(DATABASE_URL)" force $(version)

# Docker
docker-build: ## Build Docker image
	docker build -t matchtcg-backend .

docker-run: ## Run application in Docker
	docker run -p 8080:8080 --env-file .env matchtcg-backend

# Dependencies
deps: ## Download and verify dependencies
	go mod download
	go mod verify

deps-update: ## Update dependencies
	go get -u ./...
	go mod tidy

# Development utilities
fmt: ## Format code
	go fmt ./...

vet: ## Run go vet
	go vet ./...

# Environment setup
setup: ## Initial project setup
	@echo "Setting up MatchTCG Backend development environment..."
	@if [ ! -f ".env" ]; then \
		cp .env.example .env; \
		echo "Created .env file from .env.example"; \
		echo "Please update .env with your configuration"; \
	fi
	go mod download
	make dev-up
	@echo "Setup complete! Run 'make run' to start the application"

# Database utilities
db-reset: ## Reset database (drop and recreate)
	docker-compose exec postgres psql -U postgres -c "DROP DATABASE IF EXISTS matchtcg;"
	docker-compose exec postgres psql -U postgres -c "CREATE DATABASE matchtcg;"
	make migrate-up

db-shell: ## Connect to database shell
	docker-compose exec postgres psql -U postgres -d matchtcg

# Logs
logs: ## Show application logs
	docker-compose logs -f

logs-db: ## Show database logs
	docker-compose logs -f postgres

# Load environment variables
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Set default DATABASE_URL if not provided
DATABASE_URL ?= postgres://postgres:password@localhost:5432/matchtcg?sslmode=disable