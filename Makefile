# OpenAx Makefile

# Variables
BINARY_NAME=openax
MAIN_PATH=.
PKG_LIST=$$(go list ./... | grep -v /vendor/)

.PHONY: help build test test-short test-verbose test-race test-cover clean install fmt vet lint run examples bench

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build targets
build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) $(MAIN_PATH)

install: ## Install the binary to $GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	go install $(MAIN_PATH)

clean: ## Clean build artifacts
	@echo "Cleaning..."
	go clean
	rm -f $(BINARY_NAME)
	rm -f coverage.txt coverage.html

# Test targets
test: ## Run all tests
	@echo "Running tests..."
	go test $(PKG_LIST)

test-short: ## Run tests in short mode (skip integration tests)
	@echo "Running tests in short mode..."
	go test -short $(PKG_LIST)

test-verbose: ## Run tests with verbose output
	@echo "Running tests with verbose output..."
	go test -v $(PKG_LIST)

test-race: ## Run tests with race detector
	@echo "Running tests with race detector..."
	go test -race $(PKG_LIST)

test-cover: ## Run tests with coverage
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.txt -covermode=atomic $(PKG_LIST)
	go tool cover -html=coverage.txt -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-integration: ## Run only integration tests
	@echo "Running integration tests..."
	go test -v -run Integration

bench: ## Run benchmarks
	@echo "Running benchmarks..."
	go test -bench=. -benchmem $(PKG_LIST)

# Code quality targets
fmt: ## Format code
	@echo "Formatting code..."
	go fmt $(PKG_LIST)

vet: ## Run go vet
	@echo "Running go vet..."
	go vet $(PKG_LIST)

lint: ## Run linter (requires golangci-lint)
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install it with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Development targets
run: build ## Build and run the CLI tool
	@echo "Running $(BINARY_NAME)..."
	./$(BINARY_NAME) --help

run-example: build ## Run example filtering
	@echo "Running example filtering..."
	./$(BINARY_NAME) -i testdata/specs/petstore.yaml --tags pet --format json | head -20

examples: ## Run all examples
	@echo "Running library examples..."
	cd examples/library && go run main.go
	@echo ""
	@echo "Running custom filter examples..."
	cd examples/custom-filter && go run main.go

# Validation targets
validate-specs: build ## Validate all test specs
	@echo "Validating test specifications..."
	@for spec in testdata/specs/*.yaml; do \
		echo "Validating $$spec..."; \
		./$(BINARY_NAME) --validate-only -i "$$spec" || echo "❌ $$spec failed"; \
	done

# Dependencies
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod download

deps-update: ## Update dependencies
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

# Release targets
check: test-short vet lint ## Run pre-commit checks
	@echo "✅ All checks passed!"

ci: deps test-race test-cover vet ## Run CI pipeline
	@echo "✅ CI pipeline completed!"

# Git hooks
install-hooks: ## Install git hooks
	@echo "Installing git hooks..."
	@echo '#!/bin/sh\nmake check' > .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "✅ Pre-commit hook installed!"

# Docker targets (future use)
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t openax:latest .

docker-test: ## Run tests in Docker
	@echo "Running tests in Docker..."
	docker run --rm -v $$(pwd):/app -w /app golang:1.21 make test

# Documentation
docs: ## Generate documentation
	@echo "Generating documentation..."
	@echo "API documentation available at: https://pkg.go.dev/github.com/imtanmoy/openax"
	@echo "Repository: https://github.com/imtanmoy/openax"

# Show project status
status: ## Show project status
	@echo "Project Status:"
	@echo "==============="
	@echo "Binary: $(BINARY_NAME)"
	@echo "Go version: $$(go version)"
	@echo "Git branch: $$(git branch --show-current 2>/dev/null || echo 'unknown')"
	@echo "Git commit: $$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
	@echo ""
	@echo "Package structure:"
	@find pkg -name "*.go" | wc -l | xargs echo "  Go files in pkg/:"
	@find testdata -name "*.yaml" | wc -l | xargs echo "  Test specs:"
	@find examples -name "*.go" | wc -l | xargs echo "  Example files:"