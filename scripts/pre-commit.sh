#!/bin/bash

# Pre-commit script that mimics GitHub Actions CI checks
# Run this before pushing to catch issues early

set -e

echo "🔍 Running pre-commit checks..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${2}${1}${NC}"
}

print_status "1. Checking Go module consistency..." $YELLOW
go mod tidy
if ! git diff --exit-code go.mod go.sum; then
    print_status "❌ go.mod/go.sum are not tidy. Run 'go mod tidy' and commit the changes." $RED
    exit 1
fi
print_status "✅ Go modules are tidy" $GREEN

print_status "2. Verifying dependencies..." $YELLOW
go mod verify
print_status "✅ Dependencies verified" $GREEN

print_status "3. Running go vet..." $YELLOW
go vet ./...
print_status "✅ go vet passed" $GREEN

print_status "4. Running tests with race detection..." $YELLOW
go test -v -race -coverprofile=coverage.out ./...
print_status "✅ Tests passed" $GREEN

print_status "5. Checking test coverage..." $YELLOW
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
echo "Test coverage: ${COVERAGE}%"
# Use 50% for development, CI will enforce 80%
if awk "BEGIN {exit !($COVERAGE < 50)}"; then
    print_status "❌ Test coverage is below 50% (${COVERAGE}%)" $RED
    exit 1
fi
print_status "✅ Test coverage is above 50%" $GREEN

print_status "6. Running security checks..." $YELLOW
if command -v govulncheck >/dev/null 2>&1; then
    govulncheck ./...
    print_status "✅ Security checks passed" $GREEN
else
    print_status "⚠️  govulncheck not found, skipping (install with: go install golang.org/x/vuln/cmd/govulncheck@latest)" $YELLOW
fi

print_status "7. Running golangci-lint..." $YELLOW
if command -v golangci-lint >/dev/null 2>&1; then
    golangci-lint run --timeout=5m
    print_status "✅ Linting passed" $GREEN
else
    print_status "⚠️  golangci-lint not found, skipping (install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)" $YELLOW
fi

print_status "8. Building binary..." $YELLOW
make build
print_status "✅ Build successful" $GREEN

print_status "9. Testing binary..." $YELLOW
./bin/kbvault --version > /dev/null
print_status "✅ Binary test passed" $GREEN

print_status "10. Testing Docker build..." $YELLOW
if command -v docker >/dev/null 2>&1 || command -v podman >/dev/null 2>&1; then
    make docker-build > /dev/null 2>&1
    print_status "✅ Docker build successful" $GREEN
else
    print_status "⚠️  Docker/Podman not found, skipping Docker build test" $YELLOW
fi

print_status "🎉 All pre-commit checks passed! Ready to push." $GREEN