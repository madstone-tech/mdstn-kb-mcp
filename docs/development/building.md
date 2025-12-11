# Building & Testing

Complete guide to building, testing, and developing kbVault.

## Prerequisites

- **Go 1.25+** - https://golang.org/dl/
- **Make** - Build automation
- **Git** - Version control
- **golangci-lint** - Linting (optional but recommended)

## Project Structure

```
kbvault/
├── cmd/kbvault/              # CLI application
├── pkg/                       # Public packages
├── internal/                  # Private packages
├── docs/                      # Documentation
├── scripts/                   # Build scripts
├── completions/               # Shell completions
├── configs/                   # Configuration templates
├── test/                      # Test data
├── Makefile                   # Build automation
├── go.mod / go.sum           # Dependencies
└── .github/workflows/        # CI/CD pipelines
```

## Building

### Quick Build

```bash
# Build binary to bin/kbvault
make build

# Binary location
./bin/kbvault --version
```

### Build with Custom Flags

```bash
# Build with version info (recommended)
go build -ldflags "\
  -X main.version=v1.0.0 \
  -X main.commitHash=abc123 \
  -X main.buildTime=2025-12-11T10:30:00Z" \
  -o bin/kbvault ./cmd/kbvault

# Run to verify
./bin/kbvault --version
```

### Release Build

```bash
# Using GoReleaser (automated)
goreleaser build --single-target

# Manual release build
make build-release
```

## Testing

### Run All Tests

```bash
# Full test suite with coverage
make test

# Run tests with race detector
go test -race ./...

# Run tests with verbose output
go test -v ./...
```

### Run Specific Tests

```bash
# Test single package
go test -v ./pkg/types

# Test single test function
go test -v ./pkg/types -run TestNote

# Run with pattern matching
go test -v ./... -run "TestConfig*"

# Test with coverage report
go test -cover ./...
```

### Test Coverage

```bash
# Generate coverage report
make test-coverage

# Check coverage percentage
go tool cover -func=coverage.out

# Create HTML coverage report
go tool cover -html=coverage.out -o coverage.html
```

**Coverage Requirements:**
- Minimum: 50% (enforced by CI)
- Target: 70%+
- Critical paths: 90%+

## Linting

### Run Linter

```bash
# Full linting
golangci-lint run

# Lint with timeout
golangci-lint run --timeout=5m

# Lint specific package
golangci-lint run ./cmd/kbvault
```

### Common Issues

**Line too long:**
```bash
# gofmt enforces 80-column style
gofmt -l ./cmd ./pkg ./internal
```

**Unused imports:**
```bash
# Use goimports to clean up imports
goimports -w ./cmd/kbvault/main.go
```

**Interface implementation:**
```bash
# Verify interface implementation
go vet ./...
```

## Formatting

### Format Code

```bash
# Auto-format all Go files
make fmt

# Format specific directory
gofmt -w ./cmd

# Format and simplify code
gofmt -s -w ./cmd

# Check formatting without modifying
gofmt -l ./cmd
```

## Dependencies

### View Dependencies

```bash
# List all dependencies
go list -m all

# Show dependency tree
go mod graph

# Check for updates
go list -u -m all
```

### Update Dependencies

```bash
# Update specific dependency
go get -u github.com/spf13/cobra@latest

# Update all dependencies
go get -u ./...

# Update to specific version
go get github.com/aws/aws-sdk-go-v2@v1.20.0
```

### Verify Dependencies

```bash
# Verify go.mod and go.sum
go mod verify

# Clean up unused dependencies
go mod tidy

# Check for vulnerabilities
go list -json -m all | nancy sleuth
```

## Development Workflow

### 1. Set Up Development Environment

```bash
# Clone repository
git clone https://github.com/madstone-tech/mdstn-kb-mcp.git
cd mdstn-kb-mcp

# Install dependencies
go mod download

# Verify setup
go test ./...
```

### 2. Make Changes

```bash
# Create feature branch
git checkout -b feature/my-feature

# Make code changes
# Add tests
# Update documentation
```

### 3. Test Locally

```bash
# Run affected tests
go test -v ./pkg/config

# Run linter
golangci-lint run

# Format code
make fmt

# Build binary
make build

# Test manual functionality
./bin/kbvault --version
```

### 4. Commit Changes

```bash
# Stage changes
git add .

# Commit with message
git commit -m "feat: add new feature"

# Push to remote
git push origin feature/my-feature
```

### 5. CI/CD Pipeline

```bash
# GitHub Actions runs automatically
# Checks:
# - go test ./...
# - golangci-lint run
# - gofmt verification
# - coverage >= 50%
```

## Debugging

### Enable Debug Logging

```bash
# Set debug environment variable (when supported)
KBVAULT_DEBUG=true ./bin/kbvault list

# Or use verbose flag
./bin/kbvault -v list
```

### Debug Tests

```bash
# Run single test with debugging
go test -v -run TestName -count=1 ./pkg/types

# Debug output
t.Logf("Debug info: %v", value)

# Temporary: Use fmt.Println for debugging
fmt.Println("DEBUG: value =", value)
```

### Profiling

```bash
# CPU profiling
go test -cpuprofile=cpu.prof ./pkg/storage

# Memory profiling
go test -memprofile=mem.prof ./pkg/storage

# Analyze profile
go tool pprof cpu.prof
```

## Common Make Targets

```bash
make              # Show all targets
make build        # Build binary
make test         # Run all tests
make fmt          # Format code
make check        # Run linter and format check
make clean        # Remove build artifacts
make test-coverage # Generate coverage report
make help         # Show help
```

## Git Workflow

### Commit Message Format

Follow conventional commits:

```
<type>: <description>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `style`: Code style
- `refactor`: Code refactoring
- `test`: Tests
- `chore`: Maintenance

**Examples:**
```
feat: add profile management system
fix: resolve S3 credential validation error
docs: update API documentation
refactor: simplify storage interface
test: add comprehensive profile tests
```

### Branch Naming

```
feature/name-of-feature
bugfix/name-of-bug
docs/what-is-documented
refactor/what-is-refactored
```

## Code Style

### Import Order

```go
import (
    // Standard library
    "fmt"
    "os"
    
    // Third-party packages
    "github.com/spf13/cobra"
    "github.com/aws/aws-sdk-go-v2/aws"
    
    // Local packages
    "github.com/madstone-tech/mdstn-kb-mcp/pkg/types"
)
```

### Naming Conventions

```go
// Constants
const MaxCacheSize = 1024

// Interfaces (end with -er)
type Reader interface {}

// Error types (end with Error)
type StorageError struct {}

// Functions (camelCase)
func readNote() {}

// Private functions (camelCase, lowercase first letter)
func loadConfig() {}

// Public functions (PascalCase)
func LoadConfig() {}
```

### Error Handling

```go
// Check all errors
if err != nil {
    return fmt.Errorf("failed to do something: %w", err)
}

// Use typed errors
if _, ok := err.(*types.NotFoundError); ok {
    // Handle specific error
}
```

## Documentation

### Code Comments

```go
// LoadNote loads a note from storage.
// It returns an error if the note doesn't exist.
func LoadNote(id string) (*Note, error) {
    // implementation
}
```

### Package Documentation

```go
// Package types contains core data structures for kbVault.
package types
```

### Keep Docs Updated

When making changes:
1. Update relevant documentation files
2. Update code comments
3. Update CHANGELOG (if applicable)
4. Update package README if adding new features

## Performance Optimization

### Identify Bottlenecks

```bash
# Profile CPU usage
go test -cpuprofile=cpu.prof -bench BenchmarkSearch ./pkg/search

# Analyze results
go tool pprof -http=:8080 cpu.prof
```

### Benchmarking

```go
func BenchmarkSearch(b *testing.B) {
    // Setup
    engine := search.NewEngine(...)
    
    // Run benchmark
    for i := 0; i < b.N; i++ {
        engine.Search(ctx, "query")
    }
}
```

**Run benchmarks:**
```bash
go test -bench=. -benchmem ./pkg/search
```

## Troubleshooting

### Build Fails

```bash
# Clean build artifacts
make clean

# Verify Go installation
go version

# Update dependencies
go mod tidy

# Rebuild
make build
```

### Tests Fail

```bash
# Run single failing test
go test -v -run TestName ./package

# Run with race detector
go test -race ./package

# Check environment
env | grep KBVAULT
```

### Linting Fails

```bash
# Check what golangci-lint complains about
golangci-lint run --print-issued-lines ./...

# Auto-fix where possible
gofmt -w .
```

## CI/CD Pipeline

The repository uses GitHub Actions for:

1. **Testing** - Runs `go test -race ./...`
2. **Linting** - Runs golangci-lint
3. **Coverage** - Ensures minimum 50% coverage
4. **Building** - Builds binaries for multiple platforms
5. **Release** - Creates releases on version tags

See `.github/workflows/` for configuration.

## Security

### Dependency Scanning

```bash
# Check for known vulnerabilities
go list -json -m all | nancy sleuth

# Update vulnerable packages
go get -u package@version
```

### Code Review

All changes require:
- Tests with good coverage
- Documentation updates
- Code review approval
- Green CI checks

---

## See Also

- [Architecture Overview](../architecture/overview.md)
- [Contributing Guide](../../CONTRIBUTING.md)
- [CLI Reference](../guides/cli-reference.md)
