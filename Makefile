.PHONY: help build test test-coverage lint clean migrate-up migrate-down docker-build docker-run dev deps

# Default target
help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Build commands
build: ## Build the application
	@echo "Building application..."
	@go build -o bin/fowergram cmd/server/main.go

build-linux: ## Build for Linux
	@echo "Building for Linux..."
	@GOOS=linux GOARCH=amd64 go build -o bin/fowergram-linux cmd/server/main.go

build-windows: ## Build for Windows
	@echo "Building for Windows..."
	@GOOS=windows GOARCH=amd64 go build -o bin/fowergram.exe cmd/server/main.go

build-all: build build-linux build-windows ## Build for all platforms

# Development commands
dev: ## Run the application in development mode
	@echo "Starting development server..."
	@air -c .air.toml || go run cmd/server/main.go

run: ## Run the application
	@echo "Starting application..."
	@go run cmd/server/main.go

# Dependency management
deps: ## Download and verify dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod verify

deps-update: ## Update all dependencies
	@echo "Updating dependencies..."
	@go get -u ./...
	@go mod tidy

deps-clean: ## Clean module cache
	@echo "Cleaning module cache..."
	@go clean -modcache

# Testing commands
test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

test-short: ## Run tests with short flag
	@echo "Running short tests..."
	@go test -short -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-race: ## Run tests with race detector
	@echo "Running tests with race detector..."
	@go test -race -v ./...

test-bench: ## Run benchmark tests
	@echo "Running benchmark tests..."
	@go test -bench=. -benchmem ./...

# Code quality commands
lint: ## Run linters
	@echo "Running linters..."
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	@golangci-lint run

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

# Security commands
security: ## Run security checks
	@echo "Running security checks..."
	@which gosec > /dev/null || (echo "Installing gosec..." && go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest)
	@gosec ./...

# Database commands
migrate-install: ## Install migrate tool
	@echo "Installing migrate tool..."
	@which migrate > /dev/null || go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

migrate-up: migrate-install ## Run database migrations
	@echo "Running database migrations..."
	@migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down: migrate-install ## Rollback database migrations
	@echo "Rolling back database migrations..."
	@migrate -path migrations -database "$(DATABASE_URL)" down

migrate-force: migrate-install ## Force migration version
	@echo "Forcing migration version $(VERSION)..."
	@migrate -path migrations -database "$(DATABASE_URL)" force $(VERSION)

migrate-create: migrate-install ## Create new migration (usage: make migrate-create NAME=migration_name)
	@echo "Creating migration: $(NAME)"
	@migrate create -ext sql -dir migrations -seq $(NAME)

# Docker commands
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t fowergram-backend:latest .

docker-build-dev: ## Build Docker image for development
	@echo "Building development Docker image..."
	@docker build -f Dockerfile.dev -t fowergram-backend:dev .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	@docker run -p 8000:8000 --env-file .env fowergram-backend:latest

docker-push: ## Push Docker image to registry
	@echo "Pushing Docker image..."
	@docker push fowergram-backend:latest

# Docker Compose commands
compose-up: ## Start all services with docker-compose
	@echo "Starting services with docker-compose..."
	@docker-compose up -d

compose-down: ## Stop all services
	@echo "Stopping services..."
	@docker-compose down

compose-logs: ## View logs
	@echo "Viewing logs..."
	@docker-compose logs -f

compose-rebuild: ## Rebuild and restart services
	@echo "Rebuilding services..."
	@docker-compose down
	@docker-compose build --no-cache
	@docker-compose up -d

# GraphQL commands
graphql-generate: ## Generate GraphQL code
	@echo "Generating GraphQL code..."
	@which gqlgen > /dev/null || go install github.com/99designs/gqlgen@latest
	@go run github.com/99designs/gqlgen generate

graphql-introspect: ## Generate GraphQL introspection
	@echo "Generating GraphQL introspection..."
	@curl -X POST -H "Content-Type: application/json" -d '{"query":"query IntrospectionQuery { __schema { queryType { name } mutationType { name } subscriptionType { name } types { ...FullType } directives { name description locations args { ...InputValue } } } } fragment FullType on __Type { kind name description fields(includeDeprecated: true) { name description args { ...InputValue } type { ...TypeRef } isDeprecated deprecationReason } inputFields { ...InputValue } interfaces { ...TypeRef } enumValues(includeDeprecated: true) { name description isDeprecated deprecationReason } possibleTypes { ...TypeRef } } fragment InputValue on __InputValue { name description type { ...TypeRef } defaultValue } fragment TypeRef on __Type { kind name ofType { kind name ofType { kind name ofType { kind name ofType { kind name ofType { kind name ofType { kind name ofType { kind name } } } } } } } }"}' http://localhost:8000/graphql | jq . > schema.json

# API Documentation commands
docs-serve: ## Serve API documentation locally
	@echo "Serving API documentation at http://localhost:3000/docs"
	@python3 -m http.server 3000 --directory api || python -m SimpleHTTPServer 3000

docs-validate: ## Validate OpenAPI specification
	@echo "Validating OpenAPI specification..."
	@which swagger > /dev/null || (echo "Installing swagger..." && go install github.com/go-swagger/go-swagger/cmd/swagger@latest)
	@swagger validate api/openapi.yaml

docs-generate: ## Generate API docs from code annotations
	@echo "Generating API documentation from code..."
	@go run scripts/generate-api-docs.go

docs-update: docs-generate docs-validate ## Update and validate API documentation
	@echo "API documentation updated and validated successfully!"

docs-open: ## Open API documentation in browser
	@echo "Opening API documentation..."
	@which open > /dev/null && open http://localhost:3000/docs/stoplight.html || echo "Please visit http://localhost:3000/docs/stoplight.html"

# Monitoring commands
prometheus-up: ## Start Prometheus
	@echo "Starting Prometheus..."
	@docker run -d -p 9090:9090 -v $(PWD)/monitoring/prometheus.yml:/etc/prometheus/prometheus.yml prom/prometheus

grafana-up: ## Start Grafana
	@echo "Starting Grafana..."
	@docker run -d -p 3001:3000 grafana/grafana

jaeger-up: ## Start Jaeger
	@echo "Starting Jaeger..."
	@docker run -d -p 16686:16686 -p 14268:14268 jaegertracing/all-in-one:latest

# Load testing commands
load-test: ## Run load tests
	@echo "Running load tests..."
	@which hey > /dev/null || (echo "Installing hey..." && go install github.com/rakyll/hey@latest)
	@hey -n 1000 -c 10 http://localhost:8000/health

load-test-graphql: ## Run GraphQL load tests
	@echo "Running GraphQL load tests..."
	@hey -n 1000 -c 10 -m POST -H "Content-Type: application/json" -d '{"query":"query { __typename }"}' http://localhost:8000/graphql

# Cleanup commands
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@go clean -cache -testcache -modcache

clean-docker: ## Clean Docker artifacts
	@echo "Cleaning Docker artifacts..."
	@docker system prune -f
	@docker volume prune -f

# Release commands
release: clean test lint build ## Prepare release
	@echo "Release prepared successfully"

tag: ## Create and push git tag (usage: make tag VERSION=v1.0.0)
	@echo "Creating tag $(VERSION)..."
	@git tag -a $(VERSION) -m "Release $(VERSION)"
	@git push origin $(VERSION)

# Environment setup
setup: ## Set up development environment
	@echo "Setting up development environment..."
	@go mod download
	@which air > /dev/null || go install github.com/cosmtrek/air@latest
	@which golangci-lint > /dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@which migrate > /dev/null || go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@which gqlgen > /dev/null || go install github.com/99designs/gqlgen@latest
	@cp env.example .env
	@echo "Development environment setup complete!"

# Performance profiling
profile-cpu: ## Run CPU profiling
	@echo "Running CPU profiling..."
	@go test -cpuprofile=cpu.prof -bench=. ./...
	@go tool pprof cpu.prof

profile-mem: ## Run memory profiling
	@echo "Running memory profiling..."
	@go test -memprofile=mem.prof -bench=. ./...
	@go tool pprof mem.prof

# Documentation
docs: ## Generate documentation
	@echo "Generating documentation..."
	@which godoc > /dev/null || go install golang.org/x/tools/cmd/godoc@latest
	@echo "Documentation server starting at http://localhost:6060"
	@godoc -http=:6060

# Version info
version: ## Show version information
	@echo "Go version: $(shell go version)"
	@echo "Git commit: $(shell git rev-parse HEAD)"
	@echo "Build date: $(shell date)"

# Default environment variables
DATABASE_URL ?= postgres://fowergram:password@localhost:5432/fowergram?sslmode=disable
REDIS_URL ?= redis://localhost:6379 