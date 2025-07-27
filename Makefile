# kbVault Makefile

.PHONY: help build test clean dev lint fmt vet install

# Build configuration
BINARY_NAME=kbvault
BINARY_DIR=bin
MAIN_PATH=./cmd/kbvault
VERSION?=dev
COMMIT_HASH?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME?=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Go configuration
GOFLAGS=-ldflags="-X main.Version=${VERSION} -X main.CommitHash=${COMMIT_HASH} -X main.BuildTime=${BUILD_TIME}"

help: ## Show this help message
	@echo "kbVault - Knowledge Base Vault"
	@echo ""
	@echo "Available commands:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the binary
	@echo "Building kbVault..."
	@mkdir -p $(BINARY_DIR)
	@go build $(GOFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Binary built: $(BINARY_DIR)/$(BINARY_NAME)"

test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf $(BINARY_DIR)
	@rm -f coverage.out coverage.html
	@go clean

dev: ## Run in development mode
	@echo "Running in development mode..."
	@go run $(MAIN_PATH) --dev

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

install: build ## Install binary to GOPATH/bin
	@echo "Installing kbVault..."
	@cp $(BINARY_DIR)/$(BINARY_NAME) $(GOPATH)/bin/
	@echo "Installed: $(GOPATH)/bin/$(BINARY_NAME)"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

deps-update: ## Update dependencies
	@echo "Updating dependencies..."
	@go get -u ./...
	@go mod tidy

# Build for multiple platforms
build-all: ## Build for all platforms
	@echo "Building for all platforms..."
	@mkdir -p $(BINARY_DIR)
	@GOOS=linux GOARCH=amd64 go build $(GOFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	@GOOS=darwin GOARCH=amd64 go build $(GOFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	@GOOS=darwin GOARCH=arm64 go build $(GOFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	@GOOS=windows GOARCH=amd64 go build $(GOFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "Built binaries:"
	@ls -la $(BINARY_DIR)/

# Development shortcuts
run: ## Run the application
	@go run $(MAIN_PATH)

debug: ## Run with debug logging
	@KBVAULT_LOG_LEVEL=DEBUG go run $(MAIN_PATH)

# Initialize development environment
init-dev: ## Initialize development environment
	@echo "Initializing development environment..."
	@go mod download
	@mkdir -p configs
	@mkdir -p test/fixtures
	@echo "Development environment ready!"

# Example vault for testing
test-vault: ## Create example vault for testing
	@echo "Creating test vault..."
	@mkdir -p test/example-vault/notes
	@echo "---\ntitle: Example Note\ntags: [example]\n---\n\n# Example Note\n\nThis is an example note for testing." > test/example-vault/notes/example.md
	@echo "Test vault created: test/example-vault"
