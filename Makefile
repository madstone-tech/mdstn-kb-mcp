# kbVault Makefile
.PHONY: help build test clean lint fmt vet deps dev-deps tools install

# Build information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

# Go build flags
LDFLAGS = -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildTime=$(BUILD_TIME)
BUILD_FLAGS = -ldflags "$(LDFLAGS)"

# Directories
BIN_DIR = bin
PKG_DIR = pkg
INTERNAL_DIR = internal
CMD_DIR = cmd

# Default target
help: ## Show this help message
	@echo "kbVault - High-performance Go knowledge management tool"
	@echo ""
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build targets
build: ## Build the kbvault binary
	@echo "Building kbvault..."
	@mkdir -p $(BIN_DIR)
	go build $(BUILD_FLAGS) -o $(BIN_DIR)/kbvault ./$(CMD_DIR)/kbvault

build-all: ## Build binaries for all platforms
	@echo "Building for all platforms..."
	@mkdir -p $(BIN_DIR)
	GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o $(BIN_DIR)/kbvault-linux-amd64 ./$(CMD_DIR)/kbvault
	GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS) -o $(BIN_DIR)/kbvault-darwin-amd64 ./$(CMD_DIR)/kbvault
	GOOS=darwin GOARCH=arm64 go build $(BUILD_FLAGS) -o $(BIN_DIR)/kbvault-darwin-arm64 ./$(CMD_DIR)/kbvault

# Development targets
dev: ## Build and run in development mode
	go run $(BUILD_FLAGS) ./$(CMD_DIR)/kbvault

install: ## Install kbvault to GOPATH/bin
	go install $(BUILD_FLAGS) ./$(CMD_DIR)/kbvault

# Testing targets
test: ## Run all tests
	go test -v -race ./...

test-coverage: ## Run tests with coverage
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

test-integration: ## Run integration tests (requires test environment)
	go test -v -race -tags=integration ./...

benchmark: ## Run benchmarks
	go test -v -bench=. -benchmem ./...

# Code quality targets
fmt: ## Format Go code
	go fmt ./...

vet: ## Run go vet
	go vet ./...

lint: ## Run golangci-lint (requires golangci-lint to be installed)
	golangci-lint run

check: fmt vet lint test ## Run all checks (format, vet, lint, test)

# Dependency management
deps: ## Download dependencies
	go mod download
	go mod tidy

dev-deps: ## Install development dependencies
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

tools: dev-deps ## Install all development tools

# Utility targets
clean: ## Clean build artifacts
	rm -rf $(BIN_DIR)
	rm -f coverage.out coverage.html

version: ## Show version information
	@echo "Version: $(VERSION)"
	@echo "Commit:  $(COMMIT)"
	@echo "Built:   $(BUILD_TIME)"

# Development workflow targets
setup: tools deps ## Set up development environment
	@echo "Development environment ready!"

quick: fmt vet test ## Quick development check (format, vet, test)

# CI targets (for GitHub Actions)
ci-test: ## CI test target
	go test -v -race -coverprofile=coverage.out ./...

ci-build: ## CI build target
	go build $(BUILD_FLAGS) -o $(BIN_DIR)/kbvault ./$(CMD_DIR)/kbvault

# Generate targets (for future use)
generate: ## Run go generate
	go generate ./...