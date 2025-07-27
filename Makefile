# kbVault Makefile
.PHONY: help build build-all test test-coverage test-integration benchmark clean lint fmt vet deps dev-deps tools install \
        dev version setup quick ci-test ci-build generate \
        docker-build docker-build-multi docker-run docker-shell docker-push docker-clean \
        release-dry release-snapshot release-check \
        compose-up compose-down compose-logs \
        security-scan security-deps vuln-check

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

# Docker targets
DOCKER_TAG ?= kbvault:latest
DOCKER_REGISTRY ?= ghcr.io/madstone-tech
DOCKER_IMAGE = $(DOCKER_REGISTRY)/kbvault
CONTAINER_ENGINE ?= $(shell command -v podman 2> /dev/null || echo docker)

docker-build: ## Build Docker image
	@echo "Building Docker image with $(CONTAINER_ENGINE)..."
	$(CONTAINER_ENGINE) build -t $(DOCKER_TAG) .

docker-build-multi: ## Build multi-architecture Docker images
	@echo "Building multi-architecture Docker images..."
	docker buildx build --platform linux/amd64,linux/arm64 -t $(DOCKER_TAG) .

docker-run: ## Run Docker container
	@echo "Running container with $(CONTAINER_ENGINE)..."
	$(CONTAINER_ENGINE) run --rm -it -p 8080:8080 -p 9090:9090 $(DOCKER_TAG)

docker-shell: ## Open shell in Docker container (not available with scratch image)
	@echo "Note: Shell not available in scratch-based image"
	@echo "Running container version instead..."
	$(CONTAINER_ENGINE) run --rm $(DOCKER_TAG) --version

docker-push: ## Push Docker image to registry
	@echo "Pushing Docker image to registry with $(CONTAINER_ENGINE)..."
	$(CONTAINER_ENGINE) tag $(DOCKER_TAG) $(DOCKER_IMAGE):$(VERSION)
	$(CONTAINER_ENGINE) tag $(DOCKER_TAG) $(DOCKER_IMAGE):latest
	$(CONTAINER_ENGINE) push $(DOCKER_IMAGE):$(VERSION)
	$(CONTAINER_ENGINE) push $(DOCKER_IMAGE):latest

docker-clean: ## Clean Docker images and containers
	@echo "Cleaning images and containers with $(CONTAINER_ENGINE)..."
	$(CONTAINER_ENGINE) system prune -f
	$(CONTAINER_ENGINE) rmi $(DOCKER_TAG) 2>/dev/null || true
	$(CONTAINER_ENGINE) rmi $(DOCKER_IMAGE):$(VERSION) 2>/dev/null || true
	$(CONTAINER_ENGINE) rmi $(DOCKER_IMAGE):latest 2>/dev/null || true

# Release targets
release-dry: ## Dry run release (for testing)
	@echo "Running GoReleaser in dry-run mode..."
	goreleaser release --snapshot --clean --skip-publish

release-snapshot: ## Create snapshot release
	@echo "Creating snapshot release..."
	goreleaser release --snapshot --clean

release-check: ## Check release configuration
	@echo "Checking GoReleaser configuration..."
	goreleaser check

# Compose targets (for development)
compose-up: ## Start development environment with docker-compose
	@echo "Starting development environment..."
	docker-compose up -d

compose-down: ## Stop development environment
	@echo "Stopping development environment..."
	docker-compose down

compose-logs: ## Show logs from development environment
	docker-compose logs -f

# Security targets
security-scan: ## Run security scan with gosec
	@echo "Running security scan..."
	gosec ./...

security-deps: ## Check for vulnerable dependencies
	@echo "Checking for vulnerable dependencies..."
	go list -json -deps ./... | nancy sleuth

vuln-check: ## Check for vulnerabilities using govulncheck
	@echo "Running vulnerability check..."
	govulncheck ./...