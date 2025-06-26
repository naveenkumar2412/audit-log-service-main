# Makefile for Audit Log Service

.PHONY: help build run test clean docker-build docker-run docker-stop migrate deps lint fmt vet security install-tools

# Variables
BINARY_NAME=audit-log-service
MAIN_PATH=./cmd/api/main.go
DOCKER_IMAGE=audit-log-service:latest
DOCKER_COMPOSE_FILE=./docker/docker-compose.yml

# Default target
help: ## Show available commands
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Development
build: ## Build the application
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_NAME) $(MAIN_PATH)

run: build ## Build and run the application
	@echo "Running $(BINARY_NAME)..."
	@./$(BINARY_NAME)

dev: ## Run the application in development mode with hot reload (requires air)
	@echo "Starting development server..."
	@air -c .air.toml

# Testing
test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-race: ## Run tests with race detection
	@echo "Running tests with race detection..."
	@go test -race ./...

benchmark: ## Run benchmarks
	@echo "Running benchmarks..."
	@go test -bench=. ./...

# Database
migrate-up: ## Run database migrations up
	@echo "Running database migrations up..."
	@./scripts/migrate.sh up

migrate-down: ## Run database migrations down
	@echo "Running database migrations down..."
	@./scripts/migrate.sh down

migrate-create: ## Create a new migration file (requires NAME variable)
	@echo "Creating migration file..."
	@migrate create -ext sql -dir internal/database/migrations -seq $(NAME)

# Dependencies
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

deps-upgrade: ## Upgrade dependencies
	@echo "Upgrading dependencies..."
	@go get -u ./...
	@go mod tidy

# Code quality
lint: ## Run linters
	@echo "Running linters..."
	@golangci-lint run ./...

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

security: ## Run security checks
	@echo "Running security checks..."
	@gosec ./...

# Docker
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -f docker/Dockerfile -t $(DOCKER_IMAGE) .

docker-run: ## Run application in Docker
	@echo "Running application in Docker..."
	@docker-compose -f $(DOCKER_COMPOSE_FILE) up -d

docker-stop: ## Stop Docker containers
	@echo "Stopping Docker containers..."
	@docker-compose -f $(DOCKER_COMPOSE_FILE) down

docker-logs: ## View Docker logs
	@docker-compose -f $(DOCKER_COMPOSE_FILE) logs -f

docker-clean: ## Clean Docker images and containers
	@echo "Cleaning Docker images and containers..."
	@docker-compose -f $(DOCKER_COMPOSE_FILE) down -v
	@docker rmi $(DOCKER_IMAGE) 2>/dev/null || true

# Tools installation
install-tools: ## Install development tools
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install honnef.co/go/tools/cmd/staticcheck@latest
	@go install github.com/air-verse/air@latest
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Production
build-prod: ## Build for production
	@echo "Building for production..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o $(BINARY_NAME) $(MAIN_PATH)

# Cleanup
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@go clean
	@rm -f $(BINARY_NAME)
	@rm -f coverage.out coverage.html

clean-all: clean ## Clean everything including Docker
	@echo "Cleaning everything..."
	@docker-compose -f $(DOCKER_COMPOSE_FILE) down -v --remove-orphans 2>/dev/null || true
	@docker rmi $(DOCKER_IMAGE) 2>/dev/null || true

# Configuration
config-example: ## Copy example config
	@echo "Copying example config..."
	@cp configs/config.example.yaml configs/config.yaml

# Health check
health: ## Check application health
	@echo "Checking application health..."
	@curl -f http://localhost:9025/health || echo "Service not responding"

# API testing
test-api: ## Test API endpoints
	@echo "Testing API endpoints..."
	@curl -X POST http://localhost:9025/api/v1/audit \
		-H "Content-Type: application/json" \
		-H "X-API-Key: api-key-1" \
		-d '{"tenant_id":"test","user_id":"user1","resource":"users","event":"CREATE","method":"POST","ip":"127.0.0.1","environment":"development"}'